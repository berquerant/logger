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

func TestMapperList(t *testing.T) {
	ev1 := logger.NewEvent(logger.Linfo, "log %s", []any{"val"})

	t.Run("identity without mapping", func(t *testing.T) {
		ml := logger.NewMapperList()
		got, err := ml.Map(ev1)
		assert.Nil(t, err)
		eventEqual(t, ev1, got)
	})

	t.Run("identity with identity mapper", func(t *testing.T) {
		ml := logger.NewMapperList()
		ml.Append(func(ev logger.Event) (logger.Event, error) {
			return ev, nil
		})
		got, err := ml.Map(ev1)
		assert.Nil(t, err)
		eventEqual(t, ev1, got)
	})

	ev2 := logger.NewEvent(logger.Lwarn, "log %s", []any{"val"})

	t.Run("modify event", func(t *testing.T) {
		ml := logger.NewMapperList()
		ml.Append(func(ev logger.Event) (logger.Event, error) {
			return logger.NewEvent(logger.Lwarn, ev.Format(), ev.Args()), nil
		})
		got, err := ml.Map(ev1)
		assert.Nil(t, err)
		eventEqual(t, ev2, got)
	})

	ev3 := logger.NewEvent(logger.Lwarn, "log event %s", []any{"val"})

	t.Run("modify 2 times", func(t *testing.T) {
		ml := logger.NewMapperList()
		ml.Append(func(ev logger.Event) (logger.Event, error) {
			return logger.NewEvent(logger.Lwarn, ev.Format(), ev.Args()), nil
		})
		ml.Append(func(ev logger.Event) (logger.Event, error) {
			return logger.NewEvent(ev.Level(), "log event %s", ev.Args()), nil
		})
		got, err := ml.Map(ev1)
		assert.Nil(t, err)
		eventEqual(t, ev3, got)
	})

	t.Run("cancel by error", func(t *testing.T) {
		errCanceled := errors.New("canceled")
		ml := logger.NewMapperList()
		ml.Append(func(ev logger.Event) (logger.Event, error) {
			return nil, errCanceled
		})
		ml.Append(func(ev logger.Event) (logger.Event, error) {
			return logger.NewEvent(ev.Level(), "log event %s", ev.Args()), nil
		})
		_, err := ml.Map(ev1)
		assert.ErrorIs(t, err, errCanceled)
	})
}

type mockMapperList struct {
	got    logger.Event
	result logger.Event
	err    error
}

func (*mockMapperList) Append(_ logger.Mapper) {}
func (m *mockMapperList) Map(ev logger.Event) (logger.Event, error) {
	m.got = ev
	return m.result, m.err
}

