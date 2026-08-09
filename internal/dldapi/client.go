package dldapi

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
	BaseURL = "https://dldapi.waqov.com"
)

// DLDAPI Documentation & Setup:
// - API Docs: https://dldapi.waqov.com/docs
// - To get API token: Contact via product landing page
// - Masaar Pro users: Credits included with subscription
// - Open-source users: Contact for custom pricing
//
// Client wraps DLDAPI with caching
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	cache      *redis.Client
	cacheTTL   time.Duration
}

// NewClient creates a new DLDAPI client.
// baseURL overrides the default production URL (pass "" to use the default).
// rdb is optional - if provided, responses will be cached.
func NewClient(token, baseURL string, rdb *redis.Client) *Client {
	if baseURL == "" {
		baseURL = BaseURL
	}
	return &Client{
		baseURL: baseURL,
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

// ── Insights (v1.3.0 — market intelligence) ──────────────────────────────────

func (c *Client) GetMarketOverview(ctx context.Context, period, propertyType string) (map[string]interface{}, error) {
	params := url.Values{}
	if period != "" {
		params.Set("period", period)
	}
	if propertyType != "" {
		params.Set("property_type", propertyType)
	}
	var result map[string]interface{}
	err := c.getWithCache(ctx, "/public/realestate/insights/market-overview", params, &result, 30*time.Minute)
	return result, err
}

func (c *Client) GetAreaComparison(ctx context.Context, areas, propertyType string) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("areas", areas)
	if propertyType != "" {
		params.Set("property_type", propertyType)
	}
	var result map[string]interface{}
	err := c.getWithCache(ctx, "/public/realestate/insights/area-comparison", params, &result, 30*time.Minute)
	return result, err
}

func (c *Client) GetPriceTrends(ctx context.Context, area, propertyType, granularity string) (map[string]interface{}, error) {
	params := url.Values{}
	if area != "" {
		params.Set("area", area)
	}
	if propertyType != "" {
		params.Set("property_type", propertyType)
	}
	if granularity != "" {
		params.Set("granularity", granularity)
	}
	var result map[string]interface{}
	err := c.getWithCache(ctx, "/public/realestate/insights/price-trends", params, &result, 1*time.Hour)
	return result, err
}

func (c *Client) GetTopAreas(ctx context.Context, metric, propertyType string, limit int) (map[string]interface{}, error) {
	params := url.Values{}
	if metric != "" {
		params.Set("metric", metric)
	}
	if propertyType != "" {
		params.Set("property_type", propertyType)
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	var result map[string]interface{}
	err := c.getWithCache(ctx, "/public/realestate/insights/top-areas", params, &result, 1*time.Hour)
	return result, err
}

// ── Brokers ───────────────────────────────────────────────────────────────────

func (c *Client) GetBrokers(ctx context.Context, limit, offset int) ([]map[string]interface{}, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("offset", fmt.Sprintf("%d", offset))
	var result []map[string]interface{}
	err := c.getWithCache(ctx, "/public/realestate/brokers", params, &result, 12*time.Hour)
	return result, err
}

// ── Transaction details ───────────────────────────────────────────────────────

func (c *Client) GetTransaction(ctx context.Context, transactionID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.getWithCache(ctx, fmt.Sprintf("/public/realestate/transactions/%s", transactionID), url.Values{}, &result, 24*time.Hour)
	return result, err
}

func (c *Client) GetEnrichedTransaction(ctx context.Context, transactionID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.getWithCache(ctx, fmt.Sprintf("/public/realestate/transactions/%s/enriched", transactionID), url.Values{}, &result, 24*time.Hour)
	return result, err
}

func (c *Client) GetTransactionsByProject(ctx context.Context, projectName string, filters map[string]interface{}) (map[string]interface{}, error) {
	params := url.Values{}
	if v, ok := filters["rooms"].(string); ok && v != "" {
		params.Set("rooms", v)
	}
	if v, ok := filters["limit"].(int); ok {
		params.Set("limit", fmt.Sprintf("%d", v))
	}
	if v, ok := filters["offset"].(int); ok {
		params.Set("offset", fmt.Sprintf("%d", v))
	}
	var result map[string]interface{}
	err := c.getWithCache(ctx, fmt.Sprintf("/public/realestate/transactions/by-project/%s", url.PathEscape(projectName)), params, &result, 1*time.Hour)
	return result, err
}

func (c *Client) GetAreaTransactionSummary(ctx context.Context, areaName string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.getWithCache(ctx, fmt.Sprintf("/public/realestate/transactions/area/%s/summary", url.PathEscape(areaName)), url.Values{}, &result, 1*time.Hour)
	return result, err
}

// ── Rental sub-endpoints ──────────────────────────────────────────────────────

func (c *Client) GetRentalStats(ctx context.Context, filters map[string]interface{}) (map[string]interface{}, error) {
	params := url.Values{}
	if v, ok := filters["area"].(string); ok && v != "" {
		params.Set("area", v)
	}
	if v, ok := filters["property_type"].(string); ok && v != "" {
		params.Set("property_type", v)
	}
	if v, ok := filters["rooms"].(string); ok && v != "" {
		params.Set("rooms", v)
	}
	var result map[string]interface{}
	err := c.getWithCache(ctx, "/public/realestate/rentals/stats", params, &result, 1*time.Hour)
	return result, err
}

func (c *Client) GetRentalAreas(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.getWithCache(ctx, "/public/realestate/rentals/areas", url.Values{}, &result, 2*time.Hour)
	return result, err
}

func (c *Client) GetRentalsByProject(ctx context.Context, projectName string, limit, offset int) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("offset", fmt.Sprintf("%d", offset))
	var result map[string]interface{}
	err := c.getWithCache(ctx, fmt.Sprintf("/public/realestate/rentals/project/%s", url.PathEscape(projectName)), params, &result, 1*time.Hour)
	return result, err
}

func (c *Client) GetRentalsByBuilding(ctx context.Context, buildingName string, limit, offset int) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("offset", fmt.Sprintf("%d", offset))
	var result map[string]interface{}
	err := c.getWithCache(ctx, fmt.Sprintf("/public/realestate/rentals/building/%s", url.PathEscape(buildingName)), params, &result, 1*time.Hour)
	return result, err
}

