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

// ─── WebSocket Events ─────────────────────────────────────────────────────────

export interface WSEvent {
  type: 'lead.created' | 'lead.stage_changed' | 'whatsapp.message' | 'notification'
  payload: unknown
}
