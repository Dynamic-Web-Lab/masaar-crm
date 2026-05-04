## 2025-02-21 - [Path Traversal in File Uploads]
**Vulnerability:** Found unsanitized user-provided filenames (`file.Filename`) being used directly to construct file URLs and storage paths during file upload in `internal/api/handler/bank_statement.go`. This could lead to Path Traversal vulnerabilities or arbitrary file writes if the file is saved locally, and could be exploited to cause XSS via crafted file names.
**Learning:** Developers often trust user-provided filenames from multipart/form-data directly. The framework (Fiber) doesn't sanitize `file.Filename` by default.
**Prevention:** Always sanitize any user-provided filename using `filepath.Base(file.Filename)` before using it in storage, paths, or URLs.
