package goflow

type StopCommand struct{}

func Stop() StopCommand { return StopCommand{} }

type NoneCommand struct{}

func None() NoneCommand { return NoneCommand{} }

type DoneCommand struct{}

func Done() DoneCommand { return DoneCommand{} }

type CallCommand struct {
	Data   interface{}
	result chan interface{}
}

func Call(data interface{}) CallCommand {
	return CallCommand{
		Data:   data,
		result: make(chan interface{}, 1),
	}
}

func (cc *CallCommand) Reply(v interface{}) {
	cc.result <- v
	close(cc.result)
}

func (cc *CallCommand) Result() <-chan interface{} {
	return cc.result
}

type PullCommand struct{}

func Pull() PullCommand { return PullCommand{} }

type CancelCommand struct{}

func Cancel() CancelCommand { return CancelCommand{} }

type PushCommand struct {
	Data interface{}
}

func Push(data interface{}) PushCommand { return PushCommand{data} }

type ErrorCommand struct {
	Error error
}

func ErrorCmd(err error) ErrorCommand { return ErrorCommand{err} }

type CompleteCommand struct{}

func Complete() CompleteCommand { return CompleteCommand{} }
