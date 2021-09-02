package jsonutil_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yusufsyaifudin/jsonutil"
)

const sampleComplexJsonData = `{"real_string":"123","real_int":123,"real_float":12.3,"real_object":{"foo":{"any":{"float":1.1,"int":1},"float":{"float":1.1},"int":{"int":1}}},"real_array":["abc"],"numeric":{"float32":3.2,"float64":6.4,"int":1,"int8":8,"int16":16,"int32":32,"int64":64,"uint":10,"uint8":80,"uint16":160,"uint32":320,"uint64":640,"uintptr":9999},"null":null}`

type Numeric struct {
	Float32 float32 `json:"float32"`
	Float64 float64 `json:"float64"`
	Int     int     `json:"int"`
	Int8    int8    `json:"int8"`
	Int16   int16   `json:"int16"`
	Int32   int32   `json:"int32"`
	Int64   int64   `json:"int64"`
	UInt    uint    `json:"uint"`
	UInt8   uint8   `json:"uint8"`
	UInt16  uint16  `json:"uint16"`
	UInt32  uint32  `json:"uint32"`
	UInt64  uint64  `json:"uint64"`
	Uintptr uintptr `json:"uintptr"`
}

type Complex struct {
	RealString jsonutil.Value `json:"real_string"`
	RealInt    jsonutil.Value `json:"real_int"`
	RealFloat  jsonutil.Value `json:"real_float"`
	RealObject jsonutil.Value `json:"real_object"`
	RealArray  jsonutil.Value `json:"real_array"`
	Numeric    Numeric        `json:"numeric"`
	Null       jsonutil.Value `json:"null"`
}

var obj = map[string]interface{}{
	"foo": map[string]interface{}{
		"int": map[string]int{
			"int": 1,
		},
		"float": map[string]float64{
			"float": 1.1,
		},
		"any": map[string]interface{}{
			"int":   1,
			"float": 1.1,
		},
	},
}

var arrData = []string{"abc"}

var numericData = Numeric{
	Float32: 3.2,
	Float64: 6.4,
	Int:     1,
	Int8:    8,
	Int16:   16,
	Int32:   32,
	Int64:   64,
	UInt:    10,
	UInt8:   80,
	UInt16:  160,
	UInt32:  320,
	UInt64:  640,
	Uintptr: 9999,
}

