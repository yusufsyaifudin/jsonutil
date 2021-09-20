package jsonutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// Value is a raw encoded JSON value.
// It implements json.Marshaler and json.Unmarshaler and can
// be used to delay JSON decoding or precompute a JSON encoding.
type Value struct {
	str string
	raw interface{}
	json.RawMessage
}

var _ json.Marshaler = (*Value)(nil)
var _ json.Unmarshaler = (*Value)(nil)

func NewValue(value interface{}) Value {
	return Value{
		str: fmt.Sprintf("%v", value),
		raw: value,
	}
}

// MarshalJSON returns v as the JSON encoding of v.
func (v Value) MarshalJSON() ([]byte, error) {
	if v.raw == nil {
		return []byte("null"), nil
	}
	return json.Marshal(v.raw)
}

// UnmarshalJSON sets *v to a copy of data.
func (v *Value) UnmarshalJSON(data []byte) error {
	if v == nil {
		return errors.New("jsonutil.Value: UnmarshalJSON on nil pointer")
	}

	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if raw == nil {
		return nil
	}

	switch raw.(type) {
	case string:
		v.str = raw.(string)
	case float32, float64, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		v.str = fmt.Sprint(raw)
	default:
		v.str = fmt.Sprintf("%v", raw)
	}

	// always write as raw
	v.raw = raw
	return nil
}

func (v Value) String() string {
	if v.raw == nil {
		return ""
	}

	return fmt.Sprintf("%v", v.raw)
}

func (v Value) Int64() (int64, error) {
	return strconv.ParseInt(v.str, 10, 64)
}

func (v Value) Float64() (float64, error) {
	return strconv.ParseFloat(v.str, 64)
}

func (v Value) Interface() interface{} {
	return v.raw
}
