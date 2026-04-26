-- +goose Up
CREATE TABLE payment_reminders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    reminder_type VARCHAR(50) NOT NULL CHECK (reminder_type IN ('30_days', '15_days', '7_days', '1_day_late', '5_days_late', '10_days_late')),
    reminder_date DATE NOT NULL,
    sent_at TIMESTAMP WITH TIME ZONE,
    delivery_method VARCHAR(50) NOT NULL CHECK (delivery_method IN ('whatsapp', 'email', 'sms')),
    delivery_status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (delivery_status IN ('pending', 'sent', 'failed')),
    delivery_error TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(payment_id, reminder_type, delivery_method)
);

CREATE INDEX idx_payment_reminders_company_id ON payment_reminders(company_id);
CREATE INDEX idx_payment_reminders_payment_id ON payment_reminders(payment_id);
CREATE INDEX idx_payment_reminders_reminder_date ON payment_reminders(reminder_date);
CREATE INDEX idx_payment_reminders_delivery_status ON payment_reminders(delivery_status, reminder_date);

-- +goose Down
DROP TABLE IF EXISTS payment_reminders;
