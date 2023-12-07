package commands

import (
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
)

// httpCmd starts an HTTP listener for incoming connections, whether it be from c2 profiles or agents directly
func httpCmd(stop bool) {
	if stop {
		if _, err := console.Rpc.HttpClose(ctx, &clientpb.Empty{}); err != nil {
			cLogger.Error("%v", err)
			return
		}
	}
	notif, _ := console.Rpc.HttpOpen(ctx, &clientpb.Empty{})
	log.NumericalLevel(cLogger, uint16(notif.LogLevel), notif.Msg)
}

// same as httpCmd but starts an HTTPS listener
func httpsCmd(stop bool) {
	if stop {
		if _, err := console.Rpc.HttpsClose(ctx, &clientpb.Empty{}); err != nil {
			cLogger.Error("%v", err)
			return
		}
	}
	notif, _ := console.Rpc.HttpsOpen(ctx, &clientpb.Empty{})
	log.NumericalLevel(cLogger, uint16(notif.LogLevel), notif.Msg)
}
