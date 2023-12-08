package commands

import (
	"encoding/base64"
	"encoding/json"
	"os"

	"github.com/google/uuid"
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/crypto"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/teamserver"
)

func coopCmd(stop bool) {
	if stop {
		teamserver.Stop()
		cLogger.Info("turned off co-op mode")
		return
	}
	go func() {
		if err := teamserver.Start(); err != nil {
			cLogger.Error("couldn't start teamserver: %v", err)
		}
		return
	}()
}

func playerNewCmd(name, lhost string) {
	// don't save certificate deliberately, we don't need to and could be an issue if
	// they get leaked
	certPEM, keyPEM, err := crypto.NewClientCertificate(name)
	if err != nil {
		cLogger.Error("failed to generate player certificates: %v", err)
		return
	}
	uid := uuid.New().String()
	clientConfig := &config.MonarchClientConfig{
		UUID:    uid,
		Name:    name,
		RHost:   lhost,
		RPort:   config.MainConfig.MultiplayerPort,
		CertPEM: certPEM,
		KeyPEM:  keyPEM,
	}
	b64Cert := base64.StdEncoding.EncodeToString(certPEM)
	player := &db.Player{
		UUID:     uid,
		Username: name,
		ClientCA: b64Cert,
	}
	bytes, err := json.Marshal(clientConfig)
	if err != nil {
		cLogger.Error("couldn't marshal config: %v", err)
		return
	}
	if err := os.WriteFile(name+"-monarch-client.config", bytes, 0600); err != nil {
		cLogger.Error("failed to create configuration file: %v", err)
		return
	}
	if err := db.Create(player); err != nil {
		cLogger.Error("failed to create new player: %v", err)
		return
	}
}

func playerKickCmd(name string) {

}
