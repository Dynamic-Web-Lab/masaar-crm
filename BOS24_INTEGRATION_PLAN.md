# BuyOrSell24 API Integration Plan for Masaar CRM

## Overview

**BuyOrSell24 Engine** is a Dubai real estate data API with:
- 📊 Transaction history (sales, rentals, Ejari)
- 🏢 Building & project database
- 🎯 AI-powered natural language search
- 📍 Maps & location data
- 📈 Market analytics & statistics

**Perfect for**: Real estate agents, brokers, and WhatsApp CRM workflows

---

## Integration Opportunities

### 1. **Enrich Contact/Lead Data** ⭐ HIGH PRIORITY
**Use Case**: When an agent adds a lead, automatically fetch property data

```
Flow:
1. Agent mentions property in WhatsApp: "Customer interested in 2BR Marina"
2. Masaar parses intent and queries BuyOrSell24 API
3. Returns recent transactions, comparable properties, price trends
4. Auto-enriches lead with market data
```

**Implementation**:
- New endpoint: `POST /api/v1/properties/search`
- Calls BuyOrSell24 NL search: `/public/search/properties`
- Stores results in lead notes/context

### 2. **Transaction Lookup** ⭐ HIGH PRIORITY
**Use Case**: Agent asks about specific property history

```
Example:
Input: "What sold in downtown last month?"
Flow:
1. Parse query (area=Downtown, trans_type=Sell, start_date=last_month)
2. Query: GET /public/realestate/transactions
3. Return comparable sales with prices/dates
4. Display in WhatsApp chat
```

**Implementation**:
- Handler: `internal/api/handler/property.go`
- Endpoints:
  - `GET /api/v1/properties/transactions` (filter by area/type/date)
  - `GET /api/v1/properties/comparables/{property_id}` (similar properties)

### 3. **Building/Project Directory** ⭐ MEDIUM PRIORITY
**Use Case**: Quick building info during conversations

```
Example:
Input: "Info on DAMAC Maison?", "Schools near Marina?"
Flow:
1. Autocomplete search: GET /public/search/buildings/autocomplete?q=DAMAC
2. Get building details: GET /public/realestate/buildings/{id}
3. Get nearby schools: GET /public/realestate/schools/nearby
4. Return formatted info for WhatsApp
```

**Implementation**:
- Endpoints:
  - `GET /api/v1/properties/buildings` (search + list)
  - `GET /api/v1/properties/buildings/{id}` (details)
  - `GET /api/v1/properties/schools/nearby` (proximity search)

### 4. **Rental Yield Analysis** ⭐ MEDIUM PRIORITY
**Use Case**: Investors analyzing ROI on properties

```
Example:
Input: "Yield on 2BR apartments in Downtown?"
Flow:
1. Query Ejari rentals: GET /public/realestate/rentals/ejari/stats
2. Get comparable sales: GET /public/realestate/transactions
3. Calculate: annual_rent / avg_sale_price = yield %
4. Show comparison: "Downtown average yield: 4.2% vs Marina: 3.8%"
```

**Implementation**:
- Endpoint: `GET /api/v1/properties/yield-analysis`
- Combines:
  - `/public/realestate/rentals/ejari/stats`
  - `/public/realestate/transactions/stats`

### 5. **Price Trends & Heatmaps** 🟢 LOW PRIORITY
**Use Case**: Market insights dashboard

```
Implementation:
- Endpoint: GET /api/v1/properties/market-trends
- Returns: price history, volume, per-sqm trends by area
- Display as chart/heatmap in dashboard
```

---

## Architecture

### New Handler: `property.go`

```go
type PropertyHandler struct {
    bos24Client *bos24.Client  // BuyOrSell24 API client
    repo        *repo.PropertyRepo  // Cache layer
}

// Handlers
- SearchProperties()      // NL search
- GetTransactions()       // Search by filters
- GetComparables()        // Similar properties
- GetBuilding()           // Building details
- SearchBuildings()       // Autocomplete
- GetYieldAnalysis()      // Rental yield
- GetMarketTrends()       // Statistics
```

### New Service: `internal/bos24/client.go`

