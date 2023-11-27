package commands

import (
	"fmt"
	"github.com/pygrum/monarch/pkg/db"
	"time"
)

// agentsCmd lists compiled agents
func agentsCmd(names []string) {
	var agents []db.Agent
	if len(names) > 0 {
		if err := db.FindConditional("agent_id IN ?", names, &agents); err != nil {
			cLogger.Error("failed to retrieve the specified agents: %v", err)
			return
		}
		if len(agents) == 0 {
			if err := db.FindConditional("name IN ?", names, &agents); err != nil {
				cLogger.Error("failed to retrieve the specified agents: %v", err)
				return
			}
		}
	} else {
		if err := db.Find(&agents); err != nil {
			cLogger.Error("failed to find agent(s): %v", err)
			return
		}
	}
	header := "ID\tNAME\tVERSION\tPLATFORM\tBUILDER\tFILE\tCREATED AT\t"
	_, _ = fmt.Fprintln(w, header)
	for _, agent := range agents {
		line := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t",
			agent.AgentID,
			agent.Name,
			agent.Version,
			agent.OS+"/"+agent.Arch,
			agent.Builder,
			agent.File,
			agent.CreatedAt.Format(time.DateTime),
		)
		_, _ = fmt.Fprintln(w, line)
	}
	_ = w.Flush()
}
