package inithook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// used for doc
type Attr = string
type SetterName = string

// some builtin attrs
const (
	AppName = "app_name"
	Version = "version"
)

// AttrSetter used to set attr in library
type AttrSetter[T any] func(ctx context.Context, value T) error

// RegisterAttrSetter used to register AttrSetter, and should be used in library init
// NOTE: the execution order of setters cannot be guaranteed!
func RegisterAttrSetter[T any](attr, setterName string, setter AttrSetter[T]) error {
	if setter == nil {
		return fmt.Errorf("inithook: nil attr setter")
	}
	err := attrConstructors.Register(context.Background(), attr, NewConstructor[T]())
	if err != nil && !errors.Is(err, ErrAlreadyExists) {
		return err
	}
	settersLock.Lock()
	if setters[attr] == nil {
		setters[attr] = map[SetterName]any{}
	}
	setters[attr][setterName] = genericAttrSetter(func(ctx context.Context, value any) error {
		if typed, ok := value.(T); ok {
			return setter(ctx, typed)
		}
		return fmt.Errorf("attr %s setter value type should be %T but got %T", attr, *new(T), value)
	})
	settersLock.Unlock()
	return nil
}

// ExecuteMapAttrSetters execute a map of attr setters with json format value, used in app code
func ExecuteMapAttrSetters(ctx context.Context, attrsData map[Attr]json.RawMessage) error {
	for attr, data := range attrsData {
		fn := GetAttrConstructor(attr)
		if fn == nil {
			return fmt.Errorf("inithook: attr %s constructor not found", attr)
		}
		value := fn()
		err := json.Unmarshal(data, &value)
		if err != nil {
			return fmt.Errorf("inithook: attr %s unmarshal with data %v failed: %v", attr, data, err)
		}
		err = ExecuteAttrSetters(ctx, attr, value)
		if err != nil {
			return err
		}
	}
	return nil
}

// ExecuteAttrSetters execute attr setters, used in app code
func ExecuteAttrSetters(ctx context.Context, attr string, value any) error {
	settersLock.Lock()
	attrSetters := setters[attr]
	settersLock.Unlock()
	for name, setter := range attrSetters {
		gas, ok := setter.(genericAttrSetter)
		if !ok {
			return fmt.Errorf("inithook: attr %s setter %s should be `func(context.Context, any) error` but got %T", attr, name, setter)
		}
		err := gas(ctx, value)
		if err != nil {
			return fmt.Errorf("inithook: attr %s setters %s executed failed: %v", attr, name, err)
		}
	}
	settersUsed.Set(context.Background(), attr, struct{}{})
	return nil
}

// AttrsNotSetted return a slice of attr which has not seted, used to alert in app
func AttrsNotSetted() []Attr {
	var attrNotUsed []Attr
	settersLock.Lock()
	for attr := range setters {
		if !settersUsed.Has(context.Background(), attr) {
			attrNotUsed = append(attrNotUsed, attr)
		}
	}
	settersLock.Unlock()
	return attrNotUsed
}

// NewConstructor returns a constructor which return a new T's instance constructor
// which is a new created or a cached constructor, it's used to assist the serialization procedure,
// and the constructor will indirect reflect.Ptr recursively to ensure not return nil pointer,
func NewConstructor[T any]() func() any {
	typ := reflect.TypeOf(new(T)).Elem()
	constructorsCacheLock.Lock()
	defer constructorsCacheLock.Unlock()
	if fn, ok := constructorsCache[typ]; ok {
		return fn
	}
	fn := func() any {
		var level int
		typ := reflect.TypeOf(new(T)).Elem()
		for ; typ.Kind() == reflect.Ptr; typ = typ.Elem() {
			level++
		}
		if level == 0 {
			return *new(T)
		}
		value := reflect.Zero(typ)
		for i := 0; i < level; i++ {
			p := reflect.New(value.Type())
			p.Elem().Set(value)
			value = p
		}
		return value.Interface().(T)
	}
	constructorsCache[typ] = fn
	return fn
}

// GetAttrConstructor return attr value constructor used to assist the serialization procedure
func GetAttrConstructor(attr string) func() any {
	value, _ := attrConstructors.Get(context.Background(), attr)
	return value
}

var (
	setters     = map[Attr]map[SetterName]any{}
	settersLock sync.Mutex

	constructorsCache     = map[reflect.Type]func() any{}
	constructorsCacheLock sync.Mutex
	// constructorsCache = NewMap[reflect.Type, func() any]() // NOTE: refelct.Type is not comparable!

	attrConstructors = NewMap[Attr, func() any]()

	settersUsed = NewMap[Attr, struct{}]()
)

type genericAttrSetter func(ctx context.Context, value any) error
