# Phase 15-16: Strategic Roadmap - Masaar CRM Enhancement

**Status:** Planning Phase
**Target:** Complete core feature gaps and prepare for Phase 17 (SaaS multi-tenant platform)
**Estimated Timeline:** Phase 15 (4 weeks), Phase 16 (4 weeks)

---

## Executive Summary

Masaar CRM has successfully implemented 40+ domain entities, 60+ API endpoints, and complete infrastructure (Docker, Kubernetes). Phase 15-16 will fill critical feature gaps identified in the real estate CRM space:

**Priority Gaps:**
1. **Expense Tracking** - Currently no operational expense logging (only payment revenue)
2. **Inspection & Maintenance Scheduling** - Properties need scheduled inspections and maintenance tasks
3. **Tenant Analytics Dashboard** - Missing portfolio-level analytics for property performance
4. **Lease Renewal Automation** - No automated renewal workflows
5. **Commission Tracking** - Missing agent commission calculations on deals/rentals
6. **Document Management** - Limited document handling (only invoice PDFs)

---

## Phase 15: Core Operations & Analytics (Weeks 1-4)

### 15.1: Expense Tracking System

**Goal:** Complete financial picture for all operational and property expenses.

**Schema Design:**
```sql
-- Migration: 00024_create_expense_management.sql
CREATE TABLE expense_categories (
    id UUID PRIMARY KEY,
    company_id UUID NOT NULL,
    category_name VARCHAR(100) NOT NULL,
    category_type ENUM('property_maintenance', 'utilities', 'insurance', 'cleaning', 'repairs', 'staff', 'other'),
    description TEXT,
    created_at TIMESTAMP,
    UNIQUE(company_id, category_name)
);

CREATE TABLE expenses (
    id UUID PRIMARY KEY,
    company_id UUID NOT NULL,
    category_id UUID NOT NULL REFERENCES expense_categories(id),
    property_id UUID REFERENCES rental_properties(id),
    tenant_id UUID REFERENCES tenants(id),
    amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'AED',
    expense_date DATE NOT NULL,
    description TEXT,
    vendor_name VARCHAR(150),
    vendor_contact VARCHAR(255),
    payment_method ENUM('cash', 'bank_transfer', 'credit_card', 'check', 'other'),
    payment_status ENUM('pending', 'paid', 'refunded'),
    receipt_url VARCHAR(500),
    notes TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE expense_approvals (
    id UUID PRIMARY KEY,
    expense_id UUID NOT NULL UNIQUE REFERENCES expenses(id),
    approval_status ENUM('pending', 'approved', 'rejected'),
    approved_by UUID REFERENCES users(id),
    approval_comments TEXT,
    approval_date TIMESTAMP
);

CREATE INDEX idx_expenses_company_property ON expenses(company_id, property_id);
CREATE INDEX idx_expenses_date_range ON expenses(expense_date);
```

**Domain Models:**
```go
type ExpenseCategory struct {
    ID            uuid.UUID
    CompanyID     uuid.UUID
    CategoryName  string
    CategoryType  ExpenseCategoryType // enum
    Description   string
}

type Expense struct {
    ID            uuid.UUID
    CompanyID     uuid.UUID
    CategoryID    uuid.UUID
    PropertyID    *uuid.UUID
    TenantID      *uuid.UUID
    Amount        decimal.Decimal
    Currency      string
    ExpenseDate   time.Time
    Description   string
    VendorName    string
    VendorContact string
    PaymentMethod ExpensePaymentMethod // enum
    PaymentStatus ExpensePaymentStatus // enum
    ReceiptURL    string
    Notes         string
    CreatedBy     uuid.UUID
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DeletedAt     *time.Time
}

type ExpenseApproval struct {
    ID               uuid.UUID
    ExpenseID        uuid.UUID
    ApprovalStatus   ExpenseApprovalStatus // enum
    ApprovedBy       *uuid.UUID
    ApprovalComments string
    ApprovalDate     *time.Time
}
```

