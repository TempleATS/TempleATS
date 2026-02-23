package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/temple-ats/TempleATS/internal/handler"
	mw "github.com/temple-ats/TempleATS/internal/middleware"
	"github.com/temple-ats/TempleATS/internal/testutil"
)

func setupCareersRouter(srv *handler.Server) *chi.Mux {
	r := chi.NewRouter()
	// Auth routes for setup
	r.Post("/api/auth/signup", srv.Signup)
	r.Group(func(r chi.Router) {
		r.Use(mw.RequireAuth)
		r.Post("/api/jobs", srv.CreateJob)
		r.Put("/api/jobs/{jobId}", srv.UpdateJob)
	})
	// Public careers routes
	r.Get("/api/careers/{orgSlug}", srv.CareersListJobs)
	r.Get("/api/careers/{orgSlug}/jobs/{jobId}", srv.CareersGetJob)
	r.Post("/api/careers/{orgSlug}/jobs/{jobId}/apply", srv.CareersApply)
	return r
}

// createOrgAndOpenJob signs up, creates a job, publishes it, returns cookie + job ID.
func createOrgAndOpenJob(t *testing.T, router *chi.Mux, email, orgSlug, jobTitle string) (*http.Cookie, string) {
	t.Helper()
	cookie := signupAndGetCookie(t, router, email, orgSlug)

	// Create job
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, authedRequest("POST", "/api/jobs",
		`{"title":"`+jobTitle+`","description":"A great job"}`, cookie))
	if w1.Code != http.StatusCreated {
		t.Fatalf("create job failed: %d %s", w1.Code, w1.Body.String())
	}
	var job map[string]interface{}
	json.NewDecoder(w1.Body).Decode(&job)
	jobID := job["id"].(string)

	// Publish job (set status to open)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, authedRequest("PUT", "/api/jobs/"+jobID,
		`{"title":"`+jobTitle+`","description":"A great job","status":"open"}`, cookie))
	if w2.Code != http.StatusOK {
		t.Fatalf("publish job failed: %d %s", w2.Code, w2.Body.String())
	}

	return cookie, jobID
}

func TestCareersListJobs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupCareersRouter(srv)
	cookie, _ := createOrgAndOpenJob(t, router, "careers@test.com", "careerorg", "Go Developer")

	// Create a draft job (should not appear in public listing)
	router.ServeHTTP(httptest.NewRecorder(), authedRequest("POST", "/api/jobs",
		`{"title":"Draft Job","description":"Not visible"}`, cookie))

	// Public careers listing
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/careers/careerorg", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	org := resp["organization"].(map[string]interface{})
	if org["name"] != "TestOrg" {
		t.Errorf("expected org name TestOrg, got %v", org["name"])
	}

	jobs := resp["jobs"].([]interface{})
	if len(jobs) != 1 {
		t.Errorf("expected 1 open job (not draft), got %d", len(jobs))
	}
}

func TestCareersListJobs_NonexistentOrg(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupCareersRouter(srv)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/careers/nonexistent", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCareersGetJob(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupCareersRouter(srv)
	_, jobID := createOrgAndOpenJob(t, router, "getjob@test.com", "getjoborg", "React Developer")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/careers/getjoborg/jobs/"+jobID, nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var job map[string]interface{}
	json.NewDecoder(w.Body).Decode(&job)
	if job["title"] != "React Developer" {
		t.Errorf("expected title React Developer, got %v", job["title"])
	}
}

func TestCareersApply_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupCareersRouter(srv)
	_, jobID := createOrgAndOpenJob(t, router, "apply@test.com", "applyorg", "Backend Dev")

	body := `{"name":"Jane Doe","email":"jane@example.com","phone":"+1234567890"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/careers/applyorg/jobs/"+jobID+"/apply",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["applicationId"] == nil || resp["applicationId"] == "" {
		t.Error("expected applicationId in response")
	}
	if resp["candidateId"] == nil || resp["candidateId"] == "" {
		t.Error("expected candidateId in response")
	}
}

func TestCareersApply_Duplicate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupCareersRouter(srv)
	_, jobID := createOrgAndOpenJob(t, router, "dup@test.com", "duporg", "Dev")

	body := `{"name":"John","email":"john@example.com"}`

	// First application
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("POST", "/api/careers/duporg/jobs/"+jobID+"/apply",
		bytes.NewBufferString(body))
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("first apply failed: %d %s", w1.Code, w1.Body.String())
	}

	// Duplicate application
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/api/careers/duporg/jobs/"+jobID+"/apply",
		bytes.NewBufferString(body))
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusConflict {
		t.Errorf("expected 409 for duplicate application, got %d: %s", w2.Code, w2.Body.String())
	}
}

func TestCareersApply_MissingFields(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupCareersRouter(srv)
	_, jobID := createOrgAndOpenJob(t, router, "missing@test.com", "missingorg", "Dev")

	body := `{"name":"NoEmail"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/careers/missingorg/jobs/"+jobID+"/apply",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCareersApply_RecordsStageTransition(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupCareersRouter(srv)
	_, jobID := createOrgAndOpenJob(t, router, "transition@test.com", "transorg", "Dev")

	body := `{"name":"Alice","email":"alice@example.com"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/careers/transorg/jobs/"+jobID+"/apply",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	appID := resp["applicationId"].(string)

	// Verify stage transition was recorded
	transitions, err := tdb.Queries.ListTransitionsByApplication(req.Context(), appID)
	if err != nil {
		t.Fatalf("failed to list transitions: %v", err)
	}
	if len(transitions) != 1 {
		t.Fatalf("expected 1 transition, got %d", len(transitions))
	}
	if transitions[0].ToStage != "applied" {
		t.Errorf("expected to_stage 'applied', got %q", transitions[0].ToStage)
	}
}
