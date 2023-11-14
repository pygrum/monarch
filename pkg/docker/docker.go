package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/pygrum/monarch/pkg/consts"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/translator"
	"strconv"
	"strings"
)

var (
	l log.Logger
)

func init() {
	l, _ = log.NewLogger(log.ConsoleLogger, "")
}

// RPCAddresses returns the rpc addresses of builder-translator pairs given an agent ID.
func RPCAddresses(cli *client.Client, ctx context.Context, BuilderID string) (string, string, error) {
	var builder db.Builder
	var trltr db.Translator
	var builderAddress, translatorAddress string
	if err := db.FindOneConditional("agent_id = ?", BuilderID, &builder); err != nil {
		return "", "", err
	}
	if len(builder.BuilderID) == 0 {
		return "", "", fmt.Errorf("no builder with ID %s exists", BuilderID)
	}
	// Use builder's linked translator as ID
	if err := db.FindOneConditional("translator_id = ?", builder.TranslatorID, &trltr); err != nil {
		return "", "", err
	}
	if len(builder.BuilderID) == 0 {
		return "", "", fmt.Errorf("no builder with ID %s exists", BuilderID)
	}
	cJson, err := cli.ContainerInspect(ctx, builder.ContainerID)
	if err != nil {
		return "", "", fmt.Errorf("failed to inspect builder %s: %v", BuilderID, err)
	}
	builderAddress = cJson.NetworkSettings.IPAddress + ":" + strconv.Itoa(translator.ListenPort)
	cJson, err = cli.ContainerInspect(ctx, trltr.ContainerID)
	if err != nil {
		return "", "", fmt.Errorf("failed to inspect builder %s: %v", BuilderID, err)
	}
	translatorAddress = cJson.NetworkSettings.IPAddress + ":" + strconv.Itoa(translator.ListenPort)
	return builderAddress, translatorAddress, nil
}

// StartContainers starts the pair of containers (or, just the builder) that are created when a new agent is installed
func StartContainers(cli *client.Client, ctx context.Context, builderImageTag,
	translatorImageTag string) (string, string, error) {
	// Run container with same name as image
	var builderID, translatorID string
	bContainerName := strings.Split(builderImageTag, ":")[0]
	tContainerName := strings.Split(translatorImageTag, ":")[0]
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: builderImageTag,
		Tty:   false,
	}, &container.HostConfig{RestartPolicy: container.RestartPolicy{Name: "unless-stopped"}},
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{consts.MonarchNet: {
				NetworkID: consts.MonarchNet,
			}},
		}, nil, bContainerName)
	if err != nil {
		return "", "", err
	}
	// Start builder container
	if err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", "", err
	}
	l.Info("started builder container %s", bContainerName)
	builderID = resp.ID

	// Only start translator image if it exists
	if len(translatorImageTag) != 0 {
		resp, err := cli.ContainerCreate(ctx, &container.Config{
			Image: translatorImageTag,
			Tty:   false,
		}, &container.HostConfig{RestartPolicy: container.RestartPolicy{Name: "unless-stopped"}},
			&network.NetworkingConfig{
				EndpointsConfig: map[string]*network.EndpointSettings{consts.MonarchNet: {
					NetworkID: consts.MonarchNet,
				}},
			}, nil, tContainerName)
		if err != nil {
			return "", "", err
		}
		// Start builder container
		if err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			return "", "", err
		}
		l.Info("started translator container %s", tContainerName)
		translatorID = resp.ID
	}
	return builderID, translatorID, nil
}
