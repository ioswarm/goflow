package goflow

import (
	"fmt"
	"time"
)

var (
	ErrorRouteHandleNotFound *Error = newError("RouteHandler not found", "RouteHandle not found in registrated RouteHandler list", "GF-0101")
)

func newError(message string, desc string, code string) *Error {
	return &Error{
		Message:     message,
		Description: desc,
		Code:        code,
		Meta:        make(map[string]interface{}),
		Timestamp:   time.Now().UTC(),
	}
}

func NewError(message string) *Error {
	return newError(message, "", "")
}

func Errorf(format string, a ...interface{}) *Error {
	return NewError(fmt.Sprintf(format, a...))
}

func Errorln(a ...interface{}) *Error {
	return NewError(fmt.Sprintln(a...))
}

type Error struct {
	Message     string                 `json:"error"`
	Description string                 `json:"error_description,omitempty"`
	Code        string                 `json:"error_code,omitempty"`
	Timestamp   time.Time              `json:"timestamp,omitempty"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) WithCode(code string) *Error {
	e.Code = code
	return e
}

func (e *Error) WithDescription(desc string) *Error {
	e.Description = desc
	return e
}

func (e *Error) AddMeta(name string, value interface{}) *Error {
	if e.Meta == nil {
		e.Meta = make(map[string]interface{})
	}
	e.Meta[name] = value
	return e
}

func (e *Error) GetMeta(name string) (interface{}, bool) {
	if e.Meta != nil {
		value, ok := e.Meta[name]
		return value, ok
	}
	return nil, false
}
