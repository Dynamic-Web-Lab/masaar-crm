## 2026-04-01 - [Missing Token Revocation Validation]
**Vulnerability:** JWT tokens are not checked against the Redis blacklist during authentication, allowing revoked tokens to remain active until expiration.
**Learning:** Relying solely on JWT signature and expiration without checking the blacklist (session revocation store) leads to session replay vulnerabilities upon logout.
**Prevention:** Implement a CheckBlacklist middleware that queries the revocation store for the token extracted from the Authorization header.
## 2026-04-14 - [Fail-Closed Enforcement on Token Revocation]
**Vulnerability:** The Redis blacklist verification (`CheckBlacklist`) silently ignored connection/query errors, allowing requests with potentially revoked tokens to proceed if Redis was unreachable. Similarly, the `Logout` handler falsely returned success even if the revocation operations failed.
**Learning:** Security-critical checks (like session validation) and state changes (like session revocation) must fail-closed. If the system cannot verify a token's validity, it must reject the request. If it cannot complete a revocation, it must return an error to prevent silent failure.
**Prevention:** Always explicitly check for and handle errors returned by the datastore (e.g., `rdb.Exists`, `rdb.Set`, `rdb.Del`) during token operations, returning an appropriate 5xx HTTP status code rather than proceeding.
## 2026-04-09 - [Fail Closed on Session Revocation]
**Vulnerability:** Redis operations for session revocation (`Del`, `Set` in logout) and token blacklisting check (`Exists` in middleware) were failing open (errors ignored), which allowed revoked tokens to be used during Redis outages.
**Learning:** Security-critical operations must always fail closed. Ignoring errors from revocation stores creates a dangerous fallback where users are assumed authenticated when the state cannot be verified.
**Prevention:** Explicitly check for errors on all Redis commands involved in session management and return a 500 error, rather than continuing or returning success when the operation actually failed.
## 2026-04-24 - [Secure Blacklist Check]
**Vulnerability:** JWT token blacklist checking was failing open when Redis failed, allowing revoked tokens access.
**Learning:** Security controls like token revocation checks must fail closed on infrastructure errors to prevent bypasses.
**Prevention:** Explicitly check for errors on infrastructure dependencies (like Redis `Exists`) and return 500 Internal Server Error when they occur.
