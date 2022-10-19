package container_test

import (
	"testing"

	"github.com/berquerant/logger/container"
	"github.com/stretchr/testify/assert"
)

func TestUpdateMap(t *testing.T) {
	for _, tc := range []struct {
		title string
		a     map[string]int
		b     map[string]int
		want  map[string]int
	}{
		{
			title: "nil and nil",
			a:     nil,
			b:     nil,
			want:  map[string]int{},
		},
		{
			title: "right nil",
			a: map[string]int{
				"x": 1,
			},
			b: nil,
			want: map[string]int{
				"x": 1,
			},
		},
		{
			title: "left nil",
			a:     nil,
			b: map[string]int{
				"x": 1,
			},
			want: map[string]int{
				"x": 1,
			},
		},
		{
			title: "add",
			a: map[string]int{
				"x": 1,
			},
			b: map[string]int{
				"y": 2,
			},
			want: map[string]int{
				"x": 1,
				"y": 2,
			},
		},
		{
			title: "right wins",
			a: map[string]int{
				"x": 1,
				"y": 2,
			},
			b: map[string]int{
				"y": 3,
				"z": 4,
			},
			want: map[string]int{
				"x": 1,
				"y": 3,
				"z": 4,
			},
		},
	} {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			assert.Equal(t, tc.want, container.UpdateMap(tc.a, tc.b))
		})
	}
}
