# Contributing to Masaar CRM

Thank you for your interest in contributing to Masaar CRM! This document provides guidelines and instructions for contributing.

## Code of Conduct

Please read our [Code of Conduct](CODE_OF_CONDUCT.md) before contributing. We expect all contributors to follow it.

## How to Contribute

### Reporting Bugs

1. Check [existing issues](https://github.com/dynamicweblab/masaar-crm/issues) to avoid duplicates.
2. Open a new issue with a clear title and description.
3. Include steps to reproduce, expected behavior, and actual behavior.
4. Mention your OS, browser, and deployment method (Docker, manual, etc.).

### Suggesting Features

1. Open an issue with the `enhancement` label.
2. Describe the feature, why it's needed, and how it should work.
3. For UAE-specific features, mention relevant regulations (PDPL, VAT, etc.).

### Submitting Changes

1. Fork the repository.
2. Create a feature branch from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. Make your changes following the coding standards below.
4. Test your changes locally with Docker Compose.
5. Commit with a clear message:
   ```bash
   git commit -m "feat: add new feature description"
   ```
6. Push to your fork and open a Pull Request.

## Development Setup

### Prerequisites

- Docker and Docker Compose
- Go 1.22+ (for backend development)
- Node.js 18+ (for frontend development)

### Local Development

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/masaar-crm.git
cd masaar-crm

# Copy environment file
cp .env.example .env

# Start dependencies only
docker compose up -d postgres redis ollama

# Run the backend
go run ./cmd/server

# In another terminal, run the frontend
cd web
npm install
npm run dev
```

### Available Commands

| Command | Description |
|---------|-------------|
| `docker compose up` | Start all services |
| `go run ./cmd/server` | Run backend only |
| `cd web && npm run dev` | Run frontend only |
| `swag init -g cmd/server/main.go` | Regenerate Swagger docs |

## Coding Standards

### Go (Backend)

- Follow [Effective Go](https://go.dev/doc/effective_go) conventions.
- Use the handler-repo pattern for new endpoints.
- Always capture both return values from `uuid.Parse()`:
  ```go
  // Correct
  id, _ := uuid.Parse(c.Locals("company_id").(string))
  
  // Wrong - compile error
  id := uuid.Parse(c.Locals("company_id").(string))
  ```
- Scope all queries by `company_id` from JWT claims.
- Add Swagger comments for new endpoints.

### TypeScript/React (Frontend)

- Use functional components with hooks.
- Use the `useLang()` context for bilingual text.
- Use the API client in `web/lib/api.ts` for all API calls.
- Follow the existing component patterns in `web/components/`.

### Database

- Migrations go in `migrations/` with sequential numbering.
- Use goose format for migration files.
- Never modify applied migrations.

## Project Structure

```
masaar-crm/
├── cmd/server/          # Go backend entry point
├── internal/
│   ├── api/             # HTTP handlers and middleware
│   ├── domain/          # Models and types
│   ├── repo/            # PostgreSQL repositories
│   └── ws/              # WebSocket hub
├── migrations/          # SQL migrations (goose)
├── web/                 # Next.js frontend
└── docker/              # Docker configuration
```

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).

## Questions?

Open an issue for any questions about contributing. We're happy to help!
