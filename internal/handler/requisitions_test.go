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

func setupCRUDRouter(srv *handler.Server) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/api/auth/signup", srv.Signup)
	r.Post("/api/auth/login", srv.Login)
	r.Group(func(r chi.Router) {
		r.Use(mw.RequireAuth)
		r.Get("/api/reqs", srv.ListRequisitions)
		r.Post("/api/reqs", srv.CreateRequisition)
		r.Get("/api/reqs/{reqId}", srv.GetRequisition)
		r.Put("/api/reqs/{reqId}", srv.UpdateRequisition)
		r.Post("/api/reqs/{reqId}/jobs", srv.AttachJobToRequisition)
		r.Get("/api/jobs", srv.ListJobs)
		r.Post("/api/jobs", srv.CreateJob)
		r.Get("/api/jobs/{jobId}", srv.GetJob)
		r.Put("/api/jobs/{jobId}", srv.UpdateJob)
	})
	return r
}

// signupAndGetCookie creates an org+user and returns the token cookie.
func signupAndGetCookie(t *testing.T, router *chi.Mux, email, orgSlug string) *http.Cookie {
	t.Helper()
	body := `{"email":"` + email + `","name":"Test","password":"pass123","orgName":"TestOrg","orgSlug":"` + orgSlug + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("signup failed: %d %s", w.Code, w.Body.String())
	}
	for _, c := range w.Result().Cookies() {
		if c.Name == "token" {
			return c
		}
	}
	t.Fatal("no token cookie")
	return nil
}

func authedRequest(method, url string, body string, cookie *http.Cookie) *http.Request {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, url, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, url, nil)
	}
	req.AddCookie(cookie)
	return req
}

func TestCreateRequisition(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupCRUDRouter(srv)
	cookie := signupAndGetCookie(t, router, "req@test.com", "reqorg")

	body := `{"title":"Senior Engineer","level":"L5","department":"Engineering","targetHires":2}`
	w := httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("POST", "/api/reqs", body, cookie))

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["title"] != "Senior Engineer" {
		t.Errorf("expected title Senior Engineer, got %v", resp["title"])
	}
}

func TestListRequisitions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupCRUDRouter(srv)
	cookie := signupAndGetCookie(t, router, "list@test.com", "listorg")

	// Create 2 reqs
	router.ServeHTTP(httptest.NewRecorder(), authedRequest("POST", "/api/reqs", `{"title":"Req 1"}`, cookie))
	router.ServeHTTP(httptest.NewRecorder(), authedRequest("POST", "/api/reqs", `{"title":"Req 2"}`, cookie))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, authedRequest("GET", "/api/reqs", "", cookie))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var reqs []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&reqs)
	if len(reqs) != 2 {
		t.Errorf("expected 2 reqs, got %d", len(reqs))
	}
}

func TestGetRequisitionWithJobs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupCRUDRouter(srv)
	cookie := signupAndGetCookie(t, router, "get@test.com", "getorg")

	// Create req
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, authedRequest("POST", "/api/reqs", `{"title":"Backend Eng"}`, cookie))
	var req map[string]interface{}
	json.NewDecoder(w1.Body).Decode(&req)
	reqID := req["id"].(string)

	// Create job and attach to req
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, authedRequest("POST", "/api/jobs", `{"title":"Go Developer","description":"Write Go code"}`, cookie))
	var job map[string]interface{}
	json.NewDecoder(w2.Body).Decode(&job)
	jobID := job["id"].(string)

	// Attach job
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, authedRequest("POST", "/api/reqs/"+reqID+"/jobs", `{"jobId":"`+jobID+`"}`, cookie))
	if w3.Code != http.StatusOK {
		t.Fatalf("attach failed: %d %s", w3.Code, w3.Body.String())
	}

	// Get req with jobs
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, authedRequest("GET", "/api/reqs/"+reqID, "", cookie))
	if w4.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w4.Code, w4.Body.String())
	}

	var result map[string]interface{}
	json.NewDecoder(w4.Body).Decode(&result)
	jobs := result["jobs"].([]interface{})
	if len(jobs) != 1 {
		t.Errorf("expected 1 attached job, got %d", len(jobs))
	}
}

func TestRequisitionOrgIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupCRUDRouter(srv)

	// Org A creates a req
	cookieA := signupAndGetCookie(t, router, "orgA@test.com", "orga")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, authedRequest("POST", "/api/reqs", `{"title":"OrgA Req"}`, cookieA))
	var reqA map[string]interface{}
	json.NewDecoder(w1.Body).Decode(&reqA)
	reqAID := reqA["id"].(string)

	// Org B tries to see org A's req
	cookieB := signupAndGetCookie(t, router, "orgB@test.com", "orgb")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, authedRequest("GET", "/api/reqs/"+reqAID, "", cookieB))
	if w2.Code != http.StatusNotFound {
		t.Errorf("expected 404 for cross-org access, got %d", w2.Code)
	}

	// Org B should see empty list
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, authedRequest("GET", "/api/reqs", "", cookieB))
	var reqs []map[string]interface{}
	json.NewDecoder(w3.Body).Decode(&reqs)
	if len(reqs) != 0 {
		t.Errorf("expected 0 reqs for org B, got %d", len(reqs))
	}
}
