package commands

import (
	"fmt"
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
)

func agentInfoCmd(agent string, build bool) {
	agents, err := console.Rpc.Agents(ctx, &clientpb.AgentRequest{AgentId: []string{agent}})
	if err != nil {
		cLogger.Error("failed to get agents: %v", err)
		return
	}
	if len(agents.Agents) == 0 {
		cLogger.Info("no agent with that name or ID exists")
	}
	a := agents.Agents[0]
	if !build {
		header := "ID\tNAME\tVERSION\tPLATFORM\tBUILDER\tCREATED AT\t"
		_, _ = fmt.Fprintln(w, header)
		line := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t",
			a.AgentId,
			a.Name,
			a.Version,
			a.OS+"/"+a.Arch,
			a.Builder,
			a.CreatedAt,
		)
		_, _ = fmt.Fprintln(w, line)
		_ = w.Flush()
		return
	}
	fmt.Print(a.AgentInfo)
}
