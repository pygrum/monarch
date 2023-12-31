package main

import (
	"github.com/pygrum/monarch/pkg/commands"
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/db"

	"github.com/pygrum/monarch/pkg/console"
)

func main() {
	config.Initialize()
	db.Initialize()

	commands.ConsoleInitCTX()
	console.Run(true, commands.ServerConsoleCommands, commands.BuildCommands)
}
