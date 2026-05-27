# Masaar CRM — System Overview

Reference document for Claude Code sessions. Covers the full codebase as of v0.3.0 (2026-05-27).

---

## Architecture at a Glance

```
Meta WhatsApp API ──► POST /webhooks/whatsapp ──► Go/Fiber backend ──► PostgreSQL
                                                        │
Browser ──────────────────────────────────────────► Next.js 14
                                                        │
                                                   WebSocket (/ws/notifications)
                                                        │
                                                    Redis (sessions + blacklist + rate limits)
                                                        │
                                                    Ollama (local LLM, PII-safe)
                                                    Gemini  (cloud LLM, non-PII only)
```

- Backend: Go 1.22 + Fiber v2 — single binary at `cmd/server/main.go`
- Frontend: Next.js 14 App Router + Tailwind CSS — `web/` directory
- Database: PostgreSQL 16 + pgvector extension
- Migrations: goose v3, files in `migrations/`, auto-run at startup
- Auth: JWT HS256 (15 min access / 7 day refresh) + Redis blacklist

---

## Startup Sequence (`cmd/server/main.go`)

1. Load config (`internal/config`)
2. Connect pgxpool → run `goose.Up()` (all pending migrations)
3. Connect Redis → ping
4. Instantiate all repos
5. Instantiate email service (SMTP or Azure)
6. Instantiate WhatsApp sender (optional — only if `WA_PHONE_NUMBER_ID` set)
7. Start WebSocket hub
8. Instantiate AI clients: `sensitiveAI` (Ollama, always), `cloudAI` (Gemini, optional)
9. Build `ScoringService`, `TaggingService`, `PaymentReminderService`, `PaymentConfirmationService`
10. Instantiate all handlers → `api.Handlers` struct
11. Register routes (`api.RegisterRoutes`)
12. Start background jobs (payment reminder generation every 12h, delivery every 1h)
13. Listen (graceful shutdown on SIGINT/SIGTERM)

---

## Role-Based Access Control

Three roles enforced per-handler via `middleware.RequireRole()`:

| Role    | Capabilities |
|---------|-------------|
| `admin` | Full access — delete, approve, admin routes |
| `agent` | Create + update on most resources, no delete |
| `viewer`| Read-only (no middleware on GET routes) |

Middleware chain for protected routes:
```
apiLimiter → JWT(secret) → CheckBlacklist(rdb) → ExtractClaims(companyID) → [RequireRole(...)]
```

---

## Rate Limiters

| Limiter | Max | Window | Applied to |
|---------|-----|--------|-----------|
| `loginLimiter` | 10 | 1 min | Login, forgot-password |
| `magicLinkLimiter` | 3 | 1 min | Magic link request |
| `smsOTPLimiter` | 3 | 1 min | SMS OTP request |
| `webhookLimiter` | 300 | 1 min | WhatsApp + lead webhooks |
| `apiLimiter` | 100 | 1 min | All `/api/v1/*` |
| `apiKeyLimiter` | 300 | 1 min | API key endpoints (Redis sliding window) |

---

## Complete Route Map

### Public (no auth)

```
POST /api/v1/auth/login                   loginLimiter
POST /api/v1/auth/magic-link/request      magicLinkLimiter
POST /api/v1/auth/magic-link/verify
POST /api/v1/auth/sms/request             smsOTPLimiter
POST /api/v1/auth/sms/verify
POST /api/v1/auth/refresh
POST /api/v1/auth/forgot-password         loginLimiter
POST /api/v1/auth/reset-password
GET  /webhooks/whatsapp                   Meta verify handshake
POST /webhooks/whatsapp                   webhookLimiter — Meta inbound messages
POST /webhooks/stripe                     Stripe billing webhook
GET  /api/public/sign/:id                 Public document signature view
POST /api/public/sign/:id                 Public document signing
POST /webhooks/leads                      webhookLimiter + API key auth + lead:create scope
GET  /health                              DB + Redis connectivity check
GET  /docs/*                              Swagger UI (BasicAuth in production)
```

### WebSocket (JWT required)

```
GET /ws/notifications                     Personal real-time notifications per user
```

### Authenticated `/api/v1/*` (all require JWT + non-blacklisted)

#### Auth
```
DELETE /auth/logout
```

#### Stats
```
GET /stats                                Dashboard KPIs (all roles)
```

