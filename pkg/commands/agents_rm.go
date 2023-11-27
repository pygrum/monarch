package commands

import "github.com/pygrum/monarch/pkg/db"

func cmdRm(names []string) {
	var agents []db.Agent
	if err := db.FindConditional("agent_id IN ?", names, &agents); err != nil {
		if err = db.FindConditional("name IN ?", names, &agents); err != nil {
			cLogger.Error("failed to find agent(s): %v", err)
			return
		}
	}
	for _, agent := range agents {
		if err := db.DeleteOne(&agent); err != nil {
			cLogger.Error("failed to delete %s: %v", agent.Name, err)
		}
		cLogger.Success("deleted %v", agent.Name+":"+agent.AgentID)
	}
}
