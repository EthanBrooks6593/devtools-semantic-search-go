package main

import (
	"reflect"
	"testing"
)

func TestSearchRequestDefaults(t *testing.T) {
	tests := []struct {
		name string
		in   searchRequest
		want int
	}{
		{"explicit", searchRequest{Query: "vector search endpoint", TopK: 3}, 3},
		{"default", searchRequest{Query: "semantic search api go"}, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.TopK
			if got == 0 {
				got = 5
			}
			if got != tt.want {
				t.Fatalf("top_k=%d, want %d", got, tt.want)
			}
		})
	}
}

func TestDecodeOpenAIEmbeddingResponse(t *testing.T) {
	var got struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := decodeOpenAIResponse([]byte(`{"data":[{"embedding":[0.25,-0.5]}]}`), &got); err != nil {
		t.Fatal(err)
	}
	want := []float64{0.25, -0.5}
	if len(got.Data) != 1 || !reflect.DeepEqual(got.Data[0].Embedding, want) {
		t.Fatalf("embedding=%v, want %v", got.Data, want)
	}
}
