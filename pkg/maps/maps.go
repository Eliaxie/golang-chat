package maps

import (
	"sync"
)

var LockMap map[interface{}]*sync.Mutex = make(map[interface{}]*sync.Mutex)

func Store[T comparable, G any](_map *map[T]G, key T, value G) {
	if LockMap[_map] == nil {
		LockMap[_map] = &sync.Mutex{}
	}
	LockMap[_map].Lock()
	(*_map)[key] = value
	LockMap[_map].Unlock()
}

func Load[T comparable, G any](_map *map[T]G, key T) G {
	if LockMap[_map] == nil {
		LockMap[_map] = &sync.Mutex{}
	}
	LockMap[_map].Lock()
	value := (*_map)[key]
	LockMap[_map].Unlock()
	return value
}

func LoadAndCheck[T comparable, G any](_map *map[T]G, key T) (G, bool) {
	if LockMap[_map] == nil {
		LockMap[_map] = &sync.Mutex{}
	}
	LockMap[_map].Lock()
	value, found := (*_map)[key]
	LockMap[_map].Unlock()
	return value, found
}

func Delete[T comparable, G any](_map *map[T]G, key T) {
	if LockMap[_map] == nil {
		LockMap[_map] = &sync.Mutex{}
	}
	LockMap[_map].Lock()
	delete((*_map), key)
	LockMap[_map].Unlock()
}

func Keys[T comparable, G any](_map *map[T]G) []T {
	if LockMap[_map] == nil {
		LockMap[_map] = &sync.Mutex{}
	}
	LockMap[_map].Lock()
	keys := make([]T, 0, len(*_map))
	for k := range *_map {
		keys = append(keys, k)
	}
	LockMap[_map].Unlock()
	return keys
}

func Values[T comparable, G any](_map *map[T]G) []G {
	if LockMap[_map] == nil {
		LockMap[_map] = &sync.Mutex{}
	}
	LockMap[_map].Lock()
	values := make([]G, 0, len(*_map))
	for _, v := range *_map {
		values = append(values, v)
	}
	LockMap[_map].Unlock()
	return values
}

func KeysValues[T comparable, G any](_map *map[T]G) (keys []T, values []G) {
	if LockMap[_map] == nil {
		LockMap[_map] = &sync.Mutex{}
	}
	LockMap[_map].Lock()
	keys = make([]T, 0, len(*_map))
	values = make([]G, 0, len(*_map))
	for k, v := range *_map {
		keys = append(keys, k)
		values = append(values, v)
	}
	LockMap[_map].Unlock()
	return keys, values
}

func Copy[T comparable, G any](_map *map[T]G) (keys []T, values []G) {
	if LockMap[_map] == nil {
		LockMap[_map] = &sync.Mutex{}
	}
	LockMap[_map].Lock()
	keys = make([]T, 0, len(*_map))
	values = make([]G, 0, len(*_map))
	for k, v := range *_map {
		keys = append(keys, k)
		values = append(values, v)
	}
	LockMap[_map].Unlock()
	return keys, values
}

func Clone[T comparable, G any](_map *map[T]G) map[T]G {
	if LockMap[_map] == nil {
		LockMap[_map] = &sync.Mutex{}
	}
	LockMap[_map].Lock()
	clone := make(map[T]G)
	for k, v := range *_map {
		clone[k] = v
	}
	LockMap[_map].Unlock()
	return clone
}
