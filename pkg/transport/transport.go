package transport

import "github.com/pygrum/monarch/pkg/rpcpb"

const (
	DestFile int32 = iota
	DestStdout
)

// GenericHTTPRequest is the structure sent by the C2 when an operator requests for an agent to perform a task.
type GenericHTTPRequest struct {
	AgentID   string
	RequestID string
	Opcode    int32
	Args      [][]byte
}

type ResponseDetail struct {
	Dest int32  // Where to send response to (file, stdout)
	Name string // Name of file to save, if applicable
	Data []byte // file or output data
}

// GenericHTTPResponse is the structure received from an agent after a task is performed
type GenericHTTPResponse struct {
	AgentID   string
	RequestID string
	Status    rpcpb.Status
	Responses []ResponseDetail
}
