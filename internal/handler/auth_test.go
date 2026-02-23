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

func setupAuthRouter(srv *handler.Server) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/api/auth/signup", srv.Signup)
	r.Post("/api/auth/login", srv.Login)
	r.Post("/api/auth/logout", srv.Logout)
	r.Group(func(r chi.Router) {
		r.Use(mw.RequireAuth)
		r.Get("/api/auth/me", srv.Me)
	})
	return r
}

func TestSignup_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupAuthRouter(srv)

	body := `{"email":"test@example.com","name":"Test User","password":"password123","orgName":"Acme Corp","orgSlug":"acme"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["email"] != "test@example.com" {
		t.Errorf("expected email test@example.com, got %v", resp["email"])
	}
	if resp["orgSlug"] != "acme" {
		t.Errorf("expected orgSlug acme, got %v", resp["orgSlug"])
	}
	if resp["role"] != "admin" {
		t.Errorf("expected role admin, got %v", resp["role"])
	}

	// Should have set a token cookie
	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "token" && c.Value != "" {
			found = true
		}
	}
	if !found {
		t.Error("expected token cookie to be set")
	}
}

func TestSignup_DuplicateEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupAuthRouter(srv)

	body := `{"email":"dup@example.com","name":"User 1","password":"pass123","orgName":"Org1","orgSlug":"org1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("first signup failed: %d %s", w.Code, w.Body.String())
	}

	// Second signup with same email
	body2 := `{"email":"dup@example.com","name":"User 2","password":"pass123","orgName":"Org2","orgSlug":"org2"}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewBufferString(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Errorf("expected 409 for duplicate email, got %d: %s", w2.Code, w2.Body.String())
	}
}

func TestSignup_DuplicateSlug(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupAuthRouter(srv)

	body := `{"email":"user1@example.com","name":"User 1","password":"pass123","orgName":"Org1","orgSlug":"sameslug"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("first signup failed: %d %s", w.Code, w.Body.String())
	}

	body2 := `{"email":"user2@example.com","name":"User 2","password":"pass123","orgName":"Org2","orgSlug":"sameslug"}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewBufferString(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Errorf("expected 409 for duplicate slug, got %d: %s", w2.Code, w2.Body.String())
	}
}

func TestSignup_MissingFields(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupAuthRouter(srv)

	body := `{"email":"test@example.com","password":"pass123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLogin_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupAuthRouter(srv)

	// First signup
	signup := `{"email":"login@example.com","name":"Login User","password":"mypassword","orgName":"LoginOrg","orgSlug":"loginorg"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewBufferString(signup))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("signup failed: %d %s", w.Code, w.Body.String())
	}

	// Now login
	login := `{"email":"login@example.com","password":"mypassword"}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(login))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w2.Code, w2.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w2.Body).Decode(&resp)

	if resp["email"] != "login@example.com" {
		t.Errorf("expected email login@example.com, got %v", resp["email"])
	}

	// Should have token cookie
	cookies := w2.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "token" && c.Value != "" {
			found = true
		}
	}
	if !found {
		t.Error("expected token cookie to be set")
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupAuthRouter(srv)

	signup := `{"email":"badpass@example.com","name":"User","password":"correct","orgName":"Org","orgSlug":"badpassorg"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewBufferString(signup))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	login := `{"email":"badpass@example.com","password":"wrong"}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(login))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong password, got %d", w2.Code)
	}
}

func TestLogin_NonexistentUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupAuthRouter(srv)

	login := `{"email":"nobody@example.com","password":"pass"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(login))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for nonexistent user, got %d", w.Code)
	}
}

func TestMe_Authenticated(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupAuthRouter(srv)

	// Signup first
	signup := `{"email":"me@example.com","name":"Me User","password":"pass123","orgName":"MeOrg","orgSlug":"meorg"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewBufferString(signup))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Extract token cookie
	var tokenCookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "token" {
			tokenCookie = c
		}
	}
	if tokenCookie == nil {
		t.Fatal("no token cookie from signup")
	}

	// Call /me with the cookie
	req2 := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req2.AddCookie(tokenCookie)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w2.Code, w2.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w2.Body).Decode(&resp)

	if resp["email"] != "me@example.com" {
		t.Errorf("expected email me@example.com, got %v", resp["email"])
	}
	if resp["orgSlug"] != "meorg" {
		t.Errorf("expected orgSlug meorg, got %v", resp["orgSlug"])
	}
}

func TestMe_Unauthenticated(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup()

	srv := handler.NewServer(tdb.Pool)
	router := setupAuthRouter(srv)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestLogout(t *testing.T) {
	srv := handler.NewServer(nil)
	router := setupAuthRouter(srv)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Token cookie should be cleared (MaxAge -1)
	for _, c := range w.Result().Cookies() {
		if c.Name == "token" && c.MaxAge != -1 {
			t.Error("expected token cookie to be cleared (MaxAge -1)")
		}
	}
}
