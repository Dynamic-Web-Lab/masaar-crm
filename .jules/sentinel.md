## 2026-04-01 - [Missing Token Revocation Validation]
**Vulnerability:** JWT tokens are not checked against the Redis blacklist during authentication, allowing revoked tokens to remain active until expiration.
**Learning:** Relying solely on JWT signature and expiration without checking the blacklist (session revocation store) leads to session replay vulnerabilities upon logout.
**Prevention:** Implement a CheckBlacklist middleware that queries the revocation store for the token extracted from the Authorization header.
## 2024-04-22 - [Fail-Open Redis Token Verification]
**Vulnerability:** Redis token verification operations (e.g., `CheckBlacklist` middleware and `Logout` handler) failed open if Redis returned an error, allowing revoked tokens to remain valid or logout procedures to silently fail.
**Learning:** Security-critical infrastructure calls like session revocation checks must handle connection or execution errors by failing closed, otherwise a database outage can lead to an authentication bypass.
**Prevention:** Always check `err != nil` for Redis operations (`rdb.Exists`, `rdb.Del`, `rdb.Set`) during token validation/revocation and explicitly return a 500 HTTP error instead of proceeding.
