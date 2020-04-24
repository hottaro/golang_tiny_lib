package htevent

import "sync"

// HTDispatcher is an event htDispatcher.
type HTDispatcher interface {
	Handler(name string, args ...interface{}) error
	// f is a function
	On(name string, f interface{}) error
	Off(name string, f interface{}) error
	// Destroy a event
	Destroy(name string) error
}

type htDispatcher struct {
	events map[string]HTEvent
	mu     sync.RWMutex
}

// NewHTDispatcher creates a new event htDispatcher.
func NewHTDispatcher() HTDispatcher {
	return &htDispatcher{
		events: map[string]HTEvent{},
	}
}

func (t *htDispatcher) Handler(name string, args ...interface{}) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	ev, ok := t.events[name]
	if !ok {
		return newHTEventNotDefined(name)
	}

	return ev.Handler(args...)
}

func (t *htDispatcher) On(name string, f interface{}) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	ev, ok := t.events[name]
	if !ok {
		ev = New()
		t.events[name] = ev
	}
	return ev.On(f)
}

func (t *htDispatcher) Off(name string, f interface{}) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	e, ok := t.events[name]
	if !ok {
		return newHTEventNotDefined(name)
	}

	return e.Off(f)
}

func (t *htDispatcher) Destroy(name string) error {
	if _, ok := t.events[name]; !ok {
		return newHTEventNotDefined(name)
	}
	delete(t.events, name)
	return nil
}

var _ HTDispatcher = &htDispatcher{}
