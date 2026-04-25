package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/maidulcu/masaar-crm/internal/bos24"
)

type PropertyHandler struct {
	bos24Client *bos24.Client
}

func NewPropertyHandler(bos24Client *bos24.Client) *PropertyHandler {
	return &PropertyHandler{
		bos24Client: bos24Client,
	}
}

// SearchProperties performs natural language property search
// @Summary Search properties by natural language query
// @Description Search for properties using natural language (e.g., "2BR apartments in Marina").
// Requires BOS24_API_TOKEN configured. See https://data.buyorsell24.com/redoc for API details.
// @Tags Property Search
// @Accept json
// @Produce json
// @Param request body SearchPropertiesRequest true "Search query"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /api/v1/properties/search [post]
// @Security Bearer
func (h *PropertyHandler) SearchProperties(c *fiber.Ctx) error {
	if h.bos24Client == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "real estate service not enabled",
		})
	}

	var req SearchPropertiesRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	if req.Query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "query is required"})
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	result, err := h.bos24Client.SearchProperties(c.Context(), req.Query, limit)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "failed to search properties: " + err.Error(),
		})
	}

	return c.JSON(result)
}

// GetTransactions retrieves recent property transactions with optional filters
// @Summary Get property transactions
// @Description Get transaction history with optional filters (area, property_type, price range, etc.)
// @Tags Transactions
// @Produce json
// @Param area query string false "Area/district name"
// @Param property_type query string false "Property type (Unit, Villa, etc.)"
// @Param trans_type query string false "Transaction type (Sell, Rent)"
// @Param min_price query number false "Minimum price"
// @Param max_price query number false "Maximum price"
// @Param limit query integer false "Result limit (default 20)"
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} map[string]string
// @Router /api/v1/properties/transactions [get]
// @Security Bearer
func (h *PropertyHandler) GetTransactions(c *fiber.Ctx) error {
	if h.bos24Client == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "real estate service not enabled",
		})
	}

	filters := make(map[string]interface{})

	if area := c.Query("area"); area != "" {
		filters["area"] = area
	}
	if propertyType := c.Query("property_type"); propertyType != "" {
		filters["property_type"] = propertyType
	}
	if transType := c.Query("trans_type"); transType != "" {
		filters["trans_type"] = transType
	}
	if minPrice := c.Query("min_price"); minPrice != "" {
		if price, err := strconv.ParseFloat(minPrice, 64); err == nil {
			filters["min_price"] = price
		}
	}
	if maxPrice := c.Query("max_price"); maxPrice != "" {
		if price, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			filters["max_price"] = price
		}
	}
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filters["limit"] = l
		}
	}

	result, err := h.bos24Client.GetTransactions(c.Context(), filters)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "failed to get transactions: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"transactions": result,
	})
}

// GetBuildings searches buildings with autocomplete
// @Summary Search buildings
// @Description Autocomplete search for buildings and projects
// @Tags Buildings
// @Produce json
// @Param q query string true "Building name search term"
// @Param limit query integer false "Result limit (default 10)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /api/v1/properties/buildings [get]
// @Security Bearer
func (h *PropertyHandler) GetBuildings(c *fiber.Ctx) error {
	if h.bos24Client == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "real estate service not enabled",
		})
	}

	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "search query required"})
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	result, err := h.bos24Client.SearchBuildings(c.Context(), query, limit)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "failed to search buildings: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"buildings": result,
	})
}

// GetBuildingByID retrieves detailed building information
// @Summary Get building details
// @Description Get complete details for a specific building by ID
// @Tags Buildings
// @Produce json
// @Param id path string true "Building ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /api/v1/properties/buildings/:id [get]
// @Security Bearer
func (h *PropertyHandler) GetBuildingByID(c *fiber.Ctx) error {
	if h.bos24Client == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "real estate service not enabled",
		})
	}

	idStr := c.Params("id")
	buildingID, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid building ID"})
	}

	result, err := h.bos24Client.GetBuilding(c.Context(), buildingID)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "failed to get building: " + err.Error(),
		})
	}

	return c.JSON(result)
}

// GetNearbySchools finds schools near a location
// @Summary Find schools and amenities nearby
// @Description Get list of schools and amenities near coordinates
// @Tags Location
// @Produce json
// @Param lat query number true "Latitude"
// @Param lng query number true "Longitude"
// @Param radius_km query number false "Search radius in km (default 2)"
// @Param limit query integer false "Result limit (default 10)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /api/v1/properties/schools/nearby [get]
// @Security Bearer
func (h *PropertyHandler) GetNearbySchools(c *fiber.Ctx) error {
	if h.bos24Client == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "real estate service not enabled",
		})
	}

	latStr := c.Query("lat")
	lngStr := c.Query("lng")

	if latStr == "" || lngStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "latitude and longitude required",
		})
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid latitude"})
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid longitude"})
	}

	radius := 2.0
	if r := c.Query("radius_km"); r != "" {
		if parsed, err := strconv.ParseFloat(r, 64); err == nil && parsed > 0 {
			radius = parsed
		}
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	result, err := h.bos24Client.GetNearbySchools(c.Context(), lat, lng, radius, limit)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "failed to get schools: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"schools": result,
	})
}

// Request/Response types
type SearchPropertiesRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}
