-- +goose Up
-- Seed a read-only demo user for public product demos.
-- Password is irrelevant — the /auth/demo endpoint bypasses it when DEMO_MODE=true.
-- The Viewer role enforces read-only access via RBAC even if the password is guessed.
INSERT INTO users (id, name, email, password_hash, role, lang_pref, company_id)
VALUES (
    '00000000-0000-0000-0000-000000000099',
    'Demo User',
    'demo@masaar.local',
    '$2a$10$x962uUSG.SNiL3LSNsOmMuZnRXndRGXQtHpOWVm3eMnFmiLU0JJWu',
    'viewer',
    'en',
    '00000000-0000-0000-0000-000000000001'
) ON CONFLICT (email) DO NOTHING;

-- +goose Down
DELETE FROM users WHERE email = 'demo@masaar.local';
