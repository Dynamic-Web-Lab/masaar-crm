# Masaar CRM — Multi-Tenant SaaS & UAE Real Estate Go-to-Market Plan

> **Target market**: UAE real estate agencies (1–50 agents), Arabic-first, WhatsApp-native
> **Goal**: Convert single-tenant open-source CRM into a sellable multi-tenant SaaS

---

## Current State Assessment

**What's already built and strong:**
- WhatsApp inbox (inbound + outbound via Meta Cloud API)
- Lead/contact pipeline with Kanban stages
- Lease management with Ejari number field
- Tenant management with ID verification
- Rental property management
- Payment tracking, invoices with UAE 5% VAT
- Maintenance tasks and inspections
- Document management with digital signatures
- Commission tracking (fixed, percentage, tiered)
- Bank statement import and reconciliation
- Expense management
- Analytics (tenant, property, financial, maintenance)
- AI lead scoring and thread summaries (Ollama)
- BOS24 UAE property data integration (transactions, buildings, areas)
- Arabic/RTL native UI
- Role-based access control (Admin, Agent, Viewer)
- Audit logging (PDPL-ready)
- PDF invoice generation

**What's blocking revenue today:**
1. No multi-tenancy — one company only, cannot sell to multiple customers
2. No PDC (Post-Dated Cheque) management — the #1 UAE rental workflow
3. No background job runner — reminders and trial expiry never fire
4. Stripe keys not configured — billing UI exists but can't collect payment
5. No team invitation flow — admins can't add their agents
6. No signup/onboarding flow — new customers have nowhere to start

---

## Competitive Landscape

| Feature | Propspace | Property Monitor | Zoho CRM | **Masaar** |
|---------|-----------|-----------------|----------|------------|
| PDC management | ✅ | ✅ | ❌ | ❌ Phase 1 |
| Mobile app | ✅ | ✅ | ✅ | ❌ Phase 3 |
| Portal listing sync | ✅ | ❌ | ❌ | ❌ Phase 3 |
| Off-plan pipeline | ✅ | ✅ | ❌ | ❌ Phase 2 |
| WhatsApp CRM native | ❌ | ❌ | ❌ | ✅ |
| Arabic native | partial | ❌ | partial | ✅ |
| AI lead scoring | ❌ | ❌ | ❌ | ✅ |
| Bank reconciliation | ❌ | partial | ❌ | ✅ |
| Ejari integration | ❌ | ✅ | ❌ | partial |
| Open source / self-host | ❌ | ❌ | ❌ | ✅ |

**Pricing comparison:** Propspace ~$80–200/user/month. Property Monitor ~$150+/month. Our target: $29–199/company/month (not per-user), which is a strong angle for small agencies.

---

## Pricing Plans (current `internal/billing/plans.go`)

| Plan | Price | BOS24 calls | AI requests | PDF exports |
|------|-------|-------------|-------------|-------------|
| Community | Free | 0 | 0 | 0 |
| Starter | $29/mo | 200 | 100 | 10 |
| Pro | $79/mo | 1,000 | 500 | Unlimited |
| Business | $199/mo | Unlimited | Unlimited | Unlimited |

> Open question: Confirm pricing. $29/$79/$199 USD is aggressive vs. competitors — consider AED pricing on the marketing site (AED 109 / AED 299 / AED 749) to feel local.

---

## Phase 0 — Fix Critical Bugs Before Anything Else (1 week)

These prevent the backend from working correctly today. Must be done before any phase.

| # | Issue | File | Fix |
|---|-------|------|-----|
| 0.1 | `ExtractClaims` hardcodes `company_id` from `APP_COMPANY_ID` env var | `middleware/auth.go:67` | Read from JWT claims after Phase 1 |
| 0.2 | `billing.GetPlan()` uses `LIMIT 1`, not filtered by `company_id` | `repo/billing.go:30` | Add `company_id` param in Phase 1 |
| 0.3 | `quota.go` calls `GetPlan()` with no `companyID` arg | `middleware/quota.go:18` | Update after billing repo change |
| 0.4 | No background job runner — `PaymentReminder` rows exist but nothing sends them | none | Add in Phase 1 |
| 0.5 | Magic link / SMS OTP flows create users without `company_id` | `handler/auth.go` | Guard in Phase 1 |

---

## Phase 1 — Multi-Tenancy + Revenue Infrastructure (3–4 weeks)

**Goal:** Multiple companies can sign up, pay, and use the product in isolation.

### 1.1 Database Migration (`migrations/0045_saas_multi_tenancy.sql`)

