package commands

import (
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
	"io"
)

func uninstallCmd(args []string, deleteSource bool) {
	stream, err := console.Rpc.Uninstall(ctx, &clientpb.UninstallRequest{
		Builders:     &clientpb.BuilderRequest{BuilderId: args},
		RemoveSource: deleteSource,
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