```go
type Client struct {
    baseURL    string
    token      string
    httpClient *http.Client
    cache      redis.Client  // Cache API responses
}

// Methods
- SearchProperties(query, limit)
- GetTransactions(filters)
- GetBuilding(id)
- SearchBuildings(query)
- GetEjariStats(filters)
- GetTransactionStats(filters)
```

### Cache Strategy (Redis)

```
Key Format: bos24:{endpoint}:{params_hash}
TTL Examples:
- Building details: 7 days (static)
- Transaction stats: 1 hour (updates daily)
- Recent transactions: 1 hour
- Ejari stats: 1 hour
- Autocomplete: 24 hours
```

---

## API Endpoints to Add

### Property Search

**POST /api/v1/properties/search**
```
Request:
{
  "query": "2 bedroom apartment in Marina",
  "limit": 20
}

Response:
{
  "results": [
    {
      "id": 123,
      "building": "Marina Residences",
      "area": "Marina",
      "rooms": "2BR",
      "price": 1200000,
      "price_per_sqm": 12000
    }
  ],
  "total": 45,
  "ai_explanation": "Found 2BR apartments in Marina area with recent transactions"
}
```

### Transaction Lookup

**GET /api/v1/properties/transactions**
```
Params:
- area: "Downtown"
- property_type: "Unit"
- trans_type: "Sell"
- start_date: "2026-04-01"
- end_date: "2026-04-25"
- min_price: 500000
- max_price: 2000000
- limit: 20

Response:
{
  "transactions": [
    {
      "id": "1-102-2026-2432",
      "building": "Burj Khalifa",
      "area": "Downtown",
      "rooms": "2BR",
      "price": 1500000,
      "date": "2026-04-20",
      "price_per_sqm": 15000
    }
  ],
  "total": 142,
  "avg_price": 1450000,
  "stats": {...}
}
```

### Comparables

**GET /api/v1/properties/comparables**
```
Params:
- building_id: 456
- rooms: "2BR"
- radius_km: 2
- limit: 10

Response:
{
  "target_property": {...},
  "comparables": [
    {similar_properties}
  ],
  "analysis": {
    "avg_price": 1400000,
    "price_range": [1200000, 1600000],
    "price_per_sqm": [12000, 14000]
  }
}
```

### Yield Analysis

**GET /api/v1/properties/yield-analysis**
```
Params:
- area: "Downtown"
- property_type: "Unit"
- rooms: "2BR"

Response:
{
  "area": "Downtown",
  "rental_market": {
    "avg_annual_rent": 120000,
    "rental_volume": 234
  },
  "sales_market": {
    "avg_price": 1450000,
    "sales_volume": 156
  },
  "yield": {
    "gross_yield_percent": 8.27,
    "comparison": "Higher than Dubai average (6.5%)"
  }
}
```

---

## Cost Implications

### BuyOrSell24 Credit System

**Per Query Cost:**
| Query Type | Credits | Example |
|-----------|---------|---------|
| List/search | 1 | GET /areas, /transactions |
| Building details | 1 | GET /buildings/{id} |
| Rental stats | 5 | GET /rentals/ejari/stats |
| Transaction stats | 25 | GET /transactions/stats (expensive) |
| School search | 1 | GET /schools/nearby |

**Example Usage for Masaar:**
- 100 agents × 5 queries/day = 500 queries/day
- If mostly search (1 credit) = ~500 credits/day = 15,000/month
- **Cost**: Lite Plan (165 credits/hour = 1,000/month) insufficient
  - **Recommendation**: Startup Plan (5,000/month) for 100+ agents

### Implementation Cost Estimate

| Task | Hours | Priority |
|------|-------|----------|
| BOS24 client library | 8 | HIGH |
| Property handler endpoints | 12 | HIGH |
| Caching layer (Redis) | 4 | HIGH |
| Unit tests | 6 | MEDIUM |
| Integration tests | 4 | MEDIUM |
| Documentation | 3 | MEDIUM |
| **Total** | **37 hours** | |

---

## Implementation Roadmap

