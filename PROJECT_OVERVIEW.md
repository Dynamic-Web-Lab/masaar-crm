# Masaar CRM - Complete Project Overview

**Last Updated:** April 25, 2026  
**Current Status:** Phase 6 Complete (Automatic Scoring + Auto-Tagging)

---

## 📊 Project Summary

**Masaar CRM** is an open-source, self-hosted WhatsApp CRM for UAE real estate teams. It combines a Go backend with a Next.js frontend, featuring real-time WhatsApp messaging, lead management with AI scoring, property intelligence, and complete communication history.

**Technology Stack:**
- **Backend:** Go 1.22 + Fiber framework
- **Frontend:** Next.js 14 + React 18 + Tailwind CSS
- **Database:** PostgreSQL 16 with pgvector
- **Cache:** Redis 7
- **AI/LLM:** Ollama (local LLM for thread summaries, intent parsing, lead enrichment, scoring)
- **Integrations:** Meta WhatsApp Business API, BuyOrSell24 (Dynamic Web Lab real estate data), SMTP email

---

## ✅ IMPLEMENTED FEATURES

### Phase 1: Core WhatsApp Integration
- [x] Webhook receiver for Meta Cloud API
- [x] Thread management (list, get, close)
- [x] Message retrieval and storage
- [x] Contact linking
- [x] WebSocket real-time events
- [x] Rate limiting (300 req/min)
- [x] Webhook verification (Meta handshake)

### Phase 2: Lead Management & Analytics
- [x] Kanban board (lead pipeline visualization)
- [x] Lead creation and stage transitions (new → contacted → qualified → proposal → won/lost)
- [x] Lead notes (add/update)
- [x] AI lead scoring (via Ollama - manual endpoint)
- [x] WebSocket broadcasts for real-time updates
- [x] Contact linking to leads
- [x] Property search (AI natural language + filters)
- [x] Yield analysis (rental vs sales stats)
- [x] Comparables (similar properties for pricing)
- [x] Market trends (area analytics)
- [x] Nearby schools/amenities

### Phase 3: Email Integration & Company Settings
- [x] SMTP email service (configurable)
- [x] Email sending endpoint
- [x] HTML email templates (invoice template included)
- [x] Email history log (pending/sent/failed/bounced)
- [x] Company settings for UAE compliance
  - Company name, VAT number, address
  - Bank details storage
  - Audit trail (who changed, when)
- [x] Invoice generation with proper VAT
- [x] PDF download endpoint

### Phase 4: WhatsApp Intelligence
- [x] Message intent parsing (property search, yield inquiry, price check, comparables, contact request)
- [x] Lead enrichment from conversation (budget, timeline, interests, property preferences)
- [x] Agent assist (suggest next action)
- [x] Automatic lead creation from thread
  - Extracts property interests from messages
  - Parses budget ranges
  - Auto-populates lead notes with enrichment data
  - Sets deal value from budget

### Phase 5a: WhatsApp Outbound Messaging
- [x] Send text messages via Meta API
- [x] Send template messages (pre-approved)
- [x] Send media (images, documents, audio, video)
- [x] Message status tracking (pending/sent/delivered/read/failed)
- [x] Outbound message history
- [x] Error handling and logging

### Phase 5b-d: Lead Intelligence & Communication History
- [x] Lead auto-tagging based on enrichment
  - Quality tags: "hot", "warm", "cold"
  - Interest tags: property type, areas ("marina", "downtown", etc.)
  - Timeline tags: "urgent", "1-3 months", "flexible"
  - Budget segments: "under-1M", "1M-2M", "2M+"
  - Category-based organization
- [x] Automatic lead scoring (0-100)
  - Triggered on message analysis, stage changes
  - Tracks score update timestamp
  - Color-coded display (0-30 red, 30-70 yellow, 70-100 green)
- [x] Unified communication history
  - Single timeline for all interactions
  - Types: WhatsApp inbound/outbound, email sent/received, calls
  - Status tracking per message
  - Linked to leads and contacts
  - External ID mapping for delivery tracking

