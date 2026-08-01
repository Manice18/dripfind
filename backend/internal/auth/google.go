package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleProfile struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
}

func (s *Service) googleConfigured() bool {
	return s.GoogleClientID != "" && s.GoogleClientSecret != ""
}

func (s *Service) googleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     s.GoogleClientID,
		ClientSecret: s.GoogleClientSecret,
		RedirectURL:  strings.TrimRight(s.APIPublicURL, "/") + "/auth/google/callback",
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

func (s *Service) GoogleAuthURL(state string) (string, error) {
	if !s.googleConfigured() {
		return "", ErrGoogleNotConf
	}
	return s.googleOAuthConfig().AuthCodeURL(state, oauth2.AccessTypeOnline), nil
}

func (s *Service) CompleteGoogle(ctx context.Context, code string) (*User, string, time.Time, error) {
	if !s.googleConfigured() {
		return nil, "", time.Time{}, ErrGoogleNotConf
	}
	tok, err := s.googleOAuthConfig().Exchange(ctx, code)
	if err != nil {
		return nil, "", time.Time{}, fmt.Errorf("google token exchange: %w", err)
	}

	profile, err := fetchGoogleProfile(ctx, tok.AccessToken)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	if profile.Email == "" || profile.ID == "" {
		return nil, "", time.Time{}, errors.New("google profile missing email")
	}

	user, err := s.Store.FindOrCreateGoogle(ctx, profile.Email, profile.ID, profile.Name)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	token, expires, err := s.createSession(ctx, user.ID)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	return user, token, expires, nil
}

func fetchGoogleProfile(ctx context.Context, accessToken string) (*GoogleProfile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("google userinfo %s: %s", res.Status, string(body))
	}
	var profile GoogleProfile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

func NewOAuthState() (string, error) {
	return NewSessionToken()
}

func FrontendRedirect(frontendURL, path string, q url.Values) string {
	u, err := url.Parse(strings.TrimRight(frontendURL, "/") + path)
	if err != nil {
		return frontendURL
	}
	if q != nil {
		u.RawQuery = q.Encode()
	}
	return u.String()
}
