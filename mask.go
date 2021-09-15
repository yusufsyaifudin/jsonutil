package jsonutil

import (
	"context"
	"reflect"
)

const maskedStr = "xxx"

type MaskFunc func(ctx context.Context, value string) string

func DefaultMaskFunc(ctx context.Context, value string) string {
	return maskedStr
}

type Config struct {
	Keys map[string]struct{}
}

type Masking struct {
	Keys map[string]struct{}
}

func NewMasking(conf Config) *Masking {

	return &Masking{
		Keys: conf.Keys,
	}
}

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

	return altered, nil
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
			if _, shouldMasked := m.Keys[mapRange.Key().String()]; !shouldMasked {
				altered.SetMapIndex(mapRange.Key(), mapRange.Value())
				continue
			}

			altered.SetMapIndex(mapRange.Key(), reflect.ValueOf(maskedStr))

		case map[string]interface{}:
			v := m.maskMapInterface(ctx, mapRange.Value().Interface().(map[string]interface{}))
			altered.SetMapIndex(mapRange.Key(), reflect.ValueOf(v))

		case []interface{}:
			values := mapRange.Value().Interface().([]interface{})
			newArr := m.maskSliceInterface(ctx, values)

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
			if _, shouldMasked := m.Keys[k]; !shouldMasked {
				myMap[k] = v
				continue
			}

			myMap[k] = maskedStr

		case map[string]interface{}:
			// No need to check if key is in whitelist or not, because we do recursive call.
			// Hence, only when the final value is string or slice
			// we must check whether we should continue to mask or not.
			myMap[k] = m.maskMapInterface(ctx, v.(map[string]interface{}))

		case []interface{}:

			// TODO: some bug occur here
			if _, shouldMasked := m.Keys[k]; !shouldMasked {
				myMap[k] = v
				continue
			}

			myMap[k] = m.maskSliceInterface(ctx, v.([]interface{}))

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
			v := m.maskSliceInterface(ctx, value.Interface().([]interface{}))
			altered.Index(i).Set(reflect.ValueOf(v))
		default:
			altered.Index(i).Set(value)
		}
	}

	return
}

func (m *Masking) maskSliceInterface(ctx context.Context, slices []interface{}) []interface{} {
	newSlices := make([]interface{}, len(slices))
	for i, v := range slices {
		switch v.(type) {
		case string:
			newSlices[i] = maskedStr

		case map[string]interface{}:
			newSlices[i] = m.maskMapInterface(ctx, v.(map[string]interface{}))

		case []interface{}:
			newSlices[i] = m.maskSliceInterface(ctx, v.([]interface{}))

		default:
			newSlices[i] = v
		}

	}

	return newSlices
}
