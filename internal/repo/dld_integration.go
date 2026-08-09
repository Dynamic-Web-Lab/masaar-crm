package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/dynamicweblab/masaar-crm/internal/domain"
)

// DLDIntegrationRepo handles storage for per-company DLD integration settings
// and provides idempotent upsert helpers for synced listings and leads.
type DLDIntegrationRepo struct {
	db *pgxpool.Pool
}

func NewDLDIntegrationRepo(db *pgxpool.Pool) *DLDIntegrationRepo {
	return &DLDIntegrationRepo{db: db}
}

// GetSettings returns the DLD integration settings for a company.
// Returns a zeroed struct (no error) if no row exists yet.
func (r *DLDIntegrationRepo) GetSettings(ctx context.Context, companyID uuid.UUID) (*domain.DLDSettings, error) {
	const q = `
		SELECT id, company_id, api_key, webhook_secret, webhook_id, last_sync_at, updated_at
		FROM dld_integration_settings
		WHERE company_id = $1
	`
	s := &domain.DLDSettings{}
	err := r.db.QueryRow(ctx, q, companyID).Scan(
		&s.ID, &s.CompanyID,
		&s.APIKey, &s.WebhookSecret, &s.WebhookID,
		&s.LastSyncAt, &s.UpdatedAt,
	)
	if err != nil {
		// No row is not an error — return an empty settings object
		return &domain.DLDSettings{CompanyID: companyID}, nil
	}
	return s, nil
}

// SaveSettings upserts DLD integration settings for a company.
func (r *DLDIntegrationRepo) SaveSettings(ctx context.Context, s *domain.DLDSettings) error {
	const q = `
		INSERT INTO dld_integration_settings
			(id, company_id, api_key, webhook_secret, webhook_id, last_sync_at, updated_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, NOW())
		ON CONFLICT (company_id) DO UPDATE SET
			api_key        = EXCLUDED.api_key,
			webhook_secret = EXCLUDED.webhook_secret,
			webhook_id     = EXCLUDED.webhook_id,
			last_sync_at   = EXCLUDED.last_sync_at,
			updated_at     = NOW()
	`
	_, err := r.db.Exec(ctx, q,
		s.CompanyID, s.APIKey, s.WebhookSecret, s.WebhookID, s.LastSyncAt,
	)
	return err
}

// UpdateLastSyncAt sets the last_sync_at timestamp for a company's DLD settings.
func (r *DLDIntegrationRepo) UpdateLastSyncAt(ctx context.Context, companyID uuid.UUID, t time.Time) error {
	const q = `
		UPDATE dld_integration_settings SET last_sync_at = $1, updated_at = NOW()
		WHERE company_id = $2
	`
	_, err := r.db.Exec(ctx, q, t, companyID)
	return err
}

