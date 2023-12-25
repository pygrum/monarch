package commands

import (
	"fmt"
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
	"strings"
)

func stageCmd(arg string, as string) {
	if len(arg) == 0 {
		s, err := console.Rpc.StageView(ctx, &clientpb.Empty{})
		if err != nil {
			cLogger.Error("%v", err)
			return
		}
		stagefmt := "%s (%s)\tstaged as %s\t"
		if len(s.Stage) == 0 {
			cLogger.Info("nothing staged")
			return
		}
		for k, v := range s.Stage {
			_, _ = fmt.Fprintln(w, fmt.Sprintf(stagefmt, v.Path, v.Agent,
				strings.ReplaceAll(s.Endpoint, "{file}", k)))
		}
		w.Flush()
		return
	}
	notif, err := console.Rpc.StageAdd(ctx, &clientpb.StageAddRequest{Agent: arg, Alias: as})
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
