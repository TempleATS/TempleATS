package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/temple-ats/TempleATS/internal/auth"
	"github.com/temple-ats/TempleATS/internal/calendar"
	"github.com/temple-ats/TempleATS/internal/db"
	"github.com/temple-ats/TempleATS/internal/email"
	mw "github.com/temple-ats/TempleATS/internal/middleware"
	"golang.org/x/oauth2"
)

func getGoogleConfig() *oauth2.Config {
	return calendar.GoogleOAuthConfig(
		os.Getenv("GOOGLE_CLIENT_ID"),
		os.Getenv("GOOGLE_CLIENT_SECRET"),
		os.Getenv("GOOGLE_REDIRECT_URI"),
	)
}

func generateToken() string {
	b := make([]byte, 24)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GoogleAuthURL returns the Google OAuth URL for calendar connection.
func (s *Server) GoogleAuthURL(w http.ResponseWriter, r *http.Request) {
	config := getGoogleConfig()
	if config.ClientID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Google Calendar not configured"})
		return
	}

	state := generateToken()
	url := config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	// Store state in a short-lived cookie for CSRF protection
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}

// GoogleAuthCallback handles the Google OAuth callback, stores tokens, and redirects.
func (s *Server) GoogleAuthCallback(w http.ResponseWriter, r *http.Request) {
	// Validate state
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	// Clear state cookie
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", Value: "", Path: "/", MaxAge: -1})

	// Get user from JWT cookie (they must be logged in)
	tokenCookie, err := r.Cookie("token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	claims, err := auth.ValidateToken(tokenCookie.Value)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	config := getGoogleConfig()
	token, err := calendar.ExchangeCode(r.Context(), config, code)
	if err != nil {
		log.Printf("[google-oauth] exchange error: %v", err)
		http.Redirect(w, r, "/account?error=oauth_failed", http.StatusFound)
		return
	}

	// Get the user's Google email
	calEmail, err := calendar.GetUserEmail(r.Context(), config, token)
	if err != nil {
		log.Printf("[google-oauth] get email error: %v", err)
		calEmail = "unknown"
	}

	// Store the connection
	_, err = s.Queries.UpsertCalendarConnection(r.Context(), db.UpsertCalendarConnectionParams{
		UserID:       claims.UserID,
		Provider:     "google",
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenExpiry:  pgtype.Timestamptz{Time: token.Expiry, Valid: true},
		CalendarEmail: calEmail,
	})
	if err != nil {
		log.Printf("[google-oauth] save connection error: %v", err)
		http.Redirect(w, r, "/account?error=save_failed", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/account?calendar=connected", http.StatusFound)
}

// GetCalendarConnectionHandler returns the user's calendar connection status.
func (s *Server) GetCalendarConnectionHandler(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserID(r.Context())
	conn, err := s.Queries.GetCalendarConnectionByUser(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusOK, nil)
		return
	}
	writeJSON(w, http.StatusOK, conn)
}

// DisconnectCalendar removes the user's calendar connection.
func (s *Server) DisconnectCalendar(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserID(r.Context())
	err := s.Queries.DeleteCalendarConnection(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to disconnect"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "disconnected"})
}

type checkAvailabilityRequest struct {
	InterviewerIDs []string `json:"interviewerIds"`
	StartDate      string   `json:"startDate"` // RFC3339
	EndDate        string   `json:"endDate"`   // RFC3339
}

type availableBlock struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// CheckAvailability queries Google Calendar freebusy for interviewers and returns available blocks.
func (s *Server) CheckAvailability(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req checkAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid startDate"})
		return
	}
	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid endDate"})
		return
	}

	config := getGoogleConfig()

	// Collect busy blocks from all interviewers
	var allBusy []calendar.BusyBlock
	for _, uid := range req.InterviewerIDs {
		conn, err := s.Queries.GetCalendarConnection(ctx, db.GetCalendarConnectionParams{
			UserID: uid, Provider: "google",
		})
		if err != nil {
			continue // interviewer not connected, skip
		}

		expiry := conn.TokenExpiry.Time
		busy, newTok, err := calendar.GetFreeBusy(ctx, config, conn.AccessToken, conn.RefreshToken, expiry, conn.CalendarEmail, startDate, endDate)
		if err != nil {
			log.Printf("[availability] freebusy error for user %s: %v", uid, err)
			continue
		}

		// If token was refreshed, update it
		if newTok != nil && newTok.AccessToken != conn.AccessToken {
			s.Queries.UpsertCalendarConnection(ctx, db.UpsertCalendarConnectionParams{
				UserID:        uid,
				Provider:      "google",
				AccessToken:   newTok.AccessToken,
				RefreshToken:  newTok.RefreshToken,
				TokenExpiry:   pgtype.Timestamptz{Time: newTok.Expiry, Valid: true},
				CalendarEmail: conn.CalendarEmail,
			})
		}

		allBusy = append(allBusy, busy...)
	}

	// Compute available blocks: invert busy blocks within [startDate, endDate]
	// during business hours (9 AM - 6 PM in UTC, adjustable later)
	available := computeAvailableBlocks(startDate, endDate, allBusy, 30*time.Minute)

	writeJSON(w, http.StatusOK, available)
}

// computeAvailableBlocks returns free time blocks within business hours (9am-6pm each day)
// that don't overlap with any busy blocks.
func computeAvailableBlocks(start, end time.Time, busy []calendar.BusyBlock, slotDuration time.Duration) []availableBlock {
	// Sort busy blocks by start time
	sort.Slice(busy, func(i, j int) bool { return busy[i].Start.Before(busy[j].Start) })

	// Merge overlapping busy blocks
	var merged []calendar.BusyBlock
	for _, b := range busy {
		if len(merged) > 0 && !b.Start.After(merged[len(merged)-1].End) {
			if b.End.After(merged[len(merged)-1].End) {
				merged[len(merged)-1].End = b.End
			}
		} else {
			merged = append(merged, b)
		}
	}

	var available []availableBlock

	// Iterate each day
	for day := start; day.Before(end); day = day.Add(24 * time.Hour) {
		dayStart := time.Date(day.Year(), day.Month(), day.Day(), 9, 0, 0, 0, day.Location())
		dayEnd := time.Date(day.Year(), day.Month(), day.Day(), 18, 0, 0, 0, day.Location())

		if dayStart.Before(start) {
			dayStart = start
		}
		if dayEnd.After(end) {
			dayEnd = end
		}

		// Skip weekends
		if dayStart.Weekday() == time.Saturday || dayStart.Weekday() == time.Sunday {
			continue
		}

		// Find free slots within this day by subtracting busy blocks
		cursor := dayStart
		for _, b := range merged {
			if b.End.Before(cursor) || b.End.Equal(cursor) {
				continue
			}
			if b.Start.After(dayEnd) || b.Start.Equal(dayEnd) {
				break
			}

			// Free time before this busy block
			if b.Start.After(cursor) {
				freeEnd := b.Start
				if freeEnd.After(dayEnd) {
					freeEnd = dayEnd
				}
				if freeEnd.Sub(cursor) >= slotDuration {
					available = append(available, availableBlock{Start: cursor, End: freeEnd})
				}
			}

			if b.End.After(cursor) {
				cursor = b.End
			}
		}

		// Free time after all busy blocks
		if cursor.Before(dayEnd) && dayEnd.Sub(cursor) >= slotDuration {
			available = append(available, availableBlock{Start: cursor, End: dayEnd})
		}
	}

	return available
}

type createScheduleRequest struct {
	Slots           []slotInput `json:"slots"`
	DurationMinutes int32       `json:"durationMinutes"`
	Location        string      `json:"location"`
	MeetingUrl      string      `json:"meetingUrl"`
	Notes           string      `json:"notes"`
	InterviewerIDs  []string    `json:"interviewerIds"`
}

type slotInput struct {
	Start string `json:"start"` // RFC3339
	End   string `json:"end"`   // RFC3339
}

// CreateSchedule creates a new interview schedule with proposed slots and emails the candidate.
func (s *Server) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "appId")
	orgID := mw.GetOrgID(r.Context())
	userID := mw.GetUserID(r.Context())
	ctx := r.Context()

	var req createScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if len(req.Slots) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "at least one slot is required"})
		return
	}
	if req.DurationMinutes <= 0 {
		req.DurationMinutes = 60
	}

	// Verify application belongs to this org
	appDetails, err := s.Queries.GetApplicationWithDetails(ctx, appID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "application not found"})
		return
	}
	_, err = s.Queries.GetJobByID(ctx, db.GetJobByIDParams{ID: appDetails.JobID, OrganizationID: orgID})
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "application not found"})
		return
	}

	token := generateToken()
	schedule, err := s.Queries.CreateInterviewSchedule(ctx, db.CreateInterviewScheduleParams{
		ApplicationID:   appID,
		Token:           token,
		DurationMinutes: req.DurationMinutes,
		Location:        pgtype.Text{String: req.Location, Valid: req.Location != ""},
		MeetingUrl:      pgtype.Text{String: req.MeetingUrl, Valid: req.MeetingUrl != ""},
		Notes:           pgtype.Text{String: req.Notes, Valid: req.Notes != ""},
		CreatedBy:       userID,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create schedule"})
		return
	}

	// Create slots
	for _, slot := range req.Slots {
		startTime, _ := time.Parse(time.RFC3339, slot.Start)
		endTime, _ := time.Parse(time.RFC3339, slot.End)
		s.Queries.CreateInterviewSlot(ctx, db.CreateInterviewSlotParams{
			ScheduleID: schedule.ID,
			StartTime:  pgtype.Timestamptz{Time: startTime, Valid: true},
			EndTime:    pgtype.Timestamptz{Time: endTime, Valid: true},
		})
	}

	// Add interviewers
	for _, uid := range req.InterviewerIDs {
		s.Queries.AddScheduleInterviewer(ctx, schedule.ID, uid)
	}

	// Send booking email to candidate
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	bookingLink := fmt.Sprintf("%s/schedule/%s", baseURL, token)

	subject := fmt.Sprintf("Schedule your interview for %s at %s", appDetails.JobTitle, appDetails.OrgName)
	body := fmt.Sprintf(
		`<p>Hi %s,</p>
<p>We'd like to schedule an interview with you for the <strong>%s</strong> position at <strong>%s</strong>.</p>
<p>Please click the link below to choose a time that works for you:</p>
<p><a href="%s" style="display:inline-block;padding:12px 24px;background:#4f46e5;color:#fff;text-decoration:none;border-radius:6px;">Choose Interview Time</a></p>
<p>If you have any questions, please don't hesitate to reach out.</p>
<p>Best regards,<br/>The %s Team</p>`,
		appDetails.CandidateName, appDetails.JobTitle, appDetails.OrgName, bookingLink, appDetails.OrgName,
	)

	s.Email.SendAsync(orgID, email.SendParams{
		To:            appDetails.CandidateEmail,
		RecipientName: appDetails.CandidateName,
		Subject:       subject,
		Body:          body,
		Type:          "interview_schedule",
		ApplicationID: appID,
		TriggeredByID: userID,
	})

	// Return schedule with slots
	slots, _ := s.Queries.ListSlotsBySchedule(ctx, schedule.ID)
	interviewers, _ := s.Queries.ListScheduleInterviewers(ctx, schedule.ID)

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"schedule":     schedule,
		"slots":        slots,
		"interviewers": interviewers,
		"bookingLink":  bookingLink,
	})
}