#### Users (admin-only for management; self-service for /me)
```
GET    /users                             admin
POST   /users                             admin
POST   /users/invite                      admin
GET    /users/me                          any
PATCH  /users/me/password                 any
PATCH  /users/me/lang                     any
PATCH  /users/:id                         admin
PATCH  /users/:id/active                  admin
DELETE /users/:id                         admin
```

#### Settings
```
GET    /settings/bos24                    admin
PATCH  /settings/bos24                    admin
GET    /settings/company                  admin
PATCH  /settings/company                  admin
GET    /settings/api-keys                 admin
POST   /settings/api-keys                 admin
DELETE /settings/api-keys/:id             admin
GET    /settings/webhooks                 admin
POST   /settings/webhooks                 admin
DELETE /settings/webhooks/:id             admin
POST   /settings/webhooks/:id/test        admin
```

#### Billing
```
GET  /billing                             any
GET  /billing/usage                       any
POST /billing/checkout                    admin
POST /billing/portal                      admin
```

#### Contacts
```
GET    /contacts                          any
GET    /contacts/:id                      any
POST   /contacts                          agent+admin
PATCH  /contacts/:id                      agent+admin
DELETE /contacts/:id                      admin
```

#### Leads / Pipeline
```
GET    /leads                             any — Kanban board format
GET    /leads/search                      any — flat list
GET    /leads/:id                         any
POST   /leads                             agent+admin
PATCH  /leads/:id/stage                   agent+admin
PATCH  /leads/:id/notes                   agent+admin
PATCH  /leads/:id/assign                  agent+admin
DELETE /leads/:id                         admin
GET    /leads/:id/communications          any
GET    /leads/:id/tags                    any
POST   /leads/:id/tags                    agent+admin
DELETE /leads/:id/tags/:tag               agent+admin
```

#### WhatsApp Inbox
```
GET  /threads                             any
GET  /threads/:id                         any
GET  /threads/:id/messages                any
POST /threads/:id/close                   agent+admin
POST /threads/:id/reopen                  agent+admin
```

#### WhatsApp Outbound
```
POST /threads/:id/send-message            agent+admin
POST /threads/:id/send-template           agent+admin
GET  /threads/:id/outbound-messages       agent+admin
POST /threads/:id/send-media              agent+admin
```

#### AI (quota-checked)
```
POST /ai/summarize/:thread_id             agent+admin — Ollama, PII-safe
POST /ai/extract-buyer-profile/:thread_id agent+admin — Ollama, PII-safe
POST /ai/score-lead/:id                   agent+admin — Ollama, PII-safe
POST /ai/score-contact/:id                agent+admin — Ollama, persists to contacts.lead_score
POST /ai/draft-reply/:thread_id           agent+admin — Ollama, PII-safe
POST /ai/describe-listing                 agent+admin — Gemini (non-PII), falls back to Ollama
```

#### Message Analysis (AI quota-checked)
```
POST /messages/analyze                    agent+admin
POST /messages/suggest-action             agent+admin
POST /messages/auto-create-lead           agent+admin
```

#### Properties (BuyOrSell24, quota-checked)
```
POST /properties/search
POST /properties/report/pdf
POST /properties/projects/search
POST /properties/ai/describe
GET  /properties/transactions
GET  /properties/transactions/areas
GET  /properties/buildings
GET  /properties/buildings/:id
GET  /properties/areas
GET  /properties/areas/:slug/summary
GET  /properties/areas/:slug/buildings
GET  /properties/map/areas
GET  /properties/pois
GET  /properties/schools/nearby
GET  /properties/rentals
GET  /properties/rentals/ejari
GET  /properties/rentals/ejari/yield
GET  /properties/developers
GET  /properties/projects
GET  /properties/units
GET  /properties/valuations
GET  /properties/yield-analysis
GET  /properties/comparables
GET  /properties/market-trends
GET  /properties/insights/market-overview
GET  /properties/insights/area-comparison
GET  /properties/insights/price-trends
GET  /properties/insights/top-areas
GET  /properties/brokers
GET  /properties/transactions/by-project/:name
GET  /properties/transactions/area/:name/summary
GET  /properties/transactions/:id
GET  /properties/transactions/:id/enriched
GET  /properties/rentals/stats
GET  /properties/rentals/areas
GET  /properties/rentals/project/:name
GET  /properties/rentals/building/:name
GET  /properties/lands
GET  /properties/lands/:id
GET  /properties/map/config
GET  /properties/map/bounds
GET  /properties/map/poi-categories
GET  /properties/map/heatmap
GET  /properties/map/area/:name
GET  /properties/areas/:id
GET  /properties/developers/:id
GET  /properties/projects/:id
GET  /properties/units/:id
GET  /properties/valuations/:id
```

