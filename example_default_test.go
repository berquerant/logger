package logger_test

import (
	"log"
	"os"

	"github.com/berquerant/logger"
)

func ExampleNewDefault() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)
	l := logger.NewDefault(logger.Lwarn)
	l.Info("information")
	l.Error("error")
	l.Debug("debug")
	l.Warn("warn")
	l.Debug("last line")
	// Output:
	// E | error
	// W | warn
}
