package commands

import (
	"fmt"
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
	"strconv"
)

func sessionsCmd(sessionIDs []string) {
	var sessIDs = make([]int32, len(sessionIDs))
	for i, id := range sessionIDs {
		intID, err := strconv.Atoi(id)
		if err != nil {
			cLogger.Error("'%v' is not a valid session ID", id)
			return
		}
		sessIDs[i] = int32(intID)
	}
	sessions, err := console.Rpc.Sessions(ctx, &clientpb.SessionsRequest{IDs: sessIDs})
	if err != nil {
		cLogger.Error("%v", err)
		return
	}
	header := "ID\tAGENT ID\tAGENT NAME\tQUEUE SIZE\tLAST ACTIVE\tSTATUS\t"
	_, _ = fmt.Fprintln(w, header)
	for _, session := range sessions.Sessions {
		line := fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v\t",
			session.Id,
			session.AgentId,
			session.AgentName,
			session.QueueSize,
			session.LastActive,
			session.Status,
		)
		_, _ = fmt.Fprintln(w, line)
	}
	_ = w.Flush()
}
