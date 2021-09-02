package jsonutil_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/yusufsyaifudin/jsonutil"
)

const allJSONType = `
{
  "string_only": "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
  "string_quoted": "\"",
  "uint": -1,
  "uint8": -1,
  "uint16": -1,
  "uint32": -1,
  "uint64": -1,
  "int": -1,
  "int8": -1,
  "int16": -1,
  "int32": -1,
  "int64": -1,
  "float32": 1.1,
  "float64": 1.1,
  "array_string": [
    "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
    "Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
    "Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.",
    "Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."
  ],
  "map": {
    "string_only": "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
    "uint": -1,
    "uint8": -1,
    "uint16": -1,
    "uint32": -1,
    "uint64": -1,
    "int": -1,
    "int8": -1,
    "int16": -1,
    "int32": -1,
    "int64": -1,
    "float32": 1.1,
    "float64": 1.1,
    "array_string": [
      "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
      "Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
      "Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.",
      "Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."
    ],
    "map": {
      "foo": "bar"
    }
  }
}
`

func TestTruncateString(t *testing.T) {
	type TestCase struct {
		Name  string
		Input string
		Error bool
	}

	var testCases = []TestCase{
		{
			Name:  "complex json string",
			Input: allJSONType,
			Error: false,
		},
		{
			Name:  "valid escaped string",
			Input: `"\""`,
			Error: false,
		},
		{
			Name:  "valid multiple escaped string",
			Input: `"\"\"\""`,
			Error: false,
		},
		{
			Name:  "valid escaped string space",
			Input: `"\"  \n \"  "`,
			Error: false,
		},
		{
			Name:  "invalid escaped string",
			Input: `"\"\"`,
			Error: true,
		},
		// TODO handle this
		// {
		// 	Name:   "invalid escaped string using space",
		// 	Input:  `"\"\   "`,
		// 	Error:  true,
		// },
		{
			Name:  "invalid string json",
			Input: `""""`,
			Error: false, // TODO must be error if invalid json
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			out, err := jsonutil.TruncateJsonString(context.Background(), []byte(testCase.Input), 0)
			if testCase.Error {
				if err == nil {
					t.Error("must return error, but got no error")
					return
				}

				return
			}

			if string(out) != testCase.Input {
				t.Errorf("\nwant %s\ngot  %s", testCase.Input, out)
			}

			out, err = jsonutil.TruncateJsonValueString(context.Background(), []byte(testCase.Input), 0)
			if testCase.Error {
				if err == nil {
					t.Error("must return error, but got no error")
					return
				}

				return
			}

			if string(out) != testCase.Input {
				t.Errorf("\nwant %s\ngot  %s", testCase.Input, out)
			}

		})
	}
}

