'use client'
import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { api } from '@/lib/api'
import { isLoggedIn } from '@/lib/auth'
import { useAuthStore } from '@/store/auth'
import { useLang } from '@/context/LangContext'
import type { LoginResponse } from '@/types'

type LoginMode = 'password' | 'magic-link'

export default function LoginPage() {
  const [mode, setMode] = useState<LoginMode>('password')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [magicSent, setMagicSent] = useState(false)
  const { setSession, init, token } = useAuthStore()
  const router = useRouter()
  const { lang, setLang, t } = useLang()

  useEffect(() => { init() }, [init])
  useEffect(() => {
    if (token && isLoggedIn()) router.replace('/pipeline')
  }, [token, router])

  const handlePasswordLogin = async (e: React.FormEvent) => {
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

  const handleMagicLinkRequest = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await api.auth.requestMagicLink(email, lang)
      setMagicSent(true)
    } catch (err: any) {
      setError(err.message || t('حدث خطأ', 'Something went wrong'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="relative min-h-screen flex items-center justify-center px-4 bg-surface-50 overflow-hidden">
      {/* Soft ambient gradient backdrop */}
      <div aria-hidden="true" className="pointer-events-none absolute inset-0 -z-10">
        <div className="absolute -top-32 -start-32 w-[420px] h-[420px] rounded-full bg-primary-200/40 blur-3xl" />
        <div className="absolute -bottom-32 -end-32 w-[420px] h-[420px] rounded-full bg-gold-200/40 blur-3xl" />
      </div>

      <div className="w-full max-w-sm">
        {/* Logo */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center gap-2.5 mb-3">
            <span className="inline-flex items-center justify-center w-10 h-10 rounded-2xl bg-primary-600 text-white text-lg font-bold shadow-card">
              M
            </span>
            <span className="text-2xl font-bold tracking-tight text-surface-900">
              {lang === 'ar' ? 'مسار' : 'Masaar'}
            </span>
          </div>
          <p className="text-surface-500 text-sm">
            {t('نظام إدارة علاقات العملاء الإماراتي', 'CRM built for the UAE')}
          </p>
        </div>

        {magicSent ? (
          <div className="bg-white rounded-2xl shadow-card border border-surface-200/70 p-8 space-y-4 text-center">
            <div className="mx-auto w-12 h-12 rounded-2xl bg-primary-50 text-primary-600 flex items-center justify-center mb-1">
              <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.75} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
              </svg>
            </div>
            <h3 className="text-lg font-semibold text-surface-900">
              {t('تم إرسال رابط الدخول', 'Magic link sent!')}
            </h3>
            <p className="text-sm text-surface-500">
              {t(
                'تحقق من بريدك الإلكتروني وانقر على الرابط لتسجيل الدخول',
                'Check your email and click the link to sign in'
              )}
            </p>
            <button
              onClick={() => { setMagicSent(false); setError('') }}
              className="text-sm font-medium text-primary-600 hover:text-primary-700"
            >
              {t('العودة لتسجيل الدخول', 'Back to sign in')}
            </button>
          </div>
        ) : (
          <form
            onSubmit={mode === 'password' ? handlePasswordLogin : handleMagicLinkRequest}
            className="bg-white rounded-2xl shadow-card border border-surface-200/70 p-7 space-y-4"
          >
            <div>
              <label className="block text-[13px] font-medium text-surface-700 mb-1.5">
                {t('البريد الإلكتروني', 'Email')}
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                autoComplete="email"
                className="w-full px-3.5 py-2.5 border border-surface-200 rounded-xl text-sm bg-white placeholder:text-surface-400 focus:outline-none focus:ring-2 focus:ring-primary-500/40 focus:border-primary-400 transition-shadow"
                placeholder={t('أدخل بريدك الإلكتروني', 'Enter your email')}
              />
            </div>

            {mode === 'password' && (
              <div>
                <label className="block text-[13px] font-medium text-surface-700 mb-1.5">
                  {t('كلمة المرور', 'Password')}
                </label>
                <input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  autoComplete="current-password"
                  className="w-full px-3.5 py-2.5 border border-surface-200 rounded-xl text-sm bg-white placeholder:text-surface-400 focus:outline-none focus:ring-2 focus:ring-primary-500/40 focus:border-primary-400 transition-shadow"
                  placeholder="••••••••"
                />
              </div>
            )}

            {error && (
              <p className="text-red-600 text-xs bg-red-50 border border-red-100 px-3 py-2 rounded-lg">{error}</p>
            )}

            <button
              type="submit"
              disabled={loading}
              className="w-full py-2.5 bg-primary-600 text-white font-medium rounded-xl text-sm shadow-card hover:bg-primary-700 hover:shadow-card-hover transition-all duration-200 ease-soft disabled:opacity-60 disabled:cursor-not-allowed"
            >
              {loading
                ? t('جاري المعالجة...', 'Processing...')
                : mode === 'password'
                  ? t('تسجيل الدخول', 'Sign in')
                  : t('إرسال رابط الدخول', 'Send magic link')}
            </button>

            <div className="text-center pt-1">
              <button
                type="button"
                onClick={() => { setMode(mode === 'password' ? 'magic-link' : 'password'); setError('') }}
                className="text-xs font-medium text-primary-600 hover:text-primary-700"
              >
                {mode === 'password'
                  ? t('تسجيل الدخول بدون كلمة مرور', 'Sign in with magic link')
                  : t('العودة لتسجيل الدخول بالكلمة', 'Back to password login')}
              </button>
            </div>
          </form>
        )}

        {/* Lang toggle */}
        <div className="mt-5 text-center">
          <button
            onClick={() => setLang(lang === 'ar' ? 'en' : 'ar')}
            className="text-xs text-surface-500 hover:text-surface-900 transition-colors"
          >
            {lang === 'ar' ? 'Switch to English' : 'التبديل إلى العربية'}
          </button>
        </div>

        <p className="mt-6 text-center text-[11px] text-surface-400">
          {t('مصنوع بعناية للسوق الإماراتي', 'Built for the UAE market')} ·{' '}
          <a href="https://dynamicweblab.com" className="hover:text-primary-600 transition-colors">
            Dynamic Web Lab
          </a>
        </p>
      </div>
    </div>
  )
}
