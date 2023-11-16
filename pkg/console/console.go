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

// BuildMode switches the console to build mode. The name parameter is the name of the agent to be built.
// Only creating a new menu each time because I want it to be named
func BuildMode(name string, commands []*cobra.Command) {
	buildMenu := server.App.NewMenu(name)
	buildMenu.AddCommand(commands...)
	server.App.SwitchMenu(name)
}

func Run(commands []*cobra.Command) error {
	srvMenu := server.App.ActiveMenu()
	srvMenu.AddCommand(commands...)
	srvMenu.CompletionOptions.HiddenDefaultCmd = true
	return server.App.Start()
}
