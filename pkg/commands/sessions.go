package commands

import (
	"fmt"
	"github.com/pygrum/monarch/pkg/handlers/xhttp"
	"strconv"
	"time"
)

func sessionsCmd(sessionIDs []string) {
	var sessIDs = make([]int, len(sessionIDs))
	for i, id := range sessionIDs {
		intID, err := strconv.Atoi(id)
		if err != nil {
			cLogger.Error("'%v' is not a valid session ID", id)
			return
		}
		sessIDs[i] = intID
	}
	sessions := xhttp.Handler.Sessions(sessIDs)
	header := "ID\tAGENT ID\tAGENT NAME\tQUEUE SIZE\tLAST ACTIVE\t"
	_, _ = fmt.Fprintln(w, header)
	for _, session := range sessions {
		line := fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t",
			session.ID,
			session.Agent.AgentID,
			session.Agent.Name,
			session.RequestQueue.Size(),
			session.LastActive.Format(time.DateTime),
		)
		_, _ = fmt.Fprintln(w, line)
	}
	_ = w.Flush()
}
