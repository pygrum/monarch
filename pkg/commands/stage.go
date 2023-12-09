package commands

import (
	"fmt"
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
	"strings"
)

func stageCmd(args []string, format, as string) {
	if len(args) == 0 {
		s, err := console.Rpc.StageView(ctx, &clientpb.Empty{})
		if err != nil {
			cLogger.Error("%v", err)
			return
		}
		stagefmt := "%s (%s)\tstaged as %s\t"
		for k, v := range s.Stage {
			_, _ = fmt.Fprintln(w, fmt.Sprintf(stagefmt, v.Path, v.Agent,
				strings.ReplaceAll(s.Endpoint, "{file}", k)))
		}
		w.Flush()
		return
	}
	name := args[0]
	notif, err := console.Rpc.StageAdd(ctx, &clientpb.StageAddRequest{Agent: name, Alias: as})
	if err != nil {
		cLogger.Error("%v", err)
		return
	}
	log.NumericalLevel(cLogger, uint16(notif.LogLevel), notif.Msg)
}

func unstageCmd(name string) {
	if _, err := console.Rpc.Unstage(ctx, &clientpb.UnstageRequest{Alias: name}); err != nil {
		cLogger.Error("%v", err)
		return
	}
	cLogger.Info("unstaged %s", name)
}
