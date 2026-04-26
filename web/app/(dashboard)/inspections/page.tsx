'use client';

import { useEffect, useState } from 'react';
import api from '@/lib/api';
import { useLang } from '@/context/LangContext';

interface Inspection {
  id: string;
  property_id: string;
  inspection_type: string;
  scheduled_date: string;
  status: string;
  severity_level?: string;
}

const labels = {
  en: {
    scheduled: 'Scheduled',
    inProgress: 'In Progress',
    completed: 'Completed',
    cancelled: 'Cancelled',
    green: 'Good',
    yellow: 'Minor Issues',
    red: 'Critical',
    noInspections: 'No inspections found',
    loading: 'Loading...',
  },
  ar: {
    scheduled: 'مجدولة',
    inProgress: 'قيد التنفيذ',
    completed: 'مكتملة',
    cancelled: 'ملغاة',
    green: 'جيد',
    yellow: 'مشاكل بسيطة',
    red: 'حرج',
    noInspections: 'لم يتم العثور على فحوصات',
    loading: 'جاري التحميل...',
  },
};

export default function InspectionsPage() {
  const { lang } = useLang();
  const t = labels[lang as keyof typeof labels];
  const [inspections, setInspections] = useState<Inspection[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(0);
  const limit = 10;

  useEffect(() => {
    loadInspections();
  }, [page]);

  const loadInspections = async () => {
    setLoading(true);
    try {
      const res = (await api.inspection.list(limit, page * limit)) as any;
      setInspections(res.data || []);
    } catch (error) {
      console.error('Failed to fetch inspections:', error);
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return 'bg-green-100 text-green-800';
      case 'in_progress':
        return 'bg-blue-100 text-blue-800';
      case 'scheduled':
        return 'bg-yellow-100 text-yellow-800';
      case 'cancelled':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getSeverityColor = (severity?: string) => {
    switch (severity) {
      case 'green':
        return 'text-green-600';
      case 'yellow':
        return 'text-yellow-600';
      case 'red':
        return 'text-red-600';
      default:
        return 'text-gray-600';
    }
  };

  if (loading && page === 0) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-gray-500">{t.loading}</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
            <thead className="bg-gray-50 dark:bg-gray-900">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 dark:text-gray-300 uppercase tracking-wider">
                  Inspection Type
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 dark:text-gray-300 uppercase tracking-wider">
                  Scheduled Date
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 dark:text-gray-300 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 dark:text-gray-300 uppercase tracking-wider">
                  Severity
                </th>
              </tr>
            </thead>
            <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
              {inspections.length === 0 ? (
                <tr>
                  <td colSpan={4} className="px-6 py-4 text-center text-gray-500">
                    {t.noInspections}
                  </td>
                </tr>
              ) : (
                inspections.map((inspection) => (
                  <tr key={inspection.id} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">
                      {inspection.inspection_type}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600 dark:text-gray-400">
                      {new Date(inspection.scheduled_date).toLocaleDateString()}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`px-3 py-1 rounded-full text-xs font-medium ${getStatusColor(inspection.status)}`}>
                        {inspection.status}
                      </span>
                    </td>
                    <td className={`px-6 py-4 whitespace-nowrap text-sm font-medium ${getSeverityColor(inspection.severity_level)}`}>
                      {inspection.severity_level ? t[inspection.severity_level as keyof typeof t] : '-'}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {inspections.length > 0 && (
        <div className="flex items-center justify-between">
          <button
            onClick={() => setPage(Math.max(0, page - 1))}
            disabled={page === 0}
            className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md disabled:opacity-50"
          >
            Previous
          </button>
          <p className="text-sm text-gray-600 dark:text-gray-400">Page {page + 1}</p>
          <button
            onClick={() => setPage(page + 1)}
            disabled={inspections.length < limit}
            className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md disabled:opacity-50"
          >
            Next
          </button>
        </div>
      )}
    </div>
  );
}
