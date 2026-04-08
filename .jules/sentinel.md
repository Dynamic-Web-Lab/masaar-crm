## 2026-04-01 - [Missing Token Revocation Validation]
**Vulnerability:** JWT tokens are not checked against the Redis blacklist during authentication, allowing revoked tokens to remain active until expiration.
**Learning:** Relying solely on JWT signature and expiration without checking the blacklist (session revocation store) leads to session replay vulnerabilities upon logout.
**Prevention:** Implement a CheckBlacklist middleware that queries the revocation store for the token extracted from the Authorization header.

## 2026-04-01 - [Missing Fail-Closed Mechanism in Token Validation]
**Vulnerability:** The CheckBlacklist middleware ignored Redis errors (e.g., connection issues) and failed open, allowing requests to proceed even if token revocation couldn't be verified.
**Learning:** Security-critical middlewares, especially those checking token revocation, must fail closed. Ignoring errors in these checks can lead to auth bypasses during service degradation.
**Prevention:** Always handle errors from backend stores (like Redis or databases) in security middlewares and return appropriate error statuses (e.g., 500 Internal Server Error) to block access when verification fails.
