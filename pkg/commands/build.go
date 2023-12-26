package commands

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/desertbit/grumble"
	"github.com/pygrum/monarch/pkg/completion"
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/consts"
	"github.com/pygrum/monarch/pkg/crypto"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/protobuf/builderpb"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
	"google.golang.org/grpc"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"
)

var (
	l             log.Logger
	internals     = []string{"id", "ca_cert"}
	ValidTypes    = []string{"bool", "int", "string", "float"}
	builderConfig *BuilderConfig
)

func init() {
	l, _ = log.NewLogger(log.ConsoleLogger, "")
}

type BuilderConfig struct {
	builderID string // ID of builder
	name      string
	version   string
	ID        string // ID of resulting agent
	request   *builderpb.BuildRequest
	options   []*builderpb.Option
}

func buildCheck() error {
	if builderConfig == nil {
		return errors.New("no builder has been loaded yet")
	}
	return nil
}

// buildCmd to start the interactive agent builder
func buildCmd(builderName string) {
	builders, err := console.Rpc.Builders(ctx, &clientpb.BuilderRequest{BuilderId: []string{builderName}})
	if err != nil {
		cLogger.Error("%v", err)
		return
	}
	if len(builders.Builders) == 0 {
		l.Error("could not find a builder with the specified name or ID (%s)", builderName)
		return
	}
	builder := builders.Builders[0]
	if err := loadBuildOptions(builder); err != nil {
		l.Error("failed to load build options for %s: %v", builderName, err)
		return
	}
	prompt := "monarch " + "(" + builderName + ") > "
	console.App.SetPrompt(prompt)
}

// Returns default builder options
func defaultOptions() []*builderpb.Option {
	var options []*builderpb.Option
	ID := &builderpb.Option{
		Name:        "id",
		Description: "[internal] the agent ID assigned to the build (immutable)",
		Default:     builderConfig.ID,
		Type:        "string",
		Required:    true,
	}
	name := &builderpb.Option{
		Name:        "name",
		Description: "name of this particular agent instance",
		Default:     builderConfig.ID,
		Type:        "string",
		Required:    false,
	}
	OS := &builderpb.Option{
		Name:        "os",
		Description: "the OS that the build targets",
		Default:     runtime.GOOS,
		Required:    false,
	}
	arch := &builderpb.Option{
		Name:        "arch",
		Description: "the platform architecture that the build targets",
		Default:     "amd64",
		Type:        "string",
		Required:    false,
	}
	def := config.MainConfig.Interface
	if def == "" {
		def = config.ClientConfig.RHost
	}
	host := &builderpb.Option{
		Name:        "host",
		Description: "the host that the agent calls back to",
		Default:     def,
		Type:        "string",
		Required:    false,
	}
	port := &builderpb.Option{
		Name:        "port",
		Description: "the port on which to connect to the host on callback",
		Default:     "8000",
		Type:        "int",
		Required:    false,
	}
	out := &builderpb.Option{
		Name:        "outfile",
		Description: "the name of the resulting binary",
		Default:     "",
		Type:        "string",
		Required:    false,
	}
	caCert := &builderpb.Option{
		Name:        "ca_cert",
		Description: "[internal] the certificate authority used to sign server certificates (immutable)",
		Default:     "***",
		Type:        "string",
		Required:    true,
	}
	options = append(options, ID, name, OS, arch, host, port, caCert, out)
	return options
}

// loadBuildOptions loads the build options into the BuildRequest in the builderConfig variable
func loadBuildOptions(b *clientpb.Builder) error {
	// initialize pointer
	builderConfig = &BuilderConfig{
		ID: agentID(),
		request: &builderpb.BuildRequest{
			Options: make(map[string]string),
		},
	}
	optionsReply, err := console.Rpc.Options(ctx, &builderpb.OptionsRequest{
		BuilderId: builderConfig.ID + b.BuilderId,
	})
	if err != nil {
		return err
	}
	builderConfig.options = optionsReply.GetOptions()
	// Add default options
	builderConfig.options = append(builderConfig.options, defaultOptions()...)
	for i, k := range builderConfig.options {
		k.Type = strings.ToLower(k.Type)
		if !slices.Contains(ValidTypes, k.Type) && k.Type != "" {
			return fmt.Errorf("%s has an invalid type '%s'. Please report this issue to the maintainer",
				k.Name, k.Type)
		}
		k.Name = strings.ToLower(k.Name)
		builderConfig.options[i].Name = strings.ToLower(k.Name)
		_, ok := builderConfig.request.Options[k.Name]
		if ok {
			l.Warn("duplicate instance(s) of option: %s", k.Name)
			continue
		}
		builderConfig.request.Options[k.Name] = k.Default
	}
	builderConfig.builderID = b.BuilderId
	builderConfig.name = b.Name
	builderConfig.version = b.Version
	return nil
}

