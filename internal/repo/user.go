package repo

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

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

func (r *UserRepo) List(ctx context.Context) ([]domain.User, error) {
	const q = `
		SELECT id, name, email, password_hash, role, lang_pref, wa_number, created_at
		FROM users ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash,
			&u.Role, &u.LangPref, &u.WANumber, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	const q = `UPDATE users SET password_hash = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, q, passwordHash, id)
	return err
}

// CreatePasswordResetToken generates a 32-byte random token, stores its SHA-256
// hash in the DB (1-hour TTL), and returns the plaintext token to send by email.
func (r *UserRepo) CreatePasswordResetToken(ctx context.Context, userID uuid.UUID) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate reset token: %w", err)
	}
	plaintext := hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(plaintext))
	hash := hex.EncodeToString(sum[:])

	// Invalidate any previous tokens for this user first
	_, _ = r.db.Exec(ctx, `DELETE FROM password_reset_tokens WHERE user_id = $1`, userID)

	const q = `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`
	if _, err := r.db.Exec(ctx, q, userID, hash, time.Now().Add(time.Hour)); err != nil {
		return "", fmt.Errorf("store reset token: %w", err)
	}
	return plaintext, nil
}

// ConsumePasswordResetToken validates the plaintext token, marks it used, and
// returns the associated user_id. Returns an error if expired or already used.
func (r *UserRepo) ConsumePasswordResetToken(ctx context.Context, plaintext string) (uuid.UUID, error) {
	sum := sha256.Sum256([]byte(plaintext))
	hash := hex.EncodeToString(sum[:])

	var userID uuid.UUID
	var expiresAt time.Time
	var usedAt *time.Time

	const q = `
		SELECT user_id, expires_at, used_at
		FROM password_reset_tokens
		WHERE token_hash = $1
	`
	err := r.db.QueryRow(ctx, q, hash).Scan(&userID, &expiresAt, &usedAt)
	if err != nil {
		return uuid.Nil, fmt.Errorf("token not found")
	}
	if usedAt != nil {
		return uuid.Nil, fmt.Errorf("token already used")
	}
	if time.Now().After(expiresAt) {
		return uuid.Nil, fmt.Errorf("token expired")
	}

	// Mark as used
	_, _ = r.db.Exec(ctx, `UPDATE password_reset_tokens SET used_at = NOW() WHERE token_hash = $1`, hash)
	return userID, nil
}

func (r *UserRepo) UpdateLangPref(ctx context.Context, id uuid.UUID, lang string) error {
	const q = `UPDATE users SET lang_pref = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, q, lang, id)
	return err
}

func (r *UserRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, q, email).Scan(&exists)
	return exists, err
}

func (r *UserRepo) CreateWithDefaults(ctx context.Context, email string, langPref string) (*domain.User, error) {
	const q = `
		INSERT INTO users (id, name, email, password_hash, role, lang_pref, wa_number)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at
	`
	u := &domain.User{
		ID:           uuid.New(),
		Name:         email, // Use email as initial name
		Email:        email,
		PasswordHash: "", // No password for magic link users
		Role:         domain.RoleViewer, // Default role for magic link signups
		LangPref:     langPref,
		WANumber:     "",
	}
	err := r.db.QueryRow(ctx, q,
		u.ID, u.Name, u.Email, u.PasswordHash,
		u.Role, u.LangPref, u.WANumber,
	).Scan(&u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}
