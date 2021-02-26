package goflow

type Marshal func(interface{}) ([]byte, error)

type Unmarshal func([]byte, interface{}) error

