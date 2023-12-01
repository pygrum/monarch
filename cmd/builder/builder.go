package main

import (
	"github.com/pygrum/monarch/pkg/builder"
	"github.com/pygrum/monarch/pkg/log"
)

var l log.Logger

func init() {
	l, _ = log.NewLogger(log.ConsoleLogger, "")
}
func main() {
	if err := builder.Run(); err != nil {
		l.Fatal("%v", err)
	}
}
