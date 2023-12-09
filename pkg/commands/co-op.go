package commands

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/pygrum/monarch/pkg/protobuf/rpcpb"
	"github.com/pygrum/monarch/pkg/types"
	"os"
	"time"

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

func playersCmd(names []string) {
	var players []db.Player
	if len(names) > 0 {
		if err := db.Where("username IN ?", names).Find(&players).Error; err != nil {
			cLogger.Error("query failed: %v", err)
			return
		}
	} else {
		if err := db.Find(&players); err != nil {
			cLogger.Error("query failed: %v", err)
		}
	}
	header := "USERNAME\tACCOUNT CREATION DATE\t"
	_, _ = fmt.Fprintln(w, header)
	for _, player := range players {
		if player.Username == "console" {
			continue
		}
		line := fmt.Sprintf("%s\t%s\t",
			player.Username,
			player.CreatedAt.Format(time.RFC850),
		)
		_, _ = fmt.Fprintln(w, line)
	}
	_ = w.Flush()
}

func playersNewCmd(name, lhost string) {
	if len(name) == 0 {
		cLogger.Error("you must specify the player name")
		return
	}
	if len(lhost) == 0 {
		cLogger.Error("you must specify the server host")
		return
	}
	// don't save certificate deliberately, we don't need to and could be an issue if
	// they get leaked
	certPEM, keyPEM, err := crypto.NewClientCertificate(name)
	if err != nil {
		cLogger.Error("failed to generate player certificates: %v", err)
		return
	}
	uid := uuid.New().String()
	caCertPEM, _, err := crypto.CaCertKeyPair()
	if err != nil {
		cLogger.Error("failed to get CA certificate: %v", err)
		return
	}
	secret := crypto.RandomBytes(32)
	challenge := hex.EncodeToString(crypto.RandomBytes(128))
	clientConfig := &config.MonarchClientConfig{
		UUID:      uid,
		Name:      name,
		RHost:     lhost,
		RPort:     config.MainConfig.MultiplayerPort,
		CertPEM:   certPEM,
		KeyPEM:    keyPEM,
		CaCertPEM: caCertPEM,
		Secret:    secret,
		Challenge: challenge,
	}
	b64Cert := base64.StdEncoding.EncodeToString(certPEM)
	player := &db.Player{
		UUID:      uid,
		Username:  name,
		ClientCA:  b64Cert,
		Challenge: challenge,
		Secret:    hex.EncodeToString(secret),
	}
	bytes, err := json.Marshal(clientConfig)
	if err != nil {
		cLogger.Error("couldn't marshal config: %v", err)
		return
	}
	if err := db.Create(player); err != nil {
		cLogger.Error("failed to create new player: %v", err)
		return
	}
	if err := os.WriteFile(name+"-monarch-client.config", bytes, 0600); err != nil {
		cLogger.Error("failed to create configuration file: %v", err)
		return
	}
	cLogger.Success("saved player config to ./" + name + "-monarch-client.config")
}

func playersKickCmd(name string) {
	player := &db.Player{}
	if err := db.FindOneConditional("username = ?", name, &player); err != nil {
		cLogger.Error("query failed: %v", err)
		return
	}
	if len(player.UUID) == 0 {
		cLogger.Error("player '%s' doesn't exist")
		return
	}
	queue, ok := types.NotifQueues[player.UUID]
	if ok {
		_ = queue.Enqueue(&rpcpb.Notification{LogLevel: rpcpb.LogLevel_LevelError, Msg: types.NotificationKickPlayer})
	}
	if err := db.DeleteOne(player); err != nil {
		cLogger.Error("failed to remove player: %v", err)
		return
	}
	cLogger.Info("kicked %s from the operation", player.Username)
}
