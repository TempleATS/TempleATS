package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"strings"
	"text/template"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/temple-ats/TempleATS/internal/db"
)

// Service handles email sending for organizations.
type Service struct {
	Queries *db.Queries
}

// NewService creates a new email service.
func NewService(queries *db.Queries) *Service {
	return &Service{Queries: queries}
}

// Config holds resolved SMTP settings.
type Config struct {
	Host      string
	Port      int32
	Username  string
	Password  string
	FromEmail string
	FromName  string
	TLS       bool
}

// GetConfig retrieves SMTP config for an org. Returns nil if not configured.
func (s *Service) GetConfig(ctx context.Context, orgID string) *Config {
	settings, err := s.Queries.GetSmtpSettings(ctx, orgID)
	if err != nil {
		return nil
	}
	return &Config{
		Host:      settings.Host,
		Port:      settings.Port,
		Username:  settings.Username,
		Password:  settings.Password,
		FromEmail: settings.FromEmail,
		FromName:  settings.FromName,
		TLS:       settings.TlsEnabled,
	}
}

// SendEmail sends a single email using the provided SMTP config.
func SendEmail(config *Config, to, subject, body string) error {
	from := config.FromEmail
	if config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", config.FromName, config.FromEmail)
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		from, to, subject, body)

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)

	if config.TLS {
		conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
		if err != nil {
			return fmt.Errorf("dial: %w", err)
		}
		c, err := smtp.NewClient(conn, config.Host)
		if err != nil {
			conn.Close()
			return fmt.Errorf("smtp client: %w", err)
		}
		defer c.Close()

		tlsConfig := &tls.Config{ServerName: config.Host}
		if err = c.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
		if err = c.Auth(auth); err != nil {
			return fmt.Errorf("auth: %w", err)
		}
		if err = c.Mail(config.FromEmail); err != nil {
			return fmt.Errorf("mail from: %w", err)
		}
		if err = c.Rcpt(to); err != nil {
			return fmt.Errorf("rcpt to: %w", err)
		}
		w, err := c.Data()
		if err != nil {
			return fmt.Errorf("data: %w", err)
		}
		_, err = w.Write([]byte(msg))
		if err != nil {
			return fmt.Errorf("write: %w", err)
		}
		if err = w.Close(); err != nil {
			return fmt.Errorf("close: %w", err)
		}
		return c.Quit()
	}

	return smtp.SendMail(addr, auth, config.FromEmail, []string{to}, []byte(msg))
}

// Attachment represents an email file attachment.
type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

// SendEmailWithAttachment sends an email with file attachments using multipart MIME.
func SendEmailWithAttachment(config *Config, to, subject, body string, attachments []Attachment) error {
	from := config.FromEmail
	if config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", config.FromName, config.FromEmail)
	}

	boundary := fmt.Sprintf("==BOUNDARY_%d==", time.Now().UnixNano())

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", to))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n", boundary))

	// HTML body part
	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
	buf.WriteString(body)
	buf.WriteString("\r\n")

	// Attachment parts
	for _, att := range attachments {
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		buf.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n", att.ContentType, att.Filename))
		buf.WriteString("Content-Transfer-Encoding: base64\r\n")
		buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n\r\n", att.Filename))
		encoded := base64.StdEncoding.EncodeToString(att.Data)
		// Write in 76-char lines per MIME spec
		for i := 0; i < len(encoded); i += 76 {
			end := i + 76
			if end > len(encoded) {
				end = len(encoded)
			}
			buf.WriteString(encoded[i:end])
			buf.WriteString("\r\n")
		}
	}
	buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	msg := buf.Bytes()
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)

	if config.TLS {
		conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
		if err != nil {
			return fmt.Errorf("dial: %w", err)
		}
		c, err := smtp.NewClient(conn, config.Host)
		if err != nil {
			conn.Close()
			return fmt.Errorf("smtp client: %w", err)
		}
		defer c.Close()

		tlsConfig := &tls.Config{ServerName: config.Host}
		if err = c.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
		if err = c.Auth(auth); err != nil {
			return fmt.Errorf("auth: %w", err)
		}
		if err = c.Mail(config.FromEmail); err != nil {
			return fmt.Errorf("mail from: %w", err)
		}
		if err = c.Rcpt(to); err != nil {
			return fmt.Errorf("rcpt to: %w", err)
		}
		w, err := c.Data()
		if err != nil {
			return fmt.Errorf("data: %w", err)
		}
		if _, err = w.Write(msg); err != nil {
			return fmt.Errorf("write: %w", err)
		}
		if err = w.Close(); err != nil {
			return fmt.Errorf("close: %w", err)
		}
		return c.Quit()
	}

	return smtp.SendMail(addr, auth, config.FromEmail, []string{to}, msg)
}

// SendParams holds parameters for async email sending.
type SendParams struct {
	To            string
	RecipientName string
	Subject       string
	Body          string
	Type          string // "mention", "stage_change", "freeform"
	ApplicationID string
	NoteID        string
	TriggeredByID string
}

// SendAsync sends email in a goroutine and logs a notification to DB.
func (s *Service) SendAsync(orgID string, params SendParams) {
	go func() {
		ctx := context.Background()
		config := s.GetConfig(ctx, orgID)
		if config == nil {
			log.Printf("[email] SMTP not configured for org %s, skipping", orgID)
			return
		}

		status := "sent"
		var errMsg string
		if err := SendEmail(config, params.To, params.Subject, params.Body); err != nil {
			status = "failed"
			errMsg = err.Error()
			log.Printf("[email] failed to send to %s: %v", params.To, err)
		} else {
			log.Printf("[email] sent %s email to %s", params.Type, params.To)
		}

		s.Queries.CreateNotification(ctx, db.CreateNotificationParams{
			OrganizationID: orgID,
			Type:           params.Type,
			RecipientEmail: params.To,
			RecipientName:  params.RecipientName,
			Subject:        params.Subject,
			Body:           params.Body,
			Status:         status,
			ErrorMessage:   pgText(errMsg),
			ApplicationID:  pgText(params.ApplicationID),
			NoteID:         pgText(params.NoteID),
			TriggeredByID:  pgText(params.TriggeredByID),
		})
	}()
}

// RenderTemplate renders an email template with Go template variables.
func RenderTemplate(tmpl string, data map[string]string) (string, error) {
	t, err := template.New("email").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ParseMentions extracts @mentioned users from note content by matching against team members.
func ParseMentions(content string, teamMembers []db.User) []db.User {
	lower := strings.ToLower(content)
	var mentioned []db.User
	for _, member := range teamMembers {
		if strings.Contains(lower, "@"+strings.ToLower(member.Name)) {
			mentioned = append(mentioned, member)
		}
	}
	return mentioned
}

// pgText converts a string to pgtype.Text, returning invalid for empty strings.
func pgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}
