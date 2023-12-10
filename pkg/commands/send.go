package commands

import (
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/protobuf/rpcpb"
)

func sendCmd(to, msg string, all bool) {
	if len(to) == 0 {
		if !all {
			cLogger.Error("player not specified with -to, please specify a player name or --all")
			return
		}
	}
	if _, err := console.Rpc.SendMessage(ctx, &rpcpb.Message{To: to, Msg: msg}); err != nil {
		cLogger.Error("%v", err)
		return
	}
	cLogger.Info("sent")
}
