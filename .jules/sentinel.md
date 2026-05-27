## 2025-02-21 - [Path Traversal in File Uploads]
**Vulnerability:** Found unsanitized user-provided filenames (`file.Filename`) being used directly to construct file URLs and storage paths during file upload in `internal/api/handler/bank_statement.go`. This could lead to Path Traversal vulnerabilities or arbitrary file writes if the file is saved locally, and could be exploited to cause XSS via crafted file names.
**Learning:** Developers often trust user-provided filenames from multipart/form-data directly. The framework (Fiber) doesn't sanitize `file.Filename` by default.
**Prevention:** Always sanitize any user-provided filename using `filepath.Base(file.Filename)` before using it in storage, paths, or URLs.

## 2026-05-05 - [SSRF via Outbound Webhooks]
**Vulnerability:** Found an SSRF vulnerability where user-provided webhook URLs could be used to send requests to restricted IP addresses (private, loopback, unspecified, and link-local) from the application server. This could lead to accessing internal services or scanning internal networks.
**Learning:** Developers frequently implement webhooks or fetch remote URLs using a default `http.Client`. Without restricting the resolved IP address, this leaves the application vulnerable to Server-Side Request Forgery and DNS Rebinding.
**Prevention:** Use a custom `net.Dialer` with a `Control` hook to check the resolved IP and block restricted ranges before the connection is established. This ensures both security and performance by blocking the connection at the transport layer while preserving connection pooling and HTTP/2.
