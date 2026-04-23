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
