## 2026-04-01 - [Missing Token Revocation Validation]
**Vulnerability:** JWT tokens are not checked against the Redis blacklist during authentication, allowing revoked tokens to remain active until expiration.
**Learning:** Relying solely on JWT signature and expiration without checking the blacklist (session revocation store) leads to session replay vulnerabilities upon logout.
**Prevention:** Implement a CheckBlacklist middleware that queries the revocation store for the token extracted from the Authorization header.

## 2026-04-01 - [Fail-Open Error Handling in Token Revocation]
**Vulnerability:** The Redis interactions for token revocation checking (CheckBlacklist middleware) and invalidation (Logout handler) ignored errors. If Redis became unreachable, they would fail open, allowing revoked tokens to continue accessing endpoints and falsely acknowledging successful logouts.
**Learning:** Security-critical operations involving external stores (like a Redis blacklist) must explicitly handle database or cache errors and fail closed. Ignoring errors can easily bypass token revocation mechanisms.
**Prevention:** Always check the error returned by cache/database operations (e.g., `Exists`, `Del`, `Set`). If an error occurs, return an appropriate HTTP status (e.g., 500 Internal Server Error) to deny the request or indicate the failure explicitly.
