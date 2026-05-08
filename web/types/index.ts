// ─── Auth ────────────────────────────────────────────────────────────────────

export type UserRole = 'admin' | 'agent' | 'viewer'

export interface User {
  id: string
  name: string
  email: string
  role: UserRole
  lang_pref: 'ar' | 'en'
  wa_number: string
  is_active: boolean
  created_at: string
}

export interface AuthUser {
  id: string
  name: string
  email: string
  role: 'admin' | 'agent' | 'viewer'
  lang_pref: 'ar' | 'en'
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  user: AuthUser
}

// ─── Contact ─────────────────────────────────────────────────────────────────

export interface Contact {
  id: string
  phone_wa: string
  full_name: string
  email: string
  language: 'ar' | 'en'
  lead_score: number
  assigned_to: string | null
  created_at: string
  updated_at: string
}

export interface PaginatedResult<T> {
  data: T[]
  total: number
  page: number
  limit: number
}

// ─── Lead ─────────────────────────────────────────────────────────────────────

export type LeadStage = 'new' | 'contacted' | 'qualified' | 'proposal' | 'won' | 'lost'
export type LeadSource = 'whatsapp' | 'web' | 'referral' | 'event'

export interface Lead {
  id: string
  contact_id: string
  stage: LeadStage
  source: LeadSource
  deal_value: number
  currency: string
  notes: string
  lead_score?: number
  tags?: string[]
  created_at: string
  updated_at: string
  contact?: Contact
}

export type KanbanBoard = Partial<Record<LeadStage, Lead[]>>

// ─── Communication History ────────────────────────────────────────────────────

export type CommunicationType = 'whatsapp_inbound' | 'whatsapp_outbound' | 'email_sent' | 'email_received' | 'call'
export type CommunicationStatus = 'pending' | 'sent' | 'delivered' | 'read' | 'failed' | 'bounced'

export interface CommunicationHistory {
  id: number
  lead_id: string
  contact_id: string
  communication_type: CommunicationType
  direction: 'inbound' | 'outbound'
  body: string
  from_identifier: string
  to_identifier: string
  external_id: string
  status: CommunicationStatus
  metadata?: Record<string, unknown>
  created_at: string
  created_by?: string
}

// ─── WhatsApp ─────────────────────────────────────────────────────────────────

export type ThreadStatus = 'open' | 'closed' | 'pending'
export type MessageDirection = 'inbound' | 'outbound'

export interface WhatsAppThread {
  id: string
  contact_id: string
  wa_account_id: string
  thread_status: ThreadStatus
  last_message_at: string | null
  message_count: number
  ai_summary: string
  created_at: string
  contact?: Contact
}

export interface WhatsAppMessage {
  id: string
  thread_id: string
  direction: MessageDirection
  body: string
  media_url: string
  wa_message_id: string
  sent_at: string
}

// ─── Deal ─────────────────────────────────────────────────────────────────────

export type DealStage = 'open' | 'won' | 'lost'
export type InvoiceStatus = 'draft' | 'sent' | 'paid'

export interface Deal {
  id: string
  lead_id: string
  title: string
  stage: DealStage
  amount: number
  currency: string
  close_date: string | null
  probability: number
  owner_id: string
  created_at: string
}

export interface VATInvoice {
  id: string
  deal_id: string
  invoice_no: string
  subtotal: number
  vat_rate: number
  vat_amount: number
  total: number
  status: InvoiceStatus
  issued_at: string
}

// ─── Notification ─────────────────────────────────────────────────────────────

export interface Notification {
  id: string
  user_id: string
  type: string
  title: string
  body: string
  read: boolean
  data: string
  created_at: string
}

// ─── Rental Property ──────────────────────────────────────────────────────────

export type PropertyType = 'villa' | 'apartment' | 'townhouse' | 'commercial'
export type PropertyStatus = 'active' | 'inactive' | 'sold' | 'maintenance'
export type OccupancyStatus = 'vacant' | 'occupied' | 'maintenance'

export interface RentalProperty {
  id: string
  company_id: string
  name: string
  description: string
  property_type: PropertyType
  units_count: number
  area: string
  street_address: string
  building_number: string
  unit_number: string
  city: string
  emirate: string
  postal_code: string
  total_sqft: number
  bedrooms: number
  bathrooms: number
  parking_spaces: number
  amenities: string[]
  purchase_price: number
  purchase_date: string | null
  market_value: number
  currency: string
  status: PropertyStatus
  occupancy_status: OccupancyStatus
  total_occupied_units: number
  property_deed_url: string
  title_deed_number: string
  municipality_registration: string
  created_at: string
  updated_at: string
  created_by: string | null
  updated_by: string | null
}

// ─── Tenant ───────────────────────────────────────────────────────────────────

export type IDType = 'emirati_id' | 'passport' | 'driving_license' | 'trade_license'
export type EmploymentStatus = 'employed' | 'self_employed' | 'retired' | 'student'
export type TenantStatus = 'active' | 'inactive' | 'blacklisted'
export type VerificationStatus = 'pending' | 'verified' | 'rejected'

