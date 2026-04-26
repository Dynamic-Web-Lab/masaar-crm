'use client'
import { useEffect, useState } from 'react'
import { Header } from '@/components/layout/Header'
import { useLang } from '@/context/LangContext'
import { api } from '@/lib/api'
import type { Tenant, PaginatedResult } from '@/types'

export default function TenantsPage() {
  const { t } = useLang()
  const [tenants, setTenants] = useState<Tenant[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)

  const limit = 20

  useEffect(() => {
    loadTenants()
  }, [page])

  const loadTenants = async () => {
    setLoading(true)
    try {
      const result = (await api.tenants.list({ page, limit })) as PaginatedResult<Tenant>
      setTenants(result.data ?? [])
      setTotal(result.total ?? 0)
    } catch (err) {
      console.error('Failed to load tenants:', err)
      setTenants([])
    } finally {
      setLoading(false)
    }
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
      <Header title={t('المستأجرون', 'Tenants')} />

      <div className="flex-1 overflow-auto p-6">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-lg font-semibold text-gray-800">
            {t('المستأجرون', 'Tenants')} ({total})
          </h2>
          <button className="px-4 py-2 bg-green-600 text-white text-sm font-medium rounded-lg hover:bg-green-700 transition-colors">
            + {t('مستأجر جديد', 'New Tenant')}
          </button>
        </div>

        <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
          {tenants.length === 0 ? (
            <div className="p-6 text-center text-gray-400">
              {t('لا يوجد مستأجرون حتى الآن', 'No tenants yet')}
            </div>
          ) : (
            <table className="w-full text-sm">
              <thead className="bg-gray-50 border-b border-gray-200">
                <tr>
                  <th className="px-4 py-3 text-left font-medium text-gray-700">{t('الاسم', 'Name')}</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-700">{t('البريد الإلكتروني', 'Email')}</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-700">{t('الهاتف', 'Phone')}</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-700">{t('التحقق', 'Verified')}</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-700">{t('الحالة', 'Status')}</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-700">{t('الإجراءات', 'Actions')}</th>
                </tr>
              </thead>
              <tbody>
                {tenants.map((tenant) => (
                  <tr key={tenant.id} className="border-b border-gray-200 hover:bg-gray-50 transition-colors">
                    <td className="px-4 py-3 text-gray-800 font-medium">{tenant.full_name_en}</td>
                    <td className="px-4 py-3 text-gray-600 text-xs">{tenant.email}</td>
                    <td className="px-4 py-3 text-gray-600">{tenant.phone_wa}</td>
                    <td className="px-4 py-3">
                      {tenant.is_verified ? (
                        <span className="px-2 py-1 bg-green-100 text-green-700 rounded text-xs font-medium">
                          ✓ {t('محقق', 'Verified')}
                        </span>
                      ) : (
                        <span className="px-2 py-1 bg-yellow-100 text-yellow-700 rounded text-xs font-medium">
                          {t('قيد الانتظار', 'Pending')}
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`px-2 py-1 rounded text-xs font-medium ${
                          tenant.status === 'active'
                            ? 'bg-green-100 text-green-700'
                            : tenant.status === 'inactive'
                              ? 'bg-gray-100 text-gray-700'
                              : 'bg-red-100 text-red-700'
                        }`}
                      >
                        {tenant.status}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <button className="text-brand-600 hover:text-brand-700 text-sm font-medium">
                        {t('عرض', 'View')}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>

        {/* Pagination */}
        {total > limit && (
          <div className="flex justify-center gap-2 mt-6">
            <button
              onClick={() => setPage(Math.max(1, page - 1))}
              disabled={page === 1}
              className="px-4 py-2 border border-gray-200 rounded-lg disabled:opacity-50 hover:bg-gray-50"
            >
              {t('السابق', 'Prev')}
            </button>
            <div className="flex items-center px-4 py-2 text-sm text-gray-600">
              {page} / {Math.ceil(total / limit)}
            </div>
            <button
              onClick={() => setPage(page + 1)}
              disabled={page * limit >= total}
              className="px-4 py-2 border border-gray-200 rounded-lg disabled:opacity-50 hover:bg-gray-50"
            >
              {t('التالي', 'Next')}
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
