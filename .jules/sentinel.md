## 2026-04-01 - [Missing Token Revocation Validation]
**Vulnerability:** JWT tokens are not checked against the Redis blacklist during authentication, allowing revoked tokens to remain active until expiration.
**Learning:** Relying solely on JWT signature and expiration without checking the blacklist (session revocation store) leads to session replay vulnerabilities upon logout.
**Prevention:** Implement a CheckBlacklist middleware that queries the revocation store for the token extracted from the Authorization header.
## 2026-04-01 - [Missing Fail-Closed Mechanism in Redis Checks]
**Vulnerability:** The CheckBlacklist middleware and Logout handler ignored errors from Redis when checking or setting the token blacklist. If Redis was unavailable, the operations would silently fail, potentially allowing revoked tokens to remain valid (failing open).
**Learning:** Security-critical operations like token revocation checks must use a 'fail-closed' mechanism. If the system cannot verify the revocation status due to an infrastructure failure (e.g., Redis down), it must assume the token is invalid or block the request, rather than allowing it to proceed.
**Prevention:** Always handle errors from external stores (databases, caches) in security middleware and handlers. Return an appropriate HTTP error (like 500 Internal Server Error) instead of ignoring the error.
