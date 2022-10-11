# logger

A light-weight wrapper of standard `log`.

## Static logger

``` go
logger.G().Info("message")
```

writes like `2022/09/20 10:00:00 I | message` to stderr.

## Logger instance

``` go
l := logger.NewDefault(logger.Lerror)
l.Info("message")
```

## Customized logger

``` go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/berquerant/logger"
)

func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)
	l := &logger.Logger{
		Proxy: logger.NewProxy(
			logger.MustNewMapperFunc(func(ev logger.Event) logger.Event {
				switch ev.Level() {
				case logger.Linfo, logger.Lwarn, logger.Lerror:
					// select info, warn, error
					return ev
				default:
					// ignore other levels
					fmt.Printf("Ignore: %v", ev)
					return nil
				}
			}).Next(func(ev logger.Event) {
				if ev.Level() == logger.Lerror { // consume only error logs
					fmt.Printf("Got an error: %v\n", ev)
				}
			}).Next(logger.StandardLogConsumer),
		),
	}
	l.Info("info msg")
	l.Error("error msg")
	l.Trace("trace msg")
}
// Output:
// info msg
// Got an error: error msg
// error msg
// Ignore: trace msg
```
