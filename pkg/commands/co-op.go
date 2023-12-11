package commands

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/console"
	"github.com/pygrum/monarch/pkg/consts"
	"github.com/pygrum/monarch/pkg/crypto"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
	"github.com/pygrum/monarch/pkg/protobuf/rpcpb"
	"github.com/pygrum/monarch/pkg/teamserver"
	"github.com/pygrum/monarch/pkg/teamserver/roles"
	"github.com/pygrum/monarch/pkg/types"
	"os"
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
	cLogger.Info("co-op mode activated (%s:%d)", config.MainConfig.Interface, config.MainConfig.MultiplayerPort)
}

func playersCmd(names []string) {
	players, err := console.Rpc.Players(ctx, &clientpb.PlayerRequest{Names: names})
	if err != nil {
		cLogger.Error("%v", err)
		return
	}
	header := "USERNAME\tROLE\tACCOUNT CREATION DATE\t"
	_, _ = fmt.Fprintln(w, header)
	for _, player := range players.Players {
		if player.Username == consts.ServerUser {
			continue
		}
		line := fmt.Sprintf("%s\t%s\t%s\t",
			player.Username,
			player.Role,
			player.Registered,
		)
		_, _ = fmt.Fprintln(w, line)
	}
	_ = w.Flush()
}

func playersNewCmd(name, lhost, role string) {
	if !roles.ValidRole(role) {
		cLogger.Error("'%s' is not a valid role", role)
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
		Role:      roles.Role(role),
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
	cLogger.Info("account '%s' created with role '%s'", name, role)
	cLogger.Success("saved player config to ./" + name + "-monarch-client.config")
}

func playersKickCmd(name string) {
	if name == consts.ServerUser {
		cLogger.Warn("you cannot kick yourself")
		return
	}
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
