package envmap

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEnvVarString(t *testing.T) {
	nonUtf8 := &bytes.Buffer{}
	nonUtf8.WriteString("${")
	nonUtf8.WriteString("ENV")
	nonUtf8.WriteByte(216)
	nonUtf8.WriteByte(1)
	nonUtf8.WriteByte(220)
	nonUtf8.WriteByte(55)

	nonUtf8.WriteString("VAR")
	nonUtf8.WriteString("}")

	testCases := []struct {
		String        string
		ExpectedKey   string
		ExpectedKind  Kind
		ExpectedError bool
	}{
		{
			// length must minimum 3
			String:        "${}",
			ExpectedKey:   "",
			ExpectedKind:  kindUnknown,
			ExpectedError: true,
		},
		{
			// must have prefix ${
			String:        "ENV_VAR}",
			ExpectedKey:   "",
			ExpectedKind:  kindUnknown,
			ExpectedError: true,
		},
		{
			// must have suffix }
			String:        "${ENV_VAR",
			ExpectedKey:   "",
			ExpectedKind:  kindUnknown,
			ExpectedError: true,
		},
		{
			String:        nonUtf8.String(),
			ExpectedKey:   "",
			ExpectedKind:  kindUnknown,
			ExpectedError: true,
		},
		{
			// not valid, must starts with alphabet
			String:        "${1A}",
			ExpectedKey:   "",
			ExpectedKind:  kindUnknown,
			ExpectedError: true,
		},
		{
			// not valid, must starts with alphabet
			String:        "${_A}",
			ExpectedKey:   "",
			ExpectedKind:  kindUnknown,
			ExpectedError: true,
		},
		{
			// not valid, must ends with no _
			String:        "${A_}",
			ExpectedKey:   "",
			ExpectedKind:  kindUnknown,
			ExpectedError: true,
		},
		{
			// not valid, must use :[],
			// err: string contains non alphanumeric character
			String:        "${A[]}",
			ExpectedKey:   "",
			ExpectedKind:  kindUnknown,
			ExpectedError: true,
		},

		// positive
		{
			String:        "${A}",
			ExpectedKey:   "A",
			ExpectedKind:  KindString,
			ExpectedError: false,
		},
		{
			String:        "${A:[]}",
			ExpectedKey:   "A",
			ExpectedKind:  KindArray,
			ExpectedError: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.String, func(t *testing.T) {
			key, kind, err := IsEnvVarString(context.Background(), testCase.String)
			if testCase.ExpectedError {
				assert.Empty(t, key)
				assert.Equal(t, testCase.ExpectedKind, kind)
				assert.Error(t, err)
				return
			}

			assert.Equal(t, testCase.ExpectedKey, key)
			assert.Equal(t, testCase.ExpectedKind, kind)
			assert.NoError(t, err)
		})
	}
}