#### Notifications
```
GET    /notifications                     any
PATCH  /notifications/read-all            any
PATCH  /notifications/:id/read            any
```

#### Deals
```
GET    /deals                             any
GET    /deals/:id                         any
GET    /deals/:id/invoices                any
POST   /deals                             agent+admin
PATCH  /deals/:id                         agent+admin
PATCH  /deals/:id/stage                   agent+admin
DELETE /deals/:id                         admin
```

#### Invoices
```
GET    /invoices                          any
GET    /invoices/:id                      any
GET    /invoices/:id/pdf                  any
POST   /invoices                          agent+admin
POST   /invoices/:id/send                 admin
PATCH  /invoices/:id/status               admin
```

#### Email
```
GET  /emails                              agent+admin — all email history
POST /emails/send                         agent+admin
GET  /emails/history                      agent+admin — scoped history
```

#### Rental Properties
```
GET    /rental-properties                 any
GET    /rental-properties/:id             any
POST   /rental-properties                 agent+admin
PATCH  /rental-properties/:id             agent+admin
DELETE /rental-properties/:id             admin
```

#### Tenants
```
GET    /tenants                           any
GET    /tenants/:id                       any
POST   /tenants                           agent+admin
PATCH  /tenants/:id                       agent+admin
DELETE /tenants/:id                       admin
POST   /tenants/:id/verify                admin
```

#### Lease Templates
```
GET    /lease-templates                   any
GET    /lease-templates/:id               any
POST   /lease-templates                   admin
PATCH  /lease-templates/:id               admin
DELETE /lease-templates/:id               admin
```

#### Message Templates
```
GET    /message-templates                 any
GET    /message-templates/:id             any
POST   /message-templates                 admin
PATCH  /message-templates/:id             admin
DELETE /message-templates/:id             admin
```

#### Leases
```
GET    /leases                            any
GET    /leases/:id                        any
POST   /leases                            agent+admin
PATCH  /leases/:id                        agent+admin
DELETE /leases/:id                        admin
```

#### Payments
```
GET    /payments                          any
GET    /payments/:id                      any
POST   /payments                          agent+admin
PATCH  /payments/:id                      agent+admin
DELETE /payments/:id                      admin
GET    /payments/:payment_id/confirmation agent+admin
POST   /payments/:payment_id/send-confirmation agent+admin
```

#### Bank Integrations
```
GET    /bank-integrations                 any (but content scoped to company)
GET    /bank-integrations/:id             any
POST   /bank-integrations                 admin
PATCH  /bank-integrations/:id             admin
DELETE /bank-integrations/:id             admin
```

#### Bank Statements
```
GET    /bank-statements                   agent+admin
GET    /bank-statements/:id               agent+admin
POST   /bank-statements/upload            agent+admin
DELETE /bank-statements/:id               admin
```

#### Analytics
```
GET /analytics/tenant-overview            any
GET /analytics/properties                 any
GET /analytics/properties/:propertyID     any
GET /analytics/tenants                    any
GET /analytics/tenants/:tenantID          any
GET /analytics/financial                  any
GET /analytics/maintenance                any
```

#### Expenses
```
GET    /expense-categories                any
POST   /expense-categories                admin
GET    /expenses                          any
GET    /expenses/:id                      any
POST   /expenses                          agent+admin
PATCH  /expenses/:id                      agent+admin
DELETE /expenses/:id                      admin
POST   /expenses/:id/approve              admin
```

#### Inspections
```
GET    /inspection-templates              any
POST   /inspection-templates              admin
GET    /inspections                       any
GET    /inspections/:id                   any
POST   /inspections                       agent+admin
PATCH  /inspections/:id                   agent+admin
POST   /inspections/:id/complete          agent+admin
```

#### Maintenance Tasks
```
GET    /maintenance-tasks                 any
GET    /maintenance-tasks/:id             any
POST   /maintenance-tasks                 agent+admin
PATCH  /maintenance-tasks/:id             agent+admin
POST   /maintenance-tasks/:id/complete    agent+admin
POST   /maintenance-tasks/:id/photos      agent+admin
GET    /maintenance-tasks/:id/photos      any
DELETE /maintenance-tasks/:id             admin
```

