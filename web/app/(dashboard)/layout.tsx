'use client'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAuthStore } from '@/store/auth'
import { Sidebar } from '@/components/layout/Sidebar'
import { useLang } from '@/context/LangContext'

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const { token, init } = useAuthStore()
  const router = useRouter()
  const { t } = useLang()
  const [isDemo, setIsDemo] = useState(false)

  useEffect(() => { init() }, [init])
  useEffect(() => {
    if (token === null && typeof window !== 'undefined') {
      const stored = localStorage.getItem('masaar_access_token')
      if (!stored) router.replace('/login')
    }
  }, [token, router])

  useEffect(() => {
    if (typeof window !== 'undefined') {
      setIsDemo(localStorage.getItem('masaar_is_demo') === '1')
    }
  }, [])

  const handleExitDemo = () => {
    localStorage.removeItem('masaar_is_demo')
    localStorage.removeItem('masaar_access_token')
    localStorage.removeItem('masaar_refresh_token')
    router.replace('/login')
  }

  return (
    <div className="flex min-h-screen flex-col">
      {/* Demo banner */}
      {isDemo && (
        <div className="bg-amber-400 text-amber-900 px-4 py-2 flex items-center justify-between text-xs font-medium z-50">
          <div className="flex items-center gap-2">
            <span>👀</span>
            <span>
              {t(
                'أنت في وضع العرض التوضيحي — القراءة فقط، لا يمكن تعديل أي بيانات',
                'You\'re in demo mode — read-only, no data can be modified'
              )}
            </span>
          </div>
          <button
            onClick={handleExitDemo}
            className="px-3 py-1 bg-amber-900 text-amber-50 rounded-md hover:bg-amber-800 transition-colors text-xs"
          >
            {t('الخروج', 'Exit Demo')}
          </button>
        </div>
      )}

      <div className="flex flex-1 min-h-0">
        <Sidebar />
        <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
          {children}
        </div>
      </div>
    </div>
  )
}
