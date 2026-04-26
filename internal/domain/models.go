package domain

import (
	"time"

	"github.com/google/uuid"
)

// ─── User ────────────────────────────────────────────────────────────────────

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleAgent  Role = "agent"
	RoleViewer Role = "viewer"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	LangPref     string    `json:"lang_pref"` // "ar" | "en"
	WANumber     string    `json:"wa_number"`
	CreatedAt    time.Time `json:"created_at"`
}

// ─── Contact ─────────────────────────────────────────────────────────────────

type Contact struct {
	ID         uuid.UUID  `json:"id"`
	PhoneWA    string     `json:"phone_wa"`
	FullName   string     `json:"full_name"`
	Email      string     `json:"email"`
	Language   string     `json:"language"` // "ar" | "en"
	LeadScore  int        `json:"lead_score"`
	AssignedTo *uuid.UUID `json:"assigned_to"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// ─── Lead ─────────────────────────────────────────────────────────────────────

type LeadStage string

const (
	StageNew       LeadStage = "new"
	StageContacted LeadStage = "contacted"
	StageQualified LeadStage = "qualified"
	StageProposal  LeadStage = "proposal"
	StageWon       LeadStage = "won"
	StageLost      LeadStage = "lost"
)

type LeadSource string

const (
	SourceWhatsApp LeadSource = "whatsapp"
	SourceWeb      LeadSource = "web"
	SourceReferral LeadSource = "referral"
	SourceEvent    LeadSource = "event"
)

type Lead struct {
	ID          uuid.UUID  `json:"id"`
	ContactID   uuid.UUID  `json:"contact_id"`
	Stage       LeadStage  `json:"stage"`
	Source      LeadSource `json:"source"`
	DealValue   float64    `json:"deal_value"`
	Currency    string     `json:"currency"` // default: AED
	Notes       string     `json:"notes"`
	LeadScore   int        `json:"lead_score"`
	ScoreUpdatedAt *time.Time `json:"score_updated_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Joined
	Contact *Contact `json:"contact,omitempty"`
	Tags    []string `json:"tags,omitempty"`
}

// ─── WhatsApp ─────────────────────────────────────────────────────────────────

type ThreadStatus string

const (
	ThreadOpen    ThreadStatus = "open"
	ThreadClosed  ThreadStatus = "closed"
	ThreadPending ThreadStatus = "pending"
)

type WhatsAppThread struct {
	ID            uuid.UUID    `json:"id"`
	ContactID     uuid.UUID    `json:"contact_id"`
	WAAccountID   string       `json:"wa_account_id"`
	ThreadStatus  ThreadStatus `json:"thread_status"`
	LastMessageAt *time.Time   `json:"last_message_at"`
	MessageCount  int          `json:"message_count"`
	AISummary     string       `json:"ai_summary"`
	CreatedAt     time.Time    `json:"created_at"`

	// Joined
	Contact  *Contact          `json:"contact,omitempty"`
	Messages []WhatsAppMessage `json:"messages,omitempty"`
}

type MessageDirection string

const (
	DirectionInbound  MessageDirection = "inbound"
	DirectionOutbound MessageDirection = "outbound"
)

type WhatsAppMessage struct {
	ID          uuid.UUID        `json:"id"`
	ThreadID    uuid.UUID        `json:"thread_id"`
	Direction   MessageDirection `json:"direction"`
	Body        string           `json:"body"`
	MediaURL    string           `json:"media_url"`
	WAMessageID string           `json:"wa_message_id"`
	SentAt      time.Time        `json:"sent_at"`
}

type OutboundStatus string

const (
	OutboundPending   OutboundStatus = "pending"
	OutboundSent      OutboundStatus = "sent"
	OutboundDelivered OutboundStatus = "delivered"
	OutboundRead      OutboundStatus = "read"
	OutboundFailed    OutboundStatus = "failed"
)

type WhatsAppOutbound struct {
	ID            int64          `json:"id"`
	ThreadID      uuid.UUID      `json:"thread_id"`
	ToNumber      string         `json:"to_number"`
	MessageBody   string         `json:"message_body"`
	MediaURL      string         `json:"media_url"`
	WAMessageID   string         `json:"wa_message_id"`
	Status        OutboundStatus `json:"status"`
	ErrorMsg      string         `json:"error_message"`
	ScheduledAt   *time.Time     `json:"scheduled_at"`
	SentAt        *time.Time     `json:"sent_at"`
	CreatedAt     time.Time      `json:"created_at"`
	CreatedBy     *uuid.UUID     `json:"created_by"`
	Metadata      map[string]any `json:"metadata"`
}

