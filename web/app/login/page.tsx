'use client'
import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { api } from '@/lib/api'
import { isLoggedIn } from '@/lib/auth'
import { useAuthStore } from '@/store/auth'
import { useLang } from '@/context/LangContext'
import type { LoginResponse } from '@/types'

export default function LoginPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [demoLoading, setDemoLoading] = useState(false)
  const [demoEnabled, setDemoEnabled] = useState(false)
  const { setSession, init, token } = useAuthStore()
  const router = useRouter()
  const { lang, setLang, t } = useLang()

  useEffect(() => { init() }, [init])
  useEffect(() => {
    if (token && isLoggedIn()) router.replace('/pipeline')
  }, [token, router])

  // Check if demo mode is enabled on this server
  useEffect(() => {
    api.health().then((h) => setDemoEnabled(h.demo_enabled)).catch(() => {})
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const res = await api.auth.login(email, password) as LoginResponse
      setSession(res.access_token, res.refresh_token, res.user)
      if (res.user.lang_pref) setLang(res.user.lang_pref)
      router.push('/pipeline')
    } catch (err: any) {
      setError(err.message || t('حدث خطأ', 'Something went wrong'))
    } finally {
      setLoading(false)
    }
  }

  const handleDemo = async () => {
    setError('')
    setDemoLoading(true)
    try {
      const res = await api.auth.demo() as LoginResponse & { demo?: boolean }
      setSession(res.access_token, res.refresh_token, res.user)
      if (res.demo) localStorage.setItem('masaar_is_demo', '1')
      router.push('/pipeline')
    } catch (err: any) {
      setError(err.message || t('حدث خطأ', 'Something went wrong'))
    } finally {
      setDemoLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
      <div className="w-full max-w-sm">
        {/* Logo */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center gap-2 mb-2">
            <span className="text-brand-600 text-3xl font-bold">M</span>
            <span className="text-2xl font-bold text-gray-900">
              {lang === 'ar' ? 'مسار' : 'Masaar'}
            </span>
          </div>
          <p className="text-gray-500 text-sm">
            {t('نظام إدارة علاقات العملاء الإماراتي', 'CRM built for the UAE')}
          </p>
        </div>

        <form onSubmit={handleSubmit} className="bg-white rounded-2xl shadow-sm border border-gray-100 p-8 space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              {t('البريد الإلكتروني', 'Email')}
            </label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              autoComplete="email"
              className="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-brand-500 focus:border-transparent"
              placeholder={t('أدخل بريدك الإلكتروني', 'Enter your email')}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              {t('كلمة المرور', 'Password')}
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              autoComplete="current-password"
              className="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-brand-500 focus:border-transparent"
              placeholder="••••••••"
            />
          </div>

          {error && (
            <p className="text-red-500 text-xs bg-red-50 px-3 py-2 rounded-lg">{error}</p>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full py-2.5 bg-brand-600 text-white font-medium rounded-lg text-sm hover:bg-brand-700 transition-colors disabled:opacity-60"
          >
            {loading
              ? t('جاري تسجيل الدخول...', 'Signing in...')
              : t('تسجيل الدخول', 'Sign in')}
          </button>

          {/* Demo login — only shown when server has DEMO_MODE=true */}
          {demoEnabled && (
            <>
              <div className="relative">
                <div className="absolute inset-0 flex items-center">
                  <div className="w-full border-t border-gray-100" />
                </div>
                <div className="relative flex justify-center text-xs">
                  <span className="px-2 bg-white text-gray-400">
                    {t('أو', 'or')}
                  </span>
                </div>
              </div>
              <button
                type="button"
                onClick={handleDemo}
                disabled={demoLoading}
                className="w-full py-2.5 border-2 border-dashed border-amber-300 text-amber-700 font-medium rounded-lg text-sm hover:bg-amber-50 transition-colors disabled:opacity-60 flex items-center justify-center gap-2"
              >
                {demoLoading ? (
                  <div className="w-4 h-4 border-2 border-amber-500 border-t-transparent rounded-full animate-spin" />
                ) : (
                  <span>👀</span>
                )}
                {demoLoading
                  ? t('جاري الدخول...', 'Loading demo...')
                  : t('تجربة العرض التوضيحي', 'Try Demo — read-only')}
              </button>
            </>
          )}
        </form>

        {/* Lang toggle */}
        <div className="mt-4 text-center">
          <button
            onClick={() => setLang(lang === 'ar' ? 'en' : 'ar')}
            className="text-xs text-gray-400 hover:text-gray-600"
          >
            {lang === 'ar' ? 'Switch to English' : 'التبديل إلى العربية'}
          </button>
        </div>

        <p className="mt-6 text-center text-xs text-gray-400">
          {t('مصنوع بعناية للسوق الإماراتي', 'Built for the UAE market')} ·{' '}
          <a href="https://dynamicweblab.com" className="hover:text-brand-600 transition-colors">
            Dynamic Web Lab
          </a>
        </p>
      </div>
    </div>
  )
}
