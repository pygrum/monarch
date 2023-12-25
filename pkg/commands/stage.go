package commands

import (
	"fmt"
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
	"os"
	"path/filepath"
	"strings"
)

func stageCmd(agent string, as string) {
	if len(agent) == 0 {
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
	notif, err := console.Rpc.StageAdd(ctx, &clientpb.StageAddRequest{Agent: agent, Alias: as})
	if err != nil {
		cLogger.Error("%v", err)
		return
	}
	log.NumericalLevel(cLogger, uint16(notif.LogLevel), notif.Msg)
}

func stageLocalCmd(file string, as string) {
	data, err := os.ReadFile(file)
	if err != nil {
		cLogger.Error("%v", err)
		return
	}
	req := &clientpb.StageLocalRequest{
		Filename: filepath.Base(file),
		Data:     data,
		Alias:    as,
	}
	notif, err := console.Rpc.StageLocal(ctx, req)
	if err != nil {
		cLogger.Error("%v", err)
		return
	}
	log.NumericalLevel(cLogger, uint16(notif.LogLevel), notif.Msg)
}

func unstageCmd(name string) {
	if _, err := console.Rpc.Unstage(ctx, &clientpb.UnstageRequest{Alias: name}); err != nil {
		cLogger.Info("nothing is staged on %s", name)
		return
	}
	cLogger.Info("unstaged %s", name)
}
