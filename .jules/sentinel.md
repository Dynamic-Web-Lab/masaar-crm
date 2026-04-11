## 2026-04-01 - [Missing Token Revocation Validation]
**Vulnerability:** JWT tokens are not checked against the Redis blacklist during authentication, allowing revoked tokens to remain active until expiration.
**Learning:** Relying solely on JWT signature and expiration without checking the blacklist (session revocation store) leads to session replay vulnerabilities upon logout.
**Prevention:** Implement a CheckBlacklist middleware that queries the revocation store for the token extracted from the Authorization header.
## 2025-04-11 - [Fail-Open Vulnerability in Blacklist Check]
**Vulnerability:** The CheckBlacklist middleware ignored Redis connection or lookup errors. If the Redis server was unreachable, the check failed open, allowing revoked tokens to bypass authentication.
**Learning:** Security controls that rely on external systems (like a database or cache) must fail securely. Ignoring errors leads to implicit trust.
**Prevention:** Always explicitly check for errors when interacting with security-critical dependencies (e.g., `err != nil`) and return an appropriate HTTP error status (like 500 Internal Server Error) to fail closed.
