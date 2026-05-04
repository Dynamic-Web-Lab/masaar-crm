-- +goose Up
CREATE TABLE bank_integrations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,

  -- Bank details
  bank_name VARCHAR(255) NOT NULL,
  bank_code VARCHAR(50),
  account_number VARCHAR(100),
  account_name VARCHAR(255),
  iban VARCHAR(50),

  -- Integration settings
  integration_type VARCHAR(50) NOT NULL, -- api, csv_upload, manual
  status VARCHAR(50) DEFAULT 'active', -- active, inactive, error

  -- API credentials (encrypted in production)
  api_key_encrypted VARCHAR(500),
  api_secret_encrypted VARCHAR(500),
  api_endpoint VARCHAR(255),

  -- Sync settings
  auto_sync BOOLEAN DEFAULT TRUE,
  last_sync_date TIMESTAMP,
  sync_interval_hours INTEGER DEFAULT 24,

  -- Error tracking
  last_sync_error TEXT,
  sync_error_count INTEGER DEFAULT 0,

  -- Connection status
  is_connected BOOLEAN DEFAULT FALSE,
  connection_test_date TIMESTAMP,

  -- Metadata
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  created_by UUID REFERENCES users(id),
  updated_by UUID REFERENCES users(id)
);

CREATE INDEX idx_bank_integrations_company_id ON bank_integrations(company_id);
CREATE INDEX idx_bank_integrations_bank_name ON bank_integrations(bank_name);
CREATE INDEX idx_bank_integrations_status ON bank_integrations(status);
CREATE INDEX idx_bank_integrations_auto_sync ON bank_integrations(auto_sync);

-- +goose Down
DROP TABLE bank_integrations;
