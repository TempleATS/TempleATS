package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/temple-ats/TempleATS/internal/handler"
	mw "github.com/temple-ats/TempleATS/internal/middleware"
	"github.com/temple-ats/TempleATS/internal/testutil"
)

func setupCandidatesRouter(srv *handler.Server) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/api/auth/signup", srv.Signup)
	r.Post("/api/careers/{orgSlug}/jobs/{jobId}/apply", srv.CareersApply)
	r.Group(func(r chi.Router) {
		r.Use(mw.RequireAuth)
		r.Post("/api/jobs", srv.CreateJob)
		r.Put("/api/jobs/{jobId}", srv.UpdateJob)
		r.Get("/api/candidates", srv.ListCandidates)
		r.Get("/api/candidates/{candidateId}", srv.GetCandidate)
		r.Get("/api/applications/{appId}", srv.GetApplication)
		r.Post("/api/applications/{appId}/notes", srv.AddNote)
	})
	return r
}

func TestListCandidates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupCandidatesRouter(srv)

	cookie, jobID := createOrgAndOpenJob(t, router, "cand@test.com", "candorg", "Engineer")

	// Apply 2 candidates
	applyCandidate(t, router, "candorg", jobID, "Alice", "alice@cand.com")
	applyCandidate(t, router, "candorg", jobID, "Bob", "bob@cand.com")

	// List candidates
	w := httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("GET", "/api/candidates", "", cookie))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var candidates []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&candidates)
	if len(candidates) != 2 {
		t.Errorf("expected 2 candidates, got %d", len(candidates))
	}
}

func TestSearchCandidates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupCandidatesRouter(srv)

	cookie, jobID := createOrgAndOpenJob(t, router, "search@test.com", "searchorg", "Dev")

	applyCandidate(t, router, "searchorg", jobID, "Alice Smith", "alice@search.com")
	applyCandidate(t, router, "searchorg", jobID, "Bob Jones", "bob@search.com")

	// Search by name
	w := httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("GET", "/api/candidates?q=Alice", "", cookie))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var candidates []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&candidates)
	if len(candidates) != 1 {
		t.Errorf("expected 1 candidate, got %d", len(candidates))
	}
}

func TestGetCandidate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupCandidatesRouter(srv)

	cookie, jobID := createOrgAndOpenJob(t, router, "detail@test.com", "detailorg", "Dev")
	applyCandidate(t, router, "detailorg", jobID, "Carol", "carol@test.com")

	// Get candidates to find ID
	w := httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("GET", "/api/candidates", "", cookie))
	var candidates []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&candidates)
	candID := candidates[0]["id"].(string)

	// Get candidate detail
	w = httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("GET", "/api/candidates/"+candID, "", cookie))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	cand := resp["candidate"].(map[string]interface{})
	if cand["name"] != "Carol" {
		t.Errorf("expected name Carol, got %v", cand["name"])
	}

	apps := resp["applications"].([]interface{})
	if len(apps) != 1 {
		t.Errorf("expected 1 application, got %d", len(apps))
	}
}

func TestCandidateOrgIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupCandidatesRouter(srv)

	// Org A
	cookieA, jobA := createOrgAndOpenJob(t, router, "orgA@test.com", "orgaisolation", "Dev")
	applyCandidate(t, router, "orgaisolation", jobA, "Alice", "alice@iso.com")

	// Org B
	cookieB, jobB := createOrgAndOpenJob(t, router, "orgB@test.com", "orgbisolation", "Dev")
	applyCandidate(t, router, "orgbisolation", jobB, "Bob", "bob@iso.com")

	// Org A should only see Alice
	w := httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("GET", "/api/candidates", "", cookieA))
	var candsA []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&candsA)
	if len(candsA) != 1 {
		t.Errorf("org A expected 1 candidate, got %d", len(candsA))
	}

	// Org B should only see Bob
	w = httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("GET", "/api/candidates", "", cookieB))
	var candsB []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&candsB)
	if len(candsB) != 1 {
		t.Errorf("org B expected 1 candidate, got %d", len(candsB))
	}
}

func TestAddNote(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupCandidatesRouter(srv)

	cookie, jobID := createOrgAndOpenJob(t, router, "note@test.com", "noteorg", "Dev")
	appID := applyCandidate(t, router, "noteorg", jobID, "Dave", "dave@note.com")

	// Add a note
	w := httptest.NewRecorder()
	body := `{"content":"Strong candidate, good Go experience"}`
	router.ServeHTTP(w, authedRequest("POST", "/api/applications/"+appID+"/notes",
		body, cookie))
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// Get application - should include note
	w = httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("GET", "/api/applications/"+appID, "", cookie))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	notes := resp["notes"].([]interface{})
	if len(notes) != 1 {
		t.Errorf("expected 1 note, got %d", len(notes))
	}
	note := notes[0].(map[string]interface{})
	if note["content"] != "Strong candidate, good Go experience" {
		t.Errorf("unexpected note content: %v", note["content"])
	}
}

func TestAddNote_EmptyContent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupCandidatesRouter(srv)

	cookie, jobID := createOrgAndOpenJob(t, router, "empty@test.com", "emptyorg", "Dev")
	appID := applyCandidate(t, router, "emptyorg", jobID, "Eve", "eve@empty.com")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("POST", "/api/applications/"+appID+"/notes",
		`{"content":""}`, cookie))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
