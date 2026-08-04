package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
)

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type otpRequest struct {
	Email string `json:"email"`
}

type otpVerifyRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func (s *Service) Routes(r chi.Router) {
	r.Post("/auth/register", s.handleRegister)
	r.Post("/auth/login", s.handleLogin)
	r.Post("/auth/logout", s.handleLogout)
	r.Get("/auth/me", s.handleMe)
	r.Delete("/auth/account", s.handleDeleteAccount)
	r.Post("/auth/otp/request", s.handleOTPRequest)
	r.Post("/auth/otp/verify", s.handleOTPVerify)
	r.Get("/auth/google", s.handleGoogleStart)
	r.Get("/auth/google/callback", s.handleGoogleCallback)
}

func (s *Service) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	user, token, expires, err := s.Register(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		s.writeAuthError(w, err)
		return
	}
	s.setSessionCookie(w, token, expires)
	writeJSON(w, http.StatusCreated, map[string]any{"user": user})
}

func (s *Service) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	user, token, expires, err := s.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		s.writeAuthError(w, err)
		return
	}
	s.setSessionCookie(w, token, expires)
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (s *Service) handleLogout(w http.ResponseWriter, r *http.Request) {
	c, _ := r.Cookie(SessionCookieName)
	token := ""
	if c != nil {
		token = c.Value
	}
	_ = s.Logout(r.Context(), token)
	s.clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Service) handleMe(w http.ResponseWriter, r *http.Request) {
	user, err := s.userFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (s *Service) handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	user, err := s.userFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	if err := s.DeleteAccount(r.Context(), user.ID); err != nil {
		s.Log.Error("delete account", "error", err, "user_id", user.ID)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete account"})
		return
	}
	s.clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Service) handleOTPRequest(w http.ResponseWriter, r *http.Request) {
	var req otpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	if err := s.RequestOTP(r.Context(), req.Email); err != nil {
		s.writeAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "sent",
		"message": "If the email is valid, a code has been sent",
	})
}

func (s *Service) handleOTPVerify(w http.ResponseWriter, r *http.Request) {
	var req otpVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	user, token, expires, err := s.VerifyOTP(r.Context(), req.Email, req.Code)
	if err != nil {
		s.writeAuthError(w, err)
		return
	}
	s.setSessionCookie(w, token, expires)
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (s *Service) handleGoogleStart(w http.ResponseWriter, r *http.Request) {
	state, err := NewOAuthState()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to start google auth"})
		return
	}
	authURL, err := s.GoogleAuthURL(state)
	if err != nil {
		s.writeAuthError(w, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "dripfind_oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((10 * time.Minute).Seconds()),
	})
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (s *Service) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	q := url.Values{}
	fail := func(msg string) {
		q.Set("error", msg)
		http.Redirect(w, r, FrontendRedirect(s.FrontendURL, "/auth", q), http.StatusFound)
	}

	stateCookie, err := r.Cookie("dripfind_oauth_state")
	if err != nil || stateCookie.Value == "" || stateCookie.Value != r.URL.Query().Get("state") {
		fail("invalid_oauth_state")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "dripfind_oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		fail(errMsg)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		fail("missing_code")
		return
	}

	_, token, expires, err := s.CompleteGoogle(r.Context(), code)
	if err != nil {
		s.Log.Error("google auth", "error", err)
		fail("google_auth_failed")
		return
	}
	s.setSessionCookie(w, token, expires)
	http.Redirect(w, r, FrontendRedirect(s.FrontendURL, "/studio", nil), http.StatusFound)
}

func (s *Service) setSessionCookie(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  expires,
	})
}

func (s *Service) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   s.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func (s *Service) writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrEmailTaken):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "email already registered"})
	case errors.Is(err, ErrInvalidCreds):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
	case errors.Is(err, ErrNoPassword):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "use Google or email code to sign in to this account"})
	case errors.Is(err, ErrInvalidOTP):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired code"})
	case errors.Is(err, ErrOTPTooMany):
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "too many attempts, request a new code"})
	case errors.Is(err, ErrGoogleNotConf):
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "google sign-in is not configured"})
	default:
		msg := err.Error()
		if msg == "email is required" || msg == "invalid email" ||
			msg == "password must be at least 8 characters" || msg == "password is too long" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": msg})
			return
		}
		s.Log.Error("auth error", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "authentication failed"})
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
