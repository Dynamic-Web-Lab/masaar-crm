package dldapi

// IntegrationClient calls the DLD CRM integration API at dldapi.waqov.com.
// This is separate from Client (dldapi.waqov.com — market intelligence).
// The integration API requires a per-company key provisioned by the DLD admin.

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

const IntegrationBaseURL = "https://dldapi.waqov.com/api"

// ── Domain types ──────────────────────────────────────────────────────────────

type DLDCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type DLDCity struct {
	ID   int    `json:"id"`
	Slug string `json:"slug"`
}

type DLDSeller struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

type DLDImage struct {
	URL       string `json:"url"`
	URLLarge  string `json:"url_large"`
	URLMedium string `json:"url_medium"`
	URLThumb  string `json:"url_thumb"`
	SortOrder int    `json:"sort_order"`
}

type DLDListing struct {
	UUID        string      `json:"uuid"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Price       float64     `json:"price"`
	Currency    string      `json:"currency"`
	Status      string      `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Category    DLDCategory `json:"category"`
	City        DLDCity   `json:"city"`
	Seller      DLDSeller `json:"seller"`
	PrimaryImage *DLDImage `json:"primary_image"`
	Images      []DLDImage `json:"images"`
}

type DLDRegisteredUser struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

type DLDInquiryListing struct {
	UUID   string `json:"uuid"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

type DLDInquiry struct {
	ID            int                  `json:"id"`
	BuyerName     string               `json:"buyer_name"`
	BuyerEmail    string               `json:"buyer_email"`
	BuyerPhone    string               `json:"buyer_phone"`
	Message       string               `json:"message"`
	CreatedAt     time.Time            `json:"created_at"`
	Listing       DLDInquiryListing  `json:"listing"`
	RegisteredUser *DLDRegisteredUser `json:"registered_user"`
}

type DLDUser struct {
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

type DLDPageMeta struct {
	CurrentPage int `json:"current_page"`
	LastPage    int `json:"last_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
}

type DLDListingsPage struct {
	Data []DLDListing `json:"data"`
	Meta DLDPageMeta  `json:"meta"`
}

type DLDInquiriesPage struct {
	Data []DLDInquiry `json:"data"`
	Meta DLDPageMeta  `json:"meta"`
}

type DLDWebhookRegistration struct {
	ID     int      `json:"id"`
	URL    string   `json:"url"`
	Events []string `json:"events"`
	Secret string   `json:"secret"`
}

// Inbound webhook payload shapes

type DLDWebhookResource struct {
	Type string `json:"type"`
	UUID string `json:"uuid"`
	ID   int    `json:"id"`
}

type DLDWebhookEvent struct {
	Event      string               `json:"event"`
	OccurredAt time.Time            `json:"occurred_at"`
	Resource   DLDWebhookResource `json:"resource"`
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

// FetchListings returns a page of DLD marketplace listings.
// status: "published", "active", etc. — empty = all.
// updatedSince: ISO 8601 date string — empty = no filter.
func (c *IntegrationClient) FetchListings(ctx context.Context, status, updatedSince string, page, perPage int) (*DLDListingsPage, error) {
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

	var result DLDListingsPage
	if err := c.get(ctx, "/v1/integrations/listings", params, &result); err != nil {
		return nil, fmt.Errorf("dld FetchListings: %w", err)
	}
	return &result, nil
}

// FetchListing returns a single DLD listing by UUID.
func (c *IntegrationClient) FetchListing(ctx context.Context, uuid string) (*DLDListing, error) {
	var wrapper struct {
		Data DLDListing `json:"data"`
	}
	if err := c.get(ctx, "/v1/integrations/listings/"+uuid, url.Values{}, &wrapper); err != nil {
		return nil, fmt.Errorf("dld FetchListing %s: %w", uuid, err)
	}
	return &wrapper.Data, nil
}

// FetchInquiries returns a page of buyer inquiries, newest first.
// createdSince: ISO 8601 date string — empty = no filter.
func (c *IntegrationClient) FetchInquiries(ctx context.Context, createdSince string, page, perPage int) (*DLDInquiriesPage, error) {
	params := url.Values{}
	if createdSince != "" {
		params.Set("created_since", createdSince)
	}
	if perPage <= 0 {
		perPage = 100
	}
	params.Set("per_page", fmt.Sprintf("%d", perPage))
	params.Set("page", fmt.Sprintf("%d", page))

	var result DLDInquiriesPage
	if err := c.get(ctx, "/v1/integrations/inquiries", params, &result); err != nil {
		return nil, fmt.Errorf("dld FetchInquiries: %w", err)
	}
	return &result, nil
}

// FetchUser returns the public profile for a DLD user by UUID.
func (c *IntegrationClient) FetchUser(ctx context.Context, userUUID string) (*DLDUser, error) {
	var wrapper struct {
		Data DLDUser `json:"data"`
	}
	if err := c.get(ctx, "/v1/integrations/users/"+userUUID, url.Values{}, &wrapper); err != nil {
		return nil, fmt.Errorf("dld FetchUser %s: %w", userUUID, err)
	}
	return &wrapper.Data, nil
}

// RegisterWebhook registers a webhook URL with DLD for the given events.
// events: e.g. ["listing.created","listing.published","listing.updated","inquiry.created"]
func (c *IntegrationClient) RegisterWebhook(ctx context.Context, webhookURL string, events []string, secret string) (*DLDWebhookRegistration, error) {
	body := map[string]interface{}{
		"url":    webhookURL,
		"events": events,
		"secret": secret,
	}
	var wrapper struct {
		Data DLDWebhookRegistration `json:"data"`
	}
	if err := c.post(ctx, "/v1/integrations/webhooks/register", body, &wrapper); err != nil {
		return nil, fmt.Errorf("dld RegisterWebhook: %w", err)
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
		return fmt.Errorf("dld rate limit exceeded (429) — retry after %s", resp.Header.Get("Retry-After"))
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("dld api error: status %d — %s", resp.StatusCode, string(bodyBytes))
	}

	return json.Unmarshal(bodyBytes, out)
}