// ─── Deal ─────────────────────────────────────────────────────────────────────

type DealStage string

const (
	DealStageOpen DealStage = "open"
	DealStageWon  DealStage = "won"
	DealStageLost DealStage = "lost"
)

type Deal struct {
	ID          uuid.UUID  `json:"id"`
	LeadID      uuid.UUID  `json:"lead_id"`
	Title       string     `json:"title"`
	Stage       DealStage  `json:"stage"`
	Amount      float64    `json:"amount"`
	Currency    string     `json:"currency"`
	CloseDate   *time.Time `json:"close_date"`
	Probability int        `json:"probability"`
	OwnerID     uuid.UUID  `json:"owner_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ─── VAT Invoice ──────────────────────────────────────────────────────────────

type InvoiceStatus string

const (
	InvoiceDraft InvoiceStatus = "draft"
	InvoiceSent  InvoiceStatus = "sent"
	InvoicePaid  InvoiceStatus = "paid"
)

type VATInvoice struct {
	ID        uuid.UUID     `json:"id"`
	DealID    uuid.UUID     `json:"deal_id"`
	InvoiceNo string        `json:"invoice_no"`
	Subtotal  float64       `json:"subtotal"`
	VATRate   float64       `json:"vat_rate"`
	VATAmount float64       `json:"vat_amount"`
	Total     float64       `json:"total"`
	QRPayload string        `json:"qr_payload"`
	Status    InvoiceStatus `json:"status"`
	IssuedAt  time.Time     `json:"issued_at"`
}

// ─── Audit Log ────────────────────────────────────────────────────────────────

type AuditLog struct {
	ID         int64      `json:"id"`
	EntityType string     `json:"entity_type"`
	EntityID   uuid.UUID  `json:"entity_id"`
	Action     string     `json:"action"`
	ActorID    *uuid.UUID `json:"actor_id"`
	Diff       any        `json:"diff"`
	Timestamp  time.Time  `json:"ts"`
}

// ─── Pagination ───────────────────────────────────────────────────────────────

type PaginatedResult[T any] struct {
	Data  []T `json:"data"`
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// ─── Notification ────────────────────────────────────────────────────────────

type Notification struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Type      string    `json:"type"` // "lead_assigned", "lead_stage_change", "new_message"
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Read      bool      `json:"read"`
	Data      string    `json:"data,omitempty"` // JSON payload for navigation
	CreatedAt time.Time `json:"created_at"`
}

// ─── Stats ────────────────────────────────────────────────────────────────────

type Stats struct {
	TotalContacts  int     `json:"total_contacts"`
	ActiveLeads    int     `json:"active_leads"`
	NewLeadsWeek   int     `json:"new_leads_week"`
	OpenThreads    int     `json:"open_threads"`
	OpenDeals      int     `json:"open_deals"`
	OpenDealsValue float64 `json:"open_deals_value"`
	WonDeals       int     `json:"won_deals"`
	WonDealsValue  float64 `json:"won_deals_value"`
}

// ─── API Settings ─────────────────────────────────────────────────────────────

type APISetting struct {
	ID           uuid.UUID  `json:"id"`
	SettingKey   string     `json:"setting_key"`
	SettingValue string     `json:"setting_value"`
	Description  string     `json:"description"`
	UpdatedAt    time.Time  `json:"updated_at"`
	UpdatedBy    *uuid.UUID `json:"updated_by"`
}

// ─── Email ────────────────────────────────────────────────────────────────────

type EmailStatus string

const (
	EmailPending  EmailStatus = "pending"
	EmailSent     EmailStatus = "sent"
	EmailFailed   EmailStatus = "failed"
	EmailBounced  EmailStatus = "bounced"
)

type EmailHistory struct {
	ID          int64          `json:"id"`
	FromEmail   string         `json:"from_email"`
	ToEmail     string         `json:"to_email"`
	Subject     string         `json:"subject"`
	Body        string         `json:"body"`
	HTMLBody    string         `json:"html_body"`
	Status      EmailStatus    `json:"status"`
	ErrorMsg    string         `json:"error_message"`
	RelatedTo   string         `json:"related_to"`   // invoice, proposal, followup
	RelatedID   *int64         `json:"related_id"`
	SentAt      *time.Time     `json:"sent_at"`
	CreatedAt   time.Time      `json:"created_at"`
	CreatedBy   *uuid.UUID     `json:"created_by"`
	Metadata    map[string]any `json:"metadata"`
}

// ─── Lead Tags ────────────────────────────────────────────────────────────────

type LeadTag struct {
	ID          int64      `json:"id"`
	LeadID      uuid.UUID  `json:"lead_id"`
	Tag         string     `json:"tag"`
	Category    string     `json:"category"` // segment, quality, interest, timeline, status
	AutoApplied bool       `json:"auto_applied"`
	CreatedAt   time.Time  `json:"created_at"`
	CreatedBy   *uuid.UUID `json:"created_by"`
}

// ─── Communication History ─────────────────────────────────────────────────────

type CommunicationType string

const (
	CommWhatsAppInbound  CommunicationType = "whatsapp_inbound"
	CommWhatsAppOutbound CommunicationType = "whatsapp_outbound"
	CommEmailSent        CommunicationType = "email_sent"
	CommEmailReceived    CommunicationType = "email_received"
	CommCall             CommunicationType = "call"
)

type CommunicationHistory struct {
	ID                int64              `json:"id"`
	LeadID            uuid.UUID          `json:"lead_id"`
	ContactID         uuid.UUID          `json:"contact_id"`
	CommunicationType CommunicationType  `json:"communication_type"`
	Direction         string             `json:"direction"` // inbound, outbound
	Body              string             `json:"body"`
	FromIdentifier    string             `json:"from_identifier"`
	ToIdentifier      string             `json:"to_identifier"`
	ExternalID        string             `json:"external_id"`
	Status            string             `json:"status"`
	Metadata          map[string]any     `json:"metadata"`
	CreatedAt         time.Time          `json:"created_at"`
	CreatedBy         *uuid.UUID         `json:"created_by"`
}

// ─── Company Settings ─────────────────────────────────────────────────────────

type CompanySettings struct {
	ID              int        `json:"id"`
	Name            string     `json:"name"`
	VATNumber       string     `json:"vat_number"`
	BusinessAddress string     `json:"business_address"`
	BusinessPhone   string     `json:"business_phone"`
	BusinessEmail   string     `json:"business_email"`
	BankName        string     `json:"bank_name"`
	BankAccount     string     `json:"bank_account"`
	BankIBAN        string     `json:"bank_iban"`
	UpdatedAt       time.Time  `json:"updated_at"`
	UpdatedBy       *uuid.UUID `json:"updated_by"`
}

// ─── Rental Properties ────────────────────────────────────────────────────────

type PropertyType string

const (
	PropertyTypeVilla       PropertyType = "villa"
	PropertyTypeApartment   PropertyType = "apartment"
	PropertyTypeTownhouse   PropertyType = "townhouse"
	PropertyTypeCommercial  PropertyType = "commercial"
)

type PropertyStatus string

const (
	PropertyStatusActive      PropertyStatus = "active"
	PropertyStatusInactive    PropertyStatus = "inactive"
	PropertyStatusSold        PropertyStatus = "sold"
	PropertyStatusMaintenance PropertyStatus = "maintenance"
)

type OccupancyStatus string

const (
	OccupancyVacant      OccupancyStatus = "vacant"
	OccupancyOccupied    OccupancyStatus = "occupied"
	OccupancyMaintenance OccupancyStatus = "maintenance"
)

type RentalProperty struct {
	ID                  uuid.UUID       `json:"id"`
	CompanyID           uuid.UUID       `json:"company_id"`
	Name                string          `json:"name"`
	Description         string          `json:"description"`
	PropertyType        PropertyType    `json:"property_type"`
	UnitsCount          int             `json:"units_count"`
	Area                string          `json:"area"`
	StreetAddress       string          `json:"street_address"`
	BuildingNumber      string          `json:"building_number"`
	UnitNumber          string          `json:"unit_number"`
	City                string          `json:"city"`
	Emirate             string          `json:"emirate"`
	PostalCode          string          `json:"postal_code"`
	TotalSqft           float64         `json:"total_sqft"`
	Bedrooms            int             `json:"bedrooms"`
	Bathrooms           int             `json:"bathrooms"`
	ParkingSpaces       int             `json:"parking_spaces"`
	Amenities           []string        `json:"amenities"`
	PurchasePrice       float64         `json:"purchase_price"`
	PurchaseDate        *time.Time      `json:"purchase_date"`
	MarketValue         float64         `json:"market_value"`
	Currency            string          `json:"currency"`
	Status              PropertyStatus  `json:"status"`
	OccupancyStatus     OccupancyStatus `json:"occupancy_status"`
	TotalOccupiedUnits  int             `json:"total_occupied_units"`
	PropertyDeedURL     string          `json:"property_deed_url"`
	TitleDeedNumber     string          `json:"title_deed_number"`
	MunicipalityRegNum  string          `json:"municipality_registration"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
	CreatedBy           *uuid.UUID      `json:"created_by"`
	UpdatedBy           *uuid.UUID      `json:"updated_by"`
}

