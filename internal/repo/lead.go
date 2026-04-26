package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maidulcu/masaar-crm/internal/domain"
)

type LeadRepo struct {
	db *pgxpool.Pool
}

func NewLeadRepo(db *pgxpool.Pool) *LeadRepo {
	return &LeadRepo{db: db}
}

// KanbanBoard returns leads grouped by stage with joined contact.
func (r *LeadRepo) KanbanBoard(ctx context.Context) (map[domain.LeadStage][]domain.Lead, error) {
	const q = `
		SELECT l.id, l.contact_id, l.stage, l.source, l.deal_value, l.currency, l.notes,
		       l.created_at, l.updated_at,
		       c.id, c.phone_wa, c.full_name, c.email, c.language, c.lead_score
		FROM leads l
		JOIN contacts c ON c.id = l.contact_id
		ORDER BY l.stage, l.created_at DESC
	`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("kanban query: %w", err)
	}
	defer rows.Close()

	board := map[domain.LeadStage][]domain.Lead{}
	for rows.Next() {
		var l domain.Lead
		var c domain.Contact
		if err := rows.Scan(
			&l.ID, &l.ContactID, &l.Stage, &l.Source,
			&l.DealValue, &l.Currency, &l.Notes,
			&l.CreatedAt, &l.UpdatedAt,
			&c.ID, &c.PhoneWA, &c.FullName, &c.Email, &c.Language, &c.LeadScore,
		); err != nil {
			return nil, fmt.Errorf("scan kanban row: %w", err)
		}
		l.Contact = &c
		board[l.Stage] = append(board[l.Stage], l)
	}
	return board, nil
}

func (r *LeadRepo) Create(ctx context.Context, l *domain.Lead) error {
	const q = `
		INSERT INTO leads (id, contact_id, stage, source, deal_value, currency, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING created_at, updated_at
	`
	l.ID = uuid.New()
	return r.db.QueryRow(ctx, q,
		l.ID, l.ContactID, l.Stage, l.Source, l.DealValue, l.Currency, l.Notes,
	).Scan(&l.CreatedAt, &l.UpdatedAt)
}

func (r *LeadRepo) UpdateNotes(ctx context.Context, id uuid.UUID, notes string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE leads SET notes=$1, updated_at=NOW() WHERE id=$2`,
		notes, id,
	)
	return err
}

func (r *LeadRepo) UpdateStage(ctx context.Context, id uuid.UUID, stage domain.LeadStage) error {
	_, err := r.db.Exec(ctx,
		`UPDATE leads SET stage=$1, updated_at=NOW() WHERE id=$2`,
		stage, id,
	)
	return err
}

func (r *LeadRepo) UpdateScore(ctx context.Context, id uuid.UUID, score int) error {
	_, err := r.db.Exec(ctx,
		`UPDATE leads SET lead_score=$1, score_updated_at=NOW(), updated_at=NOW() WHERE id=$2`,
		score, id,
	)
	return err
}

func (r *LeadRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Lead, error) {
	const q = `
		SELECT l.id, l.contact_id, l.stage, l.source, l.deal_value, l.currency, l.notes,
		       l.lead_score, l.score_updated_at, l.created_at, l.updated_at
		FROM leads l WHERE l.id = $1
	`
	l := &domain.Lead{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&l.ID, &l.ContactID, &l.Stage, &l.Source,
		&l.DealValue, &l.Currency, &l.Notes,
		&l.LeadScore, &l.ScoreUpdatedAt, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get lead by id: %w", err)
	}
	return l, nil
}
