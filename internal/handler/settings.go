package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/temple-ats/TempleATS/internal/db"
	"github.com/temple-ats/TempleATS/internal/email"
	mw "github.com/temple-ats/TempleATS/internal/middleware"
)

type updateOrgDefaultsRequest struct {
	DefaultCompanyBlurb     string `json:"defaultCompanyBlurb"`
	DefaultClosingStatement string `json:"defaultClosingStatement"`
}

type updateOrgNameRequest struct {
	Name string `json:"name"`
}

// GetOrgName returns the current organization name.
func (s *Server) GetOrgName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID := mw.GetOrgID(ctx)

	org, err := s.Queries.GetOrganizationByID(ctx, orgID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get organization"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"name": org.Name})
}

// UpdateOrgName updates the organization name.
func (s *Server) UpdateOrgName(w http.ResponseWriter, r *http.Request) {
	var req updateOrgNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}

	ctx := r.Context()
	orgID := mw.GetOrgID(ctx)

	org, err := s.Queries.GetOrganizationByID(ctx, orgID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get organization"})
		return
	}

	_, err = s.Queries.UpdateOrganization(ctx, db.UpdateOrganizationParams{
		ID:      orgID,
		Name:    req.Name,
		Slug:    org.Slug,
		LogoUrl: org.LogoUrl,
		Website: org.Website,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update organization name"})
		return
	}

	go db.InsertAuditLog(context.Background(), s.Pool, orgID, mw.GetUserID(ctx), "update", "organization", orgID, map[string]string{"name": req.Name})
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// GetOrgDefaults returns the org's default company blurb and closing statement.
func (s *Server) GetOrgDefaults(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID := mw.GetOrgID(ctx)

	defaults, err := s.Queries.GetOrgDefaults(ctx, orgID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get defaults"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"defaultCompanyBlurb":     defaults.DefaultCompanyBlurb,
		"defaultClosingStatement": defaults.DefaultClosingStatement,
	})
}

// UpdateOrgDefaults updates the org's default company blurb and closing statement.
func (s *Server) UpdateOrgDefaults(w http.ResponseWriter, r *http.Request) {
	var req updateOrgDefaultsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	ctx := r.Context()
	orgID := mw.GetOrgID(ctx)

	_, err := s.Queries.UpdateOrgDefaults(ctx, db.UpdateOrgDefaultsParams{
		ID:                      orgID,
		DefaultCompanyBlurb:     req.DefaultCompanyBlurb,
		DefaultClosingStatement: req.DefaultClosingStatement,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update defaults"})
		return
	}

	go db.InsertAuditLog(context.Background(), s.Pool, orgID, mw.GetUserID(ctx), "update", "settings", orgID, map[string]string{"field": "org_defaults"})
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// --- SMTP Settings ---

type smtpSettingsRequest struct {
	Host      string `json:"host"`
	Port      int32  `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	FromEmail string `json:"fromEmail"`
	FromName  string `json:"fromName"`
	TLS       bool   `json:"tls"`
}

// GetSmtpSettings returns the org's SMTP configuration (password masked).
func (s *Server) GetSmtpSettings(w http.ResponseWriter, r *http.Request) {
	orgID := mw.GetOrgID(r.Context())
	settings, err := s.Queries.GetSmtpSettings(r.Context(), orgID)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"configured": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"configured": true,
		"host":       settings.Host,
		"port":       settings.Port,
		"username":   settings.Username,
		"password":   "••••••••",
		"fromEmail":  settings.FromEmail,
		"fromName":   settings.FromName,
		"tls":        settings.TlsEnabled,
	})
}

// UpdateSmtpSettings creates or updates the org's SMTP configuration.
func (s *Server) UpdateSmtpSettings(w http.ResponseWriter, r *http.Request) {
	var req smtpSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Host == "" || req.Username == "" || req.Password == "" || req.FromEmail == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "host, username, password, and fromEmail are required"})
		return
	}
	if req.Port == 0 {
		req.Port = 587
	}

	orgID := mw.GetOrgID(r.Context())

	// If password is masked, keep the existing one
	if req.Password == "••••••••" {
		existing, err := s.Queries.GetSmtpSettings(r.Context(), orgID)
		if err == nil {
			req.Password = existing.Password
		}
	}

	_, err := s.Queries.UpsertSmtpSettings(r.Context(), db.UpsertSmtpSettingsParams{
		OrganizationID: orgID,
		Host:           req.Host,
		Port:           req.Port,
		Username:       req.Username,
		Password:       req.Password,
		FromEmail:      req.FromEmail,
		FromName:       req.FromName,
		TlsEnabled:     req.TLS,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save SMTP settings"})
		return
	}
	go db.InsertAuditLog(context.Background(), s.Pool, orgID, mw.GetUserID(r.Context()), "update", "settings", orgID, map[string]string{"field": "smtp"})
	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

// TestSmtpSettings sends a test email to the current user.
func (s *Server) TestSmtpSettings(w http.ResponseWriter, r *http.Request) {
	orgID := mw.GetOrgID(r.Context())
	userID := mw.GetUserID(r.Context())

	config := s.Email.GetConfig(r.Context(), orgID)
	if config == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "SMTP not configured"})
		return
	}

	user, err := s.Queries.GetUserByID(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get user"})
		return
	}

	subject := "TempleATS - Test Email"
	body := fmt.Sprintf("<p>Hi %s,</p><p>This is a test email from TempleATS. Your SMTP settings are working correctly!</p>", user.Name)

	if err := email.SendEmail(config, user.Email, subject, body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "SMTP test failed: " + err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "sent", "sentTo": user.Email})
}

// --- Email Templates ---

type emailTemplateRequest struct {
	Stage   string `json:"stage"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Enabled bool   `json:"enabled"`
}

// GetEmailTemplates returns all stage email templates for the org.
func (s *Server) GetEmailTemplates(w http.ResponseWriter, r *http.Request) {
	orgID := mw.GetOrgID(r.Context())
	templates, err := s.Queries.ListEmailTemplatesByOrg(r.Context(), orgID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list templates"})
		return
	}
	writeJSON(w, http.StatusOK, templates)
}

// UpdateEmailTemplate creates or updates a stage email template.
func (s *Server) UpdateEmailTemplate(w http.ResponseWriter, r *http.Request) {
	var req emailTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Stage == "" || req.Subject == "" || req.Body == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "stage, subject, and body are required"})
		return
	}

	orgID := mw.GetOrgID(r.Context())
	tmpl, err := s.Queries.UpsertEmailTemplate(r.Context(), db.UpsertEmailTemplateParams{
		OrganizationID: orgID,
		Stage:          req.Stage,
		Subject:        req.Subject,
		Body:           req.Body,
		Enabled:        req.Enabled,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save template"})
		return
	}
	writeJSON(w, http.StatusOK, tmpl)
}