**API Endpoints:**
```
POST   /api/v1/expense-categories          (Admin only)
GET    /api/v1/expense-categories          (with filtering)
PUT    /api/v1/expense-categories/{id}     (Admin only)

POST   /api/v1/expenses                    (Agent+)
GET    /api/v1/expenses                    (filtered by company)
GET    /api/v1/expenses/{id}               (Agent+)
PUT    /api/v1/expenses/{id}               (Agent+, if pending)
DELETE /api/v1/expenses/{id}               (soft-delete, if pending)
POST   /api/v1/expenses/bulk-import        (Admin, CSV import)

POST   /api/v1/expenses/{id}/approve       (Admin only)
POST   /api/v1/expenses/{id}/reject        (Admin only)
```

**Implementation Tasks:**
- [ ] Create migration 00024 with expense tables
- [ ] Add domain models (ExpenseCategory, Expense, ExpenseApproval, enums)
- [ ] Create repositories: ExpenseCategoryRepo, ExpenseRepo, ExpenseApprovalRepo
- [ ] Create handler: ExpenseHandler with CRUD + bulk import
- [ ] Add routes to router.go
- [ ] Frontend: Create /expenses page with category filter, date range, amount totals, vendor search
- [ ] Frontend: Add expense creation modal with receipt upload
- [ ] Frontend: Approval workflow UI (Admin only)
- [ ] Add real-time notification on expense approval/rejection

---

### 15.2: Inspection & Maintenance Scheduling

**Goal:** Manage property inspections and maintenance tasks with automated reminders.

**Schema Design:**
```sql
-- Migration: 00025_create_inspection_and_maintenance.sql
CREATE TABLE inspection_templates (
    id UUID PRIMARY KEY,
    company_id UUID NOT NULL,
    template_name VARCHAR(150) NOT NULL,
    inspection_type ENUM('general', 'pre_lease', 'end_lease', 'damage_assessment', 'safety'),
    checklist_items JSONB, -- Array of {id, description, critical}
    estimated_duration_minutes INT,
    created_at TIMESTAMP,
    UNIQUE(company_id, template_name)
);

CREATE TABLE inspections (
    id UUID PRIMARY KEY,
    company_id UUID NOT NULL,
    property_id UUID NOT NULL REFERENCES rental_properties(id),
    template_id UUID REFERENCES inspection_templates(id),
    inspection_type VARCHAR(50),
    scheduled_date TIMESTAMP NOT NULL,
    completed_date TIMESTAMP,
    inspector_id UUID REFERENCES users(id),
    tenant_id UUID REFERENCES tenants(id),
    status ENUM('scheduled', 'in_progress', 'completed', 'cancelled'),
    findings TEXT,
    severity_level ENUM('green', 'yellow', 'red'),
    photos_urls TEXT[], -- Array of photo URLs
    checklist_results JSONB, -- {item_id: {status: pass/fail, notes}}
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE maintenance_tasks (
    id UUID PRIMARY KEY,
    company_id UUID NOT NULL,
    property_id UUID NOT NULL REFERENCES rental_properties(id),
    inspection_id UUID REFERENCES inspections(id),
    maintenance_type ENUM('plumbing', 'electrical', 'hvac', 'flooring', 'painting', 'structural', 'other'),
    description TEXT NOT NULL,
    priority ENUM('low', 'medium', 'high', 'urgent'),
    scheduled_date DATE,
    due_date DATE,
    completion_date DATE,
    contractor_name VARCHAR(150),
    contractor_contact VARCHAR(255),
    estimated_cost DECIMAL(15,2),
    actual_cost DECIMAL(15,2),
    status ENUM('pending', 'scheduled', 'in_progress', 'completed', 'cancelled'),
    assigned_to UUID REFERENCES users(id),
    notes TEXT,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE maintenance_photos (
    id UUID PRIMARY KEY,
    task_id UUID NOT NULL REFERENCES maintenance_tasks(id),
    photo_url VARCHAR(500) NOT NULL,
    uploaded_at TIMESTAMP,
    photo_stage ENUM('before', 'during', 'after')
);

CREATE INDEX idx_inspections_property_date ON inspections(property_id, scheduled_date);
CREATE INDEX idx_maintenance_property_status ON maintenance_tasks(property_id, status);
```

