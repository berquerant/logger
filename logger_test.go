package logger_test

import (
	"fmt"
	"testing"

	"github.com/berquerant/logger"
	"github.com/stretchr/testify/assert"
)

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
