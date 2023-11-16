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
	"text/tabwriter"
)

var (
	l             log.Logger
	builderConfig struct {
		client  rpcpb.BuilderClient
		request *rpcpb.BuildRequest
		options []*rpcpb.Option
	}
	w = tabwriter.NewWriter(os.Stdout, 1, 1, 2, ' ', 0)
)

func init() {
	l, _ = log.NewLogger(log.ConsoleLogger, "")
}

// BuildCmd to start the interactive agent builder
func BuildCmd(builderName string) {
	builder := &db.Builder{}
	if err := db.FindOneConditional("name = ?", builderName, &builder); err != nil {
		// Search using builderName as either name or ID
		if err = db.FindOneConditional("builder_id = ?", builderName, &builder); err != nil {
			l.Error("failed to retrieve specified agent: %v", err)
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
	console.BuildMode(builder.Name)
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
	for _, k := range optionsReply.GetOptions() {
		// Do this so that we can quickly check if an option is valid using map indexing, check is done in SetCmd
		builderConfig.request.Options[k.Name] = ""
	}
	builderConfig.options = optionsReply.GetOptions()
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
	// TODO:builder service request (RPC)
}

func ConsoleCommands() []*cobra.Command {
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
