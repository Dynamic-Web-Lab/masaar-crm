package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Company struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type CompanyRepo struct {
	db *pgxpool.Pool
}

func NewCompanyRepo(db *pgxpool.Pool) *CompanyRepo {
	return &CompanyRepo{db: db}
}

func (r *CompanyRepo) List(ctx context.Context) ([]Company, error) {
	const q = `SELECT id, name FROM companies ORDER BY created_at ASC`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list companies: %w", err)
	}
	defer rows.Close()

	var companies []Company
	for rows.Next() {
		var c Company
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, fmt.Errorf("scan company: %w", err)
		}
		companies = append(companies, c)
	}
	return companies, nil
}