```sql
-- Add SaaS fields to companies
ALTER TABLE companies
  ADD COLUMN subdomain          VARCHAR(100) UNIQUE,
  ADD COLUMN plan               VARCHAR(20)  NOT NULL DEFAULT 'starter',
  ADD COLUMN trial_started_at   TIMESTAMPTZ,
  ADD COLUMN trial_ends_at      TIMESTAMPTZ,
  ADD COLUMN on_trial           BOOLEAN      NOT NULL DEFAULT TRUE,
  ADD COLUMN is_active          BOOLEAN      NOT NULL DEFAULT TRUE,
  ADD COLUMN stripe_customer_id VARCHAR(100),
  ADD COLUMN stripe_sub_id      VARCHAR(100);

-- Bind users to companies
ALTER TABLE users ADD COLUMN company_id UUID REFERENCES companies(id);
CREATE INDEX idx_users_company_id ON users(company_id);

-- Backfill existing data to the seed company
UPDATE companies SET
  plan = COALESCE((SELECT plan FROM company_settings LIMIT 1), 'community'),
  is_active = TRUE,
  on_trial = FALSE
WHERE id = '00000000-0000-0000-0000-000000000001';

UPDATE users SET company_id = '00000000-0000-0000-0000-000000000001'
WHERE company_id IS NULL;

-- Now enforce NOT NULL (safe after backfill)
ALTER TABLE users ALTER COLUMN company_id SET NOT NULL;

-- Create company_settings row for existing company (if not exists)
INSERT INTO company_settings (company_id)
VALUES ('00000000-0000-0000-0000-000000000001')
ON CONFLICT DO NOTHING;

-- Remove billing fields from company_settings (now live on companies table)
ALTER TABLE company_settings
  DROP COLUMN IF EXISTS plan,
  DROP COLUMN IF EXISTS stripe_customer_id,
  DROP COLUMN IF EXISTS stripe_sub_id,
  DROP COLUMN IF EXISTS plan_started_at,
  DROP COLUMN IF EXISTS plan_expires_at;
```

### 1.2 Domain Models (`internal/domain/models.go`)

Add to `Company` struct:
```go
type Company struct {
    ID               uuid.UUID  `json:"id"`
    Name             string     `json:"name"`
    Subdomain        string     `json:"subdomain"`
    Plan             string     `json:"plan"`
    TrialStartedAt   *time.Time `json:"trial_started_at"`
    TrialEndsAt      *time.Time `json:"trial_ends_at"`
    OnTrial          bool       `json:"on_trial"`
    IsActive         bool       `json:"is_active"`
    StripeCustomerID string     `json:"stripe_customer_id,omitempty"`
    StripeSubID      string     `json:"stripe_sub_id,omitempty"`
    CreatedAt        time.Time  `json:"created_at"`
}
```

Add `CompanyID uuid.UUID` to `User` struct.

### 1.3 Company Repository (`internal/repo/company.go`)

Expand the current minimal repo (which only has `List()`):

| Method | Description |
|--------|-------------|
| `Create(ctx, name, subdomain string) (*Company, error)` | Creates company + `company_settings` row; sets 90-day trial on `starter` |
| `GetByID(ctx, id uuid.UUID) (*Company, error)` | Full company with plan/trial info |
| `GetBySubdomain(ctx, subdomain string) (*Company, error)` | Used at login to resolve tenant |
| `SetPlan(ctx, id uuid.UUID, plan, stripeCustomerID, stripeSubID string) error` | After Stripe webhook |
| `EndTrial(ctx, id uuid.UUID) error` | Downgrades to `community`, sets `on_trial=false` |
| `Deactivate(ctx, id uuid.UUID) error` | Suspends on non-payment |
| `ListExpiredTrials(ctx) ([]Company, error)` | Used by background job |

### 1.4 User Repository Updates (`internal/repo/user.go`)

| Method | Change |
|--------|--------|
| `CreateWithCompany(ctx, name, email, hash string, companyID uuid.UUID, role Role) (*User, error)` | New — for registration |
| `ListByCompany(ctx, companyID uuid.UUID, page, limit int) ([]User, int, error)` | New — team management |
| `InviteUser(ctx, email, name string, companyID uuid.UUID, role Role) (*User, error)` | New — invite flow |

### 1.5 Billing Repository Updates (`internal/repo/billing.go`)

- `GetPlan(ctx, companyID uuid.UUID)` — filter by `company_id`, not `LIMIT 1`
- `SetPlan(ctx, companyID uuid.UUID, plan, customerID, subID string) error` — write to `companies` table
- `IncrUsage` / `GetUsage` — already keyed by `company_id`, no changes

### 1.6 JWT & Middleware

