'use client'
import { useEffect, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { Header } from '@/components/layout/Header'
import { api } from '@/lib/api'
import { useLang } from '@/context/LangContext'
import type { WhatsAppThread, WhatsAppMessage } from '@/types'
import clsx from 'clsx'

interface Suggestion {
  ar: string
  en: string
}

export default function ThreadPage() {
  const { id } = useParams<{ id: string }>()
  const [thread, setThread] = useState<WhatsAppThread | null>(null)
  const [messages, setMessages] = useState<WhatsAppMessage[]>([])
  const [summary, setSummary] = useState('')
  const [summarizing, setSummarizing] = useState(false)
  const [loading, setLoading] = useState(true)
  const { lang, t } = useLang()
  const router = useRouter()

  // Reply suggestions
  const [suggestions, setSuggestions] = useState<Suggestion[]>([])
  const [suggestionsLoading, setSuggestionsLoading] = useState(false)
  const [showSuggestions, setShowSuggestions] = useState(false)
  const [copiedIdx, setCopiedIdx] = useState<number | null>(null)

  // Inline translation — map of message id → translated text
  const [translations, setTranslations] = useState<Record<string, string>>({})
  const [translatingId, setTranslatingId] = useState<string | null>(null)

  useEffect(() => {
    if (!id) return
    Promise.all([
      api.threads.get(id) as Promise<WhatsAppThread>,
      api.threads.messages(id) as Promise<WhatsAppMessage[]>,
    ]).then(([t, msgs]) => {
      setThread(t)
      setMessages(Array.isArray(msgs) ? msgs : [])
    }).finally(() => setLoading(false))
  }, [id])

  const handleSummarize = async () => {
    setSummarizing(true)
    try {
      const res = await api.ai.summarize(id) as { summary: string }
      setSummary(res.summary)
    } catch {
      setSummary(t('حدث خطأ أثناء التلخيص', 'Summarization failed'))
    } finally {
      setSummarizing(false)
    }
  }

  const handleReplySuggestions = async () => {
    if (showSuggestions && suggestions.length > 0) {
      setShowSuggestions(false)
      return
    }
    setSuggestionsLoading(true)
    setShowSuggestions(true)
    try {
      const res = await api.ai.replySuggestions(id) as { suggestions: Suggestion[] }
      setSuggestions(res.suggestions || [])
    } catch {
      setSuggestions([])
    } finally {
      setSuggestionsLoading(false)
    }
  }

  const handleCopySuggestion = (text: string, idx: number) => {
    navigator.clipboard.writeText(text).catch(() => {})
    setCopiedIdx(idx)
    setTimeout(() => setCopiedIdx(null), 2000)
  }

  const handleTranslate = async (msg: WhatsAppMessage) => {
    if (translations[msg.id]) {
      // Toggle off
      setTranslations((prev) => {
        const next = { ...prev }
        delete next[msg.id]
        return next
      })
      return
    }
    setTranslatingId(msg.id)
    try {
      const targetLang = lang === 'ar' ? 'en' : 'ar'
      const res = await api.ai.translate(msg.body, targetLang as 'ar' | 'en') as { translated: string }
      setTranslations((prev) => ({ ...prev, [msg.id]: res.translated }))
    } catch {
      // silently ignore
    } finally {
      setTranslatingId(null)
    }
  }

  const handleClose = async () => {
    await api.threads.close(id).catch(() => {})
    router.push('/inbox')
  }

  if (loading) {
    return (
      <div className="flex-1 flex items-center justify-center text-gray-400 text-sm">
        {t('جاري التحميل...', 'Loading...')}
      </div>
    )
  }

  return (
    <div className="flex flex-col flex-1 overflow-hidden">
      <Header title={thread?.contact?.full_name ?? t('المحادثة', 'Thread')} />

      {/* Thread meta bar */}
      <div className="bg-white border-b border-gray-100 px-6 py-3 flex items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-full bg-brand-100 text-brand-700 font-semibold text-sm flex items-center justify-center">
            {thread?.contact?.full_name?.[0]?.toUpperCase() ?? '?'}
          </div>
          <div>
            <p className="font-medium text-sm text-gray-900">{thread?.contact?.full_name}</p>
            <p className="text-xs text-gray-400">{thread?.contact?.phone_wa}</p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={handleSummarize}
            disabled={summarizing || messages.length === 0}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium bg-indigo-50 text-indigo-700 rounded-lg hover:bg-indigo-100 transition-colors disabled:opacity-50"
          >
            {summarizing ? '⏳' : '✨'}
            {summarizing
              ? t('جاري التلخيص...', 'Summarizing...')
              : t('تلخيص بالذكاء الاصطناعي', 'AI Summarize')}
          </button>

          <button
            onClick={handleReplySuggestions}
            disabled={suggestionsLoading || messages.length === 0}
            className={clsx(
              'flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg transition-colors disabled:opacity-50',
              showSuggestions
                ? 'bg-emerald-600 text-white hover:bg-emerald-700'
                : 'bg-emerald-50 text-emerald-700 hover:bg-emerald-100'
            )}
          >
            {suggestionsLoading ? '⏳' : '💬'}
            {suggestionsLoading
              ? t('جاري التحليل...', 'Thinking...')
              : t('اقتراحات الرد', 'Reply Suggestions')}
          </button>

          {thread?.thread_status !== 'closed' && (
            <button
              onClick={handleClose}
              className="px-3 py-1.5 text-xs font-medium bg-gray-100 text-gray-600 rounded-lg hover:bg-gray-200 transition-colors"
            >
              {t('إغلاق', 'Close thread')}
            </button>
          )}
        </div>
      </div>

      {/* AI Summary */}
      {summary && (
        <div className="mx-6 mt-4 p-4 bg-indigo-50 border border-indigo-100 rounded-xl text-sm text-indigo-800">
          <p className="font-semibold text-xs text-indigo-500 mb-1">{t('ملخص الذكاء الاصطناعي', 'AI Summary')}</p>
          <p>{summary}</p>
        </div>
      )}

      {/* Reply Suggestions Panel */}
      {showSuggestions && (
        <div className="mx-6 mt-3 p-4 bg-emerald-50 border border-emerald-100 rounded-xl">
          <p className="font-semibold text-xs text-emerald-600 mb-3">
            {t('اقتراحات الرد بالذكاء الاصطناعي', 'AI Reply Suggestions')}
          </p>
          {suggestionsLoading ? (
            <div className="flex items-center gap-2 text-xs text-emerald-600">
              <div className="w-3 h-3 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin" />
              {t('يفكر الذكاء الاصطناعي...', 'AI is thinking...')}
            </div>
          ) : suggestions.length === 0 ? (
            <p className="text-xs text-gray-400">{t('لم تتوفر اقتراحات', 'No suggestions available')}</p>
          ) : (
            <div className="space-y-2">
              {suggestions.map((s, idx) => {
                const text = lang === 'ar' ? s.ar : s.en
                const subText = lang === 'ar' ? s.en : s.ar
                return (
                  <div key={idx} className="bg-white border border-emerald-100 rounded-lg p-3">
                    <p className="text-sm text-gray-800 mb-1">{text}</p>
                    <p className="text-xs text-gray-400 italic mb-2">{subText}</p>
                    <button
                      onClick={() => handleCopySuggestion(text, idx)}
                      className="text-xs px-2 py-1 bg-emerald-600 text-white rounded hover:bg-emerald-700 transition-colors"
                    >
                      {copiedIdx === idx
                        ? t('تم النسخ ✓', 'Copied ✓')
                        : t('نسخ', 'Copy')}
                    </button>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      )}

      {/* Messages */}
      <div className="flex-1 overflow-y-auto px-6 py-4 space-y-3">
        {messages.length === 0 ? (
          <p className="text-center text-sm text-gray-400 mt-10">
            {t('لا توجد رسائل', 'No messages yet')}
          </p>
        ) : (
          messages.map((msg) => {
            const isInbound = msg.direction === 'inbound'
            const translatedText = translations[msg.id]
            const isTranslating = translatingId === msg.id
            return (
              <div
                key={msg.id}
                className={clsx('flex flex-col', isInbound ? 'items-start' : 'items-end')}
              >
                <div
                  className={clsx(
                    'max-w-xs md:max-w-md px-4 py-2.5 rounded-2xl text-sm',
                    isInbound
                      ? 'bg-white border border-gray-100 text-gray-800 rounded-tl-sm'
                      : 'bg-brand-600 text-white rounded-tr-sm'
                  )}
                >
                  <p className="leading-relaxed">{msg.body}</p>
                  {translatedText && (
                    <p className={clsx(
                      'text-xs mt-1.5 pt-1.5 border-t leading-relaxed italic',
                      isInbound ? 'text-gray-500 border-gray-100' : 'text-blue-100 border-blue-400'
                    )}>
                      {translatedText}
                    </p>
                  )}
                  <div className="flex items-center justify-between mt-1 gap-2">
                    <p className={clsx('text-[10px]', isInbound ? 'text-gray-400' : 'text-blue-100')}>
                      {new Date(msg.sent_at).toLocaleTimeString(lang === 'ar' ? 'ar-AE' : 'en-AE', {
                        hour: '2-digit', minute: '2-digit'
                      })}
                    </p>
                    <button
                      onClick={() => handleTranslate(msg)}
                      disabled={isTranslating}
                      className={clsx(
                        'text-[10px] px-1.5 py-0.5 rounded transition-colors disabled:opacity-50',
                        isInbound
                          ? 'text-gray-400 hover:text-indigo-600 hover:bg-indigo-50'
                          : 'text-blue-200 hover:text-white hover:bg-blue-500'
                      )}
                    >
                      {isTranslating ? '...' : translatedText ? t('إخفاء', 'Hide') : t('ترجمة', 'Translate')}
                    </button>
                  </div>
                </div>
              </div>
            )
          })
        )}
      </div>

      {/* Read-only notice */}
      <div className="px-6 py-3 bg-gray-50 border-t border-gray-100 text-center">
        <p className="text-xs text-gray-400">
          {t(
            'إرسال الرسائل متوفر في الإصدار Enterprise',
            'Sending messages is available in the Enterprise edition'
          )}
        </p>
      </div>
    </div>
  )
}
