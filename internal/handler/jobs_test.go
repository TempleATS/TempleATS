package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/temple-ats/TempleATS/internal/handler"
	"github.com/temple-ats/TempleATS/internal/testutil"
)

func TestCreateJob(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupCRUDRouter(srv)
	cookie := signupAndGetCookie(t, router, "job@test.com", "joborg")

	body := `{"title":"Go Developer","description":"Write Go microservices","location":"Remote","department":"Engineering","salary":"$120k-150k"}`
	w := httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("POST", "/api/jobs", body, cookie))

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["title"] != "Go Developer" {
		t.Errorf("expected title Go Developer, got %v", resp["title"])
	}
	if resp["status"] != "draft" {
		t.Errorf("expected default status draft, got %v", resp["status"])
	}
}

func TestCreateJob_MissingFields(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupCRUDRouter(srv)
	cookie := signupAndGetCookie(t, router, "jobmissing@test.com", "jobmissingorg")

	body := `{"title":"No Description"}`
	w := httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("POST", "/api/jobs", body, cookie))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestListJobs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupCRUDRouter(srv)
	cookie := signupAndGetCookie(t, router, "listjob@test.com", "listjoborg")

	// Create 2 jobs
	router.ServeHTTP(httptest.NewRecorder(), authedRequest("POST", "/api/jobs", `{"title":"Job 1","description":"Desc 1"}`, cookie))
	router.ServeHTTP(httptest.NewRecorder(), authedRequest("POST", "/api/jobs", `{"title":"Job 2","description":"Desc 2"}`, cookie))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("GET", "/api/jobs", "", cookie))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var jobs []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&jobs)
	if len(jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestUpdateJobStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupCRUDRouter(srv)
	cookie := signupAndGetCookie(t, router, "update@test.com", "updateorg")

	// Create job
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, authedRequest("POST", "/api/jobs", `{"title":"Draft Job","description":"Will be published"}`, cookie))
	var job map[string]interface{}
	json.NewDecoder(w1.Body).Decode(&job)
	jobID := job["id"].(string)

	// Update to open
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, authedRequest("PUT", "/api/jobs/"+jobID, `{"title":"Draft Job","description":"Will be published","status":"open"}`, cookie))

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w2.Code, w2.Body.String())
	}

	var updated map[string]interface{}
	json.NewDecoder(w2.Body).Decode(&updated)
	if updated["status"] != "open" {
		t.Errorf("expected status open, got %v", updated["status"])
	}
}

func TestJobOrgIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupCRUDRouter(srv)

	// Org A creates a job
	cookieA := signupAndGetCookie(t, router, "jobA@test.com", "jobisoorga")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, authedRequest("POST", "/api/jobs", `{"title":"OrgA Job","description":"Secret job"}`, cookieA))
	var jobA map[string]interface{}
	json.NewDecoder(w1.Body).Decode(&jobA)
	jobAID := jobA["id"].(string)

	// Org B tries to see org A's job
	cookieB := signupAndGetCookie(t, router, "jobB@test.com", "jobisoorgb")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, authedRequest("GET", "/api/jobs/"+jobAID, "", cookieB))
	if w2.Code != http.StatusNotFound {
		t.Errorf("expected 404 for cross-org job access, got %d", w2.Code)
	}

	// Org B should see empty list
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, authedRequest("GET", "/api/jobs", "", cookieB))
	var jobs []map[string]interface{}
	json.NewDecoder(w3.Body).Decode(&jobs)
	if len(jobs) != 0 {
		t.Errorf("expected 0 jobs for org B, got %d", len(jobs))
	}
}
