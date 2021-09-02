package jsonutil

import (
	"bytes"
	"context"
	"errors"
	"fmt"
)

const (
	stringToken        = '"'
	escapedStringToken = '\\'
)

// TruncateJsonString will truncate all string in json between two char `"`.
// if the string length more than maxChars. Zero means no truncate will performed.
// Some chars will preserve into the output.
func TruncateJsonString(ctx context.Context, data []byte, maxChars int) ([]byte, error) {
	padding := 20
	if maxChars < padding {
		padding = maxChars / 2
	}

	if maxChars <= 0 {
		maxChars = len(data)
	}

	var (
		buf      = &bytes.Buffer{}
		str      = make([]byte, 0)
		begin    = false
		idxStart = 0
		idxEnd   = 0
	)

	for i := 0; i < len(data); i++ {
		datum := data[i]

		prevIdx := i
		if i != 0 {
			prevIdx = i - 1
		}

		if datum == stringToken && idxStart == 0 {
			idxStart = i + 1
			begin = true
			continue
		}

		if begin {

			// if found matched string token and previous token is not escaped char \
			if datum == stringToken && data[prevIdx] != escapedStringToken {
				// get the string value
				// fmt.Println(string(str))

				buf.WriteByte(stringToken)
				if len(str) >= maxChars {
					buf.WriteString(
						fmt.Sprintf("%s **escaped %d chars at [%d:%d]** %s",
							str[0:padding], len(str)-maxChars, idxStart, idxEnd, str[len(str)-padding:]),
					)
				} else {
					buf.Write(str)
				}

				buf.WriteByte(stringToken)

				begin = false
				str = make([]byte, 0)
				idxStart = 0
				idxEnd = 0
				continue
			}

			if i == len(data)-1 {
				return nil, errors.New("error token is not closed")
			}

			str = append(str, datum)
			idxEnd = i + 1
			continue
		}

		buf.WriteByte(datum)
	}

	return buf.Bytes(), nil
}

// TruncateJsonValueString unlike TruncateJsonString which will truncate all json string either it is key or value,
// this function will only truncate the string if it is a value.
// Key string will not be truncated.
func TruncateJsonValueString(ctx context.Context, data []byte, maxChars int) ([]byte, error) {
	padding := 20
	if maxChars < padding {
		padding = maxChars / 2
	}

	if maxChars <= 0 {
		maxChars = len(data)
	}

	var (
		buf      = &bytes.Buffer{}
		str      = make([]byte, 0)
		begin    = false
		idxStart = 0
		idxEnd   = 0
	)

	for i := 0; i < len(data); i++ {
		datum := data[i]

		prevIdx := i
		if i != 0 {
			prevIdx = i - 1
		}

		if datum == stringToken && idxStart == 0 {
			idxStart = i + 1
			begin = true
			continue
		}

		if begin {

			// if found matched string token and previous token is not escaped char \
			if datum == stringToken && data[prevIdx] != escapedStringToken {
				// get the string value
				// fmt.Println(string(str))

				// ensure that this string is not a key string by looking for : char after this
				isKey := false
				isFoundStringTokenAgain := false
				for j := idxEnd + 1; j < len(data); j++ {
					if data[j] == stringToken {
						isFoundStringTokenAgain = true
						continue
					}

					if !isFoundStringTokenAgain && data[j] == ':' {
						isKey = true
						break
					}
				}

				if isKey {
					buf.WriteByte(stringToken)
					buf.Write(str)
					buf.WriteByte(stringToken)

					begin = false
					str = make([]byte, 0)
					idxStart = 0
					idxEnd = 0
					continue
				}

				buf.WriteByte(stringToken)
				if len(str) >= maxChars {
					buf.WriteString(
						fmt.Sprintf("%s **escaped %d chars at [%d:%d]** %s",
							str[0:padding], len(str)-maxChars, idxStart, idxEnd, str[len(str)-padding:]),
					)
				} else {
					buf.Write(str)
				}

				buf.WriteByte(stringToken)

				begin = false
				str = make([]byte, 0)
				idxStart = 0
				idxEnd = 0
				continue
			}

			if i == len(data)-1 {
				return nil, errors.New("error token is not closed")
			}

			str = append(str, datum)
			idxEnd = i + 1
			continue
		}

		buf.WriteByte(datum)
	}

	return buf.Bytes(), nil
}
