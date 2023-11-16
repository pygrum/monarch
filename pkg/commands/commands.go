package commands

import (
	"github.com/pygrum/monarch/pkg/log"
	"github.com/spf13/cobra"
)

var cLogger log.Logger

func init() {
	cLogger, _ = log.NewLogger(log.ConsoleLogger, "")
}

// ConsoleCommands returns all commands used by the console
func ConsoleCommands() []*cobra.Command {

	var root []*cobra.Command

	cmdExit := &cobra.Command{
		Use:   "exit",
		Short: "shutdown the monarch server",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			exitCmd()
		},
	}
	cmdBuild := &cobra.Command{
		Use:   "build [agent]",
		Short: "start the interactive builder for an installed agent",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			buildCmd(args[0])
		},
	}

	var buildersShowTranslator bool
	cmdBuilders := &cobra.Command{
		Use:   "builders",
		Short: "list installed builders",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			buildersCmd(buildersShowTranslator)
		},
	}
	cmdBuilders.Flags().BoolVarP(&buildersShowTranslator, "show-translator", "t", false,
		"show corresponding translator IDs")

	cmdAgents := &cobra.Command{
		Use:   "agents [names...]",
		Short: "list compiled agents",
		Run: func(cmd *cobra.Command, args []string) {
			agentsCmd(args)
		},
	}

	root = append(root, cmdAgents, cmdBuilders, cmdBuild, cmdExit)
	return root
}
