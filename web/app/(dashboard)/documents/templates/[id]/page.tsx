'use client'

import { useEffect, useState } from 'react'
import { useRouter, useParams } from 'next/navigation'
import api from '@/lib/api'

interface DocumentTemplate {
  id: string
  template_name: string
  document_type: string
  template_content: string
  language?: string
  signature_required?: boolean
  signature_fields?: string[]
}

export default function EditTemplatePage() {
  const router = useRouter()
  const params = useParams()
  const templateId = params.id as string

  const [template, setTemplate] = useState<DocumentTemplate | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [formData, setFormData] = useState({
    template_name: '',
    template_content: '',
  })

  useEffect(() => {
    loadTemplate()
  }, [templateId])

  async function loadTemplate() {
    setLoading(true)
    try {
      const data = await api.documents.templates.get(templateId)
      if (data) {
        setTemplate(data as DocumentTemplate)
        setFormData({
          template_name: (data as DocumentTemplate).template_name,
          template_content: (data as DocumentTemplate).template_content,
        })
      }
    } catch (err) {
      console.error('Failed to load template:', err)
      alert('Failed to load template')
    } finally {
      setLoading(false)
    }
  }

  async function handleSave() {
    if (!formData.template_name || !formData.template_content) {
      alert('Template name and content are required')
      return
    }

    setSaving(true)
    try {
      await api.documents.templates.update(templateId, {
        template_name: formData.template_name,
        template_content: formData.template_content,
      })
      alert('Template updated successfully')
      router.push('/documents')
    } catch (err) {
      console.error('Failed to update template:', err)
      alert('Failed to update template')
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 p-6">
        <div className="flex justify-center items-center h-96">
          <div className="text-center">
            <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
            <p className="text-gray-600 mt-4">Loading template...</p>
          </div>
        </div>
      </div>
    )
  }

  if (!template) {
    return (
      <div className="min-h-screen bg-gray-50 p-6">
        <div className="max-w-4xl mx-auto">
          <button onClick={() => router.back()} className="text-blue-600 hover:text-blue-700 mb-6">
            ← Back
          </button>
          <div className="bg-white rounded-lg shadow-md p-12 text-center">
            <p className="text-gray-600">Template not found</p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <button onClick={() => router.back()} className="text-blue-600 hover:text-blue-700 mb-4 font-medium">
            ← Back to Templates
          </button>
          <h1 className="text-3xl font-bold text-gray-900">Edit Template</h1>
          <p className="text-gray-600 mt-1">Type: {template.document_type}</p>
        </div>

        {/* Form */}
        <div className="bg-white rounded-lg shadow-md p-6">
          <div className="space-y-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">Template Name</label>
              <input
                type="text"
                value={formData.template_name}
                onChange={(e) => setFormData({ ...formData, template_name: e.target.value })}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">Template Content</label>
              <p className="text-xs text-gray-500 mb-2">{'Use {{field_name}} for placeholders (e.g., {{tenant_name}})'}</p>
              <textarea
                value={formData.template_content}
                onChange={(e) => setFormData({ ...formData, template_content: e.target.value })}
                rows={12}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono text-sm"
              />
            </div>

            <div className="flex gap-3">
              <button
                onClick={handleSave}
                disabled={saving}
                className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-400 font-medium"
              >
                {saving ? 'Saving...' : 'Save Changes'}
              </button>
              <button
                onClick={() => router.back()}
                className="px-6 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 font-medium"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>

        {/* Preview */}
        <div className="mt-8 bg-white rounded-lg shadow-md p-6">
          <h2 className="text-xl font-semibold mb-4">Preview</h2>
          <div className="bg-gray-50 p-4 rounded border border-gray-200 font-serif text-gray-800 whitespace-pre-wrap break-words">
            {formData.template_content || 'No content yet'}
          </div>
        </div>
      </div>
    </div>
  )
}
