package tcp

import (
	"encoding/binary"
	"fmt"
	"github.com/google/uuid"
	"github.com/pygrum/monarch/pkg/consts"
	"github.com/pygrum/monarch/pkg/transport"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	testAgentID        = "0123456789abcdef"
	testOS             = "windows"
	testArch           = "amd64"
	testUsername       = "WINDOWS11\\bob"
	testHostname       = "WINDOWS11"
	testUID            = "1000"
	testGID            = "2000"
	testPID            = "3000"
	testHomeDir        = "C:\\Users\\bob"
	testOpcode   int32 = 10
)

var (
	testRequestID = uuid.New().String()
	testArgs      = []string{
		"ABC",
		"DEF",
	}
	testResponses = []transport.ResponseDetail{
		{
			0, 0, "ABC", []byte("ABC"),
		},
		{
			1, 1, "", []byte("DEF"),
		},
	}
)

func TestMarshalRequest(t *testing.T) {
	req := &transport.GenericHTTPRequest{
		AgentID:   testAgentID,
		RequestID: testRequestID,
		Opcode:    testOpcode,
	}
	for _, a := range testArgs {
		req.Args = append(req.Args, []byte(a))
	}
	data, err := MarshalRequest(req)
	if err != nil {
		t.Fatalf("MarshalRequest failed: %v", err)
	}
	newReq, err := parseRequest(data)
	if err != nil {
		t.Fatalf("couldn't parse request: %v", err)
	}
	checkRequest(t, newReq)
}

func TestParseResponse(t *testing.T) {
	resp := &transport.GenericHTTPResponse{
		AgentID:   testAgentID,
		RequestID: testRequestID,
		Responses: testResponses,
	}
	data, err := marshalResponse(resp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	newResp, err := ParseResponse(data)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}
	checkResponse(t, newResp)
}

func TestParseRegistration(t *testing.T) {
	reg := &transport.Registration{
		AgentID:  testAgentID,
		OS:       testOS,
		Arch:     testArch,
		Username: testUsername,
		Hostname: testHostname,
		UID:      testUID,
		GID:      testGID,
		PID:      testPID,
		HomeDir:  testHomeDir,
	}
	data, err := marshalRegistration(reg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	newReg, _, err := ParseRegistration(data)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}
	checkRegistration(t, newReg)
}

func checkRequest(t *testing.T, req *transport.GenericHTTPRequest) {
	assert.Equal(t, testAgentID, req.AgentID, "original and parsed agent IDs do not match")
	assert.Equal(t, testRequestID, req.RequestID, "original and parsed request IDs do not match")
	assert.Equal(t, testOpcode, req.Opcode, "original and parsed request IDs do not match")
	assert.Equal(t, len(testArgs), len(req.Args), "original and parsed arg counts do not match")
	for i, arg := range req.Args {
		assert.Equal(t, testArgs[i], string(arg), "original and parsed args at index %d do not match", i)
	}
}

func checkResponse(t *testing.T, resp *transport.GenericHTTPResponse) {
	assert.Equal(t, testAgentID, resp.AgentID, "original and parsed agent IDs do not match")
	assert.Equal(t, testRequestID, resp.RequestID, "original and parsed request IDs do not match")
	assert.Equal(t, len(testArgs), len(resp.Responses), "original and parsed response counts do not match")
	for i, r := range resp.Responses {
		assert.Equal(t, testResponses[i].Status, r.Status, "original and parsed statuses at %d do not match", i)
		assert.Equal(t, testResponses[i].Dest, r.Dest, "original and parsed destinations at %d do not match", i)
		assert.Equal(t, testResponses[i].Name, r.Name, "original and parsed names at %d do not match", i)
		assert.Equal(t, testResponses[i].Data, r.Data, "original and parsed data at %d do not match", i)
	}
}

func checkRegistration(t *testing.T, reg *transport.Registration) {
	assert.Equal(t, testAgentID, reg.AgentID, "original and parsed agent IDs do not match")
	assert.Equal(t, testOS, reg.OS, "original and parsed OSes do not match")
	assert.Equal(t, testArch, reg.Arch, "original and parsed agent architectures do not match")
	assert.Equal(t, testUsername, reg.Username, "original and parsed usernames do not match")
	assert.Equal(t, testHostname, reg.Hostname, "original and parsed hostnames do not match")
	assert.Equal(t, testUID, reg.UID, "original and parsed UIDs do not match")
	assert.Equal(t, testGID, reg.GID, "original and parsed GIDs do not match")
	assert.Equal(t, testPID, reg.PID, "original and parsed PIDs do not match")
	assert.Equal(t, testHomeDir, reg.HomeDir, "original and parsed home directories do not match")
}

