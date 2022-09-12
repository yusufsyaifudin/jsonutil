package envmap

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"gopkg.in/yaml.v3"
)

type S struct {
	ValStr StrOrArr  `json:"val_str" yaml:"val_str" bson:"val_str"`
	PtrStr *StrOrArr `json:"ptr_str" yaml:"ptr_str" bson:"ptr_str"`

	ValArr StrOrArr  `json:"val_arr" yaml:"val_arr" bson:"val_arr"`
	PtrArr *StrOrArr `json:"ptr_arr" yaml:"ptr_arr" bson:"ptr_arr"`
}

var (
	fixtureJson                = `{"val_str":"${VAR}","ptr_str":"${VAR}","val_arr":["${VAR1}","${VAR2}"],"ptr_arr":["${VAR1}","${VAR2}"]}`
	fixtureJsonQuoted          = `{"val_str":"\"str\"","ptr_str":"${VAR}","val_arr":["${VAR1}","${VAR2}"],"ptr_arr":["${VAR1}","${VAR2}"]}`
	fixtureJsonMultilineQuoted = `{"val_str":"\"str\n newline\"","ptr_str":"${VAR}","val_arr":["${VAR1}","${VAR2}"],"ptr_arr":["${VAR1}","${VAR2}"]}`

	fixtureYaml = `
val_str: ${VAR}
ptr_str: ${VAR}
val_arr:
    - ${VAR1}
    - ${VAR2}
ptr_arr:
    - ${VAR1}
    - ${VAR2}
`

	fixtureYamlSingleQuote = `
val_str: '${VAR}'
ptr_str: '${VAR}'
val_arr:
    - '${VAR1}'
    - '${VAR2}'
ptr_arr:
    - '${VAR1}'
    - '${VAR2}'
`

	fixtureYamlDoubleQuote = `
val_str: "${VAR}"
ptr_str: "${VAR}"
val_arr:
    - "${VAR1}"
    - "${VAR2}"
ptr_arr:
    - "${VAR1}"
    - "${VAR2}"
`

	fixtureYamlDoubleComplex = `
val_str: "my string \"is quoted\""
ptr_str: "my string \"is quoted\""
val_arr:
    - "my string \"is quoted\""
    - 'my string "is quoted"'
    - |
      multi line string 
      "quoted"
      string
ptr_arr:
    - "my string \"is quoted\""
    - 'my string "is quoted"'
    - |
      multi line string 
      "quoted"
      string
`

	fixtureYamlDoubleComplexExpected = `
val_str: my string "is quoted"
ptr_str: my string "is quoted"
val_arr:
    - my string "is quoted"
    - my string "is quoted"
    - "multi line string \n\"quoted\"\nstring\n"
ptr_arr:
    - my string "is quoted"
    - my string "is quoted"
    - "multi line string \n\"quoted\"\nstring\n"
`
)

func TestStrOrArr_JSON(t *testing.T) {
	testCases := []struct {
		Name   string
		String string
	}{
		{
			Name:   "quoted",
			String: fixtureJsonQuoted,
		},
		{
			Name:   "normal",
			String: fixtureJson,
		},
		{
			Name:   "multiline",
			String: fixtureJsonMultilineQuoted,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			var s S
			err := json.Unmarshal([]byte(testCase.String), &s)
			assert.NoError(t, err)

			//t.Logf("unmarshal fixture %+v\n", s)

			sBytes, err := json.Marshal(s)
			assert.NotNil(t, sBytes)
			assert.NoError(t, err)
			assert.EqualValues(t, testCase.String, string(sBytes))

			var newS S
			err = json.Unmarshal(sBytes, &newS)
			assert.NoError(t, err)
			assert.EqualValues(t, s, newS)
		})
	}
}

func TestStrOrArr_YAML(t *testing.T) {
	testCases := []struct {
		Name           string
		Input          string
		ExpectedOutput string
	}{
		{
			Name:           "normal",
			Input:          fixtureYaml,
			ExpectedOutput: fixtureYaml,
		},
		{
			Name:           "single quote",
			Input:          fixtureYamlSingleQuote,
			ExpectedOutput: fixtureYaml, // even the input is single quote, it always be generated as non-quoted by default
		},
		{
			Name:           "double quote",
			Input:          fixtureYamlDoubleQuote,
			ExpectedOutput: fixtureYaml, // even the input is double quote, it always be generated as non-quoted by default
		},
		{
			Name:           "complex",
			Input:          fixtureYamlDoubleComplex,
			ExpectedOutput: fixtureYamlDoubleComplexExpected,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			var s S
			err := yaml.Unmarshal([]byte(testCase.Input), &s)
			assert.NoError(t, err)

			//t.Logf("unmarshal fixture %+v\n", s)

			sBytes, err := yaml.Marshal(s)
			assert.NotNil(t, sBytes)
			assert.NoError(t, err)
			assert.EqualValues(t, strings.TrimPrefix(testCase.ExpectedOutput, "\n"), string(sBytes))

			//t.Logf("yaml: \n%s\n", string(sBytes))

			var newS S
			err = yaml.Unmarshal(sBytes, &newS)
			assert.NoError(t, err)
			assert.EqualValues(t, s, newS)
		})
	}
}

func TestStrOrArr_BSON(t *testing.T) {
	testCases := []struct {
		Name string
		Data S
	}{
		{
			Name: "simple",
			Data: S{
				ValStr: *String("${VAR}"),
				PtrStr: String("${VAR}"),
				ValArr: *StringArray([]string{"${VAR1}", "${VAR2}"}),
				PtrArr: StringArray([]string{"${VAR1}", "${VAR2}"}),
			},
		},

		{
			Name: "multiline",
			Data: S{
				ValStr: *String("\n new \n line"),
				PtrStr: String("\n new \n line"),
				ValArr: *StringArray([]string{"\n new \n line", "${VAR2}"}),
				PtrArr: StringArray([]string{"\n new \n line", "${VAR2}"}),
			},
		},

		{
			Name: "quoted",
			Data: S{
				ValStr: *String("\"quoted\""),
				PtrStr: String("\"quoted\""),
				ValArr: *StringArray([]string{"\"quoted\"", "${VAR2}"}),
				PtrArr: StringArray([]string{"\"quoted\"", "${VAR2}"}),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			dataBytes, err := bson.Marshal(testCase.Data)
			assert.NotNil(t, dataBytes)
			assert.NoError(t, err)

			var actual S
			err = bson.Unmarshal(dataBytes, &actual)

			//t.Logf("bytes  %T: %s", testCase.Data, dataBytes)
			//t.Logf("actual %T: \n%+v\n", actual, actual)

			assert.NoError(t, err)
			assert.EqualValues(t, testCase.Data, actual)
		})
	}
}
