package install

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/pygrum/monarch/pkg/consts"
	"github.com/pygrum/monarch/pkg/db"
	"os"
	"path/filepath"
	"testing"
)

var dir string

func init() {
	var err error
	dir, err = os.MkdirTemp("", "tester-*")
	if err != nil {
		panic(err)
	}
	dockerDir := filepath.Join(dir, consts.DockerfilesPath)
	builderDir := filepath.Join(dir, consts.DockerfilesPath, "builder")
	translateDir := filepath.Join(dir, consts.DockerfilesPath, "translator")
	err = os.MkdirAll(builderDir, 0777)
	checkErr(err)
	err = os.Mkdir(translateDir, 0777)
	checkErr(err)

	templateDir := filepath.Join("..", "..", "templates")
	// Write build and translate Dockerfiles to respective files in simulated directory
	buildBytes, err := os.ReadFile(filepath.Join(templateDir, consts.BuilderDockerfile))
	checkErr(err)
	trBytes, err := os.ReadFile(filepath.Join(templateDir, consts.TranslatorDockerfile))
	checkErr(err)
	err = os.WriteFile(filepath.Join(dockerDir, consts.BuilderDockerfile), buildBytes, 0666)
	checkErr(err)
	err = os.WriteFile(filepath.Join(dockerDir, consts.TranslatorDockerfile), trBytes, 0666)
	checkErr(err)

	// Create config file in folder
	royalBytes, err := os.ReadFile(filepath.Join("..", "..", "configs", configName))
	checkErr(err)
	err = os.WriteFile(filepath.Join(dir, configName), royalBytes, 0666)
	checkErr(err)

}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func cleanup(agent *db.Builder) {
	err := os.RemoveAll(dir)
	ctx := context.Background()
	checkErr(err)
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	checkErr(err)
	if agent != nil {
		err = cli.ContainerRemove(ctx, agent.ContainerID, types.ContainerRemoveOptions{Force: true})
		checkErr(err)
	}
}

func TestSetup(t *testing.T) {
	a, err := setup(dir)
	if err != nil {
		cleanup(a)
		t.Fatalf("setup(%s): failed with error %v", dir, err)
	}
	cleanup(a)
}
