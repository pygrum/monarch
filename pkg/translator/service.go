package translator

import "github.com/pygrum/monarch/pkg/rpcpb"

// The json object used to request a translation of c2->agent msgs (localhost:20000)
type toAgentServiceRequest struct {
	AgentID   string   `json:"agent_id"`
	RequestID int32    `json:"request_id"`
	Opcode    int32    `json:"opcode"`
	Args      [][]byte `json:"args"`
}

// The json object received after a c2->agent translation request is fulfilled
type toAgentServiceResponse struct {
	Success  bool
	ErrorMsg string `json:"error_msg"`
	Message  []byte
}

// The json object received after an agent->c2 translation request is fulfilled
type fromAgentServiceResponse struct {
	rpcpb.Reply
	Success  bool
	ErrorMsg string `json:"error_msg"`
}
