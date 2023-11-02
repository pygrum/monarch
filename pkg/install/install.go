package install

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/uuid"
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/consts"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/log"
	"io"
	"path/filepath"
	"reflect"
	"strings"
)

const (
	configName       = "royal.yaml"
	nativeTranslator = "native"
)

var l log.Logger

func init() {
	l, _ = log.NewLogger(log.ConsoleLogger, "")
}

type imageBuildResponse struct {
	Aux aux `json:"aux"`
}

type aux struct {
	ID string `json:"ID"`
}

// NewRepo and use GitHub credentials if repository is private
func NewRepo(url string, private bool) error {
	c := config.MonarchConfig{}
	if err := config.EnvConfig(&c); err != nil {
		return err
	}
	if c.Debug {
		_ = l.SetLogLevel(log.LevelDebug)
	}
	o := &git.CloneOptions{
		URL: url,
	}
	if len(c.GitUsername) == 0 || len(c.GitPAT) == 0 {
		if private {
			return errors.New("github credentials not configured")
		}
		if !c.IgnoreConsoleWarnings {
			l.Warn("Your GitHub credentials have not been configured")
		}
	}
	if private {
		o.Auth = &http.BasicAuth{
			Username: c.GitUsername,
			Password: c.GitPAT,
		}
	}
	clonePath := filepath.Join(c.InstallDir, strings.TrimSuffix(filepath.Base(url),
		filepath.Ext(filepath.Base(url))))
	_, err := git.PlainClone(clonePath, false, o)
	if err != nil {
		return err
	}
	// TODO:Find config, then find agent builder Dockerfile (and if necessary, translator) and start containers.
	// Build arguments passed with environment variables.
	return setup(clonePath)
}

func setup(path string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	configPath := filepath.Join(path, configName)
	royal := config.ProjectConfig{}
	if err := config.YamlConfig(configPath, &royal); err != nil {
		return err
	}
	l.Info("success! installing: %s %s\n", royal.Name, royal.Version)
	// TODO:ImageBuild using dockerfile for agent, save image/container info to DB

	reader, err := archive.Tar(path, archive.Gzip)
	if err != nil {
		return err
	}

	var builderImageID, translatorImageID, builderImageTag, translatorImageTag string
	resp, err := cli.ImageBuild(ctx, reader, types.ImageBuildOptions{
		Dockerfile: filepath.Join(path, consts.DockerfilesPath, consts.BuilderDockerfile),
		Tags:       []string{royal.Name + ":" + royal.Version},
		PullParent: true, // I think this means pull the container that the dockerfile builds on
	})
	if err != nil {
		return fmt.Errorf("failed to build agent-builder image: %v", err)
	}
	bytes, _ := io.ReadAll(resp.Body)

	_ = resp.Body.Close()
	obj := imageBuildResponse{}
	if err = json.Unmarshal(bytes, &obj); err != nil {
		return err
	}
	if !reflect.DeepEqual(obj.Aux, aux{}) {
		builderImageID = obj.Aux.ID
	}
	builderImageTag = royal.Name + ":" + royal.Version

	// Create translator image if translator is inbuilt
	if royal.TranslatorType == nativeTranslator {
		resp, err = cli.ImageBuild(ctx, reader, types.ImageBuildOptions{
			Dockerfile: filepath.Join(path, consts.DockerfilesPath, consts.TranslatorDockerfile),
			Tags:       []string{royal.TranslatorName + ":" + royal.Version},
			PullParent: true,
		})
		if err != nil {
			return fmt.Errorf("failed to build translator image: %v", err)
		}

		bytes, _ = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		obj := imageBuildResponse{}
		if err = json.Unmarshal(bytes, &obj); err != nil {
			return err
		}
		if !reflect.DeepEqual(obj.Aux, aux{}) {
			translatorImageID = obj.Aux.ID
		}
		translatorImageTag = royal.TranslatorName + ":" + royal.Version
	}
	buildContainerID, trContainerID, err := startContainers(cli, ctx, builderImageTag, translatorImageTag)
	if err != nil {
		return fmt.Errorf("failed to start agent services: %v", err)
	}

	agentID := uuid.New().String()
	translatorID := uuid.New().String()
	agent := &db.Agent{
		AgentID:            agentID,
		Name:               royal.Name,
		Version:            royal.Version,
		InstalledAt:        path,
		BuilderImageID:     builderImageID,
		BuilderContainerID: buildContainerID,
		TranslatorID:       translatorID,
	}
	translator := &db.Translator{
		TranslatorID: translatorID,
		Name:         royal.TranslatorName,
		Version:      royal.Version,
		InstalledAt:  path,
		ImageID:      translatorImageID,
		ContainerID:  trContainerID,
	}
	err = db.Create(agent)
	if err != nil {
		return err
	}
	return db.Create(translator)
}

func startContainers(cli *client.Client, ctx context.Context, builderImageTag, translatorImageTag string) (string, string, error) {
	// Run container with same name as image
	var builderID, translatorID string

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: builderImageTag,
		Tty:   false,
	}, nil, nil, nil, builderImageTag)
	if err != nil {
		return "", "", err
	}
	// Start builder container
	if err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", "", err
	}
	l.Info("started builder container %s", builderImageTag)
	builderID = resp.ID

	// Only start translator image if it exists
	if len(translatorImageTag) != 0 {
		resp, err = cli.ContainerCreate(ctx, &container.Config{
			Image: translatorImageTag,
			Tty:   false,
		}, nil, nil, nil, translatorImageTag)
		if err != nil {
			return "", "", err
		}
		// Start builder container
		if err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			return "", "", err
		}
		l.Info("started translator container %s", translatorImageTag)
		translatorID = resp.ID
	}
	return builderID, translatorID, nil
}
