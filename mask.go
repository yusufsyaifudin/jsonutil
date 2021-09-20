package jsonutil

import (
	"context"
	"encoding/json"
	"reflect"
)

type (
	MaskFunc     func(ctx context.Context, value string) string
	TruncateFunc func(ctx context.Context, value string) string
)

var DefaultMaskFunc MaskFunc = func(ctx context.Context, value string) string {
	return "xxx"
}

var DefaultTruncateFunc TruncateFunc = func(ctx context.Context, value string) string {
	return "xxx"
}

type Config struct {
	Keys         map[string]MaskFunc
	TruncateFunc TruncateFunc

	// you can define your own json marshal or unmarshal for speed
	JSONMarshal   func(v interface{}) ([]byte, error)
	JSONUnmarshal func(data []byte, v interface{}) error
}

type Masking struct {
	Config Config
}

func NewMasking(conf Config) *Masking {

	for s, maskFunc := range conf.Keys {
		if maskFunc == nil {
			maskFunc = DefaultMaskFunc
		}

		conf.Keys[s] = maskFunc
	}

	if conf.JSONMarshal == nil {
		conf.JSONMarshal = json.Marshal
	}

	if conf.JSONUnmarshal == nil {
		conf.JSONUnmarshal = json.Unmarshal
	}

	return &Masking{Config: conf}
}

func (m *Masking) MaskByte(ctx context.Context, b []byte) ([]byte, error) {
	var data interface{}
	err := m.Config.JSONUnmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	out, err := m.Mask(ctx, data)
	if err != nil {
		return nil, err
	}

	return m.Config.JSONMarshal(out)
}

// Mask will handle masking of JSON string value only.
// Any value like object, array, number and null will not be masked.
// This function will walk to every JSON array element and object value.
// Means that if you have an object `{a: {b: ""}}` then you can mask the value on key b.
// This also applies in array [{a: {b: ""}}].
// In case you have an array of string like this ["", ""] it will not be masked,
// because it is top level and does not have key.
func (m *Masking) Mask(ctx context.Context, data interface{}) (interface{}, error) {
	original := reflect.ValueOf(data)
	kind := original.Kind()
	altered := reflect.New(original.Type()).Elem()

	switch kind {
	case reflect.Map:
		altered = m.maskMap(ctx, original)
	case reflect.Slice, reflect.Array:
		altered = m.maskSlice(ctx, original)
	default:
		altered.Set(original)
	}

	return altered.Interface(), nil
}

func (m *Masking) maskMap(ctx context.Context, elem reflect.Value) (altered reflect.Value) {
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
			// if key is not in the list of masked
			if maskFunc, shouldMasked := m.Config.Keys[mapRange.Key().String()]; shouldMasked {
				v := maskFunc(ctx, mapRange.Value().String())
				altered.SetMapIndex(mapRange.Key(), reflect.ValueOf(v))
				continue
			}

			altered.SetMapIndex(mapRange.Key(), mapRange.Value())

		case map[string]interface{}:
			v := m.maskMapInterface(ctx, mapRange.Value().Interface().(map[string]interface{}))
			altered.SetMapIndex(mapRange.Key(), reflect.ValueOf(v))

		case []interface{}:
			values := mapRange.Value().Interface().([]interface{})
			newArr := m.maskSliceInterface(ctx, mapRange.Key().String(), values)

			altered.SetMapIndex(mapRange.Key(), reflect.ValueOf(newArr))

		default:
			altered.SetMapIndex(mapRange.Key(), mapRange.Value())
			// fmt.Printf("%v %v %T\n", mapRange.Key(), mapRange.Value(), mapRange.Value().Interface())

		}

	}

	return
}

func (m *Masking) maskMapInterface(ctx context.Context, myMap map[string]interface{}) map[string]interface{} {
	for k, v := range myMap {

		switch v.(type) {
		case string:
			if maskFunc, shouldMasked := m.Config.Keys[k]; shouldMasked {
				myMap[k] = maskFunc(ctx, v.(string))
				continue
			}

			myMap[k] = v

		case map[string]interface{}:
			// No need to check if key is in whitelist or not, because we do recursive call.
			// Hence, only when the final value is string or slice
			// we must check whether we should continue to mask or not.
			myMap[k] = m.maskMapInterface(ctx, v.(map[string]interface{}))

		case []interface{}:
			myMap[k] = m.maskSliceInterface(ctx, k, v.([]interface{}))

		default:
			myMap[k] = v
		}

	}

	return myMap
}

func (m *Masking) maskSlice(ctx context.Context, elem reflect.Value) (altered reflect.Value) {
	altered = reflect.MakeSlice(elem.Type(), elem.Len(), elem.Len())
	for i := 0; i < elem.Len(); i++ {
		value := elem.Index(i)

		switch value.Interface().(type) {
		case string:
			// On top level array, if the array contain string, then don't mask it unless it has key.
			// altered.Index(i).Set(reflect.ValueOf(maskedStr))
			altered.Index(i).Set(value)
		case map[string]interface{}:
			v := m.maskMapInterface(ctx, value.Interface().(map[string]interface{}))
			altered.Index(i).Set(reflect.ValueOf(v))
		case []interface{}:
			// top level array doesn't have key
			v := m.maskSliceInterface(ctx, "", value.Interface().([]interface{}))
			altered.Index(i).Set(reflect.ValueOf(v))
		default:
			altered.Index(i).Set(value)
		}
	}

	return
}

func (m *Masking) maskSliceInterface(ctx context.Context, key string, slices []interface{}) []interface{} {
	newSlices := make([]interface{}, len(slices))
	for i, v := range slices {
		switch v.(type) {
		case string:
			if maskFunc, shouldMasked := m.Config.Keys[key]; shouldMasked {
				newSlices[i] = maskFunc(ctx, v.(string))
				continue
			}

			newSlices[i] = v

		case map[string]interface{}:
			newSlices[i] = m.maskMapInterface(ctx, v.(map[string]interface{}))

		case []interface{}:
			newSlices[i] = m.maskSliceInterface(ctx, key, v.([]interface{}))

		default:
			newSlices[i] = v
		}

	}

	return newSlices
}
