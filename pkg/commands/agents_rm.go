package commands

import (
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
)

func cmdRm(names []string) {
	if _, err := console.Rpc.RmAgents(CTX, &clientpb.AgentRequest{AgentId: names}); err != nil {
		cLogger.Error("failed to remove all agents: %v", err)
		return
	}
}
