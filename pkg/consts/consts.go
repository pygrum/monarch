package consts

const (
	DockerfilesPath   = "docker"
	BuilderDockerfile = "builder/Dockerfile"

	MonarchNet = "monarch-net"
	Version    = "0.1.0" // track with container
	ServerUser = "server"

	CoopHelpGroup    = "Co-op / Multiplayer"
	GeneralHelpGroup = "General"
	AdminHelpGroup   = "Admin"
	BuildHelpGroup   = "Build"

	DefaultPrompt   = "monarch > "
	AgentIDSize     = 8
	RequestIDLength = 36
	OpcodeLength    = 4 // uint32

	ProfileTCP          = "tcp"
	ProfileHTTP         = "http"
	ProfileHTTPS        = "https"
	TypeInternalProfile = "internal"

	OpTypeBool   = "bool"
	OpTypeInt    = "int"
	OpTypeFloat  = "float"
	OpTypeString = "string"

	OpLPort = "lport"
)
