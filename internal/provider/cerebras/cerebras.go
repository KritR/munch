package cerebras

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/krithikr/munch/internal/protocol"
	"github.com/krithikr/munch/internal/provider"
)

type Client struct {
	BaseURL    string
	APIKey     string
	Model      string
	Timeout    time.Duration
	MaxRetries int
	HTTPClient *http.Client
}

func (c Client) Name() string {
	return "cerebras"
}
func (c Client) Generate(req provider.GenerationRequest) (provider.Response, error) {
	baseURL := strings.TrimRight(c.BaseURL, "/")
	if baseURL == "" {
		return provider.Response{}, fmt.Errorf("base URL is required")
	}
	if c.APIKey == "" {
		return provider.Response{}, fmt.Errorf("api key is required")
	}
	if c.Model == "" {
		return provider.Response{}, fmt.Errorf("model is required")
	}

	timeout := c.Timeout
	if timeout <= 0 {
		timeout = 4 * time.Second
	}

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}

	payload := chatCompletionRequest{
		Model: c.Model,
		Messages: []message{
			{Role: "system", Content: req.SystemPrompt},
			{Role: "user", Content: req.UserPrompt},
		},
		ResponseFormat: responseFormat{
			Type: "json_schema",
			JSONSchema: jsonSchemaEnvelope{
				Name:   "munch_suggestions",
				Strict: true,
				Schema: suggestionsJSONSchema(),
			},
		},
		Stream: boolPtr(false),
	}

	var lastErr error
	attempts := max(1, c.MaxRetries+1)
	for attempt := 0; attempt < attempts; attempt++ {
		resp, err := c.doRequest(httpClient, baseURL, payload)
		if err == nil {
			return resp, nil
		}
		lastErr = err
	}
	return provider.Response{}, lastErr
}

func (c Client) doRequest(httpClient *http.Client, baseURL string, payload chatCompletionRequest) (provider.Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return provider.Response{}, err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return provider.Response{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		return provider.Response{}, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return provider.Response{}, fmt.Errorf("cerebras request failed: status=%d body=%s", res.StatusCode, strings.TrimSpace(string(raw)))
	}

	var chatResp chatCompletionResponse
	if err := json.NewDecoder(res.Body).Decode(&chatResp); err != nil {
		return provider.Response{}, err
	}
	if len(chatResp.Choices) == 0 {
		return provider.Response{}, fmt.Errorf("cerebras response contained no choices")
	}
	if strings.TrimSpace(chatResp.Choices[0].Message.Content) == "" {
		return provider.Response{}, fmt.Errorf("cerebras response contained empty content")
	}

	var content suggestionsEnvelope
	if err := json.Unmarshal([]byte(chatResp.Choices[0].Message.Content), &content); err != nil {
		return provider.Response{}, fmt.Errorf("invalid structured output: %w", err)
	}

	return provider.Response{Suggestions: content.Suggestions}, nil
}

type chatCompletionRequest struct {
	Model          string         `json:"model"`
	Messages       []message      `json:"messages"`
	ResponseFormat responseFormat `json:"response_format"`
	Stream         *bool          `json:"stream,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type       string             `json:"type"`
	JSONSchema jsonSchemaEnvelope `json:"json_schema"`
}

type jsonSchemaEnvelope struct {
	Name   string         `json:"name"`
	Strict bool           `json:"strict"`
	Schema map[string]any `json:"schema"`
}

type chatCompletionResponse struct {
	Choices []choice `json:"choices"`
}

type choice struct {
	Message responseMessage `json:"message"`
}

type responseMessage struct {
	Content string `json:"content"`
}

type suggestionsEnvelope struct {
	Suggestions []protocol.Suggestion `json:"suggestions"`
}

func suggestionsJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"suggestions"},
		"properties": map[string]any{
			"suggestions": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"required":             []string{"command", "description", "risk", "assumptions", "uses_tools"},
					"properties": map[string]any{
						"command":     map[string]any{"type": "string"},
						"description": map[string]any{"type": "string"},
						"risk": map[string]any{
							"type": "string",
							"enum": []string{"low", "medium", "high"},
						},
						"assumptions": map[string]any{
							"type":  "array",
							"items": map[string]any{"type": "string"},
						},
						"uses_tools": map[string]any{
							"type":  "array",
							"items": map[string]any{"type": "string"},
						},
						"confidence": map[string]any{
							"type": []string{"number", "null"},
						},
					},
				},
			},
		},
	}
}

func boolPtr(v bool) *bool {
	return &v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
