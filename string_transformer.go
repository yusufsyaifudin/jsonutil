package jsonutil

import (
	"context"
	"encoding/json"
	"reflect"
)

type Type int

const (
	Object Type = iota
	Array
)

type KVInfo struct {
	IsTopLevel bool
	Inside     Type // Inside specify whether current Value is inside Object or Array.
	Key        string
	Value      string
}

// StringTransformer is a function to replace value to new value.
type StringTransformer func(ctx context.Context, info KVInfo) string

// DefaultStringTransformer will not Transform any value.
var DefaultStringTransformer StringTransformer = func(ctx context.Context, info KVInfo) string {
	return info.Value
}

type Config struct {
	StringTransformer StringTransformer

	// you can define your own json marshal or unmarshal for speed.
	JSONMarshal   func(v interface{}) ([]byte, error)
	JSONUnmarshal func(data []byte, v interface{}) error
}

type Transformer struct {
	Config Config
}

func NewTransformer(conf Config) *Transformer {
	if conf.StringTransformer == nil {
		conf.StringTransformer = DefaultStringTransformer
	}

	if conf.JSONMarshal == nil {
		conf.JSONMarshal = json.Marshal
	}

	if conf.JSONUnmarshal == nil {
		conf.JSONUnmarshal = json.Unmarshal
	}

	return &Transformer{Config: conf}
}

func (m *Transformer) TransformBytes(ctx context.Context, b []byte) ([]byte, error) {
	var data interface{}
	err := m.Config.JSONUnmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	out, err := m.Transform(ctx, data)
	if err != nil {
		return nil, err
	}

	return m.Config.JSONMarshal(out)
}

// Transform will handle masking of JSON string value only.
// Any value like object, array, number and null will not be masked.
// This function will walk to every JSON array element and object value.
// Means that if you have an object `{a: {b: ""}}` then you can mask the value on key b.
// This also applies in array [{a: {b: ""}}].
func (m *Transformer) Transform(ctx context.Context, data interface{}) (interface{}, error) {
	original := reflect.ValueOf(data)
	kind := original.Kind()
	altered := reflect.New(original.Type()).Elem()

	switch kind {
	case reflect.Map:
		altered = m.maskMap(ctx, original)
	case reflect.Slice, reflect.Array:
		altered = m.maskSlice(ctx, original)
	default:
		// string only such as "abc" is a valid JSON.
		altered.Set(original)
	}

	return altered.Interface(), nil
}

// maskMap will always call when we found top level object, so isTopElem wil always true.
func (m *Transformer) maskMap(ctx context.Context, elem reflect.Value) (altered reflect.Value) {
	altered = reflect.MakeMapWithSize(elem.Type(), len(elem.MapKeys()))
	mapRange := elem.MapRange()
	for mapRange.Next() {

		// key must be string, the valid JSON must have string as a key
		if _, ok := mapRange.Key().Interface().(string); !ok {
			altered.SetMapIndex(mapRange.Key(), mapRange.Value())
			continue
		}

		// value must be string in order to mask
		switch mapRange.Value().Interface().(type) {
		case string:
			// top level kv string, e.g: {"a": "b"}
			// this will handle on value part: "b"
			v := m.Config.StringTransformer(ctx, KVInfo{
				IsTopLevel: true,
				Inside:     Object,
				Key:        mapRange.Key().Interface().(string),
				Value:      mapRange.Value().Interface().(string),
			})

			altered.SetMapIndex(mapRange.Key(), reflect.ValueOf(v))

		case map[string]interface{}:
			// top level kv, with v contains object, e.g: {"foo": {"a": "b"}}
			// this will handle on value part: {"a": "b"}
			v := m.maskMapInterface(ctx, mapRange.Value().Interface().(map[string]interface{}))
			altered.SetMapIndex(mapRange.Key(), reflect.ValueOf(v))

		case []interface{}:
			// top level kv with v contains mixed element on array, e.g: {"foo": ["a",1]}
			// this will handle on part ["a",1]
			values := mapRange.Value().Interface().([]interface{})
			newArr := m.maskSliceInterface(ctx, mapRange.Key().String(), values)

			altered.SetMapIndex(mapRange.Key(), reflect.ValueOf(newArr))

		default:
			// top level kv, with v contains type but not string,
			// e.g: {"foo": 1}
			// this will handle on value part: 1
			altered.SetMapIndex(mapRange.Key(), mapRange.Value())
		}

	}

	return
}

