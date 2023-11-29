package main

import (
	"github.com/pygrum/monarch/pkg/lint"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "royal-lint [path/to/royal.yaml]",
		Short: "Lint Monarch project configuration files",
		Long: "royal-lint is a tool for developers to verify that their project configuration file 'royal.yaml' " +
			"is valid. This includes checking general fields (e.g. version for semantic versioning) as well as " +
			"special fields that only allow certain variables (such as the 'type' field when specifying " +
			"build arguments).",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			lint.Lint(args[0])
		},
	}
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