#### Lease Renewals
```
GET  /lease-renewals                      any
GET  /lease-renewals/:id                  any
POST /lease-renewals/:lease_id/initiate   admin
PUT  /lease-renewals/:id/propose          agent+admin
POST /lease-renewals/:id/send-offer       agent+admin
PUT  /lease-renewals/:id/accept           agent+admin
PUT  /lease-renewals/:id/reject           agent+admin
POST /lease-renewals/:id/counter-offer    agent+admin
```

#### Renewal Templates
```
GET    /renewal-templates                 any
POST   /renewal-templates                 admin
PATCH  /renewal-templates/:id             admin
DELETE /renewal-templates/:id             admin
```

#### Document Templates
```
GET    /documents/templates               any
GET    /documents/templates/:id           any
POST   /documents/templates               agent+admin
PATCH  /documents/templates/:id           agent+admin
DELETE /documents/templates/:id           admin
```

#### Documents
```
GET    /documents                         any
GET    /documents/:id                     any
POST   /documents                         agent+admin
POST   /documents/:id/request-signature   agent+admin
PATCH  /documents/signatures/:id/mark-signed agent+admin
DELETE /documents/:id                     agent+admin
```

#### Audit Log
```
GET /audit-logs                           admin
```

---

## Handler → Repo → Route Mapping

| Handler file | Repo(s) | Key routes |
|---|---|---|
| `auth.go` | `UserRepo`, `AuditLogRepo` | `/auth/*` |
| `user.go` | `UserRepo`, `AuditLogRepo` | `/users/*` |
| `contact.go` | `ContactRepo`, `AuditLogRepo` | `/contacts/*` |
| `lead.go` | `LeadRepo`, `ContactRepo`, `CommHistRepo`, `LeadTagRepo` | `/leads/*` |
| `whatsapp.go` | `WhatsAppRepo`, `ContactRepo` | `/threads/*` (read) |
| `whatsapp_outbound.go` | `WhatsAppOutboundRepo`, `WhatsAppRepo` | `/threads/:id/send-*` |
| `ai.go` | `ContactRepo`, `LeadRepo`, `WhatsAppRepo` | `/ai/*` |
| `message.go` | `WhatsAppRepo`, `LeadRepo`, `ContactRepo`, `CommHistRepo`, `LeadTagRepo` | `/messages/*` |
| `deal.go` | `DealRepo`, `InvoiceRepo`, `AuditLogRepo` | `/deals/*` |
| `invoice.go` | `InvoiceRepo`, `DealRepo`, `CompanySettingsRepo` | `/invoices/*` |
| `email.go` | `EmailRepository` | `/emails/*` |
| `notification.go` | `NotificationRepo` | `/notifications/*` |
| `stats.go` | `StatsRepo` | `/stats` |
| `property.go` | BOS24 client, `CompanySettingsRepo` | `/properties/*` |
| `settings.go` | `SettingsRepo`, `CompanySettingsRepo` | `/settings/*` |
| `rental_property.go` | `RentalPropertyRepo` | `/rental-properties/*` |
| `tenant.go` | `TenantRepo` | `/tenants/*` |
| `lease_template.go` | `LeaseTemplateRepo` | `/lease-templates/*` |
| `lease.go` | `LeaseRepo` | `/leases/*` |
| `payment.go` | `PaymentRepo` | `/payments/*` |
| `bank_integration.go` | `BankIntegrationRepo` | `/bank-integrations/*` |
| `bank_statement.go` | `BankStatementRepo` | `/bank-statements/*` |
| `payment_confirmation.go` | `PaymentConfirmationRepo` | `/payments/:id/confirmation*` |
| `analytics.go` | `AnalyticsRepository` | `/analytics/*` |
| `expense.go` | `ExpenseRepository` | `/expenses/*`, `/expense-categories/*` |
| `inspection.go` | `InspectionTemplateRepo`, `InspectionRepo` | `/inspections/*`, `/inspection-templates/*` |
| `maintenance.go` | `MaintenanceTaskRepo` | `/maintenance-tasks/*` |
| `lease_renewal.go` | `LeaseRenewalRepo`, `RenewalTemplateRepo`, `RenewalCommunicationLogRepo` | `/lease-renewals/*`, `/renewal-templates/*` |
| `document.go` | `DocumentRepo`, `AuditLogRepo` | `/documents/*`, public sign |
| `api_key.go` | `ApiKeyRepo` | `/settings/api-keys/*` |
| `webhook_sub.go` | `WebhookRepo` | `/settings/webhooks/*` |
| `billing.go` | `BillingRepo`, `CompanySettingsRepo` | `/billing/*`, `/webhooks/stripe` |
| `message_template.go` | `MessageTemplateRepo` | `/message-templates/*` |
| `public_lead.go` | `ContactRepo`, `LeadRepo` | `/webhooks/leads` |
| `audit.go` | `AuditLogRepo` | `/audit-logs` |

