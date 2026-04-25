# Bug Report: Masaar CRM

## Critical Bugs Found

### 🔴 BUG #1: Duplicate `CheckBlacklist` Function Declaration (COMPILATION ERROR)

**Location:** `internal/api/middleware/auth.go`
- Lines 13-28: First implementation using `c.Locals("user")`
- Lines 81-94: Second implementation using `BearerToken(c)`

**Issue:** Go does not allow two functions with the same name in the same package. This will cause a compilation error: `redeclared in this block`.

**Impact:** 
- Code does not compile
- Application cannot start

**Fix:** Remove the first (less secure) implementation at lines 13-28 and keep the second implementation that correctly uses `BearerToken(c)`.

---

### 🔴 BUG #2: Fail-Open Token Revocation in Logout Handler

**Location:** `internal/api/handler/auth.go`, lines 145-150

```go
if body.RefreshToken != "" {
    h.redis.Del(context.Background(), fmt.Sprintf("refresh:%s", body.RefreshToken))  // Error ignored!
}
if token := middleware.BearerToken(c); token != "" {
    h.redis.Set(context.Background(), "blacklist:"+token, "1",
        time.Duration(h.config.JWTAccessExpiryMin)*time.Minute)  // Error ignored!
}
```

**Issue:** 
- Both `h.redis.Del()` and `h.redis.Set()` calls ignore errors
- If Redis fails, logout returns HTTP 204 (success) but tokens are NOT revoked
- User believes they're logged out when they're actually still authenticated

**Impact:** 
- Security vulnerability: Revoked tokens remain valid if Redis is unavailable
- User session management fails silently

**Severity:** CRITICAL

**Fix:** Check errors from Redis operations and return HTTP 500 if they fail:
```go
if body.RefreshToken != "" {
    if err := h.redis.Del(context.Background(), fmt.Sprintf("refresh:%s", body.RefreshToken)).Err(); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "logout failed"})
    }
}
if token := middleware.BearerToken(c); token != "" {
    if err := h.redis.Set(context.Background(), "blacklist:"+token, "1",
        time.Duration(h.config.JWTAccessExpiryMin)*time.Minute).Err(); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "logout failed"})
    }
}
```

---

### 🔴 BUG #3: Fail-Open Token Blacklist Validation

**Location:** `internal/api/middleware/auth.go`, lines 84-85

```go
token := BearerToken(c)
if token != "" {
    exists, _ := rdb.Exists(c.Context(), "blacklist:"+token).Result()  // Error ignored!
    if exists > 0 {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "token revoked",
        })
    }
}
```

**Issue:** 
- Error from `rdb.Exists()` is ignored with `_`
- If Redis fails or is unavailable, `exists` defaults to 0
- Middleware proceeds as if token is valid (fail-open)

**Impact:** 
- Security vulnerability: Revoked tokens bypass validation during Redis outages
- Users with revoked tokens can still access the API

**Severity:** CRITICAL

**Fix:** Check Redis errors and fail closed:
```go
exists, err := rdb.Exists(c.Context(), "blacklist:"+token).Result()
if err != nil {
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to verify token"})
}
if exists > 0 {
    return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "token revoked"})
}
```

---

### 🟠 BUG #4: Security: JWT Token Exposed in URL Query Parameter

**Location:** `web/lib/api.ts`, lines 131-134

```typescript
getPDF: (id: string) => {
  const token = getToken()
  return `${BASE}/api/v1/invoices/${id}/pdf`
    + (token ? `?token=${token}` : '')
}
```

**Issue:** 
- JWT token is passed as a URL query parameter
- Tokens in URLs are logged in:
  - Browser history
  - Server access logs
  - Proxy/cache logs
  - Referrer headers

**Impact:** 
- Token exposure across logs and caches
- Increased risk of token compromise

**Severity:** HIGH

**Fix:** Use Authorization header instead by having the backend serve PDF with proper authentication:
```typescript
getPDF: (id: string) => {
  return `${BASE}/api/v1/invoices/${id}/pdf`
}
```
And ensure the backend handler checks the Authorization header.

---

### 🟠 BUG #5: Missing Validation of Lead/Deal Stage Values

**Location:** 
- `internal/api/handler/lead.go`, lines 87-98 (UpdateStage)
- `internal/api/handler/deal.go` (similar issue)

**Issue:** 
- Handlers accept any `stage` value from the client
- No validation against enum of allowed stages
- Invalid stages can be stored in database

**Example:**
```go
var body struct {
    Stage domain.LeadStage `json:"stage"`
}
if err := c.BodyParser(&body); err != nil || body.Stage == "" {
    return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "stage is required"})
}
// No validation that body.Stage is one of: new|contacted|qualified|proposal|won|lost
if err := h.leads.UpdateStage(c.Context(), id, body.Stage); err != nil {
```

**Impact:** 
- Invalid application state (leads in non-existent stages)
- Potential data corruption
- API contract violation (Swagger docs specify allowed stages)

**Severity:** MEDIUM

**Fix:** Validate stage against allowed values:
```go
validStages := map[domain.LeadStage]bool{
    domain.StageNew:       true,
    domain.StageContacted: true,
    domain.StageQualified: true,
    domain.StageProposal:  true,
    domain.StageWon:       true,
    domain.StageLost:      true,
}
if !validStages[body.Stage] {
    return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
        "error": "invalid stage",
    })
}
```

---

### 🟡 BUG #6: No Context Timeout on WhatsApp Webhook Processing

**Location:** `internal/api/handler/whatsapp.go`, line 136 (Receive function)

**Issue:** 
- Webhook handler uses `c.Context()` which inherits from the HTTP request context
- If the handler takes too long, Fiber might timeout and close the connection
- WhatsApp doesn't get a 200 response, causing retries
- No explicit timeout on database operations inside webhook handling

**Impact:** 
- Webhook message loss during slow database operations
- Unnecessary Meta API retries
- Potential duplicate message processing

**Severity:** MEDIUM

**Fix:** Consider adding explicit timeout handling or async processing of webhook data.

---

## Summary

| Bug | Severity | Type | Fixed? |
|-----|----------|------|--------|
| Duplicate CheckBlacklist | CRITICAL | Compilation Error | ❌ |
| Logout fail-open | CRITICAL | Security | ❌ |
| Blacklist validation fail-open | CRITICAL | Security | ❌ |
| JWT in URL | HIGH | Security | ❌ |
| Missing stage validation | MEDIUM | Data Integrity | ❌ |
| No webhook timeout | MEDIUM | Reliability | ❌ |

## Recommended Priority

1. **Fix BUG #1** immediately - code won't compile
2. **Fix BUG #2 & #3** immediately - critical security issues
3. **Fix BUG #4** before production - security best practice
4. **Fix BUG #5** before next feature release - data integrity
5. **Review BUG #6** - depends on webhook processing requirements
