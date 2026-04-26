// ─── Auth ────────────────────────────────────────────────────────────────────

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

// ─── WebSocket Events ─────────────────────────────────────────────────────────

export interface WSEvent {
  type: 'lead.created' | 'lead.stage_changed' | 'whatsapp.message' | 'notification'
  payload: unknown
}
