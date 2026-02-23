package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/temple-ats/TempleATS/internal/handler"
)

func TestHealth(t *testing.T) {
	srv := handler.NewServer(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	srv.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	expected := `{"status":"ok"}`
	if w.Body.String() != expected {
		t.Errorf("expected body %q, got %q", expected, w.Body.String())
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}
