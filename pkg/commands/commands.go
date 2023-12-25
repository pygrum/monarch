package commands

import (
	"context"
	"fmt"
	"github.com/desertbit/grumble"
	"github.com/pygrum/monarch/pkg/completion"
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/consts"
	"github.com/pygrum/monarch/pkg/crypto"
	"github.com/pygrum/monarch/pkg/teamserver/roles"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"strings"

	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
)

var (
	ctx        = context.Background()
	cLogger    log.Logger
	cmdPlayers *grumble.Command
)

func init() {
	cLogger, _ = log.NewLogger(log.ConsoleLogger, "")
}

func requiredFlag(flag string) {
	_, _ = fmt.Fprintln(console.App.Stderr(), fmt.Sprintf("'%s' is a required flag", flag))
}

func InitCTX() {
	m := make(map[string]string)
	m["uid"] = config.ClientConfig.UUID
	challenge, err := crypto.EncryptAES(config.ClientConfig.Secret, config.ClientConfig.Challenge)
	if err != nil {
		logrus.Fatalf("couldn't encrypt challenge for auth: %v", err)
	}
	m["challenge"] = challenge
	md := metadata.New(m)
	ctx = metadata.NewOutgoingContext(ctx, md)
}

func ConsoleInitCTX() {
	m := make(map[string]string)
	m["uid"] = config.ClientConfig.UUID
	m["username"] = consts.ServerUser
	m["role"] = string(roles.RoleAdmin)
	md := metadata.New(m)
	ctx = metadata.NewOutgoingContext(ctx, md)
}

func ServerConsoleCommands() []*grumble.Command {
	root := ConsoleCommands()
	cmdCoop := &grumble.Command{
		Name: "co-op",
		Help: "start / stop co-op mode",
		Flags: func(f *grumble.Flags) {
			f.Bool("s", "stop", false, "turn off co-op mode")
		},
		Run: func(c *grumble.Context) error {
			coopCmd(c.Flags.Bool("stop"))
			return nil
		},
		HelpGroup: consts.CoopHelpGroup,
	}

	cmdPlayers.AddCommand(&grumble.Command{
		Name: "new",
		Help: "generate a configuration file for a new player",
		Flags: func(f *grumble.Flags) {
			f.String("u", "username", "", "username of the new player")
			f.String("l", "lhost", "", "the hostname the player connects to")
			f.String("r", "role", "player", "the player role (see autocomplete options)")
		},
		Run: func(c *grumble.Context) error {
			uname, lhost, role := c.Flags.String("username"), c.Flags.String("lhost"), c.Flags.String("role")
			if len(lhost) == 0 {
				requiredFlag("lhost")
				return nil
			}
			if len(uname) == 0 {
				requiredFlag("uname")
				return nil
			}
			playersNewCmd(
				uname,
				lhost,
				role,
			)
			return nil
		},
		Completer: func(prefix string, args []string) []string {
			return completion.Players(prefix, ctx)
		},
	})

	cmdPlayers.AddCommand(&grumble.Command{
		Name: "kick",
		Help: "kick a player from the server",
		Args: func(a *grumble.Args) {
			a.String("name", "the name of the player to kick")
		},
		Run: func(c *grumble.Context) error {
			playersKickCmd(c.Args.String("name"))
			return nil
		},
		Completer: func(prefix string, args []string) []string {
			return completion.Players(prefix, ctx)
		},
	})

	root = append(root, cmdCoop)
	return root
}

