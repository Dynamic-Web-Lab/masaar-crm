package bos24

// IntegrationClient calls the BOS24 CRM integration API at api.buyorsell24.com.
// This is separate from Client (data.buyorsell24.com — market intelligence).
// The integration API requires a per-company key provisioned by the BOS24 admin.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const IntegrationBaseURL = "https://api.buyorsell24.com/api"

// ── Domain types ──────────────────────────────────────────────────────────────

type BOS24Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type BOS24City struct {
	ID   int    `json:"id"`
	Slug string `json:"slug"`
}

type BOS24Seller struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

type BOS24Image struct {
	URL       string `json:"url"`
	URLLarge  string `json:"url_large"`
	URLMedium string `json:"url_medium"`
	URLThumb  string `json:"url_thumb"`
	SortOrder int    `json:"sort_order"`
}

type BOS24Listing struct {
	UUID        string      `json:"uuid"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Price       float64     `json:"price"`
	Currency    string      `json:"currency"`
	Status      string      `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Category    BOS24Category `json:"category"`
	City        BOS24City   `json:"city"`
	Seller      BOS24Seller `json:"seller"`
	PrimaryImage *BOS24Image `json:"primary_image"`
	Images      []BOS24Image `json:"images"`
}

type BOS24RegisteredUser struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

type BOS24InquiryListing struct {
	UUID   string `json:"uuid"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

type BOS24Inquiry struct {
	ID            int                  `json:"id"`
	BuyerName     string               `json:"buyer_name"`
	BuyerEmail    string               `json:"buyer_email"`
	BuyerPhone    string               `json:"buyer_phone"`
	Message       string               `json:"message"`
	CreatedAt     time.Time            `json:"created_at"`
	Listing       BOS24InquiryListing  `json:"listing"`
	RegisteredUser *BOS24RegisteredUser `json:"registered_user"`
}

type BOS24User struct {
	UUID             string  `json:"uuid"`
	Name             string  `json:"name"`
	CompanyName      string  `json:"company_name"`
	IsVerified       bool    `json:"is_verified"`
	WhatsAppVerified bool    `json:"whatsapp_verified"`
	PhoneVerified    bool    `json:"phone_verified"`
	TrustScore       float64 `json:"trust_score"`
	AccountType      string  `json:"account_type"`
	ListingsCount    int     `json:"listings_count"`
	CreatedAt        time.Time `json:"created_at"`
}

type BOS24PageMeta struct {
	CurrentPage int `json:"current_page"`
	LastPage    int `json:"last_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
}

type BOS24ListingsPage struct {
	Data []BOS24Listing `json:"data"`
	Meta BOS24PageMeta  `json:"meta"`
}

type BOS24InquiriesPage struct {
	Data []BOS24Inquiry `json:"data"`
	Meta BOS24PageMeta  `json:"meta"`
}

type BOS24WebhookRegistration struct {
	ID     int      `json:"id"`
	URL    string   `json:"url"`
	Events []string `json:"events"`
	Secret string   `json:"secret"`
}

// Inbound webhook payload shapes

type BOS24WebhookResource struct {
	Type string `json:"type"`
	UUID string `json:"uuid"`
	ID   int    `json:"id"`
}

type BOS24WebhookEvent struct {
	Event      string               `json:"event"`
	OccurredAt time.Time            `json:"occurred_at"`
	Resource   BOS24WebhookResource `json:"resource"`
	Data       json.RawMessage      `json:"data"` // listing or inquiry shape
}

// ── Client ────────────────────────────────────────────────────────────────────

type IntegrationClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewIntegrationClient(apiKey string) *IntegrationClient {
	return &IntegrationClient{
		baseURL: IntegrationBaseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// FetchListings returns a page of BOS24 marketplace listings.
// status: "published", "active", etc. — empty = all.
// updatedSince: ISO 8601 date string — empty = no filter.
func (c *IntegrationClient) FetchListings(ctx context.Context, status, updatedSince string, page, perPage int) (*BOS24ListingsPage, error) {
	params := url.Values{}
	if status != "" {
		params.Set("status", status)
	}
	if updatedSince != "" {
		params.Set("updated_since", updatedSince)
	}
	if perPage <= 0 {
		perPage = 100
	}
	params.Set("per_page", fmt.Sprintf("%d", perPage))
	params.Set("page", fmt.Sprintf("%d", page))

	var result BOS24ListingsPage
	if err := c.get(ctx, "/v1/integrations/listings", params, &result); err != nil {
		return nil, fmt.Errorf("bos24 FetchListings: %w", err)
	}
	return &result, nil
}

// FetchListing returns a single BOS24 listing by UUID.
func (c *IntegrationClient) FetchListing(ctx context.Context, uuid string) (*BOS24Listing, error) {
	var wrapper struct {
		Data BOS24Listing `json:"data"`
	}
	if err := c.get(ctx, "/v1/integrations/listings/"+uuid, url.Values{}, &wrapper); err != nil {
		return nil, fmt.Errorf("bos24 FetchListing %s: %w", uuid, err)
	}
	return &wrapper.Data, nil
}

// FetchInquiries returns a page of buyer inquiries, newest first.
// createdSince: ISO 8601 date string — empty = no filter.
func (c *IntegrationClient) FetchInquiries(ctx context.Context, createdSince string, page, perPage int) (*BOS24InquiriesPage, error) {
	params := url.Values{}
	if createdSince != "" {
		params.Set("created_since", createdSince)
	}
	if perPage <= 0 {
		perPage = 100
	}
	params.Set("per_page", fmt.Sprintf("%d", perPage))
	params.Set("page", fmt.Sprintf("%d", page))

	var result BOS24InquiriesPage
	if err := c.get(ctx, "/v1/integrations/inquiries", params, &result); err != nil {
		return nil, fmt.Errorf("bos24 FetchInquiries: %w", err)
	}
	return &result, nil
}

// FetchUser returns the public profile for a BOS24 user by UUID.
func (c *IntegrationClient) FetchUser(ctx context.Context, userUUID string) (*BOS24User, error) {
	var wrapper struct {
		Data BOS24User `json:"data"`
	}
	if err := c.get(ctx, "/v1/integrations/users/"+userUUID, url.Values{}, &wrapper); err != nil {
		return nil, fmt.Errorf("bos24 FetchUser %s: %w", userUUID, err)
	}
	return &wrapper.Data, nil
}

// RegisterWebhook registers a webhook URL with BOS24 for the given events.
// events: e.g. ["listing.created","listing.published","listing.updated","inquiry.created"]
func (c *IntegrationClient) RegisterWebhook(ctx context.Context, webhookURL string, events []string, secret string) (*BOS24WebhookRegistration, error) {
	body := map[string]interface{}{
		"url":    webhookURL,
		"events": events,
		"secret": secret,
	}
	var wrapper struct {
		Data BOS24WebhookRegistration `json:"data"`
	}
	if err := c.post(ctx, "/v1/integrations/webhooks/register", body, &wrapper); err != nil {
		return nil, fmt.Errorf("bos24 RegisterWebhook: %w", err)
	}
	return &wrapper.Data, nil
}

// ── HTTP helpers ──────────────────────────────────────────────────────────────

func (c *IntegrationClient) get(ctx context.Context, path string, params url.Values, out interface{}) error {
	fullURL := c.baseURL + path
	if enc := params.Encode(); enc != "" {
		fullURL += "?" + enc
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")
	return c.do(req, out)
}

func (c *IntegrationClient) post(ctx context.Context, path string, body interface{}, out interface{}) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return c.do(req, out)
}

func (c *IntegrationClient) do(req *http.Request, out interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("bos24 rate limit exceeded (429) — retry after %s", resp.Header.Get("Retry-After"))
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("bos24 api error: status %d — %s", resp.StatusCode, string(bodyBytes))
	}

	return json.Unmarshal(bodyBytes, out)
}
