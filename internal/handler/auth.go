package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/temple-ats/TempleATS/internal/auth"
	"github.com/temple-ats/TempleATS/internal/db"
	mw "github.com/temple-ats/TempleATS/internal/middleware"
)

type signupRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
	OrgName  string `json:"orgName"`
	OrgSlug  string `json:"orgSlug"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Role    string `json:"role"`
	OrgID   string `json:"orgId"`
	OrgSlug string `json:"orgSlug"`
	OrgName string `json:"orgName"`
}

// Signup creates a new organization and user, then returns a JWT.
func (s *Server) Signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" || req.OrgName == "" || req.OrgSlug == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "all fields are required"})
		return
	}

	req.OrgSlug = strings.ToLower(strings.TrimSpace(req.OrgSlug))

	ctx := r.Context()

	// Create org and user in a transaction
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer tx.Rollback(ctx)

	qtx := s.Queries.WithTx(tx)

	org, err := qtx.CreateOrganization(ctx, db.CreateOrganizationParams{
		Name: req.OrgName,
		Slug: req.OrgSlug,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "organization slug already taken"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create organization"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	user, err := qtx.CreateUser(ctx, db.CreateUserParams{
		Email:          req.Email,
		Name:           req.Name,
		PasswordHash:   string(hash),
		Role:           "admin",
		OrganizationID: org.ID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "email already registered"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	token, err := auth.GenerateToken(user.ID, org.ID, org.Slug, user.Role)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		return
	}

	setTokenCookie(w, token)
	writeJSON(w, http.StatusCreated, userResponse{
		ID:      user.ID,
		Email:   user.Email,
		Name:    user.Name,
		Role:    user.Role,
		OrgID:   org.ID,
		OrgSlug: org.Slug,
		OrgName: org.Name,
	})
}

// Login validates credentials and returns a JWT.
func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
		return
	}

	ctx := r.Context()

	user, err := s.Queries.GetUserByEmail(ctx, req.Email)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	org, err := s.Queries.GetOrganizationByID(ctx, user.OrganizationID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	token, err := auth.GenerateToken(user.ID, org.ID, org.Slug, user.Role)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		return
	}

	setTokenCookie(w, token)
	writeJSON(w, http.StatusOK, userResponse{
		ID:      user.ID,
		Email:   user.Email,
		Name:    user.Name,
		Role:    user.Role,
		OrgID:   org.ID,
		OrgSlug: org.Slug,
		OrgName: org.Name,
	})
}

// Me returns the current authenticated user's info.
func (s *Server) Me(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := mw.GetUserID(ctx)

	user, err := s.Queries.GetUserByID(ctx, userID)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "user not found"})
		return
	}

	org, err := s.Queries.GetOrganizationByID(context.Background(), user.OrganizationID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, userResponse{
		ID:      user.ID,
		Email:   user.Email,
		Name:    user.Name,
		Role:    user.Role,
		OrgID:   org.ID,
		OrgSlug: org.Slug,
		OrgName: org.Name,
	})
}

// Logout clears the auth cookie.
func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func setTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		MaxAge:   86400, // 24 hours
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
