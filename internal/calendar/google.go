package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gocal "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func decodeJSON(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

// BusyBlock represents a time range when someone is busy.
type BusyBlock struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// GoogleOAuthConfig builds an OAuth2 config for Google Calendar.
func GoogleOAuthConfig(clientID, clientSecret, redirectURI string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar.freebusy",
			"https://www.googleapis.com/auth/calendar.events",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}
}

// ExchangeCode exchanges an authorization code for a token.
func ExchangeCode(ctx context.Context, config *oauth2.Config, code string) (*oauth2.Token, error) {
	return config.Exchange(ctx, code)
}

// tokenSource creates a token source that auto-refreshes from stored tokens.
func tokenSource(config *oauth2.Config, accessToken, refreshToken string, expiry time.Time) oauth2.TokenSource {
	tok := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       expiry,
		TokenType:    "Bearer",
	}
	return config.TokenSource(context.Background(), tok)
}

// GetFreeBusy queries Google Calendar for busy blocks in a time range.
func GetFreeBusy(ctx context.Context, config *oauth2.Config, accessToken, refreshToken string, expiry time.Time, calendarEmail string, timeMin, timeMax time.Time) ([]BusyBlock, *oauth2.Token, error) {
	ts := tokenSource(config, accessToken, refreshToken, expiry)
	// Get potentially refreshed token
	newTok, err := ts.Token()
	if err != nil {
		return nil, nil, fmt.Errorf("token refresh: %w", err)
	}

	svc, err := gocal.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, newTok, fmt.Errorf("create calendar service: %w", err)
	}

	req := &gocal.FreeBusyRequest{
		TimeMin: timeMin.Format(time.RFC3339),
		TimeMax: timeMax.Format(time.RFC3339),
		Items: []*gocal.FreeBusyRequestItem{
			{Id: calendarEmail},
		},
	}

	resp, err := svc.Freebusy.Query(req).Context(ctx).Do()
	if err != nil {
		return nil, newTok, fmt.Errorf("freebusy query: %w", err)
	}

	var blocks []BusyBlock
	if cal, ok := resp.Calendars[calendarEmail]; ok {
		for _, busy := range cal.Busy {
			start, _ := time.Parse(time.RFC3339, busy.Start)
			end, _ := time.Parse(time.RFC3339, busy.End)
			blocks = append(blocks, BusyBlock{Start: start, End: end})
		}
	}

	return blocks, newTok, nil
}

// CreateEvent creates a Google Calendar event and returns the event ID.
func CreateEvent(ctx context.Context, config *oauth2.Config, accessToken, refreshToken string, expiry time.Time, summary, description string, start, end time.Time, attendees []string) (string, *oauth2.Token, error) {
	ts := tokenSource(config, accessToken, refreshToken, expiry)
	newTok, err := ts.Token()
	if err != nil {
		return "", nil, fmt.Errorf("token refresh: %w", err)
	}

	svc, err := gocal.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return "", newTok, fmt.Errorf("create calendar service: %w", err)
	}

	event := &gocal.Event{
		Summary:     summary,
		Description: description,
		Start: &gocal.EventDateTime{
			DateTime: start.Format(time.RFC3339),
		},
		End: &gocal.EventDateTime{
			DateTime: end.Format(time.RFC3339),
		},
	}

	for _, email := range attendees {
		event.Attendees = append(event.Attendees, &gocal.EventAttendee{Email: email})
	}

	created, err := svc.Events.Insert("primary", event).Context(ctx).Do()
	if err != nil {
		return "", newTok, fmt.Errorf("create event: %w", err)
	}

	return created.Id, newTok, nil
}

// GetUserEmail retrieves the email associated with the Google account using the People API userinfo.
func GetUserEmail(ctx context.Context, config *oauth2.Config, token *oauth2.Token) (string, error) {
	client := config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return "", fmt.Errorf("get userinfo: %w", err)
	}
	defer resp.Body.Close()

	// Parse JSON response
	var info struct {
		Email string `json:"email"`
	}
	if err := decodeJSON(resp.Body, &info); err != nil {
		return "", fmt.Errorf("decode userinfo: %w", err)
	}
	return info.Email, nil
}