func TestValue(t *testing.T) {

	t.Run("marshal_unmarshal", func(t *testing.T) {
		expected := Complex{
			RealString: jsonutil.NewValue("123"),
			RealInt:    jsonutil.NewValue(123),
			RealFloat:  jsonutil.NewValue(12.3),
			RealObject: jsonutil.NewValue(obj),
			RealArray:  jsonutil.NewValue(arrData),
			Numeric:    numericData,
			Null:       jsonutil.NewValue(nil),
		}

		bytes, err := json.Marshal(expected)
		assert.NotNil(t, bytes)
		assert.NoError(t, err)

		var actual Complex
		err = json.Unmarshal(bytes, &actual)
		assert.NoError(t, err)

		// assert each field
		assert.EqualValues(t, expected.RealString, actual.RealString)

		// int will save as raw float64 after unmarshal, so we get the actual value instead of comparing struct Value
		assert.EqualValues(t, expected.RealInt.Interface(), actual.RealInt.Interface())
		assert.EqualValues(t, expected.RealFloat, actual.RealFloat)

		// For type interface such as map, slice or struct,
		// when created using NewValue it uses real type such as map[string]string or []string{}
		// but after unmarshal, it becomes map[string]interface{} or []interface{}
		// This is expected and handling this is negligible since the important thing is the value, not the type.
		assert.ObjectsAreEqual(expected.RealObject, actual.RealObject)
		assert.ObjectsAreEqual(expected.RealArray, actual.RealArray)
		assert.EqualValues(t, expected.Numeric, actual.Numeric)
	})

	t.Run("unmarshal_marshal", func(t *testing.T) {
		var data Complex
		err := json.Unmarshal([]byte(sampleComplexJsonData), &data)
		assert.NoError(t, err)

		bytes, err := json.Marshal(data)
		assert.NotNil(t, bytes)
		assert.NoError(t, err)
		assert.EqualValues(t, sampleComplexJsonData, string(bytes))
	})

	t.Run("unmarshal", func(t *testing.T) {
		var data Complex
		err := json.Unmarshal([]byte(sampleComplexJsonData), &data)
		assert.NoError(t, err)

		// assert data is string but can convert into int64 or float64
		assert.EqualValues(t, "123", data.RealString.String())

		i, err := data.RealString.Int64()
		assert.EqualValues(t, int64(123), i)
		assert.NoError(t, err)

		f, err := data.RealString.Float64()
		assert.EqualValues(t, 123, f)
		assert.NoError(t, err)

		// assert data actual is int64, but can be converted as string or float64
		assert.EqualValues(t, "123", data.RealInt.String())

		i, err = data.RealInt.Int64()
		assert.EqualValues(t, int64(123), i)
		assert.NoError(t, err)

		f, err = data.RealInt.Float64()
		assert.EqualValues(t, 123, f)
		assert.NoError(t, err)

		// assert data actual is float64, but can be converted as string
		assert.EqualValues(t, "12.3", data.RealFloat.String())

		// float64 cannot be converted into int64, hence return 0 and error
		i, err = data.RealFloat.Int64()
		assert.EqualValues(t, int64(0), i)
		assert.Error(t, err)

		f, err = data.RealFloat.Float64()
		assert.EqualValues(t, 12.3, f)
		assert.NoError(t, err)

		// object,
		// actual is converted to map[]interface while the expected as map[]float64 or any concrete type
		// so, use ObjectsAreEqualValues to compare
		assert.ObjectsAreEqualValues(obj, data.RealObject.Interface())
		assert.ObjectsAreEqualValues(arrData, data.RealArray.Interface())
		assert.EqualValues(t, numericData, data.Numeric)
	})
}

func TestValue_MarshalJSON(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		data := jsonutil.NewValue(nil)
		b, err := json.Marshal(data)
		assert.EqualValues(t, []byte("null"), b)
		assert.NoError(t, err)
	})
}

func BenchmarkValue_MarshalJSON(b *testing.B) {
	complexData := Complex{
		RealString: jsonutil.NewValue("123"),
		RealInt:    jsonutil.NewValue(123),
		RealFloat:  jsonutil.NewValue(12.3),
		RealObject: jsonutil.NewValue(obj),
		RealArray:  jsonutil.NewValue(arrData),
		Numeric:    numericData,
		Null:       jsonutil.NewValue(nil),
	}

	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(complexData)
		if err != nil {
			b.Fatal(err)
			return
		}
	}
}

func Benchmark_UnmarshalJSON(b *testing.B) {
	complexData := Complex{
		RealString: jsonutil.NewValue("123"),
		RealInt:    jsonutil.NewValue(123),
		RealFloat:  jsonutil.NewValue(12.3),
		RealObject: jsonutil.NewValue(obj),
		RealArray:  jsonutil.NewValue(arrData),
		Numeric:    numericData,
		Null:       jsonutil.NewValue(nil),
	}

	data, err := json.Marshal(complexData)
	if err != nil {
		b.Fatal(err)
		return
	}

	for i := 0; i < b.N; i++ {
		var c Complex
		err = json.Unmarshal(data, &c)
		if err != nil {
			b.Fatal(err)
			return
		}
	}
}

func TestSample(t *testing.T) {
	type TestCase struct {
		Name  string
		Value string
	}

	testCases := []TestCase{
		{
			Name:  "string",
			Value: `"string"`,
		},
		{
			Name:  "numeric",
			Value: `1`,
		},
		{
			Name:  "float",
			Value: `1.1`,
		},
		{
			Name:  "object",
			Value: `{"foo": {"foo": "bar"}}`,
		},
		{
			Name:  "multi type array",
			Value: `["array", 1, {"foo": "bar"}]`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			var value jsonutil.Value
			err := json.Unmarshal([]byte(testCase.Value), &value)
			assert.NoError(t, err)
			t.Log(value)
		})
	}

}
