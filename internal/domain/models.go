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
	ID        uuid.UUID  `json:"id"`
	ContactID uuid.UUID  `json:"contact_id"`
	Stage     LeadStage  `json:"stage"`
	Source    LeadSource `json:"source"`
	DealValue float64    `json:"deal_value"`
	Currency  string     `json:"currency"` // default: AED
	Notes     string     `json:"notes"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`

	// Joined
	Contact *Contact `json:"contact,omitempty"`
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
