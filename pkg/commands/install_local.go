package commands

import (
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
	"io"
)

// installs local repositories / folders
func localCmd(path string) {
	stream, err := console.Rpc.Install(ctx, &clientpb.InstallRequest{
		Path:   path,
		Source: clientpb.InstallRequest_Local,
	})
	if err != nil {
		cLogger.Error("%v", err)
		return
	}
	for {
		notif, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			cLogger.Error("failed to receive notification: %v", err)
		}
		log.NumericalLevel(cLogger, uint16(notif.LogLevel), notif.Msg)
	}
}
