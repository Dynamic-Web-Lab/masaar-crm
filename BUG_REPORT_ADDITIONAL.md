# Additional Bug Report: Masaar CRM (After Merge)

## 3 New Bugs Found Post-Merge

### BUG #7: MEDIUM - Missing Invoice Status Validation

**Location:** `internal/api/handler/invoice.go`, lines 122-137 (UpdateStatus)

**Issue:**
```go
var body struct {
    Status domain.InvoiceStatus `json:"status"`
}
if err := c.BodyParser(&body); err != nil || body.Status == "" {
    return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "status required"})
}
// No validation - accepts any status value!
if err := h.invoices.UpdateStatus(c.Context(), id, body.Status); err != nil {
```

**Problem:** No enum validation for invoice status. Invalid statuses can be stored in database.

**Valid Values:** `draft | sent | paid` (from domain.InvoiceStatus)

**Impact:** Data integrity issue - invoices can be set to invalid states

**Severity:** MEDIUM

**Fix:**
```go
validStatuses := map[domain.InvoiceStatus]bool{
    domain.InvoiceDraft: true,
    domain.InvoiceSent:  true,
    domain.InvoicePaid:  true,
}
if !validStatuses[body.Status] {
    return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid status"})
}
```

---

### BUG #8: LOW - String Comparison Instead of Domain Constant

**Location:** `internal/api/handler/ai.go`, line 61 (DraftReply)

**Issue:**
```go
for _, m := range msgs {
    prefix := "Agent"
    if m.Direction == "inbound" {  // ❌ String literal comparison
        prefix = "Customer"
    }
    bodies = append(bodies, prefix+": "+m.Body)
}
```

**Problem:** 
- Uses string literal `"inbound"` instead of domain constant `domain.DirectionInbound`
- If MessageDirection constant values change, this breaks silently
- Harder to maintain and prone to typos

**Correct constant value:** `domain.DirectionInbound = MessageDirection("inbound")`

**Severity:** LOW (Low impact, but bad practice)

**Fix:**
```go
if m.Direction == domain.DirectionInbound {
    prefix = "Customer"
}
```

---

### BUG #9: LOW - Potential Nil Check Missing

**Location:** `internal/api/handler/ai.go`, lines 67-69 (DraftReply)

**Issue:**
```go
threads, err := h.wa.ListThreads(c.Context(), "", 1, 1)
if err != nil || len(threads) == 0 {  // ✓ Length check is safe, but...
    return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "thread not found"})
}

summary, _ := h.ollama.SummarizeThread(c.Context(), bodies)
contact, _ := h.contacts.GetByID(c.Context(), threads[0].ContactID)  // threads[0] safe here
```

**Problem:**
- Lines 72-73 ignore errors with `_` on potentially failing operations
- If SummarizeThread or GetByID fail, errors are silently discarded
- Subsequent code uses `summary` and `contact` without knowing if they're valid

**Severity:** LOW (Non-critical paths, but poor error handling)

**Fix:**
```go
summary, err := h.ollama.SummarizeThread(c.Context(), bodies)
if err != nil {
    return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "AI service unavailable"})
}

contact, err := h.contacts.GetByID(c.Context(), threads[0].ContactID)
if err != nil {
    // Handle appropriately - contact may be optional
    contact = nil
}
```

---

## Summary Table

| Bug # | Issue | Severity | Type | File |
|-------|-------|----------|------|------|
| 7 | Missing invoice status validation | MEDIUM | Data Integrity | invoice.go |
| 8 | String literal vs domain constant | LOW | Code Quality | ai.go |
| 9 | Errors ignored in AI handler | LOW | Error Handling | ai.go |

## Total Bugs Found
- **Critical:** 3 (fixed)
- **High:** 1 (fixed)
- **Medium:** 2 (1 fixed, 1 new)
- **Low:** 2 (new)

**Total:** 8 bugs identified (5 fixed, 3 remaining)
