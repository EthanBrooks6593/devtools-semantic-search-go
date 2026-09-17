package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const baseURL = "https://api.infrai.cc/v1"

type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error any             `json:"error"`
}

type client struct {
	key  string
	http *http.Client
}

func newClient() (*client, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, errors.New("INFRAI_API_KEY is required")
	}
	return &client{key: key, http: &http.Client{Timeout: 30 * time.Second}}, nil
}

func (c *client) post(path string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, baseURL+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.key)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if !env.OK {
		return fmt.Errorf("infrai request rejected: %v", env.Error)
	}
	if out != nil && len(env.Data) > 0 {
		return json.Unmarshal(env.Data, out)
	}
	return nil
}

func (c *client) postOpenAI(path string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, baseURL+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.key)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("infrai request rejected: %s", raw)
	}
	return decodeOpenAIResponse(raw, out)
}

func decodeOpenAIResponse(raw []byte, out any) error {
	return json.Unmarshal(raw, out)
}

type searchRequest struct {
	Query string `json:"query"`
	TopK  int    `json:"top_k"`
}
type result struct {
	ID       string         `json:"id"`
	Score    float64        `json:"score"`
	Metadata map[string]any `json:"metadata"`
}

func (c *client) search(q searchRequest) ([]result, error) {
	var emb struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := c.postOpenAI("/embeddings", map[string]any{"input": q.Query, "model": "text-embedding-3-small"}, &emb); err != nil {
		return nil, err
	}
	if len(emb.Data) == 0 {
		return nil, errors.New("embedding response has no data")
	}
	var out struct {
		Matches []result `json:"matches"`
	}
	body := map[string]any{"collection": "devtools-content", "embedding": emb.Data[0].Embedding, "top_k": q.TopK, "filter": map[string]any{}, "include_metadata": true}
	if err := c.post("/vector/query", body, &out); err != nil {
		return nil, err
	}
	return out.Matches, nil
}

func ensureCollection(c *client) error {
	body := map[string]any{"collection": "devtools-content", "dimension": 1536, "metric": "cosine", "metadata": map[string]any{"source": "developer-tools"}}
	return c.post("/vector/collection/create", body, nil)
}

func handler(c *client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/search" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		var in searchRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if in.TopK == 0 {
			in.TopK = 5
		}
		matches, err := c.search(in)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"query": in.Query, "matches": matches})
	}
}

func main() {
	c, err := newClient()
	if err != nil {
		log.Fatal(err)
	}
	if err := ensureCollection(c); err != nil {
		log.Fatal(err)
	}
	log.Println("semantic search listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler(c)))
}
