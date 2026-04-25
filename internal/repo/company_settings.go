package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maidulcu/masaar-crm/internal/domain"
)

type CompanySettingsRepo struct {
	pool *pgxpool.Pool
}

func NewCompanySettingsRepo(pool *pgxpool.Pool) *CompanySettingsRepo {
	return &CompanySettingsRepo{pool: pool}
}

func (r *CompanySettingsRepo) Get(ctx context.Context) (*domain.CompanySettings, error) {
	query := `SELECT id, name, vat_number, business_address, business_phone, business_email,
	bank_name, bank_account, bank_iban, updated_at, updated_by
	FROM company_settings LIMIT 1`

	var settings domain.CompanySettings

	err := r.pool.QueryRow(ctx, query).Scan(
		&settings.ID,
		&settings.Name,
		&settings.VATNumber,
		&settings.BusinessAddress,
		&settings.BusinessPhone,
		&settings.BusinessEmail,
		&settings.BankName,
		&settings.BankAccount,
		&settings.BankIBAN,
		&settings.UpdatedAt,
		&settings.UpdatedBy,
	)

	if err != nil {
		return nil, err
	}

	return &settings, nil
}

func (r *CompanySettingsRepo) Update(ctx context.Context, settings *domain.CompanySettings, userID *uuid.UUID) error {
	query := `UPDATE company_settings SET
	name = $1,
	vat_number = $2,
	business_address = $3,
	business_phone = $4,
	business_email = $5,
	bank_name = $6,
	bank_account = $7,
	bank_iban = $8,
	updated_at = $9,
	updated_by = $10
	WHERE id = $11`

	_, err := r.pool.Exec(ctx, query,
		settings.Name,
		settings.VATNumber,
		settings.BusinessAddress,
		settings.BusinessPhone,
		settings.BusinessEmail,
		settings.BankName,
		settings.BankAccount,
		settings.BankIBAN,
		time.Now(),
		userID,
		settings.ID,
	)

	return err
}
