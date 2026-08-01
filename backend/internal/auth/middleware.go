package auth

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxKey int

const userCtxKey ctxKey = 1

func ContextWithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userCtxKey, user)
}

func UserFromContext(ctx context.Context) (*User, bool) {
	u, ok := ctx.Value(userCtxKey).(*User)
	return u, ok && u != nil
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	u, ok := UserFromContext(ctx)
	if !ok {
		return uuid.Nil, false
	}
	return u.ID, true
}

func (s *Service) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := s.userFromRequest(r)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		next.ServeHTTP(w, r.WithContext(ContextWithUser(r.Context(), user)))
	})
}

func (s *Service) OptionalUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := s.userFromRequest(r)
		if err == nil {
			r = r.WithContext(ContextWithUser(r.Context(), user))
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Service) userFromRequest(r *http.Request) (*User, error) {
	c, err := r.Cookie(SessionCookieName)
	if err != nil || c.Value == "" {
		return nil, ErrNotFound
	}
	return s.UserFromSession(r.Context(), c.Value)
}