// ListSchedules returns all schedules for an application.
func (s *Server) ListSchedules(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "appId")
	ctx := r.Context()

	schedules, err := s.Queries.ListSchedulesByApplication(ctx, appID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list schedules"})
		return
	}

	// Enrich each schedule with slots
	type enrichedSchedule struct {
		db.ListSchedulesByApplicationRow
		Slots        []db.InterviewSlot               `json:"slots"`
		Interviewers []db.ListScheduleInterviewersRow `json:"interviewers"`
	}

	result := make([]enrichedSchedule, 0)
	for _, sch := range schedules {
		slots, _ := s.Queries.ListSlotsBySchedule(ctx, sch.ID)
		interviewers, _ := s.Queries.ListScheduleInterviewers(ctx, sch.ID)
		result = append(result, enrichedSchedule{
			ListSchedulesByApplicationRow: sch,
			Slots:                         slots,
			Interviewers:                  interviewers,
		})
	}

	writeJSON(w, http.StatusOK, result)
}

// GetPublicSchedule returns schedule data for the public booking page.
func (s *Server) GetPublicSchedule(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	ctx := r.Context()

	schedule, err := s.Queries.GetScheduleByToken(ctx, token)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "schedule not found"})
		return
	}

	slots, _ := s.Queries.ListSlotsBySchedule(ctx, schedule.ID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":               schedule.ID,
		"status":           schedule.Status,
		"duration_minutes": schedule.DurationMinutes,
		"location":         schedule.Location,
		"meeting_url":      schedule.MeetingUrl,
		"notes":            schedule.Notes,
		"created_at":       schedule.CreatedAt,
		"confirmed_at":     schedule.ConfirmedAt,
		"job_title":        schedule.JobTitle,
		"candidate_name":   schedule.CandidateName,
		"org_name":         schedule.OrgName,
		"slots":            slots,
	})
}

