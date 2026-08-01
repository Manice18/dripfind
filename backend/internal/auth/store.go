package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrEmailTaken    = errors.New("email already registered")
	ErrInvalidCreds  = errors.New("invalid email or password")
	ErrInvalidOTP    = errors.New("invalid or expired code")
	ErrOTPTooMany    = errors.New("too many attempts")
	ErrNoPassword    = errors.New("password login not available for this account")
	ErrGoogleNotConf = errors.New("google auth is not configured")
)

type User struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) CreateUser(ctx context.Context, email, passwordHash, name string) (*User, error) {
	email = normalizeEmail(email)
	var u User
	err := s.pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, name, email_verified)
		VALUES ($1, $2, $3, FALSE)
		RETURNING id, email, name, email_verified, created_at
	`, email, nullIfEmpty(passwordHash), strings.TrimSpace(name)).Scan(
		&u.ID, &u.Email, &u.Name, &u.EmailVerified, &u.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrEmailTaken
		}
		return nil, err
	}
	return &u, nil
}

func (s *Store) FindByEmail(ctx context.Context, email string) (*User, string, error) {
	email = normalizeEmail(email)
	var u User
	var hash *string
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, name, email_verified, created_at, password_hash
		FROM users WHERE email=$1
	`, email).Scan(&u.ID, &u.Email, &u.Name, &u.EmailVerified, &u.CreatedAt, &hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", ErrNotFound
	}
	if err != nil {
		return nil, "", err
	}
	if hash == nil {
		return &u, "", nil
	}
	return &u, *hash, nil
}

func (s *Store) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var u User
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, name, email_verified, created_at
		FROM users WHERE id=$1
	`, id).Scan(&u.ID, &u.Email, &u.Name, &u.EmailVerified, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) FindOrCreateGoogle(ctx context.Context, email, googleID, name string) (*User, error) {
	email = normalizeEmail(email)
	var u User
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, name, email_verified, created_at
		FROM users WHERE google_id=$1
	`, googleID).Scan(&u.ID, &u.Email, &u.Name, &u.EmailVerified, &u.CreatedAt)
	if err == nil {
		return &u, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// Link Google to an existing email account if present.
	err = s.pool.QueryRow(ctx, `
		UPDATE users
		SET google_id=$2, name=CASE WHEN name='' THEN $3 ELSE name END,
		    email_verified=TRUE, updated_at=NOW()
		WHERE email=$1
		RETURNING id, email, name, email_verified, created_at
	`, email, googleID, strings.TrimSpace(name)).Scan(
		&u.ID, &u.Email, &u.Name, &u.EmailVerified, &u.CreatedAt,
	)
	if err == nil {
		return &u, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	err = s.pool.QueryRow(ctx, `
		INSERT INTO users (email, google_id, name, email_verified)
		VALUES ($1, $2, $3, TRUE)
		RETURNING id, email, name, email_verified, created_at
	`, email, googleID, strings.TrimSpace(name)).Scan(
		&u.ID, &u.Email, &u.Name, &u.EmailVerified, &u.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrEmailTaken
		}
		return nil, err
	}
	return &u, nil
}

func (s *Store) MarkEmailVerified(ctx context.Context, userID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE users SET email_verified=TRUE, updated_at=NOW() WHERE id=$1
	`, userID)
	return err
}

func (s *Store) CreateSession(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO sessions (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, tokenHash, expiresAt)
	return err
}

func (s *Store) UserBySessionToken(ctx context.Context, tokenHash string) (*User, uuid.UUID, error) {
	var u User
	var sessionID uuid.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT u.id, u.email, u.name, u.email_verified, u.created_at, s.id
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash=$1 AND s.expires_at > NOW()
	`, tokenHash).Scan(&u.ID, &u.Email, &u.Name, &u.EmailVerified, &u.CreatedAt, &sessionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, uuid.Nil, ErrNotFound
	}
	if err != nil {
		return nil, uuid.Nil, err
	}
	return &u, sessionID, nil
}

func (s *Store) DeleteSessionByToken(ctx context.Context, tokenHash string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash=$1`, tokenHash)
	return err
}

func (s *Store) DeleteExpiredSessions(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at <= NOW()`)
	return err
}

func (s *Store) UpsertOTP(ctx context.Context, email, codeHash, purpose string, expiresAt time.Time) error {
	email = normalizeEmail(email)
	_, err := s.pool.Exec(ctx, `DELETE FROM email_otps WHERE email=$1 AND purpose=$2`, email, purpose)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO email_otps (email, code_hash, purpose, expires_at)
		VALUES ($1, $2, $3, $4)
	`, email, codeHash, purpose, expiresAt)
	return err
}

type otpRow struct {
	ID        uuid.UUID
	CodeHash  string
	Attempts  int
	ExpiresAt time.Time
}

func (s *Store) GetOTP(ctx context.Context, email, purpose string) (*otpRow, error) {
	email = normalizeEmail(email)
	var row otpRow
	err := s.pool.QueryRow(ctx, `
		SELECT id, code_hash, attempts, expires_at
		FROM email_otps
		WHERE email=$1 AND purpose=$2
		ORDER BY created_at DESC
		LIMIT 1
	`, email, purpose).Scan(&row.ID, &row.CodeHash, &row.Attempts, &row.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Store) IncrementOTPAttempts(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `UPDATE email_otps SET attempts=attempts+1 WHERE id=$1`, id)
	return err
}

func (s *Store) DeleteOTP(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM email_otps WHERE id=$1`, id)
	return err
}

// DeleteUser removes the user and relies on FK CASCADE for sessions, outfits,
// history, clothing_items, and product_matches. OTP rows (keyed by email) are
// cleared explicitly.
func (s *Store) DeleteUser(ctx context.Context, id uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var email string
	err = tx.QueryRow(ctx, `SELECT email FROM users WHERE id=$1`, id).Scan(&email)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM email_otps WHERE email=$1`, email); err != nil {
		return err
	}
	ct, err := tx.Exec(ctx, `DELETE FROM users WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return tx.Commit(ctx)
}

func (s *Store) FindOrCreateByEmail(ctx context.Context, email string) (*User, error) {
	email = normalizeEmail(email)
	u, _, err := s.FindByEmail(ctx, email)
	if err == nil {
		return u, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	var created User
	err = s.pool.QueryRow(ctx, `
		INSERT INTO users (email, email_verified)
		VALUES ($1, TRUE)
		RETURNING id, email, name, email_verified, created_at
	`, email).Scan(&created.ID, &created.Email, &created.Name, &created.EmailVerified, &created.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			u2, _, err2 := s.FindByEmail(ctx, email)
			return u2, err2
		}
		return nil, err
	}
	return &created, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "23505")
}
