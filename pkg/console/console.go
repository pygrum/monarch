package console

import (
	"fmt"
	"github.com/pygrum/monarch/pkg/consts"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/reeflective/console"
	"github.com/spf13/cobra"
)

type server struct {
	App *console.Console
}

var monarchServer *server

func init() {
	monarchServer = &server{
		App: console.New("monarch"),
	}
	db.Initialize()
	log.Initialize(monarchServer.App.TransientPrintf)
}

// NamedMenu switches the console to a new menu with the provided name.
func NamedMenu(name string, commands func() *cobra.Command) {
	namedMenu := monarchServer.App.NewMenu(name)
	namedMenu.SetCommands(commands)
	monarchServer.App.SwitchMenu(name)
}

// Run entrypoint for the entire application
func Run(rootCmd func() *cobra.Command) error {
	srvMenu := monarchServer.App.ActiveMenu()
	srvMenu.SetCommands(rootCmd)
	monarchServer.App.SetPrintLogo(func(_ *console.Console) {
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
	return monarchServer.App.Start()
}

// MainMenu switches back to the main menu
func MainMenu() {
	monarchServer.App.SwitchMenu("")
}
