package parameter

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type paramType interface {
	int | int32 | int64 | float32 | float64 | bool | string
}

var (
	ErrValNotFound      = errors.New("value is not found")
	ErrValNotAllowed    = errors.New("value outside opts - ignoring set")
	ErrWrongStringVal   = errors.New("could not convert paramTypeString to string")
	ErrUnknownParamType = errors.New("unknown parameter type")
	ErrWrongIntVal      = errors.New("received param type that cannot be converted to int")
	ErrWrongFloatVal    = errors.New("received param type that cannot be converted to float")
	ErrWrongBoolVal     = errors.New("received param type that cannot be converted to bool")
	ErrWrongTypeVal     = errors.New("received value with invalid type")
	ErrUnhandledTypeVal = errors.New("unhandled type")
)

// Parameter interface is responsible for wrapping ConcreteParameter struct, exposing all methods required.
type Parameter interface {
	SetValue(any) error
	Value() any
	String() string
	Opts() []string
}

// ConcreteParameter[T paramType] hold the actual concrete value for each parameter created with New constructor.
type ConcreteParameter[T paramType] struct {
	Parameter
	typ  reflect.Kind
	val  T
	opts []T
	m    sync.RWMutex
}

// as we are using getter and setter it make more sense to have constuctor for Parameter so this can be used outside the module more easily
// e.g.

// Parameter constructor, the constructor will automatically create the ConcreteParameter instance base on the value passed on in the params
func New(val any, opt string) (Parameter, error) {
	switch reflect.TypeOf(val).Kind() {
	case reflect.Int:
		return newParameter[int](reflect.Int, val, opt)
	case reflect.Int32:
		return newParameter[int32](reflect.Int32, val, opt)
	case reflect.Int64:
		return newParameter[int64](reflect.Int64, val, opt)
	case reflect.Float32:
		return newParameter[float32](reflect.Float32, val, opt)
	case reflect.Float64:
		return newParameter[float64](reflect.Float64, val, opt)
	case reflect.String:
		return newParameter[string](reflect.String, val, opt)
	case reflect.Bool:
		return newParameter[bool](reflect.Bool, val, opt)
	}

	return nil, ErrUnknownParamType
}

func newParameter[T paramType](typ reflect.Kind, val any, opt string) (*ConcreteParameter[T], error) {
	opts, err := buildOptions[T](typ, opt)
	if err != nil {
		return nil, err
	}

	param := &ConcreteParameter[T]{
		typ:  typ,
		opts: opts,
	}

	err = param.SetValue(val)
	if err != nil {
		return nil, err
	}

	return param, nil
}

// Value setter
func (p *ConcreteParameter[T]) SetValue(val any) error {
	valT, ok := val.(T)
	if !ok {
		valStr, ok := val.(string)
		if !ok {
			return ErrWrongTypeVal
		}

		pVal, err := convertStringToVal[T](p.typ, valStr)
		if err != nil {
			return err
		}

		valT = *pVal
	}

	if len(p.opts) > 0 {
		var isFound bool
		for _, t := range p.opts {
			if t == val {
				isFound = true
				break
			}
		}

		if !isFound {
			return ErrValNotAllowed
		}
	}

	p.m.Lock()
	p.val = valT
	p.m.Unlock()
	return nil
}

// Type getter
func (p *ConcreteParameter[T]) Type() reflect.Kind {
	return p.typ
}

// Value getter
func (p *ConcreteParameter[T]) Value() any {
	p.m.RLock()
	defer p.m.RUnlock()
	return p.val
}

// To String representation
func (p *ConcreteParameter[T]) String() string {
	p.m.RLock()
	defer p.m.RUnlock()
	return fmt.Sprintf("%v", p.val)
}

func (p *ConcreteParameter[T]) Opts() []string {
	var opts []string
	for _, opt := range p.opts {
		opts = append(opts, fmt.Sprintf("%v", opt))
	}
	return opts
}
func convertStringToVal[T paramType](typ reflect.Kind, val string) (*T, error) {
	switch typ {
	case reflect.Int:
		if intVal, err := strconv.Atoi(val); err == nil {
			return interface{}(&intVal).(*T), nil
		} else {
			return nil, ErrWrongIntVal
		}
	case reflect.Int32:
		if intVal, err := strconv.ParseInt(val, 10, 32); err == nil {
			int32Val := int32(intVal)
			return interface{}(&int32Val).(*T), nil
		} else {
			return nil, ErrWrongIntVal
		}
	case reflect.Int64:
		if intVal, err := strconv.ParseInt(val, 10, 64); err == nil {
			return interface{}(&intVal).(*T), nil
		} else {
			return nil, ErrWrongIntVal
		}
	case reflect.Float32:
		if floatVal, err := strconv.ParseFloat(val, 32); err == nil {
			float32Val := float32(floatVal)
			return interface{}(&float32Val).(*T), nil
		} else {
			return nil, ErrWrongFloatVal
		}
	case reflect.Float64:
		if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
			return interface{}(&floatVal).(*T), nil
		} else {
			return nil, ErrWrongFloatVal
		}
	case reflect.Bool:
		var boolVal bool
		switch val {
		case "true":
			boolVal = true
		case "false":
			boolVal = false
		default:
			return nil, ErrWrongBoolVal
		}

		return interface{}(&boolVal).(*T), nil
	default:
		return nil, ErrUnhandledTypeVal
	}
}

func buildOptions[T paramType](typ reflect.Kind, opt string) ([]T, error) {
	opts := []T{}
	if opt != "" {
		splits := strings.Split(opt, "|")
		switch typ {
		case reflect.Int:
			for _, val := range splits {
				if intVal, err := strconv.Atoi(val); err == nil {
					opts = append(opts, interface{}(intVal).(T))
				} else {
					return nil, ErrWrongIntVal
				}
			}
		case reflect.Int32:
			for _, val := range splits {
				if intVal, err := strconv.ParseInt(val, 10, 32); err == nil {
					opts = append(opts, interface{}(int32(intVal)).(T))
				} else {
					return nil, ErrWrongIntVal
				}
			}
		case reflect.Int64:
			for _, val := range splits {
				if intVal, err := strconv.ParseInt(val, 10, 64); err == nil {
					opts = append(opts, interface{}(intVal).(T))
				} else {
					return nil, ErrWrongIntVal
				}
			}
		case reflect.Float32:
			for _, val := range splits {
				if floatVal, err := strconv.ParseFloat(val, 32); err == nil {
					opts = append(opts, interface{}(float32(floatVal)).(T))
				} else {
					return nil, ErrWrongFloatVal
				}
			}
		case reflect.Float64:
			for _, val := range splits {
				if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
					opts = append(opts, interface{}(floatVal).(T))
				} else {
					return nil, ErrWrongFloatVal
				}
			}
		case reflect.String:
			for _, val := range splits {
				opts = append(opts, interface{}(val).(T))
			}
		case reflect.Bool:
			for _, valStr := range splits {
				var val bool
				switch valStr {
				case "true":
					val = true
				case "false":
					val = false
				}
				opts = append(opts, interface{}(val).(T))
			}
		default:
			return opts, ErrUnhandledTypeVal
		}
	}

	return opts, nil
}