---

## Database Tables (44 migrations)

| Migration | Tables created |
|-----------|----------------|
| 0001 | `companies` |
| 0002 | `users` |
| 0003 | `contacts` |
| 0004 | `whatsapp_threads`, `whatsapp_messages` |
| 0005 | `leads`, `lead_notes` |
| 0006 | `deals` |
| 0007 | `vat_invoices`, `invoice_line_items` |
| 0008 | `audit_logs` |
| 0009 | `notifications` |
| 0010 | `api_settings` (BOS24 token) |
| 0011 | `email_history` |
| 0012 | `company_settings` (VAT/bank/invoice config) |
| 0013 | `whatsapp_outbound` |
| 0014 | `lead_tags` |
| 0015 | `communication_history` |
| 0016 | `rental_properties` |
| 0017 | `tenants` |
| 0018 | `lease_templates` |
| 0019 | `leases` |
| 0020 | `bank_integrations` |
| 0021 | `bank_transactions` |
| 0022 | `payments` |
| 0023 | cross-reference columns (bank ↔ payment) |
| 0024 | `users.role` default change |
| 0025 | `bank_statements`, `payment_confirmations` |
| 0026 | `inspection_templates`, `inspections`, `maintenance_tasks`, `maintenance_photos` |
| 0027 | `lease_renewal_workflows`, `renewal_communication_templates`, `renewal_communication_log` |
| 0028 | `commission_structures`, `agent_commissions`, `commission_transactions` |
| 0029 | `document_templates`, `documents`, `document_signatures`, `document_audit_log` |
| 0030 | `custom_fields` |
| 0031 | `bulk_operations` |
| 0032 | `expense_categories`, `expenses`, `expense_approvals` |
| 0033 | `api_keys` |
| 0034 | `webhook_subscriptions` |
| 0036 | leads improvements (additional columns) |
| 0037 | `password_reset_tokens` |
| 0038 | `billing_records`, `subscription_plans` |
| 0039 | `users.is_active` column |
| 0040 | `users.wa_number` NOT NULL change |
| 0042 | production indexes |
| 0043 | `users.phone` column |
| 0044 | `message_templates` |

Note: migration 0035 is missing/skipped — this is intentional or was removed.

---

## Frontend Pages

All pages live under `web/app/(dashboard)/` unless noted.

### CRM Section
| Page | Route | Notes |
|------|-------|-------|
| Dashboard | `/dashboard` | KPI overview |
| Pipeline | `/pipeline` | Kanban drag-drop (dnd-kit); AI scoring button |
| Inbox | `/inbox`, `/inbox/[id]` | WhatsApp threads + messages |
| Msg Templates | `/message-templates` | CRUD; WhatsApp bubble preview |
| Contacts | `/contacts`, `/contacts/[id]` | Detail has AI score button |
| Deals | `/deals`, `/deals/[id]` | Stage tracking, linked invoices |
| Invoices | `/invoices`, `/invoices/[id]` | Summary cards, PDF download, status actions |
| Email History | `/email-history` | Status filters, click-to-detail |

### Properties Section
| Page | Route | Notes |
|------|-------|-------|
| Search | `/properties` | BOS24 market data |
| Rentals | `/rentals`, `/rentals/[id]` | Rental property management |
| Tenants | `/tenants`, `/tenants/[id]` | Tenant profiles |
| Leases | `/leases`, `/leases/[id]` | Lease lifecycle |
| Lease Templates | `/lease-templates` | Admin-only CRUD |
| Documents | `/documents`, `/documents/[id]`, `/documents/templates`, `/documents/templates/[id]` | e-signature framework |
| Renewal Templates | `/renewal-templates` | Admin-only card-grid CRUD |
| Renewals | `/renewals`, `/renewals/[id]` | Workflow tracking |
| Expenses | `/expenses`, `/expenses/[id]` | Approval workflow |
| Payments | `/payments`, `/payments/[id]` | Payment tracking |
| Bank Statements | `/bank-statements` | Upload + reconcile |
| Inspections | `/inspections`, `/inspections/[id]` | Checklist-based |
| Maintenance | `/maintenance`, `/maintenance/[id]` | Task + photo tracking |

