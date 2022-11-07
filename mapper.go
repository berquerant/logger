package logger

import (
	"errors"
	"fmt"
	"reflect"
)

// Level is the threshold for logging.
type Level int

//go:generate go run github.com/berquerant/dataclass@latest -type Event -field "Level Level,Format string,Args []any" -output event_generated.go

func (e *event) String() string { return fmt.Sprintf(e.format, e.args...) }

// MapperFunc converts and/or filters the log event.
type MapperFunc func(Event) (Event, error)

var (
	ErrInvalidMapperFunc = errors.New("InvalidMapperFunc")
	ErrNilMapperFunc     = errors.New("NilMapperFunc")
	ErrNilEvent          = errors.New("NilEvent")
)

func NewMapperFunc(f any) (MapperFunc, error) { return intoMapperFunc(f) }
func MustNewMapperFunc(f any) MapperFunc {
	m, err := intoMapperFunc(f)
	if err != nil {
		panic(err)
	}
	return m
}

func intoMapperFunc(f any) (MapperFunc, error) {
	switch f := f.(type) {
	case MapperFunc:
		return f, nil
	case func(Event) (Event, error):
		return f, nil
	case func(Event):
		return func(ev Event) (Event, error) {
			f(ev)
			return ev, nil
		}, nil
	case func(Event) error:
		return func(ev Event) (Event, error) {
			if err := f(ev); err != nil {
				return nil, err
			}
			return ev, nil
		}, nil
	case func(Event) Event:
		return func(ev Event) (Event, error) { return f(ev), nil }, nil
	default:
		return nil, fmt.Errorf("%w %v", ErrInvalidMapperFunc, reflect.TypeOf(f))
	}
}

// Call the function.
// Return ErrNilMapperFunc if it's nil.
// Return ErrNilEvent if event is nil.
func (m MapperFunc) Call(event Event) (Event, error) {
	if event == nil {
		return nil, ErrNilEvent
	}
	if m == nil {
		return nil, ErrNilMapperFunc
	}
	return m(event)
}

// Next appends a MapperFunc.
// The returned function calls this, and f with the returned event of this, if no errors.
//
// Available signatures of f:
//   func(Event)
//   func(Event) Event
//   func(Event) error
//   func(Event) (Event, error)
// Otherwise f is evaluated as nil MapperFunc.
func (m MapperFunc) Next(f any) MapperFunc {
	mapper, _ := intoMapperFunc(f)
	if m == nil {
		return mapper
	}
	if mapper == nil {
		return m
	}
	return func(event Event) (Event, error) {
		event, err := m.Call(event)
		if err != nil {
			return nil, err
		}
		return mapper.Call(event)
	}
}

// Via appends a MapperFunc.
// The returned function calls this, and f with the returned event of this, if no errors,
// but ignores the values from f.
//
// Available signatures of f:
//   func(Event)
//   func(Event) Event
//   func(Event) error
//   func(Event) (Event, error)
// Otherwise f is evaluated as nil MapperFunc.
func (m MapperFunc) Via(f any) MapperFunc {
	mapper, _ := intoMapperFunc(f)
	if m == nil {
		return mapper
	}
	if mapper == nil {
		return m
	}
	return func(event Event) (Event, error) {
		event, err := m.Call(event)
		if err != nil {
			return nil, err
		}
		_, _ = mapper.Call(event)
		return event, nil
	}
}