**JWT claims** — Add to access token in `handler/auth.go:generateTokenPair`:
```go
"company_id": user.CompanyID.String(),
"plan":       company.Plan,
"on_trial":   company.OnTrial,
```

**`middleware/auth.go` — `ExtractClaims` fix:**
Read `company_id` from JWT claims, not from `cfg.CompanyID`:
```go
func ExtractClaims() fiber.Handler {
    return func(c *fiber.Ctx) error {
        claims := ClaimsFromCtx(c)
        sub, _ := claims["sub"].(string)
        userID, _ := uuid.Parse(sub)
        companyID, _ := uuid.Parse(claims["company_id"].(string))
        c.Locals("user_id", userID)
        c.Locals("company_id", companyID.String()) // keep as string for compat
        if roleStr, ok := claims["role"].(string); ok {
            c.Locals("role", domain.Role(roleStr))
        }
        return c.Next()
    }
}
```

**NEW: `middleware/trial.go`** — reads `company_id` directly from JWT claims (NOT from Locals — this avoids the circular dependency with `ExtractClaims`):
```go
func TrialCheck(companyRepo *repo.CompanyRepo) fiber.Handler {
    return func(c *fiber.Ctx) error {
        claims := middleware.ClaimsFromCtx(c)
        companyID, _ := uuid.Parse(claims["company_id"].(string))
        company, err := companyRepo.GetByID(c.Context(), companyID)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "company not found"})
        }
        if !company.IsActive {
            return c.Status(403).JSON(fiber.Map{"error": "account_suspended"})
        }
        if company.OnTrial && company.TrialEndsAt != nil && company.TrialEndsAt.Before(time.Now()) {
            companyRepo.EndTrial(c.Context(), companyID)
            company.Plan = "community"
        }
        c.Locals("plan", company.Plan)
        c.Locals("on_trial", company.OnTrial)
        return c.Next()
    }
}
```

**Middleware order** in `router.go` (important — `TrialCheck` reads from JWT directly, so order is safe):
1. `middleware.JWT(cfg.JWTSecret)`
2. `middleware.CheckBlacklist(rdb)`
3. `middleware.TrialCheck(companyRepo)` ← reads JWT claims directly
4. `middleware.ExtractClaims()` ← reads JWT claims + sets Locals

### 1.7 Registration Endpoint (`POST /api/v1/auth/register`)

Rate-limited: 3 req/min/IP. Public route.

**Request:**
```json
{
  "name": "Ahmed Al Maktoum",
  "email": "ahmed@example.com",
  "password": "SecurePass1!",
  "company_name": "Example Properties LLC",
  "subdomain": "example",
  "turnstile_token": "..."
}
```

**Flow:**
1. Check `ALLOW_REGISTRATION=true` (return 503 if false — for self-hosters who disable public signup)
2. Validate Turnstile token via Cloudflare siteverify API (`https://challenges.cloudflare.com/turnstile/v0/siteverify`)
3. Validate input: name, email format, password strength (min 8 chars, 1 uppercase, 1 number)
4. Check email uniqueness (`userRepo.FindByEmail`)
5. Check subdomain uniqueness (`companyRepo.GetBySubdomain`)
6. Hash password (bcrypt cost 12)
7. `companyRepo.Create(ctx, companyName, subdomain)` — creates company + `company_settings` row
8. `userRepo.CreateWithCompany(ctx, name, email, hash, companyID, RoleAdmin)`
9. Generate JWT pair (with `company_id`, `plan`, `on_trial` in claims)
10. Send welcome email async (email service is built — SMTP / Azure Communication Services)
11. Log audit event

**Response (201):**
```json
{
  "access_token": "...",
  "refresh_token": "...",
  "expires_in": 900,
  "user": { "id": "...", "name": "...", "email": "...", "role": "admin" },
  "company": {
    "id": "...", "name": "...", "subdomain": "example",
    "plan": "starter", "on_trial": true,
    "trial_ends_at": "2026-08-26T00:00:00Z", "days_remaining": 90
  }
}
```

### 1.8 Login Update (`POST /api/v1/auth/login`)

After password verify:
1. Load company: `companyRepo.GetByID(ctx, user.CompanyID)`
2. If `!company.IsActive` → 403 "account_suspended"
3. If trial expired → auto-downgrade via `companyRepo.EndTrial`
4. Include `company_id`, `plan`, `on_trial` in JWT
5. Return `company` object with `days_remaining` in response

Also update magic link and SMS OTP flows: these must not create users without a `company_id`. If the user doesn't exist, they must go through `/register` first.

### 1.9 Team Invitation Flow (NEW — was missing from original plan)

Without this, the admin who registers cannot add their agents.

