package commands

import (
	"fmt"
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/transport"
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

	var yesExit bool
	cmdExit := &cobra.Command{
		Use:   "exit",
		Short: "shutdown the monarch server",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			exitCmd(yesExit)
		},
	}
	cmdExit.Flags().BoolVarP(&yesExit, "yes", "y", false, "confirm exit")

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
		Use:   "agents [flags] AGENTS...",
		Short: "list compiled agents",
		Run: func(cmd *cobra.Command, args []string) {
			agentsCmd(args)
		},
	}
	cmdAgentsRm := &cobra.Command{
		Use:   "rm [flags] AGENTS...",
		Short: "remove compiled agents from listing",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmdRm(args)
		},
	}
	cmdAgents.AddCommand(cmdAgentsRm)

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
	var httpStop bool
	cmdHttp := &cobra.Command{
		Use:   "http",
		Short: "start an http listener",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			httpCmd(httpStop)
		},
	}
	cmdHttp.Flags().BoolVarP(&httpStop, "stop", "s", false, "stop the http listener")

	var httpsStop bool
	cmdHttps := &cobra.Command{
		Use:   "https",
		Short: "start an https listener",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			httpsCmd(httpsStop)
		},
	}
	cmdHttps.Flags().BoolVarP(&httpsStop, "stop", "s", false, "stop the https listener")

	var installPrivate bool
	var branch string
	cmdInstall := &cobra.Command{
		Use:   "install [flags] REPO",
		Short: "install a builder from a Git repository",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			installCmd(args[0], branch, installPrivate)
		},
	}
	cmdInstall.Flags().BoolVarP(&installPrivate, "use-creds", "c", false,
		"use GitHub credentials for installation")
	cmdInstall.Flags().StringVarP(&branch, "branch", "b", "", "the branch you wish to "+
		"install")

	cmdLocal := &cobra.Command{
		Use:   "local [flags] FOLDER",
		Short: "install a builder from a local folder",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			localCmd(args[0])
		},
	}
	// it's a subcommand of the 'install' command
	cmdInstall.AddCommand(cmdLocal)

	var purge bool
	cmdUninstall := &cobra.Command{
		Use:   "uninstall [flags] BUILDERS...",
		Short: "uninstall builder(s) by name or ID",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			uninstallCmd(args, purge)
		},
	}
	cmdUninstall.Flags().BoolVarP(&purge, "delete-data", "p", false, "delete the source"+
		" folder that was saved to disk when installed")
	cmdVersion := &cobra.Command{
		Use:   "version",
		Short: "view installed monarch version",
		Run: func(cmd *cobra.Command, args []string) {
			versionCmd()
		},
	}
	var stageAs string
	var format string
	cmdStage := &cobra.Command{
		Use:   "stage [agent]",
		Short: "stage an agent on the configured staging endpoint, or view currently staged agents",
		Args:  cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			stageCmd(args, format, stageAs)
		},
	}
	cmdStage.Flags().StringVar(&stageAs, "as", "", "the file to stage your agent as (e.g. index.php)")
	cmdStage.Flags().StringVarP(&format, "format", "f", "",
		"the format of the staged file - shellcode")

	cmdUnstage := &cobra.Command{
		Use:   "unstage [agent-alias]",
		Short: "unstage a staged agent, by specifying its stage alias (e.g. index.php)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			unstageCmd(args[0])
		},
	}

	root.AddCommand(cmdSessions, cmdUse, cmdHttp, cmdHttps, cmdAgents, cmdBuilders, cmdBuild, cmdInstall, cmdUninstall,
		cmdStage, cmdUnstage, cmdVersion, cmdExit)
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

func info(systemInfo transport.Registration) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "view information about the agent's host",
		Long: "Information is typically compiled and sent by an agent when it first connects to the server. " +
			"This information includes details such as the user running the process, the process ID, UID, GID, " +
			"IP address, and more; however if an agent doesn't transmit this information, you'd have to find out " +
			"yourself.",
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println()
			fmt.Println("System Information")
			fmt.Println("==================")
			_, _ = fmt.Fprintln(w, fmt.Sprintf("%v\t%v\t", "Agent ID:", systemInfo.AgentID))
			_, _ = fmt.Fprintln(w, fmt.Sprintf("%v\t%v\t", "Host OS:", systemInfo.OS))
			_, _ = fmt.Fprintln(w, fmt.Sprintf("%v\t%v\t", "Architecture:", systemInfo.Arch))
			_, _ = fmt.Fprintln(w, fmt.Sprintf("%v\t%v\t", "Username:", systemInfo.Username))
			_, _ = fmt.Fprintln(w, fmt.Sprintf("%v\t%v\t", "Hostname:", systemInfo.Hostname))
			_, _ = fmt.Fprintln(w, fmt.Sprintf("%v\t%v\t", "UID:", systemInfo.UID))
			_, _ = fmt.Fprintln(w, fmt.Sprintf("%v\t%v\t", "GID:", systemInfo.GID))
			_, _ = fmt.Fprintln(w, fmt.Sprintf("%v\t%v\t", "PID:", systemInfo.PID))
			_, _ = fmt.Fprintln(w, fmt.Sprintf("%v\t%v\t", "Home Directory:", systemInfo.HomeDir))
			_ = w.Flush()
			fmt.Println()
		},
	}
	return cmd
}
