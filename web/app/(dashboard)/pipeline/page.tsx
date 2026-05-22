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
import { AgentAssist } from '@/components/agent/AgentAssist'
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
  const [activeTab, setActiveTab] = useState<'notes' | 'history' | 'assist'>('history')

  // Agent assist (Phase 8)
  const [threadId, setThreadId] = useState<string>('')
  const [threadConversation, setThreadConversation] = useState('')

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

    // Fetch thread data for agent assist (Phase 8)
    try {
      const threads = await api.threads.list({ limit: 1 })
      const threadsList = (threads as any).data || []
      const matchingThread = threadsList.find((t: any) => t.contact_id === lead.contact_id)
      if (matchingThread) {
        setThreadId(matchingThread.id)
        const messages = await api.threads.messages(matchingThread.id)
        const messagesList = Array.isArray(messages) ? messages : (messages as any).data || []
        const conversation = messagesList.map((m: any) => m.body).join('\n')
        setThreadConversation(conversation)
      }
    } catch (err) {
      console.error('Failed to fetch thread data:', err)
      setThreadId('')
      setThreadConversation('')
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
      <div className="flex-1 flex items-center justify-center text-surface-400 text-sm">
        {t('جاري التحميل...', 'Loading...')}
      </div>
    )
  }

  return (
    <div className="flex flex-col flex-1 overflow-hidden">
      <Header title={t('خط الأنابيب', 'Sales Pipeline')} />

      <div className="flex-1 overflow-x-auto p-6 bg-surface-50">
        <div className="flex justify-end mb-5">
          <button
            onClick={handleOpenAddModal}
            className="inline-flex items-center gap-1.5 px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-xl shadow-card hover:bg-primary-700 hover:shadow-card-hover transition-all duration-200 ease-soft"
          >
            <svg className="w-4 h-4" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
              <path fillRule="evenodd" d="M10 5a1 1 0 011 1v3h3a1 1 0 110 2h-3v3a1 1 0 11-2 0v-3H6a1 1 0 110-2h3V6a1 1 0 011-1z" clipRule="evenodd" />
            </svg>
            {t('Lead جديد', 'New Lead')}
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
              className="w-full px-3.5 py-2.5 border border-surface-200 rounded-xl text-sm bg-white placeholder:text-surface-400 focus:outline-none focus:ring-2 focus:ring-primary-500/40 focus:border-primary-400 transition-shadow"
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
              className="w-full px-3.5 py-2.5 border border-surface-200 rounded-xl text-sm bg-white placeholder:text-surface-400 focus:outline-none focus:ring-2 focus:ring-primary-500/40 focus:border-primary-400 transition-shadow"
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
              className="w-full px-3.5 py-2.5 border border-surface-200 rounded-xl text-sm bg-white placeholder:text-surface-400 focus:outline-none focus:ring-2 focus:ring-primary-500/40 focus:border-primary-400 transition-shadow"
            />
          </FormField>

          {submitError && <FormError error={submitError} />}

          <div className="flex gap-2 pt-2">
            <button
              type="button"
              onClick={() => setShowAddModal(false)}
              className="flex-1 px-4 py-2.5 border border-surface-200 text-surface-700 text-sm font-medium rounded-xl bg-white hover:bg-surface-50 transition-colors"
            >
              {t('إلغاء', 'Cancel')}
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="flex-1 px-4 py-2.5 bg-primary-600 text-white text-sm font-medium rounded-xl shadow-card hover:bg-primary-700 hover:shadow-card-hover transition-all disabled:opacity-60"
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
            <div className="grid grid-cols-2 gap-2.5 text-sm">
              <div className="bg-surface-50 border border-surface-200/70 rounded-xl p-3">
                <p className="text-[11px] uppercase tracking-wide text-surface-500 mb-1 font-medium">{t('المرحلة', 'Stage')}</p>
                <p className="font-semibold text-surface-900 capitalize">{selectedLead.stage}</p>
              </div>
              <div className="bg-surface-50 border border-surface-200/70 rounded-xl p-3">
                <p className="text-[11px] uppercase tracking-wide text-surface-500 mb-1 font-medium">{t('القيمة', 'Value')}</p>
                <p className="font-semibold text-surface-900 tabular-nums">
                  {selectedLead.currency} {selectedLead.deal_value.toLocaleString()}
                </p>
              </div>
              {selectedLead.contact?.phone_wa && (
                <div className="bg-surface-50 border border-surface-200/70 rounded-xl p-3 col-span-2">
                  <p className="text-[11px] uppercase tracking-wide text-surface-500 mb-1 font-medium">{t('واتساب', 'WhatsApp')}</p>
                  <p className="font-semibold text-surface-900 font-mono">{selectedLead.contact.phone_wa}</p>
                </div>
              )}
            </div>

            {/* Tabs */}
            <div className="border-b border-surface-200 flex gap-5">
              <button
                onClick={() => setActiveTab('history')}
                className={`py-2 px-0.5 text-sm font-medium border-b-2 -mb-px transition-colors ${
                  activeTab === 'history'
                    ? 'text-primary-600 border-primary-600'
                    : 'text-surface-500 border-transparent hover:text-surface-900'
                }`}
              >
                {t('السجل', 'History')}
              </button>
              <button
                onClick={() => setActiveTab('assist')}
                className={`py-2 px-0.5 text-sm font-medium border-b-2 -mb-px transition-colors ${
                  activeTab === 'assist'
                    ? 'text-primary-600 border-primary-600'
                    : 'text-surface-500 border-transparent hover:text-surface-900'
                }`}
              >
                💡 {t('المساعد', 'Assist')}
              </button>
              <button
                onClick={() => setActiveTab('notes')}
                className={`py-2 px-0.5 text-sm font-medium border-b-2 -mb-px transition-colors ${
                  activeTab === 'notes'
                    ? 'text-primary-600 border-primary-600'
                    : 'text-surface-500 border-transparent hover:text-surface-900'
                }`}
              >
                {t('الملاحظات', 'Notes')}
              </button>
            </div>

            {/* Tab content */}
            {activeTab === 'history' && (
              <div className="max-h-96 overflow-y-auto">
                <CommunicationHistoryComponent
                  communications={communications}
                  loading={commLoading}
                />
              </div>
            )}

            {activeTab === 'assist' && threadId && (
              <div className="max-h-96 overflow-y-auto">
                <AgentAssist
                  leadId={selectedLead.id}
                  threadId={threadId}
                  contactName={selectedLead.contact?.full_name || 'Unknown'}
                  conversation={threadConversation}
                  onActionClick={(action, message) => {
                    console.log('Action clicked:', action, message)
                    // Handle action (send message, etc.)
                  }}
                />
              </div>
            )}

            {activeTab === 'assist' && !threadId && (
              <div className="py-8 text-center text-surface-400 text-sm">
                {t('لا توجد محادثة لهذا العميل', 'No conversation for this lead')}
              </div>
            )}

            {activeTab === 'notes' && (
              <div>
                <label className="block text-[13px] font-medium text-surface-700 mb-1.5">
                  {t('الملاحظات', 'Notes')}
                </label>
                <textarea
                  value={notes}
                  onChange={(e) => { setNotes(e.target.value); setNotesSaved(false) }}
                  rows={5}
                  placeholder={t('أضف ملاحظات حول هذا العميل المحتمل...', 'Add notes about this lead...')}
                  className="w-full text-sm border border-surface-200 rounded-xl px-3.5 py-2.5 bg-white placeholder:text-surface-400 focus:outline-none focus:ring-2 focus:ring-primary-500/40 focus:border-primary-400 resize-none transition-shadow"
                />
              </div>
            )}

            {notesError && (
              <p className="text-xs text-red-600 bg-red-50 border border-red-100 px-3 py-2 rounded-lg">{notesError}</p>
            )}
            {notesSaved && (
              <p className="text-xs text-emerald-600 font-medium">{t('تم الحفظ', 'Saved')}</p>
            )}

            <div className="flex gap-2 pt-1">
              <button
                type="button"
                onClick={() => setSelectedLead(null)}
                className="flex-1 px-4 py-2.5 border border-surface-200 text-surface-700 text-sm font-medium rounded-xl bg-white hover:bg-surface-50 transition-colors"
              >
                {t('إغلاق', 'Close')}
              </button>
              {activeTab === 'notes' && (
                <button
                  type="button"
                  onClick={handleSaveNotes}
                  disabled={notesSaving}
                  className="flex-1 px-4 py-2.5 bg-primary-600 text-white text-sm font-medium rounded-xl shadow-card hover:bg-primary-700 hover:shadow-card-hover disabled:opacity-60 transition-all"
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