// setCmd - for users to edit build configuration
func setCmd(name, value string) {
	name = strings.ToLower(name)
	_, ok := builderConfig.request.Options[name]
	if !ok {
		l.Error("'%s' is not an option", name)
		return
	}
	// o(n^2) search for valid option ;(
	for _, o := range builderConfig.options {
		if o.Name == name {
			found := false
			if len(o.Choices) == 0 {
				found = true
			}
			for _, c := range o.Choices {
				if c == value {
					found = true
					break
				}
			}
			if !found {
				l.Error("'%s' is not a valid option. choose between one of the options below:", name)
				fmt.Println(strings.Join(o.Choices, ", "))
				return
			}
			if err := TypeVerify(o.Type, value); err != nil {
				l.Error("error setting %s as '%s': %v", name, value, err)
				return
			}
		}
	}
	if slices.Contains(internals, name) {
		l.Error("'%s' is used internally by the engine and cannot be changed", name)
		return
	}
	builderConfig.request.Options[name] = value
}

func TypeVerify(t, value string) error {
	switch t {
	case "int":
		if _, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("not a valid integer")
		}

	case "bool":
		if _, err := strconv.ParseBool(value); err != nil {
			return fmt.Errorf("not a valid boolean")
		}

	case "float":
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return fmt.Errorf("not a valid float")
		}

	// Accept anything if it is a string type, or no type set
	default:
		break
	}
	return nil
}

// unsetCmd - unsets variables in build config
func unsetCmd(name string) {
	// must still do option checks as we do not want map to grow by spamming unsets, or add invalid options
	setCmd(name, "")
}

// optionsCmd - returns all build configuration options
func optionsCmd() {
	header := "NAME\tVALUE\tDESCRIPTION\tTYPE\tREQUIRED\tCHOICES\t"
	_, _ = fmt.Fprintln(w, header)
	for _, option := range builderConfig.options {
		tableLine := fmt.Sprintf("%s\t%s\t%s\t%v\t%v\t%v\t",
			option.Name,
			// color set options green
			builderConfig.request.Options[option.Name],
			option.Description,
			option.Type,
			option.Required,
			strings.Join(option.Choices, ", "))

		_, _ = fmt.Fprintln(w, tableLine)
	}
	// Print full table
	_ = w.Flush()
}

// build actually builds. duh
func build() {
	var required []string
	for _, option := range builderConfig.options {
		// if unset, enforce requirement
		if builderConfig.request.Options[option.Name] == "" {
			if option.Required {
				required = append(required, option.Name)
			}
		}
	}
	if len(required) != 0 {
		l.Error("the following requiredFlag options have not been set:")
		for _, o := range required {
			fmt.Println(o)
		}
		return
	}
	// add internal option 'ca_cert'
	var caCert []byte
	var caErr error
	// check with three as it is represented by three asterisks
	if len(config.ClientConfig.CaCertPEM) <= 3 {
		caCert, _, caErr = crypto.CaCertKeyPair()
		if caErr != nil {
			cLogger.Error("failed to set internal parameter ca_cert: %v", caErr)
		}
	} else {
		caCert = config.ClientConfig.CaCertPEM
	}
	builderConfig.request.Options["ca_cert"] = base64.StdEncoding.EncodeToString(caCert)
	buildCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	// uses both agent id and builder id for unique identifier for each build session
	// receive large bins
	builderConfig.request.BuilderId = builderConfig.ID + builderConfig.builderID
	maxSizeOption := grpc.MaxCallRecvMsgSize(32 * 10e6)
	buildResponse, err := console.Rpc.Build(buildCtx, builderConfig.request, maxSizeOption)
	if err != nil {
		l.Error("[RPC] failed to build agent: %v", err)
		return
	}
	// reset CA cert to default
	builderConfig.request.Options["ca_cert"] = "***"
	resp := buildResponse.Reply
	if resp.Status == builderpb.Status_FailedWithMessage {
		l.Error("build failed: %s", resp.Error)
		return
	}
	outfile := filepath.Join(os.TempDir(), builderConfig.request.Options["outfile"])
	var out *os.File
	if len(builderConfig.request.Options["outfile"]) == 0 {
		out, err = os.CreateTemp("", "*."+builderConfig.name)
	} else {
		out, err = os.Create(outfile)
	}
	if err != nil {
		l.Error("creating temp file failed: %v", err)
		return
	}
	defer out.Close()
	_, err = out.Write(resp.Build)
	if err != nil {
		l.Error("failed to save build to %s: %v", out.Name(), err)
		return
	}
	// save to agents table
	agent := &clientpb.Agent{
		AgentId:   builderConfig.ID,
		Name:      builderConfig.request.Options["name"],
		Version:   builderConfig.version,
		OS:        builderConfig.request.Options["os"],
		Arch:      builderConfig.request.Options["arch"],
		Host:      builderConfig.request.Options["host"],
		Port:      builderConfig.request.Options["port"],
		Builder:   builderConfig.builderID,
		File:      buildResponse.ServerFile,
		CreatedBy: config.ClientConfig.UUID,
	}
	if _, err = console.Rpc.NewAgent(ctx, agent); err != nil {
		l.Error("%v", err)
		return
	}
	l.Success("build complete. saved to %s", out.Name())
}

