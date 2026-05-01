package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const geminiBaseURL = "https://generativelanguage.googleapis.com/v1beta/models"

// NewGeminiClient returns a Client backed by Google Gemini.
// model is typically "gemini-2.0-flash" or "gemini-1.5-pro".
func NewGeminiClient(apiKey, model string) *Client {
	return &Client{provider: &geminiProvider{
		apiKey: apiKey,
		model:  model,
		http:   &http.Client{Timeout: 60 * time.Second},
	}}
}

// ── Gemini provider ──────────────────────────────────────────────────────────

type geminiProvider struct {
	apiKey string
	model  string
	http   *http.Client
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *geminiProvider) generate(ctx context.Context, prompt string) (string, error) {
	url := fmt.Sprintf("%s/%s:generateContent?key=%s", geminiBaseURL, p.model, p.apiKey)

	payload := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: prompt}}},
		},
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("gemini: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini: request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	var result geminiResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("gemini: decode response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("gemini: API error: %s", result.Error.Message)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini: empty response")
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}