### Phase 1: Foundation (Week 1)
- [ ] Create `bos24/client.go` with HTTP client
- [ ] Add auth token management
- [ ] Implement caching layer
- [ ] Create `property.go` handler
- [ ] Add 2 core endpoints (search, transactions)

### Phase 2: Core Features (Week 2)
- [ ] Building/project directory endpoints
- [ ] Comparables analysis
- [ ] Yield analysis
- [ ] Unit tests for all endpoints
- [ ] Error handling & rate limit management

### Phase 3: Polish (Week 3)
- [ ] Integration tests with real BOS24 API
- [ ] Caching optimization
- [ ] Documentation & API docs
- [ ] WhatsApp intent parsing for common queries
- [ ] Dashboard/UI for property data

### Phase 4: Enhancement (Week 4)
- [ ] Heatmaps & market trends
- [ ] Export reports (PDF)
- [ ] Saved searches/favorites
- [ ] Alerts on new transactions

---

## Example Workflows

### Workflow 1: Lead Enrichment
```
Agent WhatsApp: "New lead interested in 2BR Marina apartment"
↓
Masaar parses: area="Marina", rooms="2BR", type="Unit"
↓
Queries BOS24: GET /public/realestate/transactions?area=Marina&rooms=2BR
↓
Returns: Recent sales, avg price, price trends
↓
Masaar adds to lead context: "Market insight: 2BR Marina avg 1.4M AED"
```

### Workflow 2: Price Negotiation
```
Buyer: "Is 1.2M fair for this 2BR apartment?"
↓
Agent: "Check comparables"
↓
Masaar queries: GET /api/v1/properties/comparables?building_id=456
↓
Shows: "Last 10 sales in this building: 1.25M - 1.45M"
↓
Agent: "Market suggests 1.25M-1.35M would be fair"
```

### Workflow 3: Rental Analysis
```
Investor: "What's the yield on Downtown apartments?"
↓
Agent: "Let me check"
↓
Masaar: GET /api/v1/properties/yield-analysis?area=Downtown
↓
Response: "Gross yield: 8.3%, Higher than Dubai average 6.5%"
↓
Agent shares detailed analysis with investor
```

---

## Security Considerations

### Rate Limiting
- Implement per-user rate limiting in Masaar
- Cache responses to reduce API calls
- Batch queries where possible
- Monitor credit usage

### Authentication
- Store BOS24 token in environment variable (`.env`)
- Rotate token periodically
- Never expose token in logs/errors

### Data Privacy
- Cache only aggregated data in Redis
- Don't store individual transaction details
- Respect user data privacy laws (PDPL)

---

## Success Metrics

| Metric | Target |
|--------|--------|
| API response time | <500ms (incl. caching) |
| Cache hit rate | >70% |
| Agent adoption | >50% use property features |
| Time saved per query | 5 minutes → 30 seconds |
| Monthly credit usage | <4,000 (within Startup plan) |

---

## Risk Mitigation

| Risk | Mitigation |
|------|-----------|
| API downtime | Fallback to cached data, graceful degradation |
| Rate limit exceeded | Queue non-urgent requests, alert admins |
| Stale data | Clear cache on schedule, show data age |
| High credit usage | Aggressive caching, limit free queries |

---

## Decision: Should We Integrate?

### ✅ YES, if:
- Target market includes real estate agents
- Need market data to compete
- Budget for API subscription ($166-500/month)
- Time to implement (37+ hours)

### ❌ NOT NOW, if:
- Current users don't need property data
- Budget constraints
- Focus on other features first

### 🟡 HYBRID Approach:
- Start with read-only endpoints
- Implement core search only
- Monitor usage → decide on expansion
- Upgrade plan if successful

---

## Next Steps

1. **Get API token** from BuyOrSell24
2. **Test API** with sample queries
3. **Estimate actual credit usage** with your user base
4. **Decide implementation scope** (Phase 1 only? Full?)
5. **Allocate development time** (1-2 sprints)
6. **Create integration tests** with real API

Would you like me to:
- Build the BOS24 client library?
- Create sample handlers with tests?
- Design the WhatsApp intent parser?
- Estimate actual costs for your user base?
