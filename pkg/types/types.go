package types

import (
	"fmt"
	"github.com/pygrum/monarch/pkg/protobuf/rpcpb"
)

type Queue interface {
	Enqueue(interface{}) error
	Dequeue() interface{}
	Size() int
}

type NotificationQueue struct {
	Channel chan *rpcpb.Notification
}

const (
	NotificationKickPlayer = "you have been kicked from the operation"
)

var NotifQueues map[string]Queue

func (r *NotificationQueue) Enqueue(req interface{}) error {
	select {
	case r.Channel <- req.(*rpcpb.Notification):
		return nil
	default:
		return fmt.Errorf("queue is full - max capacity of 10")
	}
}

func (r *NotificationQueue) Dequeue() interface{} {
	// Must block, as we wait for a request to queue
	select {
	case req := <-r.Channel:
		return req
	}
}

func (r *NotificationQueue) Size() int {
	return len(r.Channel)
}
