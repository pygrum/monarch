// Package config is responsible for managing configuration variables that are used by the application.
// These variables are set as environment variables, and are utilised both by the main application and connected
// containers.
package config

import "github.com/kelseyhightower/envconfig"

const appName = "monarch"

type MonarchConfig struct {
	// Specifies whether Monarch is in debug mode or not.
	Debug bool
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
}

// Config fetches the app configuration from environment variables and attempts to unmarshal them using the
// provided configuration template pointer.
func Config(config interface{}) error {
	return envconfig.Process(appName, config)
}
