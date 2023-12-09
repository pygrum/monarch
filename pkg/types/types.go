package types

type Queue interface {
	Enqueue(interface{}) error
	Dequeue() interface{}
	Size() int
}
