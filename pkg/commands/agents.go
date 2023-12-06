package commands

import (
	"fmt"
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
)

// agentsCmd lists compiled agents
func agentsCmd(names []string) {
	agents, err := console.Rpc.Agents(ctx, &clientpb.AgentRequest{AgentId: names})
	if err != nil {
		cLogger.Error("failed to get agents: %v", err)
		return
	}
	header := "ID\tNAME\tVERSION\tPLATFORM\tBUILDER\tFILE\tCREATED AT\t"
	_, _ = fmt.Fprintln(w, header)
	for _, agent := range agents.Agents {
		line := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t",
			agent.AgentId,
			agent.Name,
			agent.Version,
			agent.OS+"/"+agent.Arch,
			agent.Builder,
			agent.File,
			agent.CreatedAt,
		)
		_, _ = fmt.Fprintln(w, line)
	}
	_ = w.Flush()
}
