'use client'
import { useEffect, useState } from 'react'
import { Header } from '@/components/layout/Header'
import { useLang } from '@/context/LangContext'
import { api } from '@/lib/api'
import type { Payment, PaginatedResult } from '@/types'

export default function PaymentsPage() {
  const { t } = useLang()
  const [payments, setPayments] = useState<Payment[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)

  const limit = 20

  useEffect(() => {
    loadPayments()
  }, [page])

  const loadPayments = async () => {
    setLoading(true)
    try {
      const result = (await api.payments.list({ page, limit })) as PaginatedResult<Payment>
      setPayments(result.data ?? [])
      setTotal(result.total ?? 0)
    } catch (err) {
      console.error('Failed to load payments:', err)
      setPayments([])
    } finally {
      setLoading(false)
    }
  }

  const formatDate = (date: string | null) => {
    if (!date) return '-'
    return new Date(date).toLocaleDateString()
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'received':
        return 'bg-green-100 text-green-700'
      case 'pending':
        return 'bg-yellow-100 text-yellow-700'
      case 'overdue':
        return 'bg-red-100 text-red-700'
      case 'failed':
        return 'bg-red-100 text-red-700'
      case 'refunded':
        return 'bg-blue-100 text-blue-700'
      default:
        return 'bg-gray-100 text-gray-700'
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
      <Header title={t('الدفعات', 'Payments')} />

      <div className="flex-1 overflow-auto p-6">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-lg font-semibold text-gray-800">
            {t('الدفعات', 'Payments')} ({total})
          </h2>
          <button className="px-4 py-2 bg-green-600 text-white text-sm font-medium rounded-lg hover:bg-green-700 transition-colors">
            + {t('دفعة جديدة', 'New Payment')}
          </button>
        </div>

        <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
          {payments.length === 0 ? (
            <div className="p-6 text-center text-gray-400">
              {t('لا توجد دفعات حتى الآن', 'No payments yet')}
            </div>
          ) : (
            <table className="w-full text-sm">
              <thead className="bg-gray-50 border-b border-gray-200">
                <tr>
                  <th className="px-4 py-3 text-left font-medium text-gray-700">{t('المبلغ', 'Amount')}</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-700">{t('تاريخ الاستحقاق', 'Due Date')}</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-700">{t('طريقة الدفع', 'Method')}</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-700">{t('تاريخ الاستلام', 'Received Date')}</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-700">{t('الحالة', 'Status')}</th>
                  <th className="px-4 py-3 text-left font-medium text-gray-700">{t('الإجراءات', 'Actions')}</th>
                </tr>
              </thead>
              <tbody>
                {payments.map((payment) => (
                  <tr key={payment.id} className="border-b border-gray-200 hover:bg-gray-50 transition-colors">
                    <td className="px-4 py-3 text-gray-800 font-medium">
                      {payment.currency} {payment.amount.toLocaleString()}
                    </td>
                    <td className="px-4 py-3 text-gray-600 text-xs">{formatDate(payment.due_date)}</td>
                    <td className="px-4 py-3 text-gray-600 text-xs">
                      <span className="capitalize">{payment.payment_method}</span>
                    </td>
                    <td className="px-4 py-3 text-gray-600 text-xs">{formatDate(payment.received_date)}</td>
                    <td className="px-4 py-3">
                      <span className={`px-2 py-1 rounded text-xs font-medium ${getStatusColor(payment.status)}`}>
                        {t(payment.status === 'pending' ? 'قيد الانتظار' :
                           payment.status === 'received' ? 'مستلمة' :
                           payment.status === 'overdue' ? 'متأخرة' :
                           payment.status === 'failed' ? 'فشلت' :
                           payment.status === 'refunded' ? 'مرتجعة' : payment.status,
                           payment.status === 'pending' ? 'Pending' :
                           payment.status === 'received' ? 'Received' :
                           payment.status === 'overdue' ? 'Overdue' :
                           payment.status === 'failed' ? 'Failed' :
                           payment.status === 'refunded' ? 'Refunded' : payment.status)}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-xs text-gray-500">
                      <button className="text-blue-600 hover:text-blue-800">{t('تفاصيل', 'View')}</button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>

        {/* Pagination */}
        {total > limit && (
          <div className="flex items-center justify-between mt-4 text-sm text-gray-600">
            <button
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
              className="px-3 py-1 disabled:opacity-50 hover:bg-gray-100 rounded"
            >
              {t('السابق', 'Previous')}
            </button>
            <span>{t(`صفحة ${page}`, `Page ${page}`)}</span>
            <button
              onClick={() => setPage(p => p + 1)}
              disabled={page * limit >= total}
              className="px-3 py-1 disabled:opacity-50 hover:bg-gray-100 rounded"
            >
              {t('التالي', 'Next')}
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
