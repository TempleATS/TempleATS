package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/temple-ats/TempleATS/internal/db"
	mw "github.com/temple-ats/TempleATS/internal/middleware"
)

var validRoles = map[string]bool{
	"super_admin":    true,
	"admin":          true,
	"recruiter":      true,
	"hiring_manager": true,
	"interviewer":    true,
}

type inviteRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type updateMemberRequest struct {
	Role string `json:"role"`
}

// ListTeam returns all users and pending invitations for the org.
func (s *Server) ListTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID := mw.GetOrgID(ctx)

	users, err := s.Queries.ListUsersByOrg(ctx, orgID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list users"})
		return
	}

	invitations, err := s.Queries.ListInvitationsByOrg(ctx, orgID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list invitations"})
		return
	}

	// Strip password hashes from response
	type safeUser struct {
		ID             string             `json:"id"`
		Email          string             `json:"email"`
		Name           string             `json:"name"`
		Role           string             `json:"role"`
		OrganizationID string             `json:"organization_id"`
		CreatedAt      pgtype.Timestamptz `json:"created_at"`
	}
	safeUsers := make([]safeUser, len(users))
	for i, u := range users {
		safeUsers[i] = safeUser{
			ID:             u.ID,
			Email:          u.Email,
			Name:           u.Name,
			Role:           u.Role,
			OrganizationID: u.OrganizationID,
			CreatedAt:      u.CreatedAt,
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"members":     safeUsers,
		"invitations": invitations,
	})
}

// InviteUser creates a new invitation for the org.
func (s *Server) InviteUser(w http.ResponseWriter, r *http.Request) {
	var req inviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Email == "" || req.Role == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and role are required"})
		return
	}

	if !validRoles[req.Role] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid role"})
		return
	}

	ctx := r.Context()
	orgID := mw.GetOrgID(ctx)
	callerRole := mw.GetRole(ctx)

	// Recruiters can only invite hiring_manager or interviewer
	if callerRole == "recruiter" && req.Role != "interviewer" && req.Role != "hiring_manager" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "recruiters can only invite hiring managers and interviewers"})
		return
	}

	// Admins can only invite recruiter, hiring_manager, or interviewer
	if callerRole == "admin" && (req.Role == "super_admin" || req.Role == "admin") {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "admins cannot invite admin or super admin"})
		return
	}

	invitation, err := s.Queries.CreateInvitation(ctx, db.CreateInvitationParams{
		Email:          req.Email,
		Role:           req.Role,
		OrganizationID: orgID,
		ExpiresAt:      pgtype.Timestamptz{Time: time.Now().Add(7 * 24 * time.Hour), Valid: true},
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create invitation"})
		return
	}

	writeJSON(w, http.StatusCreated, invitation)
}

// UpdateTeamMember updates a user's role.
func (s *Server) UpdateTeamMember(w http.ResponseWriter, r *http.Request) {
	var req updateMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if !validRoles[req.Role] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid role"})
		return
	}

	ctx := r.Context()
	orgID := mw.GetOrgID(ctx)
	callerRole := mw.GetRole(ctx)
	callerID := mw.GetUserID(ctx)
	targetID := chi.URLParam(r, "userId")

	// Prevent self-demotion
	if targetID == callerID {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot change your own role"})
		return
	}

	// Admins cannot promote to super_admin or admin
	if callerRole == "admin" && (req.Role == "super_admin" || req.Role == "admin") {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "admins cannot promote to admin or super admin"})
		return
	}

	// Admins cannot modify super_admins or other admins
	if callerRole == "admin" {
		target, err := s.Queries.GetUserByID(ctx, targetID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
			return
		}
		if target.Role == "super_admin" || target.Role == "admin" {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "admins cannot modify admin or super admin users"})
			return
		}
	}

	updated, err := s.Queries.UpdateUserRole(ctx, db.UpdateUserRoleParams{
		ID:             targetID,
		Role:           req.Role,
		OrganizationID: orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":    updated.ID,
		"email": updated.Email,
		"name":  updated.Name,
		"role":  updated.Role,
	})
}

// RemoveTeamMember removes a user from the org.
func (s *Server) RemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID := mw.GetOrgID(ctx)
	callerRole := mw.GetRole(ctx)
	callerID := mw.GetUserID(ctx)
	targetID := chi.URLParam(r, "userId")

	// Cannot remove self
	if targetID == callerID {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot remove yourself"})
		return
	}

	// Admins cannot remove super_admins or other admins
	if callerRole == "admin" {
		target, err := s.Queries.GetUserByID(ctx, targetID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
			return
		}
		if target.Role == "super_admin" || target.Role == "admin" {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "admins cannot remove admin or super admin users"})
			return
		}
	}

	err := s.Queries.DeleteUser(ctx, db.DeleteUserParams{
		ID:             targetID,
		OrganizationID: orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

// AssignInterviewer assigns an interviewer to an application.
func (s *Server) AssignInterviewer(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "appId")
	ctx := r.Context()
	orgID := mw.GetOrgID(ctx)

	var req struct {
		UserID string `json:"userId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Verify application belongs to this org
	app, err := s.Queries.GetApplicationByID(ctx, appID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "application not found"})
		return
	}
	_, err = s.Queries.GetJobByID(ctx, db.GetJobByIDParams{ID: app.JobID, OrganizationID: orgID})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "application not found"})
		return
	}

	assignment, err := s.Queries.CreateInterviewAssignment(ctx, db.CreateInterviewAssignmentParams{
		ApplicationID: appID,
		InterviewerID: req.UserID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to assign interviewer"})
		return
	}

	writeJSON(w, http.StatusCreated, assignment)
}

// RemoveInterviewer removes an interviewer assignment.
func (s *Server) RemoveInterviewer(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "appId")
	userID := chi.URLParam(r, "userId")
	ctx := r.Context()
	orgID := mw.GetOrgID(ctx)

	// Verify application belongs to this org
	app, err := s.Queries.GetApplicationByID(ctx, appID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "application not found"})
		return
	}
	_, err = s.Queries.GetJobByID(ctx, db.GetJobByIDParams{ID: app.JobID, OrganizationID: orgID})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "application not found"})
		return
	}

	err = s.Queries.DeleteInterviewAssignment(ctx, db.DeleteInterviewAssignmentParams{
		ApplicationID: appID,
		InterviewerID: userID,
	})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "assignment not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

// ListInterviewers returns interviewers assigned to an application.
func (s *Server) ListInterviewers(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "appId")
	ctx := r.Context()

	interviewers, err := s.Queries.ListInterviewersByApplication(ctx, appID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list interviewers"})
		return
	}

	writeJSON(w, http.StatusOK, interviewers)
}

// MyInterviews returns interview assignments for the current user.
func (s *Server) MyInterviews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := mw.GetUserID(ctx)
	orgID := mw.GetOrgID(ctx)

	apps, err := s.Queries.ListApplicationsByInterviewer(ctx, db.ListApplicationsByInterviewerParams{
		InterviewerID:  userID,
		OrganizationID: orgID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list interviews"})
		return
	}

	writeJSON(w, http.StatusOK, apps)
}