// ── Lands ─────────────────────────────────────────────────────────────────────

func (c *Client) GetLands(ctx context.Context, limit, offset int) ([]map[string]interface{}, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("offset", fmt.Sprintf("%d", offset))
	var result []map[string]interface{}
	err := c.getWithCache(ctx, "/public/realestate/lands", params, &result, 4*time.Hour)
	return result, err
}

func (c *Client) GetLand(ctx context.Context, landID int) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.getWithCache(ctx, fmt.Sprintf("/public/realestate/lands/%d", landID), url.Values{}, &result, 4*time.Hour)
	return result, err
}

// ── Map extras ────────────────────────────────────────────────────────────────

func (c *Client) GetMapConfig(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.getWithCache(ctx, "/public/maps/config", url.Values{}, &result, 24*time.Hour)
	return result, err
}

func (c *Client) GetMapBounds(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.getWithCache(ctx, "/public/maps/bounds", url.Values{}, &result, 24*time.Hour)
	return result, err
}

func (c *Client) GetPOICategories(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.getWithCache(ctx, "/public/maps/pois/categories", url.Values{}, &result, 24*time.Hour)
	return result, err
}

func (c *Client) GetPropertyHeatmap(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.getWithCache(ctx, "/public/maps/heatmap/properties", url.Values{}, &result, 2*time.Hour)
	return result, err
}

func (c *Client) GetAreaLocation(ctx context.Context, areaName string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.getWithCache(ctx, fmt.Sprintf("/public/maps/areas/%s", url.PathEscape(areaName)), url.Values{}, &result, 24*time.Hour)
	return result, err
}

// ── Single record lookups ─────────────────────────────────────────────────────

func (c *Client) GetAreaByID(ctx context.Context, areaID int) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.getWithCache(ctx, fmt.Sprintf("/public/realestate/areas/%d", areaID), url.Values{}, &result, 24*time.Hour)
	return result, err
}

func (c *Client) GetDeveloper(ctx context.Context, developerID int) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.getWithCache(ctx, fmt.Sprintf("/public/realestate/developers/%d", developerID), url.Values{}, &result, 12*time.Hour)
	return result, err
}

func (c *Client) GetProject(ctx context.Context, projectID int) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.getWithCache(ctx, fmt.Sprintf("/public/realestate/projects/%d", projectID), url.Values{}, &result, 6*time.Hour)
	return result, err
}

func (c *Client) GetUnit(ctx context.Context, unitID int) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.getWithCache(ctx, fmt.Sprintf("/public/realestate/units/%d", unitID), url.Values{}, &result, 2*time.Hour)
	return result, err
}

func (c *Client) GetValuation(ctx context.Context, valuationID int) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.getWithCache(ctx, fmt.Sprintf("/public/realestate/valuations/%d", valuationID), url.Values{}, &result, 2*time.Hour)
	return result, err
}

// Helper methods

func (c *Client) getWithCache(ctx context.Context, path string, params url.Values, result interface{}, ttl time.Duration) error {
	cacheKey := fmt.Sprintf("dldapi:get:%s:%s", path, params.Encode())

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
		return fmt.Errorf("dldapi error: %d - %s", resp.StatusCode, string(body))
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

	cacheKey := fmt.Sprintf("dldapi:post:%s:%s", path, string(payloadBytes))

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
		return fmt.Errorf("dldapi error: %d - %s", resp.StatusCode, string(body))
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

// IsEnabled checks if DLDAPI integration is enabled
// Returns true only if token is set and non-empty
func IsEnabled(token string) bool {
	return token != "" && len(token) > 0
}
