-- +goose Up
-- +goose StatementBegin

-- Users: fast lookup by email (login, password reset, magic link)
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Contacts: fast lookup by WhatsApp number (webhook message routing)
CREATE INDEX IF NOT EXISTS idx_contacts_phone_wa ON contacts(phone_wa);

-- WhatsApp threads: filter by status (open/pending/closed inbox views)
CREATE INDEX IF NOT EXISTS idx_wa_threads_status ON whatsapp_threads(status);

-- WhatsApp threads: contact drilldown
CREATE INDEX IF NOT EXISTS idx_wa_threads_contact_id ON whatsapp_threads(contact_id);

-- WhatsApp messages: conversation history (most common query pattern)
CREATE INDEX IF NOT EXISTS idx_wa_messages_thread_id ON whatsapp_messages(thread_id);

-- Leads: pipeline board (stage column is the primary kanban filter)
CREATE INDEX IF NOT EXISTS idx_leads_stage ON leads(stage);

-- Leads: contact ownership lookups
CREATE INDEX IF NOT EXISTS idx_leads_contact_id ON leads(contact_id);

-- Leads: assigned agent filter
CREATE INDEX IF NOT EXISTS idx_leads_assigned_to ON leads(assigned_to);

-- Deals: stage and contact lookups
CREATE INDEX IF NOT EXISTS idx_deals_stage ON deals(stage);
CREATE INDEX IF NOT EXISTS idx_deals_contact_id ON deals(contact_id);

-- Notifications: per-user unread count (polled frequently)
CREATE INDEX IF NOT EXISTS idx_notifications_user_id_read ON notifications(user_id, is_read);

-- Payments: lease drilldown (payment history per lease)
CREATE INDEX IF NOT EXISTS idx_payments_lease_id ON payments(lease_id);

-- Payments: status filter (overdue detection)
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);

-- Leases: property and tenant lookups
CREATE INDEX IF NOT EXISTS idx_leases_property_id ON leases(property_id);
CREATE INDEX IF NOT EXISTS idx_leases_tenant_id ON leases(tenant_id);

-- Documents: entity scoping (most common document query)
CREATE INDEX IF NOT EXISTS idx_documents_entity ON documents(related_entity_type, related_entity_id);

-- Document signatures: lookup by status for dashboard counts
CREATE INDEX IF NOT EXISTS idx_document_signatures_status ON document_signatures(signature_status);

-- Audit logs: user activity trail
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_contacts_phone_wa;
DROP INDEX IF EXISTS idx_wa_threads_status;
DROP INDEX IF EXISTS idx_wa_threads_contact_id;
DROP INDEX IF EXISTS idx_wa_messages_thread_id;
DROP INDEX IF EXISTS idx_leads_stage;
DROP INDEX IF EXISTS idx_leads_contact_id;
DROP INDEX IF EXISTS idx_leads_assigned_to;
DROP INDEX IF EXISTS idx_deals_stage;
DROP INDEX IF EXISTS idx_deals_contact_id;
DROP INDEX IF EXISTS idx_notifications_user_id_read;
DROP INDEX IF EXISTS idx_payments_lease_id;
DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_leases_property_id;
DROP INDEX IF EXISTS idx_leases_tenant_id;
DROP INDEX IF EXISTS idx_documents_entity;
DROP INDEX IF EXISTS idx_document_signatures_status;
DROP INDEX IF EXISTS idx_audit_logs_user_id;
-- +goose StatementEnd
