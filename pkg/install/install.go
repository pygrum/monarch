package install

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/uuid"
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/consts"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/docker"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/protobuf/rpcpb"
	"github.com/pygrum/monarch/pkg/utils"
	"io"
	"path/filepath"
	"strings"
	"sync"
)

const (
	configName       = "royal.yaml"
	nativeTranslator = "native"
)

var (
	l  log.Logger
	mu sync.Mutex
)

func init() {
	l, _ = log.NewLogger(log.ConsoleLogger, "")
}

// NewRepo and use GitHub credentials if repository is private
func NewRepo(url, branch string, private bool, stream rpcpb.Monarch_InstallServer) error {
	c := config.MainConfig
	o := &git.CloneOptions{
		URL: url,
	}
	if len(branch) != 0 {
		o.SingleBranch = true
		o.ReferenceName = plumbing.ReferenceName(branch)
	}
	if len(c.GitUsername) == 0 || len(c.GitPAT) == 0 {
		if private {
			return errors.New("github credentials not configured")
		}
		if !c.IgnoreConsoleWarnings {
			if flag.Lookup("test.v") == nil {
				_ = stream.Send(&rpcpb.Notification{
					LogLevel: rpcpb.LogLevel_LevelWarn,
					Msg:      "Your GitHub credentials have not been configured",
				})
			}
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
		if errors.Is(err, git.ErrRepositoryAlreadyExists) {
			return fmt.Errorf("a repository with that name already exists")
		}
		return err
	}
	// Build arguments passed with environment variables.
	a, err := Setup(clonePath, stream)
	if err != nil {
		return err
	}
	return db.Create(a)
}

func Setup(path string, stream rpcpb.Monarch_InstallServer) (*db.Builder, error) {
	mu.Lock()
	defer mu.Unlock()
	ctx := context.Background()
	configPath := filepath.Join(path, configName)
	royal := config.ProjectConfig{}
	if err := config.YamlConfig(configPath, &royal); err != nil {
		return nil, err
	}
	if flag.Lookup("test.v") == nil {
		_ = stream.Send(&rpcpb.Notification{
			LogLevel: rpcpb.LogLevel_LevelSuccess,
			Msg:      fmt.Sprintf("success! installing: %s v%s", royal.Name, royal.Version),
		})
	}
	if len(royal.Name) == 0 || len(royal.Version) == 0 {
		return nil, fmt.Errorf("name and / or version missing (configuration file at %s)", path)
	}
	royal.Name = strings.ToLower(royal.Name)
	b := &db.Builder{}
	if flag.Lookup("test.v") == nil {
		if err := db.FindOneConditional("name = ?", royal.Name, b); err == nil {
			// just to check that we actually returned sum
			if b.Name == royal.Name {
				if err = utils.Cleanup(b, stream); err != nil {
					return nil, fmt.Errorf("failed to delete existing builder: %v", err)
				}
			}
		}
	}
	reader, err := archive.Tar(path, archive.Gzip)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	var builderImageID, builderImageTag string
	// Build builder image
	resp, err := docker.Cli.ImageBuild(ctx, reader, types.ImageBuildOptions{
		Dockerfile: filepath.Join(consts.DockerfilesPath, consts.BuilderDockerfile),
		Tags:       []string{royal.Name + ":" + royal.Version},
		PullParent: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build builder image: %v", err)
	}
	bytes, _ := io.ReadAll(resp.Body)

	_ = resp.Body.Close()

	ID, buildErr, err := parseResponse(string(bytes))
	if buildErr != nil {
		return nil, fmt.Errorf("build error while installing builder: %v", buildErr)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read build response: %v", err)
	}
	builderImageID = ID
	builderImageTag = royal.Name + ":" + royal.Version
	if flag.Lookup("test.v") == nil {
		_ = stream.Send(&rpcpb.Notification{
			LogLevel: rpcpb.LogLevel_LevelSuccess,
			Msg:      fmt.Sprintf("successfully created builder image: %s", builderImageID),
		})
	}
	buildContainerID, err := docker.StartContainer(docker.Cli, ctx, builderImageTag)
	if err != nil {
		return nil, fmt.Errorf("failed to start builder services: %v", err)
	}

	builderID := uuid.New().String()
	builder := &db.Builder{
		BuilderID:   builderID,
		Name:        royal.Name,
		Version:     royal.Version,
		Author:      royal.Author,
		Url:         royal.URL,
		InstalledAt: path,
		ImageID:     builderImageID,
		ContainerID: buildContainerID,
		SupportedOS: strings.Join(royal.SupportedOSes, ", "),
	}
	// Return empty to avoid nil pointer dereference
	return builder, nil
}

// parseResponse returns the image ID, build errors and any errors that occurred during parsing
func parseResponse(s string) (string, error, error) {
	responses := strings.Split(s, "\n")
	for _, response := range responses {
		if len(response) == 0 {
			continue
		}
		responseMap := make(map[string]interface{})
		if err := json.Unmarshal([]byte(response), &responseMap); err != nil {
			return "", nil, err
		}
		errMsg, ok := responseMap["error"]
		if ok {
			return "", errors.New(errMsg.(string)), nil
		}
		success, ok := responseMap["aux"].(map[string]interface{})
		if !ok {
			continue
		}
		return success["ID"].(string), nil, nil
	}
	return "", errors.New("could not retrieve image ID"), nil
}
