package goflow

import (
	"context"
	"sync"
)

type Behavior interface {
	Handle(interface{}) (interface{}, error)
}

type BehaviorFunc func(interface{}) (interface{}, error)

func (f BehaviorFunc) Handle(v interface{}) (interface{}, error) {
	return f(v)
}

type Ref interface {
	IN() chan<- interface{}
	OUT() <-chan interface{}
	Close()
	RequestChan(interface{}) <-chan interface{}
	Request(interface{}) (interface{}, error)
	RequestContext(context.Context, interface{}) (interface{}, error)
	Send(interface{})
}

func newRef(behaviorChan chan<- interface{}) Ref {
	r := &ref{
		inChan:       make(chan interface{}), // TODO configure size
		closeChan:    make(chan bool),
		behaviorChan: behaviorChan,
	}
	r.run()
	return r
}

type ref struct {
	sync.Mutex
	inChan       chan interface{}
	outChan      chan interface{}
	closeChan    chan bool
	behaviorChan chan<- interface{}
}

func (r *ref) run() {
	go func() {
		for {
			select {
			case <-r.closeChan:
				close(r.outChan)
				return
			case cmd := <-r.inChan:
				switch cmd.(type) {
				case StopCommand:
					r.behaviorChan <- cmd
				case CallCommand:
					cincmd := cmd.(CallCommand)
					ccmd := Call(cincmd.Data)
					r.behaviorChan <- ccmd
					result := <-ccmd.Result()
					cincmd.Reply(result)
					if r.outChan != nil {
						r.outChan <- result
					}
				default:
					ccmd := Call(cmd)
					r.behaviorChan <- ccmd
					if r.outChan != nil {
						r.outChan <- <-ccmd.Result()
					}
				}
			}
		}
	}()
}

func (r *ref) IN() chan<- interface{} {
	return r.inChan
}

func (r *ref) OUT() <-chan interface{} {
	if r.outChan == nil {
		r.Lock()
		defer r.Unlock()
		r.outChan = make(chan interface{}) // TODO configure size
	}
	return r.outChan
}

func (r *ref) Close() {
	r.closeChan <- true
	close(r.closeChan)
}

func (r *ref) RequestChan(v interface{}) <-chan interface{} {
	ccmd := Call(v)
	r.inChan <- ccmd
	return ccmd.Result()
}

func (r *ref) Request(v interface{}) (interface{}, error) {
	return r.RequestContext(context.Background(), v)
}

func (r *ref) RequestContext(ctx context.Context, v interface{}) (interface{}, error) {
	result := r.RequestChan(v)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-result:
		if err, ok := res.(error); ok {
			return nil, err
		}
		return res, nil
	}
}

func (r *ref) Send(v interface{}) {
	go func() {
		r.RequestChan(v)
	}()
}

type BehaviorHandler interface {
	Ref() Ref
}

func NewBehaviorHandler(behavior Behavior) BehaviorHandler {
	bh := &behaviorHandler{
		channel:  make(chan interface{}), // TODO configure size
		behavior: behavior,
		poolSize: 1, // TODO configure poolSize
	}
	bh.run()
	return bh
}

type behaviorHandler struct {
	channel  chan interface{}
	behavior Behavior
	poolSize int
	jobs     chan CallCommand
}

func worker(num int, jobs <-chan CallCommand, behavior Behavior) {
	for call := range jobs {
		res, err := behavior.Handle(call.Data)
		if err != nil {
			call.Reply(err)
			continue
		}
		call.Reply(res)
	}
}

func (b *behaviorHandler) run() {
	b.jobs = make(chan CallCommand, b.poolSize)
	for i := 0; i < b.poolSize; i++ {
		go worker(i, b.jobs, b.behavior)
	}
	go func() {
		for {
			cmd, open := <-b.channel
			if !open {
				return
			}
			switch cmd.(type) {
			case StopCommand:
				close(b.jobs)
				return
			case CallCommand:
				call := cmd.(CallCommand)
				b.jobs <- call
			}
		}
	}()
}

func (b *behaviorHandler) Ref() Ref {
	return newRef(b.channel)
}
