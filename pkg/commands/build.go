package commands

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/docker"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/rpcpb"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

var (
	l          log.Logger
	immutables = []string{"id"}
)

var builderConfig struct {
	builderID string // ID of builder
	name      string
	version   string
	ID        string // ID of resulting agent
	client    rpcpb.BuilderClient
	request   *rpcpb.BuildRequest
	options   []*rpcpb.Option
}

func init() {
	l, _ = log.NewLogger(log.ConsoleLogger, "")
}

// buildCmd to start the interactive agent builder
func buildCmd(builderName string) {
	builder := &db.Builder{}
	if err := db.FindOneConditional("name = ?", builderName, &builder); err != nil {
		// Search using builderName as either name or ID
		if err = db.FindOneConditional("builder_id = ?", builderName, &builder); err != nil {
			l.Error("could not find builder for '%s': %v", builderName, err)
			return
		}
	}
	if len(builder.BuilderID) == 0 {
		l.Error("could not find a builder with the specified name or ID (%s)", builderName)
		return
	}
	if err := loadBuildOptions(builder); err != nil {
		l.Error("failed to load build options for %s: %v", builderName, err)
	}
	console.NamedMenu(builder.Name, consoleCommands())
}

// Returns default builder options
func defaultOptions() []*rpcpb.Option {
	var options []*rpcpb.Option
	ID := &rpcpb.Option{
		Name:        "id",
		Description: "[immutable] the agent ID assigned to the build",
		Default:     builderConfig.ID,
		Required:    true,
	}
	name := &rpcpb.Option{
		Name:        "name",
		Description: "name of this particular agent instance",
		Default:     builderConfig.ID,
		Required:    false,
	}
	OS := &rpcpb.Option{
		Name:        "os",
		Description: "the OS that the build targets",
		Default:     "",
		Required:    false,
	}
	arch := &rpcpb.Option{
		Name:        "arch",
		Description: "the platform architecture that the build targets",
		Default:     "x64",
		Required:    false,
	}
	host := &rpcpb.Option{
		Name:        "host",
		Description: "the host that the agent calls back to",
		Default:     config.MainConfig.Interface,
		Required:    false,
	}
	port := &rpcpb.Option{
		Name:        "port",
		Description: "the port on which to connect to the host on callback",
		Default:     "",
		Required:    false,
	}
	out := &rpcpb.Option{
		Name:        "outfile",
		Description: "the name of the resulting binary",
		Default:     "",
		Required:    false,
	}
	options = append(options, ID, name, OS, arch, host, port, out)
	return options
}

// loadBuildOptions loads the build options into the BuildRequest in the builderConfig variable
func loadBuildOptions(b *db.Builder) error {
	ctx := context.Background()
	builderRPC, err := docker.RPCAddresses(docker.Cli, ctx, b.BuilderID)
	if err != nil {
		return err
	}
	conn, err := grpc.Dial(builderRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	client := rpcpb.NewBuilderClient(conn)
	optionsReply, err := client.GetOptions(ctx, &rpcpb.OptionsRequest{})
	if err != nil {
		return err
	}
	builderConfig.options = optionsReply.GetOptions()
	// Add default options
	builderConfig.options = append(builderConfig.options, defaultOptions()...)
	for i, k := range builderConfig.options {
		k.Name = strings.ToLower(k.Name)
		builderConfig.options[i].Name = strings.ToLower(k.Name)

		_, ok := builderConfig.request.Options[k.Name]
		if ok {
			l.Warn("duplicate instance(s) of option: %s", k.Name)
			continue
		}
		// Do this so that we can quickly check if an option is valid using map indexing, check is done in setCmd
		builderConfig.request.Options[k.Name] = ""
	}
	builderConfig.builderID = b.BuilderID
	builderConfig.name = b.Name
	builderConfig.version = b.Version
	builderConfig.ID = agentID()
	builderConfig.client = client
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
	if slices.Contains(immutables, name) {
		l.Error("'%s' is enforced by the engine and cannot be changed", name)
		return
	}
	builderConfig.request.Options[name] = value
}

// unsetCmd - unsets variables in build config
func unsetCmd(name string) {
	// must still do option checks as we do not want map to grow by spamming unsets, or add invalid options
	setCmd(name, "")
}

// optionsCmd - returns all build configuration options
func optionsCmd() {
	headers := "NAME\tVALUE\tDESCRIPTION\tREQUIRED\t"
	_, _ = fmt.Fprintln(w, headers)
	for _, option := range builderConfig.options {
		tableLine := fmt.Sprintf("%s\t%s\t%s\t%v\t",
			option.Name,
			// color set options green
			fmt.Sprintf("%s%s%s", "\033[32m", builderConfig.request.Options[option.Name], "\033[0m"),
			option.Description, option.Required)

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
	resp, err := builderConfig.client.BuildAgent(context.Background(), builderConfig.request)
	if err != nil {
		l.Error("[RPC] failed to build agent: %v", err)
		return
	}
	if resp.Status == rpcpb.Status_FailedWithMessage {
		l.Error("build failed: %s", resp.Error)
		return
	}
	outfile := filepath.Join(os.TempDir(), builderConfig.request.Options["outfile"])
	var out *os.File
	if len(outfile) == 0 {
		out, err = os.CreateTemp(os.TempDir(), "*."+builderConfig.name)
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
	agent := &db.Agent{
		AgentID:   builderConfig.ID,
		Name:      builderConfig.request.Options["name"],
		Version:   builderConfig.version,
		OS:        builderConfig.request.Options["os"],
		Arch:      builderConfig.request.Options["arch"],
		Host:      builderConfig.request.Options["host"],
		Port:      builderConfig.request.Options["port"],
		Builder:   builderConfig.ID,
		File:      out.Name(),
		CreatedAt: time.Now(),
	}
	if err = db.Create(agent); err != nil {
		l.Error("failed to save agent instance: %v", err)
	}
}

// agentID generates an ID for an agent.
func agentID() string {
	idBytes := make([]byte, 16) // 32l
	_, _ = rand.Read(idBytes)
	return hex.EncodeToString(idBytes)
}

func consoleCommands() []*cobra.Command {
	var cmds []*cobra.Command
	cmdOptions := &cobra.Command{
		Use:   "options",
		Args:  cobra.NoArgs,
		Short: "view all modifiable build configuration options",
		Run: func(cmd *cobra.Command, args []string) {
			optionsCmd()
		},
	}
	cmdSet := &cobra.Command{
		Use:   "set [option] [value]",
		Args:  cobra.ExactArgs(2),
		Short: "set a build option to the provided value",
		Run: func(cmd *cobra.Command, args []string) {
			setCmd(args[0], args[1])
		},
	}
	cmdUnset := &cobra.Command{
		Use:   "unset [option]",
		Args:  cobra.ExactArgs(1),
		Short: "unsets the provided option",
		Run: func(cmd *cobra.Command, args []string) {
			unsetCmd(args[0])
		},
	}
	cmdBuild := &cobra.Command{
		Use:   "build",
		Args:  cobra.NoArgs,
		Short: "builds the agent using the provided configuration options",
		Run: func(cmd *cobra.Command, args []string) {
			build()
		},
	}
	cmds = append(cmds, cmdBuild, cmdOptions, cmdSet, cmdUnset)
	return cmds
}
