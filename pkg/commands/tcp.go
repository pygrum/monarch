package commands

import (
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
)

// tcpCmd starts a raw (TLS-secured) TCP listener for incoming connections, whether it be from c2 profiles or agents directly
func tcpCmd(stop bool) {
	if stop {
		if _, err := console.Rpc.TcpClose(ctx, &clientpb.Empty{}); err != nil {
			cLogger.Error("%v", err)
		}
		return
	}
	notif, err := console.Rpc.TcpOpen(ctx, &clientpb.Empty{})
	if err != nil {
		cLogger.Error("%v", err)
		return
	}
	log.NumericalLevel(cLogger, uint16(notif.LogLevel), notif.Msg)
}