**Domain Models:**
```go
type InspectionTemplate struct {
    ID                      uuid.UUID
    CompanyID               uuid.UUID
    TemplateName            string
    InspectionType          InspectionType // enum
    ChecklistItems          []ChecklistItem // JSON
    EstimatedDurationMinutes int
}

type Inspection struct {
    ID             uuid.UUID
    CompanyID      uuid.UUID
    PropertyID     uuid.UUID
    TemplateID     *uuid.UUID
    InspectionType string
    ScheduledDate  time.Time
    CompletedDate  *time.Time
    InspectorID    *uuid.UUID
    TenantID       *uuid.UUID
    Status         InspectionStatus // enum
    Findings       string
    SeverityLevel  SeverityLevel // enum
    PhotosURLs     []string
    ChecklistResults map[string]ChecklistResult
    CreatedBy      uuid.UUID
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

type MaintenanceTask struct {
    ID               uuid.UUID
    CompanyID        uuid.UUID
    PropertyID       uuid.UUID
    InspectionID     *uuid.UUID
    MaintenanceType  MaintenanceType // enum
    Description      string
    Priority         TaskPriority // enum
    ScheduledDate    *time.Time
    DueDate          *time.Time
    CompletionDate   *time.Time
    ContractorName   string
    ContractorContact string
    EstimatedCost    *decimal.Decimal
    ActualCost       *decimal.Decimal
    Status           MaintenanceStatus // enum
    AssignedTo       *uuid.UUID
    Notes            string
    CreatedBy        uuid.UUID
    CreatedAt        time.Time
    UpdatedAt        time.Time
    DeletedAt        *time.Time
}
```

**API Endpoints:**
```
-- Inspection Templates
GET    /api/v1/inspection-templates       (read-only)
POST   /api/v1/inspection-templates       (Admin only)
PUT    /api/v1/inspection-templates/{id}  (Admin only)

-- Inspections
POST   /api/v1/inspections                (Agent+)
GET    /api/v1/inspections                (filtered by company/property)
GET    /api/v1/inspections/{id}           (Agent+)
PUT    /api/v1/inspections/{id}           (Inspector, if status=in_progress)
POST   /api/v1/inspections/{id}/complete  (Inspector)
POST   /api/v1/inspections/{id}/photos    (multipart upload)

-- Maintenance Tasks
POST   /api/v1/maintenance-tasks          (Agent+)
GET    /api/v1/maintenance-tasks          (filtered by property/status)
GET    /api/v1/maintenance-tasks/{id}
PUT    /api/v1/maintenance-tasks/{id}     (Agent+, if status != completed)
POST   /api/v1/maintenance-tasks/{id}/complete
POST   /api/v1/maintenance-tasks/{id}/photos
```

**Background Jobs:**
- Inspection reminder emails (24 hours before scheduled)
- Maintenance task reminders (3 days, 1 day before due date)
- Auto-generate maintenance tasks from critical inspection findings

**Implementation Tasks:**
- [ ] Create migration 00025
- [ ] Add domain models and enums
- [ ] Create repositories: InspectionTemplateRepo, InspectionRepo, MaintenanceTaskRepo
- [ ] Create handler: InspectionHandler, MaintenanceTaskHandler
- [ ] Add background jobs for reminders
- [ ] Frontend: Create /inspections and /maintenance pages
- [ ] Frontend: Calendar view for scheduled inspections/tasks
- [ ] Frontend: Photo gallery and before/after comparison
- [ ] WebSocket notifications for inspection/task status changes

---

### 15.3: Tenant Analytics Dashboard

**Goal:** Portfolio-level insights on tenant performance, occupancy, and revenue trends.

**Analytics Queries:**
```go
// tenant_analytics.go
type TenantAnalytics struct {
    TotalTenants        int
    ActiveTenants       int
    VacantUnits         int
    OccupancyRate       float64 // percentage
    AverageRentPerUnit  decimal.Decimal
    TotalMonthlyRevenue decimal.Decimal
    CollectionRate      float64 // percentage
    OverdueDuesAmount   decimal.Decimal
    UpcomingRenewals    int // within 60 days
    TenantChurnRate     float64 // last 12 months
}

type PropertyAnalytics struct {
    PropertyID          uuid.UUID
    PropertyName        string
    TotalUnits          int
    OccupiedUnits       int
    OccupancyRate       float64
    MonthlyRevenue      decimal.Decimal
    OperatingExpenses   decimal.Decimal
    NetOperatingIncome  decimal.Decimal
    MaintenanceNeeded   int // count of pending tasks
    AverageMaintenanceResponseDays int
}

type TenantPerformanceMetrics struct {
    TenantID            uuid.UUID
    TenantName          string
    RentalHistory       int // number of properties rented
    AverageStay         float64 // days
    PaymentOnTimeRate   float64 // percentage
    DisputeCount        int
    RiskScore           int // 0-100
}
```

