package goflow

import (
	"context"
	"io"
	"reflect"
)

var (
	errorType   = reflect.TypeOf((*error)(nil)).Elem()
	contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
)

func isErrorType(gtype reflect.Type) bool {
	return gtype.Implements(errorType)
}

func isContextType(gtype reflect.Type) bool {
	return gtype.Implements(contextType)
}

func checkImplements(atype reflect.Type, btype reflect.Type) bool {
	if btype.Kind() == reflect.Interface {
		return atype.Implements(btype)
	} else if atype.Kind() == reflect.Interface {
		return btype.Implements(atype)
	}
	return false
}

func compareType(atype reflect.Type, btype reflect.Type) bool {
	return atype == btype || checkImplements(atype, btype)
}

func variadicMapFunc(varf interface{}) func(interface{}) (interface{}, error) {
	varfType := reflect.TypeOf(varf)
	if varfType.Kind() != reflect.Func {
		panic("Given parameter must be a func")
	}
	if varfType.NumIn() != 1 {
		panic("Given function must accept one parameter")
	}
	if varfType.NumOut() > 2 || (varfType.NumOut() == 2 && !compareType(varfType.Out(1), errorType)) {
		panic("Given function must return at maximum two result, where the secondone must of type error")
	}

	inType := varfType.In(0)
	varfValue := reflect.ValueOf(varf)

	return func(v interface{}) (interface{}, error) {
		vType := reflect.TypeOf(v)
		if !compareType(vType, inType) {
			return nil, Errorf("Could not cast %v to %v", vType, inType)
		}

		result := varfValue.Call([]reflect.Value{reflect.ValueOf(v)})
		if varfType.NumOut() > 1 {
			res := result[1].Interface()
			if res != nil {
				return nil, res.(error)
			}
		}
		return result[0].Interface(), nil
	}
}

func variadicFilterFunc(varf interface{}) func(interface{}) (interface{}, error) {
	varfType := reflect.TypeOf(varf)
	if varfType.Kind() != reflect.Func {
		panic("Given parameter must be a func")
	}
	if varfType.NumIn() != 1 {
		panic("Given function must accept one parameter")
	}
	if varfType.NumOut() == 1 && varfType.Out(0).Kind() != reflect.Bool {
		panic("Given function must return a bool")
	}

	inType := varfType.In(0)
	varfValue := reflect.ValueOf(varf)

	return func(v interface{}) (interface{}, error) {
		vType := reflect.TypeOf(v)
		if !compareType(vType, inType) {
			return nil, io.EOF
		}

		result := varfValue.Call([]reflect.Value{reflect.ValueOf(v)})
		if !result[0].Bool() {
			return nil, io.EOF
		}
		return v, nil
	}
}
