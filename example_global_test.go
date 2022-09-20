package logger_test

import (
	"log"
	"os"

	"github.com/berquerant/logger"
)

func ExampleG() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)
	logger.G().Info("information")
	logger.G().Error("error")
	logger.G().Debug("debug")
	logger.G().SetLevel(logger.Ldebug)
	logger.G().Info("change level")
	logger.G().Debug("last line")
	// Output:
	// I | information
	// E | error
	// I | change level
	// D | last line
}