export interface Tenant {
  id: string
  company_id: string
  full_name_en: string
  full_name_ar: string
  email: string
  phone: string
  phone_wa: string
  id_type: IDType
  id_number: string
  id_expiry_date: string | null
  id_document_url: string
  is_verified: boolean
  verification_status: VerificationStatus
  verification_date: string | null
  verified_by: string | null
  verification_notes: string
  employment_status: EmploymentStatus
  employer_name: string
  annual_income: number
  income_currency: string
  salary_certificate_url: string
  nationality: string
  country_of_origin: string
  permanent_address: string
  emergency_contact_name: string
  emergency_contact_phone: string
  status: TenantStatus
  notes: string
  created_at: string
  updated_at: string
  created_by: string | null
  updated_by: string | null
}

// ─── Lease Management ────────────────────────────────────────────────────────

export type PaymentFrequency = 'monthly' | 'quarterly' | 'semi_annual' | 'annual'
export type LeaseStatus = 'active' | 'renewed' | 'terminated' | 'expired'

export interface LeaseTemplate {
  id: string
  company_id: string
  name: string
  description: string
  is_default: boolean
  payment_frequency: PaymentFrequency
  payment_day_of_month: number
  auto_generate_payments: boolean
  default_security_deposit_percent: number
  default_utility_charges: number
  default_late_fee_percent: number
  default_lease_duration_months: number
  default_notice_period_days: number
  default_renewal_duration_months: number
  template_document_url: string
  terms_conditions: string
  status: string
  created_at: string
  updated_at: string
  created_by: string | null
  updated_by: string | null
}

export interface Lease {
  id: string
  company_id: string
  property_id: string
  tenant_id: string
  template_id: string | null
  start_date: string
  end_date: string
  renewal_start_date: string | null
  renewal_end_date: string | null
  monthly_rent: number
  currency: string
  security_deposit: number
  utility_charges: number
  late_fee_percent: number
  payment_frequency: PaymentFrequency
  payment_day_of_month: number
  auto_generate_payments: boolean
  last_generated_payment_date: string | null
  notice_period_days: number
  move_out_date: string | null
  move_out_inspection_date: string | null
  lease_document_url: string
  signed_by_landlord_date: string | null
  signed_by_tenant_date: string | null
  ejari_number: string
  ejari_url: string
  status: LeaseStatus
  termination_reason: string
  termination_date: string | null
  notes: string
  created_at: string
  updated_at: string
  created_by: string | null
  updated_by: string | null
  property?: RentalProperty
  tenant?: Tenant
  template?: LeaseTemplate
}

// ─── Payment Management ───────────────────────────────────────────────────────

export type PaymentStatus = 'pending' | 'received' | 'overdue' | 'failed' | 'refunded'
export type PaymentMethod = 'transfer' | 'check' | 'cash' | 'card' | 'other'

export interface Payment {
  id: string
  company_id: string
  lease_id: string
  amount: number
  currency: string
  due_date: string
  payment_method: PaymentMethod
  status: PaymentStatus
  received_date: string | null
  received_amount: number
  late_fee: number
  bank_transaction_id: string | null
  notes: string
  created_at: string
  updated_at: string
  created_by: string | null
  updated_by: string | null
}

// ─── Bank Integration ─────────────────────────────────────────────────────────

export type TransactionType = 'credit' | 'debit' | 'transfer' | 'check'

export interface BankTransaction {
  id: string
  company_id: string
  bank_integration_id: string
  external_id: string
  amount: number
  currency: string
  transaction_date: string
  description: string
  transaction_type: TransactionType
  matched_payment_id: string | null
  match_confidence: number
  created_at: string
  updated_at: string
}

export interface BankIntegration {
  id: string
  company_id: string
  bank_name: string
  bank_code: string
  account_number: string
  account_name: string
  iban: string
  integration_type: string
  status: string
  api_key_encrypted?: string
  api_secret_encrypted?: string
  api_endpoint: string
  auto_sync: boolean
  last_sync_date: string | null
  sync_interval_hours: number
  last_sync_error: string | null
  sync_error_count: number
  is_connected: boolean
  connection_test_date: string | null
  created_at: string
  updated_at: string
  created_by: string | null
  updated_by: string | null
}

// ─── Bank Statements & Payment Confirmations ───────────────────────────────

export type FileFormat = 'csv' | 'pdf' | 'xlsx'
export type ProcessingStatus = 'pending' | 'processing' | 'completed' | 'failed'
export type DataClassification = 'public' | 'internal' | 'confidential'

export interface BankStatement {
  id: string
  company_id: string
  bank_integration_id: string
  file_name: string
  file_size_bytes: number
  file_url: string
  file_format: FileFormat
  uploaded_by: string
  upload_date: string
  processing_status: ProcessingStatus
  transactions_imported: number
  import_error: string | null
  data_classification: DataClassification
  retention_until: string | null
  created_at: string
  updated_at: string
  deleted_at: string | null
}

export type ConfirmationDeliveryStatus = 'pending' | 'sent' | 'failed' | 'bounced'
export type ConfirmationDeliveryMethod = 'email' | 'whatsapp' | 'sms'

export interface PaymentConfirmation {
  id: string
  company_id: string
  payment_id: string
  confirmation_number: string
  tenant_email: string
  tenant_phone: string | null
  sent_at: string | null
  delivery_status: ConfirmationDeliveryStatus
  delivery_method: ConfirmationDeliveryMethod
  pdf_url: string | null
  data_classification: DataClassification
  retention_until: string | null
  created_at: string
  updated_at: string
  deleted_at: string | null
}

// ─── WebSocket Events ─────────────────────────────────────────────────────────

export interface WSEvent {
  type: 'lead.created' | 'lead.stage_changed' | 'whatsapp.message' | 'notification'
  payload: unknown
}
