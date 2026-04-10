## 2026-04-01 - [Missing Token Revocation Validation]
**Vulnerability:** JWT tokens are not checked against the Redis blacklist during authentication, allowing revoked tokens to remain active until expiration.
**Learning:** Relying solely on JWT signature and expiration without checking the blacklist (session revocation store) leads to session replay vulnerabilities upon logout.
**Prevention:** Implement a CheckBlacklist middleware that queries the revocation store for the token extracted from the Authorization header.

## 2026-04-01 - [Fail-Closed Security Operations]
**Vulnerability:** Security-critical operations like `CheckBlacklist` middleware and `Logout` handlers were ignoring or bypassing errors when connecting to Redis, allowing revoked tokens to remain active or failing to fully revoke tokens on logout.
**Learning:** Failing open or silently ignoring errors during authentication or session management creates a false sense of security, leading to severe vulnerabilities like session replay.
**Prevention:** Always implement a fail-closed mechanism in security-critical paths; if a database or cache query fails during authentication or revocation, the operation must immediately return a 500 Internal Server Error rather than proceeding.
