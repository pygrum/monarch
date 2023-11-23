package commands

import (
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/handlers/xhttp"
)

// httpCmd starts an HTTP listener for incoming connections, whether it be from c2 profiles or agents directly
func httpCmd(stop bool) {
	if stop {
		l.Info("stopping http listener")
		if err := xhttp.Handler.Stop(); err != nil {
			l.Error("%v", err)
		}
		return
	}
	if xhttp.Handler.IsActive() {
		cLogger.Warn("http listener is already active")
		return
	}
	cLogger.Info("starting http listener on %s:%d", config.MainConfig.Interface, config.MainConfig.HttpPort)
	go xhttp.Handler.Serve()
}

// same as httpCmd but starts an HTTPS listener
func httpsCmd(stop bool) {
	if stop {
		l.Info("stopping https listener")
		if err := xhttp.Handler.StopTLS(); err != nil {
			l.Error("%v", err)
		}
		return
	}
	if xhttp.Handler.IsActiveTLS() {
		cLogger.Warn("https listener is already active")
		return
	}
	cLogger.Info("starting https listener on %s:%d", config.MainConfig.Interface, config.MainConfig.HttpsPort)
	go xhttp.Handler.ServeTLS()
}
