package jsonutil_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/yusufsyaifudin/jsonutil"
)

func TestExample(t *testing.T) {
	mask()
	truncate()
}

func mask() {
	jsonStr := `
{
	"user_login": "user_email@example.com",
    "user_password": "this is sensitive information"
}
`

	transform := jsonutil.NewTransformer(jsonutil.Config{
		StringTransformer: func(ctx context.Context, info jsonutil.KVInfo) string {
			if info.Key == "user_password" {
				return "xxx"
			}

			return info.Value
		},
	})

	out, err := transform.TransformBytes(context.Background(), []byte(jsonStr))
	if err != nil {
		panic(err)
	}

	// will return: {"user_login":"user_email@example.com","user_password":"xxx"}
	fmt.Println(string(out))
}

func truncate() {
	jsonStr := `
{
    "the_long_paragraph": "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."
}
`

	transform := jsonutil.NewTransformer(jsonutil.Config{
		StringTransformer: func(ctx context.Context, info jsonutil.KVInfo) string {
			const padding = 20
			const maxChars = 10
			stringVal := info.Value
			length := len(stringVal)

			return fmt.Sprintf("%s **escaped %d chars** %s",
				stringVal[0:padding], length-maxChars, stringVal[length-padding:],
			)
		},
	})

	out, err := transform.TransformBytes(context.Background(), []byte(jsonStr))
	if err != nil {
		panic(err)
	}

	// will return: {"the_long_paragraph":"Lorem ipsum dolor si **escaped 435 chars** anim id est laborum."}
	fmt.Println(string(out))
}