// ─── Tenants ──────────────────────────────────────────────────────────────────

type IDType string

const (
	IDTypeEmiratiID     IDType = "emirati_id"
	IDTypePassport      IDType = "passport"
	IDTypeDrivingLicense IDType = "driving_license"
	IDTypeTradeLicense  IDType = "trade_license"
)

type EmploymentStatus string

const (
	EmploymentEmployed      EmploymentStatus = "employed"
	EmploymentSelfEmployed  EmploymentStatus = "self_employed"
	EmploymentRetired       EmploymentStatus = "retired"
	EmploymentStudent       EmploymentStatus = "student"
)

type TenantStatus string

const (
	TenantStatusActive      TenantStatus = "active"
	TenantStatusInactive    TenantStatus = "inactive"
	TenantStatusBlacklisted TenantStatus = "blacklisted"
)

type VerificationStatus string

const (
	VerificationPending  VerificationStatus = "pending"
	VerificationVerified VerificationStatus = "verified"
	VerificationRejected VerificationStatus = "rejected"
)

type Tenant struct {
	ID                      uuid.UUID          `json:"id"`
	CompanyID               uuid.UUID          `json:"company_id"`
	FullNameEN              string             `json:"full_name_en"`
	FullNameAR              string             `json:"full_name_ar"`
	Email                   string             `json:"email"`
	Phone                   string             `json:"phone"`
	PhoneWA                 string             `json:"phone_wa"`
	IDType                  IDType             `json:"id_type"`
	IDNumber                string             `json:"id_number"`
	IDExpiryDate            *time.Time         `json:"id_expiry_date"`
	IDDocumentURL           string             `json:"id_document_url"`
	IsVerified              bool               `json:"is_verified"`
	VerificationStatus      VerificationStatus `json:"verification_status"`
	VerificationDate        *time.Time         `json:"verification_date"`
	VerifiedBy              *uuid.UUID         `json:"verified_by"`
	VerificationNotes       string             `json:"verification_notes"`
	EmploymentStatus        EmploymentStatus   `json:"employment_status"`
	EmployerName            string             `json:"employer_name"`
	AnnualIncome            float64            `json:"annual_income"`
	IncomeCurrency          string             `json:"income_currency"`
	SalaryCertificateURL    string             `json:"salary_certificate_url"`
	Nationality             string             `json:"nationality"`
	CountryOfOrigin         string             `json:"country_of_origin"`
	PermanentAddress        string             `json:"permanent_address"`
	EmergencyContactName    string             `json:"emergency_contact_name"`
	EmergencyContactPhone   string             `json:"emergency_contact_phone"`
	Status                  TenantStatus       `json:"status"`
	Notes                   string             `json:"notes"`
	CreatedAt               time.Time          `json:"created_at"`
	UpdatedAt               time.Time          `json:"updated_at"`
	CreatedBy               *uuid.UUID         `json:"created_by"`
	UpdatedBy               *uuid.UUID         `json:"updated_by"`
}

