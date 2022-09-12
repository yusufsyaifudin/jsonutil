package envmap

import (
	"bytes"
	"regexp"
	"strings"
)

var (
	envRegex        = regexp.MustCompile(`\${[0-9A-Za-z_.]+(:((\${[^}]+})|[^}])+)?}`)
	escapedEnvRegex = regexp.MustCompile(`\${({[0-9A-Za-z_.]+(:((\${[^}]+})|[^}])+)?})}`)
)

// ReplaceEnvVariables will search a blob of data for the pattern `${FOO:bar}`,
// where `FOO` is an environment variable name and `bar` is a default value. The
// `bar` section (including the colon) can be left out if there is no
// appropriate default value for the field.
//
// For each aforementioned pattern found in the blob the contents of the
// respective environment variable will be read and will replace the pattern. If
// the environment variable is empty or does not exist then either the default
// value is used or the field will be left empty.
//
// This is copied version from:
// https://github.com/benthosdev/benthos/blob/0c6a2e570999be7539926cb463d31fe92e166c72/internal/config/env_vars.go#L24
func ReplaceEnvVariables(inBytes []byte, valuesMap map[string]string) []byte {
	replaced := envRegex.ReplaceAllFunc(inBytes, func(content []byte) []byte {
		var value string
		if len(content) > 3 {
			if colonIndex := bytes.IndexByte(content, ':'); colonIndex == -1 {
				value = valuesMap[string(content[2:len(content)-1])]
			} else {
				targetVar := content[2:colonIndex]
				defaultVal := content[colonIndex+1 : len(content)-1]

				value = valuesMap[string(targetVar)]
				if value == "" {
					value = string(defaultVal)
				}
			}
			// Escape newlines, otherwise there's no way that they would work
			// within a config.
			value = strings.ReplaceAll(value, "\n", "\\n")
		}
		return []byte(value)
	})
	replaced = escapedEnvRegex.ReplaceAll(replaced, []byte("$$$1"))
	return replaced
}
