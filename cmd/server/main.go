package main

import (
	"github.com/pygrum/monarch/pkg/server"
	"log"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
