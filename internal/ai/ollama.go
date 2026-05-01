package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// generator is the single method both Ollama and Gemini providers must satisfy.
type generator interface {
	generate(ctx context.Context, prompt string) (string, error)
}

// Client is the AI facade used by all handlers. It delegates raw generation to
// whichever provider was selected at startup (Ollama or Gemini).
type Client struct {
	provider generator
}

// NewClient returns a Client backed by a local Ollama instance.
func NewClient(baseURL, model string) *Client {
	return &Client{provider: &ollamaProvider{
		baseURL: baseURL,
		model:   model,
		http:    &http.Client{Timeout: 60 * time.Second},
	}}
}

// Generate delegates to the underlying provider.
func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	return c.provider.generate(ctx, prompt)
}

// ── Ollama provider ──────────────────────────────────────────────────────────

type ollamaProvider struct {
	baseURL string
	model   string
	http    *http.Client
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func (p *ollamaProvider) generate(ctx context.Context, prompt string) (string, error) {
	body, _ := json.Marshal(ollamaRequest{Model: p.model, Prompt: prompt, Stream: false})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ollama: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama: request failed: %w", err)
	}
	defer resp.Body.Close()

	var result ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("ollama: decode response: %w", err)
	}
	return result.Response, nil
}
