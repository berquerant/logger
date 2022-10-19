package container

import (
	"encoding/json"

	"github.com/berquerant/logger"
)

// UpdateMap returns a new map with the merged keys and values of a and b.
func UpdateMap[K comparable, V any](a, b map[K]V) map[K]V {
	minSize := func() int {
		if len(a) > len(b) {
			return len(a)
		}
		return len(b)
	}()
	c := make(map[K]V, minSize)
	for k, v := range a {
		c[k] = v
	}
	for k, v := range b {
		c[k] = v
	}
	return c
}

type Map[K comparable, V any] map[K]V

func NewMap[K comparable, V any]() Map[K, V] {
	var m Map[K, V]
	return m
}

func (m Map[K, V]) Set(key K, value V) Map[K, V] {
	m[key] = value
	return m
}

func (m Map[K, V]) Get(key K) (V, bool) {
	v, ok := m[key]
	return v, ok
}

func (m Map[K, V]) Update(other Map[K, V]) Map[K, V] { return Map[K, V](UpdateMap(m, other)) }
func (m Map[K, V]) Clone() Map[K, V]                 { return UpdateMap(m, nil) }

// StructMapper appends the map as JSON at the tail.
func (m Map[K, V]) StructMapper(ev logger.Event) logger.Event {
	b, err := json.Marshal(m)
	if err != nil {
		return logger.NewEvent(ev.Level(), ev.Format()+" | %v", append(ev.Args(), err))
	}
	return logger.NewEvent(ev.Level(), ev.Format()+" | %s", append(ev.Args(), b))
}
