package build

import (
	"context"
	"fmt"
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
	"text/tabwriter"
)

var (
	l log.Logger
	w = tabwriter.NewWriter(os.Stdout, 1, 1, 3, ' ', 0)
)

var builderConfig struct {
	name    string // Name of agent
	client  rpcpb.BuilderClient
	request *rpcpb.BuildRequest
	options []*rpcpb.Option
}

func init() {
	l, _ = log.NewLogger(log.ConsoleLogger, "")
}

// BuildCmd to start the interactive agent builder
func BuildCmd(builderName string) {
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
	console.BuildMode(builder.Name, consoleCommands())
}

// Returns default builder options
func defaultOptions() []*rpcpb.Option {
	var options []*rpcpb.Option
	OS := &rpcpb.Option{
		Name:        "OS",
		Description: "the OS that the build targets",
		Default:     "",
		Required:    true, // must still set required so that they can't unset the value then build
	}
	arch := &rpcpb.Option{
		Name:        "arch",
		Description: "the platform architecture that the build targets",
		Default:     "",
		Required:    true,
	}
	out := &rpcpb.Option{
		Name:        "outfile",
		Description: "the name of the resulting binary",
		Default:     "",
		Required:    false,
	}
	options = append(options, OS, arch, out)
	return options
}

// loadBuildOptions loads the build options into the BuildRequest in the builderConfig variable
func loadBuildOptions(b *db.Builder) error {
	ctx := context.Background()
	builderRPC, _, err := docker.RPCAddresses(docker.Cli, ctx, b.BuilderID)
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
	for _, k := range builderConfig.options {
		_, ok := builderConfig.request.Options[k.Name]
		if ok {
			l.Warn("duplicate instance(s) of option: %s", k.Name)
			continue
		}
		// Do this so that we can quickly check if an option is valid using map indexing, check is done in SetCmd
		builderConfig.request.Options[k.Name] = ""
		builderConfig.name = b.Name
	}
	builderConfig.client = client
	return nil
}

// SetCmd - for users to edit build configuration
func SetCmd(name, value string) {
	_, ok := builderConfig.request.Options[name]
	if !ok {
		l.Error("'%s' is not an option", name)
		return
	}
	builderConfig.request.Options[name] = value
}

// UnsetCmd - unsets variables in build config
func UnsetCmd(name string) {
	// must still do option checks as we do not want map to grow by spamming unsets, or add invalid options
	SetCmd(name, "")
}

// OptionsCmd - returns all build configuration options
func OptionsCmd() {
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

// Build actually builds. duh
func Build() {
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
}

func consoleCommands() []*cobra.Command {
	var cmds []*cobra.Command
	cmdOptions := &cobra.Command{
		Use:   "options",
		Args:  cobra.NoArgs,
		Short: "view all modifiable build configuration options",
		Run: func(cmd *cobra.Command, args []string) {
			OptionsCmd()
		},
	}
	cmdSet := &cobra.Command{
		Use:   "set [option] [value]",
		Args:  cobra.ExactArgs(2),
		Short: "set a build option to the provided value",
		Run: func(cmd *cobra.Command, args []string) {
			SetCmd(args[0], args[1])
		},
	}
	cmdUnset := &cobra.Command{
		Use:   "unset [option]",
		Args:  cobra.ExactArgs(1),
		Short: "unsets the provided option",
		Run: func(cmd *cobra.Command, args []string) {
			UnsetCmd(args[0])
		},
	}
	cmdBuild := &cobra.Command{
		Use:   "build",
		Args:  cobra.NoArgs,
		Short: "builds the agent using the provided configuration options",
		Run: func(cmd *cobra.Command, args []string) {
			Build()
		},
	}
	cmds = append(cmds, cmdBuild, cmdOptions, cmdSet, cmdUnset)
	return cmds
}