- `POST /api/v1/users/invite` — Admin sends invite email with a token
- `POST /api/v1/auth/accept-invite/:token` — Invited user sets password, joins company
- Frontend: Settings → Team → "Invite Agent" button
- Invite token stored in `password_reset_tokens` table (reuse existing infrastructure)

### 1.10 Background Job Runner (NEW — was missing from original plan)

Add a simple cron scheduler in `cmd/server/main.go` using `robfig/cron/v3`:

```go
c := cron.New()
c.AddFunc("@daily",  jobs.ExpireTrials(companyRepo))      // downgrade expired trials
c.AddFunc("@hourly", jobs.SendPaymentReminders(reminderRepo, emailSvc, waSvc))
c.AddFunc("@daily",  jobs.AlertExpiringEIDs(tenantRepo, notifRepo))  // Phase 2
c.Start()
```

### 1.11 Config & Environment

Remove from `.env` / `internal/config/config.go`:
```bash
# REMOVE
APP_COMPANY_ID=00000000-0000-0000-0000-000000000001
```

Add:
```bash
# ADD
ALLOW_REGISTRATION=true
TRIAL_DURATION_DAYS=90
TRIAL_PLAN_ID=starter

# Stripe (required to collect payment)
STRIPE_SECRET_KEY=sk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...
STRIPE_PRICE_ID_STARTER=price_...
STRIPE_PRICE_ID_PRO=price_...
STRIPE_PRICE_ID_BUSINESS=price_...

# Turnstile (for registration CAPTCHA)
TURNSTILE_SECRET_KEY=...
```

### 1.12 Frontend — Signup Page (`web/app/signup/page.tsx`)

Fields: Company Name, Subdomain (with `example.masaar.app` preview), Full Name, Email, Password (with strength indicator), Confirm Password, Cloudflare Turnstile.

On success: auto-login → redirect to `/onboarding` (new wizard, see 1.13).

### 1.13 Frontend — Onboarding Wizard (NEW — was missing)

New route: `web/app/onboarding/page.tsx`

4-step wizard shown once after first login:
1. **Company details** — logo, address, TRN number
2. **Add first property** — quick form
3. **Connect WhatsApp** — show webhook URL, link to Meta setup guide
4. **Invite your team** — add agent emails

After completing (or skipping): redirect to `/pipeline`. Store `onboarding_completed: boolean` in `company_settings`.

### 1.14 Frontend — Billing Trial Banner

In `web/app/(dashboard)/settings/billing/page.tsx`:
- Green banner (30+ days): "You're on a 90-day Starter trial — X days remaining"
- Yellow banner (7–30 days): "Trial ends in X days — upgrade to keep access"
- Red banner (<7 days): "Trial expires soon — upgrade now to avoid losing features"

### 1.15 Frontend — Auth Store & API Client

Add `company` to Zustand store in `web/store/auth.ts`:
```typescript
interface AuthState {
  user: AuthUser | null
  company: {
    id: string; name: string; plan: string
    on_trial: boolean; trial_ends_at: string | null; days_remaining: number
  } | null
  // ...
}
```

Add `auth.register(data)` to `web/lib/api.ts`.
Add `users.invite(data)` to `web/lib/api.ts`.

### Phase 1 Implementation Order

| # | Task | Files | Est. |
|---|------|-------|------|
| 1 | DB migration 0045 | `migrations/0045_saas_multi_tenancy.sql` | 2h |
| 2 | Domain models | `internal/domain/models.go` | 30m |
| 3 | Company repo (full) | `internal/repo/company.go` | 2h |
| 4 | User repo additions | `internal/repo/user.go` | 1h |
| 5 | Billing repo fix | `internal/repo/billing.go` | 45m |
| 6 | JWT claims update | `internal/api/handler/auth.go` | 30m |
| 7 | Trial middleware (JWT-direct) | `internal/api/middleware/trial.go` | 1h |
| 8 | ExtractClaims fix | `internal/api/middleware/auth.go` | 30m |
| 9 | Registration endpoint + Turnstile | `internal/api/handler/auth.go` | 3h |
| 10 | Login update + company checks | `internal/api/handler/auth.go` | 1h |
| 11 | Team invitation endpoint | `internal/api/handler/user.go` | 2h |
| 12 | Background job runner | `cmd/server/main.go`, `internal/jobs/` | 2h |
| 13 | Router update + middleware chain | `internal/api/router.go` | 30m |
| 14 | Config + env vars | `internal/config/config.go` | 30m |
| 15 | Quota middleware fix | `internal/api/middleware/quota.go` | 30m |
| 16 | Wire Stripe (live keys + test) | `internal/billing/stripe.go` | 1h |
| 17 | Frontend: signup page | `web/app/signup/page.tsx` | 2h |
| 18 | Frontend: onboarding wizard | `web/app/onboarding/page.tsx` | 3h |
| 19 | Frontend: login signup link | `web/app/login/page.tsx` | 15m |
| 20 | Frontend: billing trial banner | `web/app/(dashboard)/settings/billing/` | 1h |
| 21 | Frontend: auth store + API client | `web/store/auth.ts`, `web/lib/api.ts` | 1h |
| 22 | Frontend: team invite UI | `web/app/(dashboard)/settings/` | 2h |
| 23 | Manual testing + edge cases | — | 3h |

