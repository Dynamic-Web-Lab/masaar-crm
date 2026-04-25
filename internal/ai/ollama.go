package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

func NewClient(baseURL, model string) *Client {
	return &Client{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
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

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	body, _ := json.Marshal(ollamaRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ollama: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
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

// ScoreLead asks the LLM to score a lead from 0–100 based on context.
func (c *Client) ScoreLead(ctx context.Context, contactName, conversationSummary, source string) (string, error) {
	prompt := fmt.Sprintf(`You are a UAE real estate & B2B sales expert.

Score this lead from 0 to 100 and explain why in 2-3 sentences.
Return JSON: {"score": <int>, "reasoning": "<text>"}

Lead details:
- Name: %s
- Source: %s
- Conversation summary: %s

Respond only with the JSON object.`, contactName, source, conversationSummary)

	return c.Generate(ctx, prompt)
}

// DraftReply generates a context-aware WhatsApp reply suggestion.
func (c *Client) DraftReply(ctx context.Context, contactName, language, threadSummary string) (string, error) {
	lang := "English"
	if language == "ar" {
		lang = "Arabic"
	}

	prompt := fmt.Sprintf(`You are a professional UAE sales agent.

Write a warm, concise WhatsApp follow-up message in %s for this contact.
Keep it under 100 words. Be professional but friendly.

Contact: %s
Conversation context: %s

Reply only with the message text, nothing else.`, lang, contactName, threadSummary)

	return c.Generate(ctx, prompt)
}

// SummarizeThread creates a brief summary of a WhatsApp conversation.
func (c *Client) SummarizeThread(ctx context.Context, messages []string) (string, error) {
	if len(messages) == 0 {
		return "", nil
	}

	conversation := ""
	for i, m := range messages {
		conversation += fmt.Sprintf("%d. %s\n", i+1, m)
	}

	prompt := fmt.Sprintf(`Summarize this WhatsApp sales conversation in 2-3 sentences.
Focus on: customer interest, concerns raised, and next action needed.

Conversation:
%s

Summary:`, conversation)

	return c.Generate(ctx, prompt)
}

// ParseIntent extracts actionable intent from a message
// Returns structured JSON with detected intent type and parameters
func (c *Client) ParseIntent(ctx context.Context, message string) (string, error) {
	prompt := fmt.Sprintf(`You are a UAE real estate sales intent analyzer.

Analyze this WhatsApp message and extract the customer's intent.
Return JSON ONLY: {
  "intent": "property_search|yield_inquiry|price_check|comparables|amenities|contact_request|proposal_request|other",
  "property_type": "2BR|3BR|villa|townhouse|studio|commercial|null",
  "area": "Marina|Downtown|JBR|DIFC|BusinessBay|JVC|Other area name or null",
  "budget_min": number or null,
  "budget_max": number or null,
  "transaction_type": "buy|rent|both|null",
  "urgency": "urgent|asap|flexible|null",
  "next_action": "search_properties|analyze_yield|get_comparables|send_proposal|schedule_call|null"
}

Message: "%s"

Return only valid JSON, no other text.`, message)

	return c.Generate(ctx, prompt)
}

// EnrichLead extracts lead information from a conversation
// Returns structured data about budget, interests, timeline, property preferences
func (c *Client) EnrichLead(ctx context.Context, contactName, conversation string) (string, error) {
	prompt := fmt.Sprintf(`You are a UAE real estate lead enrichment expert.

Analyze this conversation and extract lead information.
Return JSON ONLY: {
  "property_interests": ["2BR apartment", "villa", "commercial space"],
  "budget": {
    "min_aed": number or null,
    "max_aed": number or null,
    "currency": "AED"
  },
  "timeline": "immediate|1-3 months|3-6 months|flexible|null",
  "transaction_type": "buy|rent|both",
  "preferred_areas": ["Marina", "Downtown"],
  "family_size": number or null,
  "investment_focused": true or false,
  "pain_points": ["price", "location", "amenities"],
  "next_steps": "site visit|proposal|more info|call back",
  "lead_quality": "hot|warm|cold",
  "enrichment_notes": "Key insights for agent"
}

Contact: %s
Conversation: %s

Return only valid JSON.`, contactName, conversation)

	return c.Generate(ctx, prompt)
}

// SuggestAction recommends next action based on message and conversation
// Used for agent assist feature
func (c *Client) SuggestAction(ctx context.Context, message, threadSummary string) (string, error) {
	prompt := fmt.Sprintf(`You are a UAE real estate sales coach.

Based on the customer message and conversation, suggest the best next action for the agent.
Return JSON ONLY: {
  "action": "property_search|send_proposal|schedule_viewing|call_customer|send_comparables|analyze_yield|get_market_data|send_invoice",
  "reasoning": "Brief explanation why this action",
  "suggested_message": "Optional: Sample response the agent could send",
  "properties_to_show": ["property_id_1", "property_id_2"] or null
}

Customer message: "%s"
Conversation context: %s

Return only valid JSON.`, message, threadSummary)

	return c.Generate(ctx, prompt)
}
