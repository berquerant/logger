package logger_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/berquerant/logger"
	"github.com/stretchr/testify/assert"
)

func eventToString(ev logger.Event) string {
	return fmt.Sprintf("%v %s", ev.Level(), fmt.Sprintf(ev.Format(), ev.Args()...))
}

func eventEqual(t *testing.T, want, got logger.Event) {
	assert.Equal(t, eventToString(want), eventToString(got))
}

type mockMapperFunc struct {
	arg logger.Event
	ret logger.Event
	err error
}

func (m *mockMapperFunc) call(event logger.Event) (logger.Event, error) {
	m.arg = event
	return m.ret, m.err
}

func newMockMapperFunc(ret logger.Event, err error) *mockMapperFunc {
	return &mockMapperFunc{
		ret: ret,
		err: err,
	}
}

func TestMapperFuncVia(t *testing.T) {
	var (
		err1     = errors.New("err1")
		ev1      = logger.NewEvent(10, "msg", nil)
		ev2      = logger.NewEvent(11, "msg", nil)
		ev3      = logger.NewEvent(12, "msg", nil)
		newF1    = func() *mockMapperFunc { return newMockMapperFunc(ev2, nil) }
		newF2    = func() *mockMapperFunc { return newMockMapperFunc(ev3, nil) }
		newF1Err = func() *mockMapperFunc { return newMockMapperFunc(nil, err1) }
	)

	t.Run("ignore error", func(t *testing.T) {
		var (
			f1 = newF1()
			f2 = newF1Err()
			f3 = newF2()
		)
		got, err := logger.MustNewMapperFunc(f1.call).Via(f2.call).Next(f3.call).Call(ev1)
		eventEqual(t, ev3, got)
		assert.Nil(t, err)
		eventEqual(t, ev1, f1.arg)
		eventEqual(t, ev2, f2.arg)
		eventEqual(t, ev2, f3.arg)
	})

	t.Run("cancel chain", func(t *testing.T) {
		var (
			f1 = newF1Err()
			f2 = newF2()
		)
		got, err := logger.MustNewMapperFunc(f1.call).Via(f2.call).Call(ev1)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, err1)
		eventEqual(t, ev1, f1.arg)
		assert.Nil(t, f2.arg)
	})

	t.Run("2 chains", func(t *testing.T) {
		var (
			f1 = newF1()
			f2 = newF2()
		)
		got, err := logger.MustNewMapperFunc(f1.call).Via(f2.call).Call(ev1)
		assert.Nil(t, err)
		eventEqual(t, ev2, got)
		eventEqual(t, ev1, f1.arg)
		eventEqual(t, ev2, f2.arg)
	})
}

func TestMapperFuncNext(t *testing.T) {
	var (
		err1     = errors.New("err1")
		ev1      = logger.NewEvent(10, "msg", nil)
		ev2      = logger.NewEvent(11, "msg", nil)
		ev3      = logger.NewEvent(12, "msg", nil)
		newF1    = func() *mockMapperFunc { return newMockMapperFunc(ev2, nil) }
		newF2    = func() *mockMapperFunc { return newMockMapperFunc(ev3, nil) }
		newF1Err = func() *mockMapperFunc { return newMockMapperFunc(nil, err1) }
	)

	t.Run("error at the end", func(t *testing.T) {
		var (
			f1 = newF1()
			f2 = newF1Err()
		)
		got, err := logger.MustNewMapperFunc(f1.call).Next(f2.call).Call(ev1)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, err1)
		eventEqual(t, ev1, f1.arg)
		eventEqual(t, ev2, f2.arg)
	})

	t.Run("cancel chains", func(t *testing.T) {
		var (
			f1 = newF1Err()
			f2 = newF2()
		)
		got, err := logger.MustNewMapperFunc(f1.call).Next(f2.call).Call(ev1)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, err1)
		eventEqual(t, ev1, f1.arg)
		assert.Nil(t, f2.arg)
	})

	t.Run("2 chains", func(t *testing.T) {
		var (
			f1 = newF1()
			f2 = newF2()
		)
		got, err := logger.MustNewMapperFunc(f1.call).Next(f2.call).Call(ev1)
		assert.Nil(t, err)
		eventEqual(t, ev3, got)
		eventEqual(t, ev1, f1.arg)
		eventEqual(t, ev2, f2.arg)
	})
}

func TestMapperFuncCall(t *testing.T) {
	ev1 := logger.NewEvent(10, "msg", nil)

	t.Run("nil", func(t *testing.T) {
		var f logger.MapperFunc
		_, err := f.Call(ev1)
		assert.ErrorIs(t, err, logger.ErrNilMapperFunc)
	})

	t.Run("without chain", func(t *testing.T) {
		got, err := logger.MustNewMapperFunc(func(event logger.Event) (logger.Event, error) {
			return logger.NewEvent(event.Level()+1, event.Format(), event.Args()), nil
		}).Call(ev1)
		assert.Nil(t, err)
		eventEqual(t, logger.NewEvent(11, ev1.Format(), nil), got)
	})
}

func TestNewMapperFunc(t *testing.T) {
	ev1 := logger.NewEvent(10, "msg", nil)

	for _, tc := range []struct {
		title string
		f     any
		err   error
	}{
		{
			title: "invalid signature",
			f:     func() {},
			err:   logger.ErrInvalidMapperFunc,
		},
		{
			title: "normalized",
			f:     func(event logger.Event) (logger.Event, error) { return event, nil },
		},
		{
			title: "no return values",
			f:     func(_ logger.Event) {},
		},
		{
			title: "return error only",
			f:     func(_ logger.Event) error { return nil },
		},
		{
			title: "return event only",
			f:     func(event logger.Event) logger.Event { return event },
		},
	} {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			got, err := logger.NewMapperFunc(tc.f)
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
				return
			}
			assert.Nil(t, err)
			ret, err := got.Call(ev1)
			assert.Nil(t, err)
			eventEqual(t, ev1, ret)
		})
	}
}