func TestProxy(t *testing.T) {
	var (
		errFirst    = errors.New("first")
		errSecond   = errors.New("second")
		firstEvent  = logger.NewEvent(logger.Ldebug, "log 1", nil)
		secondEvent = logger.NewEvent(logger.Linfo, "log 2", nil)
	)

	t.Run("no mappers", func(t *testing.T) {
		p := logger.NewProxy()
		var gotErr error
		p.SetErrConsumer(func(err error) { gotErr = err })
		p.Put(firstEvent)
		assert.Nil(t, gotErr)
	})

	t.Run("single mapper", func(t *testing.T) {
		f := &mockMapperList{
			result: firstEvent,
		}
		p := logger.NewProxy(f)
		p.Put(firstEvent)

		eventEqual(t, firstEvent, f.got)
	})

	t.Run("error on second mappers", func(t *testing.T) {
		f := &mockMapperList{
			result: secondEvent,
		}
		s := &mockMapperList{
			err: errSecond,
		}
		p := logger.NewProxy(f, s)
		var gotErr error
		p.SetErrConsumer(func(err error) { gotErr = err })
		p.Put(firstEvent)

		eventEqual(t, firstEvent, f.got)
		eventEqual(t, secondEvent, s.got)
		assert.ErrorIs(t, errSecond, gotErr)
	})

	t.Run("error on first mappers without error consumer", func(t *testing.T) {
		f := &mockMapperList{
			result: secondEvent,
			err:    errFirst,
		}
		s := &mockMapperList{}
		p := logger.NewProxy(f, s)
		p.Put(firstEvent)

		eventEqual(t, firstEvent, f.got)
		assert.Nil(t, s.got)
	})

	t.Run("error on first mappers", func(t *testing.T) {
		f := &mockMapperList{
			result: secondEvent,
			err:    errFirst,
		}
		s := &mockMapperList{}
		p := logger.NewProxy(f, s)
		var gotErr error
		p.SetErrConsumer(func(err error) { gotErr = err })
		p.Put(firstEvent)

		eventEqual(t, firstEvent, f.got)
		assert.ErrorIs(t, gotErr, errFirst)
		assert.Nil(t, s.got)
	})

	t.Run("pass without error consumer", func(t *testing.T) {
		f := &mockMapperList{
			result: secondEvent,
		}
		s := &mockMapperList{}
		p := logger.NewProxy(f, s)
		p.Put(firstEvent)

		eventEqual(t, firstEvent, f.got)
		eventEqual(t, secondEvent, s.got)
	})

	t.Run("pass", func(t *testing.T) {
		f := &mockMapperList{
			result: secondEvent,
		}
		s := &mockMapperList{}
		p := logger.NewProxy(f, s)
		var gotErr error
		p.SetErrConsumer(func(err error) { gotErr = err })
		p.Put(firstEvent)

		eventEqual(t, firstEvent, f.got)
		eventEqual(t, secondEvent, s.got)
		assert.Nil(t, gotErr)
	})

	t.Run("modify proxy", func(t *testing.T) {
		var gotErr error
		p := logger.NewProxy()
		p.SetErrConsumer(func(err error) { gotErr = err })
		p.Put(firstEvent)
		assert.Nil(t, gotErr)
		_, ok := p.At(0)
		assert.False(t, ok)

		f := &mockMapperList{}
		p.Append(f)
		p.Put(firstEvent)
		eventEqual(t, firstEvent, f.got)
		assert.Nil(t, gotErr)
		x, ok := p.At(0)
		assert.True(t, ok)
		assert.Equal(t, f, x)
	})
}

func TestLogLevelToPrefixMapper(t *testing.T) {
	newEv := func(lv logger.Level) logger.Event {
		return logger.NewEvent(lv, "change %s", []any{"color"})
	}
	wantEv := func(lv logger.Level, prefix string) logger.Event {
		return logger.NewEvent(lv, prefix+"change %s", []any{"color"})
	}

	for _, tc := range []struct {
		title string
		level logger.Level
		want  logger.Event
	}{
		{
			title: "info",
			level: logger.Linfo,
			want:  wantEv(logger.Linfo, "I | "),
		},
		{
			title: "warn",
			level: logger.Lwarn,
			want:  wantEv(logger.Lwarn, "W | "),
		},
		{
			title: "error",
			level: logger.Lerror,
			want:  wantEv(logger.Lerror, "E | "),
		},
		{
			title: "debug",
			level: logger.Ldebug,
			want:  wantEv(logger.Ldebug, "D | "),
		},
		{
			title: "trace",
			level: logger.Ltrace,
			want:  wantEv(logger.Ltrace, "T | "),
		},
	} {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			got, err := logger.LogLevelToPrefixMapper(newEv(tc.level))
			assert.Nil(t, err)
			eventEqual(t, tc.want, got)
		})
	}
}

func TestLogLevelFilter(t *testing.T) {
	const (
		s  = logger.Lsilent
		e  = logger.Lerror
		w  = logger.Lwarn
		i  = logger.Linfo
		d  = logger.Ldebug
		tr = logger.Ltrace
	)

	newSet := func(v ...logger.Level) map[logger.Level]bool {
		set := make(map[logger.Level]bool)
		for _, x := range v {
			set[x] = true
		}
		return set
	}

	wantMetrix := map[logger.Level]map[logger.Level]bool{
		s:  newSet(),
		e:  newSet(e),
		w:  newSet(e, w),
		i:  newSet(e, w, i),
		d:  newSet(e, w, i, d),
		tr: newSet(e, w, i, d, tr),
	}
	testTargets := []logger.Level{e, w, i, d, tr}
	newEv := func(lv logger.Level) logger.Event {
		return logger.NewEvent(lv, "v", nil)
	}

	for lv, wants := range wantMetrix {
		lv := lv
		t.Run(fmt.Sprint(lv), func(t *testing.T) {
			f := logger.LogLevelFilter(lv)
			for _, tl := range testTargets {
				tl := tl
				t.Run(fmt.Sprint(tl), func(t *testing.T) {
					got, _ := f(newEv(tl))
					assert.Equal(t, wants[tl], got != nil)
				})
			}
		})
	}
}
