## 2026-04-01 - [Missing Token Revocation Validation]
**Vulnerability:** JWT tokens are not checked against the Redis blacklist during authentication, allowing revoked tokens to remain active until expiration.
**Learning:** Relying solely on JWT signature and expiration without checking the blacklist (session revocation store) leads to session replay vulnerabilities upon logout.
**Prevention:** Implement a CheckBlacklist middleware that queries the revocation store for the token extracted from the Authorization header.
## 2026-04-12 - [Fail Open in Middleware Revocation Check]
**Vulnerability:** In `CheckBlacklist`, the error returned by Redis when checking the token blacklist was ignored (`exists, _ := rdb.Exists(...)`). If Redis failed (e.g., connection issue), the check would fail-open, allowing potentially revoked tokens to bypass validation and access authenticated routes.
**Learning:** Security-critical mechanisms, especially token validation, must not assume success upon encountering an error in an external dependency like a cache or database. They must be explicitly designed to fail-closed.
**Prevention:** Always check for and handle errors from external systems during authentication and authorization checks. If an error occurs, reject the request (e.g., return HTTP 500) rather than proceeding to `c.Next()`.
