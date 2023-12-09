package commands

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/pygrum/monarch/pkg/completion"
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/protobuf/builderpb"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
	"github.com/rsteube/carapace"
	"github.com/spf13/cobra"
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
	immutables    = []string{"id"}
	ValidTypes    = []string{"bool", "int", "string", "float"}
	builderConfig BuilderConfig
)

type BuilderConfig struct {
	builderID string // ID of builder
	name      string
	version   string
	ID        string // ID of resulting agent
	request   *builderpb.BuildRequest
	options   []*builderpb.Option
}

func init() {
	l, _ = log.NewLogger(log.ConsoleLogger, "")
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
	console.NamedMenu(builder.Name, consoleCommands)
}

// Returns default builder options
func defaultOptions() []*builderpb.Option {
	var options []*builderpb.Option
	ID := &builderpb.Option{
		Name:        "id",
		Description: "[immutable] the agent ID assigned to the build",
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
	host := &builderpb.Option{
		Name:        "host",
		Description: "the host that the agent calls back to",
		Default:     config.MainConfig.Interface,
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
	options = append(options, ID, name, OS, arch, host, port, out)
	return options
}

// loadBuildOptions loads the build options into the BuildRequest in the builderConfig variable
func loadBuildOptions(b *clientpb.Builder) error {
	// initialize pointer
	builderConfig = BuilderConfig{
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
	if slices.Contains(immutables, name) {
		l.Error("'%s' is enforced by the engine and cannot be changed", name)
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
	//TODO:Implement build (RPC, requirement checks etc.)
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
		l.Error("the following required options have not been set:")
		for _, o := range required {
			fmt.Println(o)
		}
		return
	}
	buildCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	// uses both agent id and builder id for unique identifier for each build session
	// receive large bins
	builderConfig.request.BuilderId = builderConfig.ID + builderConfig.builderID
	maxSizeOption := grpc.MaxCallRecvMsgSize(32 * 10e6)
	resp, err := console.Rpc.Build(buildCtx, builderConfig.request, maxSizeOption)
	if err != nil {
		l.Error("[RPC] failed to build agent: %v", err)
		return
	}
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
	l.Success("build complete. saved to %s", out.Name())
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
		File:      out.Name(),
		CreatedBy: config.ClientConfig.UUID,
	}
	if _, err = console.Rpc.NewAgent(ctx, agent); err != nil {
		l.Error("%v", err)
		return
	}
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
		if slices.Contains(immutables, k) {
			continue
		}
		options = append(options, k)
	}
	return options
}

func consoleCommands() *cobra.Command {
	rootCmd := &cobra.Command{}
	cmdOptions := &cobra.Command{
		Use:   "options",
		Args:  cobra.NoArgs,
		Short: "view all modifiable build configuration options",
		Run: func(cmd *cobra.Command, args []string) {
			optionsCmd()
		},
	}
	cmdSet := &cobra.Command{
		Use:   "set OPTION VALUE",
		Args:  cobra.ExactArgs(2),
		Short: "set a build option to the provided value",
		Run: func(cmd *cobra.Command, args []string) {
			setCmd(args[0], args[1])
		},
	}
	carapace.Gen(cmdSet).PositionalCompletion(completion.Options(allOptions()))

	cmdUnset := &cobra.Command{
		Use:   "unset [option]",
		Args:  cobra.ExactArgs(1),
		Short: "unsets the provided option",
		Run: func(cmd *cobra.Command, args []string) {
			unsetCmd(args[0])
		},
	}
	carapace.Gen(cmdUnset).PositionalCompletion(completion.Options(allOptions()))

	cmdBuild := &cobra.Command{
		Use:   "build",
		Args:  cobra.NoArgs,
		Short: "builds the agent using the provided configuration options",
		Run: func(cmd *cobra.Command, args []string) {
			build()
		},
	}
	rootCmd.AddCommand(cmdBuild, cobraProfilesCmd(), cmdOptions, cmdSet, cmdUnset,
		exit("exit the interactive builder", "build"))
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	return rootCmd
}