// ─── Lease Management ──────────────────────────────────────────────────────

type PaymentFrequency string

const (
	FrequencyMonthly      PaymentFrequency = "monthly"
	FrequencyQuarterly    PaymentFrequency = "quarterly"
	FrequencySemiAnnual   PaymentFrequency = "semi_annual"
	FrequencyAnnual       PaymentFrequency = "annual"
)

type LeaseStatus string

const (
	LeaseStatusActive      LeaseStatus = "active"
	LeaseStatusRenewed     LeaseStatus = "renewed"
	LeaseStatusTerminated  LeaseStatus = "terminated"
	LeaseStatusExpired     LeaseStatus = "expired"
)

type LeaseTemplate struct {
	ID                          uuid.UUID         `json:"id"`
	CompanyID                   uuid.UUID         `json:"company_id"`
	Name                        string            `json:"name"`
	Description                 string            `json:"description"`
	IsDefault                   bool              `json:"is_default"`
	PaymentFrequency            PaymentFrequency  `json:"payment_frequency"`
	PaymentDayOfMonth           int               `json:"payment_day_of_month"`
	AutoGeneratePayments        bool              `json:"auto_generate_payments"`
	DefaultSecurityDepositPct   float64           `json:"default_security_deposit_percent"`
	DefaultUtilityCharges       float64           `json:"default_utility_charges"`
	DefaultLateFeePercent       float64           `json:"default_late_fee_percent"`
	DefaultLeaseDurationMonths  int               `json:"default_lease_duration_months"`
	DefaultNoticePeriodDays     int               `json:"default_notice_period_days"`
	DefaultRenewalDurationMonths int              `json:"default_renewal_duration_months"`
	TemplateDocumentURL         string            `json:"template_document_url"`
	TermsConditions             string            `json:"terms_conditions"`
	Status                      string            `json:"status"`
	CreatedAt                   time.Time         `json:"created_at"`
	UpdatedAt                   time.Time         `json:"updated_at"`
	CreatedBy                   *uuid.UUID        `json:"created_by"`
	UpdatedBy                   *uuid.UUID        `json:"updated_by"`
}