func TestMapValue(t *testing.T) {
	testCases := []struct {
		Name          string
		StrOrArr      *StrOrArr
		Values        map[string]string
		Expected      *StrOrArr
		ExpectedError bool
	}{
		{
			Name:          "nil values",
			StrOrArr:      nil,
			Values:        nil,
			Expected:      nil,
			ExpectedError: true,
		},
		{
			Name:          "not mapped",
			StrOrArr:      String("KAFKA_BROKERS"),
			Values:        nil,
			Expected:      String("KAFKA_BROKERS"),
			ExpectedError: false,
		},
		{
			Name:     "simple env",
			StrOrArr: String("${KAFKA_BROKER}"),
			Values: map[string]string{
				"KAFKA_BROKER": "localhost:9092",
			},
			Expected:      String("localhost:9092"),
			ExpectedError: false,
		},
		{
			Name:     "simple env quoted",
			StrOrArr: String("${KAFKA_BROKER}"),
			Values: map[string]string{
				"KAFKA_BROKER": "\"localhost:9092\" another string",
			},
			Expected:      String("\"localhost:9092\" another string"),
			ExpectedError: false,
		},
		{
			Name:     "simple env using comma separated",
			StrOrArr: String("${KAFKA_BROKER}"),
			Values: map[string]string{
				"KAFKA_BROKER": "localhost:9092,localhost:9092",
			},
			Expected:      String("localhost:9092,localhost:9092"),
			ExpectedError: false,
		},
		{
			Name:     "simple env but not found",
			StrOrArr: String("${KAFKA_BROKER}"),
			Values: map[string]string{
				"TYPO_ENV_NAME": "localhost:9092",
			},
			Expected:      String("${KAFKA_BROKER}"),
			ExpectedError: false,
		},
		{
			Name:     "simple env mapped to array",
			StrOrArr: String("${KAFKA_BROKERS:[]}"),
			Values: map[string]string{
				"KAFKA_BROKERS": "localhost:9092,localhost:9093",
			},
			Expected:      StringArray([]string{"localhost:9092", "localhost:9093"}),
			ExpectedError: false,
		},
		{
			Name:     "simple env mapped to array not found",
			StrOrArr: String("${KAFKA_BROKERS:[]}"),
			Values: map[string]string{
				"TYPO_ENV_NAME": "localhost:9092,localhost:9093",
			},
			Expected:      String("${KAFKA_BROKERS:[]}"),
			ExpectedError: false,
		},
		{
			Name:          "empty string",
			StrOrArr:      String(""),
			Values:        nil,
			Expected:      String(""),
			ExpectedError: false,
		},
		{
			Name:     "array without env var",
			StrOrArr: StringArray([]string{"KAFKA_BROKER"}),
			Values: map[string]string{
				"KAFKA_BROKER": "localhost:9092",
			},
			Expected:      StringArray([]string{"KAFKA_BROKER"}),
			ExpectedError: false,
		},
		{
			Name:     "array simple",
			StrOrArr: StringArray([]string{"${KAFKA_BROKER}"}),
			Values: map[string]string{
				"KAFKA_BROKER": "localhost:9092",
			},
			Expected:      StringArray([]string{"localhost:9092"}),
			ExpectedError: false,
		},
		{
			Name:     "array simple but not exist in values map",
			StrOrArr: StringArray([]string{"${KAFKA_BROKER}"}),
			Values: map[string]string{
				"TYPO_ENV_NAME": "localhost:9092",
			},
			Expected:      StringArray([]string{"${KAFKA_BROKER}"}),
			ExpectedError: false,
		},
		{
			Name:     "array simple quoted",
			StrOrArr: StringArray([]string{"${KAFKA_BROKER}"}),
			Values: map[string]string{
				"KAFKA_BROKER": "\"localhost:9092\"",
			},
			Expected:      StringArray([]string{"\"localhost:9092\""}),
			ExpectedError: false,
		},
		{
			Name:     "array contains env var type array will not be mapped",
			StrOrArr: StringArray([]string{"${KAFKA_BROKER:[]}"}),
			Values: map[string]string{
				"KAFKA_BROKER": "localhost:9092",
			},
			Expected:      StringArray([]string{"${KAFKA_BROKER:[]}"}),
			ExpectedError: false,
		},
		{
			Name:     "array values and var",
			StrOrArr: StringArray([]string{"${KAFKA_BROKER}", "localhost:9093"}),
			Values: map[string]string{
				"KAFKA_BROKER": "localhost:9092",
			},
			Expected:      StringArray([]string{"localhost:9092", "localhost:9093"}),
			ExpectedError: false,
		},
		{
			Name:          "unknown type",
			StrOrArr:      &StrOrArr{str: "not-nil", arrStr: []string{"value"}},
			Values:        nil,
			Expected:      nil,
			ExpectedError: true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			actual, err := MapValue(context.Background(), testCase.StrOrArr, testCase.Values)
			if testCase.ExpectedError {
				assert.Empty(t, actual)
				assert.Error(t, err)
				return
			}

			assert.Equal(t, testCase.Expected, actual)
			assert.NoError(t, err)
		})
	}

}

func TestLabelCleaner(t *testing.T) {
	testCases := []struct {
		String   string
		Expected string
	}{
		{
			String:   "3333",
			Expected: "333",
		},
		{
			String:   "3abc",
			Expected: "abc",
		},
		{
			String:   "abc3",
			Expected: "abc3",
		},
		{
			String:   "abc.def",
			Expected: "abc__def",
		},
		{
			String:   "ABC.def",
			Expected: "abc__def",
		},
		{
			String:   "abc_def",
			Expected: "abc_def",
		},
		{
			String:   "_abc_def",
			Expected: "abc_def",
		},
		{
			String:   "abc_def_",
			Expected: "abc_def",
		},
		{
			String:   "${ABC_DEF}",
			Expected: "abc_def",
		},
		{
			String:   "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz",
			Expected: "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.String, func(t *testing.T) {
			actual := LabelCleaner(tc.String)
			assert.Equal(t, tc.Expected, actual)
		})
	}
}
