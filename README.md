# jsonutil

Golang utility to work with JSON string value.

## Features

* [x] [Transform string inside a JSON: Mask or Truncate or any string transform](#transform-string-inside-a-json)
* [x] [Handle dynamic unpredictable value using JSON Value](#json-value)

### Transform string inside a JSON

You can either mask a value on certain key or truncate string if meet the max length.

## Example Truncate and Mask String

```go
package main

import (
	"context"
	"fmt"

	"github.com/yusufsyaifudin/jsonutil"
)

func main() {
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
```


### JSON Value

Useful when you consume an API that return inconsistent data type.
For example, sometimes it returns number: `{"amount": 1}`, and at the same time it can return string `{"amount":"20000"}`,
or maybe float number `{"amount": 100.00}` or if you don't lucky enough it can be object `{"amount": {"currency": "IDR", "value": 1200}}`
or array `{"amount": ["IDR", 100]}`.

I know, that API is bad, but when you don't have control to change the API response and forced to consume it,
the only thing you can do is **handle it**.

The `jsonutil.Value` is intended to handle that.

## Benchmark

```shell
go test -run=. -bench=. -benchmem -cpuprofile=cpu.out -memprofile=mem.out -trace=trace.out ./... | tee bench.txt
```

```shell
goos: darwin
goarch: amd64
pkg: github.com/yusufsyaifudin/jsonutil
cpu: Intel(R) Core(TM) i5-8279U CPU @ 2.40GHz
BenchmarkTransformer_Transform/large_array-8              300424              4137 ns/op             688 B/op         41 allocs/op
BenchmarkTransformer_Transform/all_JSON_type-8            140460              7396 ns/op            3584 B/op        110 allocs/op
BenchmarkTransformer_Transform/nested_100_object-8        168968              6386 ns/op             704 B/op         14 allocs/op
BenchmarkTransformer_Transform/nested_1000_object-8        21242             57727 ns/op             723 B/op         14 allocs/op
BenchmarkValue_MarshalJSON-8                              179325              7189 ns/op            2627 B/op         48 allocs/op
BenchmarkValue_UnmarshalJSON-8                             72892             18011 ns/op            5267 B/op        114 allocs/op
PASS
ok      github.com/yusufsyaifudin/jsonutil      15.688s
```