package model

import "sync"

type SafeMap[K comparable, V any] struct {
	mutex sync.Mutex
	m     map[K]V
}

func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	return &SafeMap[K, V]{
		m: make(map[K]V),
	}
}

func (sm *SafeMap[K, V]) Get(key K) (value V, ok bool) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	value, ok = sm.m[key]
	return value, ok
}

func (sm *SafeMap[K, V]) Put(key K, value V) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.m[key] = value
}

type AtomicCounter struct {
	counter int
	mutex   sync.Mutex
}

func (ac *AtomicCounter) increment() int {
	ac.mutex.Lock()
	defer ac.mutex.Unlock()
	ac.counter++
	return ac.counter
}

func MapOver[T any, V any](input []T, f func(T) V) []V {
	output := make([]V, len(input))
	for i, value := range input {
		output[i] = f(value)
	}
	return output
}

type Receiver[M any] struct {
	Channel chan<- M
}
type Sender[M any] struct {
	Channel <-chan M
}
