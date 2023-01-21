package inithook

import (
	"context"
	"reflect"
	"sync"

	"github.com/pkg/errors"
)

var (
	// ErrNotFound defines not found error
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists defines already exists error
	ErrAlreadyExists = errors.New("already exists")
)

// Map is a instances map of specified Type
type Map[K comparable, V any] struct {
	instances map[K]V
	lock      sync.RWMutex
}

// NewMap creates a new map
func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{
		instances: make(map[K]V),
	}
}

// MustRegister register a V's instance with key, if failed(e.g. already exists) then panic
func (m *Map[K, V]) MustRegister(ctx context.Context, key K, value V) {
	err := m.Register(ctx, key, value)
	if err != nil {
		panic(err)
	}
}

// Register register a V's instance with key, if exists then return `ErrAlreadyExists` error(use `errors.Is` to assert)
func (m *Map[K, V]) Register(ctx context.Context, key K, value V) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.instances[key]; ok {
		return errors.WithMessagef(ErrAlreadyExists, "type %T instance %v", value, key)
	}
	m.instances[key] = value
	return nil
}

// MustSet set a V's instance with key, if exists then override, if failed then panic
func (m *Map[K, V]) MustSet(ctx context.Context, key K, value V) {
	err := m.Set(ctx, key, value)
	if err != nil {
		panic(err)
	}
}

// Set set a V's instance with key, if exists then override
func (m *Map[K, V]) Set(ctx context.Context, key K, value V) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.instances[key] = value
	return nil
}

// MustDelete delete a V's instance specified by key, if failed then panic
func (m *Map[K, V]) MustDelete(ctx context.Context, key K) {
	err := m.Delete(ctx, key)
	if err != nil {
		panic(err)
	}
}

// Delete delete a V's instance specified by key
func (m *Map[K, V]) Delete(ctx context.Context, key K) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.instances, key)
	return nil
}

// MustClear clear all V's instances, if failed then panic
func (m *Map[K, V]) MustClear(ctx context.Context) {
	err := m.Clear(ctx)
	if err != nil {
		panic(err)
	}
}

// Clear clear all V's instances
func (m *Map[K, V]) Clear(ctx context.Context) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.instances = make(map[K]V)
	return nil
}

// GetDefault get a V's instance by key, if not found return `NotFound` error(use `errors.Is` to assert)
func (m *Map[K, V]) Get(ctx context.Context, key K) (V, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if v, ok := m.instances[key]; ok {
		return v, nil
	}
	value := *new(V)
	return value, errors.WithMessagef(ErrNotFound, "type %T instance %v", value, key)
}

// GetDefault get a V's instance by key, if not found, then try to returns a default one
func (m *Map[K, V]) GetDefault(ctx context.Context, key K) (V, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if v, ok := m.instances[key]; ok {
		return v, nil
	}
	return m.Default(ctx, key)
}

// Default returns V's default value if it implement the `DefaultLoader` or `Default`, otherwise return `Zero[V]()`
func (m *Map[K, V]) Default(ctx context.Context, key K) (V, error) {
	var value = Zero[V]()
	if defLoader, ok := any(value).(DefaultLoader[V]); ok {
		return defLoader.LoadDefault(ctx, key)

	}
	if def, ok := any(value).(Default[V]); ok {
		return def.Default(), nil
	}
	return value, nil
}

// Has tells if map has key
func (m *Map[K, V]) Has(ctx context.Context, key K) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	_, ok := m.instances[key]
	return ok
}

// Range calls f sequentially for each key and value present in the map. If f returns false, range stops the iteration.
func (m *Map[K, V]) Range(ctx context.Context, fn func(key, value any) bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	for k, v := range m.instances {
		shouldContinue := fn(k, v)
		if !shouldContinue {
			return
		}
	}
}

// DefaultLoader load default instance of V according to key
type DefaultLoader[V any] interface {
	LoadDefault(ctx context.Context, key any) (V, error)
}

// Default giving a type a useful default value.
type Default[V any] interface {
	Default() V
}

// Zero create a new V's instance, and New will indirect reflect.Ptr recursively to ensure not return nil pointer
func Zero[V any]() V {
	var level int
	typ := reflect.TypeOf(new(V)).Elem()
	for ; typ.Kind() == reflect.Ptr; typ = typ.Elem() {
		level++
	}
	if level == 0 {
		return *new(V)
	}
	value := reflect.Zero(typ)
	for i := 0; i < level; i++ {
		p := reflect.New(value.Type())
		p.Elem().Set(value)
		value = p
	}
	return value.Interface().(V)
}
