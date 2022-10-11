package logger

import (
	"fmt"
	"log"
	"sync"
)

type Proxy interface {
	Put(event Event)
	SetErrConsumer(func(error))
}

type proxy struct {
	mapper      MapperFunc
	errConsumer func(error)
}

func NewProxy(mapper MapperFunc) Proxy {
	return &proxy{
		mapper: mapper,
	}
}

func (p *proxy) SetErrConsumer(errConsumer func(error)) { p.errConsumer = errConsumer }

func (p *proxy) consumeErr(err error) {
	if p.errConsumer != nil {
		p.errConsumer(err)
	}
}

func (p *proxy) Put(ev Event) {
	if _, err := p.mapper.Call(ev); err != nil {
		p.consumeErr(err)
	}
}

// LogLevelFilter ignores an event with the lower level.
func LogLevelFilter(level Level) MapperFunc {
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
			MustNewMapperFunc(LogLevelFilter(level)).Next(LogLevelToPrefixMapper).Next(StandardLogConsumer),
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
		MustNewMapperFunc(g.logLevelFilter).Next(LogLevelToPrefixMapper).Next(StandardLogConsumer),
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
