package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
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
	rootCmd = &cobra.Command{
		Use: "monarch-client",
		Run: func(cmd *cobra.Command, args []string) {
			home, _ := os.UserHomeDir()
			// default config
			configPath := filepath.Join(home, ".monarch", "monarch-client.config")

			if err := config.JsonConfig(configPath, &config.ClientConfig); err != nil {
				fmt.Println("couldn't load configuration file:", err)
				os.Exit(1)
			}
			commands.InitCTX()
			console.Run(false, commands.ConsoleCommands, commands.BuildCommands)
		},
	}

	importCmd := &cobra.Command{
		Use:   "import </path/to/config>",
		Short: "import a configuration file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			data, err := os.ReadFile(args[0])
			if err != nil {
				logrus.Fatalf("could not read supplied configuration file: %v", err)
			}
			home, _ := os.UserHomeDir()
			configPath := filepath.Join(home, ".monarch", "monarch-client.config")
			if err = os.WriteFile(configPath, data, 0600); err != nil {
				logrus.Fatalf("import failed: %v", err)
			}
			fmt.Println("successfully imported", args[0])
		},
	}

	rootCmd.AddCommand(importCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
