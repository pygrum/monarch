package agent

import (
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/console"
	"github.com/spf13/cobra"
)

type Agent struct {
	// agent specific data
	Name     string
	Commands []config.ProjectConfigCmd
}

func (a *Agent) RegisterAgentMenu() {
	agentMenu := console.Server.App.NewMenu(a.Name)
	for _, cmd := range a.Commands {
		// Runs a function to translate the command data via RPC and process it for forwarding to destined host
		c := &cobra.Command{
			Use:   cmd.Usage,
			Args:  cobra.ExactArgs(cmd.NArgs),
			Short: cmd.DescriptionShort,
			Long:  cmd.DescriptionLong,
		}
		agentMenu.AddCommand(c)
	}
}
