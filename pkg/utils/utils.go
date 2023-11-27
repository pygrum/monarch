package utils

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/docker"
	"github.com/pygrum/monarch/pkg/log"
)

var cLogger log.Logger

func init() {
	cLogger, _ = log.NewLogger(log.ConsoleLogger, "")
}

// Cleanup is used to delete all data associated with a builder
func Cleanup(builder *db.Builder) error {
	ctx := context.Background()
	if err := docker.Cli.ContainerStop(ctx, builder.ContainerID, container.StopOptions{}); err != nil {
		cLogger.Error("failed to stop container for %s: %v", builder.Name, err)
		// don't return, we will force removal
	}
	if err := docker.Cli.ContainerRemove(ctx, builder.ContainerID, types.ContainerRemoveOptions{Force: true}); err != nil {
		return fmt.Errorf("failed to remove container for %s: %v", builder.Name, err)
	}
	m, err := docker.Cli.ImageRemove(ctx, builder.ImageID, types.ImageRemoveOptions{Force: true})
	if err != nil {
		return fmt.Errorf("failed to remove image for %s: %v", builder.Name, err)
	}
	for _, i := range m {
		cLogger.Info("%s: untagged", i.Untagged)
		cLogger.Info("deleted %s", i.Deleted)
	}
	return db.Delete(builder)
}
