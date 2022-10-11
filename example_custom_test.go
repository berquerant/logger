package logger_test

import (
	"errors"
	"fmt"

	"github.com/berquerant/logger"
)

func ExampleMapper_Via() {
	levelToPrefix := func(ev logger.Event) logger.Event {
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
		)
	}
	errGotError := errors.New("GotError")
	generateError1 := func(ev logger.Event) error {
		if ev.Level() != logger.Lerror {
			return nil
		}
		return fmt.Errorf("1 %w: %s", errGotError, ev.Format())
	}
	generateError2 := func(ev logger.Event) error {
		if ev.Level() != logger.Lerror {
			return nil
		}
		return fmt.Errorf("2 %w: %s", errGotError, ev.Format())
	}
	consume1 := func(ev logger.Event) {
		fmt.Printf("Consume1: %s\n", ev)
	}
	consume2 := func(ev logger.Event) {
		fmt.Printf("Consume2: %s\n", ev)
	}

	l := &logger.Logger{
		Proxy: logger.NewProxy(
			logger.MustNewMapperFunc(levelToPrefix).
				Via(generateError1). // Via ignores an error
				Next(consume1).
				Next(generateError2). // Next catches an error and cancels the chain
				Next(consume2),
		),
	}
	l.SetErrConsumer(func(err error) {
		fmt.Printf("Err: %v\n", err)
	})
	l.Info("info msg")
	l.Error("error msg")
	// Output:
	// Consume1: INFO info msg
	// Consume2: INFO info msg
	// Consume1: ERROR error msg
	// Err: 2 GotError: ERROR error msg
}

func ExampleMapper_Next() {
	levelToPrefix := func(ev logger.Event) logger.Event {
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
		)
	}
	format := func(ev logger.Event) logger.Event {
		return logger.NewEvent(ev.Level(), fmt.Sprintf(ev.Format(), ev.Args()...), nil)
	}
	consumeInfo := func(ev logger.Event) logger.Event {
		if ev.Level() == logger.Linfo {
			fmt.Printf("ConsumeInfo: %s\n", ev.Format())
		}
		return ev
	}
	consumeAll := func(ev logger.Event) logger.Event {
		fmt.Println(ev.Format())
		return ev
	}

	l := &logger.Logger{
		Proxy: logger.NewProxy(
			logger.MustNewMapperFunc(levelToPrefix).Next(format).Next(consumeInfo).Next(consumeAll),
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
