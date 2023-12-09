package commands

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/protobuf/builderpb"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
	"github.com/spf13/cobra"
)

var (
	ctx     = context.Background()
	cLogger log.Logger
)

func init() {
	cLogger, _ = log.NewLogger(log.ConsoleLogger, "")
}

func ServerConsoleCommands() *cobra.Command {
	root := ConsoleCommands()
	var stop bool
	cmdCoop := &cobra.Command{
		Use:   "coop",
		Short: "start / stop co-op mode",
		Run: func(cmd *cobra.Command, args []string) {
			coopCmd(stop)
		},
	}
	cmdCoop.Flags().BoolVarP(&stop, "stop", "s", false, "turn off co-op mode")

	cmdPlayers := &cobra.Command{
		Use:   "players",
		Short: "list players that have been registered on the server",
		Run: func(cmd *cobra.Command, args []string) {
			playersCmd(args)
		},
	}
	var name, lhost string
	cmdPlayersNew := &cobra.Command{
		Use:   "new",
		Short: "generate a configuration file for a new player",
		Run: func(cmd *cobra.Command, args []string) {
			playersNewCmd(name, lhost)
		},
	}
	cmdPlayersNew.Flags().StringVarP(&name, "username", "u", "", "username of the new player")
	cmdPlayersNew.Flags().StringVarP(&lhost, "lhost", "l", "",
		"the hostname the player authenticates to this server using")
	cmdPlayers.AddCommand(cmdPlayersNew)

	root.AddCommand(cmdCoop, cmdPlayers)
	return root
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
	cmdClear := &cobra.Command{
		Use:   "clear",
		Short: "clear the terminal screen",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print("\033[H\033[2J")
		},
	}
	root.AddCommand(cmdSessions, cmdUse, cmdHttp, cmdHttps, cmdAgents, cmdBuilders, cmdBuild, cmdInstall, cmdUninstall,
		cmdStage, cmdUnstage, cmdVersion, cmdClear, cmdExit)
	root.CompletionOptions.HiddenDefaultCmd = true
	return root
}

// exits any named menus spawned by any commands
func exit(short string, menuType string) *cobra.Command {
	if len(short) == 0 {
		short = "exit the interactive menu"
	}
	mt := menuType
	cmd := &cobra.Command{
		Use:   "exit",
		Short: short,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			switch mt {
			case "build":
				if _, err := console.Rpc.EndBuild(ctx, &builderpb.BuildRequest{
					BuilderId: builderConfig.ID + builderConfig.builderID,
				}); err != nil {
					cLogger.Error("failed to delete builder client for %s: %v", builderConfig.builderID, err)
					return
				}
			}
			console.MainMenu()
		},
	}
	return cmd
}

func info(systemInfo *clientpb.Registration) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "view information about the agent's host",
		Long: "Information is typically compiled and sent by an agent when it first connects to the teamserver. " +
			"This information includes details such as the user running the process, the process ID, UID, GID, " +
			"IP address, and more; however if an agent doesn't transmit this information, you'd have to find out " +
			"yourself.",
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println()
			fmt.Println("System Information")
			fmt.Println("==================")
			_, _ = fmt.Fprintln(w, fmt.Sprintf("%v\t%v\t", "Agent ID:", systemInfo.AgentId))
			_, _ = fmt.Fprintln(w, fmt.Sprintf("%v\t%v\t", "Host OS:", systemInfo.Os))
			_, _ = fmt.Fprintln(w, fmt.Sprintf("%v\t%v\t", "Architecture:", systemInfo.Arch))
			_, _ = fmt.Fprintln(w, fmt.Sprintf("%v\t%v\t", "Username:", systemInfo.Username))
			_, _ = fmt.Fprintln(w, fmt.Sprintf("%v\t%v\t", "Hostname:", systemInfo.Hostname))
			_, _ = fmt.Fprintln(w, fmt.Sprintf("%v\t%v\t", "UID:", systemInfo.UID))
			_, _ = fmt.Fprintln(w, fmt.Sprintf("%v\t%v\t", "GID:", systemInfo.GID))
			_, _ = fmt.Fprintln(w, fmt.Sprintf("%v\t%v\t", "PID:", systemInfo.PID))
			_, _ = fmt.Fprintln(w, fmt.Sprintf("%v\t%v\t", "Home directory:", systemInfo.HomeDir))
			_, _ = fmt.Fprintln(w, fmt.Sprintf("%v\t%v\t", "Remote address:", systemInfo.IPAddress))
			_ = w.Flush()
			fmt.Println()
		},
	}
	return cmd
}
