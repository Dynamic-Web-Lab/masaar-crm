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
  },

  // ─── Leads ──────────────────────────────────────────────────────────────────

  leads: {
    kanban: () => request('/api/v1/leads'),
    get: (id: string) => request(`/api/v1/leads/${id}`),
    create: (data: unknown) =>
      request('/api/v1/leads', { method: 'POST', body: JSON.stringify(data) }),
    updateStage: (id: string, stage: string) =>
      request(`/api/v1/leads/${id}/stage`, {
        method: 'PATCH',
        body: JSON.stringify({ stage }),
      }),
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
  },

  // ─── WhatsApp ───────────────────────────────────────────────────────────────

  threads: {
    list: (params: { status?: string; page?: number; limit?: number } = {}) => {
      const q = new URLSearchParams()
      if (params.status) q.set('status', params.status)
      if (params.page) q.set('page', String(params.page))
      if (params.limit) q.set('limit', String(params.limit))
      return request(`/api/v1/threads?${q}`)
    },
    get: (id: string) => request(`/api/v1/threads/${id}`),
    messages: (id: string) => request(`/api/v1/threads/${id}/messages`),
    close: (id: string) =>
      request(`/api/v1/threads/${id}/close`, { method: 'POST' }),
  },

  // ─── AI ─────────────────────────────────────────────────────────────────────

  ai: {
    summarize: (threadId: string) =>
      request(`/api/v1/ai/summarize/${threadId}`, { method: 'POST' }),
  },

  // ─── Deals ──────────────────────────────────────────────────────────────────

  deals: {
    list: (params: { page?: number; limit?: number } = {}) => {
      const q = new URLSearchParams()
      if (params.page) q.set('page', String(params.page))
      if (params.limit) q.set('limit', String(params.limit))
      return request(`/api/v1/deals?${q}`)
    },
    get: (id: string) => request(`/api/v1/deals/${id}`),
    create: (data: unknown) =>
      request('/api/v1/deals', { method: 'POST', body: JSON.stringify(data) }),
    updateStage: (id: string, stage: string) =>
      request(`/api/v1/deals/${id}/stage`, {
        method: 'PATCH',
        body: JSON.stringify({ stage }),
      }),
    invoices: (id: string) => request(`/api/v1/deals/${id}/invoices`),
  },

  // ─── Invoices ────────────────────────────────────────────────────────────────

  invoices: {
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
    list: (limit?: number, offset?: number) => {
      const q = new URLSearchParams()
      if (limit) q.set('limit', String(limit))
      if (offset) q.set('offset', String(offset))
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
    list: (limit?: number, offset?: number, status?: string) => {
      const q = new URLSearchParams()
      if (limit) q.set('limit', String(limit))
      if (offset) q.set('offset', String(offset))
      if (status) q.set('status', status)
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
}

export default api
