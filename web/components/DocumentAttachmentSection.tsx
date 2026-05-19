'use client'

import { useEffect, useState } from 'react'
import api from '@/lib/api'

interface Document {
  id: string
  document_title: string
  document_type: string
  file_url?: string
  signature_status: string
  created_at: string
}

interface Signature {
  id: string
  signer_name: string
  signer_email: string
  signature_status: string
  signed_at?: string
  created_at: string
}

interface DocumentWithSignatures {
  document: Document
  signatures: Signature[]
}

interface DocumentAttachmentSectionProps {
  entityType: string
  entityId: string
  canEdit?: boolean
}

export default function DocumentAttachmentSection({
  entityType,
  entityId,
  canEdit = false,
}: DocumentAttachmentSectionProps) {
  const [documents, setDocuments] = useState<Document[]>([])
  const [loading, setLoading] = useState(true)
  const [expandedDoc, setExpandedDoc] = useState<string | null>(null)
  const [signatures, setSignatures] = useState<Record<string, Signature[]>>({})

  const [showSignatureForm, setShowSignatureForm] = useState(false)
  const [signerData, setSignerData] = useState({
    docId: '',
    signerName: '',
    signerEmail: '',
  })
  const [sendingSignature, setSendingSignature] = useState(false)

  useEffect(() => {
    loadDocuments()
  }, [entityType, entityId])

  async function loadDocuments() {
    setLoading(true)
    try {
      const docs = await api.documents.list(entityType, entityId)
      if (Array.isArray(docs)) {
        setDocuments(docs as Document[])
      }
    } catch (err) {
      console.error('Failed to load documents:', err)
    } finally {
      setLoading(false)
    }
  }

  async function loadSignatures(docId: string) {
    try {
      const result = await api.documents.get(docId)
      if (result && typeof result === 'object') {
        const data = result as { document: Document; signatures: Signature[] }
        setSignatures((prev) => ({ ...prev, [docId]: data.signatures || [] }))
      }
    } catch (err) {
      console.error('Failed to load signatures:', err)
    }
  }

  async function handleRequestSignature() {
    if (!signerData.signerName || !signerData.signerEmail) {
      alert('Signer name and email required')
      return
    }

    setSendingSignature(true)
    try {
      await api.documents.requestSignature(
        signerData.docId,
        signerData.signerName,
        signerData.signerEmail
      )
      alert('Signature request sent successfully')
      setShowSignatureForm(false)
      setSignerData({ docId: '', signerName: '', signerEmail: '' })
      loadSignatures(signerData.docId)
    } catch (err) {
      console.error('Failed to request signature:', err)
      alert('Failed to request signature')
    } finally {
      setSendingSignature(false)
    }
  }

  async function handleDeleteDocument(docId: string) {
    if (!confirm('Delete this document?')) return

    try {
      await api.documents.delete(docId)
      loadDocuments()
    } catch (err) {
      console.error('Failed to delete document:', err)
      alert('Failed to delete document')
    }
  }

  if (loading) {
    return (
      <div className="py-4">
        <div className="inline-block animate-spin rounded-full h-4 w-4 border-b-2 border-gray-400"></div>
        <span className="text-sm text-gray-600 ml-2">Loading documents...</span>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h3 className="text-lg font-semibold text-gray-900">Documents</h3>
        {canEdit && (
          <button
            onClick={() => setShowSignatureForm(!showSignatureForm)}
            className="text-sm px-3 py-1 bg-blue-100 text-blue-700 rounded hover:bg-blue-200"
          >
            {showSignatureForm ? 'Cancel' : '+ Add Document'}
          </button>
        )}
      </div>

      {/* Create Document Form */}
      {showSignatureForm && canEdit && (
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <h4 className="font-medium text-gray-900 mb-3">Request Signature</h4>
          <div className="space-y-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Document</label>
              <select
                value={signerData.docId}
                onChange={(e) => setSignerData({ ...signerData, docId: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="">Select document</option>
                {documents.map((doc) => (
                  <option key={doc.id} value={doc.id}>
                    {doc.document_title}
                  </option>
                ))}
              </select>
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Signer Name</label>
                <input
                  type="text"
                  value={signerData.signerName}
                  onChange={(e) => setSignerData({ ...signerData, signerName: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="Full name"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Email</label>
                <input
                  type="email"
                  value={signerData.signerEmail}
                  onChange={(e) => setSignerData({ ...signerData, signerEmail: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="email@example.com"
                />
              </div>
            </div>

            <button
              onClick={handleRequestSignature}
              disabled={sendingSignature}
              className="w-full px-3 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 disabled:bg-gray-400 font-medium"
            >
              {sendingSignature ? 'Sending...' : 'Request Signature'}
            </button>
          </div>
        </div>
      )}

      {/* Documents List */}
      {documents.length === 0 ? (
        <div className="text-center py-6 bg-gray-50 rounded-lg">
          <p className="text-gray-600 text-sm">No documents attached yet</p>
        </div>
      ) : (
        <div className="space-y-2">
          {documents.map((doc) => (
            <div key={doc.id} className="border border-gray-200 rounded-lg overflow-hidden">
              {/* Document Header */}
              <div
                className="flex items-center justify-between p-4 bg-gray-50 hover:bg-gray-100 cursor-pointer transition"
                onClick={() => {
                  setExpandedDoc(expandedDoc === doc.id ? null : doc.id)
                  if (expandedDoc !== doc.id && !signatures[doc.id]) {
                    loadSignatures(doc.id)
                  }
                }}
              >
                <div className="flex-1">
                  <h4 className="font-medium text-gray-900">{doc.document_title}</h4>
                  <div className="flex gap-3 mt-1">
                    <span className="text-xs px-2 py-1 bg-gray-200 text-gray-700 rounded">
                      {doc.document_type}
                    </span>
                    <span
                      className={`text-xs px-2 py-1 rounded font-medium ${
                        doc.signature_status === 'signed'
                          ? 'bg-green-100 text-green-800'
                          : 'bg-amber-100 text-amber-800'
                      }`}
                    >
                      {doc.signature_status === 'signed' ? '✓ Signed' : '○ Pending'}
                    </span>
                  </div>
                </div>
                <svg
                  className={`w-5 h-5 text-gray-400 transition-transform ${
                    expandedDoc === doc.id ? 'rotate-180' : ''
                  }`}
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 14l-7 7m0 0l-7-7m7 7V3" />
                </svg>
              </div>

              {/* Expanded Content */}
              {expandedDoc === doc.id && (
                <div className="border-t border-gray-200 p-4 space-y-4">
                  {/* Signature Requests */}
                  <div>
                    <h5 className="font-medium text-gray-900 mb-2">Signatures</h5>
                    {signatures[doc.id] && signatures[doc.id].length > 0 ? (
                      <div className="space-y-2">
                        {signatures[doc.id].map((sig) => (
                          <div key={sig.id} className="flex items-center justify-between p-2 bg-gray-50 rounded">
                            <div className="text-sm">
                              <p className="font-medium text-gray-900">{sig.signer_name}</p>
                              <p className="text-gray-600">{sig.signer_email}</p>
                            </div>
                            <div
                              className={`text-xs px-2 py-1 rounded font-medium ${
                                sig.signature_status === 'signed'
                                  ? 'bg-green-100 text-green-800'
                                  : 'bg-yellow-100 text-yellow-800'
                              }`}
                            >
                              {sig.signature_status === 'signed' ? '✓ Signed' : '○ Pending'}
                            </div>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <p className="text-sm text-gray-600">No signature requests yet</p>
                    )}
                  </div>

                  {/* Actions */}
                  {canEdit && (
                    <div className="flex gap-2 pt-2 border-t border-gray-200">
                      {doc.file_url && (
                        <a
                          href={doc.file_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-sm px-3 py-1 bg-blue-100 text-blue-700 rounded hover:bg-blue-200"
                        >
                          View File
                        </a>
                      )}
                      <button
                        onClick={() => handleDeleteDocument(doc.id)}
                        className="text-sm px-3 py-1 bg-red-100 text-red-700 rounded hover:bg-red-200"
                      >
                        Delete
                      </button>
                    </div>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
