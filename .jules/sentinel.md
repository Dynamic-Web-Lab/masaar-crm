## 2026-04-01 - [Missing Token Revocation Validation]
**Vulnerability:** JWT tokens are not checked against the Redis blacklist during authentication, allowing revoked tokens to remain active until expiration.
**Learning:** Relying solely on JWT signature and expiration without checking the blacklist (session revocation store) leads to session replay vulnerabilities upon logout.
**Prevention:** Implement a CheckBlacklist middleware that queries the revocation store for the token extracted from the Authorization header.
## 2025-05-24 - [Fail-Open in Redis Authentication Operations]
**Vulnerability:** Redis operations in `Logout` (adding to blacklist) and `CheckBlacklist` (verifying blacklist) were failing open. Errors (like connection failures) were silently ignored, potentially allowing revoked tokens to continue accessing the system.
**Learning:** Security-critical dependency operations must always handle errors explicitly and fail securely (fail-closed).
**Prevention:** Always check `.Err()` on `go-redis` operations and return a `500 Internal Server Error` if a security verification cannot be completed.
