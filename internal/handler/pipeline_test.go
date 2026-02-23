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

func setupPipelineRouter(srv *handler.Server) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/api/auth/signup", srv.Signup)
	// Public apply route
	r.Get("/api/careers/{orgSlug}", srv.CareersListJobs)
	r.Post("/api/careers/{orgSlug}/jobs/{jobId}/apply", srv.CareersApply)
	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(mw.RequireAuth)
		r.Post("/api/jobs", srv.CreateJob)
		r.Put("/api/jobs/{jobId}", srv.UpdateJob)
		r.Get("/api/jobs/{jobId}/pipeline", srv.GetPipeline)
		r.Put("/api/applications/{appId}/stage", srv.UpdateStage)
		r.Get("/api/applications/{appId}", srv.GetApplication)
	})
	return r
}

// applyCandidate creates a candidate application and returns the applicationId.
func applyCandidate(t *testing.T, router *chi.Mux, orgSlug, jobID, name, email string) string {
	t.Helper()
	body := `{"name":"` + name + `","email":"` + email + `"}`
	req := httptest.NewRequest("POST", "/api/careers/"+orgSlug+"/jobs/"+jobID+"/apply", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("apply failed: %d %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	return resp["applicationId"].(string)
}

func TestGetPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupPipelineRouter(srv)

	// Create org with open job
	cookie, jobID := createOrgAndOpenJob(t, router, "pipe@test.com", "pipeorg", "Engineer")

	// Apply 3 candidates
	app1 := applyCandidate(t, router, "pipeorg", jobID, "Alice", "alice@test.com")
	app2 := applyCandidate(t, router, "pipeorg", jobID, "Bob", "bob@test.com")
	_ = applyCandidate(t, router, "pipeorg", jobID, "Charlie", "charlie@test.com")

	// Move Alice to screening
	w := httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("PUT", "/api/applications/"+app1+"/stage",
		`{"stage":"screening"}`, cookie))
	if w.Code != http.StatusOK {
		t.Fatalf("move alice failed: %d %s", w.Code, w.Body.String())
	}

	// Move Bob to interview
	w = httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("PUT", "/api/applications/"+app2+"/stage",
		`{"stage":"screening"}`, cookie))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("PUT", "/api/applications/"+app2+"/stage",
		`{"stage":"interview"}`, cookie))

	// Get pipeline
	w = httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("GET", "/api/jobs/"+jobID+"/pipeline", "", cookie))
	if w.Code != http.StatusOK {
		t.Fatalf("get pipeline failed: %d %s", w.Code, w.Body.String())
	}

	var pipeline map[string]interface{}
	json.NewDecoder(w.Body).Decode(&pipeline)

	stages := pipeline["stages"].(map[string]interface{})

	// Charlie still in applied
	applied := stages["applied"].([]interface{})
	if len(applied) != 1 {
		t.Errorf("expected 1 in applied, got %d", len(applied))
	}

	// Alice in screening
	screening := stages["screening"].([]interface{})
	if len(screening) != 1 {
		t.Errorf("expected 1 in screening, got %d", len(screening))
	}

	// Bob in interview
	interview := stages["interview"].([]interface{})
	if len(interview) != 1 {
		t.Errorf("expected 1 in interview, got %d", len(interview))
	}
}

func TestUpdateStage_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupPipelineRouter(srv)

	cookie, jobID := createOrgAndOpenJob(t, router, "stage@test.com", "stageorg", "Dev")
	appID := applyCandidate(t, router, "stageorg", jobID, "Dave", "dave@test.com")

	// Move to screening
	w := httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("PUT", "/api/applications/"+appID+"/stage",
		`{"stage":"screening"}`, cookie))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["stage"] != "screening" {
		t.Errorf("expected stage screening, got %v", resp["stage"])
	}
}

func TestUpdateStage_Rejection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupPipelineRouter(srv)

	cookie, jobID := createOrgAndOpenJob(t, router, "rej@test.com", "rejorg", "Dev")
	appID := applyCandidate(t, router, "rejorg", jobID, "Eve", "eve@test.com")

	// Reject with reason
	w := httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("PUT", "/api/applications/"+appID+"/stage",
		`{"stage":"rejected","rejectionReason":"not_qualified","rejectionNotes":"Missing Go experience"}`, cookie))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["stage"] != "rejected" {
		t.Errorf("expected stage rejected, got %v", resp["stage"])
	}
	if resp["rejection_reason"] == nil {
		t.Error("expected rejection_reason to be set")
	}
}

func TestUpdateStage_RejectionRequiresReason(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupPipelineRouter(srv)

	cookie, jobID := createOrgAndOpenJob(t, router, "rejr@test.com", "rejrorg", "Dev")
	appID := applyCandidate(t, router, "rejrorg", jobID, "Frank", "frank@test.com")

	// Reject without reason should fail
	w := httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("PUT", "/api/applications/"+appID+"/stage",
		`{"stage":"rejected"}`, cookie))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUpdateStage_RecordsTransition(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupPipelineRouter(srv)

	cookie, jobID := createOrgAndOpenJob(t, router, "trans@test.com", "transorg", "Dev")
	appID := applyCandidate(t, router, "transorg", jobID, "Grace", "grace@test.com")

	// Move applied -> screening -> interview
	w := httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("PUT", "/api/applications/"+appID+"/stage",
		`{"stage":"screening"}`, cookie))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("PUT", "/api/applications/"+appID+"/stage",
		`{"stage":"interview"}`, cookie))

	// Get application detail - should include transitions
	w = httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("GET", "/api/applications/"+appID, "", cookie))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	transitions := resp["transitions"].([]interface{})
	// Initial (null -> applied) + screening + interview = 3 transitions
	if len(transitions) != 3 {
		t.Errorf("expected 3 transitions, got %d", len(transitions))
	}
}

func TestUpdateStage_InvalidStage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupPipelineRouter(srv)

	cookie, jobID := createOrgAndOpenJob(t, router, "inv@test.com", "invorg", "Dev")
	appID := applyCandidate(t, router, "invorg", jobID, "Hank", "hank@test.com")

	// Try invalid stage
	w := httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("PUT", "/api/applications/"+appID+"/stage",
		`{"stage":"invalid_stage"}`, cookie))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