// ConsoleCommands returns all commands used by the console
func ConsoleCommands() []*grumble.Command {
	var root []*grumble.Command

	cmdBuild := &grumble.Command{
		Name:      "build",
		Help:      "start the interactive session with an installed builder",
		HelpGroup: consts.GeneralHelpGroup,
		Args: func(a *grumble.Args) {
			a.String("agent-type", "the type of agent to build")
		},
		Run: func(c *grumble.Context) error {
			buildCmd(c.Args.String("agent-type"))
			return nil
		},
		Completer: func(prefix string, args []string) []string {
			return completion.Builders(prefix, ctx)
		},
	}

	cmdBuilders := &grumble.Command{
		Name:      "builders",
		Help:      "list installed builders",
		HelpGroup: consts.GeneralHelpGroup,
		Args: func(a *grumble.Args) {
			a.StringList("names", "list of builder names")
		},
		Run: func(c *grumble.Context) error {
			buildersCmd(c.Args.StringList("names"))
			return nil
		},
		Completer: func(prefix string, args []string) []string {
			return completion.Builders(prefix, ctx)
		},
	}

	cmdAgents := &grumble.Command{
		Name:      "agents",
		Help:      "list compiled agents",
		HelpGroup: consts.GeneralHelpGroup,
		Args: func(a *grumble.Args) {
			a.StringList("agents", "list of compiled agents")
		},
		Run: func(c *grumble.Context) error {
			agentsCmd(c.Args.StringList("agents"))
			return nil
		},
		Completer: func(prefix string, args []string) []string {
			return completion.Agents(prefix, ctx)
		},
	}

	cmdAgentsRm := &grumble.Command{
		Name:      "rm",
		Help:      "remove compiled agents from listing",
		HelpGroup: consts.GeneralHelpGroup,
		Args: func(a *grumble.Args) {
			a.StringList("agents", "list of compiled agents to delete", grumble.Min(1))
		},
		Completer: func(prefix string, args []string) []string {
			return completion.Agents(prefix, ctx)
		},
		Run: func(c *grumble.Context) error {
			cmdRm(c.Args.StringList("agents"))
			return nil
		},
	}
	cmdAgents.AddCommand(cmdAgentsRm)

	cmdSessions := &grumble.Command{
		Name:      "sessions",
		Help:      "list established agent connections",
		HelpGroup: consts.GeneralHelpGroup,
		Completer: func(prefix string, args []string) []string {
			return completion.Sessions(prefix, ctx)
		},
		Args: func(a *grumble.Args) {
			a.StringList("ids", "list of session ids")
		},
		Run: func(c *grumble.Context) error {
			sessionsCmd(c.Args.StringList("ids"))
			return nil
		},
	}

	cmdUse := &grumble.Command{
		Name:      "use",
		Help:      "initiate an interactive session with an agent",
		HelpGroup: consts.GeneralHelpGroup,
		Args: func(a *grumble.Args) {
			a.Int("id", "the ID of the session to use")
		},
		Run: func(c *grumble.Context) error {
			useCmd(c.Args.Int("id"))
			return nil
		},
		Completer: func(prefix string, args []string) []string {
			return completion.Sessions(prefix, ctx)
		},
	}

	cmdHttp := &grumble.Command{
		Name:      "http",
		Help:      "start an http listener",
		HelpGroup: consts.AdminHelpGroup,
		Flags: func(f *grumble.Flags) {
			f.Bool("s", "stop", false, "stop the http listener")
		},
		Run: func(c *grumble.Context) error {
			httpCmd(c.Flags.Bool("stop"))
			return nil
		},
	}

	cmdHttps := &grumble.Command{
		Name:      "https",
		Help:      "start an https listener",
		HelpGroup: consts.AdminHelpGroup,
		Flags: func(f *grumble.Flags) {
			f.Bool("s", "stop", false, "stop the https listener")
		},
		Run: func(c *grumble.Context) error {
			httpsCmd(c.Flags.Bool("stop"))
			return nil
		},
	}

	cmdInstall := &grumble.Command{
		Name:      "install",
		Help:      "install a builder from a Git repository",
		HelpGroup: consts.AdminHelpGroup,
		Args: func(a *grumble.Args) {
			a.String("repo", "github repo to install")
		},
		Flags: func(f *grumble.Flags) {
			f.Bool("c", "use-creds", false, "use GitHub credentials for installation")
			f.String("b", "branch", "", "target a specific branch to install")
		},
		Run: func(c *grumble.Context) error {
			installCmd(c.Args.String("repo"), c.Flags.String("branch"), c.Flags.Bool("use-creds"))
			return nil
		},
	}

	cmdLocal := &grumble.Command{
		Name:      "local",
		Help:      "install a builder from a local folder",
		HelpGroup: consts.AdminHelpGroup,
		Args: func(a *grumble.Args) {
			a.String("project", "local project folder")
		},
		Completer: func(prefix string, args []string) []string {
			return completion.LocalPathCompleter(prefix)
		},
		Run: func(c *grumble.Context) error {
			localCmd(c.Args.String("project"))
			return nil
		},
	}
	// it's a subcommand of the 'install' command
	cmdInstall.AddCommand(cmdLocal)

	cmdUninstall := &grumble.Command{
		Name:      "uninstall",
		Help:      "uninstall builder(s) by name or ID",
		HelpGroup: consts.AdminHelpGroup,
		Args: func(a *grumble.Args) {
			a.StringList("builders", "list of builders to uninstall")
		},
		Flags: func(f *grumble.Flags) {
			f.Bool("p", "purge", false, "remove local folder")
		},
		Run: func(c *grumble.Context) error {
			uninstallCmd(c.Args.StringList("builders"), c.Flags.Bool("purge"))
			return nil
		},
		Completer: func(prefix string, args []string) []string {
			return completion.Builders(prefix, ctx)
		},
	}

	cmdVersion := &grumble.Command{
		Name:      "version",
		Help:      "view installed monarch version",
		HelpGroup: consts.GeneralHelpGroup,
		Run: func(c *grumble.Context) error {
			versionCmd()
			return nil
		},
	}

	cmdStage := &grumble.Command{
		Name:      "stage",
		Help:      "stage an agent on the configured staging endpoint, or view currently staged agents",
		HelpGroup: consts.GeneralHelpGroup,
		Args: func(a *grumble.Args) {
			a.String("agent", "the name of the agent to stage", grumble.Default(""))
		},
		Flags: func(f *grumble.Flags) {
			f.StringL("as", "", "the file to stage your agent as (e.g. index.php)")
		},
		Run: func(c *grumble.Context) error {
			stageCmd(c.Args.String("agent"), c.Flags.String("as"))
			return nil
		},
		Completer: func(prefix string, args []string) []string {
			return completion.Agents(prefix, ctx)
		},
	}

	cmdUnstage := &grumble.Command{
		Name:      "unstage",
		Help:      "unstage a staged agent, by specifying its stage alias (e.g. index.php)",
		HelpGroup: consts.GeneralHelpGroup,
		Args: func(a *grumble.Args) {
			a.String("alias", "the alias of a staged agent")
		},
		Run: func(c *grumble.Context) error {
			unstageCmd(c.Args.String("alias"))
			return nil
		},
		Completer: func(prefix string, args []string) []string {
			return completion.UnStage(prefix, ctx)
		},
	}

	cmdPlayers = &grumble.Command{
		Name: "players",
		Help: "list registered players",
		Args: func(a *grumble.Args) {
			a.StringList("usernames", "player usernames")
		},
		Run: func(c *grumble.Context) error {
			playersCmd(c.Args.StringList("usernames"))
			return nil
		},
		Completer: func(prefix string, args []string) []string {
			return completion.Players(prefix, ctx)
		},
		HelpGroup: consts.CoopHelpGroup,
	}

	cmdSend := &grumble.Command{
		Name: "send",
		Help: "send a message to another online player",
		Flags: func(f *grumble.Flags) {
			f.String("t", "to", "", "player to message")
			f.Bool("a", "all", false, "message all players")
		},
		Args: func(a *grumble.Args) {
			a.StringList("message", "message to send the player")
		},
		Run: func(c *grumble.Context) error {
			sendCmd(c.Flags.String("to"), strings.Join(c.Args.StringList("message"), " "), c.Flags.Bool("all"))
			return nil
		},
		HelpGroup: consts.CoopHelpGroup,
	}

	root = append(root, cmdSessions, cmdUse, cmdHttp, cmdHttps, cmdAgents, cmdBuilders, cmdBuild, cmdInstall, cmdUninstall,
		cmdStage, cmdUnstage, cmdVersion, cmdPlayers, cmdSend)
	return root
}

func info(systemInfo *clientpb.Registration) *grumble.Command {
	cmd := &grumble.Command{
		Name:      "info",
		Help:      "view information about the agent's host",
		HelpGroup: consts.GeneralHelpGroup,
		LongHelp: "Information is typically compiled and sent by an agent when it first connects to the teamserver. " +
			"This information includes details such as the user running the process, the process ID, UID, GID, " +
			"IP address, and more; however if an agent doesn't transmit this information, you'd have to find out " +
			"yourself.",
		Run: func(*grumble.Context) error {
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
			return w.Flush()
		},
	}
	return cmd
}
