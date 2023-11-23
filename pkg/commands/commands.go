package commands

import (
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/spf13/cobra"
	"strconv"
)

var cLogger log.Logger

func init() {
	cLogger, _ = log.NewLogger(log.ConsoleLogger, "")
}

// ConsoleCommands returns all commands used by the console
func ConsoleCommands() *cobra.Command {

	root := &cobra.Command{}

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

	cmdBuilders := &cobra.Command{
		Use:   "builders [names...]",
		Short: "list installed builders",
		Run: func(cmd *cobra.Command, args []string) {
			buildersCmd(args)
		},
	}
	cmdAgents := &cobra.Command{
		Use:   "agents [names...]",
		Short: "list compiled agents",
		Run: func(cmd *cobra.Command, args []string) {
			agentsCmd(args)
		},
	}

	cmdSessions := &cobra.Command{
		Use:   "sessions [ids...]",
		Short: "list established agent connections",
		Run: func(cmd *cobra.Command, args []string) {
			sessionsCmd(args)
		},
	}

	cmdUse := &cobra.Command{
		Use:   "use [id]",
		Short: "initiate an interactive session with an agent",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				cLogger.Error("%s is not a valid session ID", args[0])
				return
			}
			useCmd(id)
		},
	}
	root.AddCommand(cmdSessions, cmdUse, cmdAgents, cmdBuilders, cmdBuild, cmdExit)
	root.CompletionOptions.HiddenDefaultCmd = true
	return root
}

// exits any named menus spawned by any commands
func exit(short string) *cobra.Command {
	if len(short) == 0 {
		short = "exit the interactive menu"
	}
	cmd := &cobra.Command{
		Use:   "exit",
		Short: short,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			console.MainMenu()
		},
	}
	return cmd
}
