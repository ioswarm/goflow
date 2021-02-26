package goflow

import (
	"io"
	"sync"
	"sync/atomic"
)

type Producer interface {
	Subscribe(Outlet)
	OnPull()
	OnCancel()
}

type Source interface {
	OnInit()
	OnPull() (interface{}, error)
	OnClose()
}

type SourceFunc func() (interface{}, error)

func (sf SourceFunc) OnInit() {}
func (sf SourceFunc) OnPull() (interface{}, error) {
	return sf()
}
func (sf SourceFunc) OnClose() {}

/* =================== */

func NewProducer(src Source) Graph {
	return NewGraph(&producer{
		source: src,
	})
}

func NewProducerFunc(f func() (interface{}, error)) Graph {
	return NewProducer(SourceFunc(f))
}

type producer struct {
	sync.Mutex
	outlet Outlet
	source Source
}

func (p *producer) Subscribe(outlet Outlet) {
	p.Lock()
	defer p.Unlock()
	p.outlet = outlet
	p.source.OnInit()
}

func (p *producer) OnPull() {
	data, err := p.source.OnPull()
	if err != nil {
		if err == io.EOF {
			p.outlet.Complete()
			return
		}
		p.outlet.Error(err)
		return
	}
	p.outlet.Push(data)
}

func (p *producer) OnCancel() {
	defer p.outlet.Complete()
	p.source.OnClose()
}

/* =================== */

func Sequence() Graph {
	return NewSequence(0, 1)
}

func NewSequence(start uint64, step uint64) Graph {
	return NewProducer(&sequence{
		seq:  start,
		step: step,
	})
}

type sequence struct {
	seq  uint64
	step uint64
}

func (s *sequence) OnInit() {}

func (s *sequence) OnPull() (interface{}, error) {
	defer atomic.AddUint64(&s.seq, s.step)
	return atomic.LoadUint64(&s.seq), nil
}

func (s *sequence) OnClose() {}

/* =================== */

func ChanProducer(in <-chan interface{}) Graph {
	return NewProducerFunc(func() (interface{}, error) {
		return <-in, nil
	})
}

/* =================== */

func BehaviorProducer(ref Ref, ackCmd interface{}) Graph {
	return NewBehaviorProducer(ref, nil, ackCmd, nil)
}

func NewBehaviorProducer(ref Ref, initCmd interface{}, ackCmd interface{}, completeCmd interface{}) Graph {
	return NewProducer(&behaviorSource{
		ref:         ref,
		initCmd:     initCmd,
		ackCmd:      ackCmd,
		completeCmd: completeCmd,
	})
}

type behaviorSource struct {
	ref         Ref
	initCmd     interface{}
	ackCmd      interface{}
	completeCmd interface{}
}

func (bp *behaviorSource) OnInit() {
	if bp.initCmd != nil {
		bp.ref.Send(bp.initCmd)
	}
}

func (bp *behaviorSource) OnPull() (interface{}, error) {
	data, err := bp.ref.Request(bp.ackCmd)
	if err != nil {
		return nil, err
	}
	if bp.completeCmd != nil && data == bp.completeCmd {
		return nil, io.EOF
	}
	return data, nil
}

func (bp *behaviorSource) OnClose() {
	bp.ref.Close()
}
