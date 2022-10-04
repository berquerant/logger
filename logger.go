package logger

import (
	"fmt"
	"log"
	"sync"
)

// Level is the threshold for logging.
type Level int

//go:generate go run github.com/berquerant/dataclass@latest -type Event -field "Level Level,Format string,Args []any" -output event_generated.go

type (
	// Mapper converts and/or filters the log event.
	Mapper func(Event) (Event, error)
	// ErrConsumer receives an error during mappings.
	ErrConsumer func(error)
)

// MapperList is a set of event conversions.
// In order from the top of the list, applies the mapper to the event.
type MapperList interface {
	// Append appends a Mapper to the end of the list.
	Append(f Mapper)
	// Map applies the mappers to the event.
	// If the mapper returns a nil event or an error then cancels the propagation.
	Map(ev Event) (Event, error)
}

type mapperList struct {
	mappers []Mapper
	mux     sync.RWMutex
}

func NewMapperList(v ...Mapper) MapperList {
	return &mapperList{
		mappers: v,
	}
}

func (ml *mapperList) Map(ev Event) (res Event, err error) {
	ml.mux.RLock()
	defer ml.mux.RUnlock()
	res = ev
	for i, m := range ml.mappers {
		if res == nil {
			return
		}
		if res, err = m(res); err != nil {
			err = fmt.Errorf("Map error idx %d ev %s %w", i, res, err)
			return
		}
	}
	return
}

func (ml *mapperList) Append(f Mapper) {
	ml.mux.Lock()
	defer ml.mux.Unlock()
	ml.mappers = append(ml.mappers, f)
}

// Proxy accepts an event and processes it.
type Proxy interface {
	// Append appends the mapper list to the end of the list.
	Append(list MapperList)
	// At returns the i-th mapper list.
	// Returns false if out of range.
	At(i int) (MapperList, bool)
	// Put starts processing of the event.
	// Apply the first mappers the event, then the second.
	Put(ev Event)
	SetErrConsumer(ErrConsumer)
}

type proxy struct {
	sync.RWMutex
	list        []MapperList
	errConsumer ErrConsumer
}

func NewProxy(list ...MapperList) Proxy {
	return &proxy{
		list: list,
	}
}

func (p *proxy) Append(list MapperList) {
	p.Lock()
	defer p.Unlock()
	p.list = append(p.list, list)
}

func (p *proxy) At(i int) (MapperList, bool) {
	p.RLock()
	defer p.RUnlock()
	if i < 0 || i >= len(p.list) {
		return nil, false
	}
	return p.list[i], true
}

func (p *proxy) SetErrConsumer(errConsumer ErrConsumer) {
	p.Lock()
	defer p.Unlock()
	p.errConsumer = errConsumer
}

func (p *proxy) consumeErr(err error) {
	p.RLock()
	defer p.RUnlock()
	if p.errConsumer != nil {
		p.errConsumer(err)
	}
}

func (p *proxy) Put(ev Event) {
	p.RLock()
	defer p.RUnlock()

	var err error
	for _, ml := range p.list {
		if ev, err = ml.Map(ev); err != nil {
			p.consumeErr(err)
			return
		}
	}
}

// LogLevelFilter ignores an event with the lower level.
func LogLevelFilter(level Level) Mapper {
	return func(ev Event) (Event, error) {
		if ev.Level() <= level {
			return ev, nil
		}
		return nil, nil
	}
}

const (
	Lsilent Level = 0
	Lerror  Level = 10
	Lwarn   Level = 20
	Linfo   Level = 30
	Ldebug  Level = 40
	Ltrace  Level = 50
)

type Logger struct {
	Proxy
}

func (l *Logger) Info(format string, v ...any) {
	l.Put(NewEvent(Linfo, format, v))
}

func (l *Logger) Warn(format string, v ...any) {
	l.Put(NewEvent(Lwarn, format, v))
}

func (l *Logger) Error(format string, v ...any) {
	l.Put(NewEvent(Lerror, format, v))
}

func (l *Logger) Debug(format string, v ...any) {
	l.Put(NewEvent(Ldebug, format, v))
}

func (l *Logger) Trace(format string, v ...any) {
	l.Put(NewEvent(Ltrace, format, v))
}

func logLevelToPrefix(level Level) string {
	switch level {
	case Linfo:
		return "I |"
	case Lwarn:
		return "W |"
	case Lerror:
		return "E |"
	case Ldebug:
		return "D |"
	case Ltrace:
		return "T |"
	default:
		return "? |"
	}
}

// LogLevelToPrefixMapper adds a prefix depending on the event level.
func LogLevelToPrefixMapper(ev Event) (Event, error) {
	return NewEvent(
		ev.Level(),
		fmt.Sprintf("%s %s", logLevelToPrefix(ev.Level()), ev.Format()),
		ev.Args(),
	), nil
}

// StandardLogConsumer writes an event by `log.Printf`.
func StandardLogConsumer(ev Event) (Event, error) {
	log.Printf(ev.Format(), ev.Args()...)
	return ev, nil
}

// NewDefault returns a new logger with `LogLevelFilter`, `LogLevelToPrefixMapper` and `StandardLogConsumer`.
func NewDefault(level Level) *Logger {
	return &Logger{
		Proxy: NewProxy(
			NewMapperList(LogLevelFilter(level), LogLevelToPrefixMapper),
			NewMapperList(StandardLogConsumer),
		),
	}
}

// GlobalLogger is a static logger instance.
// This filters logs by level, adds a prefix depending on event level
// and writes logs by `log.Printf`.
type GlobalLogger interface {
	Info(format string, v ...any)
	Warn(format string, v ...any)
	Error(format string, v ...any)
	Debug(format string, v ...any)
	Trace(format string, v ...any)
	SetLevel(level Level)
	Level() Level
}

type globalLogger struct {
	*Logger
	level Level
}

func (g *globalLogger) logLevelFilter(ev Event) (Event, error) {
	return LogLevelFilter(g.level)(ev)
}

func (g *globalLogger) SetLevel(level Level) { g.level = level }
func (g *globalLogger) Level() Level         { return g.level }

func newGlobalLogger() *globalLogger {
	g := &globalLogger{
		level:  Linfo,
		Logger: &Logger{},
	}
	g.Proxy = NewProxy(
		NewMapperList(g.logLevelFilter, LogLevelToPrefixMapper),
		NewMapperList(StandardLogConsumer),
	)
	return g
}

var (
	globalLoggerInstance GlobalLogger
	globalLoggerInitOnce sync.Once
)

func initGlobalLogger() {
	globalLoggerInstance = newGlobalLogger()
}

// G returns the `GlobalLogger`.
func G() GlobalLogger {
	globalLoggerInitOnce.Do(initGlobalLogger)
	return globalLoggerInstance
}
