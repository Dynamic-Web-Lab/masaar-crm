# Unit Testing Guide - Masaar CRM

## Overview

This document describes the test suite for Masaar CRM. Tests are written using Go's standard `testing` package and use mocks for external dependencies.

## Test Files Created

### 1. **internal/api/handler/auth_test.go**
Tests for authentication handler (Login endpoint)

**Test Cases:**
- ✅ `TestAuthLogin_Success` - Valid login with correct credentials
- ✅ `TestAuthLogin_InvalidCredentials` - Login with non-existent email
- ✅ `TestAuthLogin_WrongPassword` - Login with incorrect password
- ✅ `TestAuthLogin_InvalidRequest` - Malformed JSON request

**Key Coverage:**
- Successful token generation
- Error handling for missing users
- Error handling for incorrect passwords
- Request validation

### 2. **internal/api/middleware/auth_test.go**
Tests for authentication middleware (CheckBlacklist)

**Test Cases:**
- ✅ `TestCheckBlacklist_TokenNotBlacklisted` - Valid token access
- ✅ `TestCheckBlacklist_TokenIsBlacklisted` - Revoked token rejection
- ✅ `TestCheckBlacklist_RedisError_FailsClosed` - **CRITICAL** - Fail-closed on Redis errors
- ✅ `TestCheckBlacklist_NoToken` - Request without authorization
- ✅ `TestCheckBlacklist_MalformedAuthHeader` - Invalid header format

**Key Coverage:**
- Token revocation validation
- **Fail-closed behavior** on Redis failures (security critical)
- Bearer token extraction
- Case-insensitive header parsing

### 3. **internal/api/handler/lead_test.go**
Tests for lead stage transitions (Kanban board)

**Test Cases:**
- ✅ `TestLeadUpdateStage_ValidStage` - Update to valid stage
- ✅ `TestLeadUpdateStage_InvalidStage` - Reject invalid stage value
- ✅ `TestLeadUpdateStage_AllValidStages` - Test all 6 valid stages
- ✅ `TestLeadUpdateStage_MissingStage` - Request without stage field
- ✅ `TestLeadUpdateStage_InvalidID` - Invalid UUID format

**Valid Stages:** new, contacted, qualified, proposal, won, lost

**Key Coverage:**
- Enum validation prevents invalid states
- All valid transitions allowed
- Request validation

### 4. **internal/api/handler/invoice_test.go**
Tests for invoice status transitions

**Test Cases:**
- ✅ `TestInvoiceUpdateStatus_ValidStatus` - Update to valid status
- ✅ `TestInvoiceUpdateStatus_InvalidStatus` - Reject invalid status
- ✅ `TestInvoiceUpdateStatus_AllValidStatuses` - Test all 3 valid statuses
- ✅ `TestInvoiceUpdateStatus_MissingStatus` - Request without status field
- ✅ `TestInvoiceUpdateStatus_InvalidID` - Invalid UUID format

**Valid Statuses:** draft, sent, paid

**Key Coverage:**
- Enum validation prevents invalid states
- Status transitions validated
- Request validation

## Running Tests

### Run All Tests
```bash
go test ./...
```

### Run Tests in Specific Package
```bash
go test ./internal/api/handler/
go test ./internal/api/middleware/
```

### Run Specific Test
```bash
go test -run TestAuthLogin_Success ./internal/api/handler/
```

### Run Tests with Verbose Output
```bash
go test -v ./...
```

### Run Tests with Coverage
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Tests in Parallel
```bash
go test -parallel 4 ./...
```

## Mock Objects

The tests use mock repositories to avoid external dependencies:

### MockUserRepo
- Simulates user database operations
- Methods: `FindByEmail`, `FindByID`, `UpdatePassword`, `UpdateLangPref`

### MockLeadRepo
- Simulates lead database operations
- Methods: `GetByID`, `UpdateStage`, `Create`, `KanbanBoard`

### MockContactRepo
- Simulates contact database operations
- Methods: `GetByID`, `Create`

### MockInvoiceRepo
- Simulates invoice database operations
- Methods: `GetByID`, `UpdateStatus`, `Create`, `NextInvoiceNo`

### MockDealRepo
- Simulates deal database operations
- Methods: `GetByID`

### MockRedis (redismock)
- Uses `github.com/go-redis/redismock/v9` for Redis mocking
- Simulates Redis operations without network calls

## Test Coverage Goals

| Component | Coverage | Priority |
|-----------|----------|----------|
| Auth Handler | 100% | CRITICAL |
| Auth Middleware | 100% | CRITICAL |
| Lead Handler | 95%+ | HIGH |
| Deal Handler | 95%+ | HIGH |
| Invoice Handler | 95%+ | HIGH |
| User Handler | 90%+ | MEDIUM |
| Contact Handler | 90%+ | MEDIUM |

## Writing New Tests

### Template for Handler Tests
```go
func TestHandlerFunction_Scenario(t *testing.T) {
    // Setup
    app := fiber.New()
    mockRepo := NewMockRepo()
    handler := NewHandler(mockRepo, ...)
    app.Method("/path", handler.Function)

    // Test
    body := bytes.NewReader([]byte(`{"field":"value"}`))
    req := httptest.NewRequest(http.MethodPost, "/path", body)
    req.Header.Set("Content-Type", "application/json")

    resp, _ := app.Test(req)
    defer resp.Body.Close()

    // Verify
    if resp.StatusCode != http.StatusExpected {
        t.Fatalf("expected %d, got %d", http.StatusExpected, resp.StatusCode)
    }
}
```

### Template for Middleware Tests
```go
func TestMiddleware_Scenario(t *testing.T) {
    app := fiber.New()
    mockRedis, mock := redismock.NewClientMock()
    
    // Setup expectations
    mock.ExpectGet("key").SetVal("value")
    
    middleware := YourMiddleware(mockRedis)
    app.Get("/test", middleware, handler)
    
    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    resp, _ := app.Test(req)
    
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Error(err)
    }
}
```

## CI/CD Integration

Tests should run automatically on:
- Pull requests (before merge)
- Commits to main branch
- Pre-commit hook

Example GitHub Actions:
```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
      - run: go test -v ./...
```

## Known Limitations

1. **No end-to-end tests** - Tests use mocks, not real database
2. **No UI tests** - Frontend tests not yet implemented
3. **No performance tests** - Load testing not included
4. **No integration tests** - Real service integration not tested

## Future Test Additions

- [ ] E2E tests using testcontainers (real PostgreSQL/Redis)
- [ ] Frontend tests (Jest/React Testing Library)
- [ ] Load tests (Apache JMeter/Go-benchmark)
- [ ] Security tests (OWASP Top 10)
- [ ] API integration tests (real API clients)
- [ ] Database migration tests

## Troubleshooting

### Tests Hang
- Check for deadlocks in mock setups
- Ensure all mock expectations are met
- Check context cancellation

### Redis Mock Issues
- Verify redis mock `ExpectationsWereMet()` is called
- Check mock setup order matches actual calls
- Ensure error handling is correct

### UUID Parse Errors
- Use `uuid.New()` for test data generation
- Validate UUID format in request bodies
- Check parameter binding

## Resources

- [Go testing docs](https://golang.org/pkg/testing/)
- [testify/assert](https://github.com/stretchr/testify)
- [Redis mock](https://github.com/go-redis/redismock)
- [Fiber testing](https://docs.gofiber.io/guide/testing)