func parseRequest(data []byte) (req *transport.GenericHTTPRequest, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	var args [][]byte
	data = data[4:] // ignore prepended size
	var offset uint32
	var next uint32 = consts.AgentIDSize * 2
	agentID := string(data[offset:next])
	offset = next
	next += consts.RequestIDLength
	requestID := string(data[offset:next])
	offset = next
	next += 4
	opcode := binary.BigEndian.Uint32(data[offset:next])
	offset = next
	next += 4
	numArgs := binary.BigEndian.Uint32(data[offset:next])
	for i := 0; i < int(numArgs); i++ {
		var s string
		s, next, err = ParseField(next, data)
		if err != nil {
			return nil, err
		}
		args = append(args, []byte(s))
	}
	req = &transport.GenericHTTPRequest{
		AgentID:   agentID,
		RequestID: requestID,
		Opcode:    int32(opcode),
		Args:      args,
	}
	return
}

func marshalResponse(resp *transport.GenericHTTPResponse) (packet []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	size := getRespPacketSize(resp)
	packet = make([]byte, size)
	offset := 0
	next := offset + (consts.AgentIDSize * 2)
	copy(packet[offset:next], resp.AgentID)
	offset = next
	next += consts.RequestIDLength
	copy(packet[offset:next], resp.RequestID)
	offset = next
	next += 4
	// marshal the number of responses
	binary.BigEndian.PutUint32(packet[offset:next], uint32(len(resp.Responses)))
	for _, r := range resp.Responses {
		packet[next] = byte(r.Status)
		next++
		packet[next] = byte(r.Dest)
		next++
		offset = next
		next += 4
		binary.BigEndian.PutUint32(packet[offset:next], uint32(len(r.Name)))
		offset = next
		next += len(r.Name)
		copy(packet[offset:next], r.Name)
		offset = next
		next += 4
		binary.BigEndian.PutUint32(packet[offset:next], uint32(len(r.Data)))
		offset = next
		next += len(r.Data)
		copy(packet[offset:next], r.Data)
		offset = next
	}
	return
}

func marshalRegistration(reg *transport.Registration) (packet []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	size := getRegPacketSize(reg)
	packet = make([]byte, size)
	offset := 0
	next := offset + (consts.AgentIDSize * 2)
	copy(packet[offset:next], reg.AgentID)
	offset = next
	next += 4
	binary.BigEndian.PutUint32(packet[offset:next], uint32(len(reg.OS)))
	offset = next
	next += len(reg.OS)
	copy(packet[offset:next], reg.OS)
	offset = next
	next += 4
	binary.BigEndian.PutUint32(packet[offset:next], uint32(len(reg.Arch)))
	offset = next
	next += len(reg.Arch)
	copy(packet[offset:next], reg.Arch)
	offset = next
	next += 4
	binary.BigEndian.PutUint32(packet[offset:next], uint32(len(reg.Username)))
	offset = next
	next += len(reg.Username)
	copy(packet[offset:next], reg.Username)
	offset = next
	next += 4
	binary.BigEndian.PutUint32(packet[offset:next], uint32(len(reg.Hostname)))
	offset = next
	next += len(reg.Hostname)
	copy(packet[offset:next], reg.Hostname)
	offset = next
	next += 4
	binary.BigEndian.PutUint32(packet[offset:next], uint32(len(reg.UID)))
	offset = next
	next += len(reg.UID)
	copy(packet[offset:next], reg.UID)
	offset = next
	next += 4
	binary.BigEndian.PutUint32(packet[offset:next], uint32(len(reg.GID)))
	offset = next
	next += len(reg.GID)
	copy(packet[offset:next], reg.GID)
	offset = next
	next += 4
	binary.BigEndian.PutUint32(packet[offset:next], uint32(len(reg.PID)))
	offset = next
	next += len(reg.PID)
	copy(packet[offset:next], reg.PID)
	offset = next
	next += 4
	binary.BigEndian.PutUint32(packet[offset:next], uint32(len(reg.HomeDir)))
	offset = next
	next += len(reg.HomeDir)
	copy(packet[offset:next], reg.HomeDir)
	offset = next
	return packet, nil
}

func getRespPacketSize(resp *transport.GenericHTTPResponse) uint32 {
	var size int
	size += consts.AgentIDSize * 2
	size += consts.RequestIDLength
	size += 4 // num of responses
	for _, r := range resp.Responses {
		size += 2 // status and dest
		size += 4 // len(name) as uint32
		size += len(r.Name)
		size += 4 // len(data) as uint32
		size += len(r.Data)
	}
	return uint32(size)
}

func getRegPacketSize(reg *transport.Registration) uint32 {
	var size int
	size += consts.AgentIDSize * 2
	size += 4
	size += len(reg.OS)
	size += 4
	size += len(reg.Arch)
	size += 4
	size += len(reg.Username)
	size += 4
	size += len(reg.Hostname)
	size += 4
	size += len(reg.UID)
	size += 4
	size += len(reg.GID)
	size += 4
	size += len(reg.PID)
	size += 4
	size += len(reg.HomeDir)

	return uint32(size)
}
