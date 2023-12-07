package install

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/pygrum/monarch/pkg/db"
	"testing"
)

var dir = "testdata"

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func cleanup(agent *db.Builder) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	checkErr(err)
	if agent != nil {
		err = cli.ContainerRemove(ctx, agent.ContainerID, types.ContainerRemoveOptions{Force: true})
		checkErr(err)
	}
}

func TestSetup(t *testing.T) {
	a, err := Setup(dir, nil)
	if err != nil {
		cleanup(a)
		t.Fatalf("Setup(%s): failed with error %v", dir, err)
	}
	cleanup(a)
}
