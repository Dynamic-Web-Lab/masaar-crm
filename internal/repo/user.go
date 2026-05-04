package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maidulcu/masaar-crm/internal/domain"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `
		SELECT id, name, email, password_hash, role, lang_pref, wa_number, created_at
		FROM users
		WHERE email = $1
	`
	u := &domain.User{}
	err := r.db.QueryRow(ctx, q, strings.ToLower(strings.TrimSpace(email))).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash,
		&u.Role, &u.LangPref, &u.WANumber, &u.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return u, nil
}

func (r *UserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const q = `
		SELECT id, name, email, password_hash, role, lang_pref, wa_number, created_at
		FROM users
		WHERE id = $1
	`
	u := &domain.User{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash,
		&u.Role, &u.LangPref, &u.WANumber, &u.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return u, nil
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	const q = `
		INSERT INTO users (id, name, email, password_hash, role, lang_pref, wa_number)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING created_at
	`
	u.ID = uuid.New()
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	return r.db.QueryRow(ctx, q,
		u.ID, u.Name, u.Email, u.PasswordHash,
		u.Role, u.LangPref, u.WANumber,
	).Scan(&u.CreatedAt)
}

func (r *UserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	const q = `UPDATE users SET password_hash = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, q, passwordHash, id)
	return err
}

func (r *UserRepo) UpdateLangPref(ctx context.Context, id uuid.UUID, lang string) error {
	const q = `UPDATE users SET lang_pref = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, q, lang, id)
	return err
}
