package htevent

import "fmt"

// HTEventNotDefined is an error indicationg that the event has not been defined.
type HTEventNotDefined struct {
	name string
}

func newHTEventNotDefined(name string) *HTEventNotDefined {
	return &HTEventNotDefined{
		name: name,
	}
}

func (e *HTEventNotDefined) Error() string {
	return fmt.Sprintf("%s event has not been defined yet.", e.name)
}

// HTEventName return name of the event.
func (e *HTEventNotDefined) HTEventName() string {
	return e.name
}

var _ error = newHTEventNotDefined("none f")

