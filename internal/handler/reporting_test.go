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

func setupReportingRouter(srv *handler.Server) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/api/auth/signup", srv.Signup)
	r.Post("/api/careers/{orgSlug}/jobs/{jobId}/apply", srv.CareersApply)
	r.Group(func(r chi.Router) {
		r.Use(mw.RequireAuth)
		r.Post("/api/reqs", srv.CreateRequisition)
		r.Post("/api/jobs", srv.CreateJob)
		r.Put("/api/jobs/{jobId}", srv.UpdateJob)
		r.Post("/api/reqs/{reqId}/jobs", srv.AttachJobToRequisition)
		r.Get("/api/reqs/{reqId}/report", srv.ReqReport)
		r.Put("/api/applications/{appId}/stage", srv.UpdateStage)
	})
	return r
}

func TestReqReport_Funnel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupReportingRouter(srv)

	cookie := signupAndGetCookie(t, router, "rep@test.com", "reportorg")

	// Create req
	w := httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("POST", "/api/reqs",
		`{"title":"Backend L5","targetHires":2}`, cookie))
	var req map[string]interface{}
	json.NewDecoder(w.Body).Decode(&req)
	reqID := req["id"].(string)

	// Create 2 jobs and attach to req
	w = httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("POST", "/api/jobs",
		`{"title":"Go Dev","description":"desc"}`, cookie))
	var job1 map[string]interface{}
	json.NewDecoder(w.Body).Decode(&job1)
	job1ID := job1["id"].(string)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("POST", "/api/jobs",
		`{"title":"Rust Dev","description":"desc"}`, cookie))
	var job2 map[string]interface{}
	json.NewDecoder(w.Body).Decode(&job2)
	job2ID := job2["id"].(string)

	// Publish both
	w = httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("PUT", "/api/jobs/"+job1ID,
		`{"title":"Go Dev","description":"desc","status":"open"}`, cookie))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("PUT", "/api/jobs/"+job2ID,
		`{"title":"Rust Dev","description":"desc","status":"open"}`, cookie))

	// Attach both to req
	w = httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("POST", "/api/reqs/"+reqID+"/jobs",
		`{"jobId":"`+job1ID+`"}`, cookie))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("POST", "/api/reqs/"+reqID+"/jobs",
		`{"jobId":"`+job2ID+`"}`, cookie))

	// Apply candidates: 3 to job1, 2 to job2
	app1 := applyCandidate(t, router, "reportorg", job1ID, "Alice", "alice@rep.com")
	app2 := applyCandidate(t, router, "reportorg", job1ID, "Bob", "bob@rep.com")
	_ = applyCandidate(t, router, "reportorg", job1ID, "Charlie", "charlie@rep.com")
	app4 := applyCandidate(t, router, "reportorg", job2ID, "Dave", "dave@rep.com")
	_ = applyCandidate(t, router, "reportorg", job2ID, "Eve", "eve@rep.com")

	// Move Alice to hired
	for _, stage := range []string{"screening", "interview", "offer", "hired"} {
		w = httptest.NewRecorder()
		router.ServeHTTP(w, authedRequest("PUT", "/api/applications/"+app1+"/stage",
			`{"stage":"`+stage+`"}`, cookie))
	}

	// Reject Bob
	w = httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("PUT", "/api/applications/"+app2+"/stage",
		`{"stage":"rejected","rejectionReason":"not_qualified"}`, cookie))

	// Move Dave to screening
	w = httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("PUT", "/api/applications/"+app4+"/stage",
		`{"stage":"screening"}`, cookie))

	// Get report
	w = httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("GET", "/api/reqs/"+reqID+"/report", "", cookie))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var report map[string]interface{}
	json.NewDecoder(w.Body).Decode(&report)

	// Verify funnel
	funnel := report["funnel"].(map[string]interface{})
	// 2 still in applied (Charlie + Eve)
	if funnel["applied"].(float64) != 2 {
		t.Errorf("expected funnel applied=2, got %v", funnel["applied"])
	}
	// 1 in screening (Dave)
	if funnel["screening"].(float64) != 1 {
		t.Errorf("expected funnel screening=1, got %v", funnel["screening"])
	}
	// 1 hired (Alice)
	if funnel["hired"].(float64) != 1 {
		t.Errorf("expected funnel hired=1, got %v", funnel["hired"])
	}
	// 1 rejected (Bob)
	if funnel["rejected"].(float64) != 1 {
		t.Errorf("expected funnel rejected=1, got %v", funnel["rejected"])
	}

	// Verify rejections
	rejections := report["rejections"].(map[string]interface{})
	byReason := rejections["byReason"].(map[string]interface{})
	if byReason["not_qualified"].(float64) != 1 {
		t.Errorf("expected rejection not_qualified=1, got %v", byReason["not_qualified"])
	}

	// Verify per-job breakdown
	byJob := report["byJob"].([]interface{})
	if len(byJob) != 2 {
		t.Errorf("expected 2 jobs in breakdown, got %d", len(byJob))
	}

	// Verify fill progress
	fill := report["fillProgress"].(map[string]interface{})
	if fill["hired"].(float64) != 1 {
		t.Errorf("expected fillProgress hired=1, got %v", fill["hired"])
	}
	if fill["target"].(float64) != 2 {
		t.Errorf("expected fillProgress target=2, got %v", fill["target"])
	}
}

func TestReqReport_EmptyReq(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupReportingRouter(srv)

	cookie := signupAndGetCookie(t, router, "empty@test.com", "emptyreport")

	// Create req with no jobs
	w := httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("POST", "/api/reqs",
		`{"title":"Empty Req"}`, cookie))
	var req map[string]interface{}
	json.NewDecoder(w.Body).Decode(&req)
	reqID := req["id"].(string)

	// Get report
	w = httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("GET", "/api/reqs/"+reqID+"/report", "", cookie))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var report map[string]interface{}
	json.NewDecoder(w.Body).Decode(&report)

	funnel := report["funnel"].(map[string]interface{})
	if funnel["applied"].(float64) != 0 {
		t.Errorf("expected funnel applied=0, got %v", funnel["applied"])
	}
}
