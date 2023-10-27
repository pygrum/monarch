package console

import (
	"github.com/pygrum/monarch/pkg/commands"
	"github.com/reeflective/console"
)

type monarchServer struct {
	app *console.Console
}

func Run() error {
	// TODO:Start server console
	s := &monarchServer{
		app: console.New("monarch"),
	}
	srvMenu := s.app.ActiveMenu()
	srvMenu.AddCommand(commands.ConsoleCommands()...)

	srvMenu.CompletionOptions.HiddenDefaultCmd = true
	return s.app.Start()
}
