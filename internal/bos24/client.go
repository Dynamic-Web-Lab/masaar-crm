package bos24

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	BaseURL = "https://data.buyorsell24.com"
)

// BuyOrSell24 API Documentation & Setup:
// - API Docs: https://data.buyorsell24.com/redoc
// - Product: https://dynamicweblab.com/products/real-estate-data-api/
// - To get API token: Contact via product landing page
// - Masaar Pro users: Credits included with subscription
// - Open-source users: Contact for custom pricing
//
// Client wraps BuyOrSell24 API with caching
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	cache      *redis.Client
	cacheTTL   time.Duration
}

// NewClient creates a new BOS24 API client
// token should be your BuyOrSell24 API token
// rdb is optional - if provided, responses will be cached
func NewClient(token string, rdb *redis.Client) *Client {
	return &Client{
		baseURL: BaseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache:    rdb,
		cacheTTL: 1 * time.Hour,
	}
}

// SearchProperties performs natural language property search
func (c *Client) SearchProperties(ctx context.Context, query string, limit int) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"query": query,
		"limit": limit,
	}

	var result map[string]interface{}
	err := c.postWithCache(ctx, "/public/search/properties", payload, &result, 1*time.Hour)
	return result, err
}

// GetTransactions retrieves transactions with optional filters
func (c *Client) GetTransactions(ctx context.Context, filters map[string]interface{}) ([]map[string]interface{}, error) {
	params := url.Values{}

	// Add filters as query parameters
	if area, ok := filters["area"].(string); ok {
		params.Set("area", area)
	}
	if propertyType, ok := filters["property_type"].(string); ok {
		params.Set("property_type", propertyType)
	}
	if transType, ok := filters["trans_type"].(string); ok {
		params.Set("trans_type", transType)
	}
	if minPrice, ok := filters["min_price"].(float64); ok {
		params.Set("min_price", fmt.Sprintf("%v", minPrice))
	}
	if maxPrice, ok := filters["max_price"].(float64); ok {
		params.Set("max_price", fmt.Sprintf("%v", maxPrice))
	}
	if limit, ok := filters["limit"].(int); ok {
		params.Set("limit", fmt.Sprintf("%d", limit))
	} else {
		params.Set("limit", "20")
	}

	var result []map[string]interface{}
	err := c.getWithCache(ctx, "/public/realestate/transactions", params, &result, 1*time.Hour)
	return result, err
}

// GetBuilding retrieves building details
func (c *Client) GetBuilding(ctx context.Context, buildingID int) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.getWithCache(ctx,
		fmt.Sprintf("/public/realestate/buildings/%d", buildingID),
		url.Values{},
		&result,
		7*24*time.Hour, // Buildings are static, cache longer
	)
	return result, err
}

// SearchBuildings autocomplete search for buildings
func (c *Client) SearchBuildings(ctx context.Context, query string, limit int) ([]map[string]interface{}, error) {
	params := url.Values{
		"q":     {query},
		"limit": {fmt.Sprintf("%d", limit)},
	}

	var result []map[string]interface{}
	err := c.getWithCache(ctx, "/public/search/buildings/autocomplete", params, &result, 24*time.Hour)
	return result, err
}

// GetTransactionStats returns transaction statistics
func (c *Client) GetTransactionStats(ctx context.Context, filters map[string]interface{}) (map[string]interface{}, error) {
	params := url.Values{}

	if area, ok := filters["area"].(string); ok {
		params.Set("area", area)
	}
	if propertyType, ok := filters["property_type"].(string); ok {
		params.Set("property_type", propertyType)
	}

	var result map[string]interface{}
	err := c.getWithCache(ctx, "/public/realestate/transactions/stats", params, &result, 1*time.Hour)
	return result, err
}

// GetEjariStats returns rental statistics from Ejari contracts
func (c *Client) GetEjariStats(ctx context.Context, filters map[string]interface{}) (map[string]interface{}, error) {
	params := url.Values{}

	if area, ok := filters["area"].(string); ok {
		params.Set("area", area)
	}
	if propertyType, ok := filters["property_type"].(string); ok {
		params.Set("property_type", propertyType)
	}

	var result map[string]interface{}
	err := c.getWithCache(ctx, "/public/realestate/rentals/ejari/stats", params, &result, 1*time.Hour)
	return result, err
}

// GetNearbySchools returns schools near a location
func (c *Client) GetNearbySchools(ctx context.Context, lat, lng, radiusKm float64, limit int) ([]map[string]interface{}, error) {
	params := url.Values{
		"lat":       {fmt.Sprintf("%f", lat)},
		"lng":       {fmt.Sprintf("%f", lng)},
		"radius_km": {fmt.Sprintf("%f", radiusKm)},
		"limit":     {fmt.Sprintf("%d", limit)},
	}

	var result []map[string]interface{}
	err := c.getWithCache(ctx, "/public/realestate/schools/nearby", params, &result, 24*time.Hour)
	return result, err
}

