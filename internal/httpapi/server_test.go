package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"rehab-followup/internal/service"
	"rehab-followup/internal/store"
)

func TestHealthAndLoginEndpoints(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/api.db")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	p := service.NewPlatform(s, time.Now)
	if _, err := p.RegisterTherapist("t1", "Lin", "Rehab", "secret"); err != nil {
		t.Fatal(err)
	}
	server := NewServer(p)
	health := httptest.NewRecorder()
	server.Handler().ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/health", nil))
	if health.Code != http.StatusOK || !strings.Contains(health.Body.String(), "ok") {
		t.Fatal("health endpoint failed")
	}
	login := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"ID":"t1","Password":"secret"}`))
	server.Handler().ServeHTTP(login, request)
	if login.Code != http.StatusOK || !strings.Contains(login.Body.String(), "Token") {
		t.Fatalf("login endpoint failed: %d %s", login.Code, login.Body.String())
	}
}