type Lease struct {
	ID                      uuid.UUID         `json:"id"`
	CompanyID               uuid.UUID         `json:"company_id"`
	PropertyID              uuid.UUID         `json:"property_id"`
	TenantID                uuid.UUID         `json:"tenant_id"`
	TemplateID              *uuid.UUID        `json:"template_id"`
	StartDate               time.Time         `json:"start_date"`
	EndDate                 time.Time         `json:"end_date"`
	RenewalStartDate        *time.Time        `json:"renewal_start_date"`
	RenewalEndDate          *time.Time        `json:"renewal_end_date"`
	MonthlyRent             float64           `json:"monthly_rent"`
	Currency                string            `json:"currency"`
	SecurityDeposit         float64           `json:"security_deposit"`
	UtilityCharges          float64           `json:"utility_charges"`
	LateFeePct              float64           `json:"late_fee_percent"`
	PaymentFrequency        PaymentFrequency  `json:"payment_frequency"`
	PaymentDayOfMonth       int               `json:"payment_day_of_month"`
	AutoGeneratePayments    bool              `json:"auto_generate_payments"`
	LastGeneratedPaymentDt  *time.Time        `json:"last_generated_payment_date"`
	NoticePeriodDays        int               `json:"notice_period_days"`
	MoveOutDate             *time.Time        `json:"move_out_date"`
	MoveOutInspectionDate   *time.Time        `json:"move_out_inspection_date"`
	LeaseDocumentURL        string            `json:"lease_document_url"`
	SignedByLandlordDate    *time.Time        `json:"signed_by_landlord_date"`
	SignedByTenantDate      *time.Time        `json:"signed_by_tenant_date"`
	EjariNumber             string            `json:"ejari_number"`
	EjariURL                string            `json:"ejari_url"`
	Status                  LeaseStatus       `json:"status"`
	TerminationReason       string            `json:"termination_reason"`
	TerminationDate         *time.Time        `json:"termination_date"`
	Notes                   string            `json:"notes"`
	CreatedAt               time.Time         `json:"created_at"`
	UpdatedAt               time.Time         `json:"updated_at"`
	CreatedBy               *uuid.UUID        `json:"created_by"`
	UpdatedBy               *uuid.UUID        `json:"updated_by"`

	// Joined
	Property *RentalProperty `json:"property,omitempty"`
	Tenant   *Tenant         `json:"tenant,omitempty"`
	Template *LeaseTemplate  `json:"template,omitempty"`
}

// ─── Payment Management ────────────────────────────────────────────────────────

type PaymentMethod string

const (
	MethodTransfer PaymentMethod = "transfer"
	MethodCheck    PaymentMethod = "check"
	MethodCash     PaymentMethod = "cash"
	MethodCard     PaymentMethod = "card"
	MethodOther    PaymentMethod = "other"
)