// GetAreas returns all UAE real estate areas (used for dropdown filters)
func (c *Client) GetAreas(ctx context.Context) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	err := c.getWithCache(ctx, "/public/realestate/areas", url.Values{}, &result, 24*time.Hour)
	return result, err
}

// GetAreaBuildings returns buildings within a specific area
func (c *Client) GetAreaBuildings(ctx context.Context, areaSlug string, limit int) ([]map[string]interface{}, error) {
	params := url.Values{"limit": {fmt.Sprintf("%d", limit)}}
	var result []map[string]interface{}
	err := c.getWithCache(ctx, fmt.Sprintf("/public/realestate/areas/%s/buildings", areaSlug), params, &result, 6*time.Hour)
	return result, err
}

// GetAreaSummary returns market summary stats for a specific area (avg price, volume, trend)
func (c *Client) GetAreaSummary(ctx context.Context, areaSlug string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.getWithCache(ctx, fmt.Sprintf("/public/realestate/areas/%s/summary", areaSlug), url.Values{}, &result, 1*time.Hour)
	return result, err
}

// GetPOIs returns points of interest (metro, schools, malls, hospitals) near coordinates
func (c *Client) GetPOIs(ctx context.Context, lat, lng, radiusKm float64, category string, limit int) ([]map[string]interface{}, error) {
	params := url.Values{
		"lat":       {fmt.Sprintf("%f", lat)},
		"lng":       {fmt.Sprintf("%f", lng)},
		"radius_km": {fmt.Sprintf("%f", radiusKm)},
		"limit":     {fmt.Sprintf("%d", limit)},
	}
	if category != "" {
		params.Set("category", category)
	}
	var result []map[string]interface{}
	err := c.getWithCache(ctx, "/public/maps/pois", params, &result, 24*time.Hour)
	return result, err
}

// GetAreasWithLocations returns areas with lat/lng centroids for map rendering
func (c *Client) GetAreasWithLocations(ctx context.Context) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	err := c.getWithCache(ctx, "/public/maps/areas", url.Values{}, &result, 24*time.Hour)
	return result, err
}

// GetEjariRentals returns Ejari rental contract listings
func (c *Client) GetEjariRentals(ctx context.Context, filters map[string]interface{}) ([]map[string]interface{}, error) {
	params := url.Values{}
	if area, ok := filters["area"].(string); ok {
		params.Set("area", area)
	}
	if propertyType, ok := filters["property_type"].(string); ok {
		params.Set("property_type", propertyType)
	}
	if rooms, ok := filters["rooms"].(string); ok {
		params.Set("rooms", rooms)
	}
	if limit, ok := filters["limit"].(int); ok {
		params.Set("limit", fmt.Sprintf("%d", limit))
	} else {
		params.Set("limit", "20")
	}
	var result []map[string]interface{}
	err := c.getWithCache(ctx, "/public/realestate/rentals/ejari", params, &result, 1*time.Hour)
	return result, err
}

// GetEjariYield returns rental yield calculations per area/type
func (c *Client) GetEjariYield(ctx context.Context, filters map[string]interface{}) ([]map[string]interface{}, error) {
	params := url.Values{}
	if area, ok := filters["area"].(string); ok {
		params.Set("area", area)
	}
	if propertyType, ok := filters["property_type"].(string); ok {
		params.Set("property_type", propertyType)
	}
	var result []map[string]interface{}
	err := c.getWithCache(ctx, "/public/realestate/rentals/ejari/yield", params, &result, 2*time.Hour)
	return result, err
}

// GetRentals returns general rental listings
func (c *Client) GetRentals(ctx context.Context, filters map[string]interface{}) ([]map[string]interface{}, error) {
	params := url.Values{}
	if area, ok := filters["area"].(string); ok {
		params.Set("area", area)
	}
	if propertyType, ok := filters["property_type"].(string); ok {
		params.Set("property_type", propertyType)
	}
	if limit, ok := filters["limit"].(int); ok {
		params.Set("limit", fmt.Sprintf("%d", limit))
	} else {
		params.Set("limit", "20")
	}
	var result []map[string]interface{}
	err := c.getWithCache(ctx, "/public/realestate/rentals", params, &result, 1*time.Hour)
	return result, err
}

// GetDevelopers returns real estate developer listings
func (c *Client) GetDevelopers(ctx context.Context, query string, limit int) ([]map[string]interface{}, error) {
	params := url.Values{"limit": {fmt.Sprintf("%d", limit)}}
	if query != "" {
		params.Set("q", query)
	}
	var result []map[string]interface{}
	err := c.getWithCache(ctx, "/public/realestate/developers", params, &result, 12*time.Hour)
	return result, err
}

