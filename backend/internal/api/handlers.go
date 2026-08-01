package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/manice18/outfit_finder/backend/internal/auth"
	"github.com/manice18/outfit_finder/backend/internal/history"
	"github.com/manice18/outfit_finder/backend/internal/models"
	"github.com/manice18/outfit_finder/backend/internal/pinterest"
	"github.com/manice18/outfit_finder/backend/internal/pipeline"
)

type Server struct {
	Store    *history.Store
	Pipeline *pipeline.Pipeline
	Auth     *auth.Service
	Log      *slog.Logger
	Origins  []string
	Images   http.Handler
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(s.logRequests)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   s.Origins,
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Handle("/images/*", http.StripPrefix("/images/", s.Images))

	s.Auth.Routes(r)

	r.Group(func(r chi.Router) {
		r.Use(s.Auth.RequireUser)
		r.Post("/analyze", s.handleAnalyze)
		r.Post("/analyze/upload", s.handleAnalyzeUpload)
		r.Get("/result/{id}", s.handleResult)
		r.Get("/history", s.handleHistory)
		r.Delete("/history/{id}", s.handleDeleteHistory)
	})

	return r
}

func (s *Server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		s.Log.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"bytes", ww.BytesWritten(),
			"latency_ms", time.Since(start).Milliseconds(),
		)
	})
}

type analyzeRequest struct {
	URL string `json:"url"`
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req analyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	req.URL = strings.TrimSpace(req.URL)
	if err := pinterest.ValidateURL(req.URL); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	outfit, err := s.Store.CreateOutfit(r.Context(), userID, req.URL)
	if err != nil {
		s.Log.Error("create outfit", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create analysis job")
		return
	}

	s.Pipeline.Start(r.Context(), outfit.ID, req.URL)
	writeJSON(w, http.StatusAccepted, map[string]string{"id": outfit.ID.String()})
}

const maxUploadBytes = 12 << 20 // 12 MB

func (s *Server) handleAnalyzeUpload(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes+1024)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		writeError(w, http.StatusBadRequest, "image too large or invalid multipart form (max 12MB)")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing image file (field name: image)")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxUploadBytes+1))
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read image")
		return
	}
	if len(data) == 0 {
		writeError(w, http.StatusBadRequest, "empty image file")
		return
	}
	if len(data) > maxUploadBytes {
		writeError(w, http.StatusBadRequest, "image too large (max 12MB)")
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = http.DetectContentType(data)
	}
	if !isAllowedImageType(contentType) {
		writeError(w, http.StatusBadRequest, "unsupported image type (use JPEG, PNG, WebP, or GIF)")
		return
	}

	filename := filepath.Base(header.Filename)
	if filename == "." || filename == "" {
		filename = "photo"
	}
	sourceLabel := "upload:" + filename

	outfit, err := s.Store.CreateOutfit(r.Context(), userID, sourceLabel)
	if err != nil {
		s.Log.Error("create outfit", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create analysis job")
		return
	}

	s.Pipeline.StartUpload(r.Context(), outfit.ID, data, contentType, sourceLabel)
	writeJSON(w, http.StatusAccepted, map[string]string{"id": outfit.ID.String()})
}

func isAllowedImageType(ct string) bool {
	ct = strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
	switch ct {
	case "image/jpeg", "image/jpg", "image/png", "image/webp", "image/gif":
		return true
	default:
		return false
	}
}

func (s *Server) handleResult(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	outfit, err := s.Store.GetOutfit(r.Context(), id, userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "result not found")
		return
	}
	if outfit.ImagePath != "" {
		outfit.ImageURL = "/images/" + outfit.ImagePath
	}
	writeJSON(w, http.StatusOK, outfit)
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	entries, err := s.Store.ListHistory(r.Context(), userID, 40)
	if err != nil {
		s.Log.Error("list history", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to load history")
		return
	}
	if entries == nil {
		entries = []models.HistoryEntry{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": entries})
}

func (s *Server) handleDeleteHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := s.Store.DeleteHistory(r.Context(), id, userID); err != nil {
		writeError(w, http.StatusNotFound, "history entry not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
