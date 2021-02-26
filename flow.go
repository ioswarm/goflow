package goflow

import (
	"io"
	"sync"
)

type Flow interface {
	Consumer
	Producer
}

type Task interface {
	OnInit()
	OnHandle(interface{}) (interface{}, error)
	OnClose()
}

type TaskFunc func(interface{}) (interface{}, error)

func (TaskFunc) OnInit() {}
func (tf TaskFunc) OnHandle(v interface{}) (interface{}, error) {
	return tf(v)
}
func (TaskFunc) OnClose() {}

/* =================== */

func Map(f func(interface{}) interface{}) Flow {
	return NewFlowFunc(func(v interface{}) (interface{}, error) {
		return f(v), nil
	})
}

/* =================== */

func Filter(f func(interface{}) bool) Flow {
	return NewFlowFunc(func(v interface{}) (interface{}, error) {
		if f(v) {
			return v, nil
		}
		return nil, io.EOF
	})
}

/* =================== */

func NewFlow(task Task) Flow {
	return &flow{
		task: task,
	}
}

func NewFlowFunc(f func(interface{}) (interface{}, error)) Flow {
	return NewFlow(TaskFunc(f))
}

type flow struct {
	sync.Mutex
	inlet  Inlet
	outlet Outlet
	task   Task
}

func (f *flow) OnSubscribe(inlet Inlet) {
	f.Lock()
	defer f.Unlock()
	f.inlet = inlet
}

func (f *flow) OnPush(v interface{}) {
	result, err := f.task.OnHandle(v)
	if err != nil {
		if err == io.EOF {
			f.inlet.Pull()
			return
		}
		f.outlet.Error(err)
	}
	f.outlet.Push(result)
}

func (f *flow) OnError(err error) {
	f.outlet.Error(err)
}

func (f *flow) OnComplete() {
	f.outlet.Complete()
	f.task.OnClose()
}

func (f *flow) Subscribe(outlet Outlet) {
	f.Lock()
	defer f.Unlock()
	f.outlet = outlet
	f.task.OnInit()
}

func (f *flow) OnPull() {
	f.inlet.Pull()
}

func (f *flow) OnCancel() {
	f.inlet.Cancel()
}

/* =================== */

func Take(u uint64) Flow {
	return &take{
		step: 1,
		end:  u,
	}
}

type take struct {
	sync.Mutex
	inlet  Inlet
	outlet Outlet
	pos    uint64
	step   uint64
	end    uint64
}

func (t *take) OnSubscribe(inlet Inlet) {
	t.Lock()
	defer t.Unlock()
	t.inlet = inlet
}

func (t *take) OnPush(v interface{}) {
	t.Lock()
	defer t.Unlock()
	t.pos += t.step
	t.outlet.Push(v)
}

func (t *take) OnError(err error) {
	t.outlet.Error(err)
}

func (t *take) OnComplete() {
	t.outlet.Complete()
}

func (t *take) Subscribe(outlet Outlet) {
	t.Lock()
	defer t.Unlock()
	t.outlet = outlet
}

func (t *take) OnPull() {
	t.Lock()
	defer t.Unlock()
	if t.pos < t.end {
		t.inlet.Pull()
		return
	}
	t.inlet.Cancel()
}

func (t *take) OnCancel() {
	t.inlet.Cancel()
}
