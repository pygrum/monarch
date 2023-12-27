// Package config is responsible for managing configuration variables that are used by the application.
// These variables are set as environment variables, and are utilised both by the main application and connected
// containers.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

var (
	MainConfig        MonarchConfig
	ClientConfig      MonarchClientConfig
	C2Config          = newC2Config()
	MonarchConfigFile string
)

// TODO:MALLEABLE C2 - ALLOW ARRAY OF ENDPOINTS FOR EACH HTTP(S) ENDPOINT (LOGIN, MAIN, ETC)

func Initialize() {
	home, _ := os.UserHomeDir()
	MonarchConfigFile = filepath.Join(home, ".monarch", "monarch.yaml")

	if err := YamlConfig(MonarchConfigFile, &MainConfig); err != nil {
		panic(fmt.Errorf("%v. was monarch installed with install-monarch.sh? ", err))
	}
	MainConfig.HttpConfig = norm(MainConfig.HttpConfig)
	if err := JsonConfig(MainConfig.HttpConfig, &C2Config); err != nil {
		panic(fmt.Errorf("couldn't read C2 configuration: %v", err))
	}
	MainConfig.CertFile = norm(MainConfig.CertFile)
	MainConfig.KeyFile = norm(MainConfig.KeyFile)

	MainConfig.CaCert = norm(MainConfig.CaCert)
	MainConfig.CaKey = norm(MainConfig.CaKey)

	MainConfig.InstallDir = norm(MainConfig.InstallDir)
}

func Home() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".monarch")
}

func norm(s string) string {
	return filepath.Join(Home(), s)
}

// ServerCertificates returns the PEM-encoded monarch server key pair
func ServerCertificates() ([]byte, []byte, error) {
	certPEM, err := os.ReadFile(MainConfig.CertFile)
	if err != nil {
		return nil, nil, err
	}
	keyPEM, err := os.ReadFile(MainConfig.KeyFile)
	if err != nil {
		return nil, nil, err
	}
	return certPEM, keyPEM, nil
}

// YamlConfig will unmarshal yaml data into the provided template pointer.
func YamlConfig(path string, config interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, config)
}

func JsonConfig(path string, config interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, config)
}

func newC2Config() HttpConfig {
	loginCfg := &EndpointConfig{
		Headers: make(map[string]string),
	}
	mainCfg := &EndpointConfig{
		Headers: make(map[string]string),
	}
	stageCfg := &EndpointConfig{
		Headers: make(map[string]string),
	}
	config := HttpConfig{
		LoginEndpoint: loginCfg,
		MainEndpoint:  mainCfg,
		StageEndpoint: stageCfg,
	}
	return config
}