type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentReceived  PaymentStatus = "received"
	PaymentOverdue   PaymentStatus = "overdue"
	PaymentFailed    PaymentStatus = "failed"
	PaymentRefunded  PaymentStatus = "refunded"
)

type Payment struct {
	ID                  uuid.UUID      `json:"id"`
	CompanyID           uuid.UUID      `json:"company_id"`
	LeaseID             uuid.UUID      `json:"lease_id"`
	Amount              float64        `json:"amount"`
	Currency            string         `json:"currency"`
	DueDate             time.Time      `json:"due_date"`
	PaidDate            *time.Time     `json:"paid_date"`
	PaymentMethod       PaymentMethod  `json:"payment_method"`
	PaymentReference    string         `json:"payment_reference"`
	Status              PaymentStatus  `json:"status"`
	BankTransactionID   *uuid.UUID     `json:"bank_transaction_id"`
	ReconciledAt        *time.Time     `json:"reconciled_at"`
	ReconciledBy        *uuid.UUID     `json:"reconciled_by"`
	Notes               string         `json:"notes"`
	ReceiptURL          string         `json:"receipt_url"`
	LateFeesApplied     bool           `json:"late_fee_applied"`
	LateFeeAmount       float64        `json:"late_fee_amount"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	CreatedBy           *uuid.UUID     `json:"created_by"`
	UpdatedBy           *uuid.UUID     `json:"updated_by"`

	// Joined
	Lease *Lease `json:"lease,omitempty"`
}

type TransactionType string

const (
	TypeCredit    TransactionType = "credit"
	TypeDebit     TransactionType = "debit"
	TypeTransfer  TransactionType = "transfer"
	TypeCheck     TransactionType = "check"
)

type BankTransaction struct {
	ID                  uuid.UUID        `json:"id"`
	CompanyID           uuid.UUID        `json:"company_id"`
	BankIntegrationID   *uuid.UUID       `json:"bank_integration_id"`
	ExternalID          string           `json:"external_id"`
	TransactionDate     time.Time        `json:"transaction_date"`
	Amount              float64          `json:"amount"`
	Currency            string           `json:"currency"`
	FromAccount         string           `json:"from_account"`
	ToAccount           string           `json:"to_account"`
	FromName            string           `json:"from_name"`
	ToName              string           `json:"to_name"`
	Reference           string           `json:"reference"`
	TransactionType     TransactionType  `json:"transaction_type"`
	Status              string           `json:"status"`
	MatchedPaymentID    *uuid.UUID       `json:"matched_payment_id"`
	MatchConfidence     float64          `json:"match_confidence"`
	MatchedAt           *time.Time       `json:"matched_at"`
	ImportedAt          time.Time        `json:"imported_at"`
	LastChecked         *time.Time       `json:"last_checked"`
	SyncError           string           `json:"sync_error"`
}

type BankIntegration struct {
	ID                   uuid.UUID `json:"id"`
	CompanyID            uuid.UUID `json:"company_id"`
	BankName             string    `json:"bank_name"`
	BankCode             string    `json:"bank_code"`
	AccountNumber        string    `json:"account_number"`
	AccountName          string    `json:"account_name"`
	IBAN                 string    `json:"iban"`
	IntegrationType      string    `json:"integration_type"`
	Status               string    `json:"status"`
	APIKeyEncrypted      string    `json:"-"`
	APISecretEncrypted   string    `json:"-"`
	APIEndpoint          string    `json:"api_endpoint"`
	AutoSync             bool      `json:"auto_sync"`
	LastSyncDate         *time.Time `json:"last_sync_date"`
	SyncIntervalHours    int       `json:"sync_interval_hours"`
	LastSyncError        string    `json:"last_sync_error"`
	SyncErrorCount       int       `json:"sync_error_count"`
	IsConnected          bool      `json:"is_connected"`
	ConnectionTestDate   *time.Time `json:"connection_test_date"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	CreatedBy            *uuid.UUID `json:"created_by"`
	UpdatedBy            *uuid.UUID `json:"updated_by"`
}

// ─── Payment Reminders ────────────────────────────────────────────────────────

type ReminderType string

const (
	Reminder30Days  ReminderType = "30_days"
	Reminder15Days  ReminderType = "15_days"
	Reminder7Days   ReminderType = "7_days"
	Reminder1DayLate ReminderType = "1_day_late"
	Reminder5DaysLate ReminderType = "5_days_late"
	Reminder10DaysLate ReminderType = "10_days_late"
)

type DeliveryMethod string

