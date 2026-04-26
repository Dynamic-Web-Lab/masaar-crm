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
  auth: {
    login: (email: string, password: string) =>
      request('/api/v1/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
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
}
