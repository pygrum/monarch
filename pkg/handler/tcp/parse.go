package tcp

import (
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/pygrum/monarch/pkg/consts"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/protobuf/builderpb"
	"github.com/pygrum/monarch/pkg/transport"
)

// ParseRegistration parses data received upon initial connection
func ParseRegistration(data []byte) (reg *transport.Registration, a *db.Agent, retErr error) {
	// handle parsing errors
	defer func() {
		if r := recover(); r != nil {
			retErr = fmt.Errorf("invalid data received from connection: %v", r)
		}
	}()
	var offset uint32 = consts.AgentIDSize * 2
	// hex encoded so we multiply size in bytes by 2 to get ascii length
	agentID := string(data[:offset])
	agent := &db.Agent{}
	if flag.Lookup("test.v") == nil {
		if err := db.FindOneConditional("agent_id = ?", agentID, agent); err != nil {
			return nil, nil, err
		}
		if len(agent.Name) == 0 {
			return nil, nil, fmt.Errorf("no agent with ID %s exists", agentID)
		}
	}
	rOS, next, err := ParseField(offset, data)
	if err != nil {
		return nil, nil, err
	}
	rArch, next, err := ParseField(next, data)
	if err != nil {
		return nil, nil, err
	}
	rUser, next, err := ParseField(next, data)
	if err != nil {
		return nil, nil, err
	}
	rHost, next, err := ParseField(next, data)
	if err != nil {
		return nil, nil, err
	}
	rUID, next, err := ParseField(next, data)
	if err != nil {
		return nil, nil, err
	}
	rGID, next, err := ParseField(next, data)
	if err != nil {
		return nil, nil, err
	}
	rPID, next, err := ParseField(next, data)
	if err != nil {
		return nil, nil, err
	}
	rHome, next, err := ParseField(next, data)
	if err != nil {
		return nil, nil, err
	}
	reg = &transport.Registration{
		AgentID:  agentID,
		OS:       rOS,
		Arch:     rArch,
		Username: rUser,
		Hostname: rHost,
		UID:      rUID,
		GID:      rGID,
		PID:      rPID,
		HomeDir:  rHome,
	}
	a = agent
	return
}

func ParseResponse(data []byte) (resp *transport.GenericHTTPResponse, retErr error) {
	var (
		offset    uint32 = consts.AgentIDSize * 2
		responses []transport.ResponseDetail
		err       error
	)
	defer func() {
		if r := recover(); r != nil {
			retErr = fmt.Errorf("invalid data received from connection: %v", r)
		}
	}()
	agentID := string(data[:offset])
	next := offset + consts.RequestIDLength
	requestID := string(data[offset:next])
	offset = next
	next += 4
	c := data[offset:next]
	count := binary.BigEndian.Uint32(c)
	for i := 0; i < int(count); i++ {
		var name, d string
		status := data[next]
		next++
		dest := data[next]
		next++
		name, next, err = ParseField(next, data)
		if err != nil {
			return nil, err
		}
		d, next, err = ParseField(next, data)
		if err != nil {
			return nil, err
		}
		responses = append(responses, transport.ResponseDetail{
			Status: builderpb.Status(int(status)),
			Dest:   int32(dest),
			Name:   name,
			Data:   []byte(d),
		})
	}
	resp = &transport.GenericHTTPResponse{
		AgentID:   agentID,
		RequestID: requestID,
		Responses: responses,
	}
	return
}

func ParseField(sizeOffset uint32, data []byte) (str string, nextOffset uint32, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("invalid data received from connection: %v", r)
		}
	}()
	s := data[sizeOffset : sizeOffset+4]
	size := binary.BigEndian.Uint32(s)
	ret := data[sizeOffset+4 : sizeOffset+4+size]
	return string(ret), sizeOffset + 4 + size, nil
}
