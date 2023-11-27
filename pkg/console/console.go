package console

import (
	"fmt"
	"github.com/pygrum/monarch/pkg/consts"
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
func NamedMenu(name string, commands func() *cobra.Command) {
	namedMenu := server.App.NewMenu(name)
	namedMenu.SetCommands(commands)
	server.App.SwitchMenu(name)
}

// Run entrypoint for the entire application
func Run(rootCmd func() *cobra.Command) error {
	srvMenu := server.App.ActiveMenu()
	srvMenu.SetCommands(rootCmd)
	server.App.SetPrintLogo(func(_ *console.Console) {
		fmt.Print("\033[H\033[2J")
		fmt.Printf(`                  o 
               o^/|\^o
            o_^|\/*\/|^_o
           o\*¬'.\|/.'¬*/o
            \\\\\\|//////
             {><><@><><}
             |"""""""""|
               MONARCH
  ADVERSARY EMULATION TOOLKIT v%s
  ==================================

		`, consts.Version)
	})
	return server.App.Start()
}

// MainMenu switches back to the main menu
func MainMenu() {
	server.App.SwitchMenu("")
}
