package goflow

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	sys = newSystem()
)

func System() FlowSystem { return sys }

type FlowSystem interface {
	Logger
	Terminate()
	TerminateWithTimeout(time.Duration)
	Terminated() <-chan int
}

func newSystem() FlowSystem {
	sys := &system{
		exitChan: make(chan int, 1),
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		signal := <-sigs
		log.Printf("GoFlow ends by %v", signal)
		sys.Terminate()
	}()

	return sys
}

type system struct {
	exitChan chan int
}

func (sys *system) Terminate() {
	sys.TerminateWithTimeout(10 * time.Second) // TODO configure sys termination timeout
}

func (sys *system) TerminateWithTimeout(time.Duration) {
	// TODO close all registered behaviors
	sys.exitChan <- 0
}

func (sys *system) Terminated() <-chan int {
	return sys.exitChan
}

func (sys *system) DEBUG(msg string, a ...interface{}) {

}

func (sys *system) INFO(msg string, a ...interface{}) {

}

func (sys *system) WARN(msg string, a ...interface{}) {

}

func (sys *system) ERROR(msg string, a ...interface{}) {

}

func (sys *system) FATAL(msg string, a ...interface{}) {

}

func (sys *system) PANIC(msg string, a ...interface{}) {

}
