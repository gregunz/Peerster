package utils

import (
	"fmt"
	"sync"
)

type Map interface {
	Add(k interface{}, v interface{})
	Get(k interface{}) (interface{}, error)
	Remove(k interface{})
	Iterate(callback func(k interface{}, v interface{}))
}

type _Map struct {
	_map map[interface{}]interface{}
	mux  sync.RWMutex
}

func NewMap() Map {
	return &_Map{
		_map: map[interface{}]interface{}{},
	}
}

func (l *_Map) Add(k interface{}, v interface{}) {
	l.mux.Lock()
	defer l.mux.Unlock()

	l._map[k] = v
}

func (l *_Map) Get(w interface{}) (interface{}, error) {
	l.mux.RLock()
	defer l.mux.RUnlock()

	if v, ok := l._map[w]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("key not in map")
}

func (l *_Map) Remove(k interface{}) {
	l.mux.Lock()
	defer l.mux.Unlock()
	delete(l._map, k)
}

func (l *_Map) Iterate(callback func(k interface{}, v interface{})) {
	l.mux.RLock()
	defer l.mux.RUnlock()

	for k, v := range l._map {
		callback(k, v)
	}
}
