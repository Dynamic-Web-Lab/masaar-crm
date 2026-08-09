-- Rename BOS24 references to DLDAPI
-- This migration renames tables, columns, and indexes from BOS24 to DLD

-- 1. Rename table bos24_integration_settings → dld_integration_settings
ALTER TABLE IF EXISTS bos24_integration_settings RENAME TO dld_integration_settings;

-- 2. Rename indexes
ALTER INDEX IF EXISTS idx_bos24_settings_company RENAME TO idx_dld_settings_company;
ALTER INDEX IF EXISTS idx_bos24_settings_secret RENAME TO idx_dld_settings_secret;

-- 3. Rename column bos24_listing_uuid → dld_listing_uuid in listings table
ALTER TABLE IF EXISTS listings RENAME COLUMN bos24_listing_uuid TO dld_listing_uuid;
ALTER INDEX IF EXISTS idx_listings_bos24_uuid RENAME TO idx_listings_dld_uuid;

-- 4. Rename column bos24_inquiry_id → dld_inquiry_id in leads table
ALTER TABLE IF EXISTS leads RENAME COLUMN bos24_inquiry_id TO dld_inquiry_id;
ALTER INDEX IF EXISTS idx_leads_bos24_inquiry RENAME TO idx_leads_dld_inquiry;

-- 5. Update api_settings key from bos24_api_token → dld_api_token
UPDATE api_settings SET key = 'dld_api_token' WHERE key = 'bos24_api_token';

-- 6. Update lead source constraint to include 'dld' instead of 'bos24'
-- Drop old constraint
ALTER TABLE leads DROP CONSTRAINT IF EXISTS leads_source_check;
-- Add new constraint with 'dld' instead of 'bos24'
ALTER TABLE leads ADD CONSTRAINT leads_source_check
    CHECK (source IN ('whatsapp','web','referral','event','dld'));

-- 7. Update billing resource type
UPDATE billing_records SET resource = 'dld' WHERE resource = 'bos24';
