'use client'
import { useState, useEffect } from 'react'
import { Header } from '@/components/layout/Header'
import { api } from '@/lib/api'
import { useAuthStore } from '@/store/auth'
import { useLang } from '@/context/LangContext'
import { useRouter } from 'next/navigation'

interface CompanyData {
  company_name: string
  address: string
  phone: string
  email: string
  vat_number: string
  trn: string
  bank_name: string
  bank_account: string
  iban: string
  invoice_footer: string
  logo_url: string
}

export default function CompanySettingsPage() {
  const { user } = useAuthStore()
  const { t } = useLang()
  const router = useRouter()

  const [form, setForm] = useState<CompanyData>({
    company_name: '', address: '', phone: '', email: '',
    vat_number: '', trn: '', bank_name: '', bank_account: '',
    iban: '', invoice_footer: '', logo_url: '',
  })
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(false)

  useEffect(() => {
    if (user && user.role !== 'admin') { router.push('/dashboard'); return }
    api.settings.getCompany().then((r: any) => {
      if (r?.data) setForm(r.data)
    }).catch(() => {}).finally(() => setLoading(false))
  }, [user, router])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(''); setSuccess(false); setSubmitting(true)
    try {
      await api.settings.updateCompany(form)
      setSuccess(true)
      setTimeout(() => setSuccess(false), 3000)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to save')
    } finally { setSubmitting(false) }
  }

  const update = (k: keyof CompanyData) => (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) =>
    setForm(p => ({ ...p, [k]: e.target.value }))

  if (user?.role !== 'admin') return null

  const inputCls = 'w-full text-sm border border-gray-200 rounded-lg px-3 py-2.5 focus:outline-none focus:ring-2 focus:ring-brand-500 focus:border-transparent'
  const labelCls = 'block text-xs font-medium text-gray-600 mb-1'

  return (
    <div className="flex flex-col flex-1 min-h-0">
      <Header title={t('إعدادات الشركة', 'Company Settings')} />
      <main className="flex-1 overflow-y-auto p-6 bg-gray-50">
        <div className="max-w-2xl space-y-6">
          <div className="bg-white rounded-xl border border-gray-200 p-6">
            <h2 className="text-sm font-semibold text-gray-700 mb-4">
              {t('معلومات الشركة', 'Company Information')}
            </h2>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="col-span-2">
                  <label className={labelCls}>{t('اسم الشركة', 'Company Name')}</label>
                  <input className={inputCls} value={form.company_name} onChange={update('company_name')} />
                </div>
                <div className="col-span-2">
                  <label className={labelCls}>{t('العنوان', 'Address')}</label>
                  <textarea className={inputCls + ' resize-none'} rows={2} value={form.address} onChange={update('address')} />
                </div>
                <div>
                  <label className={labelCls}>{t('الهاتف', 'Phone')}</label>
                  <input className={inputCls} value={form.phone} onChange={update('phone')} />
                </div>
                <div>
                  <label className={labelCls}>{t('البريد الإلكتروني', 'Email')}</label>
                  <input className={inputCls} value={form.email} onChange={update('email')} />
                </div>
              </div>

              <hr className="border-gray-100" />
              <h3 className="text-sm font-semibold text-gray-700">{t('الضرائب والفواتير', 'Tax & Invoicing')}</h3>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className={labelCls}>{t('الرقم الضريبي (VAT)', 'VAT Number')}</label>
                  <input className={inputCls} value={form.vat_number} onChange={update('vat_number')} />
                </div>
                <div>
                  <label className={labelCls}>{t('TRN', 'TRN')}</label>
                  <input className={inputCls} value={form.trn} onChange={update('trn')} />
                </div>
              </div>

              <hr className="border-gray-100" />
              <h3 className="text-sm font-semibold text-gray-700">{t('معلومات البنك', 'Bank Information')}</h3>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className={labelCls}>{t('اسم البنك', 'Bank Name')}</label>
                  <input className={inputCls} value={form.bank_name} onChange={update('bank_name')} />
                </div>
                <div>
                  <label className={labelCls}>{t('رقم الحساب', 'Account Number')}</label>
                  <input className={inputCls} value={form.bank_account} onChange={update('bank_account')} />
                </div>
                <div className="col-span-2">
                  <label className={labelCls}>{t('IBAN', 'IBAN')}</label>
                  <input className={inputCls} value={form.iban} onChange={update('iban')} />
                </div>
              </div>

              <hr className="border-gray-100" />
              <div>
                <label className={labelCls}>{t('تذييل الفاتورة', 'Invoice Footer')}</label>
                <textarea className={inputCls + ' resize-none'} rows={2} value={form.invoice_footer} onChange={update('invoice_footer')} />
              </div>

              {error && <p className="text-xs text-red-500 bg-red-50 p-2 rounded">{error}</p>}
              {success && <p className="text-xs text-green-600 bg-green-50 p-2 rounded">{t('تم الحفظ', 'Saved successfully')}</p>}

              <button type="submit" disabled={submitting || loading}
                className="w-full py-2.5 bg-brand-600 text-white text-sm font-medium rounded-lg hover:bg-brand-700 disabled:opacity-50 transition-colors">
                {submitting ? t('جارٍ الحفظ...', 'Saving...') : t('حفظ الإعدادات', 'Save Settings')}
              </button>
            </form>
          </div>
        </div>
      </main>
    </div>
  )
}