### Insights Section
| Page | Route |
|------|-------|
| Analytics overview | `/analytics` |
| Financial | `/analytics/financial` |
| Properties | `/analytics/properties`, `/analytics/properties/[id]` |
| Tenants | `/analytics/tenants` |

### Account Section
| Page | Route | Notes |
|------|-------|-------|
| Team | `/admin/users` | Admin only |
| Plans & Billing | `/settings/billing` | Stripe integration |
| Notifications | `/notifications` | WebSocket-driven |
| Audit Log | `/audit-log` | Admin only |
| Settings | `/settings` | API integrations |
| Company | `/settings/company` | VAT, bank details |
| API Keys | `/settings/api-keys` | Admin only |
| Webhooks | `/settings/webhooks` | Admin only |

### Auth Pages (outside dashboard)
- `/login`, `/login/magic-link`, `/login/magic-link/verify`
- `/forgot-password`, `/reset-password`
- `/sign/[id]` — public document signing (no auth)
- `/developers` — developer docs / API key info

---

## AI Architecture

Two clients with intentional routing:

```
sensitiveAI = Ollama (local)     — ALL customer PII: messages, contacts, leads, scoring
cloudAI     = Gemini (optional)  — Non-PII only: property listing descriptions, market text
```

If `GEMINI_API_KEY` is not set, `cloudAI` falls back to `sensitiveAI`. This is enforced by `cloudOrLocal()` in `main.go`.

**Why this matters:** UAE PDPL requires customer data to stay on-server. Never route PII through Gemini.

### AI endpoints and which client they use

| Endpoint | Client | Notes |
|----------|--------|-------|
| `/ai/summarize/:thread_id` | sensitiveAI | WhatsApp thread summary |
| `/ai/extract-buyer-profile/:thread_id` | sensitiveAI | Buyer intent extraction |
| `/ai/score-lead/:id` | sensitiveAI | Returns `{"score": int, "reasoning": "..."}` |
| `/ai/score-contact/:id` | sensitiveAI | Same as above, persists to `contacts.lead_score` |
| `/ai/draft-reply/:thread_id` | sensitiveAI | Message drafting |
| `/ai/describe-listing` | cloudOrLocal | Non-PII property copy |
| `/messages/analyze` | sensitiveAI (via MessageHandler) | Intent parsing |
| `/messages/suggest-action` | sensitiveAI | Next action suggestion |
| `/messages/auto-create-lead` | sensitiveAI | Auto-lead from message |

---

## AI Quota Middleware

Two quota layers on AI endpoints:
1. `CheckQuota(billingRepo, rdb, "ai")` — company-level quota from billing plan
2. `CheckUserAIQuota(billingRepo, rdb)` — per-user rate limit

Both use Redis for tracking.

---

## Frontend Patterns

### State management
- `useAuthStore()` — Zustand store, provides `user` (id, role, name, etc.)
- `useLang()` — Language context, returns `{ lang, t }` where `t('AR text', 'EN text')` picks by lang

### Role checks in UI
```tsx
const { user } = useAuthStore()
const isAdmin = user?.role === 'admin'
const isAgent = user?.role === 'admin' || user?.role === 'agent'
```

### API client
All API calls via `web/lib/api.ts`. Namespaces mirror backend resources:
- `api.auth.*` — login, logout, refresh
- `api.contacts.*` — CRUD + score
- `api.leads.*` — Kanban, CRUD, tags, communications
- `api.whatsapp.*` — threads, messages, outbound
- `api.deals.*` — CRUD, stage, invoices
- `api.invoices.*` — list, get, create, send, updateStatus, pdf
- `api.email.*` — send, history, list
- `api.ai.*` — summarize, scoreLead, scoreContact, draftReply, describeListing
- `api.rentalProperties.*` — CRUD
- `api.tenants.*` — CRUD, verify
- `api.leases.*` — CRUD
- `api.leaseTemplates.*` — CRUD
- `api.payments.*` — CRUD, confirmation
- `api.bankIntegrations.*` — CRUD
- `api.bankStatements.*` — list, upload, delete
- `api.analytics.*` — all analytics endpoints
- `api.expenses.*` — categories + expense CRUD + approve
- `api.inspections.*` — templates + CRUD
- `api.maintenance.*` — CRUD + photos
- `api.leaseRenewals.*` — CRUD + workflow + templates
- `api.documents.*` — templates + CRUD + signatures
- `api.messageTemplates.*` — CRUD
- `api.notifications.*` — list, markRead, markAllRead
- `api.settings.*` — bos24, company
- `api.apiKeys.*` — CRUD
- `api.webhooks.*` — CRUD + test
- `api.billing.*` — info, usage, checkout, portal
- `api.stats.*` — overview
- `api.properties.*` — BOS24 market data