func TestTruncateJsonString(t *testing.T) {
	expectedAllStr := `
{
  "s **escaped 8 chars at [6:17]** y": "L **escaped 442 chars at [21:466]** .",
  "s **escaped 10 chars at [472:485]** d": "\"",
  "u **escaped 1 chars at [497:501]** t": -1,
  "u **escaped 2 chars at [511:516]** 8": -1,
  "u **escaped 3 chars at [526:532]** 6": -1,
  "u **escaped 3 chars at [542:548]** 2": -1,
  "u **escaped 3 chars at [558:564]** 4": -1,
  "i **escaped 0 chars at [574:577]** t": -1,
  "i **escaped 1 chars at [587:591]** 8": -1,
  "i **escaped 2 chars at [601:606]** 6": -1,
  "i **escaped 2 chars at [616:621]** 2": -1,
  "i **escaped 2 chars at [631:636]** 4": -1,
  "f **escaped 4 chars at [646:653]** 2": 1.1,
  "f **escaped 4 chars at [664:671]** 4": 1.1,
  "a **escaped 9 chars at [682:694]** g": [
    "L **escaped 120 chars at [704:827]** .",
    "U **escaped 104 chars at [835:942]** .",
    "D **escaped 99 chars at [950:1052]** .",
    "E **escaped 107 chars at [1060:1170]** ."
  ],
  "m **escaped 0 chars at [1180:1183]** p": {
    "s **escaped 8 chars at [1193:1204]** y": "L **escaped 442 chars at [1208:1653]** .",
    "u **escaped 1 chars at [1661:1665]** t": -1,
    "u **escaped 2 chars at [1677:1682]** 8": -1,
    "u **escaped 3 chars at [1694:1700]** 6": -1,
    "u **escaped 3 chars at [1712:1718]** 2": -1,
    "u **escaped 3 chars at [1730:1736]** 4": -1,
    "i **escaped 0 chars at [1748:1751]** t": -1,
    "i **escaped 1 chars at [1763:1767]** 8": -1,
    "i **escaped 2 chars at [1779:1784]** 6": -1,
    "i **escaped 2 chars at [1796:1801]** 2": -1,
    "i **escaped 2 chars at [1813:1818]** 4": -1,
    "f **escaped 4 chars at [1830:1837]** 2": 1.1,
    "f **escaped 4 chars at [1850:1857]** 4": 1.1,
    "a **escaped 9 chars at [1870:1882]** g": [
      "L **escaped 120 chars at [1894:2017]** .",
      "U **escaped 104 chars at [2027:2134]** .",
      "D **escaped 99 chars at [2144:2246]** .",
      "E **escaped 107 chars at [2256:2366]** ."
    ],
    "m **escaped 0 chars at [2380:2383]** p": {
      "f **escaped 0 chars at [2395:2398]** o": "b **escaped 0 chars at [2402:2405]** r"
    }
  }
}
`

	expectedOnlyValue := `
{
  "string_only": "L **escaped 442 chars at [21:466]** .",
  "string_quoted": "\"",
  "uint": -1,
  "uint8": -1,
  "uint16": -1,
  "uint32": -1,
  "uint64": -1,
  "int": -1,
  "int8": -1,
  "int16": -1,
  "int32": -1,
  "int64": -1,
  "float32": 1.1,
  "float64": 1.1,
  "array_string": [
    "L **escaped 120 chars at [704:827]** .",
    "U **escaped 104 chars at [835:942]** .",
    "D **escaped 99 chars at [950:1052]** .",
    "E **escaped 107 chars at [1060:1170]** ."
  ],
  "map": {
    "string_only": "L **escaped 442 chars at [1208:1653]** .",
    "uint": -1,
    "uint8": -1,
    "uint16": -1,
    "uint32": -1,
    "uint64": -1,
    "int": -1,
    "int8": -1,
    "int16": -1,
    "int32": -1,
    "int64": -1,
    "float32": 1.1,
    "float64": 1.1,
    "array_string": [
      "L **escaped 120 chars at [1894:2017]** .",
      "U **escaped 104 chars at [2027:2134]** .",
      "D **escaped 99 chars at [2144:2246]** .",
      "E **escaped 107 chars at [2256:2366]** ."
    ],
    "map": {
      "foo": "b **escaped 0 chars at [2402:2405]** r"
    }
  }
}
`
	out, err := jsonutil.TruncateJsonString(context.Background(), []byte(allJSONType), 3)
	if err != nil {
		t.Error("must return error, but got no error")
		return
	}

	if string(out) != expectedAllStr {
		t.Errorf("\nwant %s\ngot  %s\n", expectedAllStr, out)
		return
	}

	out, err = jsonutil.TruncateJsonValueString(context.Background(), []byte(allJSONType), 3)
	if err != nil {
		t.Error("must return error, but got no error")
		return
	}

	if string(out) != expectedOnlyValue {
		t.Errorf("\nwant %s\ngot  %s\n", expectedOnlyValue, out)
		return
	}
}

func BenchmarkTruncateJsonString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := jsonutil.TruncateJsonString(context.Background(), []byte(allJSONType), 10)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTruncateJsonValueString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := jsonutil.TruncateJsonValueString(context.Background(), []byte(allJSONType), 10)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestExampleTruncateJsonString(t *testing.T) {
	jsonStr := `
{
    "the_long_paragraph": "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."
}
`

	var i interface{}
	err := json.Unmarshal([]byte(jsonStr), &i)
	if err != nil {
		panic(err)
	}

	out, err := jsonutil.TruncateJsonValueString(context.Background(), []byte(jsonStr), 10)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(out, &i)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))
}
