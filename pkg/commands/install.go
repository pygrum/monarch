package commands

import (
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
	"io"
)

func installCmd(repoUrl, branch string, useCreds bool) {
	stream, err := console.Rpc.Install(ctx, &clientpb.InstallRequest{
		Path:     repoUrl,
		Source:   clientpb.InstallRequest_Git,
		Branch:   branch,
		UseCreds: useCreds,
	})
	if err != nil {
		cLogger.Error("%v", err)
		return
	}
	for {
		notif, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			cLogger.Error("install failed: %v", err)
			return
		}
		log.NumericalLevel(cLogger, uint16(notif.LogLevel), notif.Msg)
	}
}

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
			cLogger.Error("install failed: %v", err)
			return
		}
		log.NumericalLevel(cLogger, uint16(notif.LogLevel), notif.Msg)
	}
}
