## 2026-04-01 - [Missing Token Revocation Validation]
**Vulnerability:** JWT tokens are not checked against the Redis blacklist during authentication, allowing revoked tokens to remain active until expiration.
**Learning:** Relying solely on JWT signature and expiration without checking the blacklist (session revocation store) leads to session replay vulnerabilities upon logout.
**Prevention:** Implement a CheckBlacklist middleware that queries the revocation store for the token extracted from the Authorization header.

## 2026-04-13 - [Fail-Open in Security Critical Operations]
**Vulnerability:** Redis operations for token revocation and blacklist checking failed open silently when infrastructure encountered an error, allowing authorization bypass.
**Learning:** Security-critical operations, especially those relying on external infrastructure like Redis for token blacklist, must fail closed. If the revocation store is unreachable, we must reject the request rather than allow potentially revoked tokens to proceed.
**Prevention:** Always check error returns from data stores (like Redis) during authentication/authorization checks and session invalidations. Return a 500 error if validation cannot be completed safely.
