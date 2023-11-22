package transport

// GenericHTTPRequest is the structure sent by the C2 when an operator requests for an agent to perform a task.
type GenericHTTPRequest struct {
	AgentID   string
	RequestID int32
	Opcode    int32
	Args      [][]byte
}

type ResponseDetail struct {
	Dest int32
	Data []byte
}

// GenericHTTPResponse is the structure received from an agent after a task is performed
type GenericHTTPResponse struct {
	AgentID   string
	RequestID int32
	Status    int32
	Responses []ResponseDetail
}
