package convert

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

type key struct{ from, to reflect.Type }

func (k key) unit() bool {
	return k.to.Elem() == k.from
}

// Registry is a place where you register your simple conversion functions
type Registry struct {
	mu         sync.RWMutex
	converters map[key]Interface
}

func errorIdx(t reflect.Type) int {
	for i := 0; i < t.NumOut(); i++ {
		t.Out(i).AssignableTo(reflect.TypeOf((*error)(nil)).Elem())
		return i
	}
	return -1
}

func (r *Registry) Register(fn interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New("only functions with signature func(in Type, out *TypeOut[,i convert.Interface]) [error] accepted")
		}
	}()
	val := reflect.ValueOf(fn)
	typ := val.Type()
	usesIface := typ.NumIn() > 2
	errorIdx := errorIdx(typ)
	k := key{typ.In(0), typ.In(1)}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.converters == nil {
		r.converters = make(map[key]Interface)
	}

	r.converters[k] = Function(func(in, out interface{}, i Interface) (err error) {
		defer func() {
			if e := recover(); e != nil {
				err = fmt.Errorf("converter panic: %v", e)
			}
		}()

		var inVals []reflect.Value
		inVals = append(inVals, reflect.ValueOf(in))
		inVals = append(inVals, reflect.ValueOf(out))
		if usesIface {
			inVals = append(inVals, reflect.ValueOf(i))
		}

		outVals := val.Call(inVals)
		if errorIdx >= 0 {
			if !outVals[errorIdx].IsNil() {
				return outVals[errorIdx].Interface().(error)
			}
		}
		return nil
	})
	return nil
}

func assign(in, out interface{}, i Interface) error {
	a := reflect.ValueOf(in)
	b := reflect.ValueOf(out)
	b.Elem().Set(a)
	return nil
}
func (r *Registry) find(k key) Interface {
	if k.unit() {
		return Function(assign)
	}
	i, ok := r.converters[k]
	if !ok {
		if k.from.Kind() == reflect.Slice &&
			k.to.Kind() == reflect.Ptr &&
			k.to.Elem().Kind() == reflect.Slice {
			elkey := key{k.from.Elem(), reflect.PtrTo(k.to.Elem().Elem())}
			elConvert := r.find(elkey)
			return appendAll(elConvert)
		}
	}
	return i
}
func (r *Registry) Convert(in, out interface{}) error {
	k := key{reflect.TypeOf(in), reflect.TypeOf(out)}
	i := r.find(k)
	if i == nil {
		return errors.New("conversion not possible")
	}
	return i.Convert(in, out, Function(func(in, out interface{}, i Interface) error {
		r.mu.RLock()
		defer r.mu.RUnlock()
		k := key{reflect.TypeOf(in), reflect.TypeOf(out)}
		ifce := r.find(k)
		if ifce == nil {
			return errors.New("conversion not possible")
		}

		return ifce.Convert(in, out, i)
	}))
}

func appendAll(orig Interface) Interface {
	return Function(func(in, out interface{}, i Interface) error {
		inVal := reflect.ValueOf(in)
		outValPtr := reflect.ValueOf(out)
		outVal := outValPtr.Elem()
		outVal.Set(reflect.MakeSlice(outVal.Type(), 0, 0))
		for x := 0; x < inVal.Len(); x++ {
			s := reflect.New(outVal.Type().Elem())
			if err := orig.Convert(inVal.Index(x).Interface(), s.Interface(), i); err != nil {
				return err
			}
			outVal.Set(reflect.Append(outVal, s.Elem()))
		}

		return nil
	})
}
