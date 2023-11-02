// Package config is responsible for managing configuration variables that are used by the application.
// These variables are set as environment variables, and are utilised both by the main application and connected
// containers.
package config

import (
	"github.com/goccy/go-yaml"
	"github.com/kelseyhightower/envconfig"
	"os"
)

const appName = "monarch"

type MonarchConfig struct {
	// Specifies whether Monarch is in debug mode or not.
	Debug bool
	// Set monarch logging level, which can be one of: debug (1), informational (2), warning (3), fatal (4)
	LogLevel uint16
	// Path to the certificate file used for TLS enabled connections.
	CertFile string
	// Path to the key file used for TLS enabled connections.
	KeyFile string
	// The main interface that Monarch will bind to for operations.
	Interface string
	// The port to use for the Monarch HTTP listener.
	HttpPort int
	// The port to use for the Monarch HTTPS listener.
	HttpsPort int
	// The folder where agent and c2 repositories are installed to.
	InstallDir string
	// Credentials used by git for installing private packages
	GitUsername string
	GitPAT      string
	// Ignore console warning logs
	IgnoreConsoleWarnings bool
}

type ProjectConfig struct {
	Name    string
	Version string
	// The translator is used to translate messages between the C2 and agent.
	// A translator can use an existing translator (type=external) or the one included in the cloned project
	// (type=native). The translator is installed as a container and given the name provided by `translator_name`.
	TranslatorName string
	TranslatorType string
	// The command schema defines the possible commands that can be used with the agent.
	// If the agent doesn't use commands to operate, then this configuration parameter is not necessary.
	// On installation of the agent, the command schema is used by the translator when an operator requests to
	// view commands.
	CmdSchema []ProjectConfigCmd
	// The script used to build the agent
	BuildScript string
	// The path that the agent is created at after a successful build
	BuildPath string
	// The directory where the build routine takes place
	SourceDir string
	// These are custom build arguments that can be used for building, in addition to default build arguments provided
	// by the C2 itself.
	BuildArgs []ProjectConfigBuildArgs
}

type ProjectConfigCmd struct {
	Name  string
	Usage string
	// Specifies whether this command requires admin privileges or not
	Admin bool
	// If opcode is specified, the provided integer opcode is used in place of the command name,
	// promoting better OpSec
	Opcode           uint
	DescriptionShort string
	DescriptionLong  string
}

type ProjectConfigBuildArgs struct {
	Name     string
	Required bool
}

// EnvConfig fetches the app configuration from environment variables and attempts to unmarshal them using the
// provided configuration template pointer.
func EnvConfig(config interface{}) error {
	return envconfig.Process(appName, config)
}

// YamlConfig will unmarshal yaml data into the provided template pointer.
func YamlConfig(path string, config interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, config)
}