// agentID generates an ID for an agent.
func agentID() string {
	idBytes := make([]byte, 8) // 16l
	_, _ = rand.Read(idBytes)
	return hex.EncodeToString(idBytes)
}

func allOptions() []string {
	var options []string
	for k := range builderConfig.request.Options {
		if slices.Contains(internals, k) {
			continue
		}
		options = append(options, k)
	}
	return options
}

func BuildCommands() []*grumble.Command {
	var rootCmd []*grumble.Command
	cmdOptions := &grumble.Command{
		Name:      "options",
		Help:      "view all modifiable build configuration options",
		HelpGroup: consts.BuildHelpGroup,
		Run: func(c *grumble.Context) error {
			if err := buildCheck(); err != nil {
				return err
			}
			optionsCmd()
			return nil
		},
	}
	cmdSet := &grumble.Command{
		Name: "set",
		Args: func(a *grumble.Args) {
			a.String("option", "the build option to change")
			a.String("value", "the value to set the option to")
		},
		Help:      "set a build option to the provided value",
		HelpGroup: consts.BuildHelpGroup,
		Run: func(c *grumble.Context) error {
			if err := buildCheck(); err != nil {
				return err
			}
			setCmd(c.Args.String("option"), c.Args.String("value"))
			return nil
		},
		Completer: func(prefix string, args []string) []string {
			if err := buildCheck(); err != nil {
				return nil
			}
			if len(args) == 0 {
				return completion.Options(prefix, allOptions())
			}
			return nil
		},
	}

	cmdUnset := &grumble.Command{
		Name: "unset",
		Args: func(a *grumble.Args) {
			a.String("option", "the build option to unset")
		},
		Help:      "unset a build option",
		HelpGroup: consts.BuildHelpGroup,
		Run: func(c *grumble.Context) error {
			if err := buildCheck(); err != nil {
				return err
			}
			unsetCmd(c.Args.String("option"))
			return nil
		},
		Completer: func(prefix string, args []string) []string {
			if err := buildCheck(); err != nil {
				return nil
			}
			return completion.Options(prefix, allOptions())
		},
	}

	cmdBuild := &grumble.Command{
		Name:      "compile",
		Help:      "builds the agent using the provided configuration options",
		HelpGroup: consts.BuildHelpGroup,
		Run: func(c *grumble.Context) error {
			if err := buildCheck(); err != nil {
				return err
			}
			build()
			return nil
		},
	}
	cmdEndBuild := &grumble.Command{
		Name:      "end-build",
		Help:      "exit the interactive builder",
		HelpGroup: consts.BuildHelpGroup,
		Run: func(c *grumble.Context) error {
			if err := buildCheck(); err != nil {
				return err
			}
			if _, err := console.Rpc.EndBuild(ctx, &builderpb.BuildRequest{
				BuilderId: builderConfig.ID + builderConfig.builderID,
			}); err != nil {
				cLogger.Error("failed to delete builder client for %s: %v", builderConfig.builderID, err)
			}
			builderConfig = nil
			console.App.SetDefaultPrompt()
			return nil
		},
	}
	rootCmd = append(rootCmd, cmdBuild, cobraProfilesCmd(), cmdOptions, cmdSet, cmdUnset,
		cmdEndBuild)
	return rootCmd
}
