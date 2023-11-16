package main

import (
	"github.com/pygrum/monarch/pkg/commands"
	"log"

	"github.com/pygrum/monarch/pkg/console"
)

func init() {

}

func main() {
	if err := console.Run(commands.ConsoleCommands()); err != nil {
		log.Fatal(err)
	}
}
