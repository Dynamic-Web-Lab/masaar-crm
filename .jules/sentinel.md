## 2026-04-01 - [Missing Token Revocation Validation]
**Vulnerability:** JWT tokens are not checked against the Redis blacklist during authentication, allowing revoked tokens to remain active until expiration.
**Learning:** Relying solely on JWT signature and expiration without checking the blacklist (session revocation store) leads to session replay vulnerabilities upon logout.
**Prevention:** Implement a CheckBlacklist middleware that queries the revocation store for the token extracted from the Authorization header.

## 2026-04-01 - [Fail-Closed Security Operations]
**Vulnerability:** Security critical operations (like token validation in middleware or session invalidation on logout) were silently ignoring infrastructure errors (e.g., Redis failures), potentially failing "open" and allowing unauthorized access or failing to revoke access correctly.
**Learning:** Security operations must always implement a "fail-closed" mechanism. If an underlying service (like a database or cache) fails during an authorization or invalidation check, the system must deny access or abort the operation safely, rather than allowing it to proceed.
**Prevention:** Explicitly check for all errors during security operations and return appropriate HTTP error statuses (like 500 Internal Server Error) when infrastructure dependencies fail.
