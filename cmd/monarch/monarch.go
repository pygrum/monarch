package main

import (
	"log"

	"github.com/pygrum/monarch/pkg/console"
)

func init() {

}

func main() {
	if err := console.Run(); err != nil {
		log.Fatal(err)
	}
}
