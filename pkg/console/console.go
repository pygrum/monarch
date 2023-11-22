package console

import (
	"github.com/pygrum/monarch/pkg/db"
	"github.com/reeflective/console"
	"github.com/spf13/cobra"
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

// NamedMenu switches the console to a new menu with the provided name.
func NamedMenu(name string, commands []*cobra.Command) {
	namedMenu := server.App.NewMenu(name)
	namedMenu.AddCommand(commands...)
	server.App.SwitchMenu(name)
}

func Run(commands []*cobra.Command) error {
	srvMenu := server.App.ActiveMenu()
	srvMenu.AddCommand(commands...)
	srvMenu.CompletionOptions.HiddenDefaultCmd = true
	return server.App.Start()
}

// MainMenu switches back to the main menu
func MainMenu() {
	server.App.SwitchMenu("")
}
