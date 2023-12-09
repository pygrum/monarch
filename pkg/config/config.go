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
	MonarchConfigFile string
)

// MonarchConfig is only used by monarch itself, not clients
type MonarchConfig struct {
	// Set monarch logging level, which can be one of: debug (1), informational (2), warning (3), fatal (4)
	LogLevel uint16
	// Path to the certificate file used for TLS enabled connections.
	CertFile string
	// Path to the key file used for TLS enabled connections.
	KeyFile string
	// certificate authority x509 key pair
	CaCert string
	CaKey  string
	// The main interface that Monarch will bind to for operations.
	Interface string
	// The port to use for the Monarch HTTP listener.
	HttpPort int
	// The port to use for the Monarch HTTPS listener.
	HttpsPort int
	// RPC port for multiplayer
	MultiplayerPort int
	// The endpoint where agents receive and respond to commands
	MainEndpoint string
	// The endpoint where agents login / register themselves
	LoginEndpoint string
	// Endpoint where payloads are staged
	StageEndpoint string
	// The folder where agent and c2 repositories are installed to.
	SessionTimeout int `yaml:"session_timeout_minutes"`
	InstallDir     string
	// Credentials used by git for installing private packages
	GitUsername string
	GitPAT      string
	// Ignore console warning logs
	IgnoreConsoleWarnings bool
	MysqlAddress          string
	MysqlUsername         string
	MysqlPassword         string
}

type MonarchClientConfig struct {
	UUID    string `json:"uuid"`
	Name    string `json:"name"`
	RHost   string `json:"rhost"`
	RPort   int    `json:"rport"`
	CertPEM []byte `json:"cert_pem"`
	KeyPEM  []byte `json:"key_pem"`
}

type ProjectConfig struct {
	Name          string
	Version       string
	Author        string
	URL           string
	SupportedOSes []string `yaml:"supported_os"`
	// The command schema defines the possible commands that can be used with the agent.
	// If the agent doesn't use commands to operate, then this configuration parameter is not necessary.
	// On installation of the agent, the command schema is used by the builder when an operator requests to
	// view commands.
	CmdSchema []ProjectConfigCmd `yaml:"cmd_schema"`
	Builder   Builder            `yaml:"builder"`
}

type ProjectConfigCmd struct {
	Name    string
	Usage   string
	MinArgs int32 `yaml:"min_args"`
	MaxArgs int32 `yaml:"max_args"` // Whether NArgs represents the minimum arg count or the exact
	// Specifies whether this command requires admin privileges or not
	Admin bool
	// If opcode is specified, the provided integer opcode is used in place of the command name,
	// promoting better OpSec
	Opcode           int32
	DescriptionShort string `yaml:"description_short"`
	DescriptionLong  string `yaml:"description_long"`
}

type Builder struct {
	// The directory where the build routine takes place
	SourceDir string `yaml:"source_dir"`
	// These are custom build arguments that can be used for building, in addition to default build arguments provided
	// by the C2 itself.
	BuildArgs []ProjectConfigBuildArg `yaml:"build_args"`
}

type ProjectConfigBuildArg struct {
	Name        string
	Description string
	Default     string
	Required    bool
	Type        string
	Choices     []string
}

func Initialize() {
	home, _ := os.UserHomeDir()
	MonarchConfigFile = filepath.Join(home, ".monarch", "monarch.yaml")

	if err := YamlConfig(MonarchConfigFile, &MainConfig); err != nil {
		panic(fmt.Errorf("%v. was monarch installed with install-monarch.sh? ", err))
	}
	MainConfig.CertFile = filepath.Join(home, ".monarch", MainConfig.CertFile)
	MainConfig.KeyFile = filepath.Join(home, ".monarch", MainConfig.KeyFile)
	MainConfig.InstallDir = filepath.Join(home, ".monarch", MainConfig.InstallDir)
}

// ServerCertificates returns the PEM-encoded monarch server key pair
func ServerCertificates() ([]byte, []byte, error) {
	certPEM, err := os.ReadFile(MainConfig.CertFile)
	if err != nil {
		return nil, nil, err
	}
	keyPEM, err := os.ReadFile(MainConfig.CertFile)
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
