## 2026-04-01 - [Missing Token Revocation Validation]
**Vulnerability:** JWT tokens are not checked against the Redis blacklist during authentication, allowing revoked tokens to remain active until expiration.
**Learning:** Relying solely on JWT signature and expiration without checking the blacklist (session revocation store) leads to session replay vulnerabilities upon logout.
**Prevention:** Implement a CheckBlacklist middleware that queries the revocation store for the token extracted from the Authorization header.
## 2026-04-09 - [Fail Closed on Session Revocation]
**Vulnerability:** Redis operations for session revocation (`Del`, `Set` in logout) and token blacklisting check (`Exists` in middleware) were failing open (errors ignored), which allowed revoked tokens to be used during Redis outages.
**Learning:** Security-critical operations must always fail closed. Ignoring errors from revocation stores creates a dangerous fallback where users are assumed authenticated when the state cannot be verified.
**Prevention:** Explicitly check for errors on all Redis commands involved in session management and return a 500 error, rather than continuing or returning success when the operation actually failed.
## 2026-04-24 - [Secure Blacklist Check]
**Vulnerability:** JWT token blacklist checking was failing open when Redis failed, allowing revoked tokens access.
**Learning:** Security controls like token revocation checks must fail closed on infrastructure errors to prevent bypasses.
**Prevention:** Explicitly check for errors on infrastructure dependencies (like Redis `Exists`) and return 500 Internal Server Error when they occur.
