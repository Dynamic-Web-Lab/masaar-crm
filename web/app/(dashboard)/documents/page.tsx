'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import api from '@/lib/api'

interface DocumentTemplate {
  id: string
  company_id: string
  template_name: string
  document_type: string
  template_content: string
  language?: string
  signature_required?: boolean
  signature_fields?: string[]
  created_by: string
  created_at: string
}

interface PaginatedResult {
  data: DocumentTemplate[]
  total: number
  page: number
  limit: number
}

export default function DocumentsPage() {
  const router = useRouter()
  const [templates, setTemplates] = useState<DocumentTemplate[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const limit = 20

  const [showCreateForm, setShowCreateForm] = useState(false)
  const [creating, setCreating] = useState(false)
  const [formData, setFormData] = useState({
    template_name: '',
    document_type: '',
    template_content: '',
    language: 'en',
    signature_required: false,
  })

  useEffect(() => {
    loadTemplates()
  }, [page])

  async function loadTemplates() {
    setLoading(true)
    try {
      const result = await api.documents.templates.list(page, limit)
      if (result && typeof result === 'object') {
        const paginatedResult = result as PaginatedResult
        setTemplates(paginatedResult.data || [])
        setTotal(paginatedResult.total || 0)
      }
    } catch (err) {
      console.error('Failed to load templates:', err)
    } finally {
      setLoading(false)
    }
  }

  async function handleCreateTemplate() {
    if (!formData.template_name || !formData.document_type) {
      alert('Template name and document type are required')
      return
    }

    setCreating(true)
    try {
      await api.documents.templates.create(formData)
      setFormData({
        template_name: '',
        document_type: '',
        template_content: '',
        language: 'en',
        signature_required: false,
      })
      setShowCreateForm(false)
      setPage(1)
      loadTemplates()
    } catch (err) {
      console.error('Failed to create template:', err)
      alert('Failed to create template')
    } finally {
      setCreating(false)
    }
  }

  async function handleDeleteTemplate(id: string) {
    if (!confirm('Delete this template?')) return

    try {
      await api.documents.templates.delete(id)
      loadTemplates()
    } catch (err) {
      console.error('Failed to delete template:', err)
      alert('Failed to delete template')
    }
  }

  const totalPages = Math.ceil(total / limit)

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Document Templates</h1>
            <p className="text-gray-600 mt-1">Create and manage reusable document templates</p>
          </div>
          <button
            onClick={() => setShowCreateForm(!showCreateForm)}
            className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 font-medium"
          >
            {showCreateForm ? 'Cancel' : '+ New Template'}
          </button>
        </div>

        {/* Create Form */}
        {showCreateForm && (
          <div className="bg-white rounded-lg shadow-md p-6 mb-6">
            <h2 className="text-xl font-semibold mb-4">Create Template</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700">Template Name</label>
                <input
                  type="text"
                  value={formData.template_name}
                  onChange={(e) => setFormData({ ...formData, template_name: e.target.value })}
                  className="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="e.g., Standard Lease Agreement"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700">Document Type</label>
                <select
                  value={formData.document_type}
                  onChange={(e) => setFormData({ ...formData, document_type: e.target.value })}
                  className="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="">Select type</option>
                  <option value="lease">Lease Agreement</option>
                  <option value="contract">Contract</option>
                  <option value="invoice">Invoice</option>
                  <option value="agreement">Agreement</option>
                  <option value="other">Other</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700">Language</label>
                <select
                  value={formData.language}
                  onChange={(e) => setFormData({ ...formData, language: e.target.value })}
                  className="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="en">English</option>
                  <option value="ar">Arabic</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700">Template Content</label>
                <textarea
                  value={formData.template_content}
                  onChange={(e) => setFormData({ ...formData, template_content: e.target.value })}
                  rows={6}
                  className="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono text-sm"
                  placeholder="Enter template content with {{placeholders}} for dynamic fields"
                />
              </div>

              <div className="flex items-center">
                <input
                  type="checkbox"
                  checked={formData.signature_required}
                  onChange={(e) => setFormData({ ...formData, signature_required: e.target.checked })}
                  className="w-4 h-4 text-blue-600 rounded"
                  id="signature_required"
                />
                <label htmlFor="signature_required" className="ml-2 text-sm text-gray-700">
                  Requires signatures
                </label>
              </div>

              <button
                onClick={handleCreateTemplate}
                disabled={creating}
                className="w-full bg-blue-600 text-white py-2 rounded-lg hover:bg-blue-700 disabled:bg-gray-400 font-medium"
              >
                {creating ? 'Creating...' : 'Create Template'}
              </button>
            </div>
          </div>
        )}

        {/* Templates List */}
        {loading ? (
          <div className="text-center py-12">
            <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
            <p className="text-gray-600 mt-4">Loading templates...</p>
          </div>
        ) : templates.length === 0 ? (
          <div className="bg-white rounded-lg shadow-md p-12 text-center">
            <p className="text-gray-600">No templates yet. Create your first template to get started.</p>
          </div>
        ) : (
          <div className="space-y-4">
            {templates.map((template) => (
              <div key={template.id} className="bg-white rounded-lg shadow-md p-4 hover:shadow-lg transition">
                <div className="flex justify-between items-start">
                  <div className="flex-1">
                    <h3 className="font-semibold text-lg text-gray-900">{template.template_name}</h3>
                    <p className="text-sm text-gray-600 mt-1">Type: {template.document_type}</p>
                    {template.signature_required && (
                      <span className="inline-block mt-2 px-2 py-1 text-xs font-medium bg-amber-100 text-amber-800 rounded">
                        Requires Signatures
                      </span>
                    )}
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={() => router.push(`/documents/templates/${template.id}`)}
                      className="px-3 py-1 text-sm bg-gray-100 hover:bg-gray-200 rounded text-gray-700 font-medium"
                    >
                      Edit
                    </button>
                    <button
                      onClick={() => handleDeleteTemplate(template.id)}
                      className="px-3 py-1 text-sm bg-red-100 hover:bg-red-200 rounded text-red-700 font-medium"
                    >
                      Delete
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex justify-center gap-2 mt-8">
            <button
              onClick={() => setPage(Math.max(1, page - 1))}
              disabled={page === 1}
              className="px-3 py-1 border border-gray-300 rounded disabled:opacity-50"
            >
              Previous
            </button>
            <span className="px-3 py-1">
              Page {page} of {totalPages}
            </span>
            <button
              onClick={() => setPage(Math.min(totalPages, page + 1))}
              disabled={page === totalPages}
              className="px-3 py-1 border border-gray-300 rounded disabled:opacity-50"
            >
              Next
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