### Phase 6: Automatic Scoring & Webhook Auto-Tagging
- [x] Intelligent lead scoring algorithm
  - Stage progression scoring (0-40 points)
  - Recency decay (0-20 points)
  - Engagement signals (0-30 points): message frequency, recent activity
  - Quality tag boost (0-10 points): hot/warm tags
  - Automatically triggered on lead creation, stage change, message analysis
- [x] Webhook-triggered auto-tagging
  - Real-time analysis of incoming WhatsApp messages
  - Quality signal detection (message length, detail density)
  - Interest extraction (property type, area, budget, transaction type)
  - Timeline signal detection (urgent vs flexible)
  - Runs asynchronously during webhook processing
  - Graceful fallback if AI unavailable

### Core Features (All Phases)
- [x] User authentication (JWT + bcrypt)
- [x] Role-based access control (Admin, Agent, Viewer)
- [x] Real-time notifications (WebSocket)
- [x] Audit logging (activity tracking)
- [x] Database migrations (goose)
- [x] Swagger/OpenAPI documentation
- [x] Arabic/English bilingual UI
- [x] RTL support
- [x] PDPL-compliant (UAE data residency)

---

## 📋 CURRENT ENDPOINTS

### Authentication
- `POST /api/v1/auth/login` — User login
- `POST /api/v1/auth/refresh` — Refresh token
- `DELETE /api/v1/auth/logout` — Logout + token blacklist

### Dashboard & Stats
- `GET /api/v1/stats` — Dashboard statistics

### Contacts
- `GET /api/v1/contacts` — List all contacts
- `GET /api/v1/contacts/:id` — Get contact details
- `POST /api/v1/contacts` — Create contact
- `PATCH /api/v1/contacts/:id` — Update contact
- `DELETE /api/v1/contacts/:id` — Delete contact

### Leads
- `GET /api/v1/leads` — Kanban board view
- `GET /api/v1/leads/:id` — Get lead details
- `POST /api/v1/leads` — Create lead
- `PATCH /api/v1/leads/:id/stage` — Update lead stage
- `PATCH /api/v1/leads/:id/notes` — Update lead notes

### WhatsApp (Inbound)
- `GET /webhooks/whatsapp` — Webhook verification (Meta)
- `POST /webhooks/whatsapp` — Receive messages (Meta)
- `GET /api/v1/threads` — List all conversations
- `GET /api/v1/threads/:id` — Get thread details
- `GET /api/v1/threads/:id/messages` — Get all messages in thread
- `POST /api/v1/threads/:id/close` — Close conversation

### WhatsApp (Outbound)
- `POST /api/v1/threads/:id/send-message` — Send text message
- `POST /api/v1/threads/:id/send-template` — Send template
- `GET /api/v1/threads/:id/outbound-messages` — Get sent messages history

### Message Analysis
- `POST /api/v1/messages/analyze` — Parse intent & enrichment
- `POST /api/v1/messages/suggest-action` — Get AI-recommended next action
- `POST /api/v1/messages/auto-create-lead` — Create lead from thread

### Property Intelligence
- `POST /api/v1/properties/search` — AI natural language search
- `GET /api/v1/properties/transactions` — Property sales history
- `GET /api/v1/properties/buildings` — Building autocomplete
- `GET /api/v1/properties/buildings/:id` — Building details
- `GET /api/v1/properties/schools/nearby` — Amenities nearby
- `GET /api/v1/properties/yield-analysis` — Rental yield analysis
- `GET /api/v1/properties/comparables` — Similar properties
- `GET /api/v1/properties/market-trends` — Market analytics

### AI & Automation
- `POST /api/v1/ai/summarize/:thread_id` — Manual thread summary

### Email
- `POST /api/v1/emails/send` — Send email
- `GET /api/v1/emails/history` — Email history for entity

