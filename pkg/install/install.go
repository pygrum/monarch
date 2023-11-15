package install

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/uuid"
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/consts"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/docker"
	"github.com/pygrum/monarch/pkg/log"
	"io"
	"path/filepath"
	"strings"
)

const (
	configName       = "royal.yaml"
	nativeTranslator = "native"
)

var (
	l log.Logger
)

func init() {
	l, _ = log.NewLogger(log.ConsoleLogger, "")
}

type imageBuildResponse struct {
	Aux map[string]string
}

// NewRepo and use GitHub credentials if repository is private
func NewRepo(url string, private bool) error {
	c := config.MonarchConfig{}
	if err := config.YamlConfig(config.MonarchConfigFile, &c); err != nil {
		return err
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
	a, t, err := setup(clonePath)
	if err != nil {
		return err
	}
	if t != nil {
		if err = db.Create(t); err != nil {
			return err
		}
	}
	return db.Create(a)
}

func setup(path string) (*db.Builder, *db.Translator, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, nil, err
	}

	configPath := filepath.Join(path, configName)
	royal := config.ProjectConfig{}
	if err := config.YamlConfig(configPath, &royal); err != nil {
		return nil, nil, err
	}
	l.Success("success! installing: %s v%s", royal.Name, royal.Version)
	// TODO:ImageBuild using dockerfile for builder, save image/container info to DB

	reader, err := archive.Tar(path, archive.Gzip)
	if err != nil {
		return nil, nil, err
	}
	defer reader.Close()
	var builderImageID, translatorImageID, builderImageTag, translatorImageTag string
	// Build builder image
	resp, err := cli.ImageBuild(ctx, reader, types.ImageBuildOptions{
		Dockerfile: filepath.Join(consts.DockerfilesPath, consts.BuilderDockerfile),
		Tags:       []string{royal.Name + ":" + royal.Version},
		PullParent: false,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build builder-builder image: %v", err)
	}
	bytes, _ := io.ReadAll(resp.Body)

	_ = resp.Body.Close()

	ID, buildErr, err := parseResponse(string(bytes))
	if buildErr != nil {
		return nil, nil, fmt.Errorf("build error while installing builder: %v", buildErr)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read build response: %v", err)
	}
	builderImageID = ID
	builderImageTag = royal.Name + ":" + royal.Version
	l.Success("successfully created builder image: %s", builderImageID)
	// Create translator image if translator is inbuilt
	if royal.TranslatorType == nativeTranslator {
		l.Info("building native translator %s", royal.TranslatorName)
		// Need to create a new reader, I think that calling an image build closes the old one
		trReader, err := archive.Tar(path, archive.Gzip)
		if err != nil {
			return nil, nil, err
		}
		defer trReader.Close()
		resp, err = cli.ImageBuild(ctx, trReader, types.ImageBuildOptions{
			Dockerfile: filepath.Join(consts.DockerfilesPath, consts.TranslatorDockerfile),
			Tags:       []string{royal.TranslatorName + ":" + royal.Version},
			PullParent: false,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to build translator image: %v", err)
		}
		bytes, _ = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		ID, buildErr, err = parseResponse(string(bytes))
		if buildErr != nil {
			return nil, nil, fmt.Errorf("build error while installing builder: %v", err)
		}
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read build response: %v", err)
		}
		translatorImageID = ID
		translatorImageTag = royal.TranslatorName + ":" + royal.Version
		l.Success("successfully created translator image: %s", translatorImageID)
	}
	buildContainerID, trContainerID, err := docker.StartContainers(cli, ctx, builderImageTag, translatorImageTag)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start builder services: %v", err)
	}

	builderID := uuid.New().String()
	translatorID := uuid.New().String()
	builder := &db.Builder{
		BuilderID:    builderID,
		Name:         royal.Name,
		Version:      royal.Version,
		Author:       royal.Author,
		Url:          royal.URL,
		InstalledAt:  path,
		ImageID:      builderImageID,
		ContainerID:  buildContainerID,
		TranslatorID: translatorID,
	}
	if len(trContainerID) != 0 {
		translator := &db.Translator{
			TranslatorID: translatorID,
			Name:         royal.TranslatorName,
			Version:      royal.Version,
			Author:       royal.Author,
			Url:          royal.URL,
			InstalledAt:  path,
			ImageID:      translatorImageID,
			ContainerID:  trContainerID,
		}
		return builder, translator, nil
	}
	// Return empty to avoid nil pointer dereference
	return builder, &db.Translator{}, nil
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
