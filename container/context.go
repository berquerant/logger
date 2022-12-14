package container

import (
	"context"

	"github.com/berquerant/logger"
)

type ctxKeyType string

const ctxKey ctxKeyType = "ctxKeyValue"

func FromContext(ctx context.Context) Context { return ctx.Value(ctxKey).(Context) }

// Context holds a request-scoped data and a logger to partial structured logging.
type Context interface {
	Data() Map[string, any]
	L() *logger.Logger
	Clone() Context
	WithContext(ctx context.Context) context.Context
}

type contextImpl struct {
	data   Map[string, any]
	lgr    *logger.Logger
	mapper logger.MapperFunc // for clone
}

func New(data Map[string, any], mapper logger.MapperFunc) Context {
	return &contextImpl{
		data: data,
		lgr: &logger.Logger{
			Proxy: logger.NewProxy(logger.MustNewMapperFunc(data.StructMapper).Next(mapper)),
		},
		mapper: mapper,
	}
}

func (c *contextImpl) Data() Map[string, any] { return c.data }
func (c *contextImpl) L() *logger.Logger      { return c.lgr }
func (c *contextImpl) Clone() Context         { return New(c.data.Clone(), c.mapper) }
func (c *contextImpl) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey, c)
}
