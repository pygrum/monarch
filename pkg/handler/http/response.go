package http

import (
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/protobuf/builderpb"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
	"github.com/pygrum/monarch/pkg/transport"
	"os"
	"path/filepath"
	"time"
)

func HandleResponse(session *clientpb.Session, resp *transport.GenericHTTPResponse) {
	session.LastActive = time.Now().Format(time.RFC850)
	for _, response := range resp.Responses {
		handleResponse(response, ShortID(resp.RequestID))
	}
}

func handleResponse(response transport.ResponseDetail, rid string) {
	if response.Status == builderpb.Status_FailedWithMessage {
		if len(response.Data) == 0 {
			TranLogger.Error("request %s failed but no message was returned", rid)
			return
		}
		TranLogger.Error("%s failed: %s", rid, string(response.Data))
		return
	}
	if response.Dest == transport.DestStdout {
		log.Print(string(response.Data))
	} else if response.Dest == transport.DestFile {
		wd, err := os.Getwd()
		if err != nil {
			wd = os.TempDir()
		}
		file := filepath.Join(wd, response.Name)
		if err := os.WriteFile(file, response.Data, 0666); err != nil {
			TranLogger.Error("failed writing response to %s to file: %v", rid, err)
			return
		}
		TranLogger.Info("%s: file saved to %s", rid, file)
	}
}
