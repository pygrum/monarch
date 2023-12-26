package tcp

import (
	"encoding/binary"
	"fmt"
	"github.com/pygrum/monarch/pkg/consts"
	"github.com/pygrum/monarch/pkg/transport"
)

func MarshalRequest(req *transport.GenericHTTPRequest) (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	size := getPacketSize(req)
	packet := make([]byte, size)
	offset := 4
	binary.BigEndian.PutUint32(packet[:offset], size-uint32(offset))
	next := offset + (consts.AgentIDSize * 2)
	// agent ID
	copy(packet[offset:next], req.AgentID)
	offset = next
	next += consts.RequestIDLength
	// request ID
	copy(packet[offset:next], req.RequestID)
	offset = next
	next += 4
	// opcode
	binary.BigEndian.PutUint32(packet[offset:next], uint32(req.Opcode))
	offset = next
	next += 4
	// number of args
	binary.BigEndian.PutUint32(packet[offset:next], uint32(len(req.Args)))
	offset = next
	for _, arg := range req.Args {
		next += 4
		// arg[n]
		binary.BigEndian.PutUint32(packet[offset:next], uint32(len(arg)))
		offset = next
		next += len(arg)
		copy(packet[offset:next], arg)
		offset = next
	}
	return packet, nil
}

func getPacketSize(r *transport.GenericHTTPRequest) uint32 {
	var size uint32 = 4 // total size is prepended as uint32, hence initial value of 4
	size += consts.AgentIDSize * 2
	size += consts.RequestIDLength
	size += consts.OpcodeLength
	size += 4 // number of args as uint32
	for _, arg := range r.Args {
		size += 4 // prepend the length of each argument (uint32)
		size += uint32(len(arg))
	}
	return size
}