**Dashboard Pages:**
```
/api/v1/analytics/tenant-overview      (GET) - Company-wide analytics
/api/v1/analytics/property/{id}        (GET) - Property-specific metrics
/api/v1/analytics/tenant/{id}          (GET) - Tenant performance
/api/v1/analytics/payment-trends       (GET) - Revenue trends by month
/api/v1/analytics/occupancy-timeline   (GET) - Historical occupancy rates
/api/v1/analytics/expense-breakdown    (GET) - Expense distribution by category
```

**Frontend Analytics Pages:**
- `/analytics/overview` - Company-wide KPIs (occupancy, revenue, collection rate)
- `/analytics/properties` - Property performance table with drill-down
- `/analytics/tenants` - Tenant metrics with risk scoring
- `/analytics/financial` - Revenue vs expenses, P&L trends
- `/analytics/maintenance` - Task completion rates, response times

**Implementation Tasks:**
- [ ] Create analytics service layer
- [ ] Add analytics handlers (read-only, all roles)
- [ ] Implement caching for analytics queries (1-hour TTL in Redis)
- [ ] Frontend: Build analytics dashboard with charts (Recharts library)
- [ ] Add role-based analytics restrictions (Viewer sees all, Agent sees assigned properties only)
- [ ] Implement data export (CSV/PDF reports)

---

## Phase 16: Advanced Workflows & Documentation (Weeks 1-4)

### 16.1: Lease Renewal Automation

**Goal:** Automated lease renewal workflows with tenant communication templates.

**Schema Design:**
```sql
-- Migration: 00026_create_lease_renewals.sql
CREATE TABLE lease_renewal_workflows (
    id UUID PRIMARY KEY,
    company_id UUID NOT NULL,
    lease_id UUID NOT NULL UNIQUE REFERENCES leases(id),
    renewal_date DATE NOT NULL,
    renewal_status ENUM('pending', 'in_progress', 'offer_sent', 'accepted', 'rejected', 'expired'),
    days_before_expiry INT DEFAULT 90, -- when to start renewal process
    proposed_rent_amount DECIMAL(15,2),
    proposed_terms TEXT, -- JSON with key terms
    tenant_response ENUM('pending', 'accepted', 'rejected', 'counter_offer'),
    tenant_counter_offer DECIMAL(15,2),
    counter_offer_date TIMESTAMP,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE renewal_communication_templates (
    id UUID PRIMARY KEY,
    company_id UUID NOT NULL,
    template_name VARCHAR(150) NOT NULL,
    email_subject VARCHAR(255),
    email_body TEXT,
    whatsapp_message TEXT,
    language ENUM('ar', 'en'),
    created_at TIMESTAMP,
    UNIQUE(company_id, template_name, language)
);

CREATE TABLE renewal_communication_log (
    id UUID PRIMARY KEY,
    renewal_id UUID NOT NULL REFERENCES lease_renewal_workflows(id),
    communication_type ENUM('email', 'whatsapp', 'sms', 'in_app_notification'),
    template_id UUID REFERENCES renewal_communication_templates(id),
    sent_date TIMESTAMP,
    delivery_status ENUM('pending', 'sent', 'failed', 'read'),
    tenant_response_date TIMESTAMP,
    response_text TEXT,
    created_at TIMESTAMP
);

CREATE INDEX idx_lease_renewals_status ON lease_renewal_workflows(company_id, renewal_status);
CREATE INDEX idx_lease_renewals_date ON lease_renewal_workflows(renewal_date);
```

**Domain Models:**
```go
type LeaseRenewalWorkflow struct {
    ID                 uuid.UUID
    CompanyID          uuid.UUID
    LeaseID            uuid.UUID
    RenewalDate        time.Time
    RenewalStatus      RenewalStatus // enum
    DaysBeforeExpiry   int
    ProposedRentAmount *decimal.Decimal
    ProposedTerms      map[string]interface{} // JSON
    TenantResponse     TenantRenewalResponse // enum
    TenantCounterOffer *decimal.Decimal
    CounterOfferDate   *time.Time
    CreatedAt          time.Time
    UpdatedAt          time.Time
}

type RenewalCommunicationTemplate struct {
    ID            uuid.UUID
    CompanyID     uuid.UUID
    TemplateName  string
    EmailSubject  string
    EmailBody     string
    WhatsAppMsg   string
    Language      Language // enum
    CreatedAt     time.Time
}
```

