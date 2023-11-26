package transport

import (
	"github.com/pygrum/monarch/pkg/rpcpb"
)

const (
	DestFile int32 = iota
	DestStdout
)

// GenericHTTPRequest is the structure sent by the C2 when an operator requests for an agent to perform a task.
type GenericHTTPRequest struct {
	AgentID   string   `json:"agent_id"`
	RequestID string   `json:"request_id"`
	Opcode    int32    `json:"opcode"`
	Args      [][]byte `json:"args"`
}

type ResponseDetail struct {
	Dest int32  `json:"dest"` // Where to send response to (file, stdout)
	Name string `json:"name"` // Name of file to save, if applicable
	Data []byte `json:"data"` // file or output data
}

// Registration is the initial data that is received from a first-time authenticating agent
// Can be viewed with the 'info' command
type Registration struct {
	AgentID   string `json:"agent_id"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	Username  string `json:"username"`
	Hostname  string `json:"hostname"`
	UID       string `json:"uid"`
	GID       string `json:"gid"`
	PID       string `json:"pid"`
	HomeDir   string `json:"home_dir"`
	IPAddress string `json:"ip_address"`
	// Leftover response in case agent is de-authed but has a response, so it's still processed
	Data *GenericHTTPResponse `json:"data"`
}

// GenericHTTPResponse is the structure received from an agent after a task is performed
type GenericHTTPResponse struct {
	AgentID   string           `json:"agent_id"`
	RequestID string           `json:"request_id"`
	Status    rpcpb.Status     `json:"status"`
	Responses []ResponseDetail `json:"responses"`
}
