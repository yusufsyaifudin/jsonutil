package envmap

import (
	"encoding/json"
	"fmt"

	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"gopkg.in/yaml.v3"
)

type Kind int

const (
	kindUnknown Kind = iota
	KindString
	KindArray
)

type StrOrArr struct {
	str    string
	arrStr []string
}

func (s StrOrArr) Kind() Kind {
	if s.str != "" && len(s.arrStr) > 0 {
		return kindUnknown
	}

	switch {
	case s.str != "":
		return KindString

	case len(s.arrStr) > 0:
		return KindArray
	}

	// treat as string value by default
	return KindString
}

func (s *StrOrArr) String() string {
	return s.str
}

func (s *StrOrArr) Array() []string {
	if s.Kind() == KindString {
		return []string{}
	}

	return s.arrStr
}

func String(str string) *StrOrArr {
	return &StrOrArr{str: str}
}

func StringArray(arrStr []string) *StrOrArr {
	return &StrOrArr{arrStr: arrStr}
}

var _ fmt.Stringer = (*StrOrArr)(nil)
var _ json.Marshaler = (*StrOrArr)(nil)
var _ json.Unmarshaler = (*StrOrArr)(nil)
var _ yaml.Marshaler = (*StrOrArr)(nil)
var _ yaml.Unmarshaler = (*StrOrArr)(nil)
var _ bson.ValueMarshaler = (*StrOrArr)(nil)
var _ bson.ValueUnmarshaler = (*StrOrArr)(nil)

func (s StrOrArr) MarshalJSON() ([]byte, error) {
	if s.str != "" && len(s.arrStr) > 0 {
		return nil, fmt.Errorf("envmap.json: cannot pick str or array of str")
	}

	switch {
	case s.str != "":
		return []byte(fmt.Sprintf("%q", s.str)), nil
	case len(s.arrStr) > 0:
		arrStr := make([]string, 0)
		err := copier.Copy(&arrStr, s.arrStr)
		if err != nil {
			err = fmt.Errorf("cannot copy arr str: %w", err)
			return nil, err
		}

		return json.Marshal(arrStr)
	}

	// by default, return as a null value on json
	return nil, nil
}

func (s *StrOrArr) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case string:
		s.str = value
		return nil

	case []interface{}:
		arrStr := make([]string, 0)
		for _, val := range value {
			switch str := val.(type) {
			case string:
				arrStr = append(arrStr, str)

			default:
				return fmt.Errorf("one of array element contains non str value: (%T) %+v", val, val)
			}

		}

		s.arrStr = arrStr
		return nil
	}

	return fmt.Errorf("not support type %T on envmap.UnmarshalJSON", v)
}

func (s StrOrArr) MarshalYAML() (interface{}, error) {
	if s.str != "" && len(s.arrStr) > 0 {
		return nil, fmt.Errorf("envmap.json: cannot pick str or array of str")
	}

	switch {
	case s.str != "":
		return s.str, nil
	case len(s.arrStr) > 0:
		return s.arrStr, nil
	}

	// by default, return as a empty value on json
	return nil, nil
}

func (s *StrOrArr) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		// simple str
		s.str = value.Value
		return nil

	case yaml.SequenceNode:
		// array
		arrStr := make([]string, 0)
		for idx, node := range value.Content {
			if node.Kind != yaml.ScalarNode {
				return fmt.Errorf("elements %d contains non-str type %s %+v %+v", idx, node.Tag, node.Value, node.Content)
			}

			arrStr = append(arrStr, node.Value)
		}

		s.arrStr = arrStr
		return nil

	}

	return fmt.Errorf("not support type %d %s on envmap.UnmarshalYAML", value.Kind, value.Tag)
}

func (s StrOrArr) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if s.str != "" && len(s.arrStr) > 0 {
		return bsontype.Null, nil, fmt.Errorf("envmap.json: cannot pick str or array of str")
	}

	switch {
	case s.str != "":
		return bson.MarshalValue(s.str)
	case len(s.arrStr) > 0:
		arrStr := make([]string, 0)
		err := copier.Copy(&arrStr, s.arrStr)
		if err != nil {
			err = fmt.Errorf("cannot copy arr str: %w", err)
			return bsontype.Null, nil, err
		}

		return bson.MarshalValue(arrStr)
	}

	// by default, return as a null value on bson
	return bson.TypeNull, nil, nil
}

func (s *StrOrArr) UnmarshalBSONValue(typ bsontype.Type, b []byte) error {
	switch typ {
	case bsontype.Null:
		return nil

	case bsontype.String:

		raw := bson.RawValue{
			Type:  bsontype.String,
			Value: b,
		}

		s.str = raw.StringValue()
		return nil

	case bsontype.Array:

		raw := bson.RawValue{
			Type:  bsontype.Array,
			Value: b,
		}

		s.arrStr = make([]string, 0)

		arrVal, err := raw.Array().Values()
		if err != nil {
			err = fmt.Errorf("envmap.UnmarshalBSONValue cannot get array values: %w", err)
			return err
		}

		for _, val := range arrVal {
			s.arrStr = append(s.arrStr, val.StringValue())
		}
		return nil

	}

	return fmt.Errorf("envmap.UnmarshalBSONValue cannot unmarshal type %s: %s", typ, b)
}
