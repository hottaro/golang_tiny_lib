package htevent

import (
	"fmt"
	"reflect"
	"sync"
)

// HTEvent is an event.
type HTEvent interface {
	Handler(args ...interface{}) error
	// f is a function
	On(f interface{}) error
	Off(f interface{}) error
}

type event struct {
	// listeners are listener functions.
	listeners []reflect.Value
	lmu       sync.RWMutex

	argTypes []reflect.Type
	tmu      sync.RWMutex
}

// New creates a new event.
func New() HTEvent {
	return &event{}
}

var _ HTEvent = New()

func (p *event) Handler(args ...interface{}) error {

	arguments := make([]reflect.Value, 0, len(args))
	argTypes := make([]reflect.Type, 0, len(args))
	for _, v := range args {
		arguments = append(arguments, reflect.ValueOf(v))
		argTypes = append(argTypes, reflect.TypeOf(v))
	}

	err := p.validateArgs(argTypes)
	if err != nil {
		return err
	}

	p.lmu.RLock()
	defer p.lmu.RUnlock()

	wg := sync.WaitGroup{}
	wg.Add(len(p.listeners))
	for _, fn := range p.listeners {
		go func(f reflect.Value) {
			defer wg.Done()
			f.Call(arguments)
		}(fn)
	}

	wg.Wait()
	return nil
}

// Start to listen an event.
func (p *event) On(f interface{}) error {
	fn, err := p.checkFuncSignature(f)
	if err != nil {
		return err
	}

	p.lmu.Lock()
	defer p.lmu.Unlock()
	p.listeners = append(p.listeners, *fn)

	return nil
}

// Stop listening an event.
func (p *event) Off(f interface{}) error {
	fn := reflect.ValueOf(f)

	p.lmu.Lock()
	defer p.lmu.Unlock()
	l := len(p.listeners)
	m := l // for error check
	for i := 0; i < l; i++ {
		if fn == p.listeners[i] {
			// XXX: GC Ref: http://jxck.hatenablog.com/entry/golang-slice-internals
			p.listeners = append(p.listeners[:i], p.listeners[i+1:]...)
			l--
			i--
		}
	}

	if l == m {
		return fmt.Errorf("Listener does't exists")
	}
	return nil
}

// retrun function as reflect.Value
// retrun error if f isn't function, argument is invalid
func (p *event) checkFuncSignature(f interface{}) (*reflect.Value, error) {
	fn := reflect.ValueOf(f)
	if fn.Kind() != reflect.Func {
		return nil, fmt.Errorf("Argument should be a function")
	}

	types := fnArgTypes(fn)

	p.lmu.RLock()
	defer p.lmu.RUnlock()
	if len(p.listeners) == 0 {
		p.tmu.Lock()
		defer p.tmu.Unlock()
		p.argTypes = types
		return &fn, nil
	}

	err := p.validateArgs(types)
	if err != nil {
		return nil, err
	}

	return &fn, nil
}

// if argument size or type are different return error.
func (p *event) validateArgs(types []reflect.Type) error {
	p.tmu.RLock()
	defer p.tmu.RUnlock()
	if len(types) != len(p.argTypes) {
		return fmt.Errorf("Argument length expected %d, but got %d", len(p.argTypes), len(types))
	}
	for i, t := range types {
		if t != p.argTypes[i] {
			return fmt.Errorf("Argument Error. Args[%d] expected %s, but got %s", i, p.argTypes[i], t)
		}
	}

	return nil
}

// return argument types.
func fnArgTypes(fn reflect.Value) []reflect.Type {
	fnType := fn.Type()
	fnNum := fnType.NumIn()

	types := make([]reflect.Type, 0, fnNum)

	for i := 0; i < fnNum; i++ {
		types = append(types, fnType.In(i))
	}

	return types
}
