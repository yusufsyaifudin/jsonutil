# jsonutil

Golang utility to work with JSON string value.

## Features

### Truncate string inside a JSON

> Limitation of this feature:
>
> The function will not verify whether bytes input and output is a valid JSON string or not.
> To ensure you pass and get the good JSON string, you can do this before and after calling `TruncateJsonValueString`:
> ```go
> var i interface{}
> err := json.Unmarshal([]byte(), &i)
> ```
> 


The string must have a pair of quotes `"`. 
If String contains `"` as a value it must be quoted by `\\`, for example, `"\""`.

This follows [JSON specification](https://www.json.org/json-en.html) where:
* A key must be String, you cannot define key with number like `{1: "value one"}`, but must `{"1": "value one"}`.
* If you need quote string, use `\\`, for example:
```json
{
  "maroon_5:she_will_be_loved": "my favourite lyric is: \"It's not always rainbows and butterflies\""
}
```

Actually you can do that in key string too, but that's uncommon and weird.

```json
{
  "my_\"quote\"": "your quote"
}
```

This useful when you have a large string value in your JSON, for example:

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	
	"github.com/yusufsyaifudin/jsonutil"
)

func main() {
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

```

The key `the_long_paragraph` contains of 445 characters, and you don't want to show all of it contents. 
Using `TruncateJsonValueString(content, 10)` the `jsonStr` will become

```json
{
    "the_long_paragraph": "Lorem **escaped 435 chars at [30:475]** orum."
}
```

If you use `TruncateJsonString`, the key will truncated as well because it is more than 10 chars.

```json
{
    "the_l **escaped 8 chars at [8:26]** graph": "Lorem **escaped 435 chars at [30:475]** orum."
}
```

## Benchmark

```shell
go test -run=. -bench=. -benchmem -cpuprofile=cpu.out -memprofile=mem.out -trace=trace.out ./... | tee bench.txt
```

```shell
goos: darwin
goarch: amd64
pkg: github.com/yusufsyaifudin/jsonutil
cpu: Intel(R) Core(TM) i5-8279U CPU @ 2.40GHz
BenchmarkTruncateJsonString-8              65910             18346 ns/op            8940 B/op        172 allocs/op
BenchmarkTruncateJsonValueString-8         44377             27951 ns/op            8396 B/op        149 allocs/op
PASS
ok      github.com/yusufsyaifudin/jsonutil      4.761s
```