func (m *Transformer) maskMapInterface(ctx context.Context, myMap map[string]interface{}) map[string]interface{} {
	for k, v := range myMap {

		switch v.(type) {
		case string:
			// when passed object {"foo": "bar"}, this will handle value "bar" as string
			transformedVal := m.Config.StringTransformer(ctx, KVInfo{
				IsTopLevel: false,
				Inside:     Object,
				Key:        k,
				Value:      v.(string),
			})

			myMap[k] = transformedVal

		case map[string]interface{}:
			// When passed object contains object: {"foo":{"another_obj":{"foo":"bar"}}},
			// this will handle value {"another_obj":{"foo":"bar"}} as map[string]interface{}
			// And call this function recursively.

			// No need to check if key is in whitelist or not, because we do recursive call.
			// Hence, only when the final value is string or slice
			// we must check whether we should continue to mask or not.
			myMap[k] = m.maskMapInterface(ctx, v.(map[string]interface{}))

		case []interface{}:
			// When passed object contains array {"foo":{"another_obj":[{"foo":"bar"}]}}
			// This will handle each element on foo {"another_obj":[{"foo":"bar"}]} and call to slice interface.
			myMap[k] = m.maskSliceInterface(ctx, k, v.([]interface{}))

		default:
			// When passed object contains elements other than string, object kv string or array, it will keep default.
			// e.g: {"foo": {"foo": 1}}, this will handle {"foo": 1} and
			// detect that 1 as integer and pass the original value to myMap.
			myMap[k] = v
		}

	}

	return myMap
}

// maskSlice will always call when we found top level array, so isTopElem wil always true.
func (m *Transformer) maskSlice(ctx context.Context, elem reflect.Value) (altered reflect.Value) {
	altered = reflect.MakeSlice(elem.Type(), elem.Len(), elem.Len())
	for i := 0; i < elem.Len(); i++ {
		value := elem.Index(i)

		switch value.Interface().(type) {
		case string:
			// this is top level element, such as ["a","b"]
			v := m.Config.StringTransformer(ctx, KVInfo{
				IsTopLevel: true,
				Inside:     Array,
				Key:        "",
				Value:      value.Interface().(string),
			})

			altered.Index(i).Set(reflect.ValueOf(v))

		case map[string]interface{}:
			// top level with array of object: [{"a":"b"}]
			v := m.maskMapInterface(ctx, value.Interface().(map[string]interface{}))
			altered.Index(i).Set(reflect.ValueOf(v))

		case []interface{}:
			// top level array, contains another array, multi-dimension array, e.g: [[{"foo":"bar"}]]
			v := m.maskSliceInterface(ctx, "", value.Interface().([]interface{}))
			altered.Index(i).Set(reflect.ValueOf(v))

		default:
			// mixed content of top level array, e.g: ["amount", 100, {"a":"b"}]
			// or [1,2.2]
			altered.Index(i).Set(value)
		}
	}

	return
}

func (m *Transformer) maskSliceInterface(ctx context.Context, key string, slices []interface{}) []interface{} {
	newSlices := make([]interface{}, len(slices))
	for i, v := range slices {
		switch v.(type) {
		case string:
			// e.g: [{"foo":["a","b"]}] will iterate over a, b
			transformedVal := m.Config.StringTransformer(ctx, KVInfo{
				IsTopLevel: false,
				Inside:     Array,
				Key:        key,
				Value:      v.(string),
			})
			newSlices[i] = transformedVal

		case map[string]interface{}:
			// e.g: {"foo":[{"a":"b"},{"c":"d"}]} will iterate over foo elements
			newSlices[i] = m.maskMapInterface(ctx, v.(map[string]interface{}))

		case []interface{}:
			// array contain multidimensional array, e.g: {"mixed": [[{"foo": "bar"}]]}
			// will iterate the elements "mixed" and each value will call this func recursively
			newSlices[i] = m.maskSliceInterface(ctx, key, v.([]interface{}))

		default:
			// if element is not contain string, e.g: [1,2] will iterate over 1 and 2
			newSlices[i] = v
		}

	}

	return newSlices
}
