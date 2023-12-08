package main

import (
	"log"

	"github.com/pygrum/monarch/pkg/commands"

	"github.com/pygrum/monarch/pkg/console"
)

func init() {

}

func main() {
	if err := console.Run(commands.ConsoleCommands, true); err != nil {
		log.Fatal(err)
	}
}
