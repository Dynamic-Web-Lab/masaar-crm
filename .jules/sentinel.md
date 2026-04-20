## 2026-04-01 - [Missing Token Revocation Validation]
**Vulnerability:** JWT tokens are not checked against the Redis blacklist during authentication, allowing revoked tokens to remain active until expiration.
**Learning:** Relying solely on JWT signature and expiration without checking the blacklist (session revocation store) leads to session replay vulnerabilities upon logout.
**Prevention:** Implement a CheckBlacklist middleware that queries the revocation store for the token extracted from the Authorization header.

## 2023-10-27 - [Missing Fail-Closed Handling on Cache/Database Security Checks]
**Vulnerability:** Security-critical operations like token revocation (Logout) and verification (CheckBlacklist) ignored Redis errors, failing open and potentially allowing access to revoked tokens if Redis goes down.
**Learning:** External dependencies (like Redis/Postgres) can fail. Without proper error handling in security middleware/handlers, an infrastructure failure can result in a security bypass.
**Prevention:** Always check error returns on cache/database lookups in security-related logic, and explicitly fail-closed (e.g. return HTTP 500) if the state cannot be verified.
