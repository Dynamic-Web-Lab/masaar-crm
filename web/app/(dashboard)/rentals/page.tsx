'use client'
import { useEffect, useState } from 'react'
import { Header } from '@/components/layout/Header'
import { useLang } from '@/context/LangContext'
import { api } from '@/lib/api'
import type { RentalProperty, PaginatedResult } from '@/types'

export default function RentalsPage() {
  const { t } = useLang()
  const [properties, setProperties] = useState<RentalProperty[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)

  const limit = 20

  useEffect(() => {
    loadProperties()
  }, [page])

  const loadProperties = async () => {
    setLoading(true)
    try {
      const result = (await api.rentalProperties.list({ page, limit })) as PaginatedResult<RentalProperty>
      setProperties(result.data ?? [])
      setTotal(result.total ?? 0)
    } catch (err) {
      console.error('Failed to load properties:', err)
      setProperties([])
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
      <Header title={t('العقارات للإيجار', 'Rental Properties')} />

      <div className="flex-1 overflow-auto p-6">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-lg font-semibold text-gray-800">
            {t('العقارات', 'Properties')} ({total})
          </h2>
          <button className="px-4 py-2 bg-green-600 text-white text-sm font-medium rounded-lg hover:bg-green-700 transition-colors">
            + {t('عقار جديد', 'New Property')}
          </button>
        </div>

        <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
          {properties.length === 0 ? (
            <div className="p-6 text-center text-gray-400">
              {t('لا توجد عقارات حتى الآن', 'No properties yet')}
            </div>
          ) : (
            <table className="w-full text-sm">
              <thead className="bg-gray-50 border-b border-gray-200">
                <tr>
                  <th className="px-4 py-3 text-left font-medium text-gray-700">{t('الاسم', 'Name')}</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-700">{t('النوع', 'Type')}</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-700">{t('المنطقة', 'Area')}</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-700">{t('الوحدات', 'Units')}</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-700">{t('الحالة', 'Status')}</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-700">{t('الإجراءات', 'Actions')}</th>
                </tr>
              </thead>
              <tbody>
                {properties.map((prop) => (
                  <tr key={prop.id} className="border-b border-gray-200 hover:bg-gray-50 transition-colors">
                    <td className="px-4 py-3 text-gray-800 font-medium">{prop.name}</td>
                    <td className="px-4 py-3 text-gray-600 capitalize">{prop.property_type}</td>
                    <td className="px-4 py-3 text-gray-600">{prop.area}</td>
                    <td className="px-4 py-3 text-gray-600">
                      {prop.total_occupied_units}/{prop.units_count}
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`px-2 py-1 rounded text-xs font-medium ${
                          prop.status === 'active'
                            ? 'bg-green-100 text-green-700'
                            : 'bg-gray-100 text-gray-700'
                        }`}
                      >
                        {prop.status}
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