// GetCompanyByWebhookSecret finds the company whose DLD webhook_secret matches
// the provided token. Used to route inbound webhook requests to the correct tenant.
func (r *DLDIntegrationRepo) GetCompanyByWebhookSecret(ctx context.Context, secret string) (*domain.DLDSettings, error) {
	const q = `
		SELECT id, company_id, api_key, webhook_secret, webhook_id, last_sync_at, updated_at
		FROM dld_integration_settings
		WHERE webhook_secret = $1 AND webhook_secret <> ''
	`
	s := &domain.DLDSettings{}
	err := r.db.QueryRow(ctx, q, secret).Scan(
		&s.ID, &s.CompanyID,
		&s.APIKey, &s.WebhookSecret, &s.WebhookID,
		&s.LastSyncAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// UpsertListing inserts or updates a listing imported from DLD.
// Idempotent: ON CONFLICT (company_id, dld_listing_uuid) updates the key fields.
// Returns the listing UUID (Masaar's internal UUID).
func (r *DLDIntegrationRepo) UpsertListing(ctx context.Context,
	companyID uuid.UUID,
	dldUUID, title, description, propertyType, listingType, city, status, coverImageURL, currency string,
	price float64,
) (uuid.UUID, error) {
	const q = `
		INSERT INTO listings (
			id, company_id, dld_listing_uuid,
			title, description, property_type, listing_type,
			city, status, cover_image_url,
			price, currency,
			portal_sync_status
		) VALUES (
			uuid_generate_v4(), $1, $2,
			$3, $4, $5, $6,
			$7, $8, $9,
			$10, $11,
			'{"source":"dld"}'::jsonb
		)
		ON CONFLICT (company_id, dld_listing_uuid) DO UPDATE SET
			title          = EXCLUDED.title,
			description    = EXCLUDED.description,
			property_type  = EXCLUDED.property_type,
			listing_type   = EXCLUDED.listing_type,
			city           = EXCLUDED.city,
			status         = EXCLUDED.status,
			cover_image_url = EXCLUDED.cover_image_url,
			price          = EXCLUDED.price,
			currency       = EXCLUDED.currency,
			updated_at     = NOW()
		RETURNING id
	`
	var id uuid.UUID
	err := r.db.QueryRow(ctx, q,
		companyID, dldUUID,
		title, description, propertyType, listingType,
		city, status, coverImageURL,
		price, currency,
	).Scan(&id)
	return id, err
}

// DeactivateListing marks a DLD listing as inactive/deleted.
func (r *DLDIntegrationRepo) DeactivateListing(ctx context.Context, companyID uuid.UUID, dldUUID string) error {
	const q = `
		UPDATE listings SET status = 'inactive', updated_at = NOW()
		WHERE company_id = $1 AND dld_listing_uuid = $2
	`
	_, err := r.db.Exec(ctx, q, companyID, dldUUID)
	return err
}

// CreateLeadFromInquiry creates a lead from a DLD inquiry if it doesn't already exist.
// Idempotent: does nothing on conflict (dld_inquiry_id already imported).
// Returns true if a new lead was created, false if it was a duplicate.
func (r *DLDIntegrationRepo) CreateLeadFromInquiry(ctx context.Context,
	contactID uuid.UUID,
	dldInquiryID int,
	notes string,
) (created bool, leadID uuid.UUID, err error) {
	const q = `
		INSERT INTO leads (id, contact_id, stage, source, deal_value, currency, notes, dld_inquiry_id)
		VALUES (uuid_generate_v4(), $1, 'new', 'dld', 0, 'AED', $2, $3)
		ON CONFLICT (dld_inquiry_id) DO NOTHING
		RETURNING id
	`
	err = r.db.QueryRow(ctx, q, contactID, notes, dldInquiryID).Scan(&leadID)
	if err != nil {
		// pgx returns ErrNoRows when DO NOTHING fires — that's a duplicate, not an error
		if err.Error() == "no rows in result set" {
			return false, uuid.Nil, nil
		}
		return false, uuid.Nil, err
	}
	return true, leadID, nil
}

// UpdateContactEmailIfEmpty sets email on a contact only when the contact's
// current email is NULL or empty — used when enriching from a DLD inquiry.
func (r *DLDIntegrationRepo) UpdateContactEmailIfEmpty(ctx context.Context, contactID uuid.UUID, email string) error {
	const q = `
		UPDATE contacts SET email = $1, updated_at = NOW()
		WHERE id = $2 AND (email IS NULL OR email = '')
	`
	_, err := r.db.Exec(ctx, q, email, contactID)
	return err
}

// ListCompaniesWithDLD returns all companies that have a non-empty DLD API key.
// Used by the nightly sync goroutine to iterate over active integrations.
func (r *DLDIntegrationRepo) ListCompaniesWithDLD(ctx context.Context) ([]*domain.DLDSettings, error) {
	const q = `
		SELECT id, company_id, api_key, webhook_secret, webhook_id, last_sync_at, updated_at
		FROM dld_integration_settings
		WHERE api_key <> ''
		ORDER BY company_id
	`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.DLDSettings
	for rows.Next() {
		s := &domain.DLDSettings{}
		if err := rows.Scan(
			&s.ID, &s.CompanyID,
			&s.APIKey, &s.WebhookSecret, &s.WebhookID,
			&s.LastSyncAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, s)
	}
	return results, nil
}