**Background Job:**
```go
// Every day, check for leases expiring in 90 days
// Create LeaseRenewalWorkflow if not exists
// Send initial renewal offer via configured channel (email/WhatsApp/SMS)
```

**API Endpoints:**
```
POST   /api/v1/lease-renewals/{lease_id}/initiate  (Admin+)
GET    /api/v1/lease-renewals                      (filtered by status)
GET    /api/v1/lease-renewals/{id}
PUT    /api/v1/lease-renewals/{id}/propose         (Agent+)
POST   /api/v1/lease-renewals/{id}/send-offer      (Agent+)
PUT    /api/v1/lease-renewals/{id}/accept          (Admin+ or Tenant via link)
PUT    /api/v1/lease-renewals/{id}/reject          (Admin+ or Tenant)
POST   /api/v1/lease-renewals/{id}/counter-offer   (Agent+)

GET    /api/v1/renewal-templates
POST   /api/v1/renewal-templates                   (Admin only)
```

**Implementation Tasks:**
- [ ] Create migration 00026
- [ ] Add domain models
- [ ] Create repositories: LeaseRenewalRepo, RenewalTemplateRepo
- [ ] Create handler: LeaseRenewalHandler
- [ ] Add background job for renewal initiation (90 days before expiry)
- [ ] Frontend: Lease renewal workflow UI with tenant response tracking
- [ ] Create tenant-facing renewal acceptance link (no auth required, token-based)
- [ ] Email/WhatsApp templates with bilingual support

---

### 16.2: Commission Tracking & Agent Performance

**Goal:** Track and calculate commissions for agents on deals and rental leases.

**Schema Design:**
```sql
-- Migration: 00027_create_commission_tracking.sql
CREATE TABLE commission_structures (
    id UUID PRIMARY KEY,
    company_id UUID NOT NULL,
    structure_name VARCHAR(100) NOT NULL,
    commission_type ENUM('fixed_amount', 'percentage', 'tiered'),
    applicable_to ENUM('deals', 'leases', 'both'),
    effective_from DATE,
    effective_to DATE,
    rules JSONB, -- {percentage: 5, min_amount: 500, tiers: [{revenue_min, commission_percent}]}
    created_at TIMESTAMP,
    UNIQUE(company_id, structure_name)
);

CREATE TABLE agent_commissions (
    id UUID PRIMARY KEY,
    company_id UUID NOT NULL,
    agent_id UUID NOT NULL REFERENCES users(id),
    commission_period_start DATE NOT NULL,
    commission_period_end DATE NOT NULL,
    commission_structure_id UUID REFERENCES commission_structures(id),
    deals_count INT DEFAULT 0,
    deals_revenue DECIMAL(15,2) DEFAULT 0,
    leases_count INT DEFAULT 0,
    leases_revenue DECIMAL(15,2) DEFAULT 0,
    total_commission DECIMAL(15,2),
    status ENUM('pending', 'approved', 'paid', 'disputed'),
    approval_date TIMESTAMP,
    payment_date TIMESTAMP,
    payment_reference VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE commission_transactions (
    id UUID PRIMARY KEY,
    commission_id UUID NOT NULL REFERENCES agent_commissions(id),
    deal_id UUID REFERENCES deals(id),
    lease_id UUID REFERENCES leases(id),
    transaction_amount DECIMAL(15,2),
    transaction_type ENUM('deal_commission', 'lease_commission', 'bonus', 'deduction'),
    created_at TIMESTAMP
);

CREATE INDEX idx_agent_commissions_period ON agent_commissions(agent_id, commission_period_start);
CREATE INDEX idx_agent_commissions_status ON agent_commissions(status);
```

