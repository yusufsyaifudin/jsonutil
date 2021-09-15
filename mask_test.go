package jsonutil_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/yusufsyaifudin/jsonutil"
)

func TestMask(t *testing.T) {
	var data interface{}
	err := json.Unmarshal([]byte(`["ini", {"nest": ["hello", {"foo": {"hello": "world"}}]}]`), &data)
	// err := json.Unmarshal([]byte(`[{"str": {"nest": "ini string"}, "nest": "hello", "nesting": ["ini", "string"]}]`), &data)
	if err != nil {
		t.Error(err)
		return
	}

	mask := jsonutil.NewMasking(jsonutil.Config{Keys: map[string]struct{}{
		"nest": {},
	}})
	out, err := mask.Mask(context.Background(), data)
	if err != nil {
		t.Log(err)
		t.Error(err)
		return
	}

	t.Logf("final output: %v", out)
}

func BenchmarkMasking_Mask(b *testing.B) {
	var data interface{}
	err := json.Unmarshal([]byte(`[{"str": {"nest": "ini string"}, "nest": "hello", "nesting": ["ini", "string"]}]`), &data)
	if err != nil {
		b.Error(err)
		return
	}

	mask := jsonutil.NewMasking(jsonutil.Config{Keys: map[string]struct{}{
		"nest":    {},
		"nesting": {},
	}})

	for i := 0; i < b.N; i++ {
		_, err = mask.Mask(context.Background(), data)
		if err != nil {
			b.Log(err)
			b.Error(err)
			return
		}
	}

}