---

## Critical Patterns & Gotchas

### 1. `uuid.Parse` returns two values — ALWAYS capture both

```go
// WRONG — compile error:
companyID := uuid.Parse(c.Locals("company_id").(string))

// CORRECT:
companyID, _ := uuid.Parse(c.Locals("company_id").(string))

// CORRECT for struct fields:
task.CompanyID, _ = uuid.Parse(someString)
```

This has caused compile failures across 5 handler files historically. Any new handler that parses UUIDs from strings must use the two-value form.

### 2. Company scoping in handlers

Every handler method that returns data must scope by `company_id` from JWT claims:
```go
companyID, _ := uuid.Parse(c.Locals("company_id").(string))
```
Never query across companies without this.

### 3. Route ordering — specific before parameterized

When registering routes, specific paths MUST come before `:id` params:
```go
// CORRECT order:
v1.Get("/invoices", h.Invoice.List)          // specific first
v1.Get("/invoices/:id", h.Invoice.Get)       // then param

// WRONG — List would never match:
v1.Get("/invoices/:id", h.Invoice.Get)
v1.Get("/invoices", h.Invoice.List)
```
Same applies to nested: `/emails` before `/emails/history` before `/emails/:id`.

### 4. AI: PII never to Gemini

`sensitiveAI` (Ollama) for anything involving customer data.
`cloudOrLocal()` (Gemini or Ollama fallback) for non-PII only.

### 5. Standard list endpoint recipe

When adding a new list endpoint:
1. `repo/X.go` — add `ListAll(ctx, page, limit int) ([]domain.X, int, error)`
2. `handler/X.go` — add handler using `strconv.Atoi` for page/limit
3. `router.go` — register BEFORE `/:id` route
4. `web/lib/api.ts` — add to namespace
5. `web/types/index.ts` — add TypeScript type if needed

### 6. Handler pattern for new resources

```go
type XHandler struct {
    xRepo *repo.XRepo
}

func NewXHandler(xRepo *repo.XRepo) *XHandler {
    return &XHandler{xRepo: xRepo}
}
```

Register in `main.go` → add to `api.Handlers` struct in `router.go` → register routes.

### 7. WhatsApp outbound records

`whatsapp_outbound` rows are ALWAYS inserted before the API call. If `WA_PHONE_NUMBER_ID`/`WA_ACCESS_TOKEN` not set, handler returns 503. Check `whatsapp_outbound.status` column for delivery: `pending` → `sent` (with `wa_message_id`) or `failed` (with `error_message`).

### 8. Stripe webhook requires raw body

`POST /webhooks/stripe` must receive raw body for HMAC verification — do not use body parsers on this route.

### 9. Swagger docs

Run `swag init -g cmd/server/main.go` after adding new handlers with `// @Summary` comments.
In production, Swagger UI at `/docs/*` is gated by BasicAuth (username: `admin`, password: first 16 chars of `JWT_SECRET`).

### 10. Background jobs

Two goroutines run on startup:
- Every 12h: generate payment reminders for all companies
- Every 1h: send pending payment reminders

These use `companyRepo.List()` — adding a company automatically enrolls it.

---

## Services (internal packages beyond handlers/repos)

| Package | Purpose |
|---------|---------|
| `internal/ai` | Ollama + Gemini clients; `ScoringService`, `TaggingService`, `PaymentReminderService`, `PaymentConfirmationService` |
| `internal/email` | SMTP + Azure Communication Services abstraction |
| `internal/whatsapp` | Meta Cloud API outbound sender (text/template/media) |
| `internal/ws` | WebSocket hub — per-user notification channels |
| `internal/pdf` | Invoice PDF generation |
| `internal/webhook` | Outbound webhook dispatcher (fires on lead events) |
| `internal/billing` | Stripe integration, quota enforcement |
| `internal/sms` | SMSCountry client (OTP login) |
| `internal/bos24` | BuyOrSell24 real estate data client |
| `internal/config` | Env var loading |

