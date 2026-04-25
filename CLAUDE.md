# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Masaar CRM** is an open-source, self-hosted WhatsApp CRM built specifically for UAE businesses. It's a full-stack application combining a Go backend with a Next.js frontend, designed for Arabic-native interfaces and WhatsApp-first sales workflows.

**Key Characteristics:**
- Full Arabic/RTL support in UI
- PDPL-compliant (UAE data residency)
- Receives WhatsApp messages via Meta Cloud API webhooks
- Local AI thread summaries via Ollama LLM
- Real-time notifications over WebSockets
- Role-based access control (Admin, Agent, Viewer)

## Technology Stack

| Layer | Technology |
|-------|-----------|
| **Backend** | Go 1.22+ with Fiber web framework |
| **Frontend** | Next.js 14 with React 18, Tailwind CSS |
| **Database** | PostgreSQL 16 with pgvector extension |
| **Cache/Sessions** | Redis 7 |
| **AI/LLM** | Ollama (llama3/mistral) |
| **Migrations** | goose (SQL-based) |
| **Auth** | JWT (HS256) + bcrypt |
| **API Docs** | Swagger/OpenAPI via swaggo |
| **Real-time** | WebSocket (Fiber native) |

## Project Structure

```
masaar-crm/
├── cmd/server/             # Server entry point (main.go)
├── internal/
│   ├── api/                # HTTP handlers, middleware, routing
│   │   ├── handler/        # Endpoint handlers (auth, contacts, leads, etc)
│   │   ├── middleware/     # JWT auth, role checks, blacklist validation
│   │   ├── router.go       # Route registration and limiter setup
│   │   └── doc.go          # Swagger documentation config
│   ├── domain/             # Core models and types (User, Contact, Lead, etc)
│   ├── repo/               # PostgreSQL repository layer (data access)
│   ├── config/             # Configuration management
│   ├── ai/                 # Ollama LLM integration
│   ├── pdf/                # PDF generation for invoices
│   └── ws/                 # WebSocket hub for real-time notifications
├── web/                    # Next.js frontend
│   ├── app/                # App Router pages and layouts
│   ├── components/         # React UI components (Kanban, forms, etc)
│   ├── store/              # Zustand state management
│   ├── context/            # Language context (Arabic/English)
│   ├── hooks/              # Custom React hooks
│   └── lib/                # API client, auth utilities
├── migrations/             # SQL migration files (goose format)
├── docker/                 # Docker build files
├── docs/                   # Swagger/OpenAPI spec output
└── scripts/                # Seed and utility scripts
```

## Quick Start Commands

### Full Stack (Docker)
```bash
# Setup
cp .env.example .env
docker compose -f docker/docker-compose.yml up

# Access
- Dashboard: http://localhost:3000
- API: http://localhost:8080/api/v1
- Swagger UI: http://localhost:8080/docs
- Default login: admin@masaar.local / changeme
```

### Backend Development Only
```bash
# Start dependencies (no API)
docker compose -f docker/docker-compose.yml up -d postgres redis ollama

# Run server directly
go run ./cmd/server

# Build for production
go build -o masaar ./cmd/server
```

### Frontend Development Only
```bash
cd web
npm install
npm run dev    # Dev server on :3000
npm run build  # Production build
npm run lint   # ESLint check
```

## Development Workflows

### Backend Development

**1. Database Migrations**
- Migrations are SQL files in `migrations/` using goose format
- Naming: `00001_description.sql`
- Automatically run on server startup via `goose.Up()`
- To add a migration: create new SQL file and run server (it auto-applies)
- Goose docs: https://github.com/pressly/goose

**2. Adding a New Endpoint**
1. Add SQL schema in a migration if needed
2. Add domain models to `internal/domain/`
3. Add repository methods to `internal/repo/` for data access
4. Create handler in `internal/api/handler/` with Swagger comments
5. Register routes in `internal/api/router.go`
6. Add role-based middleware if needed via `middleware.RequireRole()`

**3. Building Swagger Docs**
```bash
# Install swag if not present
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs from handler comments
swag init -g cmd/server/main.go

# Swagger UI automatically available at /docs
```

**4. API Authentication**
- JWT middleware in `internal/api/middleware/auth.go`
- Tokens stored in Redis blacklist on logout
- All protected routes require `Authorization: Bearer <token>` header
- Role-based access via `middleware.RequireRole(domain.RoleAdmin, domain.RoleAgent)`

### Frontend Development

**1. Component Structure**
- Server and Client Components follow Next.js 14 conventions
- Use `'use client'` for interactive components
- Zustand store in `web/store/` for global state
- Language context at `web/context/` for AR/EN switching

**2. API Client**
- Located in `web/lib/` — axios instance with auth token injection
- Automatically uses `NEXT_PUBLIC_API_URL` env var

**3. Styling**
- Tailwind CSS configured in `tailwind.config.ts`
- RTL support via Tailwind's RTL plugin
- Custom CSS minimal — prefer utility classes

## Core Patterns

