package console

import (
	"github.com/pygrum/monarch/pkg/commands"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/reeflective/console"
)

type monarchServer struct {
	App *console.Console
}

var Server *monarchServer

func init() {
	Server = &monarchServer{
		App: console.New("monarch"),
	}
	db.Initialize()
}

func Run() error {
	// TODO:Start monarch console
	srvMenu := Server.App.ActiveMenu()
	srvMenu.AddCommand(commands.ConsoleCommands()...)

	srvMenu.CompletionOptions.HiddenDefaultCmd = true
	return Server.App.Start()
}