**Phase 1 total: ~30h (~3–4 weeks part-time)**

### Phase 1 Known Gotchas

| Case | Handling |
|------|----------|
| Email already registered | 409 "email_taken" |
| Subdomain taken | 409 "subdomain_taken" |
| `ALLOW_REGISTRATION=false` | 503 — for self-hosted deployments |
| Turnstile fails | 422 — do not process registration |
| Trial expired at login | Auto-downgrade, include warning in response |
| Company suspended | 403 on all requests |
| Magic link for new user | Must redirect to `/register` — cannot auto-create company |
| Existing deployment migration | Backfill script seeds existing company; re-login required for new JWT format |
| Rollback | Migration 0045 adds columns + backfills, but `NOT NULL` on `users.company_id` is not cleanly reversible — keep old image tagged |

### Phase 1 Migration Strategy (existing `masaar.dynamicweblab.com`)

1. Run migration 0045 (additive — no data loss; backfill assigns existing users to seed company)
2. Deploy new backend
3. Existing users re-login to get new JWT with `company_id`
4. Registration opens for new customers

---

## Phase 2 — UAE Real Estate Features (4–6 weeks)

**Goal:** Close the feature gap vs. Propspace and Property Monitor for the UAE rental market. These features are the primary reason UAE property managers would pay for this over a generic CRM.

### 2.1 PDC (Post-Dated Cheque) Management — HIGHEST PRIORITY

**Why:** PDC is the dominant payment method in UAE rentals. Every landlord manages 1–12 cheques per tenant per lease year. This is the most-requested feature by UAE property managers and the biggest gap vs. competitors.

**New DB migration: `0046_pdc_management.sql`**
```sql
CREATE TABLE pdc_cheques (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id        UUID NOT NULL REFERENCES companies(id),
  lease_id          UUID NOT NULL REFERENCES leases(id),
  tenant_id         UUID NOT NULL REFERENCES tenants(id),
  cheque_number     VARCHAR(50) NOT NULL,
  bank_name         VARCHAR(150) NOT NULL,
  account_number    VARCHAR(50),
  amount            NUMERIC(12,2) NOT NULL,
  due_date          DATE NOT NULL,           -- date cheque is dated
  presentation_date DATE,                    -- date presented to bank
  cleared_date      DATE,                    -- date funds received
  status            VARCHAR(20) NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','presented','cleared','bounced','cancelled','replaced')),
  bounce_reason     VARCHAR(255),
  police_case_number VARCHAR(100),
  legal_notice_sent_at TIMESTAMPTZ,
  replacement_cheque_id UUID REFERENCES pdc_cheques(id),
  notes             TEXT,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_pdc_company ON pdc_cheques(company_id);
CREATE INDEX idx_pdc_lease ON pdc_cheques(lease_id);
CREATE INDEX idx_pdc_due_date ON pdc_cheques(due_date);
CREATE INDEX idx_pdc_status ON pdc_cheques(status);
```

**API endpoints:**
- `GET /leases/:id/cheques` — list all PDCs for a lease
- `POST /leases/:id/cheques` — add cheque(s) to a lease
- `PATCH /cheques/:id/status` — mark presented / cleared / bounced
- `GET /cheques/upcoming` — company-wide view: cheques due in next 30 days
- `GET /cheques/overdue` — bounced + unpresented past-due-date

**Background job** (add to scheduler from 1.10):
- Daily: notify agents of cheques due in 7 days via WhatsApp + email

**Frontend pages:**
- `web/app/(dashboard)/leases/[id]/cheques/` — PDC schedule per lease
- `web/app/(dashboard)/cheques/` — company-wide PDC dashboard
- Color-coded status badges: grey (pending), blue (presented), green (cleared), red (bounced)

### 2.2 Emirates ID & Visa Expiry Tracking

**Why:** UAE law requires valid Emirates ID + residency visa to sign a lease. Property managers need alerts before expiry to avoid legal risk.

