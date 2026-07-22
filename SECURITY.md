# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in Masaar CRM, please report it responsibly.

**Do NOT open a public GitHub issue for security vulnerabilities.**

### How to Report

1. Email: **security@dynamicweblab.com**
2. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

### What to Expect

- Acknowledgment within 48 hours
- Assessment within 1 week
- Fix timeline communicated upon confirmation
- Credit in release notes (unless you prefer anonymity)

## Security Best Practices for Self-Hosters

### Environment Variables

- **Never** use default values for `JWT_SECRET`, `DB_PASSWORD`, or `REDIS_PASSWORD` in production.
- Generate strong secrets:
  ```bash
  openssl rand -hex 32
  ```

### Network Security

- Restrict database and Redis ports to localhost only.
- Use HTTPS in production (reverse proxy with Let's Encrypt recommended).
- Set `ALLOWED_ORIGINS` to your actual frontend domain.

### WhatsApp Webhooks

- Always configure `WA_APP_SECRET` to validate inbound webhook signatures.
- Never expose WhatsApp credentials in client-side code.

### API Keys

- API keys are stored as SHA-256 hashes in the database.
- Only the first 8 characters are stored in plaintext for identification.
- Rotate keys regularly.

### Updates

- Keep dependencies updated (`go mod tidy`, `npm audit fix`).
- Monitor [GitHub Security Advisories](https://github.com/dynamicweblab/masaar-crm/security) for updates.

## PDPL Compliance

Masaar CRM is designed to comply with UAE Personal Data Protection Law (PDPL):

- All customer data stays on your self-hosted infrastructure.
- AI processing (Ollama) runs locally - no PII sent to external services.
- Audit logs track all data access and modifications.
- Role-based access control enforces least-privilege principle.

## Supported Versions

| Version | Supported |
|---------|-----------|
| Latest  | Yes       |
| < Latest | No       |

Always run the latest version for security patches.
