package console

import (
	"github.com/pygrum/monarch/pkg/commands"
	"github.com/pygrum/monarch/pkg/commands/build"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/reeflective/console"
)

type monarchServer struct {
	App *console.Console
}

var server *monarchServer

func init() {
	server = &monarchServer{
		App: console.New("monarch"),
	}
	db.Initialize()

}

// BuildMode switches the console to build mode. The name parameter is the name of the agent to be built.
// Only creating a new menu each time because I want it to be named
func BuildMode(name string) {
	buildMenu := server.App.NewMenu(name)
	buildMenu.AddCommand(build.ConsoleCommands()...)
	server.App.SwitchMenu(name)
}

func Run() error {
	// TODO:Start monarch console
	srvMenu := server.App.ActiveMenu()
	srvMenu.AddCommand(commands.ConsoleCommands()...)
	srvMenu.CompletionOptions.HiddenDefaultCmd = true

	return server.App.Start()
}
