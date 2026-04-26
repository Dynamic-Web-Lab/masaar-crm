# Masaar CRM Workflow Diagrams & Flowcharts

**Visual documentation for key business processes and system workflows.**

---

## Table of Contents

1. [Payment Processing Workflow](#payment-processing-workflow)
2. [Lease Management Workflow](#lease-management-workflow)
3. [Inspection & Maintenance Workflow](#inspection--maintenance-workflow)
4. [Expense Approval Workflow](#expense-approval-workflow)
5. [Commission Calculation Workflow](#commission-calculation-workflow)
6. [Lease Renewal Workflow](#lease-renewal-workflow)
7. [Database Schema Relationships](#database-schema-relationships)
8. [System Architecture](#system-architecture)

---

## Payment Processing Workflow

```
┌─────────────────────────────────────────────────────────────────────────┐
│ PAYMENT COLLECTION & RENT REMINDER SYSTEM                               │
└─────────────────────────────────────────────────────────────────────────┘

Tenant Lease Agreement
        ↓
   ┌────────────────────┐
   │ Lease Active?      │ NO → END
   └────┬───────────────┘
        │ YES
        ↓
   ┌────────────────────┐
   │ Rent Due Date      │
   │ Approaches         │
   └────┬───────────────┘
        ↓
   ┌────────────────────────────────────────┐
   │ Payment Reminder Service Triggered     │ (Every 1 hour)
   │ - Check all leases for overdue rents   │
   │ - Calculate days overdue               │
   │ - Apply late fee if configured         │
   └────┬───────────────────────────────────┘
        ↓
   ┌────────────────────────────────────────┐
   │ Determine Delivery Channel             │
   │ - Email                                │
   │ - WhatsApp                             │
   │ - SMS (if configured)                  │
   │ - In-App Notification                  │
   └────┬───────────────────────────────────┘
        ↓
   ┌────────────────────────────────────────┐
   │ Send Reminder to Tenant                │
   │ Template: "Rent of AED X due by DATE"  │
   └────┬───────────────────────────────────┘
        ↓
   ┌────────────────────────────────────────┐
   │ PAYMENT RECEIVED?                      │
   └────┬──────────┬────────────────────────┘
        │ YES      │ NO
        ↓          ↓
    ┌───────┐  ┌──────────────────┐
    │ Mark  │  │ Days Overdue > 30?│
    │ PAID  │  └──────┬───────┬───┘
    │       │         │ YES   │ NO
    └───────┘         ↓       ↓
              ┌─────────────┐  │
              │ EVICTION    │  │
              │ NOTICE      │  │
              └─────────────┘  │
                                ↓
                         ┌─────────────┐
                         │ Send Final   │
                         │ Reminder     │
                         └─────────────┘

═══════════════════════════════════════════════════════════════════════════

Database Flow:
Lease → Payment Record (created on rent due date)
     ↓
Payment Reminder Record (created when payment_status = 'pending')
     ↓
Payment Confirmation (created when payment_status = 'paid')
     ↓
Analytics Updated (occupancy_rate, collection_rate, revenue)
```

---

## Lease Management Workflow

```
┌─────────────────────────────────────────────────────────────────────────┐
│ LEASE LIFECYCLE MANAGEMENT                                              │
└─────────────────────────────────────────────────────────────────────────┘

                    LEASE INITIATION
                          ↓
        ┌─────────────────────────────────────┐
        │ Create Lease:                       │
        │ - Tenant → Property                 │
        │ - Rent amount                       │
        │ - Start date                        │
        │ - End date                          │
        │ - Terms (security deposit, etc.)    │
        └────────────┬────────────────────────┘
                     ↓
        ┌─────────────────────────────────────┐
        │ Generate Lease Document             │
        │ - From template                     │
        │ - Fill placeholders                 │
        │ - Generate PDF                      │
        │ - Request e-signature               │
        └────────────┬────────────────────────┘
                     ↓
        ┌─────────────────────────────────────┐
        │ SIGNATURE WORKFLOW (Phase 16.3)     │
        │ - Tenant receives document link     │
        │ - Tenant signs digitally            │
        │ - System records timestamp, IP      │
        │ - Both parties notified             │
        └────────────┬────────────────────────┘
                     ↓
        ┌─────────────────────────────────────┐
        │ Lease ACTIVE                        │
        │ - Monthly rent due date set         │
        │ - Payment reminders auto-send       │
        │ - Inspections scheduled             │
        │ - Maintenance tasks tracked         │
        └────────────┬────────────────────────┘
                     ↓
                  (Time passes...)
                     ↓
        ┌─────────────────────────────────────┐
        │ 90 DAYS BEFORE EXPIRY               │ ← Auto-triggered
        │ Lease Renewal Workflow Initiated    │
        └────────────┬────────────────────────┘
                     │
            ┌────────┴──────────────────────────────┐
            │ RENEWAL DECISION MADE                 │
            └────────┬──────────────┬──────┬────────┘
                 RENEW          TERMINATE  SELL
                     │               │       │
           ┌─────────▼──────┐  MOVE  │   FOR SALE
           │ NEW LEASE      │  OUT   │   NOTICE
           │ Similar terms  │   └─►  │
           │ Updated rent   │        │
           │ Signature loop │    PROPERTY VACANT
           └────────────────┘

═══════════════════════════════════════════════════════════════════════════

Key Tables Involved:
- leases: Core lease record
- payments: Monthly rent tracking
- lease_renewal_workflows: Renewal process
- documents: Signed lease document
- inspections: Condition at start, during, end
- maintenance_tasks: Any repairs during lease
```

---

## Inspection & Maintenance Workflow

```
┌─────────────────────────────────────────────────────────────────────────┐
│ PROPERTY INSPECTION & MAINTENANCE CYCLE                                 │
└─────────────────────────────────────────────────────────────────────────┘

TRIGGER: Pre-lease / Periodic / Post-occupancy
        ↓
┌──────────────────────────────┐
│ Schedule Inspection          │
│ - Select property            │
│ - Select inspection type     │
│ - Assign inspector           │
│ - Set date/time              │
└────────┬─────────────────────┘
         ↓
┌──────────────────────────────┐
│ Inspector Arrives at Site    │
│ Status → IN_PROGRESS         │
└────────┬─────────────────────┘
         ↓
   ┌─────────────────────────────────────────┐
   │ Go Through Checklist Items              │
   │ Example:                                │
   │ ☐ Walls (Pass/Fail)                    │
   │ ☐ Flooring                             │
   │ ☐ Plumbing (faucets, fixtures)         │
   │ ☐ Electrical (outlets, switches)       │
   │ ☐ HVAC (cooling/heating)               │
   │ ☐ Windows/Doors                        │
   │ ☐ Appliances                           │
   │ ☐ Gas connections                      │
   └────────┬────────────────────────────────┘
            ↓
   ┌─────────────────────────────────────────┐
   │ Document Findings                       │
   │ - Photos of damage                      │
   │ - Detailed notes                        │
   │ - Severity: Green/Yellow/Red            │
   │ - Estimated repair costs                │
   └────────┬────────────────────────────────┘
            ↓
   ┌─────────────────────────────────────────┐
   │ Mark Inspection COMPLETE                │
   │ Status → COMPLETED                      │
   └────────┬────────────────────────────────┘
            ↓
   ┌──────────────────────────────┐
   │ SEVERITY ASSESSMENT          │
   └───────┬──────┬──────┬────────┘
       GREEN   YELLOW   RED
         │        │       │
         │        │       ↓
         │        │    URGENT:
         │        │    Critical
         │        │    Issues
         │        │    ↓
         │        │ Create HIGH
         │        │ Priority
         │        │ Tasks
         │        │
         │        ↓
         │     MEDIUM:
         │     Minor issues
         │     ↓
         │     Create MEDIUM
         │     Priority tasks
         │
         ↓
    NO ISSUES:
    Archive
    inspection

         ↓
   ┌────────────────────────────────────────┐
   │ AUTO-CREATE MAINTENANCE TASKS          │
   │ From inspection findings                │
   │ - Task type: Plumbing/Electrical/etc   │
   │ - Priority: Based on severity          │
   │ - Assigned to: Contractor/Staff        │
   │ - Due date: Based on priority          │
   │ - Estimated cost: From inspection      │
   └────────┬───────────────────────────────┘
            ↓
   ┌────────────────────────────────────────┐
   │ MAINTENANCE TASK EXECUTION             │
   │ 1. Contractor receives assignment      │
   │ 2. Uploads BEFORE photos               │
   │ 3. Performs work                       │
   │ 4. Uploads DURING photos               │
   │ 5. Uploads AFTER photos                │
   │ 6. Records actual cost                 │
   │ 7. Marks COMPLETED                     │
   └────────┬───────────────────────────────┘
            ↓
   ┌────────────────────────────────────────┐
   │ Task Completion                        │
   │ - Status: COMPLETED                    │
   │ - Actual cost recorded                 │
   │ - Photos archived                      │
   │ - Analytics updated                    │
   └────────────────────────────────────────┘

═══════════════════════════════════════════════════════════════════════════

Database Schema:
inspection_templates
    ↓
inspections (created from template)
    ↓
maintenance_tasks (auto-created from inspection findings)
    ↓
maintenance_photos (before, during, after)
    ↓
expenses (actual costs recorded)
```

---

## Expense Approval Workflow

```
┌─────────────────────────────────────────────────────────────────────────┐
│ EXPENSE TRACKING & APPROVAL WORKFLOW                                    │
└─────────────────────────────────────────────────────────────────────────┘

USER CREATES EXPENSE:
        ↓
┌──────────────────────────────────┐
│ New Expense Entry:               │
│ - Category (maintenance, etc.)   │
│ - Amount (AED)                   │
│ - Date incurred                  │
│ - Vendor name                    │
│ - Payment method                 │
│ - Receipt (upload file)          │
│ - Notes                          │
└────────┬─────────────────────────┘
         ↓
    ┌─────────────────┐
    │ Status: PENDING │
    └────────┬────────┘
             ↓
    ┌────────────────────────────────┐
    │ ADMIN NOTIFICATION SENT        │
    │ Real-time via WebSocket        │
    │ Email: "New expense pending"   │
    └────────┬───────────────────────┘
             ↓
    ┌────────────────────────────────┐
    │ ADMIN REVIEWS EXPENSE          │
    │ - Check receipt                │
    │ - Verify amount                │
    │ - Review vendor                │
    │ - Check category               │
    │ - Read notes                   │
    └────────┬──────────┬────────────┘
             │          │
          APPROVE    REJECT
             │          │
             ↓          ↓
    ┌──────────────┐ ┌─────────────────┐
    │ Status:      │ │ Status: REJECTED│
    │ APPROVED     │ │ Reason sent to  │
    │              │ │ requester       │
    │ ↓            │ │ ↓               │
    │ Payment      │ │ Can edit and    │
    │ Processed    │ │ resubmit        │
    │ (optional    │ │                 │
    │  bank)       │ │ ↓               │
    │              │ │ Resubmit        │
    │ ↓            │ │ Expense         │
    │ Status: PAID │ │ (loop back)     │
    └──────────────┘ └─────────────────┘

ANALYTICS IMPACT:
   ↓
Expense added to:
- Company expense report
- Property expense breakdown
- Monthly P&L
- Analytics dashboard

═══════════════════════════════════════════════════════════════════════════

Database Flow:
expense_categories (predefined)
    ↓
expenses (user creates)
    ↓
expense_approvals (admin decision)
    ↓
Analytics updated
    ↓
Reports generated
```

---

## Commission Calculation Workflow

```
┌─────────────────────────────────────────────────────────────────────────┐
│ AGENT COMMISSION CALCULATION & PAYOUT SYSTEM                            │
└─────────────────────────────────────────────────────────────────────────┘

DURING MONTH: Deals closed
        ↓
   ┌────────────────────────┐
   │ Deal Status: WON       │
   │ - Agent assigned       │
   │ - Deal value recorded  │
   │ - Close date set       │
   └────────┬───────────────┘
            ↓
        (Day 1 of next month)
            ↓
   ┌────────────────────────────────────┐
   │ AUTOMATED JOB RUNS:                │
   │ "commission_calculation_job"       │
   │ Runs at: 00:00 UTC (midnight)      │
   └────────┬───────────────────────────┘
            ↓
   ┌────────────────────────────────────┐
   │ For each agent:                    │
   │ 1. Query deals closed prev month   │
   │ 2. Query leases signed prev month  │
   │ 3. Load commission_structure       │
   │ 4. Calculate commission:           │
   └────────┬───────────────────────────┘
            ↓
   ┌────────────────────────────────────┐
   │ COMMISSION CALCULATION EXAMPLE:    │
   │                                    │
   │ Commission Structure: 5% of deals  │
   │ Agent Sales:                       │
   │ - Deal A: AED 100,000 × 5% = 5000 │
   │ - Deal B: AED 50,000 × 5%  = 2500 │
   │ - Lease A: AED 30,000 × 5% = 1500 │
   │                                    │
   │ Total Commission: AED 9,000        │
   └────────┬───────────────────────────┘
            ↓
   ┌────────────────────────────────────┐
   │ Create AgentCommission record:     │
   │ - agent_id                         │
   │ - period_start: 1st prev month     │
   │ - period_end: last day prev month  │
   │ - total_commission: 9,000          │
   │ - status: PENDING (awaiting admin) │
   │ - created_at: now                  │
   └────────┬───────────────────────────┘
            ↓
   ┌────────────────────────────────────┐
   │ ADMIN NOTIFICATION SENT            │
   │ "Commission for [Agent] pending"   │
   └────────┬───────────────────────────┘
            ↓
   ┌────────────────────────────────────┐
   │ ADMIN REVIEW & APPROVAL            │
   │ - Check calculation accuracy       │
   │ - Review deals included            │
   │ - Look for disputes                │
   │ - Approve or dispute               │
   └────────┬──────────┬────────────────┘
             │          │
          APPROVE    DISPUTE
             │          │
             ↓          ↓
    ┌─────────────────┐  │
    │ Status:         │  │
    │ APPROVED        │  │
    │ ↓               │  │
    │ Ready to pay    │  │
    │                 │  │
    │ ↓               │  │
    │ Click:          │  │
    │ "Mark as Paid"  │  │
    │ ↓               │  │
    │ Enter:          │  │
    │ - Bank ref #    │  │
    │ - Payment date  │  │
    │ ↓               │  │
    │ Status: PAID    │  │
    │                 │  │
    │ Agent notified  │  │
    └─────────────────┘  │
                         ↓
                  ┌─────────────────┐
                  │ Status: DISPUTED│
                  │ Admin notes     │
                  │ Agent contacted │
                  │ Dispute resolved│
                  └─────────────────┘

═══════════════════════════════════════════════════════════════════════════

Database Flow:
commission_structures (configured by admin)
    ↓
deals (closed during month)
    ↓
leases (signed during month)
    ↓
[AUTOMATED] agent_commissions (created)
    ↓
commission_transactions (individual entries)
    ↓
[ADMIN] Update status: APPROVED → PAID
    ↓
Analytics: Agent performance report generated
```

---

## Lease Renewal Workflow

```
┌─────────────────────────────────────────────────────────────────────────┐
│ AUTOMATED LEASE RENEWAL WORKFLOW (Phase 16.1)                           │
└─────────────────────────────────────────────────────────────────────────┘

TRIGGER: 90 days before lease expiry
        ↓
   ┌──────────────────────────────────┐
   │ DAILY BATCH JOB CHECKS:          │
   │ For all leases expiring in 90d:  │
   │ - Create LeaseRenewalWorkflow    │
   │ - Status: PENDING                │
   │ - Send initial reminder          │
   └────────┬─────────────────────────┘
            ↓
   ┌────────────────────────────────────┐
   │ ADMIN PROPOSES RENEWAL TERMS       │
   │ - Proposed rent amount (or same)   │
   │ - Contract duration                │
   │ - New terms (if any)               │
   │ - Status → PENDING_OFFER           │
   └────────┬───────────────────────────┘
            ↓
   ┌────────────────────────────────────┐
   │ SEND OFFER TO TENANT:              │
   │ - Select communication template    │
   │ - Customize message                │
   │ - Choose channel (email/WhatsApp)  │
   │ - Status → OFFER_SENT              │
   │ - Log communication in audit trail │
   └────────┬───────────────────────────┘
            ↓
        (Tenant reviews)
            ↓
   ┌──────────────────────────────────────────────┐
   │ TENANT RESPONDS: 3 PATHS                     │
   └──────┬──────────────┬───────────┬────────────┘
      ACCEPT         COUNTER-OFFER    REJECT
         │                 │            │
         ↓                 ↓            ↓
    ┌────────────┐  ┌────────────┐  ┌────────────┐
    │ ACCEPTED:  │  │ COUNTER:   │  │ REJECTED:  │
    │ Tenant OK  │  │ Tenant     │  │ Tenant     │
    │ with terms │  │ proposes   │  │ won't      │
    │            │  │ different  │  │ renew      │
    │ ↓          │  │ rent       │  │            │
    │ Update     │  │            │  │ ↓          │
    │ Lease:     │  │ ↓          │  │ Lease      │
    │ - New end  │  │ Create     │  │ expires    │
    │   date     │  │ counter-   │  │ Property   │
    │ - New rent │  │ offer      │  │ vacates    │
    │ - Sign new │  │ record     │  │            │
    │   document │  │            │  │ OR         │
    │            │  │ Admin      │  │            │
    │ ↓          │  │ decides:   │  │ ↓          │
    │ Document   │  │ Accept or  │  │ New tenant │
    │ generated  │  │ counter    │  │ found      │
    │            │  │ back       │  │ Different  │
    │ ↓          │  │            │  │ terms      │
    │ E-sign     │  │ Loop until │  │ negotiated │
    │ workflow   │  │ agreement  │  └────────────┘
    │            │  │ reached    │
    │ ↓          │  │            │
    │ Status:    │  │ ↓          │
    │ ACCEPTED   │  │ Status:    │
    │            │  │ ACCEPTED   │
    │ New lease  │  │            │
    │ active     │  │ New lease  │
    │ Payment    │  │ active     │
    │ reminders  │  │            │
    │ restart    │  │ Cycle back │
    │            │  │ to step 1  │
    │ ↓          │  │            │
    │ Sent admin │  └────────────┘
    │ notification
    └────────────┘

═══════════════════════════════════════════════════════════════════════════

Database Schema:
lease_renewal_workflows (main process)
    ↓
renewal_communication_templates (predefined messages)
    ↓
renewal_communication_log (audit trail)
    ↓
documents (new lease document)
    ↓
document_signatures (e-signatures)
    ↓
leases (updated with new end date)
```

---

## Database Schema Relationships

```
┌─────────────────────────────────────────────────────────────────────────┐
│ CORE ENTITY RELATIONSHIPS                                               │
└─────────────────────────────────────────────────────────────────────────┘

              ┌──────────────┐
              │   companies  │ (multi-tenant)
              └──────┬───────┘
                     │
      ┌──────────────┼──────────────┬──────────────┐
      │              │              │              │
      ↓              ↓              ↓              ↓
 ┌────────────┐ ┌─────────┐ ┌──────────────┐ ┌────────┐
 │users       │ │tenants  │ │rental_        │ │contacts│
 │(Admin/     │ │         │ │properties    │ │        │
 │Agent/      │ └────┬────┘ └──────┬───────┘ └───┬────┘
 │Viewer)     │      │             │             │
 └────────────┘      │             │             │
      │              │        ┌────┴────┐        │
      └──────────────┼───────→│leases   │←───────┘
                     │        └────┬────┘
                     │             │
         ┌───────────┴─────────────┼───────────┐
         │                         │           │
         ↓                         ↓           ↓
    ┌──────────┐            ┌──────────┐ ┌─────────┐
    │inspections           │payments  │ │lease_   │
    │           │ (tracks) │          │ │renewal_ │
    │           └─→────────→├──────────┤ │workflows│
    └───────┬───┘           └────┬─────┘ └────┬────┘
            │                    │            │
            ↓                    ↓            ↓
    ┌──────────────┐        ┌──────────┐ ┌──────────┐
    │maintenance_  │        │payment_  │ │documents │
    │tasks         │        │reminders │ │          │
    │              │        │          │ │          │
    └───────┬──────┘        └──────────┘ └───┬──────┘
            │                                │
            ↓                                ↓
    ┌──────────────┐                    ┌─────────────┐
    │maintenance_  │                    │document_    │
    │photos        │                    │signatures   │
    └──────────────┘                    └─────────────┘

OPERATIONAL TABLES (Finance):
    ┌──────────────┬──────────────┬─────────────┬──────────────┐
    ↓              ↓              ↓             ↓              ↓
expenses    agent_         commission_   expense_       bank_
            commissions     structures    approvals     statements
            ├→ commission_
            │  transactions
            │
            └→ payment_confirmations

REFERENCE TABLES (Configuration):
    ┌────────────────┬─────────────┬────────────────┐
    ↓                ↓             ↓                ↓
inspection_      renewal_        document_      custom_
templates        templates       templates      fields
```

---

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│ MASAAR CRM SYSTEM ARCHITECTURE (Microservices-ready)                   │
└─────────────────────────────────────────────────────────────────────────┘

                        ┌──────────────────────┐
                        │   Browser / Mobile   │
                        │   (Next.js Frontend) │
                        └──────────┬───────────┘
                                   │
                ┌──────────────────┼──────────────────┐
                │                  │                  │
                │ HTTP/HTTPS       │ WebSocket        │
                │                  │                  │
                ↓                  ↓                  ↓
         ┌────────────┐   ┌──────────────┐   ┌─────────────┐
         │API Routes  │   │WebSocket Hub │   │CORS/Auth    │
         │(REST)      │   │(Real-time)   │   │Middleware   │
         └─────┬──────┘   └────────┬─────┘   └──────┬──────┘
               │                   │                │
               └───────────────────┼────────────────┘
                                   │
                    ┌──────────────────────────────┐
                    │  Go Fiber Web Framework      │
                    │  (cmd/server/main.go)        │
                    └──────────────┬───────────────┘
                                   │
          ┌────────────────────────┼────────────────────────┐
          │                        │                        │
          ↓                        ↓                        ↓
    ┌───────────────┐       ┌──────────────┐      ┌──────────────┐
    │API Handlers   │       │Background    │      │WebSocket     │
    │/api/v1/*      │       │Jobs          │      │Handler       │
    │               │       │(Cron tasks)  │      │              │
    │- User         │       │              │      │- Send push   │
    │- Contact      │       │- Payment     │      │  notif       │
    │- Lead         │       │  reminder    │      │- Broadcast   │
    │- Deal         │       │- Commission  │      │  events      │
    │- Invoice      │       │  calc        │      │- Real-time   │
    │- Payment      │       │- Renewal     │      │  updates     │
    │- Inspection   │       │  reminder    │      │              │
    │- Expense      │       │- Analytics   │      └──────────────┘
    │- Lease        │       │  refresh     │
    │- Commission   │       │              │
    │- Document     │       └──────────────┘
    │- Analytics    │
    │- Audit Log    │
    └────────┬──────┘
             │
       ┌─────┴──────────────────────────────────┐
       │                                        │
       ↓                                        ↓
┌────────────────┐                      ┌──────────────┐
│Repository Layer│                      │Service Layer │
│(Data Access)   │                      │(Business     │
│                │                      │Logic)        │
│- UserRepo      │                      │              │
│- ContactRepo   │                      │- AI Service  │
│- LeadRepo      │                      │  (Ollama)    │
│- DealRepo      │                      │- Email       │
│- PaymentRepo   │                      │  Service     │
│- etc.          │                      │- WhatsApp    │
│                │                      │  Service     │
│(pgx pool,      │                      │- PDF         │
│prepared stmts) │                      │  Generation  │
└────────┬───────┘                      └──────────────┘
         │
         └────────────────────┬──────────────────────┐
                              │                      │
                              ↓                      ↓
                      ┌────────────────┐    ┌──────────────┐
                      │PostgreSQL 16   │    │Redis 7       │
                      │Database        │    │Cache/        │
                      │                │    │Sessions      │
                      │- All user data │    │              │
                      │- Audit logs    │    │- JWT tokens  │
                      │- Documents     │    │- Blacklist   │
                      │- Transactions  │    │- Rate limits │
                      │- 30 migrations │    │- Locks       │
                      │                │    │              │
                      │pgvector ext.   │    └──────────────┘
                      │for AI search   │
                      └────────────────┘

EXTERNAL SERVICES (Optional):
    ├→ Meta WhatsApp API (message send/receive)
    ├→ Ollama (Local LLM for AI features)
    ├→ SMTP (Email delivery)
    ├→ S3-compatible storage (Documents/photos)
    └→ DocuSign (e-signature, Phase 16.3+)

DEPLOYMENT:
    Docker Compose (Development):
    ├→ Server (Go Fiber)
    ├→ PostgreSQL
    ├→ Redis
    ├→ Ollama (optional)
    └→ Nginx (reverse proxy)

    Kubernetes (Production):
    ├→ Server pods (horizontal scaling)
    ├→ StatefulSet: PostgreSQL
    ├→ StatefulSet: Redis
    ├→ ConfigMaps: Migrations
    ├→ Secrets: Credentials
    └→ Ingress: API routing
```

---

## Request/Response Flow Example

```
USER ACTION: Agent creates new lead

1. FRONTEND (Next.js)
   ┌──────────────────────────────┐
   │ User fills lead form         │
   │ - Contact dropdown           │
   │ - Deal value input           │
   │ - Stage select               │
   │ - Tags multiselect           │
   │ Click: "Create Lead"         │
   └────────┬─────────────────────┘
            │
2. API CLIENT (fetch)
   ┌──────────────────────────────┐
   │ Construct request:           │
   │ POST /api/v1/leads           │
   │ Headers:                     │
   │ - Authorization: Bearer JWT  │
   │ - Content-Type: application/ │
   │   json                       │
   │ Body: {                      │
   │   "contact_id": "uuid",      │
   │   "deal_value": 50000,       │
   │   "currency": "AED",         │
   │   "stage": "new",            │
   │   "tags": ["interested"]     │
   │ }                            │
   └────────┬─────────────────────┘
            │
3. MIDDLEWARE (Go Fiber)
   ┌──────────────────────────────┐
   │ middleware.JWT()             │
   │ - Validate token signature   │
   │ - Check token expiry         │
   │ - Extract user_id            │
   │ - Extract company_id         │
   │                              │
   │ middleware.CheckBlacklist()  │
   │ - Check if token blacklisted │
   │   (user logged out)          │
   │                              │
   │ middleware.RequireRole()     │
   │ - Check user role allowed    │
   │ - Only Agent+                │
   └────────┬─────────────────────┘
            │
4. HANDLER (LeadHandler)
   ┌──────────────────────────────┐
   │ CreateLead(c *fiber.Ctx)     │
   │ - Parse request body         │
   │ - Validate input:            │
   │   - contact_id exists?       │
   │   - deal_value > 0?          │
   │   - currency valid?          │
   │ - Create Lead struct         │
   │ - Set defaults: lead_score=0 │
   │ - Call repository            │
   └────────┬─────────────────────┘
            │
5. REPOSITORY (LeadRepo)
   ┌──────────────────────────────┐
   │ Create(ctx, lead)            │
   │ - Prepare SQL:               │
   │   INSERT INTO leads (...)    │
   │   VALUES ($1, $2, ...)       │
   │ - Execute with pgx pool      │
   │ - Return created_at,         │
   │   updated_at                 │
   │ - Handle DB errors           │
   └────────┬─────────────────────┘
            │
6. DATABASE (PostgreSQL)
   ┌──────────────────────────────┐
   │ Execute SQL INSERT           │
   │ - Validate constraints       │
   │ - Generate IDs               │
   │ - Set timestamps             │
   │ - Return new record          │
   │ - Create audit log entry     │
   └────────┬─────────────────────┘
            │
7. HANDLER → Response
   ┌──────────────────────────────┐
   │ Return JSON:                 │
   │ {                            │
   │   "data": {                  │
   │     "id": "uuid",            │
   │     "contact_id": "uuid",    │
   │     "deal_value": 50000,     │
   │     "stage": "new",          │
   │     "lead_score": 0,         │
   │     "created_at": "2026-..." │
   │   }                          │
   │ }                            │
   │ HTTP 201 Created             │
   └────────┬─────────────────────┘
            │
8. FRONTEND (React)
   ┌──────────────────────────────┐
   │ Parse response               │
   │ Update state with new lead   │
   │ Navigate to lead detail      │
   │ Show success toast           │
   │ Trigger WebSocket update     │
   │ (notify other users)         │
   └──────────────────────────────┘

9. WEBSOCKET (Real-time)
   ┌──────────────────────────────┐
   │ Handler broadcasts:          │
   │ {                            │
   │   "type": "lead_created",    │
   │   "data": {...}              │
   │ }                            │
   │ To: All connected users      │
   │ From: Company agent          │
   └──────────────────────────────┘

10. OTHER USERS
    ┌──────────────────────────────┐
    │ Receive WebSocket message    │
    │ Check if relevant to them    │
    │ Update UI in real-time       │
    │ Show notification            │
    └──────────────────────────────┘
```

---

**Document Version**: 0.2.0
**Created**: April 2026
**For**: Dori Internet Development Team
**Usage**: Architecture planning, developer onboarding, process documentation
