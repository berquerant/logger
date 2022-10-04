package logger_test

import (
	"fmt"

	"github.com/berquerant/logger"
)

func ExampleLogger() {
	levelToPrefix := func(ev logger.Event) (logger.Event, error) {
		var p string
		switch ev.Level() {
		case logger.Linfo:
			p = "INFO"
		case logger.Lwarn:
			p = "WARN"
		case logger.Lerror:
			p = "ERROR"
		default:
			p = "?"
		}
		return logger.NewEvent(
			ev.Level(),
			fmt.Sprintf("%s %s", p, ev.Format()),
			ev.Args(),
		), nil
	}
	format := func(ev logger.Event) (logger.Event, error) {
		return logger.NewEvent(ev.Level(), fmt.Sprintf(ev.Format(), ev.Args()...), nil), nil
	}
	consumeInfo := func(ev logger.Event) (logger.Event, error) {
		if ev.Level() == logger.Linfo {
			fmt.Printf("ConsumeInfo: %s\n", ev.Format())
		}
		return ev, nil
	}
	consumeAll := func(ev logger.Event) (logger.Event, error) {
		fmt.Println(ev.Format())
		return ev, nil
	}

	l := &logger.Logger{
		Proxy: logger.NewProxy(
			logger.NewMapperList(levelToPrefix, format),
			logger.NewMapperList(consumeInfo),
			logger.NewMapperList(consumeAll),
		),
	}
	l.Info("info msg")
	l.Warn("warn msg")
	l.Error("error msg")
	l.Info("info level value is %d", logger.Linfo)
	// Output:
	// ConsumeInfo: INFO info msg
	// INFO info msg
	// WARN warn msg
	// ERROR error msg
	// ConsumeInfo: INFO info level value is 30
	// INFO info level value is 30
}
