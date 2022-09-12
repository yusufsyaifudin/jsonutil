package envmap

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	regxAlphaNum = regexp.MustCompile(`^[A-Z0-9_]+$`)
)

// IsEnvVarString return whether the str contains regex ${KEY} or ${KEY:[]}.
// If str is using that value, then it will be considered as environment variable,
// and we must treat it like as is it.
//
// If the format matched, we will return the KEY name to get the exact value from environment variable map.
//
// String will be valid as environment variable if this requirements is meet:
// 1. Prefix "${" and Suffix "}"
// 2. Only utf8 characters
// 3. Must not start with number character
// 4. Must only contain uppercase letter and _.
// 5. For type array, the suffix can be (and must be) ":[]}"
// I.e:
// ${KAFKA_BROKERS} = KAFKA_BROKERS, string, nil
// ${KAFKA_BROKERS:[]} = KAFKA_BROKERS, array, nil
// ${KAFKA_BROKERS[]} = empty string, unknown, error
func IsEnvVarString(ctx context.Context, str string) (key string, kind Kind, err error) {

	if len(str) <= 3 {
		key = ""
		err = fmt.Errorf("minimum char of env var is 4")
		return
	}

	if !strings.HasPrefix(str, "${") {
		key = ""
		err = fmt.Errorf("string not starts with '${'")
		return
	}

	if !strings.HasSuffix(str, "}") {
		key = ""
		err = fmt.Errorf("string not ends with '}'")
		return
	}

	key = str[2:]          // take prefix ${
	key = key[:len(key)-1] // take suffix }

	if strings.HasSuffix(key, ":[]") {
		kind = KindArray
		key = key[:len(key)-3] // take suffix :[]
	}

	if !utf8.ValidString(key) {
		key = ""
		err = fmt.Errorf("strings for env var cannot contain non-utf8 chars")
		return
	}

	if len(key) > 0 {
		firstChar := key[0]
		switch firstChar {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			key = ""
			err = fmt.Errorf("strings for environment variable cannot starts with number")
			return

		case '_':
			key = ""
			err = fmt.Errorf("strings for environment variable cannot starts with underscore")
			return
		}

		lastChar := key[len(key)-1]
		if lastChar == '_' {
			key = ""
			err = fmt.Errorf("strings for environment variable cannot ends with underscore")
			return
		}
	}

	if !regxAlphaNum.MatchString(key) {
		key = ""
		err = fmt.Errorf("string contains non alphanumeric character")
		return
	}

	// only set Kind if unset
	if kind == kindUnknown {
		kind = KindString
	}

	return
}

// MapValue will return new copied StrOrArr but will replace all string
// with format ${} with the actual value from values map.
// For example, string contains ${KAFKA_BROKERS} and values map:
// map["KAFKA_BROKERS"]="localhost:9092,localhost:9093"
// It then will be new StrOrArr with values StrOrArr{arrStr: [localhost:9092,localhost:9093]}
//
// values key must only contain exact string similar like we define in Environment Variable on unix system.
// To define array, use comma separator between fields.
// To define array:
// * KAFKA_BROKERS=localhost:9092,localhost:9093 (simple, preferred)
// * KAFKA_BROKERS="localhost:9092","localhost:9093" (wrong example) the whole string "localhost:9092" will be treated as value, not localhost:9092
func MapValue(ctx context.Context, s *StrOrArr, values map[string]string) (mapped *StrOrArr, err error) {
	if s == nil {
		err = fmt.Errorf("nil StrOrArr object")
		return
	}

	if values == nil {
		values = map[string]string{}
	}

	mapped = &StrOrArr{
		str:    s.str,
		arrStr: s.arrStr,
	}

	switch s.Kind() {
	case KindString:
		var (
			key  string
			kind Kind
		)

		key, kind, err = IsEnvVarString(ctx, s.str)
		if err != nil {
			// if error is not nil, then consider it as an actual value
			mapped.str = s.str
			mapped.arrStr = nil
			err = nil
			return
		}

		// if not nil, then try to map from values
		switch kind {
		case KindString:
			// if key is not found in values, then it will use default value
			actualValue, exist := values[key]
			if !exist {
				actualValue = s.str
			}

			mapped.str = actualValue
			mapped.arrStr = nil
			return

		case KindArray:
			// if key is not found in values, then it will use default value
			actualValue, exist := values[key]
			if !exist {
				mapped.str = s.str
				mapped.arrStr = nil

				return
			}

			// separator by comma
			mapped.str = ""
			mapped.arrStr = strings.Split(actualValue, ",")
		}

	case KindArray:
		actualArrValues := make([]string, 0)

		for _, str := range s.Array() {
			key, kind, _err := IsEnvVarString(ctx, str)
			if _err != nil {
				// if error is not nil, then consider it as an actual value
				actualArrValues = append(actualArrValues, str)
				continue
			}

			// if not nil, then try to map from values
			switch kind {
			case KindString:
				// if key is not found in values, then it will use default value
				actualValue, exist := values[key]
				if !exist {
					actualValue = str
				}

				actualArrValues = append(actualArrValues, actualValue)

			default:
				// for KindArray still treated as actual value, because we cannot do nested env var.
				// This adds complexity and error-prone.

				actualArrValues = append(actualArrValues, str)
			}
		}

		mapped.str = ""
		mapped.arrStr = actualArrValues

	default:
		mapped = &StrOrArr{}
		err = fmt.Errorf("cannot handle type %+v", s.Kind())
		return
	}

	return
}

func LabelCleaner(str string) string {
	cleaner := labelCleaner(str)
	newLabel := strings.Map(func(r rune) rune {
		return cleaner(r)
	}, str)

	newLabel = strings.ReplaceAll(newLabel, ".", "__")

	if len(newLabel) >= 63 {
		newLabel = newLabel[:63]
	}

	newLabel = strings.TrimSuffix(newLabel, "_")

	return newLabel
}

// Will convert to lowercase string and match with regex
// /^[a-z0-9_]+$/ and must not start with an underscore
func labelCleaner(str string) func(r rune) rune {
	idx := 0
	return func(r rune) rune {
		idx += 1

		var replacement rune
		switch r {
		case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
			'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't',
			'u', 'v', 'w', 'x', 'y', 'z':
			replacement = r

		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// return -1 to drop the string from builder
			// cannot start with number but, can end with number
			if idx == 1 {
				return -1
			}

			replacement = r

		case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
			'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T',
			'U', 'V', 'W', 'X', 'Y', 'Z':
			replacement = unicode.ToLower(r)

		case '_':
			// return -1 to drop the string from builder
			// cannot start or end with underscore
			if idx == 1 || idx == len(str) {
				return -1
			}

			replacement = r

		case '.':
			replacement = '.'

		default:
			// return -1 to drop the string from builder
			// remove unknown char
			return -1
		}

		return replacement
	}
}
