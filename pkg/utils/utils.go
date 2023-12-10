package utils

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/docker"
	"github.com/pygrum/monarch/pkg/protobuf/rpcpb"
)

// Cleanup is used to delete all data associated with a builder
func Cleanup(builder *db.Builder, stream rpcpb.Monarch_UninstallServer) error {
	ctx := context.Background()
	if err := docker.Cli.ContainerStop(ctx, builder.ContainerID, container.StopOptions{}); err != nil {
		_ = stream.Send(&rpcpb.Notification{
			LogLevel: rpcpb.LogLevel_LevelError,
			Msg:      fmt.Sprintf("failed to stop container for %s: %v", builder.Name, err),
		})
	}
	if err := docker.Cli.ContainerRemove(ctx, builder.ContainerID, types.ContainerRemoveOptions{Force: true}); err != nil {
		return fmt.Errorf("failed to remove container for %s: %v", builder.Name, err)
	}
	m, err := docker.Cli.ImageRemove(ctx, builder.ImageID, types.ImageRemoveOptions{Force: true})
	if err != nil {
		return fmt.Errorf("failed to remove image for %s: %v", builder.Name, err)
	}
	for _, i := range m {
		_ = stream.Send(&rpcpb.Notification{
			LogLevel: rpcpb.LogLevel_LevelInfo,
			Msg:      fmt.Sprintf("%s: untagged", i.Untagged),
		})
		_ = stream.Send(&rpcpb.Notification{
			LogLevel: rpcpb.LogLevel_LevelInfo,
			Msg:      fmt.Sprintf("deleted %s", i.Deleted),
		})
	}
	return db.Delete(builder)
}
