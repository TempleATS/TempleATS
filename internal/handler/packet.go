package handler

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-pdf/fpdf"
	mw "github.com/temple-ats/TempleATS/internal/middleware"
	"github.com/temple-ats/TempleATS/internal/email"
)

var recLabels = map[string]string{
	"1": "Strong No",
	"2": "No",
	"3": "Yes",
	"4": "Strong Yes",
}

// GenerateHiringPacket creates a PDF with final round feedback and emails it to the requesting user.
func (s *Server) GenerateHiringPacket(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "appId")
	orgID := mw.GetOrgID(r.Context())
	userID := mw.GetUserID(r.Context())
	ctx := r.Context()

	// Get application details
	app, err := s.Queries.GetApplicationWithDetails(ctx, appID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "application not found"})
		return
	}

	// Verify org ownership
	if app.OrgName == "" {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "application not found"})
		return
	}

	// Get requesting user's info for email
	requestingUser, err := s.Queries.GetUserByID(ctx, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not find user"})
		return
	}

	// Get all feedback, filter to final_interview
	allFeedback, err := s.Queries.ListFeedbackByApplication(ctx, appID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load feedback"})
		return
	}

	var feedback []struct {
		AuthorName    string
		InterviewType string
		Recommendation string
		Content       string
		CreatedAt     time.Time
	}
	for _, fb := range allFeedback {
		if fb.Stage != "final_interview" {
			continue
		}
		var ivType string
		if fb.InterviewType.Valid {
			ivType = fb.InterviewType.String
		}
		var createdAt time.Time
		if fb.CreatedAt.Valid {
			createdAt = fb.CreatedAt.Time
		}
		feedback = append(feedback, struct {
			AuthorName    string
			InterviewType string
			Recommendation string
			Content       string
			CreatedAt     time.Time
		}{
			AuthorName:     fb.AuthorName,
			InterviewType:  ivType,
			Recommendation: fb.Recommendation,
			Content:        fb.Content,
			CreatedAt:      createdAt,
		})
	}

	if len(feedback) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no final round feedback to include"})
		return
	}

	// Generate PDF
	dept := ""
	if app.JobDepartment.Valid {
		dept = app.JobDepartment.String
	}
	loc := ""
	if app.JobLocation.Valid {
		loc = app.JobLocation.String
	}
	pdfBytes, err := generatePacketPDF(app.OrgName, app.CandidateName, app.JobTitle, dept, loc, feedback)
	if err != nil {
		log.Printf("[packet] PDF generation failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate PDF"})
		return
	}

	// Send email with attachment
	go func() {
		bgCtx := context.Background()
		config := s.Email.GetConfig(bgCtx, orgID)
		if config == nil {
			log.Printf("[packet] SMTP not configured for org %s", orgID)
			return
		}

		filename := fmt.Sprintf("Hiring_Packet_%s_%s.pdf",
			strings.ReplaceAll(app.CandidateName, " ", "_"),
			time.Now().Format("2006-01-02"))

		subject := fmt.Sprintf("Hiring Review Packet: %s - %s", app.CandidateName, app.JobTitle)
		body := fmt.Sprintf(`<p>Hi %s,</p>
<p>Attached is the hiring review packet for <strong>%s</strong> for the <strong>%s</strong> position.</p>
<p>This packet contains %d final round interview feedback entries.</p>
<p>— %s</p>`,
			requestingUser.Name, app.CandidateName, app.JobTitle, len(feedback), app.OrgName)

		att := email.Attachment{
			Filename:    filename,
			ContentType: "application/pdf",
			Data:        pdfBytes,
		}

		if err := email.SendEmailWithAttachment(config, requestingUser.Email, subject, body, []email.Attachment{att}); err != nil {
			log.Printf("[packet] failed to send packet email to %s: %v", requestingUser.Email, err)
		} else {
			log.Printf("[packet] sent hiring packet for %s to %s", app.CandidateName, requestingUser.Email)
		}
	}()

	writeJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}