Add to `tenants` table:
```sql
ALTER TABLE tenants
  ADD COLUMN emirates_id_number  VARCHAR(20),
  ADD COLUMN emirates_id_expiry  DATE,
  ADD COLUMN visa_expiry         DATE,
  ADD COLUMN visa_type           VARCHAR(50); -- employment, investor, golden, etc.
```

Add to domain `Tenant` struct:
```go
EmiratesIDNumber string     `json:"emirates_id_number"`
EmiratesIDExpiry *time.Time `json:"emirates_id_expiry"`
VisaExpiry       *time.Time `json:"visa_expiry"`
VisaType         string     `json:"visa_type"`
```

**Background job:** Daily alert at 90/60/30 days before expiry → WhatsApp message to assigned agent + in-app notification.

**Frontend:** Red/yellow/green expiry badge on tenant card and tenant list.

### 2.3 Off-Plan Project & Sales Pipeline

**Why:** Off-plan sales are the largest revenue category for Dubai agents. Currently zero support for developer projects, unit inventory, or payment plan schedules.

**New DB migration: `0047_offplan_projects.sql`**
```sql
CREATE TABLE developer_projects (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id    UUID NOT NULL REFERENCES companies(id),
  name          VARCHAR(200) NOT NULL,
  developer     VARCHAR(200) NOT NULL,
  location      VARCHAR(200),
  area_slug     VARCHAR(100),  -- ties to BOS24 area data
  handover_date DATE,
  total_units   INT,
  status        VARCHAR(30) CHECK (status IN ('launching','under_construction','ready','handed_over')),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE project_units (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id    UUID NOT NULL REFERENCES developer_projects(id),
  unit_number   VARCHAR(50),
  floor         INT,
  bedrooms      INT,
  size_sqft     NUMERIC(10,2),
  price_aed     NUMERIC(14,2),
  status        VARCHAR(20) CHECK (status IN ('available','reserved','sold','unavailable')),
  reserved_by   UUID REFERENCES contacts(id),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE payment_plan_milestones (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  unit_id       UUID NOT NULL REFERENCES project_units(id),
  label         VARCHAR(100),  -- e.g. "On Booking", "20% on Handover"
  percentage    NUMERIC(5,2),
  amount_aed    NUMERIC(14,2),
  due_date      DATE,
  paid          BOOLEAN DEFAULT FALSE,
  paid_at       TIMESTAMPTZ
);
```

**API endpoints:**
- `GET/POST /projects` — developer projects
- `GET/POST /projects/:id/units` — unit inventory
- `PATCH /units/:id/reserve` — reserve a unit to a contact
- `PATCH /units/:id/sell` — mark sold, links to deal
- `GET/PATCH /units/:id/payment-plan` — manage milestones

**Frontend:**
- `web/app/(dashboard)/projects/` — project list + kanban by status
- `web/app/(dashboard)/projects/[id]/` — unit grid with availability heat map
- Payment plan milestone timeline per unit

### 2.4 Data Export (CSV / Excel)

**Why:** Every UAE property manager hands off monthly reports to their accountant. Without export, they maintain a parallel spreadsheet.

Add export endpoints:
- `GET /payments/export?format=csv&from=&to=` — payment history
- `GET /leases/export?format=csv` — all leases with status
- `GET /expenses/export?format=csv&from=&to=` — expense report
- `GET /cheques/export?format=csv&status=` — PDC register
- `GET /invoices/export?format=csv` — invoice ledger

Use Go's `encoding/csv` package. No new dependencies needed.

**Frontend:** "Export CSV" button on list views (payments, leases, expenses, cheques, invoices).

### 2.5 RERA Agent Registration Tracking

**Why:** Dubai RERA requires all agents to hold a valid RERA card (expires annually). Brokerages are legally responsible for ensuring agents are certified.

Add to `users` table:
```sql
ALTER TABLE users
  ADD COLUMN rera_number     VARCHAR(50),
  ADD COLUMN rera_expiry     DATE,
  ADD COLUMN rera_card_url   VARCHAR(500);
```

**Background job:** 60/30/7 day alerts to Admin before agent's RERA card expires.
**Frontend:** RERA badge on agent profile. Admin dashboard shows expiry list.

### 2.6 Tenant Portal (Read-Only)

**Why:** Tenants currently call agents for basic questions (when is my next payment, where is my lease). A self-service portal reduces agent workload significantly.

- New public route group (no company auth, tenant token instead)
- `GET /portal/:token/lease` — lease summary
- `GET /portal/:token/payments` — payment history
- `GET /portal/:token/cheques` — PDC schedule
- `GET /portal/:token/maintenance` — open maintenance requests
- Frontend: `web/app/portal/[token]/page.tsx` — simple read-only view, no login required
- Tenant token generated per lease, sent via WhatsApp