**Domain Models:**
```go
type CommissionStructure struct {
    ID           uuid.UUID
    CompanyID    uuid.UUID
    StructureName string
    CommissionType CommissionType // enum
    ApplicableTo  ApplicableToType // enum
    EffectiveFrom time.Time
    EffectiveTo   *time.Time
    Rules         map[string]interface{} // JSON
    CreatedAt     time.Time
}

type AgentCommission struct {
    ID                    uuid.UUID
    CompanyID             uuid.UUID
    AgentID               uuid.UUID
    CommissionPeriodStart time.Time
    CommissionPeriodEnd   time.Time
    CommissionStructureID *uuid.UUID
    DealsCount            int
    DealsRevenue          decimal.Decimal
    LeasesCount           int
    LeasesRevenue         decimal.Decimal
    TotalCommission       decimal.Decimal
    Status                CommissionStatus // enum
    ApprovalDate          *time.Time
    PaymentDate           *time.Time
    PaymentReference      string
    Notes                 string
    CreatedAt             time.Time
    UpdatedAt             time.Time
}
```

**Background Job:**
```go
// Monthly (on 1st day of month, for previous month)
// For each agent:
//   1. Get all deals closed in previous month
//   2. Get all leases signed in previous month
//   3. Calculate commission based on structure
//   4. Create AgentCommission record with status=pending
//   5. Notify Admin for approval
```

**API Endpoints:**
```
-- Commission Structures
GET    /api/v1/commission-structures      (read-only)
POST   /api/v1/commission-structures      (Admin only)
PUT    /api/v1/commission-structures/{id} (Admin only)

-- Agent Commissions
GET    /api/v1/agent-commissions          (filtered by period/agent)
GET    /api/v1/agent-commissions/{id}
PUT    /api/v1/agent-commissions/{id}/approve  (Admin only)
POST   /api/v1/agent-commissions/{id}/pay      (Admin only)
POST   /api/v1/agent-commissions/{id}/dispute  (Agent/Admin)

-- Commission Reports
GET    /api/v1/reports/agent-performance (monthly commission summary)
```

**Implementation Tasks:**
- [ ] Create migration 00027
- [ ] Add domain models
- [ ] Create repositories: CommissionStructureRepo, AgentCommissionRepo
- [ ] Create handler: CommissionHandler
- [ ] Add background job for monthly commission calculation
- [ ] Create commission calculation service with rule evaluation
- [ ] Frontend: Agent performance dashboard (commission by period)
- [ ] Frontend: Commission approval workflow (Admin)
- [ ] CSV export for payroll/accounting

---

### 16.3: Document Management & e-Signature Integration

**Goal:** Centralized document storage with audit trail and optional e-signature support.

**Schema Design:**
```sql
-- Migration: 00028_create_document_management.sql
CREATE TABLE document_templates (
    id UUID PRIMARY KEY,
    company_id UUID NOT NULL,
    template_name VARCHAR(150) NOT NULL,
    document_type ENUM('lease', 'offer', 'inspection_report', 'maintenance_waiver', 'custom'),
    template_content TEXT, -- Can include {{variable}} placeholders
    language ENUM('ar', 'en'),
    signature_required BOOLEAN DEFAULT false,
    signature_fields JSONB, -- [{field_name, signer_role, order}]
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP,
    UNIQUE(company_id, template_name, language)
);

CREATE TABLE documents (
    id UUID PRIMARY KEY,
    company_id UUID NOT NULL,
    document_type VARCHAR(100),
    original_template_id UUID REFERENCES document_templates(id),
    related_entity_type VARCHAR(50), -- 'lease', 'deal', 'property', 'tenant'
    related_entity_id UUID,
    document_title VARCHAR(255) NOT NULL,
    file_url VARCHAR(500) NOT NULL,
    file_size_bytes INT,
    content_hash VARCHAR(64), -- SHA256 for audit trail
    signature_status ENUM('not_required', 'pending', 'signed'),
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    data_classification VARCHAR(20), -- public/internal/confidential
    retention_until DATE,
    deleted_at TIMESTAMP
);

CREATE TABLE document_signatures (
    id UUID PRIMARY KEY,
    document_id UUID NOT NULL REFERENCES documents(id),
    signer_name VARCHAR(150) NOT NULL,
    signer_email VARCHAR(255),
    signature_field_name VARCHAR(100),
    signature_status ENUM('pending', 'signed', 'rejected'),
    signed_at TIMESTAMP,
    signature_image_url VARCHAR(500),
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP
);

CREATE TABLE document_audit_log (
    id UUID PRIMARY KEY,
    document_id UUID NOT NULL REFERENCES documents(id),
    action VARCHAR(50), -- 'created', 'viewed', 'signed', 'shared', 'downloaded'
    actor_id UUID REFERENCES users(id),
    actor_email VARCHAR(255),
    action_timestamp TIMESTAMP,
    ip_address VARCHAR(45),
    details TEXT
);

CREATE INDEX idx_documents_entity ON documents(related_entity_type, related_entity_id);
CREATE INDEX idx_documents_classification ON documents(data_classification);
CREATE INDEX idx_document_signatures_status ON document_signatures(document_id, signature_status);
```

