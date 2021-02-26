package goflow

type Graph interface {
	Via(Flow) Graph
	To(RunnableConsumer) Runnable
}

type Runnable interface {
	Run() <-chan interface{}
	Close()
}
