'use client'
import { useEffect, useState, useCallback } from 'react'
import {
  DndContext, DragEndEvent, DragOverlay, DragStartEvent,
  PointerSensor, useSensor, useSensors, closestCorners,
} from '@dnd-kit/core'
import { Header } from '@/components/layout/Header'
import { KanbanColumn } from '@/components/kanban/Column'
import { KanbanCard } from '@/components/kanban/Card'
import { Modal, FormField, FormError } from '@/components/ui/Modal'
import { CommunicationHistoryComponent } from '@/components/communication/CommunicationHistory'
import { api } from '@/lib/api'
import { useLang } from '@/context/LangContext'
import type { KanbanBoard, Lead, LeadStage, Contact, PaginatedResult, CommunicationHistory } from '@/types'

const STAGES: LeadStage[] = ['new', 'contacted', 'qualified', 'proposal', 'won', 'lost']

export default function PipelinePage() {
  const [board, setBoard] = useState<KanbanBoard>({})
  const [activeCard, setActiveCard] = useState<Lead | null>(null)
  const [loading, setLoading] = useState(true)
  const [showAddModal, setShowAddModal] = useState(false)
  const [contacts, setContacts] = useState<Contact[]>([])
  const [submitting, setSubmitting] = useState(false)
  const [submitError, setSubmitError] = useState('')
  const { t } = useLang()

  // Lead detail / notes modal
  const [selectedLead, setSelectedLead] = useState<Lead | null>(null)
  const [notes, setNotes] = useState('')
  const [notesSaving, setNotesSaving] = useState(false)
  const [notesSaved, setNotesSaved] = useState(false)
  const [notesError, setNotesError] = useState('')

  // Communication history
  const [communications, setCommunications] = useState<CommunicationHistory[]>([])
  const [commLoading, setCommLoading] = useState(false)
  const [activeTab, setActiveTab] = useState<'notes' | 'history'>('history')

  const handleOpenLead = async (lead: Lead) => {
    setSelectedLead(lead)
    setNotes(lead.notes ?? '')
    setNotesSaved(false)
    setNotesError('')
    setActiveTab('history')

    // Fetch communication history
    setCommLoading(true)
    try {
      const response = (await api.leads.communications(lead.id, 100)) as CommunicationHistory[]
      setCommunications(Array.isArray(response) ? response : [])
    } catch (err) {
      console.error('Failed to fetch communications:', err)
      setCommunications([])
    } finally {
      setCommLoading(false)
    }
  }

  const handleSaveNotes = async () => {
    if (!selectedLead) return
    setNotesSaving(true)
    setNotesSaved(false)
    setNotesError('')
    try {
      await api.leads.updateNotes(selectedLead.id, notes)
      setBoard((prev) => {
        const next = { ...prev }
        for (const stage of Object.keys(next) as LeadStage[]) {
          next[stage] = (next[stage] ?? []).map((l) =>
            l.id === selectedLead.id ? { ...l, notes } : l
          )
        }
        return next
      })
      setSelectedLead((prev) => prev ? { ...prev, notes } : null)
      setNotesSaved(true)
    } catch (err: unknown) {
      setNotesError(err instanceof Error ? err.message : t('حدث خطأ', 'Something went wrong'))
    } finally {
      setNotesSaving(false)
    }
  }

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 8 } })
  )

  const load = useCallback(async () => {
    try {
      const data = await api.leads.kanban() as KanbanBoard
      setBoard(data ?? {})
    } catch {
      // handle error silently
    } finally {
      setLoading(false)
    }
  }, [])

  const loadContacts = useCallback(async () => {
    try {
      const data = await api.contacts.list({ limit: 100 }) as PaginatedResult<Contact>
      setContacts(data.data ?? [])
    } catch {
      setContacts([])
    }
  }, [])

  const handleOpenAddModal = () => {
    loadContacts()
    setShowAddModal(true)
  }

  const handleAddLead = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setSubmitError('')
    setSubmitting(true)
    
    const formData = new FormData(e.currentTarget)
    const data = {
      contact_id: formData.get('contact_id'),
      source: formData.get('source') || 'whatsapp',
      deal_value: parseFloat(formData.get('deal_value') as string) || 0,
    }

    if (!data.contact_id) {
      setSubmitError(t('يرجى اختيار جهة اتصال', 'Please select a contact'))
      setSubmitting(false)
      return
    }

    try {
      await api.leads.create(data)
      setShowAddModal(false)
      load()
    } catch (err: any) {
      setSubmitError(err.message || t('حدث خطأ', 'Something went wrong'))
    } finally {
      setSubmitting(false)
    }
  }

  useEffect(() => { load() }, [load])

  const findCard = (id: string): Lead | null => {
    for (const leads of Object.values(board)) {
      const found = leads?.find((l) => l.id === id)
      if (found) return found
    }
    return null
  }

  const findStage = (id: string): LeadStage | null => {
    for (const [stage, leads] of Object.entries(board)) {
      if (leads?.some((l) => l.id === id)) return stage as LeadStage
    }
    return null
  }

  const handleDragStart = (e: DragStartEvent) => {
    setActiveCard(findCard(String(e.active.id)))
  }

  const handleDragEnd = async (e: DragEndEvent) => {
    setActiveCard(null)
    const { active, over } = e
    if (!over) return

    const leadId = String(active.id)
    const targetStage = (STAGES.includes(over.id as LeadStage)
      ? over.id
      : findStage(String(over.id))) as LeadStage | null

    const currentStage = findStage(leadId)
    if (!targetStage || targetStage === currentStage) return

    // Optimistic update
    setBoard((prev) => {
      const next = { ...prev }
      const card = (next[currentStage!] ?? []).find((l) => l.id === leadId)
      if (!card) return prev
      next[currentStage!] = (next[currentStage!] ?? []).filter((l) => l.id !== leadId)
      next[targetStage] = [{ ...card, stage: targetStage }, ...(next[targetStage] ?? [])]
      return next
    })

    await api.leads.updateStage(leadId, targetStage).catch(() => load())
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
      <Header title={t('خط الأنابيب', 'Sales Pipeline')} />

      <div className="flex-1 overflow-x-auto p-6">
        <div className="flex justify-end mb-4">
          <button
            onClick={handleOpenAddModal}
            className="px-4 py-2 bg-green-600 text-white text-sm font-medium rounded-lg hover:bg-green-700 transition-colors"
          >
            + {t('-lead جديد', 'New Lead')}
          </button>
        </div>

        <DndContext
          sensors={sensors}
          collisionDetection={closestCorners}
          onDragStart={handleDragStart}
          onDragEnd={handleDragEnd}
        >
          <div className="flex gap-4 min-w-max pb-4">
            {STAGES.map((stage) => (
              <KanbanColumn
                key={stage}
                stage={stage}
                leads={board[stage] ?? []}
                onOpenLead={handleOpenLead}
              />
            ))}
          </div>

          <DragOverlay>
            {activeCard && <KanbanCard lead={activeCard} />}
          </DragOverlay>
        </DndContext>
      </div>

      {/* Add Lead Modal */}
      <Modal
        open={showAddModal}
        onClose={() => setShowAddModal(false)}
        title={t('إضافة Lead جديد', 'Add New Lead')}
      >
        <form onSubmit={handleAddLead} className="space-y-4">
          <FormField label={t('جهة الاتصال *', 'Contact *')}>
            <select
              name="contact_id"
              required
              className="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-brand-500 focus:border-transparent"
            >
              <option value="">{t('اختر جهة اتصال', 'Select a contact')}</option>
              {contacts.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.full_name} ({c.phone_wa})
                </option>
              ))}
            </select>
          </FormField>
          
          <FormField label={t('المصدر', 'Source')}>
            <select
              name="source"
              defaultValue="whatsapp"
              className="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-brand-500 focus:border-transparent"
            >
              <option value="whatsapp">{t('واتساب', 'WhatsApp')}</option>
              <option value="web">{t('الموقع', 'Website')}</option>
              <option value="referral">{t('إحالة', 'Referral')}</option>
              <option value="event">{t('فعالية', 'Event')}</option>
            </select>
          </FormField>
          
          <FormField label={t('قيمة الصفقة (AED)', 'Deal Value (AED)')}>
            <input
              type="number"
              name="deal_value"
              min="0"
              step="1"
              placeholder="0"
              className="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-brand-500 focus:border-transparent"
            />
          </FormField>

          {submitError && <FormError error={submitError} />}

          <div className="flex gap-2 pt-2">
            <button
              type="button"
              onClick={() => setShowAddModal(false)}
              className="flex-1 px-4 py-2 border border-gray-200 text-gray-600 text-sm font-medium rounded-lg hover:bg-gray-50 transition-colors"
            >
              {t('إلغاء', 'Cancel')}
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="flex-1 px-4 py-2 bg-brand-600 text-white text-sm font-medium rounded-lg hover:bg-brand-700 transition-colors disabled:opacity-60"
            >
              {submitting ? t('جاري الإضافة...', 'Adding...') : t('إضافة', 'Add')}
            </button>
          </div>
        </form>
      </Modal>

      {/* Lead Detail / Notes Modal */}
      <Modal
        open={!!selectedLead}
        onClose={() => setSelectedLead(null)}
        title={selectedLead?.contact?.full_name ?? t('تفاصيل Lead', 'Lead Details')}
      >
        {selectedLead && (
          <div className="space-y-4">
            {/* Lead meta */}
            <div className="grid grid-cols-2 gap-3 text-sm">
              <div className="bg-gray-50 rounded-lg p-3">
                <p className="text-xs text-gray-400 mb-0.5">{t('المرحلة', 'Stage')}</p>
                <p className="font-medium text-gray-800 capitalize">{selectedLead.stage}</p>
              </div>
              <div className="bg-gray-50 rounded-lg p-3">
                <p className="text-xs text-gray-400 mb-0.5">{t('القيمة', 'Value')}</p>
                <p className="font-medium text-gray-800">
                  {selectedLead.currency} {selectedLead.deal_value.toLocaleString()}
                </p>
              </div>
              {selectedLead.contact?.phone_wa && (
                <div className="bg-gray-50 rounded-lg p-3 col-span-2">
                  <p className="text-xs text-gray-400 mb-0.5">{t('واتساب', 'WhatsApp')}</p>
                  <p className="font-medium text-gray-800">{selectedLead.contact.phone_wa}</p>
                </div>
              )}
            </div>

            {/* Tabs */}
            <div className="border-b border-gray-200 flex gap-4">
              <button
                onClick={() => setActiveTab('history')}
                className={`py-2 px-1 text-sm font-medium border-b-2 transition-colors ${
                  activeTab === 'history'
                    ? 'text-brand-600 border-brand-600'
                    : 'text-gray-600 border-transparent hover:text-gray-800'
                }`}
              >
                {t('السجل', 'History')}
              </button>
              <button
                onClick={() => setActiveTab('notes')}
                className={`py-2 px-1 text-sm font-medium border-b-2 transition-colors ${
                  activeTab === 'notes'
                    ? 'text-brand-600 border-brand-600'
                    : 'text-gray-600 border-transparent hover:text-gray-800'
                }`}
              >
                {t('الملاحظات', 'Notes')}
              </button>
            </div>

            {/* Tab content */}
            {activeTab === 'history' ? (
              <div className="max-h-96 overflow-y-auto">
                <CommunicationHistoryComponent
                  communications={communications}
                  loading={commLoading}
                />
              </div>
            ) : (
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  {t('الملاحظات', 'Notes')}
                </label>
                <textarea
                  value={notes}
                  onChange={(e) => { setNotes(e.target.value); setNotesSaved(false) }}
                  rows={5}
                  placeholder={t('أضف ملاحظات حول هذا العميل المحتمل...', 'Add notes about this lead...')}
                  className="w-full text-sm border border-gray-200 rounded-lg px-3 py-2.5 focus:outline-none focus:ring-2 focus:ring-brand-500 focus:border-transparent resize-none"
                />
              </div>
            )}

            {notesError && (
              <p className="text-xs text-red-500 bg-red-50 px-3 py-2 rounded-lg">{notesError}</p>
            )}
            {notesSaved && (
              <p className="text-xs text-green-600">{t('تم الحفظ', 'Saved')}</p>
            )}

            <div className="flex gap-2 pt-1">
              <button
                type="button"
                onClick={() => setSelectedLead(null)}
                className="flex-1 px-4 py-2 border border-gray-200 text-gray-600 text-sm font-medium rounded-lg hover:bg-gray-50 transition-colors"
              >
                {t('إغلاق', 'Close')}
              </button>
              {activeTab === 'notes' && (
                <button
                  type="button"
                  onClick={handleSaveNotes}
                  disabled={notesSaving}
                  className="flex-1 px-4 py-2 bg-brand-600 text-white text-sm font-medium rounded-lg hover:bg-brand-700 disabled:opacity-60 transition-colors"
                >
                  {notesSaving ? t('جاري الحفظ...', 'Saving...') : t('حفظ الملاحظات', 'Save Notes')}
                </button>
              )}
            </div>
          </div>
        )}
      </Modal>
    </div>
  )
}