**Domain Models:**
```go
type DocumentTemplate struct {
    ID                uuid.UUID
    CompanyID         uuid.UUID
    TemplateName      string
    DocumentType      DocumentType // enum
    TemplateContent   string
    Language          Language // enum
    SignatureRequired bool
    SignatureFields   []SignatureField // JSON
    CreatedBy         uuid.UUID
    CreatedAt         time.Time
}

type Document struct {
    ID                  uuid.UUID
    CompanyID           uuid.UUID
    DocumentType        string
    OriginalTemplateID  *uuid.UUID
    RelatedEntityType   string
    RelatedEntityID     *uuid.UUID
    DocumentTitle       string
    FileURL             string
    FileSizeBytes       int
    ContentHash         string
    SignatureStatus     SignatureStatus // enum
    CreatedBy           uuid.UUID
    CreatedAt           time.Time
    UpdatedAt           time.Time
    DataClassification  string
    RetentionUntil      *time.Time
    DeletedAt           *time.Time
}

type DocumentSignature struct {
    ID                 uuid.UUID
    DocumentID         uuid.UUID
    SignerName         string
    SignerEmail        string
    SignatureFieldName string
    SignatureStatus    SignatureStatus // enum
    SignedAt           *time.Time
    SignatureImageURL  string
    IPAddress          string
    UserAgent          string
    CreatedAt          time.Time
}

type DocumentAuditLog struct {
    ID                 uuid.UUID
    DocumentID         uuid.UUID
    Action             string
    ActorID            *uuid.UUID
    ActorEmail         string
    ActionTimestamp    time.Time
    IPAddress          string
    Details            string
}
```

**API Endpoints:**
```
-- Templates
GET    /api/v1/document-templates
POST   /api/v1/document-templates         (Admin only)
PUT    /api/v1/document-templates/{id}    (Admin only)
DELETE /api/v1/document-templates/{id}    (Admin only)

-- Documents
POST   /api/v1/documents/generate         (from template)
GET    /api/v1/documents                  (filtered by entity)
GET    /api/v1/documents/{id}
DELETE /api/v1/documents/{id}             (soft-delete)
POST   /api/v1/documents/{id}/request-signature
GET    /api/v1/documents/{id}/download    (logs audit trail)

-- Signatures
POST   /api/v1/documents/{id}/sign        (public link, token-based)
GET    /api/v1/documents/{id}/audit-trail
```

**Optional: e-Signature Integration**
- Integrate with DocuSign or Adobe Sign API
- Token-based signing links (no auth required)
- Webhook notifications on signature completion
- Multi-signer workflows (sequential signing)

**Implementation Tasks:**
- [ ] Create migration 00028
- [ ] Add domain models
- [ ] Create repositories: DocumentTemplateRepo, DocumentRepo
- [ ] Create handler: DocumentHandler
- [ ] Implement document generation from templates (Golang text/template)
- [ ] Implement audit logging middleware
- [ ] Create signature request email flow (bilingual)
- [ ] Frontend: Document management page (list, preview, download)
- [ ] Frontend: Document template editor (WYSIWYG with variable insertion)
- [ ] Optional: DocuSign API integration for advanced signing

---

## Phase 16: Completion & Consolidation

### 16.4: Advanced Reporting & Data Export

**Reports to Implement:**
1. **Occupancy Report** - Units by status, turnover rates
2. **Financial Report** - Revenue, expenses, P&L by property
3. **Tenant Report** - Active, moved-out, defaulting tenants
4. **Maintenance Report** - Task completion rates, contractor performance
5. **Payment Collection** - By tenant, by property, aging analysis
6. **Commission Report** - By agent, by period, pending/paid breakdown

**Export Formats:** CSV, PDF, Excel (via go-xlsx library)

