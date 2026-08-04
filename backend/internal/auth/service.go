package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	SessionCookieName = "dripfind_session"
	SessionTTL        = 30 * 24 * time.Hour
	OTPTTL            = 10 * time.Minute
	MaxOTPAttempts    = 5
)

type Service struct {
	Store              *Store
	Mailer             Mailer
	Log                *slog.Logger
	SessionSecret      string // reserved for future signed cookies; sessions are DB-backed
	CookieSecure       bool
	FrontendURL        string
	APIPublicURL       string
	GoogleClientID     string
	GoogleClientSecret string
}

func (s *Service) Register(ctx context.Context, email, password, name string) (*User, string, time.Time, error) {
	email = normalizeEmail(email)
	if err := validateEmail(email); err != nil {
		return nil, "", time.Time{}, err
	}
	if err := validatePassword(password); err != nil {
		return nil, "", time.Time{}, err
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	user, err := s.Store.CreateUser(ctx, email, hash, name)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	token, expires, err := s.createSession(ctx, user.ID)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	return user, token, expires, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*User, string, time.Time, error) {
	email = normalizeEmail(email)
	user, hash, err := s.Store.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, "", time.Time{}, ErrInvalidCreds
		}
		return nil, "", time.Time{}, err
	}
	if hash == "" {
		return nil, "", time.Time{}, ErrNoPassword
	}
	if !CheckPassword(hash, password) {
		return nil, "", time.Time{}, ErrInvalidCreds
	}
	token, expires, err := s.createSession(ctx, user.ID)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	return user, token, expires, nil
}

func (s *Service) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	return s.Store.DeleteSessionByToken(ctx, HashToken(rawToken))
}

func (s *Service) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	return s.Store.DeleteUser(ctx, userID)
}

func (s *Service) UserFromSession(ctx context.Context, rawToken string) (*User, error) {
	if rawToken == "" {
		return nil, ErrNotFound
	}
	user, _, err := s.Store.UserBySessionToken(ctx, HashToken(rawToken))
	return user, err
}

func (s *Service) RequestOTP(ctx context.Context, email string) error {
	email = normalizeEmail(email)
	if err := validateEmail(email); err != nil {
		return err
	}
	code, err := NewOTPCode()
	if err != nil {
		return err
	}
	hash, err := HashOTP(code)
	if err != nil {
		return err
	}
	if err := s.Store.UpsertOTP(ctx, email, hash, "login", time.Now().Add(OTPTTL)); err != nil {
		return err
	}
	return s.Mailer.SendOTP(ctx, email, code)
}

func (s *Service) VerifyOTP(ctx context.Context, email, code string) (*User, string, time.Time, error) {
	email = normalizeEmail(email)
	code = strings.TrimSpace(code)
	row, err := s.Store.GetOTP(ctx, email, "login")
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, "", time.Time{}, ErrInvalidOTP
		}
		return nil, "", time.Time{}, err
	}
	if time.Now().After(row.ExpiresAt) {
		_ = s.Store.DeleteOTP(ctx, row.ID)
		return nil, "", time.Time{}, ErrInvalidOTP
	}
	if row.Attempts >= MaxOTPAttempts {
		return nil, "", time.Time{}, ErrOTPTooMany
	}
	if !CheckOTP(row.CodeHash, code) {
		_ = s.Store.IncrementOTPAttempts(ctx, row.ID)
		return nil, "", time.Time{}, ErrInvalidOTP
	}
	_ = s.Store.DeleteOTP(ctx, row.ID)

	user, err := s.Store.FindOrCreateByEmail(ctx, email)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	_ = s.Store.MarkEmailVerified(ctx, user.ID)
	user.EmailVerified = true

	token, expires, err := s.createSession(ctx, user.ID)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	return user, token, expires, nil
}

func (s *Service) createSession(ctx context.Context, userID uuid.UUID) (string, time.Time, error) {
	token, err := NewSessionToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expires := time.Now().Add(SessionTTL)
	if err := s.Store.CreateSession(ctx, userID, HashToken(token), expires); err != nil {
		return "", time.Time{}, err
	}
	return token, expires, nil
}

func validateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.New("invalid email")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if len(password) > 128 {
		return errors.New("password is too long")
	}
	return nil
}
