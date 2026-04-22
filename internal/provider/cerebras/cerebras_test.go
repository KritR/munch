package cerebras

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/krithikr/munch/internal/provider"
)

func TestClientGenerateParsesStructuredOutput(t *testing.T) {
	client := Client{
		BaseURL:    "https://example.test",
		APIKey:     "test-key",
		Model:      "test-model",
		Timeout:    2 * time.Second,
		MaxRetries: 1,
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.Path != "/v1/chat/completions" {
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
				if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
					t.Fatalf("unexpected authorization header: %q", got)
				}

				resp := map[string]any{
					"choices": []map[string]any{
						{
							"message": map[string]any{
								"content": `{"suggestions":[{"command":"rg -n TODO .","description":"Search for TODOs","risk":"low","assumptions":["rg is installed"],"uses_tools":["rg"],"confidence":0.9}]}`,
							},
						},
					},
				}
				var buf bytes.Buffer
				if err := json.NewEncoder(&buf).Encode(resp); err != nil {
					t.Fatalf("Encode() error = %v", err)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(buf.Bytes())),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}

	resp, err := client.Generate(provider.GenerationRequest{
		SystemPrompt:    "system",
		UserPrompt:      "user",
		SuggestionCount: 3,
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(resp.Suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(resp.Suggestions))
	}
	if resp.Suggestions[0].Command != "rg -n TODO ." {
		t.Fatalf("unexpected command: %q", resp.Suggestions[0].Command)
	}
}

func TestClientGenerateInvalidStructuredOutput(t *testing.T) {
	client := Client{
		BaseURL: "https://example.test",
		APIKey:  "test-key",
		Model:   "test-model",
		Timeout: 2 * time.Second,
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				resp := map[string]any{
					"choices": []map[string]any{
						{
							"message": map[string]any{
								"content": `not-json`,
							},
						},
					},
				}
				var buf bytes.Buffer
				if err := json.NewEncoder(&buf).Encode(resp); err != nil {
					t.Fatalf("Encode() error = %v", err)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(buf.Bytes())),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}

	if _, err := client.Generate(provider.GenerationRequest{
		SystemPrompt: "system",
		UserPrompt:   "user",
	}); err == nil {
		t.Fatal("expected invalid structured output error")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