func generatePacketPDF(orgName, candidateName, jobTitle, department, location string, feedback []struct {
	AuthorName     string
	InterviewType  string
	Recommendation string
	Content        string
	CreatedAt      time.Time
}) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 20)
	pdf.AddPage()

	// Title
	pdf.SetFont("Helvetica", "B", 18)
	pdf.CellFormat(0, 10, orgName, "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 14)
	pdf.CellFormat(0, 8, "Hiring Review Packet", "", 1, "C", false, 0, "")
	pdf.Ln(6)

	// Candidate info
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(4)

	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(35, 7, "Candidate:", "", 0, "", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(0, 7, candidateName, "", 1, "", false, 0, "")

	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(35, 7, "Position:", "", 0, "", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(0, 7, jobTitle, "", 1, "", false, 0, "")

	if department != "" {
		pdf.SetFont("Helvetica", "B", 11)
		pdf.CellFormat(35, 7, "Department:", "", 0, "", false, 0, "")
		pdf.SetFont("Helvetica", "", 11)
		pdf.CellFormat(0, 7, department, "", 1, "", false, 0, "")
	}

	if location != "" {
		pdf.SetFont("Helvetica", "B", 11)
		pdf.CellFormat(35, 7, "Location:", "", 0, "", false, 0, "")
		pdf.SetFont("Helvetica", "", 11)
		pdf.CellFormat(0, 7, location, "", 1, "", false, 0, "")
	}

	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(35, 7, "Generated:", "", 0, "", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(0, 7, time.Now().Format("January 2, 2006"), "", 1, "", false, 0, "")

	pdf.Ln(4)

	// Summary
	counts := map[string]int{"1": 0, "2": 0, "3": 0, "4": 0}
	for _, fb := range feedback {
		if fb.Recommendation != "" {
			counts[fb.Recommendation]++
		}
	}

	pdf.SetFont("Helvetica", "B", 13)
	pdf.CellFormat(0, 9, "Summary", "", 1, "", false, 0, "")
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(3)

	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(0, 6, fmt.Sprintf("Total Feedback Entries: %d", len(feedback)), "", 1, "", false, 0, "")
	if counts["4"] > 0 {
		pdf.CellFormat(0, 6, fmt.Sprintf("  Strong Yes: %d", counts["4"]), "", 1, "", false, 0, "")
	}
	if counts["3"] > 0 {
		pdf.CellFormat(0, 6, fmt.Sprintf("  Yes: %d", counts["3"]), "", 1, "", false, 0, "")
	}
	if counts["2"] > 0 {
		pdf.CellFormat(0, 6, fmt.Sprintf("  No: %d", counts["2"]), "", 1, "", false, 0, "")
	}
	if counts["1"] > 0 {
		pdf.CellFormat(0, 6, fmt.Sprintf("  Strong No: %d", counts["1"]), "", 1, "", false, 0, "")
	}
	pdf.Ln(4)

	// Feedback entries
	pdf.SetFont("Helvetica", "B", 13)
	pdf.CellFormat(0, 9, "Final Round Feedback", "", 1, "", false, 0, "")
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(4)

	for i, fb := range feedback {
		// Check if we need a new page (at least 40mm needed for a feedback entry)
		if pdf.GetY() > 250 {
			pdf.AddPage()
		}

		// Feedback header
		pdf.SetFillColor(245, 245, 245)
		pdf.SetFont("Helvetica", "B", 11)

		header := fmt.Sprintf("%d. %s", i+1, fb.AuthorName)
		if fb.InterviewType != "" {
			typeLabel := strings.ReplaceAll(fb.InterviewType, "_", " ")
			header += fmt.Sprintf(" (%s)", strings.Title(typeLabel))
		}
		pdf.CellFormat(0, 8, header, "", 1, "", true, 0, "")

		// Recommendation and date
		pdf.SetFont("Helvetica", "", 9)
		meta := ""
		if fb.Recommendation != "" {
			if label, ok := recLabels[fb.Recommendation]; ok {
				meta = fmt.Sprintf("Recommendation: %s", label)
			}
		}
		if !fb.CreatedAt.IsZero() {
			if meta != "" {
				meta += "  |  "
			}
			meta += fb.CreatedAt.Format("Jan 2, 2006")
		}
		if meta != "" {
			pdf.SetTextColor(120, 120, 120)
			pdf.CellFormat(0, 6, meta, "", 1, "", false, 0, "")
			pdf.SetTextColor(0, 0, 0)
		}

		pdf.Ln(2)

		// Feedback content - use MultiCell for word wrapping
		pdf.SetFont("Helvetica", "", 10)
		pdf.MultiCell(0, 5, fb.Content, "", "", false)
		pdf.Ln(6)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