---

## Environment Variables Reference

```bash
PORT=8080
APP_ENV=development               # "production" enables stricter validation

DATABASE_URL=postgres://...
REDIS_URL=redis://localhost:6379

JWT_SECRET=                       # min 32 chars in production
JWT_ACCESS_EXPIRY_MIN=15
JWT_REFRESH_EXPIRY_DAYS=7

# WhatsApp (all required for outbound)
WA_VERIFY_TOKEN=                  # Meta webhook verification token
WA_API_VERSION=v19.0
WA_PHONE_NUMBER_ID=
WA_ACCESS_TOKEN=
WA_APP_SECRET=                    # HMAC validation — required in production if WA enabled
WA_BASE_URL=https://graph.facebook.com  # NOT graph.instagram.com

# AI
OLLAMA_BASE_URL=http://ollama:11434
OLLAMA_MODEL=llama3
GEMINI_API_KEY=                   # optional — non-PII tasks only
GEMINI_MODEL=gemini-1.5-flash

# Email
SMTP_HOST=
SMTP_PORT=587
SMTP_USER=
SMTP_PASSWORD=
SMTP_FROM_EMAIL=
SMTP_FROM_NAME=
AZURE_COMM_ENDPOINT=              # Azure Communication Services (alternative to SMTP)
AZURE_COMM_KEY=
AZURE_COMM_FROM_ADDRESS=

# SMS (optional — OTP login)
SMS_COUNTRY_AUTH_KEY=
SMS_COUNTRY_AUTH_TOKEN=
SMS_COUNTRY_SENDER_ID=

# BuyOrSell24 (optional — real estate market data)
BOS24_TOKEN=
BOS24_BASE_URL=

# Stripe (optional — billing)
STRIPE_SECRET_KEY=
STRIPE_WEBHOOK_SECRET=
STRIPE_PRICE_ID_STARTER=
STRIPE_PRICE_ID_PRO=
STRIPE_PRICE_ID_BUSINESS=
APP_URL=                          # your public domain (for Stripe redirects)

ALLOWED_ORIGINS=http://localhost:3000  # must not be "*" in production
COMPANY_ID=                       # single-tenant company UUID
```

---

## Known Gaps & Resolved Bugs

| # | Issue | Status | Fix |
|---|-------|--------|-----|
| 1 | `uuid.Parse` single-value in 5 handler files — backend won't compile | **Fixed v0.3.0** | `id, _ := uuid.Parse(...)` in document.go, expense.go, inspection.go, lease_renewal.go, maintenance.go |
| 2 | `ai.go` missing `"log"` import | **Fixed v0.3.0** | Added to imports |
| 3 | `GET /api/v1/invoices` missing | **Fixed v0.3.0** | Added `Invoice.List` handler + `InvoiceRepo.ListAll` |
| 4 | `GET /api/v1/emails` missing (list all) | **Fixed v0.3.0** | Added `Email.ListAllEmailHistory` + `EmailRepo.ListAll` |
| 5 | AI scoring didn't persist on contact detail | **Fixed v0.3.0** | `POST /ai/score-contact/:id` + `ContactRepo.UpdateScore` |
| 6 | `RenewalTemplateRepo` missing Update + Delete | **Fixed v0.3.0** | Added both methods + routes |
| 7 | Message Templates page missing (backend was done) | **Fixed v0.3.0** | Added `/message-templates` frontend page |
| 8 | Commission calculation engine | **Pending** | Schema ready (migration 0028), no handler |
| 9 | DocuSign e-signature | **Pending** | Schema ready (migration 0029), no integration |
| 10 | Bulk import/export job processing | **Pending** | Schema ready (migration 0031), no handler |
| 11 | Analytics Redis caching | **Pending** | Hits DB on every call |

---

## Multi-Tenant Note

Currently single-tenant: `COMPANY_ID` env var is injected as the company UUID for all requests via `ExtractClaims(cfg.CompanyID)`. All data is scoped by `company_id` in every query. The schema is multi-tenant-ready but the platform layer (Phase 17) is not yet built.