### Phase 2 Implementation Order

| # | Task | Est. |
|---|------|------|
| 1 | PDC: DB migration + domain models | 1h |
| 2 | PDC: repo layer (CRUD + upcoming/overdue queries) | 3h |
| 3 | PDC: API endpoints | 2h |
| 4 | PDC: background alert job | 1h |
| 5 | PDC: frontend (lease view + company dashboard) | 4h |
| 6 | Emirates ID / visa fields + migration | 1h |
| 7 | Emirates ID: expiry alert job | 1h |
| 8 | Emirates ID: frontend badges | 1h |
| 9 | Off-plan: DB migration | 1h |
| 10 | Off-plan: repo + API endpoints | 4h |
| 11 | Off-plan: frontend (project list + unit grid) | 5h |
| 12 | Off-plan: payment plan milestone UI | 3h |
| 13 | Data export endpoints (CSV) | 3h |
| 14 | Export buttons in frontend list views | 2h |
| 15 | RERA tracking fields + alerts | 2h |
| 16 | Tenant portal backend + frontend | 4h |

**Phase 2 total: ~38h (~4–5 weeks part-time)**

---

## Phase 3 — Growth & Market Reach (6–8 weeks)

**Goal:** Acquire users through integrations and mobile experience. These features differentiate Masaar from generic CRMs for UAE agents.

### 3.1 Property Portal Listing Sync (Bayut / Dubizzle / Property Finder)

UAE agents list properties on these portals daily. Integration = major time saving.

**How it works (XML feed, simpler than full API):**
1. Agent marks a `rental_property` as "publish to portals"
2. Backend generates an XML feed at `GET /feed/listings.xml` in Bayut's OLX XML format
3. Agent adds the feed URL in each portal's developer settings (one-time setup)
4. Portals auto-refresh every 24h

**What to build:**
- `published_to_portals: boolean` field on `rental_properties`
- XML feed generator at `/public/feed/listings.xml?api_key=xxx`
- API key validation for public feed access
- "Publish listing" toggle in property UI

For full two-way sync (lead import from Dubizzle API), that's Phase 4+.

### 3.2 Progressive Web App (PWA)

UAE agents use phones constantly — at viewings, in cars, on site.

- Add `manifest.json` and service worker to Next.js app
- Offline caching for contacts, leads, and pipeline view
- Push notifications via browser (for new WhatsApp messages)
- "Add to Home Screen" prompt
- Mobile-first UI audit of: pipeline Kanban, contact list, WhatsApp inbox, cheque dashboard

This is cheaper than a native app and unlocks 80% of mobile use cases.

### 3.3 Ejari Workflow Tracker

Currently only stores the Ejari number. Should track the full workflow.

Add to `leases` table:
```sql
ALTER TABLE leases
  ADD COLUMN ejari_status VARCHAR(30)
    CHECK (ejari_status IN ('not_started','documents_collected','submitted','registered','cancelled'))
    DEFAULT 'not_started',
  ADD COLUMN ejari_submitted_at TIMESTAMPTZ,
  ADD COLUMN ejari_registered_at TIMESTAMPTZ;
```

- Status tracker in lease detail page
- Document checklist: passport copy, visa, Emirates ID, tenancy contract, title deed, DEWA bill
- Alert when Ejari is not registered within 30 days of lease start (legal requirement)

### 3.4 WhatsApp Message Template Approval Workflow

Currently templates must be manually created in Meta Business Manager. Bridge the gap:

- Template creation UI (already has `message_templates` table and handlers)
- Add `submission_status: draft | submitted | approved | rejected` to templates
- Webhook handler for Meta template approval callbacks
- Frontend: "Submit for approval" button → shows status badge
- Alert agents when a template gets approved/rejected

### 3.5 Multi-Currency Support (AED primary, USD secondary)

Some deals (off-plan, luxury) are priced in USD. Invoices should support both.

- Add `currency: string` (default `AED`) to deals, invoices, payments
- Exchange rate display (static or via open FX API) on invoice PDF
- Frontend: currency selector on deal and invoice forms

### 3.6 Reporting Dashboard

Beyond the existing analytics API, build printable / shareable reports:

- Monthly performance report per agent (deals closed, commissions, leads handled)
- Portfolio health report (occupancy rate, upcoming vacancies, overdue cheques)
- Landlord statement PDF (per property: rental income, expenses, net yield)
- Export as PDF or email directly to landlord from the app

### 3.7 Bulk WhatsApp Campaigns

For agencies that want to broadcast to their contact lists:

- Contact segment selection (by area, property interest, lead stage)
- Template selection (approved templates only)
- Schedule send time
- Delivery report (sent, delivered, read, failed)
- Rate limiting to stay within Meta's limits

