package commands

import (
	"github.com/pygrum/monarch/pkg/log"
	"github.com/spf13/cobra"
)

var cLogger log.Logger

func init() {
	cLogger, _ = log.NewLogger(log.ConsoleLogger, "")
}

// ConsoleCommands returns all commands used by the console
func ConsoleCommands() []*cobra.Command {

	var root []*cobra.Command

	cmdShutdown := &cobra.Command{
		Use:   "shutdown",
		Short: "shutdown the monarch monarch",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			shutdownCmd()
		},
	}
	root = append(root, cmdShutdown)
	return root

}
