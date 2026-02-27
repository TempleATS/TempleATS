package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/temple-ats/TempleATS/internal/auth"
	"github.com/temple-ats/TempleATS/internal/db"
)

type acceptInviteRequest struct {
	Token    string `json:"token"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// AcceptInvitation is a public endpoint that creates a user from an invitation.
func (s *Server) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	var req acceptInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Token == "" || req.Name == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "token, name, and password are required"})
		return
	}

	ctx := r.Context()

	// Look up the invitation
	invitation, err := s.Queries.GetInvitationByToken(ctx, req.Token)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "invitation not found or expired"})
		return
	}

	// Create user in a transaction
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	defer tx.Rollback(ctx)

	qtx := s.Queries.WithTx(tx)

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	user, err := qtx.CreateUser(ctx, db.CreateUserParams{
		Email:          invitation.Email,
		Name:           req.Name,
		PasswordHash:   string(hash),
		Role:           invitation.Role,
		OrganizationID: invitation.OrganizationID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "email already registered"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
		return
	}

	// Mark invitation as accepted
	_, err = qtx.AcceptInvitation(ctx, invitation.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	// Get org info for token
	org, err := s.Queries.GetOrganizationByID(ctx, invitation.OrganizationID)
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