type confirmScheduleRequest struct {
	SlotID string `json:"slotId"`
}

// ConfirmSchedule allows a candidate to confirm a time slot via the public booking link.
func (s *Server) ConfirmSchedule(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	ctx := r.Context()

	var req confirmScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if req.SlotID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "slotId is required"})
		return
	}

	schedule, err := s.Queries.GetScheduleByToken(ctx, token)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "schedule not found"})
		return
	}
	if schedule.Status != "pending" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "schedule is already " + schedule.Status})
		return
	}

	// Mark the selected slot
	if err := s.Queries.SelectSlot(ctx, schedule.ID, req.SlotID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to select slot"})
		return
	}

	// Update schedule status
	if err := s.Queries.UpdateScheduleStatus(ctx, schedule.ID, "confirmed"); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to confirm schedule"})
		return
	}

	// Get selected slot details
	slots, _ := s.Queries.ListSlotsBySchedule(ctx, schedule.ID)
	var selectedSlot db.InterviewSlot
	for _, sl := range slots {
		if sl.ID == req.SlotID {
			selectedSlot = sl
			break
		}
	}

	// Create calendar events for interviewers (in background)
	go func() {
		bgCtx := context.Background()
		config := getGoogleConfig()
		interviewers, _ := s.Queries.ListScheduleInterviewers(bgCtx, schedule.ID)
		summary := fmt.Sprintf("Interview: %s - %s", schedule.CandidateName, schedule.JobTitle)
		description := fmt.Sprintf("Interview with %s for %s position at %s", schedule.CandidateName, schedule.JobTitle, schedule.OrgName)

		for _, interviewer := range interviewers {
			conn, err := s.Queries.GetCalendarConnection(bgCtx, db.GetCalendarConnectionParams{
				UserID: interviewer.UserID, Provider: "google",
			})
			if err != nil {
				continue
			}

			eventID, newTok, err := calendar.CreateEvent(
				bgCtx, config, conn.AccessToken, conn.RefreshToken, conn.TokenExpiry.Time,
				summary, description,
				selectedSlot.StartTime.Time, selectedSlot.EndTime.Time,
				[]string{schedule.CandidateEmail},
			)
			if err != nil {
				log.Printf("[calendar] failed to create event for user %s: %v", interviewer.UserID, err)
				continue
			}

			// Save event ID
			s.Queries.SetCalendarEventID(bgCtx, schedule.ID, interviewer.UserID, pgtype.Text{String: eventID, Valid: true})

			// Update token if refreshed
			if newTok != nil && newTok.AccessToken != conn.AccessToken {
				s.Queries.UpsertCalendarConnection(bgCtx, db.UpsertCalendarConnectionParams{
					UserID:        interviewer.UserID,
					Provider:      "google",
					AccessToken:   newTok.AccessToken,
					RefreshToken:  newTok.RefreshToken,
					TokenExpiry:   pgtype.Timestamptz{Time: newTok.Expiry, Valid: true},
					CalendarEmail: conn.CalendarEmail,
				})
			}
		}
	}()

	// Get orgID from the schedule creator for sending emails
	creator, err := s.Queries.GetUserByID(ctx, schedule.CreatedBy)
	if err != nil {
		log.Printf("[confirm] failed to get creator: %v", err)
	}
	orgID := ""
	if err == nil {
		orgID = creator.OrganizationID
	}

	// Send confirmation emails
	go func() {
		if orgID == "" {
			return
		}
		bgCtx := context.Background()

		startFormatted := selectedSlot.StartTime.Time.Format("Monday, January 2, 2006 at 3:04 PM MST")

		// Email to candidate
		s.Email.SendAsync(orgID, email.SendParams{
			To:            schedule.CandidateEmail,
			RecipientName: schedule.CandidateName,
			Subject:       fmt.Sprintf("Interview Confirmed: %s at %s", schedule.JobTitle, schedule.OrgName),
			Body: fmt.Sprintf(
				`<p>Hi %s,</p><p>Your interview for <strong>%s</strong> at <strong>%s</strong> has been confirmed for:</p><p style="font-size:18px;font-weight:bold;">%s</p><p>We look forward to meeting you!</p>`,
				schedule.CandidateName, schedule.JobTitle, schedule.OrgName, startFormatted,
			),
			Type:          "interview_confirmed",
			ApplicationID: schedule.ApplicationID,
		})

		// Email to interviewers
		interviewers, _ := s.Queries.ListScheduleInterviewers(bgCtx, schedule.ID)
		for _, interviewer := range interviewers {
			user, err := s.Queries.GetUserByID(bgCtx, interviewer.UserID)
			if err != nil {
				continue
			}
			s.Email.SendAsync(orgID, email.SendParams{
				To:            user.Email,
				RecipientName: user.Name,
				Subject:       fmt.Sprintf("Interview Confirmed: %s - %s", schedule.CandidateName, schedule.JobTitle),
				Body: fmt.Sprintf(
					`<p>Hi %s,</p><p>An interview has been confirmed with <strong>%s</strong> for the <strong>%s</strong> position:</p><p style="font-size:18px;font-weight:bold;">%s</p><p>A calendar event has been created.</p>`,
					user.Name, schedule.CandidateName, schedule.JobTitle, startFormatted,
				),
				Type:          "interview_confirmed",
				ApplicationID: schedule.ApplicationID,
			})
		}
	}()

	writeJSON(w, http.StatusOK, map[string]string{"status": "confirmed"})
}
