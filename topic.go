package goflow

import "sync"

type Topic interface {
	Subscribe(chan<- interface{})
	Unsubscribe(chan<- interface{})
	SubscribeRef(Ref)
	UnsubscribeRef(Ref)

	Publish(interface{})
	Close()
}

func NewTopic() Topic {
	t := &topic{
		inbound:  make(chan interface{}), // TODO configure size
		consumer: make(map[chan<- interface{}]bool),
	}
	t.run()
	return t
}

type topic struct {
	sync.RWMutex
	inbound  chan interface{}
	consumer map[chan<- interface{}]bool
}

func (t *topic) run() {
	go func() {
		for msg := range t.inbound {
			t.RLock()
			for c := range t.consumer {
				c <- msg
			}
			t.RUnlock()
		}
	}()
}

func (t *topic) Subscribe(c chan<- interface{}) {
	t.Lock()
	defer t.Unlock()
	if _, ok := t.consumer[c]; !ok {
		t.consumer[c] = true
	}
}

func (t *topic) Unsubscribe(c chan<- interface{}) {
	t.Lock()
	defer t.Unlock()
	if _, ok := t.consumer[c]; ok {
		delete(t.consumer, c)
	}
}

func (t *topic) SubscribeRef(ref Ref) {
	t.Subscribe(ref.IN())
}

func (t *topic) UnsubscribeRef(ref Ref) {
	t.Unsubscribe(ref.IN())
}

func (t *topic) Publish(v interface{}) {
	go func() {
		t.inbound <- v
	}()
}

func (t *topic) Close() {
	close(t.inbound)
}
