package main

import (
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/db"
	"log"

	"github.com/pygrum/monarch/pkg/commands"

	"github.com/pygrum/monarch/pkg/console"
)

func init() {

}

func main() {
	config.Initialize()
	config.ClientConfig.Name = "console"
	db.Initialize()

	if err := console.Run(commands.ServerConsoleCommands, true); err != nil {
		log.Fatal(err)
	}
}
