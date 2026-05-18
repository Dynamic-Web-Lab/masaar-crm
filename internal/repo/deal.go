package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maidulcu/masaar-crm/internal/domain"
)

type DealRepo struct {
	db *pgxpool.Pool
}

func NewDealRepo(db *pgxpool.Pool) *DealRepo {
	return &DealRepo{db: db}
}

func (r *DealRepo) List(ctx context.Context, ownerID *uuid.UUID, page, limit int) (*domain.PaginatedResult[domain.Deal], error) {
	offset := (page - 1) * limit

	const countQ = `SELECT COUNT(*) FROM deals WHERE ($1::uuid IS NULL OR owner_id = $1)`
	var total int
	if err := r.db.QueryRow(ctx, countQ, ownerID).Scan(&total); err != nil {
		return nil, fmt.Errorf("count deals: %w", err)
	}

	const q = `
		SELECT id, lead_id, title, stage, amount, currency, close_date, probability, owner_id, created_at, updated_at
		FROM deals
		WHERE ($1::uuid IS NULL OR owner_id = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, q, ownerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list deals: %w", err)
	}
	defer rows.Close()

	var deals []domain.Deal
	for rows.Next() {
		var d domain.Deal
		if err := rows.Scan(
			&d.ID, &d.LeadID, &d.Title, &d.Stage,
			&d.Amount, &d.Currency, &d.CloseDate,
			&d.Probability, &d.OwnerID,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan deal: %w", err)
		}
		deals = append(deals, d)
	}
	return &domain.PaginatedResult[domain.Deal]{
		Data: deals, Total: total, Page: page, Limit: limit,
	}, nil
}

func (r *DealRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Deal, error) {
	const q = `
		SELECT id, lead_id, title, stage, amount, currency, close_date, probability, owner_id, created_at, updated_at
		FROM deals WHERE id = $1
	`
	d := &domain.Deal{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&d.ID, &d.LeadID, &d.Title, &d.Stage,
		&d.Amount, &d.Currency, &d.CloseDate,
		&d.Probability, &d.OwnerID,
		&d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get deal: %w", err)
	}
	return d, nil
}

func (r *DealRepo) Create(ctx context.Context, d *domain.Deal) error {
	const q = `
		INSERT INTO deals (id, lead_id, title, stage, amount, currency, close_date, probability, owner_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING created_at, updated_at
	`
	d.ID = uuid.New()
	return r.db.QueryRow(ctx, q,
		d.ID, d.LeadID, d.Title, d.Stage,
		d.Amount, d.Currency, d.CloseDate,
		d.Probability, d.OwnerID,
	).Scan(&d.CreatedAt, &d.UpdatedAt)
}

func (r *DealRepo) Update(ctx context.Context, d *domain.Deal) error {
	const q = `
		UPDATE deals SET amount=$1, probability=$2, close_date=$3, updated_at=NOW()
		WHERE id=$4
	`
	_, err := r.db.Exec(ctx, q, d.Amount, d.Probability, d.CloseDate, d.ID)
	return err
}

func (r *DealRepo) UpdateStage(ctx context.Context, id uuid.UUID, stage domain.DealStage) error {
	_, err := r.db.Exec(ctx,
		`UPDATE deals SET stage=$1, updated_at=NOW() WHERE id=$2`,
		stage, id,
	)
	return err
}

func (r *DealRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM deals WHERE id=$1`, id)
	return err
}
