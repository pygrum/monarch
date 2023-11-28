package commands

import "github.com/pygrum/monarch/pkg/db"

func cmdRm(names []string) {
	var agents []db.Agent
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
	if len(agents) == 0 {
		cLogger.Error("no agents with the provided names exist")
		return
	}
	for _, agent := range agents {
		if err := db.DeleteOne(&agent); err != nil {
			cLogger.Error("failed to delete %s: %v", agent.Name, err)
		}
		cLogger.Success("deleted %v", agent.Name+":"+agent.AgentID)
	}
}
