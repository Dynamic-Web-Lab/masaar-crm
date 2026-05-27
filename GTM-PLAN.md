# Masaar CRM — Go-to-Market & Business Operations Plan

> **Scope**: Everything SAAS-PLAN.md doesn't cover — marketing, legal, compliance, support, revenue operations, and UAE-specific business requirements.
> **Target**: UAE real estate agencies, 1–50 agents, Dubai-first then Abu Dhabi

---

## 1. Who We're Selling To

### Ideal Customer Profile (ICP)

**Primary: Small-to-mid UAE real estate brokerage**
- 2–15 agents
- Licensed with RERA (Dubai) or ADRA (Abu Dhabi)
- Currently managing leads on WhatsApp + Excel
- Pain: losing track of leads, no PDC visibility, manual Ejari paperwork
- Budget: AED 400–1,500/month feels reasonable (they pay Bayut/Dubizzle AED 10,000–50,000/year)
- Decision maker: Owner or Operations Manager

**Secondary: Property management company**
- Manages 20–200 residential units
- Pain: PDC tracking, tenant communication, maintenance coordination
- Probably already using some software (Yardi, MRI) but it's expensive and not WhatsApp-native

**Not yet (Phase 4+):**
- Enterprise brokerages (50+ agents) — need SLAs, custom contracts, SSO
- Off-plan developers — different workflow, different sales cycle

### Where They Hang Out
- RERA-certified training courses (DREI — Dubai Real Estate Institute)
- Cityscape Global (annual real estate expo, Dubai World Trade Centre)
- Arabia Property Awards
- WhatsApp groups for UAE real estate agents (large, active)
- LinkedIn UAE real estate community
- Instagram (UAE agents are very active — property content, market updates)
- Arabian Business, Gulf News Property section

---

## 2. Marketing Site

**Status: Does not exist. Must be built before any paid acquisition.**

### Domain Options
- `masaar.app` — clean, memorable
- `masaarcrm.com` — descriptive
- `getmasaar.com` — action-oriented

Recommend `masaar.app`. Register all three and redirect.

### Pages Needed

| Page | Content | Priority |
|------|---------|---------|
| `/` | Hero, pain → solution, 3 key features, social proof, pricing teaser, CTA | P0 |
| `/pricing` | Plan comparison table in AED, FAQ, "Start free trial" CTAs | P0 |
| `/features` | Feature deep-dives: WhatsApp CRM, PDC management, Ejari, AI scoring | P1 |
| `/for/property-managers` | Landing page for property management use case | P1 |
| `/for/sales-agents` | Landing page for brokerage use case | P1 |
| `/blog` | Arabic + English content: UAE real estate tips, RERA guides, market updates | P2 |
| `/about` | Founder story, why UAE, open-source commitment | P2 |
| `/legal/privacy` | PDPL-compliant privacy policy | P0 (required at launch) |
| `/legal/terms` | Terms of service | P0 (required at launch) |
| `/legal/dpa` | Data Processing Agreement for enterprise | P2 |

### Copy Angle
Lead with **pain**, not features. UAE agents' #1 frustration: "My deals are in 5 WhatsApp groups and I can't find anything."

Hero headline options (Arabic + English):
- AR: "كل صفقاتك في مكان واحد" — "All your deals in one place"
- EN: "The CRM built for UAE real estate agents — WhatsApp-native, Arabic-first"
- EN: "Stop managing your listings in WhatsApp groups"

### Pricing Display (AED, not USD)

| Plan | AED/month | AED/year (2 months free) |
|------|-----------|--------------------------|
| Community | Free (self-hosted) | — |
| Starter | AED 109 | AED 981 |
| Pro | AED 299 | AED 2,690 |
| Business | AED 749 | AED 6,741 |

> Stripe supports AED billing. Display AED on marketing site, charge in AED via Stripe.
> Note: current `plans.go` shows USD ($29/$79/$199). Update `PriceUSDMonth` or add `PriceAEDMonth` — AED/USD is pegged at 3.67.

### Social Proof Needed Before Launch
- 3–5 beta agency logos / testimonials
- 1 case study: "How [Agency Name] reduced missed payments by X%"
- Screenshot walkthrough / demo video (Arabic voiceover + English subtitles)

---

## 3. Legal & Compliance

### 3.1 PDPL Compliance (UAE Federal Decree-Law No. 45 of 2021)

UAE's data protection law came into effect June 2023. Non-compliance = fines up to AED 20M.

