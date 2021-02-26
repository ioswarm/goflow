package goflow

type Parameters map[string]interface{}

func (p Parameters) GetParameter(name string) (Parameter, bool) {
	if value, ok := p[name]; ok {
		return NewParameter(name, value), ok
	}
	return NewParameter(name, nil), false
}

func (p Parameters) AddParameter(param Parameter) {
	p[param.Name] = param.Value
}

func NewParamerters(params ...Parameter) Parameters {
	result := make(Parameters)
	for _, p := range params {
		result.AddParameter(p)
	}
	return result
}

type Parameter struct {
	Name  string
	Value interface{}
}

func NewParameter(name string, value interface{}) Parameter {
	return Parameter{
		Name:  name,
		Value: value,
	}
}