const (
	DeliveryWhatsApp DeliveryMethod = "whatsapp"
	DeliveryEmail    DeliveryMethod = "email"
	DeliverySMS      DeliveryMethod = "sms"
)

type DeliveryStatus string

const (
	DeliveryPending DeliveryStatus = "pending"
	DeliverySent    DeliveryStatus = "sent"
	DeliveryFailed  DeliveryStatus = "failed"
)

type PaymentReminder struct {
	ID             uuid.UUID      `json:"id"`
	CompanyID      uuid.UUID      `json:"company_id"`
	PaymentID      uuid.UUID      `json:"payment_id"`
	ReminderType   ReminderType   `json:"reminder_type"`
	ReminderDate   time.Time      `json:"reminder_date"`
	SentAt         *time.Time     `json:"sent_at"`
	DeliveryMethod DeliveryMethod `json:"delivery_method"`
	DeliveryStatus DeliveryStatus `json:"delivery_status"`
	DeliveryError  string         `json:"delivery_error"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// ─── Bank Statements & Payment Confirmations ───────────────────────────────

type FileFormat string

const (
	FormatCSV  FileFormat = "csv"
	FormatPDF  FileFormat = "pdf"
	FormatXLSX FileFormat = "xlsx"
)

type ProcessingStatus string

const (
	StatusPending    ProcessingStatus = "pending"
	StatusProcessing ProcessingStatus = "processing"
	StatusCompleted  ProcessingStatus = "completed"
	StatusFailed     ProcessingStatus = "failed"
)

type DataClassification string

const (
	ClassPublic      DataClassification = "public"
	ClassInternal    DataClassification = "internal"
	ClassConfidential DataClassification = "confidential"
)

type BankStatement struct {
	ID                  uuid.UUID          `json:"id"`
	CompanyID           uuid.UUID          `json:"company_id"`
	BankIntegrationID   uuid.UUID          `json:"bank_integration_id"`
	FileName            string             `json:"file_name"`
	FileSizeBytes       int                `json:"file_size_bytes"`
	FileURL             string             `json:"file_url"`
	FileFormat          FileFormat         `json:"file_format"`
	UploadedBy          uuid.UUID          `json:"uploaded_by"`
	UploadDate          time.Time          `json:"upload_date"`
	ProcessingStatus    ProcessingStatus   `json:"processing_status"`
	TransactionsImported int               `json:"transactions_imported"`
	ImportError         string             `json:"import_error"`
	DataClassification  DataClassification `json:"data_classification"`
	RetentionUntil      *time.Time         `json:"retention_until"`
	CreatedAt           time.Time          `json:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at"`
	DeletedAt           *time.Time         `json:"deleted_at"`
}

type ConfirmationDeliveryStatus string

const (
	ConfPending  ConfirmationDeliveryStatus = "pending"
	ConfSent     ConfirmationDeliveryStatus = "sent"
	ConfFailed   ConfirmationDeliveryStatus = "failed"
	ConfBounced  ConfirmationDeliveryStatus = "bounced"
)

type ConfirmationDeliveryMethod string

const (
	ConfDeliveryEmail    ConfirmationDeliveryMethod = "email"
	ConfDeliveryWhatsApp ConfirmationDeliveryMethod = "whatsapp"
	ConfDeliverySMS      ConfirmationDeliveryMethod = "sms"
)

type PaymentConfirmation struct {
	ID                 uuid.UUID                      `json:"id"`
	CompanyID          uuid.UUID                      `json:"company_id"`
	PaymentID          uuid.UUID                      `json:"payment_id"`
	ConfirmationNumber string                         `json:"confirmation_number"`
	TenantEmail        string                         `json:"tenant_email"`
	TenantPhone        *string                        `json:"tenant_phone"`
	SentAt             *time.Time                     `json:"sent_at"`
	DeliveryStatus     ConfirmationDeliveryStatus     `json:"delivery_status"`
	DeliveryMethod     ConfirmationDeliveryMethod     `json:"delivery_method"`
	PDFURL             *string                        `json:"pdf_url"`
	DataClassification DataClassification            `json:"data_classification"`
	RetentionUntil     *time.Time                     `json:"retention_until"`
	CreatedAt          time.Time                      `json:"created_at"`
	UpdatedAt          time.Time                      `json:"updated_at"`
	DeletedAt          *time.Time                     `json:"deleted_at"`
}
