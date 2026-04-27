'use client'
import { create } from 'zustand'
import type { AuthUser } from '@/types'
import { getUser, getToken, saveSession, clearSession, updateUser, isLoggedIn } from '@/lib/auth'

interface AuthState {
  user: AuthUser | null
  token: string | null
  setSession: (token: string, refreshToken: string, user: AuthUser) => void
  updateUser: (updates: Partial<AuthUser>) => void
  logout: () => void
  init: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: null,

  init: () => {
    if (isLoggedIn()) {
      set({ user: getUser(), token: getToken() })
    } else {
      clearSession()
      set({ user: null, token: null })
    }
  },

  setSession: (token, refreshToken, user) => {
    saveSession(token, refreshToken, user)
    set({ user, token })
  },

  updateUser: (updates) => {
    updateUser(updates)
    set((state) => ({ user: state.user ? { ...state.user, ...updates } : null }))
  },

  logout: () => {
    clearSession()
    set({ user: null, token: null })
  },
}))