**API Endpoint:**
```
GET /api/v1/reports/{report_type}?start_date=&end_date=&format=csv|pdf|excel
```

### 16.5: Custom Fields Support

**Goal:** Allow companies to define custom properties/tenants/lease metadata.

**Schema:**
```sql
-- Migration: 00029_create_custom_fields.sql
CREATE TABLE custom_field_definitions (
    id UUID PRIMARY KEY,
    company_id UUID NOT NULL,
    entity_type VARCHAR(50), -- 'property', 'tenant', 'lease', 'deal'
    field_name VARCHAR(100) NOT NULL,
    field_type ENUM('text', 'number', 'date', 'dropdown', 'checkbox', 'textarea'),
    is_required BOOLEAN DEFAULT false,
    dropdown_options TEXT[], -- For dropdown type
    display_order INT,
    created_at TIMESTAMP,
    UNIQUE(company_id, entity_type, field_name)
);

CREATE TABLE custom_field_values (
    id UUID PRIMARY KEY,
    field_definition_id UUID NOT NULL REFERENCES custom_field_definitions(id),
    entity_id UUID NOT NULL,
    field_value TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    UNIQUE(field_definition_id, entity_id)
);
```

### 16.6: Bulk Import/Export Utilities

**Features:**
- CSV template download for each entity (properties, tenants, leases)
- Bulk import with validation and rollback on error
- Dry-run mode to preview changes
- Import history and error logging

**Implementation:**
```go
// internal/api/handler/bulk_operations.go
func (h *BulkOperationsHandler) ImportProperties(c *fiber.Ctx) error
func (h *BulkOperationsHandler) ImportTenants(c *fiber.Ctx) error
func (h *BulkOperationsHandler) ImportLeases(c *fiber.Ctx) error
func (h *BulkOperationsHandler) ExportTemplate(c *fiber.Ctx) error
```

---

## Summary: Implementation Checklist

### Phase 15 (4 weeks)
- [ ] **Week 1:** Expense tracking system (schema, repo, handlers)
- [ ] **Week 2:** Inspection & maintenance (schema, repo, handlers)
- [ ] **Week 3:** Analytics queries and caching layer
- [ ] **Week 4:** Frontend analytics dashboard with charts

### Phase 16 (4 weeks)
- [ ] **Week 1:** Lease renewal automation (schema, background job, handlers)
- [ ] **Week 2:** Commission tracking (schema, calculation service, handlers)
- [ ] **Week 3:** Document management (schema, handlers, audit logging)
- [ ] **Week 4:** Advanced reporting, custom fields, bulk import/export

---

## Architecture Notes for Phase 15-16

### Database Migrations
- All 6 new migrations (00024-00029) must maintain backward compatibility
- Use `CONSTRAINT IF NOT EXISTS` for idempotency
- Add indexes on commonly-queried columns (dates, status, company_id)

### Background Jobs
- Extend existing scheduler pattern from Phase 12
- All jobs must handle multi-company scenarios
- Implement 5-minute timeouts to prevent job pile-up
- Log job execution to audit tables

### API Response Standards
- All list endpoints paginated (limit, offset)
- Timestamps in ISO 8601 format (UTC)
- Decimal amounts as strings (avoid float precision issues)
- Include meta (total_count, page, per_page) in responses

### Frontend Patterns
- Reuse payment page components for expense tracking
- Use existing date-range filter from analytics
- Implement breadcrumb navigation for drill-down pages
- Add loading skeletons for analytics data

### Security Considerations
- Document audit logs immutable (INSERT-only, no UPDATE/DELETE)
- Soft-delete all sensitive data (expenses, documents, commissions)
- Role-based filtering on all GET endpoints (Agent sees only assigned properties)
- Encrypt sensitive fields (vendor contact, contractor details) at rest

---

## Post-Phase 16: Vision for Phase 17

**Phase 17: SaaS Multi-Tenant Platform**
- Multi-company deployment on shared database (tenant isolation via company_id)
- Subscription management (Stripe integration)
- Company self-service onboarding
- White-label support for resellers
- Advanced role permissions (company-specific custom roles)
- Usage analytics and audit reports per tenant
- API rate limiting per company tier

This roadmap positions Masaar CRM as the most comprehensive real estate management solution for UAE businesses by end of Phase 16.
