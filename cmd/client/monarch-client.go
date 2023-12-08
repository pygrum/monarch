package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/pygrum/monarch/pkg/commands"
	"github.com/pygrum/monarch/pkg/config"
	"github.com/spf13/cobra"

	"github.com/pygrum/monarch/pkg/console"
)

var rootCmd *cobra.Command

func init() {
	var configFile string
	rootCmd = &cobra.Command{
		Use: os.Args[0],
		Run: func(cmd *cobra.Command, args []string) {
			if len(configFile) == 0 {
				home, _ := os.UserHomeDir()
				// default config
				configFile = filepath.Join(home, ".monarch", "monarch-client.json")
			}
			if err := config.JsonConfig(configFile, &config.ClientConfig); err != nil {
				log.Fatalf("couldn't load client config (%s): %v", configFile, err)
			}
			if err := console.Run(commands.ConsoleCommands, false); err != nil {
				log.Fatal(err)
			}
		},
	}
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "monarch client configuration file")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
