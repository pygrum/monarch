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
type MessageQueue struct {
	Channel chan *rpcpb.Message
}

const (
	NotificationKickPlayer = "you have been kicked from the operation"
)

var NotifQueues map[string]Queue
var MessageQueues map[string]Queue

func (r *MessageQueue) Enqueue(req interface{}) error {
	select {
	case r.Channel <- req.(*rpcpb.Message):
		return nil
	default:
		return fmt.Errorf("queue is full - max capacity of 10")
	}
}

func (r *MessageQueue) Dequeue() interface{} {
	// Must block, as we wait for a request to queue
	select {
	case req := <-r.Channel:
		return req
	}
}

func (r *MessageQueue) Size() int {
	return len(r.Channel)
}

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
