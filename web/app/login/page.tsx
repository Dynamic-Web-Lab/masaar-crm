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
  const [demoEnabled, setDemoEnabled] = useState(false)
  const { setSession, init, token } = useAuthStore()
  const router = useRouter()
  const { lang, setLang, t } = useLang()

  useEffect(() => { init() }, [init])
  useEffect(() => {
    if (token && isLoggedIn()) router.replace('/pipeline')
  }, [token, router])
  useEffect(() => {
    api.health().then(res => setDemoEnabled(res.demo_enabled)).catch(() => {})
  }, [])

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

  const handleDemoLogin = async () => {
    setError('')
    setLoading(true)
    try {
      const res = await api.auth.demo() as LoginResponse
      setSession(res.access_token, res.refresh_token, res.user)
      localStorage.setItem('masaar_is_demo', '1')
      router.push('/pipeline')
    } catch (err: any) {
      setError(err.message || t('حدث خطأ', 'Something went wrong'))
    } finally {
      setLoading(false)
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

        {demoEnabled && (
          <div className="bg-amber-50 rounded-2xl border-2 border-dashed border-amber-300 p-4 mb-6 text-center">
            <p className="text-sm text-amber-900 mb-3">
              {t('جرّب النظام بدون الحاجة لكلمة مرور', 'Try the platform without logging in')}
            </p>
            <button
              onClick={handleDemoLogin}
              disabled={loading}
              className="w-full py-2 bg-amber-300 text-amber-900 font-medium rounded-lg text-sm hover:bg-amber-400 transition-colors disabled:opacity-60"
            >
              {loading
                ? t('جاري المعالجة...', 'Processing...')
                : t('جرّب البيئة التجريبية', 'Try Demo')}
            </button>
          </div>
        )}

        {magicSent ? (
          <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-8 space-y-4 text-center">
            <div className="text-brand-600 text-4xl mb-2">📧</div>
            <h3 className="text-lg font-semibold text-gray-900">
              {t('تم إرسال رابط الدخول', 'Magic link sent!')}
            </h3>
            <p className="text-sm text-gray-600">
              {t(
                'تحقق من بريدك الإلكتروني وانقر على الرابط لتسجيل الدخول',
                'Check your email and click the link to sign in'
              )}
            </p>
            <button
              onClick={() => { setMagicSent(false); setError('') }}
              className="text-sm text-brand-600 hover:text-brand-700"
            >
              {t('العودة لتسجيل الدخول', 'Back to sign in')}
            </button>
          </div>
        ) : (
          <form
            onSubmit={mode === 'password' ? handlePasswordLogin : handleMagicLinkRequest}
            className="bg-white rounded-2xl shadow-sm border border-gray-100 p-8 space-y-4"
          >
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

            {mode === 'password' && (
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
            )}

            {error && (
              <p className="text-red-500 text-xs bg-red-50 px-3 py-2 rounded-lg">{error}</p>
            )}

            <button
              type="submit"
              disabled={loading}
              className="w-full py-2.5 bg-brand-600 text-white font-medium rounded-lg text-sm hover:bg-brand-700 transition-colors disabled:opacity-60"
            >
              {loading
                ? t('جاري المعالجة...', 'Processing...')
                : mode === 'password'
                  ? t('تسجيل الدخول', 'Sign in')
                  : t('إرسال رابط الدخول', 'Send magic link')}
            </button>

            <div className="text-center">
              <button
                type="button"
                onClick={() => { setMode(mode === 'password' ? 'magic-link' : 'password'); setError('') }}
                className="text-xs text-brand-600 hover:text-brand-700"
              >
                {mode === 'password'
                  ? t('تسجيل الدخول بدون كلمة مرور', 'Sign in with magic link')
                  : t('العودة لتسجيل الدخول بالكلمة', 'Back to password login')}
              </button>
            </div>
          </form>
        )}

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