### Invoices & Deals
- `GET /api/v1/invoices/:id` — Get invoice
- `GET /api/v1/invoices/:id/pdf` — Download PDF
- `POST /api/v1/invoices` — Create invoice
- `POST /api/v1/invoices/:id/send` — Mark sent
- `PATCH /api/v1/invoices/:id/status` — Update status
- `GET /api/v1/deals` — List deals
- `POST /api/v1/deals` — Create deal
- `PATCH /api/v1/deals/:id/stage` — Update deal stage

### Settings
- `GET /api/v1/settings/bos24` — Get BuyOrSell24 token
- `PATCH /api/v1/settings/bos24` — Update BOS24 token
- `GET /api/v1/settings/company` — Get company details
- `PATCH /api/v1/settings/company` — Update company info (VAT, address, etc.)

### Notifications
- `GET /api/v1/notifications` — List notifications
- `PATCH /api/v1/notifications/:id/read` — Mark as read
- `GET /ws/notifications` — WebSocket for real-time

---

## 🚀 NEXT PRIORITIES (Phase 7+)

### HIGH PRIORITY
1. **Communication History UI**
   - Display unified timeline in lead detail
   - Filter by communication type (WhatsApp, email, calls)
   - Show delivery status and timestamps
   - Last contacted date for follow-up planning
   - Search within communication history

2. **Agent Assist UI**
   - Show suggested next action in message interface
   - Display recommended properties based on parsed interests
   - One-click actions (Send message, Send proposal, Schedule follow-up)
   - AI-drafted reply suggestions in chat

3. **Lead Analytics Dashboard**
   - Leads by source (WhatsApp, email, website)
   - Conversion funnel (new → contacted → qualified → proposal → won)
   - Average time in each stage
   - Score distribution and trends
   - Agent performance metrics

4. **WhatsApp Outbound Automation**
   - Send template on lead creation ("Thanks for inquiry, agent will contact you")
   - Send follow-up reminders (based on timeline tag)
   - Schedule messages for optimal sending times
   - Bulk messaging to segment

### MEDIUM PRIORITY
5. **Lead Analytics Dashboard**
   - Leads by source (WhatsApp, email, website)
   - Conversion funnel (new → contacted → qualified → proposal → won)
   - Average time in each stage
   - Score distribution

6. **Agent Assist UI**
   - Show suggested action in message interface
   - Suggest next message (AI-drafted)
   - Recommended property list based on parsed interests
   - One-click actions (e.g., "Send proposal")

7. **Smart Lead Routing**
   - Assign leads automatically to available agents
   - Round-robin distribution
   - Skill-based routing (by area/property type)

8. **Reply Detection & Threading**
   - Link WhatsApp reply to original message
   - Create conversation threads
   - Show reply context

### LOWER PRIORITY
9. **Call Logging**
   - Log phone calls to lead
   - Duration, outcome, notes
   - Integrated in communication history

10. **Bulk Actions**
    - Send bulk WhatsApp to segment
    - Bulk email to tag
    - Bulk stage update
    - Bulk assign to agent

11. **Reporting & Exports**
    - Lead reports (CSV, PDF)
    - Communication logs
    - Agent performance metrics
    - ROI by source

12. **Mobile App**
    - React Native app for agents
    - Push notifications
    - Quick message reply
    - Lead creation on-the-go

---

## 🗄️ DATABASE SCHEMA

### Core Tables
- **users** — Agents, admins, viewers
- **contacts** — Customer profiles
- **leads** — Sales pipeline
- **deals** — Deal tracking
- **invoices** — VAT invoices with QR codes

### WhatsApp Tables
- **whatsapp_threads** — Conversations
- **whatsapp_messages** — Inbound messages
- **whatsapp_outbound** — Outbound message tracking

### Intelligence Tables
- **lead_tags** — Auto-applied tags by category
- **communication_history** — Unified message log (WhatsApp, email, calls)

### Integration Tables
- **email_history** — Email sends/failures
- **api_settings** — BuyOrSell24 token, company details
- **company_settings** — VAT number, bank details, audit trail

### Support Tables
- **audit_logs** — Activity tracking
- **notifications** — User alerts
- **refresh_tokens** — Session management (in Redis)

---

## 🔧 CONFIGURATION