// GetProjects returns real estate project listings
func (c *Client) GetProjects(ctx context.Context, filters map[string]interface{}) ([]map[string]interface{}, error) {
	params := url.Values{}
	if area, ok := filters["area"].(string); ok {
		params.Set("area", area)
	}
	if developer, ok := filters["developer"].(string); ok {
		params.Set("developer", developer)
	}
	if status, ok := filters["status"].(string); ok {
		params.Set("status", status)
	}
	if limit, ok := filters["limit"].(int); ok {
		params.Set("limit", fmt.Sprintf("%d", limit))
	} else {
		params.Set("limit", "20")
	}
	var result []map[string]interface{}
	err := c.getWithCache(ctx, "/public/realestate/projects", params, &result, 6*time.Hour)
	return result, err
}

// SearchProjects uses AI to search for projects by natural language
func (c *Client) SearchProjects(ctx context.Context, query string, limit int) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"query": query,
		"limit": limit,
	}
	var result map[string]interface{}
	err := c.postWithCache(ctx, "/public/search/projects", payload, &result, 1*time.Hour)
	return result, err
}

// GetTransactionAreas returns transaction volume and statistics grouped by area
func (c *Client) GetTransactionAreas(ctx context.Context, filters map[string]interface{}) ([]map[string]interface{}, error) {
	params := url.Values{}
	if propertyType, ok := filters["property_type"].(string); ok {
		params.Set("property_type", propertyType)
	}
	if transType, ok := filters["trans_type"].(string); ok {
		params.Set("trans_type", transType)
	}
	if limit, ok := filters["limit"].(int); ok {
		params.Set("limit", fmt.Sprintf("%d", limit))
	} else {
		params.Set("limit", "30")
	}
	var result []map[string]interface{}
	err := c.getWithCache(ctx, "/public/realestate/transactions/areas", params, &result, 2*time.Hour)
	return result, err
}

// GetValuations returns automated property valuations
func (c *Client) GetValuations(ctx context.Context, filters map[string]interface{}) ([]map[string]interface{}, error) {
	params := url.Values{}
	if area, ok := filters["area"].(string); ok {
		params.Set("area", area)
	}
	if propertyType, ok := filters["property_type"].(string); ok {
		params.Set("property_type", propertyType)
	}
	if rooms, ok := filters["rooms"].(string); ok {
		params.Set("rooms", rooms)
	}
	var result []map[string]interface{}
	err := c.getWithCache(ctx, "/public/realestate/valuations", params, &result, 2*time.Hour)
	return result, err
}

// DescribeProperty generates an AI description for a property listing
func (c *Client) DescribeProperty(ctx context.Context, details map[string]interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.postWithCache(ctx, "/public/ai/describe-property", details, &result, 30*time.Minute)
	return result, err
}

// GetUnits returns individual unit listings in a building or project
func (c *Client) GetUnits(ctx context.Context, filters map[string]interface{}) ([]map[string]interface{}, error) {
	params := url.Values{}
	if buildingID, ok := filters["building_id"].(int); ok {
		params.Set("building_id", fmt.Sprintf("%d", buildingID))
	}
	if projectID, ok := filters["project_id"].(string); ok {
		params.Set("project_id", projectID)
	}
	if propertyType, ok := filters["property_type"].(string); ok {
		params.Set("property_type", propertyType)
	}
	if limit, ok := filters["limit"].(int); ok {
		params.Set("limit", fmt.Sprintf("%d", limit))
	} else {
		params.Set("limit", "20")
	}
	var result []map[string]interface{}
	err := c.getWithCache(ctx, "/public/realestate/units", params, &result, 1*time.Hour)
	return result, err
}

// Helper methods

func (c *Client) getWithCache(ctx context.Context, path string, params url.Values, result interface{}, ttl time.Duration) error {
	cacheKey := fmt.Sprintf("bos24:get:%s:%s", path, params.Encode())

	// Try cache first
	if c.cache != nil {
		if cached, err := c.cache.Get(ctx, cacheKey).Result(); err == nil {
			return json.Unmarshal([]byte(cached), result)
		}
	}

	// Make API call
	fullURL := fmt.Sprintf("%s%s", c.baseURL, path)
	if params.Encode() != "" {
		fullURL = fmt.Sprintf("%s?%s", fullURL, params.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bos24 api error: %d - %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, result); err != nil {
		return err
	}

	// Cache the result
	if c.cache != nil {
		c.cache.Set(ctx, cacheKey, string(body), ttl)
	}

	return nil
}

func (c *Client) postWithCache(ctx context.Context, path string, payload interface{}, result interface{}, ttl time.Duration) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("bos24:post:%s:%s", path, string(payloadBytes))

	// Try cache first
	if c.cache != nil {
		if cached, err := c.cache.Get(ctx, cacheKey).Result(); err == nil {
			return json.Unmarshal([]byte(cached), result)
		}
	}

	// Make API call
	fullURL := fmt.Sprintf("%s%s", c.baseURL, path)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bos24 api error: %d - %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, result); err != nil {
		return err
	}

	// Cache the result
	if c.cache != nil {
		c.cache.Set(ctx, cacheKey, string(body), ttl)
	}

	return nil
}

// IsEnabled checks if BOS24 integration is enabled
// Returns true only if token is set and non-empty
func IsEnabled(token string) bool {
	return token != "" && len(token) > 0
}
