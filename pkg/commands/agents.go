package commands

import (
	"fmt"
	"github.com/pygrum/monarch/pkg/db"
	"time"
)

// agentsCmd lists compiled agents
func agentsCmd(names []string) {
	var agents []db.Agent
	if err := db.FindConditional("agent_id IN ?", names, &agents); err != nil {
		cLogger.Error("failed to find agent(s): %v", err)
		return
	}
	header := "ID\tVERSION\tPLATFORM\tBUILDER\tFILE\tCREATED AT\t"
	_, _ = fmt.Fprintln(w, header)
	for _, agent := range agents {
		line := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t",
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