### Required Environment Variables

**Server:**
```bash
PORT=8080
APP_ENV=development
DATABASE_URL=postgres://user:pass@localhost:5432/masaar
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-long-random-secret
```

**WhatsApp:**
```bash
WA_VERIFY_TOKEN=your-webhook-token
WA_API_VERSION=v19.0
WA_PHONE_NUMBER_ID=your-phone-id
WA_ACCESS_TOKEN=your-business-token
```

**Email (Optional):**
```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password
```

**Real Estate Data (Optional):**
```bash
BOS24_API_TOKEN=your-api-key
```

**AI:**
```bash
OLLAMA_BASE_URL=http://localhost:11434
OLLAMA_MODEL=llama3
```

---

## 📈 Key Metrics

**What's Tracked:**
- Lead quality score (0-100)
- Lead age and stage progression
- Communication frequency (messages, emails, calls)
- Agent response time
- Conversion funnel (new → won)
- Email open/click rates (planned)
- Property search insights (areas, budgets, types)

---

## 🎯 Architecture Highlights

1. **Optional Services:** Email, WhatsApp outbound, property intelligence all gracefully disabled if not configured
2. **AI-First:** All intelligence runs locally via Ollama (no external APIs required)
3. **Real-Time:** WebSocket broadcasts keep UI in sync instantly
4. **Audit Trail:** PDPL compliance with full change tracking
5. **Bilingual:** Full Arabic/English support with RTL layout
6. **Scalable:** Role-based access, job queue pattern for async work

---

## 📝 Files Structure

```
masaar-crm/
├── cmd/server/              # Server entry point
├── internal/
│   ├── api/
│   │   ├── handler/         # HTTP request handlers
│   │   ├── middleware/      # JWT, rate limiting, auth
│   │   └── router.go        # Route registration
│   ├── domain/              # Models and constants
│   ├── repo/                # Database access
│   ├── ai/                  # Ollama LLM integration
│   ├── email/               # SMTP service
│   ├── whatsapp/            # WhatsApp sender
│   ├── bos24/               # Real estate API client
│   ├── pdf/                 # Invoice PDF generation
│   └── ws/                  # WebSocket hub
├── web/                     # Next.js frontend
├── migrations/              # SQL migrations (goose)
├── docs/                    # Swagger API docs (auto-generated)
└── docker/                  # Docker compose files
```

---

## 🚢 Deployment

**Development:**
```bash
docker compose -f docker/docker-compose.yml up
```

**Production:**
```bash
docker compose -f docker/docker-compose.prod.yml up
```

---

## 📞 Getting Help

See `/help` command in Claude Code or consult:
- **API Docs:** `/docs` endpoint (Swagger UI)
- **Codebase Guide:** `CLAUDE.md` file
- **Feature Audit:** `INTEGRATION_AUDIT.md` (what's still missing)
- **Integration Guide:** `PROJECT_OVERVIEW.md` (this file)

---

## 📄 License

MIT - Open source and free to use

---

## 🎉 Current Status

**Phase 6 Complete:**
- ✅ Automatic lead scoring updates (with age decay & engagement signals)
- ✅ Webhook-triggered auto-tagging on messages
- ✅ Intelligent scoring algorithm (stage, recency, engagement, quality)
- ✅ WhatsApp outbound messaging
- ✅ Lead auto-tagging system (Phase 5)
- ✅ Unified communication history
- ✅ Email integration with templates
- ✅ Company settings & VAT compliance

**Ready for Phase 7:**
- UI implementation for communication history in lead detail
- Agent assist interface with suggested actions
- Lead analytics dashboard
- Bulk actions on tagged leads
- Advanced message scheduling and automation

---

**Total Development Time:** ~90 hours (Phases 1-6)  
**Migrations Deployed:** 14  
**API Endpoints:** 40+  
**Core Services:** 8 (auth, contact, lead, WhatsApp, email, property, AI, billing)
**Intelligent Services:** 3 (ai.Client, ScoringService, TaggingService)
**Test Coverage:** Manual (automated tests planned)

