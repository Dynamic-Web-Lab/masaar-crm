## 2026-04-01 - [Missing Token Revocation Validation]
**Vulnerability:** JWT tokens are not checked against the Redis blacklist during authentication, allowing revoked tokens to remain active until expiration.
**Learning:** Relying solely on JWT signature and expiration without checking the blacklist (session revocation store) leads to session replay vulnerabilities upon logout.
**Prevention:** Implement a CheckBlacklist middleware that queries the revocation store for the token extracted from the Authorization header.
## 2026-04-15 - [Fail-Open Vulnerability in Session Management]
**Vulnerability:** Redis connection errors during token revocation (logout) and blacklist validation were silently ignored, causing the application to fail open and allow revoked tokens to proceed during database outages.
**Learning:** Security-critical operations must enforce a 'fail-closed' mechanism. Swallowing cache/database errors when verifying authentication tokens allows attackers to bypass revocation mechanisms.
**Prevention:** Always explicitly check for and handle errors returned by the datastore in authentication middlewares and handlers, returning appropriate HTTP 500 error statuses if the operation fails.
