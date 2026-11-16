package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestHealthzReturnsOKJSON(t *testing.T) {
	handler := NewHandler(fstest.MapFS{"index.html": {Data: []byte("home")}}, "test-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"status":"ok"`) {
		t.Fatalf("health response = %d %q", response.Code, response.Body.String())
	}
}

func TestUnknownRouteFallsBackToHome(t *testing.T) {
	handler := NewHandler(fstest.MapFS{"index.html": {Data: []byte("app home")}}, "test-token")
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/courses/example", nil)
	request.AddCookie(&http.Cookie{Name: "biblioteca_session", Value: "test-token"})
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Body.String() != "app home" {
		t.Fatalf("SPA response = %d %q", response.Code, response.Body.String())
	}
}

func TestSessionTokenBootstrapsHttpOnlyCookie(t *testing.T) {
	handler := NewHandler(fstest.MapFS{"index.html": {Data: []byte("home")}}, "test-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/?session=test-token", nil))
	if response.Code != http.StatusSeeOther {
		t.Fatalf("token bootstrap status = %d, want %d", response.Code, http.StatusSeeOther)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "biblioteca_session" || !cookies[0].HttpOnly {
		t.Fatalf("expected HttpOnly session cookie, got %#v", cookies)
	}
}

func TestFrontendRequiresSessionCookie(t *testing.T) {
	handler := NewHandler(fstest.MapFS{"index.html": {Data: []byte("home")}}, "test-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}
