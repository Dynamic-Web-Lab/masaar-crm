## 2026-04-01 - [Missing Token Revocation Validation]
**Vulnerability:** JWT tokens are not checked against the Redis blacklist during authentication, allowing revoked tokens to remain active until expiration.
**Learning:** Relying solely on JWT signature and expiration without checking the blacklist (session revocation store) leads to session replay vulnerabilities upon logout.
**Prevention:** Implement a CheckBlacklist middleware that queries the revocation store for the token extracted from the Authorization header.

## 2024-05-15 - [Redis Token Revocation Fail-Open Risk]
**Vulnerability:** Found security mechanisms (like `CheckBlacklist` middleware and `Logout` handler) that ignored errors from Redis cache (`Exists`, `Set`, `Del`), allowing the requests to proceed or succeed when the backend infrastructure failed (fail-open).
**Learning:** Security-critical operations must always verify success. Ignoring cache connection/execution errors means that if the cache goes down, all previously revoked tokens become valid again, or token revokation silently fails.
**Prevention:** Explicitly check for errors from Redis and return a 500 Internal Server Error (fail-closed) to ensure that if authorization verification fails due to an infrastructure issue, the user is denied access rather than improperly granted it.

## 2024-05-15 - [Empty Raw JWT Token Bypass]
**Vulnerability:** The previous implementation extracted the token string from `c.Locals("user").(*jwt.Token).Raw`. The `.Raw` field is sometimes empty depending on how the token was parsed, allowing the token to bypass blacklist checks.
**Learning:** Relying on the `.Raw` attribute of a parsed JWT token is not a reliable way to get the original token string for blacklist verification because standard middleware may not populate it consistently.
**Prevention:** Extract the raw token string directly from the `Authorization` HTTP header using `BearerToken(c)` whenever possible for manual token verification like checking a blacklist.
## 2026-04-01 - [Missing Fail-Closed Mechanism in Redis Checks]
**Vulnerability:** The CheckBlacklist middleware and Logout handler ignored errors from Redis when checking or setting the token blacklist. If Redis was unavailable, the operations would silently fail, potentially allowing revoked tokens to remain valid (failing open).
**Learning:** Security-critical operations like token revocation checks must use a 'fail-closed' mechanism. If the system cannot verify the revocation status due to an infrastructure failure (e.g., Redis down), it must assume the token is invalid or block the request, rather than allowing it to proceed.
**Prevention:** Always handle errors from external stores (databases, caches) in security middleware and handlers. Return an appropriate HTTP error (like 500 Internal Server Error) instead of ignoring the error.
## 2025-05-24 - [Fail-Open in Redis Authentication Operations]
**Vulnerability:** Redis operations in `Logout` (adding to blacklist) and `CheckBlacklist` (verifying blacklist) were failing open. Errors (like connection failures) were silently ignored, potentially allowing revoked tokens to continue accessing the system.
**Learning:** Security-critical dependency operations must always handle errors explicitly and fail securely (fail-closed).
**Prevention:** Always check `.Err()` on `go-redis` operations and return a `500 Internal Server Error` if a security verification cannot be completed.
## 2026-04-14 - [Fail-Closed Enforcement on Token Revocation]
**Vulnerability:** The Redis blacklist verification (`CheckBlacklist`) silently ignored connection/query errors, allowing requests with potentially revoked tokens to proceed if Redis was unreachable. Similarly, the `Logout` handler falsely returned success even if the revocation operations failed.
**Learning:** Security-critical checks (like session validation) and state changes (like session revocation) must fail-closed. If the system cannot verify a token's validity, it must reject the request. If it cannot complete a revocation, it must return an error to prevent silent failure.
**Prevention:** Always explicitly check for and handle errors returned by the datastore (e.g., `rdb.Exists`, `rdb.Set`, `rdb.Del`) during token operations, returning an appropriate 5xx HTTP status code rather than proceeding.
## 2026-04-09 - [Fail Closed on Session Revocation]
**Vulnerability:** Redis operations for session revocation (`Del`, `Set` in logout) and token blacklisting check (`Exists` in middleware) were failing open (errors ignored), which allowed revoked tokens to be used during Redis outages.
**Learning:** Security-critical operations must always fail closed. Ignoring errors from revocation stores creates a dangerous fallback where users are assumed authenticated when the state cannot be verified.
**Prevention:** Explicitly check for errors on all Redis commands involved in session management and return a 500 error, rather than continuing or returning success when the operation actually failed.
## 2026-04-24 - [Secure Blacklist Check]
**Vulnerability:** JWT token blacklist checking was failing open when Redis failed, allowing revoked tokens access.
**Learning:** Security controls like token revocation checks must fail closed on infrastructure errors to prevent bypasses.
**Prevention:** Explicitly check for errors on infrastructure dependencies (like Redis `Exists`) and return 500 Internal Server Error when they occur.
## 2025-05-24 - [Webhook Token Timing Attack]\n**Vulnerability:** Meta webhook token verification used a standard string equality check (`==`), which is vulnerable to timing attacks.\n**Learning:** Standard string comparisons fail early on a mismatch, potentially leaking the expected token character by character based on response times.\n**Prevention:** Use `crypto/subtle.ConstantTimeCompare` for verifying secrets, signatures, and tokens.

## 2024-05-27 - [Fix Denial of Service in Bank Statement Upload]
**Vulnerability:** The bank statement upload handler used string slicing `file.Filename[len(file.Filename)-4:]` to extract the file extension. If a user uploaded a file with a name shorter than 4 characters (e.g., `a.c`), it caused an index out-of-bounds panic, leading to a server crash (Denial of Service).
**Learning:** Manual string manipulation for extracting parts of user-provided filenames is prone to edge cases and errors.
**Prevention:** Always use standard library functions like `filepath.Ext(filename)` for safe and robust file extension extraction, and combine it with `strings.ToLower` for case-insensitive comparisons.