**What's required:**

| Requirement | Status | Action |
|-------------|--------|--------|
| Privacy policy (Arabic + English) | ❌ Missing | Write before launch |
| Purpose limitation — data used only for stated purposes | ✅ Audit logs exist | Document purpose per data type |
| Data subject rights (access, correction, deletion) | ❌ No UI | Add "Delete my data" flow for tenants/contacts |
| Data breach notification (72 hours) | ❌ No process | Write incident response runbook |
| Data residency — UAE resident data should stay in UAE or adequate country | ⚠️ Depends on hosting | Host on UAE-region cloud (see Section 7) |
| Consent for marketing communications | ❌ Missing | Add consent checkbox at registration |
| Data retention limits | ❌ No policy | Define + implement retention schedule |
| DPA with sub-processors (Stripe, Meta, Ollama host) | ❌ Missing | Document sub-processors in privacy policy |

**Immediate actions:**
1. Draft privacy policy listing: data collected, purpose, retention, sub-processors, data subject rights contact
2. Add `marketing_consent: boolean` field to `users` table at registration
3. Add "Delete my account data" endpoint (soft-delete contacts, anonymize personal data)
4. Document that WhatsApp message content is processed by Meta (already in Meta's TOS, but must be disclosed)

### 3.2 UAE VAT on SaaS (5%)

UAE VAT law applies to digital services sold to UAE businesses.

**Threshold:** Register for VAT once revenue exceeds AED 375,000/year (~AED 31,250/month).
**Recommended:** Register proactively if targeting UAE businesses — it signals legitimacy and most agencies are VAT registered and expect VAT invoices.

**What's needed:**
- UAE VAT Registration (via FTA — Federal Tax Authority portal)
- Get a TRN (Tax Registration Number)
- Add TRN to all invoices issued to customers
- Charge 5% VAT on subscription fees to UAE-based customers
- File quarterly VAT returns

**In-product:**
- `companies` table needs `trn_number VARCHAR(20)` for customer's TRN on their invoices from you
- Billing invoice PDFs (receipts for subscription) must show your TRN + VAT line
- For B2B UAE sales: VAT is reverse-charged (customer accounts for it) — check with UAE tax advisor

### 3.3 Terms of Service

Must cover:
- Acceptable use (no SPAM, no illegal data)
- Data ownership (customer owns their data)
- Uptime SLA (or disclaim for Starter/Community)
- Payment terms, refund policy
- Account suspension conditions (non-payment, abuse)
- Governing law: **DIFC (Dubai International Financial Centre) courts** — preferred for tech SaaS in UAE, internationally enforceable

### 3.4 RERA Compliance (for customers)

You're not a real estate broker — you're a software vendor. But:
- Do not store or process property transaction data in a way that conflicts with DLD/RERA data governance
- Ejari data (contract numbers, registration) must not be used outside customer's own context
- If you integrate DLD API in Phase 4, you'll need a data license agreement with DLD

### 3.5 Meta / WhatsApp Business Policy Compliance

Meta has strict policies for WhatsApp Business API usage that affect your customers:
- Must display to customers: "This business uses WhatsApp Business API"
- Template messages require Meta approval — your app should warn agents before they violate this
- Bulk messaging limits (conversation-based pricing) — agents need to understand cost implications
- Document this clearly in your help center: "How WhatsApp Business API billing works"

---

## 4. Payment & Revenue Operations

### 4.1 Payment Gateway

**Stripe** is already integrated in the codebase. It supports AED billing natively.

However, UAE customers strongly prefer **local payment methods**:
- Credit/debit cards via local UAE banks work fine with Stripe
- Apple Pay (very common in UAE — add Stripe's Apple Pay integration)
- **No direct debit or bank transfer option** in Stripe for UAE — agencies that want to pay via bank transfer need a manual invoice process

**Consider adding Telr or PayTabs as secondary gateway** (Phase 2+):
- UAE-headquartered payment gateways
- Support local bank direct debit
- Better for customers who distrust international processors
- Telr: Arabic dashboard, UAE bank settlement
- PayTabs: common in KSA + UAE SaaS

### 4.2 Revenue Metrics to Track from Day 1

Set up tracking in a spreadsheet or simple analytics before you have a proper dashboard:

| Metric | Why it matters |
|--------|---------------|
| MRR (Monthly Recurring Revenue) | Core health metric |
| ARR (Annual Recurring Revenue) | Investor/planning metric |
| Trial-to-paid conversion rate | Target: >15% is good for SaaS |
| Trial start → first meaningful action (< 24h) | Onboarding effectiveness |
| Churn rate (monthly) | Target: <3% for SMB SaaS |
| Average Revenue Per Account (ARPA) | Which plan mix is working |
| CAC (Customer Acquisition Cost) | Cost per paid customer |
| LTV (Lifetime Value) | LTV:CAC should be >3x |
| Feature adoption rate | Which features drive retention |

**UAE-specific:** Track separately by emirate (Dubai vs. Abu Dhabi vs. Sharjah) — different market dynamics.

### 4.3 Billing Edge Cases for UAE

| Scenario | Handling |
|----------|---------|
| Agency wants annual invoice upfront | Offer annual plan at 2 months free; Stripe supports annual subscriptions |
| Agency wants to pay by bank transfer | Manual invoice → mark paid in Stripe dashboard |
| Agency on community plan wants upgrade mid-month | Prorated via Stripe |
| Agency registered in Free Zone | No VAT difference — Free Zone businesses are still subject to UAE VAT on B2B SaaS |
| Non-UAE agency (KSA, Kuwait, Bahrain) | No UAE VAT, but check their local tax requirements |

---

## 5. Customer Support Infrastructure

### 5.1 Support Channels (Priority Order)

**WhatsApp Business** — ironic, but it's the right channel for this market. UAE agencies will message you on WhatsApp before opening a support ticket. Set up a dedicated support number from Day 1.

**In-app chat** — Intercom, Crisp, or Chaport. Crisp has a generous free tier and supports Arabic. Embed in the dashboard sidebar.

**Email** — `support@masaar.app` — route to same inbox as WhatsApp/Crisp.

**Documentation** — see Section 5.2.

**Do not:** Twitter/X DMs, Facebook — wrong platform for B2B UAE.

### 5.2 Knowledge Base / Help Center

**Must exist before public launch.** Agencies will not call you for basic questions if docs exist.

Platform options:
- **Notion (public)** — free, easy to maintain, decent search
- **GitBook** — cleaner, supports Arabic RTL
- **Intercom Articles** — integrated with support chat

**Content needed (Arabic + English):**

| Article | Priority |
|---------|---------|
| Getting started: first 15 minutes guide | P0 |
| How to connect WhatsApp Business API | P0 |
| How to add your first property and tenant | P0 |
| How to set up a lease and PDC schedule | P0 |
| How to invite your team | P0 |
| Understanding billing and the free trial | P0 |
| How to register Ejari (workflow guide, not app-specific) | P1 |
| How WhatsApp API billing works (Meta conversation fees) | P1 |
| Exporting data for your accountant | P1 |
| PDPL: how Masaar protects your data | P1 |
| API documentation (for agencies with developers) | P2 |
| Video walkthroughs (screen recordings) | P2 |

### 5.3 Support SLA by Plan

| Plan | Response time | Channel |
|------|--------------|---------|
| Community | None (community forum only) | GitHub Issues |
| Starter | 48 business hours | Email / in-app chat |
| Pro | 24 business hours | Email / in-app chat / WhatsApp |
| Business | 4 business hours | Dedicated WhatsApp + email |

### 5.4 Onboarding Call

For Pro and Business plans: offer a free 30-minute onboarding video call.
- Walk through: WhatsApp connection, first property, first tenant, first lease + PDC schedule
- Done over Google Meet or Zoom (avoid MS Teams — UAE agencies don't use it)
- Arabic-language call option is a strong differentiator

---

## 6. Go-to-Market Strategy

### 6.1 Beta Program (Before Launch)

Recruit 5–10 agencies for 3-month free access in exchange for:
- Weekly feedback calls (30 min)
- 1 written testimonial
- Logo permission for marketing site
- Introduction to 2 other agencies

**Where to find beta users:**
- RERA networking events in Dubai
- UAE real estate WhatsApp groups (request to join as a vendor, be transparent)
- LinkedIn: search "RERA certified agent Dubai", "property manager Dubai"
- Direct outreach to small agencies near DIFC, Business Bay, JLT

### 6.2 Launch Sequence

**Week 1–2 (Soft launch):**
- Deploy to production with beta agencies
- Marketing site live
- Post in UAE real estate WhatsApp groups: "We're launching Masaar — free trial, looking for feedback"
- LinkedIn post from founder: personal story + product demo video

**Week 3–4 (Public launch):**
- Product Hunt launch (list in Arabic-friendly category)
- Press release to Gulf News Property, Arabian Business, Zawya
- Post demo video on Instagram Reels (Arabic audio is key — Dubai real estate agents are heavily on Instagram)
- LinkedIn carousel: "5 things UAE real estate agents still do in WhatsApp groups that cost them deals"

**Month 2 onwards:**
- Content marketing: weekly blog post (Arabic) on real estate tips, RERA updates, market data
- YouTube channel: Arabic walkthroughs of each feature
- SEO targets: "real estate CRM UAE", "property management software Dubai", "Ejari software", "PDC tracking software UAE"

### 6.3 Acquisition Channels (Ranked by Cost/Effort)

| Channel | Cost | Effort | Expected Quality |
|---------|------|--------|-----------------|
| UAE real estate WhatsApp groups | Free | Low | High — warm audience |
| LinkedIn outreach (owner/ops manager) | Free | Medium | High |
| RERA / DREI course sponsorship | AED 3,000–10,000/event | Medium | High |
| Google Ads: "real estate CRM Dubai" | AED 10–30/click | Low | Medium |
| Instagram Reels (Arabic content) | Free (organic) | High | Medium |
| Property portal partnership (Bayut) | Rev-share or flat fee | High | High |
| Referral program (agency refers agency) | 1 free month/referral | Low | Very High |
| ProductHunt | Free | Medium | Low (not UAE-specific) |
| Cold email to RERA-licensed agencies | Free | Medium | Low (low response) |

### 6.4 Referral Program

Most UAE real estate deals happen on relationships. Build a referral program from launch:
- Referring agency gets 1 free month per paid referral
- Referred agency gets 15% off first 3 months
- Track via unique referral link
- Build this into the billing settings page

### 6.5 Partnerships

| Partner | Why | How |
|---------|-----|-----|
| DREI (Dubai Real Estate Institute) | Access to RERA training participants | Sponsor a course, speak at event |
| Bayut / Dubizzle | Mutual — they get better data, we get leads | Approach their business dev team |
| UAE Proptech Association | Credibility + network | Join as a member |
| UAE banks (Emirates NBD, FAB) | Bank feed integration drives stickiness | API partnership program |
| Law firms specializing in real estate | They refer agencies needing contract management | Referral agreement |
| Property management consultants | They advise agencies on tools | Referral program |

---

## 7. Infrastructure & Hosting

### 7.1 UAE Data Residency (PDPL)

PDPL requires adequate protection for UAE resident personal data. Safest option: host in UAE.

**UAE-region cloud options:**
| Provider | Region | Notes |
|----------|--------|-------|
| AWS | `me-central-1` (UAE) | Available since 2022, full services |
| Azure | `UAE North` (Dubai) | Strong enterprise adoption in UAE |
| Google Cloud | No UAE region yet | Not recommended for PDPL |
| G42 Cloud | UAE-only | Government and ADNOC-aligned, enterprise only |
| Khazna Data Centers | Dubai/Abu Dhabi | Local provider, less mature |

**Recommended:** AWS `me-central-1` for Phase 1. Later add Azure UAE North for enterprise customers who require Microsoft ecosystem.

Current Docker Compose setup can be deployed on a single AWS EC2 `t3.medium` (AED ~400/month) for early stage. Move to ECS/EKS when you hit 50+ companies.

### 7.2 Uptime & Reliability Targets

| Plan | Target SLA | If breached |
|------|-----------|------------|
| Community | None | No commitment |
| Starter | 99% (7.2h downtime/month) | No compensation |
| Pro | 99.5% (3.6h downtime/month) | 1 day credit |
| Business | 99.9% (43m downtime/month) | 1 week credit |

**Status page:** Use a free tool (Betterstack, UptimeRobot) to show live uptime. UAE agencies will check this before signing up.

### 7.3 Backups

| Data | Frequency | Retention | Location |
|------|-----------|-----------|---------|
| PostgreSQL | Daily snapshot + WAL | 30 days | S3 (me-central-1) |
| Redis | No backup needed | Ephemeral | — |
| Uploaded documents / files | Continuous replication | 90 days | S3 versioned |
| Audit logs | Immutable | 7 years (PDPL/UAE commercial law) | S3 Glacier |

### 7.4 Security Checklist Before Launch

- [ ] SSL/TLS everywhere (Let's Encrypt or AWS ACM)
- [ ] Web Application Firewall (AWS WAF or Cloudflare)
- [ ] DDoS protection (Cloudflare free tier minimum)
- [ ] Rate limiting on all public endpoints (already implemented for login/webhook)
- [ ] Secrets rotation schedule (JWT secret, Stripe keys, DB password)
- [ ] Penetration test or vulnerability scan before launch (can use OWASP ZAP for free)
- [ ] Dependency audit (`go mod tidy` + `npm audit`)
- [ ] Enforce HTTPS redirect at nginx/load balancer level
- [ ] HSTS header enabled
- [ ] CSP header on frontend
- [ ] No secrets in git history (run `git-secrets` or `truffleHog` scan)

---

## 8. Demo Environment

**Must exist before any sales outreach.** Asking an agency to sign up without seeing the product first is a hard sell.

### 8.1 Live Interactive Demo

A sandboxed demo account pre-loaded with UAE real estate data:
- Company: "Al Futtaim Properties Demo"
- 3 agents (Ahmed, Sara, Mohammed)
- 10 properties (Dubai Marina, JBR, Business Bay, JLT)
- 20 tenants with Emirates ID + visa fields
- 5 active leases with PDC schedules
- Pipeline with leads at various stages
- Sample WhatsApp conversation threads
- Sample invoices with 5% VAT

**Implementation:**
- Seed script: `scripts/seed_demo.sql`
- Demo login button on marketing site (email: `demo@masaar.app`, password: `demo123`)
- Demo data resets every 24h via a cron job (prevent vandalism)
- Demo account is read-only for destructive actions (no delete, no WhatsApp send)

### 8.2 Video Demo

A 3-minute screen recording showing the golden path:
- Receive a WhatsApp lead → create contact → add to pipeline → create lease → add PDC schedule → generate invoice
- Arabic voiceover + English subtitles
- Host on YouTube (unlisted) + embed on homepage

---

## 9. Product Analytics

Track in-product behavior from Day 1 to understand what drives retention and conversion.

### 9.1 Analytics Tool

Options (pick one):
- **PostHog** — open source, can self-host (PDPL advantage), generous free tier
- **Mixpanel** — industry standard, UAE region data storage available
- **Plausible** — privacy-first, GDPR/PDPL friendly, simple

Recommend **PostHog** self-hosted on the same AWS infrastructure — no data leaves UAE.

### 9.2 Key Events to Track

| Event | Why |
|-------|-----|
| `company_registered` | Top of funnel |
| `onboarding_step_completed` (step 1/2/3/4) | Onboarding dropoff |
| `whatsapp_connected` | Strong activation signal — stickiness driver |
| `first_lead_created` | Activation |
| `first_lease_created` | Deep activation |
| `pdc_schedule_created` | Power user signal |
| `invoice_generated` | Revenue workflow adopted |
| `ai_summary_used` | AI feature adoption |
| `plan_upgraded` | Conversion |
| `plan_downgraded` | Churn risk signal |
| `export_csv_clicked` | Accountant workflow adopted |
| `team_member_invited` | Collaboration signal (reduces churn) |

### 9.3 Activation Definition

**Activated user** = company that, within 7 days of signup, has:
1. Connected WhatsApp OR added 1 contact
2. Created 1 lead OR 1 property
3. Invited at least 1 team member

Companies that hit all 3 within 7 days are 3x more likely to convert to paid (industry benchmark).

---

## 10. Customer Success

### 10.1 Trial-to-Paid Conversion Playbook

| Day | Action | Channel |
|-----|--------|---------|
| 0 | Welcome email (Arabic + English) | Email |
| 1 | "Did you connect WhatsApp?" check-in if not activated | Email |
| 3 | Feature tip: "Have you tried the PDC schedule?" | In-app + email |
| 7 | Offer onboarding call if not activated | Email |
| 14 | "You're halfway through your trial" — show value metrics | Email |
| 30 | "One month left" upgrade nudge | Email + in-app banner |
| 60 | "30 days left" — show what they'd lose | Email + in-app banner |
| 75 | "15 days left" — offer 1 month extension for feedback call | Email |
| 89 | "1 day left" — final push | Email + WhatsApp if provided |
| 90 | Trial ends → downgrade to Community | Automated |

**All emails should be in both Arabic and English** based on user's `lang_pref`.

### 10.2 Churn Prevention

Warning signals to watch (trigger a check-in):
- No login in 14 days
- WhatsApp disconnected (webhook stops receiving)
- Plan downgrade event
- Support ticket with frustration language
- Failed payment (Stripe `invoice.payment_failed` webhook)

Response: personal WhatsApp message from founder/support, not automated email.

### 10.3 Expansion Revenue

Ways to grow revenue within existing accounts:
- Seat expansion: currently no per-user pricing, but user limits could be a Pro+ feature
- Annual plan upsell: offer 2 months free at month 3 (when trust is established)
- Feature unlock: BOS24 calls, AI requests — agencies that hit limits will upgrade
- Referral: happy agency refers sister agency or friend

---

## 11. Localization

### 11.1 Languages

The app has Arabic/English switching. The business needs the same.

| Asset | Arabic | English |
|-------|--------|---------|
| Marketing site | Required | Required |
| In-app UI | ✅ Done | ✅ Done |
| Email templates | ❌ Missing | ❌ Missing |
| Knowledge base / help docs | ❌ Missing | ❌ Missing |
| Invoice PDFs | ✅ (VAT invoice) | ✅ |
| Terms of service | Required | Required |
| Privacy policy | Required | Required |
| Error messages in API | ❌ English only | ✅ |

### 11.2 Arabic Content Strategy

Arabic SEO is underserved in the UAE real estate tech space. Huge opportunity:
- Write blog posts in Modern Standard Arabic (MSA) — accessible to all Arabic speakers
- Target keywords: "برنامج إدارة العقارات الإمارات", "CRM عقارات دبي", "تتبع الشيكات المؤجلة"
- Gulf-dialect social posts for Instagram/TikTok (Dubai agents respond better to Gulf Arabic than MSA in casual contexts)

---

## 12. What's Missing — Summary Checklist

### Must have before first paid customer

- [ ] Marketing website with pricing page in AED
- [ ] Privacy policy (Arabic + English, PDPL-compliant)
- [ ] Terms of service
- [ ] UAE VAT registration (TRN)
- [ ] Stripe configured with live keys + AED pricing
- [ ] Apple Pay via Stripe (UAE users expect it)
- [ ] Help center with 8–10 core articles (Arabic + English)
- [ ] Demo environment with UAE seed data
- [ ] Status page (uptime monitoring)
- [ ] Support WhatsApp number
- [ ] In-app chat widget (Crisp recommended)
- [ ] Welcome email sequence (Arabic + English)
- [ ] Data hosted in UAE region (AWS me-central-1)
- [ ] SSL + WAF + HSTS in production
- [ ] `marketing_consent` field at signup (PDPL)
- [ ] "Delete my data" flow for contacts/tenants

### Must have within 60 days of launch

- [ ] Beta agency testimonials + case study
- [ ] 3-minute demo video (Arabic voiceover)
- [ ] Referral program
- [ ] PostHog or Mixpanel analytics installed
- [ ] Trial email drip sequence (10 emails, 90 days)
- [ ] Onboarding call booking page (Calendly)
- [ ] AED annual plan pricing in Stripe
- [ ] Backups automated + tested

### Should have within 90 days

- [ ] Telr or PayTabs as secondary payment option
- [ ] Blog with 4–6 Arabic articles live
- [ ] LinkedIn company page
- [ ] Instagram account with first 10 posts
- [ ] DREI / Cityscape presence or sponsorship planned
- [ ] 3 partnership conversations started (Bayut, DREI, a UAE law firm)
- [ ] Penetration test or OWASP scan completed

---

## 13. Open Questions (Business)

| Question | Why it matters | Recommended default |
|----------|---------------|---------------------|
| What legal entity issues invoices to customers? | VAT registration, contracts | Register UAE Free Zone company (DMCC, DIFC) — straightforward for SaaS |
| Sole founder or team? | Support capacity limits how fast you can grow | Cap at 20 customers until support is handled |
| Self-hosted vs. cloud as primary offer | Pricing, support, compliance | Cloud-first; open source is for marketing/trust |
| AED or USD pricing? | AED feels local, USD simplifies Stripe setup | Use AED on marketing site, Stripe charges in AED |
| Refund policy? | Reduces signup friction | 7-day refund on first payment, no refund after |
| Free tier quotas for cloud users? | Community plan currently shows 0 for everything | Consider 50 contacts + 1 property as a generous free tier to drive signups |
| Is there a free tier for cloud or only self-hosted? | Affects top-of-funnel | Self-hosted = free; cloud = 90-day trial then paid |
| Hiring plan? | You can't do sales + support + dev alone forever | First hire: Arabic-speaking customer success (month 3 if revenue allows) |
