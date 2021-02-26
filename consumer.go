package goflow

import (
	"fmt"
	"sync"
)

type Consumer interface {
	OnSubscribe(Inlet)
	OnPush(interface{})
	OnError(error)
	OnComplete()
}

type Receiver interface {
	OnInit()
	OnPush(interface{})
	OnError(error)
	OnComplete() interface{}
}

type ReceiverFunc func(interface{})

func (ReceiverFunc) OnInit() {}
func (rf ReceiverFunc) OnPush(v interface{}) {
	rf(v)
}
func (ReceiverFunc) OnError(error)           {}
func (ReceiverFunc) OnComplete() interface{} { return Done() }

type RunnableConsumer interface {
	Consumer
	Runnable
}

/* =================== */

func Ignore() RunnableConsumer {
	return ForEach(func(interface{}) {})
}

/* =================== */

func Println() RunnableConsumer {
	return ForEach(func(v interface{}) {
		fmt.Printf("%v\n", v)
	})
}

/* =================== */

func ForEach(f func(interface{})) RunnableConsumer {
	return NewConsumerFunc(f)
}

/* =================== */

func Slice() RunnableConsumer {
	return NewConsumer(&slice{
		result: make([]interface{}, 0),
	})
}

type slice struct {
	sync.RWMutex
	result []interface{}
}

func (s *slice) OnInit() {}

func (s *slice) OnPush(v interface{}) {
	s.Lock()
	defer s.Unlock()
	s.result = append(s.result, v)
}

func (s *slice) OnError(error) {
	// TOOD handle errors
}

func (s *slice) OnComplete() interface{} {
	s.RLock()
	defer s.RUnlock()
	return s.result
}

/* =================== */

func NewConsumer(receiver Receiver) RunnableConsumer {
	return &consumer{
		result:   make(chan interface{}, 1),
		receiver: receiver,
	}
}

func NewConsumerFunc(f func(interface{})) RunnableConsumer {
	return NewConsumer(ReceiverFunc(f))
}

type consumer struct {
	sync.Mutex
	inlet    Inlet
	result   chan interface{}
	receiver Receiver
}

func (c *consumer) OnSubscribe(inlet Inlet) {
	c.Lock()
	defer c.Unlock()
	c.inlet = inlet
	c.receiver.OnInit()
}

func (c *consumer) OnPush(v interface{}) {
	c.receiver.OnPush(v)
	c.inlet.Pull()
}
func (c *consumer) OnError(err error) {
	c.receiver.OnError(err)
}

func (c *consumer) OnComplete() {
	c.result <- c.receiver.OnComplete()
	close(c.result)
}

func (c *consumer) Run() <-chan interface{} {
	// TODO is valid and everything is set
	c.inlet.Pull()
	return c.result
}

func (c *consumer) Close() {
	c.inlet.Cancel()
	//c.OnComplete()
}