> Note: requires approved WhatsApp Business Account and message templates.

### Phase 3 Implementation Order

| # | Task | Est. |
|---|------|------|
| 1 | XML listing feed + publish toggle | 3h |
| 2 | PWA manifest + service worker | 2h |
| 3 | Mobile UI audit + fixes | 8h |
| 4 | Ejari workflow tracker | 3h |
| 5 | Template approval webhook handler | 2h |
| 6 | Multi-currency on deals/invoices | 3h |
| 7 | Agent performance report | 4h |
| 8 | Landlord statement PDF | 4h |
| 9 | Bulk WhatsApp campaign | 6h |

**Phase 3 total: ~35h (~5–6 weeks part-time)**

---

## Phase 4 — Scale & Enterprise (future)

Lower priority. Build after Phase 1–3 are shipping and paying customers exist.

| Feature | Why | Notes |
|---------|-----|-------|
| Subdomain routing (`company.masaar.app`) | Tenant isolation, branding | Nginx wildcard + frontend subdomain detection |
| Custom domain support | Enterprise accounts want `crm.theircompany.ae` | SSL via Let's Encrypt + CNAME |
| Native iOS + Android app | Full mobile feature parity | React Native or Flutter |
| Dubizzle / Bayut API (two-way sync) | Import leads from portals directly | Requires partner API access |
| Trakheesi / Abu Dhabi ADRA integration | Abu Dhabi property registration | API access via TAMM |
| DLD transaction data (SPA, NOC tracking) | Off-plan sales compliance | DLD API or BOS24 endpoint |
| White-label option | Agencies want their own branding | Theme config + logo upload |
| Accounting integration (Xero / QuickBooks) | Accountant handoff automation | Xero OAuth + invoice sync |
| Landlord portal (full) | Landlords log in, view portfolio | Separate auth with landlord role |
| AI-powered lead routing | Auto-assign leads by area expertise | ML model on agent history |
| SMS OTP for agents | Some agents prefer SMS 2FA | Twilio / UAE telco |
| PDPL data subject request workflow | UAE compliance automation | Right to erasure, data portability |

---

## Open Questions (decisions needed before launch)

| Question | Impact | Default if not decided |
|----------|--------|------------------------|
| AED or USD pricing on marketing site? | Conversion rate | Use AED (AED 109 / 299 / 749) |
| Confirm $29/$79/$199 USD plan pricing? | Revenue | Current `plans.go` values |
| Email verification required at signup? | Fraud prevention vs. friction | Skip for now; just Turnstile |
| Free plan (Community) quotas for SaaS? | Currently all quotas = 0 | Keep as self-host only tier |
| Stripe configured for production? | Cannot collect money without it | Highest priority in Phase 1 |
| Welcome email template content? | Onboarding experience | Simple "You're in!" + login link |
| SMTP provider? | Email delivery | Azure Communication Services (already in codebase) |
| Self-hosted vs. cloud as primary offer? | GTM strategy | Cloud-first, open-source as marketing |
| User limits per plan? | Upsell lever | None for now — company-level pricing |
| Subdomain format? | Branding | `company.masaar.app` in Phase 4 |

---

## Summary — What's Missing by Category

### Cannot sell today (Phase 0 + 1)
- Multi-tenancy (every phase depends on this)
- Stripe live keys not set — cannot charge customers
- No signup/registration endpoint
- No team invitation flow
- No onboarding wizard
- Background job runner missing — reminders never fire
- `ExtractClaims` bug — `company_id` from env, not JWT

### UAE-specific gaps vs. competitors (Phase 2)
- **PDC management** — biggest real estate pain point in UAE
- Emirates ID / visa expiry tracking
- Off-plan project + payment plan pipeline
- RERA agent certification tracking
- Ejari workflow (beyond just storing the number)
- Data export (CSV/Excel) — accountant handoff

### Growth features (Phase 3)
- Property portal listing sync (Bayut, Dubizzle, Property Finder)
- Mobile experience (PWA minimum)
- Landlord statement PDF
- Bulk WhatsApp campaigns
- Multi-currency (AED + USD)

### Technical debt in the plan (fix before building)
- `TrialCheck` must read `company_id` from JWT directly, not Locals (circular dependency)
- Turnstile token must be validated on the **backend** (not just shown on frontend)
- `billing.GetPlan()` has no `company_id` param — breaks in multi-tenant
- `quota.go` calls `GetPlan()` without `companyID` — needs update
- Magic link / OTP flows cannot create users without a company
- `company_settings` row must be created for every new company on registration
- `AllowRegistration` config flag must be checked in the handler, not just defined in config
