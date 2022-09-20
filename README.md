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
filter := func(ev logger.Event) (logger.Event, error) {
  switch ev.Level() {
  case logger.Linfo, logger.Lwarn, logger.Lerror:
    // select info, warn, error
    return ev, nil
  default:
    // ignore other levels
    return nil. nil
  }
}
consumeError := func(ev logger.Event) (logger.Event, error) {
  if ev.Level() == logger.Lerror {
    // consume event
  }
  return ev, nil
}
l := &logger.Logger{
  Proxy: logger.NewProxy(
    logger.NewMapperList(filter),
    logger.NewMapperList(consumeError, logger.StandardLogConsumer),
  ),
}
```