### Authentication Flow
1. User logs in: `POST /api/v1/auth/login` → returns access + refresh tokens
2. Access token valid for 15 minutes (JWT_ACCESS_EXPIRY_MIN)
3. Refresh token valid for 7 days (JWT_REFRESH_EXPIRY_DAYS)
4. On logout: token added to Redis blacklist, checked on every protected route
5. WebSocket upgrade requires valid token via same JWT middleware

### Role-Based Access Control
Three roles enforced at the handler level:
- **Admin**: Full access to all operations (delete contacts, manage users)
- **Agent**: Create/update leads, contacts, deals; manage notes
- **Viewer**: Read-only access to all resources (no write operations)

Check middleware in `internal/api/middleware/auth.go` for implementation.

### Real-time Notifications
1. WebSocket hub in `internal/ws/hub.go` manages subscriptions
2. Registered at `GET /ws/notifications` with JWT auth
3. Each user gets their own notification channel
4. Handlers broadcast events by user ID: `hub.SendNotification(userID, message)`

### AI Integration (Ollama)
- Local LLM integration in `internal/ai/` for thread summaries
- Ollama service runs in Docker container, model auto-pulls on first start
- Only works if Ollama service is healthy and model downloaded
- Handler: `POST /api/v1/ai/summarize/:thread_id` (Agent/Admin only)

## Environment Variables

Essential variables in `.env`:
```bash
# Server
PORT=8080
APP_ENV=development

# Database
DATABASE_URL=postgres://masaar:masaar@localhost:5432/masaar?sslmode=disable

# Redis (cache + sessions)
REDIS_URL=redis://localhost:6379

# JWT
JWT_SECRET=change-me-in-production
JWT_ACCESS_EXPIRY_MIN=15
JWT_REFRESH_EXPIRY_DAYS=7

# WhatsApp Business API (required for webhook)
WA_VERIFY_TOKEN=masaar-webhook-token
WA_API_VERSION=v19.0
WA_PHONE_NUMBER_ID=your-phone-id
WA_ACCESS_TOKEN=your-access-token

# Ollama LLM
OLLAMA_BASE_URL=http://ollama:11434
OLLAMA_MODEL=llama3
```

## Database Schema Overview

**Users:** Authentication, roles, language preference
**Contacts:** Unified contact profiles linked to WhatsApp numbers
**WhatsApp:** Threads and messages from Meta Cloud API
**Leads:** Sales pipeline stages (New → Won/Lost) with notes
**Deals:** Deal tracking with value and stage
**Invoices:** Generated PDFs with UAE 5% VAT
**Audit Logs:** Immutable activity log for compliance
**Notifications:** User-specific events

All migrations in `migrations/` are applied sequentially on startup.

## Code Organization Conventions

### Handler Pattern
```go
// Each handler is tied to a domain (Contact, Lead, User, etc)
type ContactHandler struct {
    repo *repo.ContactRepository
}

// Methods follow REST conventions: List, Get, Create, Update, Delete
// Use middleware.RequireRole() for access control
// Return JSON with standardized error format
```

### Repository Pattern
- All database access isolated in `internal/repo/`
- Repositories use pgx connection pool for prepared statements
- Query results mapped to domain types
- Repos handle pagination and filtering logic

### API Response Format
Success: `{ "data": {...}, "meta": {...} }`
Error: `{ "error": "message", "status": 400 }`

## Performance & Deployment

### Production Considerations
1. **JWT Secret:** Must be long random string (minimum 32 chars)
2. **Database:** Ensure PostgreSQL 16 with pgvector extension
3. **Redis:** Essential for sessions + cache; single point of failure requires HA setup
4. **Ollama:** Can be scaled separately; optional for AI features
5. **Docker:** Single binary deployment via `docker compose -f docker/docker-compose.prod.yml`
6. **Rate Limiting:** Webhook endpoint limited to 300 req/min; login limited to 10 req/min

### Monitoring
- Fiber logs all requests (middleware/logger)
- Audit log stored in database for compliance
- WebSocket disconnections handled gracefully; client auto-reconnect

## Important Files to Review

- `internal/api/router.go` — Route definitions and role enforcement
- `internal/api/handler/auth.go` — Authentication logic
- `internal/repo/` — Database layer; where to add new queries
- `cmd/server/main.go` — Server initialization, middleware setup
- `web/lib/` — Frontend API client configuration
- `migrations/` — Schema evolution history

## Testing

**Note:** No automated test suite currently in place. Test manually via:
- Swagger UI at `/docs` for API endpoints
- Frontend at `http://localhost:3000` for UI
- WhatsApp webhook via Meta's Webhook Test Tool or ngrok tunnel

## Common Gotchas

1. **Ollama requires model download** — First run pulls llama3 (~5GB); plan for it
2. **JWT token format** — Must be prefixed with `Bearer ` in Authorization header
3. **Role checks are per-handler** — Middleware applied individually; review `router.go`
4. **Redis connection required** — Logout and session validation depend on it; no graceful fallback
5. **pgvector extension** — PostgreSQL requires `CREATE EXTENSION vector` (auto-applied in migrations)
6. **RTL in Next.js** — Use Tailwind RTL plugin; CSS logical properties recommended
7. **WhatsApp webhook secret** — `WA_VERIFY_TOKEN` must match Meta's configured token exactly
