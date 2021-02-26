package goflow

import "sync"

type Outlet interface {
	Push(interface{})
	Error(error)
	Complete()
}

type Inlet interface {
	Pull()
	Cancel()
}

type Pipe interface {
	Inlet
	Outlet
	Graph
}

func NewGraph(producer Producer) Graph {
	p := &pipe{
		producer: producer,
		inbound:  make(chan interface{}),
		outbound: make(chan interface{}),
	}
	producer.Subscribe(p)
	return p
}

type pipe struct {
	sync.Mutex
	consumer Consumer
	producer Producer
	inbound  chan interface{}
	outbound chan interface{}
}

func (p *pipe) run() {
	go func() {
		for {
			select {
			case incmd := <-p.inbound:
				switch incmd.(type) {
				case PushCommand:
					pc := incmd.(PushCommand)
					go func() {
						p.consumer.OnPush(pc.Data)
					}()
				case ErrorCommand:
					ec := incmd.(ErrorCommand)
					go func() {
						p.consumer.OnError(ec.Error)
					}()
				case CompleteCommand:
					go func() {
						p.consumer.OnComplete()
					}()
					return
				}
			case outcmd := <-p.outbound:
				switch outcmd.(type) {
				case PullCommand:
					go func() {
						p.producer.OnPull()
					}()
				case CancelCommand:
					go func() {
						p.producer.OnCancel()
					}()
				}
			}
		}
	}()
}

func (p *pipe) Push(v interface{}) {
	go func() {
		p.inbound <- Push(v)
	}()
}

func (p *pipe) Error(err error) {
	go func() {
		p.inbound <- ErrorCmd(err)
	}()
}

func (p *pipe) Complete() {
	// blocks till last command is piped
	p.inbound <- Complete()
}

func (p *pipe) Pull() {
	go func() {
		p.outbound <- Pull()
	}()
}

func (p *pipe) Cancel() {
	go func() {
		p.outbound <- Cancel()
	}()
}

func (p *pipe) Take(a uint64) Graph {
	return p.Via(Take(a))
}

func (p *pipe) Map(f interface{}) Graph {
	return p.Via(NewFlowFunc(variadicMapFunc(f)))
}

func (p *pipe) Filter(f interface{}) Graph {
	return p.Via(NewFlowFunc(variadicFilterFunc(f)))
}

func (p *pipe) Via(flow Flow) Graph {
	p.Lock()
	defer p.Unlock()
	p.consumer = flow
	p.consumer.OnSubscribe(p)
	p.run()
	return NewGraph(flow)
}

func (p *pipe) To(consumer RunnableConsumer) Runnable {
	p.Lock()
	defer p.Unlock()
	p.consumer = consumer
	p.consumer.OnSubscribe(p)
	p.run()
	return consumer
}
