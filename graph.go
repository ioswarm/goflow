package goflow

type Runnable interface {
	Run() <-chan interface{}
	Close()
}

type Graph interface {
	Take(uint64) Graph
	Map(interface{}) Graph
	Filter(interface{}) Graph

	Via(Flow) Graph
	To(RunnableConsumer) Runnable
}

/* =================== */

type ConsumerGraph interface {
	Graph
	WithProducer(Producer)
}

/* =================== */

type FlowGraph interface {
	ConsumerGraph
	WithConsumer(Consumer)
}

/*
type flowGraph struct {
	sync.Mutex
	inPipe  pipe
	outPipe pipe
}

func (fg *flowGraph) WithProducer(Producer) {
	fg.Lock()
	defer fg.Unlock()

}

func (fg *flowGraph) WithConsumer(Consumer) {
	fg.Lock()
	defer fg.Unlock()

}

func (fg *flowGraph) Via(flow Flow) Graph {

	return fg
}

func (fg *flowGraph) To(consumer RunnableConsumer) Runnable {

	return nil
}
*/
