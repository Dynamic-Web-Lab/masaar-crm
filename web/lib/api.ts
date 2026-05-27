import { getToken, clearSession } from './auth'

const BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const token = getToken()
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...(init.headers || {}),
  }

  const res = await fetch(`${BASE}${path}`, { ...init, headers })

  if (res.status === 401) {
    clearSession()
    window.location.href = '/login'
    throw new Error('Unauthorized')
  }

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error || 'Request failed')
  }

  if (res.status === 204) return undefined as T
  return res.json()
}

async function requestBlob(path: string, init: RequestInit = {}): Promise<Blob> {
  const token = getToken()
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...(init.headers || {}),
  }
  const res = await fetch(`${BASE}${path}`, { ...init, headers })
  if (res.status === 401) {
    clearSession()
    window.location.href = '/login'
    throw new Error('Unauthorized')
  }
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error || 'Request failed')
  }
  return res.blob()
}

// ─── Auth ─────────────────────────────────────────────────────────────────────

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, data?: unknown) => request<T>(path, { method: 'POST', body: data ? JSON.stringify(data) : undefined }),
  put: <T>(path: string, data?: unknown) => request<T>(path, { method: 'PUT', body: data ? JSON.stringify(data) : undefined }),
  delete: <T>(path: string) => request<T>(path, { method: 'DELETE' }),
  
  auth: {
    login: (email: string, password: string) =>
      request('/api/v1/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      }),
    requestMagicLink: (email: string, lang_pref?: 'ar' | 'en') =>
      request('/api/v1/auth/magic-link/request', {
        method: 'POST',
        body: JSON.stringify({ email, lang_pref }),
      }),
    verifyMagicLink: (token: string) =>
      request('/api/v1/auth/magic-link/verify', {
        method: 'POST',
        body: JSON.stringify({ token }),
      }),
    logout: (refreshToken: string) =>
      request('/api/v1/auth/logout', {
        method: 'DELETE',
        body: JSON.stringify({ refresh_token: refreshToken }),
      }),
    requestSMSOTP: (phone: string, lang?: string) =>
      request('/api/v1/auth/sms/request', {
        method: 'POST',
        body: JSON.stringify({ phone, lang }),
      }),
    verifySMSOTP: (phone: string, otp: string) =>
      request('/api/v1/auth/sms/verify', {
        method: 'POST',
        body: JSON.stringify({ phone, otp }),
      }),
    forgotPassword: (email: string) =>
      request('/api/v1/auth/forgot-password', {
        method: 'POST',
        body: JSON.stringify({ email }),
      }),
    resetPassword: (token: string, password: string) =>
      request('/api/v1/auth/reset-password', {
        method: 'POST',
        body: JSON.stringify({ token, password }),
      }),
    refresh: (refreshToken: string) =>
      request('/api/v1/auth/refresh', {
        method: 'POST',
        body: JSON.stringify({ refresh_token: refreshToken }),
      }),
    register: (data: { name: string; email: string; password: string; company_name: string; subdomain: string; turnstile_token?: string }) =>
      request('/api/v1/auth/register', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
  },

  // ─── Contacts ───────────────────────────────────────────────────────────────

  contacts: {
    list: (params: { search?: string; page?: number; limit?: number } = {}) => {
      const q = new URLSearchParams()
      if (params.search) q.set('search', params.search)
      if (params.page) q.set('page', String(params.page))
      if (params.limit) q.set('limit', String(params.limit))
      return request(`/api/v1/contacts?${q}`)
    },
    get: (id: string) => request(`/api/v1/contacts/${id}`),
    create: (data: unknown) =>
      request('/api/v1/contacts', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: unknown) =>
      request(`/api/v1/contacts/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
    delete: (id: string) =>
      request(`/api/v1/contacts/${id}`, { method: 'DELETE' }),
  },

  // ─── Leads ──────────────────────────────────────────────────────────────────

  leads: {
    kanban: () => request('/api/v1/leads'),
    get: (id: string) => request(`/api/v1/leads/${id}`),
    create: (data: unknown) =>
      request('/api/v1/leads', { method: 'POST', body: JSON.stringify(data) }),
    updateStage: (id: string, stage: string, closedReason?: string) => {
      const body: Record<string, unknown> = { stage }
      if (closedReason) body.closed_reason = closedReason
      return request(`/api/v1/leads/${id}/stage`, {
        method: 'PATCH',
        body: JSON.stringify(body),
      })
    },
    updateNotes: (id: string, notes: string) =>
      request(`/api/v1/leads/${id}/notes`, {
        method: 'PATCH',
        body: JSON.stringify({ notes }),
      }),
    communications: (id: string, limit: number = 100) => {
      const q = new URLSearchParams()
      q.set('limit', String(limit))
      return request(`/api/v1/leads/${id}/communications?${q}`)
    },
    getTags: (id: string) =>
      request(`/api/v1/leads/${id}/tags`),
    addTag: (id: string, tag: string, category?: string) =>
      request(`/api/v1/leads/${id}/tags`, {
        method: 'POST',
        body: JSON.stringify({ tag, category: category || 'manual' }),
      }),
    removeTag: (id: string, tag: string) =>
      request(`/api/v1/leads/${id}/tags/${encodeURIComponent(tag)}`, { method: 'DELETE' }),
    search: (params: { q?: string; stage?: string; source?: string; assigned_to?: string; contact_id?: string; page?: number; limit?: number } = {}) => {
      const q = new URLSearchParams()
      if (params.q) q.set('q', params.q)
      if (params.stage) q.set('stage', params.stage)
      if (params.source) q.set('source', params.source)
      if (params.assigned_to) q.set('assigned_to', params.assigned_to)
      if (params.contact_id) q.set('contact_id', params.contact_id)
      if (params.page) q.set('offset', String(((params.page || 1) - 1) * (params.limit || 50)))
      if (params.limit) q.set('limit', String(params.limit))
      return request(`/api/v1/leads/search?${q}`)
    },
    assign: (id: string, userId: string | null) =>
      request(`/api/v1/leads/${id}/assign`, {
        method: 'PATCH',
        body: JSON.stringify({ assigned_to: userId }),
      }),
    delete: (id: string) =>
      request(`/api/v1/leads/${id}`, { method: 'DELETE' }),
    reScore: (id: string) =>
      request(`/api/v1/ai/score-lead/${id}`, { method: 'POST' }),
  },

  // ─── WhatsApp ───────────────────────────────────────────────────────────────

  threads: {
    list: (params: { status?: string; contact_id?: string; page?: number; limit?: number } = {}) => {
      const q = new URLSearchParams()
      if (params.status) q.set('status', params.status)
      if (params.contact_id) q.set('contact_id', params.contact_id)
      if (params.page) q.set('page', String(params.page))
      if (params.limit) q.set('limit', String(params.limit))
      return request(`/api/v1/threads?${q}`)
    },
    get: (id: string) => request(`/api/v1/threads/${id}`),
    messages: (id: string) => request(`/api/v1/threads/${id}/messages`),
    close: (id: string) =>
      request(`/api/v1/threads/${id}/close`, { method: 'POST' }),
    reopen: (id: string) =>
      request(`/api/v1/threads/${id}/reopen`, { method: 'POST' }),
  },

  // ─── AI ─────────────────────────────────────────────────────────────────────

  ai: {
    summarize: (threadId: string) =>
      request(`/api/v1/ai/summarize/${threadId}`, { method: 'POST' }),
    extractBuyerProfile: (threadId: string) =>
      request(`/api/v1/ai/extract-buyer-profile/${threadId}`, { method: 'POST' }),
    scoreLead: (id: string) =>
      request(`/api/v1/ai/score-lead/${id}`, { method: 'POST' }),
    scoreContact: (id: string) =>
      request(`/api/v1/ai/score-contact/${id}`, { method: 'POST' }),
    draftReply: (threadId: string) =>
      request(`/api/v1/ai/draft-reply/${threadId}`, { method: 'POST' }),
    describeListing: (data: unknown) =>
      request('/api/v1/ai/describe-listing', { method: 'POST', body: JSON.stringify(data) }),
  },

  // ─── Deals ──────────────────────────────────────────────────────────────────

  deals: {
    list: (params: { page?: number; limit?: number; stage?: string; owner_id?: string } = {}) => {
      const q = new URLSearchParams()
      if (params.page) q.set('page', String(params.page))
      if (params.limit) q.set('limit', String(params.limit))
      if (params.stage) q.set('stage', params.stage)
      if (params.owner_id) q.set('owner_id', params.owner_id)
      return request(`/api/v1/deals?${q}`)
    },
    get: (id: string) => request(`/api/v1/deals/${id}`),
    create: (data: unknown) =>
      request('/api/v1/deals', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: unknown) =>
      request(`/api/v1/deals/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
    updateStage: (id: string, stage: string) =>
      request(`/api/v1/deals/${id}/stage`, {
        method: 'PATCH',
        body: JSON.stringify({ stage }),
      }),
    invoices: (id: string) => request(`/api/v1/deals/${id}/invoices`),
  },

  // ─── Invoices ────────────────────────────────────────────────────────────────

  invoices: {
    list: (page = 1, limit = 50) =>
      request(`/api/v1/invoices?page=${page}&limit=${limit}`),
    create: (data: unknown) =>
      request('/api/v1/invoices', { method: 'POST', body: JSON.stringify(data) }),
    get: (id: string) => request(`/api/v1/invoices/${id}`),
    getPDF: (id: string) => {
      return `${BASE}/api/v1/invoices/${id}/pdf`
    },
    send: (id: string) =>
      request(`/api/v1/invoices/${id}/send`, { method: 'POST' }),
    updateStatus: (id: string, status: string) =>
      request(`/api/v1/invoices/${id}/status`, {
        method: 'PATCH',
        body: JSON.stringify({ status }),
      }),
  },

  // ─── Stats ───────────────────────────────────────────────────────────────────

  stats: {
    overview: () => request('/api/v1/stats'),
  },

  // ─── User Settings ───────────────────────────────────────────────────────────

  users: {
    me: () => request('/api/v1/users/me'),
    changePassword: (currentPassword: string, newPassword: string) =>
      request('/api/v1/users/me/password', {
        method: 'PATCH',
        body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
      }),
    updateLang: (lang: 'ar' | 'en') =>
      request<{ lang_pref: string }>('/api/v1/users/me/lang', {
        method: 'PATCH',
        body: JSON.stringify({ lang }),
      }),
    list: () => request('/api/v1/users'),
    invite: (data: { name: string; email: string; role: string }) =>
      request('/api/v1/users/invite', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: { name: string; role: string }) =>
      request(`/api/v1/users/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
    setActive: (id: string, active: boolean) =>
      request(`/api/v1/users/${id}/active`, { method: 'PATCH', body: JSON.stringify({ active }) }),
    delete: (id: string) =>
      request(`/api/v1/users/${id}`, { method: 'DELETE' }),
  },

  // ─── Notifications ───────────────────────────────────────────────────────────

  notifications: {
    list: (params: { page?: number } = {}) => {
      const q = new URLSearchParams()
      if (params.page) q.set('page', String(params.page))
      return request(`/api/v1/notifications?${q}`)
    },
    markRead: (id: string) =>
      request(`/api/v1/notifications/${id}/read`, { method: 'PATCH' }),
    markAllRead: () =>
      request(`/api/v1/notifications/read-all`, { method: 'PATCH' }),
  },

  // ─── Rental Properties ────────────────────────────────────────────────────────

  rentalProperties: {
    list: (params: { page?: number; limit?: number } = {}) => {
      const q = new URLSearchParams()
      if (params.page) q.set('page', String(params.page))
      if (params.limit) q.set('limit', String(params.limit))
      return request(`/api/v1/rental-properties?${q}`)
    },
    get: (id: string) => request(`/api/v1/rental-properties/${id}`),
    create: (data: unknown) =>
      request('/api/v1/rental-properties', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: unknown) =>
      request(`/api/v1/rental-properties/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
    delete: (id: string) =>
      request(`/api/v1/rental-properties/${id}`, { method: 'DELETE' }),
  },

  // ─── Tenants ──────────────────────────────────────────────────────────────────

  tenants: {
    list: (params: { page?: number; limit?: number } = {}) => {
      const q = new URLSearchParams()
      if (params.page) q.set('page', String(params.page))
      if (params.limit) q.set('limit', String(params.limit))
      return request(`/api/v1/tenants?${q}`)
    },
    get: (id: string) => request(`/api/v1/tenants/${id}`),
    create: (data: unknown) =>
      request('/api/v1/tenants', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: unknown) =>
      request(`/api/v1/tenants/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
    delete: (id: string) =>
      request(`/api/v1/tenants/${id}`, { method: 'DELETE' }),
    verify: (id: string, verificationNotes: string) =>
      request(`/api/v1/tenants/${id}/verify`, {
        method: 'POST',
        body: JSON.stringify({ verification_notes: verificationNotes }),
      }),
  },

  // ─── Messages ─────────────────────────────────────────────────────────────────

  messages: {
    suggestAction: (data: unknown) =>
      request('/api/v1/messages/suggest-action', { method: 'POST', body: JSON.stringify(data) }),
    analyze: (data: unknown) =>
      request('/api/v1/messages/analyze', { method: 'POST', body: JSON.stringify(data) }),
    autoCreateLead: (data: unknown) =>
      request('/api/v1/messages/auto-create-lead', { method: 'POST', body: JSON.stringify(data) }),
  },

  // ─── Lease Templates ──────────────────────────────────────────────────────────

  leaseTemplates: {
    list: (params: { page?: number; limit?: number } = {}) => {
      const q = new URLSearchParams()
      if (params.page) q.set('page', String(params.page))
      if (params.limit) q.set('limit', String(params.limit))
      return request(`/api/v1/lease-templates?${q}`)
    },
    get: (id: string) => request(`/api/v1/lease-templates/${id}`),
    create: (data: unknown) =>
      request('/api/v1/lease-templates', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: unknown) =>
      request(`/api/v1/lease-templates/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
    delete: (id: string) =>
      request(`/api/v1/lease-templates/${id}`, { method: 'DELETE' }),
  },

  // ─── Leases ────────────────────────────────────────────────────────────────────

  leases: {
    list: (params: { page?: number; limit?: number } = {}) => {
      const q = new URLSearchParams()
      if (params.page) q.set('page', String(params.page))
      if (params.limit) q.set('limit', String(params.limit))
      return request(`/api/v1/leases?${q}`)
    },
    get: (id: string) => request(`/api/v1/leases/${id}`),
    create: (data: unknown) =>
      request('/api/v1/leases', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: unknown) =>
      request(`/api/v1/leases/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
    delete: (id: string) =>
      request(`/api/v1/leases/${id}`, { method: 'DELETE' }),
  },

  // ─── Payments ─────────────────────────────────────────────────────────────────

  payments: {
    list: (params: { page?: number; limit?: number } = {}) => {
      const q = new URLSearchParams()
      if (params.page) q.set('page', String(params.page))
      if (params.limit) q.set('limit', String(params.limit))
      return request(`/api/v1/payments?${q}`)
    },
    get: (id: string) => request(`/api/v1/payments/${id}`),
    create: (data: unknown) =>
      request('/api/v1/payments', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: unknown) =>
      request(`/api/v1/payments/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
    delete: (id: string) =>
      request(`/api/v1/payments/${id}`, { method: 'DELETE' }),
  },

  // ─── Bank Integrations ────────────────────────────────────────────────────────

  bankIntegrations: {
    list: (params: { page?: number; limit?: number } = {}) => {
      const q = new URLSearchParams()
      if (params.page) q.set('page', String(params.page))
      if (params.limit) q.set('limit', String(params.limit))
      return request(`/api/v1/bank-integrations?${q}`)
    },
    get: (id: string) => request(`/api/v1/bank-integrations/${id}`),
    create: (data: unknown) =>
      request('/api/v1/bank-integrations', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: unknown) =>
      request(`/api/v1/bank-integrations/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
    delete: (id: string) =>
      request(`/api/v1/bank-integrations/${id}`, { method: 'DELETE' }),
  },

  // ─── Bank Statements ──────────────────────────────────────────────────────

  bankStatements: {
    list: (params: { page?: number; limit?: number } = {}) => {
      const q = new URLSearchParams()
      if (params.page) q.set('page', String(params.page))
      if (params.limit) q.set('limit', String(params.limit))
      return request(`/api/v1/bank-statements?${q}`)
    },
    get: (id: string) => request(`/api/v1/bank-statements/${id}`),
    upload: (file: File, bankIntegrationId: string) => {
      const formData = new FormData()
      formData.append('file', file)
      formData.append('bank_integration_id', bankIntegrationId)
      const token = getToken()
      return fetch(`${BASE}/api/v1/bank-statements/upload`, {
        method: 'POST',
        headers: token ? { Authorization: `Bearer ${token}` } : {},
        body: formData,
      }).then(async (res) => {
        if (!res.ok) throw new Error((await res.json()).error || 'Upload failed')
        return res.json()
      })
    },
    delete: (id: string) =>
      request(`/api/v1/bank-statements/${id}`, { method: 'DELETE' }),
  },

  // ─── Payment Confirmations ────────────────────────────────────────────────

  paymentConfirmations: {
    getByPayment: (paymentId: string) =>
      request(`/api/v1/payments/${paymentId}/confirmation`),
    send: (paymentId: string) =>
      request(`/api/v1/payments/${paymentId}/send-confirmation`, { method: 'POST' }),
  },

  // ─── Analytics ────────────────────────────────────────────────────────────

  analytics: {
    getTenantOverview: () =>
      request('/api/v1/analytics/tenant-overview'),
    listProperties: (params: { limit?: number; offset?: number } = {}) => {
      const q = new URLSearchParams()
      if (params.limit) q.set('limit', String(params.limit))
      if (params.offset) q.set('offset', String(params.offset))
      return request(`/api/v1/analytics/properties?${q}`)
    },
    getProperty: (propertyId: string) =>
      request(`/api/v1/analytics/properties/${propertyId}`),
    listTenants: (params: { limit?: number; offset?: number } = {}) => {
      const q = new URLSearchParams()
      if (params.limit) q.set('limit', String(params.limit))
      if (params.offset) q.set('offset', String(params.offset))
      return request(`/api/v1/analytics/tenants?${q}`)
    },
    getTenant: (tenantId: string) =>
      request(`/api/v1/analytics/tenants/${tenantId}`),
    getFinancial: (startDate?: string, endDate?: string) => {
      const q = new URLSearchParams()
      if (startDate) q.set('startDate', startDate)
      if (endDate) q.set('endDate', endDate)
      return request(`/api/v1/analytics/financial?${q}`)
    },
    getMaintenance: () =>
      request('/api/v1/analytics/maintenance'),
  },

  // ─── Inspections ───────────────────────────────────────────────────────────

  inspection: {
    list: (params: { page?: number; limit?: number } = {}) => {
      const q = new URLSearchParams()
      if (params.page) q.set('offset', String((params.page - 1) * (params.limit || 20)))
      if (params.limit) q.set('limit', String(params.limit))
      return request(`/api/v1/inspections?${q}`)
    },
    get: (id: string) =>
      request(`/api/v1/inspections/${id}`),
    create: (data: any) =>
      request('/api/v1/inspections', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    update: (id: string, data: any) =>
      request(`/api/v1/inspections/${id}`, {
        method: 'PATCH',
        body: JSON.stringify(data),
      }),
    complete: (id: string) =>
      request(`/api/v1/inspections/${id}/complete`, { method: 'POST' }),
    listTemplates: () =>
      request('/api/v1/inspection-templates'),
    createTemplate: (data: any) =>
      request('/api/v1/inspection-templates', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
  },

  // ─── Maintenance Tasks ───────────────────────────────────────────────────────

  maintenance: {
    list: (params: { page?: number; limit?: number; status?: string } = {}) => {
      const q = new URLSearchParams()
      if (params.page) q.set('offset', String((params.page - 1) * (params.limit || 20)))
      if (params.limit) q.set('limit', String(params.limit))
      if (params.status) q.set('status', params.status)
      return request(`/api/v1/maintenance-tasks?${q}`)
    },
    get: (id: string) =>
      request(`/api/v1/maintenance-tasks/${id}`),
    create: (data: any) =>
      request('/api/v1/maintenance-tasks', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    update: (id: string, data: any) =>
      request(`/api/v1/maintenance-tasks/${id}`, {
        method: 'PATCH',
        body: JSON.stringify(data),
      }),
    complete: (id: string, data?: any) =>
      request(`/api/v1/maintenance-tasks/${id}/complete`, {
        method: 'POST',
        body: data ? JSON.stringify(data) : undefined,
      }),
    addPhoto: (id: string, photoUrl: string, stage: string) =>
      request(`/api/v1/maintenance-tasks/${id}/photos`, {
        method: 'POST',
        body: JSON.stringify({ photo_url: photoUrl, photo_stage: stage }),
      }),
    getPhotos: (id: string) =>
      request(`/api/v1/maintenance-tasks/${id}/photos`),
    delete: (id: string) =>
      request(`/api/v1/maintenance-tasks/${id}`, { method: 'DELETE' }),
  },

  // ─── Billing & Plans ─────────────────────────────────────────────────────────

  billing: {
    get: () => request('/api/v1/billing'),
    getUsage: () => request('/api/v1/billing/usage'),
    checkout: (plan: string) =>
      request('/api/v1/billing/checkout', { method: 'POST', body: JSON.stringify({ plan }) }),
    portal: () =>
      request('/api/v1/billing/portal', { method: 'POST' }),
  },

  // ─── Settings ────────────────────────────────────────────────────────────────

  settings: {
    getCompany: () => request('/api/v1/settings/company'),
    updateCompany: (data: unknown) =>
      request('/api/v1/settings/company', { method: 'PATCH', body: JSON.stringify(data) }),
    getBOS24: () => request('/api/v1/settings/bos24'),
    updateBOS24: (data: unknown) =>
      request('/api/v1/settings/bos24', { method: 'PATCH', body: JSON.stringify(data) }),
    apiKeys: {
      list: () => request('/api/v1/settings/api-keys'),
      create: (data: unknown) =>
        request('/api/v1/settings/api-keys', { method: 'POST', body: JSON.stringify(data) }),
      revoke: (id: string) =>
        request(`/api/v1/settings/api-keys/${id}`, { method: 'DELETE' }),
    },
    webhooks: {
      list: () => request('/api/v1/settings/webhooks'),
      create: (data: unknown) =>
        request('/api/v1/settings/webhooks', { method: 'POST', body: JSON.stringify(data) }),
      delete: (id: string) =>
        request(`/api/v1/settings/webhooks/${id}`, { method: 'DELETE' }),
      test: (id: string) =>
        request(`/api/v1/settings/webhooks/${id}/test`, { method: 'POST' }),
    },
  },

  // ─── WhatsApp Outbound ─────────────────────────────────────────────────────────

  whatsapp: {
    sendMessage: (threadId: string, body: string) =>
      request(`/api/v1/threads/${threadId}/send-message`, {
        method: 'POST',
        body: JSON.stringify({ message: body }),
      }),
    sendMedia: (threadId: string, data: { media_url: string; media_type: string; caption?: string }) =>
      request(`/api/v1/threads/${threadId}/send-media`, {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    sendTemplate: (threadId: string, templateName: string, params: string[]) =>
      request(`/api/v1/threads/${threadId}/send-template`, {
        method: 'POST',
        body: JSON.stringify({ template_name: templateName, parameters: params }),
      }),
    getOutboundMessages: (threadId: string) =>
      request(`/api/v1/threads/${threadId}/outbound-messages`),
  },

  // ─── Email ────────────────────────────────────────────────────────────────────

  email: {
    list: (page = 1, limit = 50) =>
      request(`/api/v1/emails?page=${page}&limit=${limit}`),
    send: (data: unknown) =>
      request('/api/v1/emails/send', { method: 'POST', body: JSON.stringify(data) }),
    history: (relatedTo: string, relatedId: number) =>
      request(`/api/v1/emails/history?related_to=${relatedTo}&related_id=${relatedId}`),
  },

  // ─── Expenses ─────────────────────────────────────────────────────────────────

  expenses: {
    listCategories: () => request('/api/v1/expense-categories'),
    createCategory: (data: { category_name: string; category_type: string; description?: string }) =>
      request('/api/v1/expense-categories', { method: 'POST', body: JSON.stringify(data) }),
    list: (params: { page?: number; limit?: number } = {}) => {
      const q = new URLSearchParams()
      if (params.page) q.set('offset', String((params.page - 1) * (params.limit || 20)))
      if (params.limit) q.set('limit', String(params.limit))
      return request(`/api/v1/expenses?${q}`)
    },
    get: (id: string) => request(`/api/v1/expenses/${id}`),
    create: (data: {
      category_id: string; amount: number; expense_date: string; description: string;
      property_id?: string; tenant_id?: string; vendor_name?: string;
      vendor_contact?: string; payment_method?: string; receipt_url?: string; notes?: string;
    }) => request('/api/v1/expenses', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: { amount?: number; description?: string; payment_status?: string }) =>
      request(`/api/v1/expenses/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
    delete: (id: string) =>
      request(`/api/v1/expenses/${id}`, { method: 'DELETE' }),
    approve: (id: string, data?: { comments?: string }) =>
      request(`/api/v1/expenses/${id}/approve`, { method: 'POST', body: data ? JSON.stringify(data) : undefined }),
  },

  // ─── Lease Renewals ───────────────────────────────────────────────────────────

  leaseRenewals: {
    list: (params: { page?: number; limit?: number; status?: string } = {}) => {
      const q = new URLSearchParams()
      if (params.page) q.set('page', String(params.page))
      if (params.limit) q.set('limit', String(params.limit))
      if (params.status) q.set('status', params.status)
      return request(`/api/v1/lease-renewals?${q}`)
    },
    get: (id: string) => request(`/api/v1/lease-renewals/${id}`),
    initiate: (leaseId: string) =>
      request(`/api/v1/lease-renewals/${leaseId}/initiate`, { method: 'POST' }),
    propose: (id: string, data: { proposed_rent_amount: number; proposed_terms?: Record<string, unknown> }) =>
      request(`/api/v1/lease-renewals/${id}/propose`, { method: 'PUT', body: JSON.stringify(data) }),
    sendOffer: (id: string, data: { template_id: string; communication_type: string }) =>
      request(`/api/v1/lease-renewals/${id}/send-offer`, { method: 'POST', body: JSON.stringify(data) }),
    accept: (id: string) =>
      request(`/api/v1/lease-renewals/${id}/accept`, { method: 'PUT' }),
    reject: (id: string) =>
      request(`/api/v1/lease-renewals/${id}/reject`, { method: 'PUT' }),
    counterOffer: (id: string, data: { counter_offer_amount: number }) =>
      request(`/api/v1/lease-renewals/${id}/counter-offer`, { method: 'POST', body: JSON.stringify(data) }),
    templates: {
      list: () => request('/api/v1/renewal-templates'),
      create: (data: { template_name: string; email_subject: string; email_body: string; whatsapp_message: string; language: string }) =>
        request('/api/v1/renewal-templates', { method: 'POST', body: JSON.stringify(data) }),
      update: (id: string, data: unknown) =>
        request(`/api/v1/renewal-templates/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
      delete: (id: string) =>
        request(`/api/v1/renewal-templates/${id}`, { method: 'DELETE' }),
    },
  },

  // ─── Public (no auth) ───────────────────────────────────────────────────────

  public: {
    getSignature: (signatureId: string) =>
      fetch(`${BASE}/api/public/sign/${signatureId}`).then((r) => r.json()),
    sign: (signatureId: string) =>
      fetch(`${BASE}/api/public/sign/${signatureId}`, { method: 'POST' }).then((r) => r.json()),
  },

  // ─── Documents ──────────────────────────────────────────────────────────────

  documents: {
    // Templates
    templates: {
      list: (page: number = 1, limit: number = 20) =>
        request(`/api/v1/documents/templates?page=${page}&limit=${limit}`),
      get: (id: string) => request(`/api/v1/documents/templates/${id}`),
      create: (data: unknown) =>
        request('/api/v1/documents/templates', { method: 'POST', body: JSON.stringify(data) }),
      update: (id: string, data: unknown) =>
        request(`/api/v1/documents/templates/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
      delete: (id: string) =>
        request(`/api/v1/documents/templates/${id}`, { method: 'DELETE' }),
    },

    // Documents
    list: (entityType: string, entityId: string) =>
      request(`/api/v1/documents?entity_type=${entityType}&entity_id=${entityId}`),
    get: (id: string) => request(`/api/v1/documents/${id}`),
    create: (data: unknown) =>
      request('/api/v1/documents', { method: 'POST', body: JSON.stringify(data) }),
    delete: (id: string) =>
      request(`/api/v1/documents/${id}`, { method: 'DELETE' }),

    // Signatures
    requestSignature: (docId: string, signerName: string, signerEmail: string) =>
      request(`/api/v1/documents/${docId}/request-signature`, {
        method: 'POST',
        body: JSON.stringify({ signer_name: signerName, signer_email: signerEmail }),
      }),
    markSigned: (signatureId: string) =>
      request(`/api/v1/documents/signatures/${signatureId}/mark-signed`, { method: 'PATCH' }),
  },

  // ─── BOS24 / DLD Market Data ──────────────────────────────────────────────

  bos24: {
    search: (query: string, limit: number = 20) =>
      request('/api/v1/properties/search', {
        method: 'POST',
        body: JSON.stringify({ query, limit }),
      }),
    transactions: (filters: Record<string, string | number | undefined> = {}) => {
      const q = new URLSearchParams()
      Object.entries(filters).forEach(([k, v]) => { if (v !== undefined) q.set(k, String(v)) })
      return request(`/api/v1/properties/transactions?${q}`)
    },
    reportPdf: (payload: Record<string, unknown>) =>
      requestBlob('/api/v1/properties/report/pdf', {
        method: 'POST',
        body: JSON.stringify(payload),
      }),
  },

  // ─── Message Templates ───────────────────────────────────────────────────

  // ─── Audit Log ────────────────────────────────────────────────────────────
  auditLog: {
    list: (params: {
      entity_type?: string
      entity_id?: string
      actor_id?: string
      action?: string
      page?: number
      limit?: number
    } = {}) => {
      const q = new URLSearchParams()
      if (params.entity_type) q.set('entity_type', params.entity_type)
      if (params.entity_id) q.set('entity_id', params.entity_id)
      if (params.actor_id) q.set('actor_id', params.actor_id)
      if (params.action) q.set('action', params.action)
      if (params.page) q.set('page', String(params.page))
      if (params.limit) q.set('limit', String(params.limit))
      return request(`/api/v1/audit-logs?${q}`)
    },
  },

  messageTemplates: {
    list: (params: { page?: number; limit?: number } = {}) => {
      const q = new URLSearchParams()
      if (params.page) q.set('page', String(params.page))
      if (params.limit) q.set('limit', String(params.limit))
      return request(`/api/v1/message-templates?${q}`)
    },
    get: (id: string) => request(`/api/v1/message-templates/${id}`),
    create: (data: unknown) =>
      request('/api/v1/message-templates', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: string, data: unknown) =>
      request(`/api/v1/message-templates/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
    delete: (id: string) =>
      request(`/api/v1/message-templates/${id}`, { method: 'DELETE' }),
  },
}

export default api
