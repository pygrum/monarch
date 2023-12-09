package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/pygrum/monarch/pkg/builder"
	"github.com/pygrum/monarch/pkg/consts"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/log"
	"strconv"
	"strings"
)

var (
	l   log.Logger
	Cli *client.Client // global docker client
)

func init() {
	var err error
	l, _ = log.NewLogger(log.ConsoleLogger, "")
	Cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		l.Fatal("failed to initialize internal docker client: %v", err)
	}
}

// RPCAddress returns the rpc endpoint of the builder container given an agent ID.
func RPCAddress(cli *client.Client, ctx context.Context, BuilderID string) (string, error) {
	var bldr db.Builder
	var builderAddress string
	if err := db.FindOneConditional("builder_id = ?", BuilderID, &bldr); err != nil {
		return "", err
	}
	if len(bldr.BuilderID) == 0 {
		return "", fmt.Errorf("no bldr with ID %s exists", BuilderID)
	}
	if len(bldr.BuilderID) == 0 {
		return "", fmt.Errorf("no bldr with ID %s exists", BuilderID)
	}
	cJson, err := cli.ContainerInspect(ctx, bldr.ContainerID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect bldr %s: %v", BuilderID, err)
	}
	netSettings, ok := cJson.NetworkSettings.Networks[consts.MonarchNet]
	if !ok {
		return "", fmt.Errorf("could not find %s", consts.MonarchNet)
	}
	builderAddress = netSettings.IPAddress + ":" + strconv.Itoa(builder.ListenPort)
	return builderAddress, nil
}

// StartContainer starts the builder container that is created when a new agent is installed
func StartContainer(cli *client.Client, ctx context.Context, builderImageTag string) (string, error) {
	// Run container with same name as image
	var builderID string
	bContainerName := strings.Split(builderImageTag, ":")[0]
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
		return "", err
	}
	// Start builder container
	if err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}
	builderID = resp.ID

	return builderID, nil
}
