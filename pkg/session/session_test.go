package session

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestLoadAndSave(t *testing.T) {
	sm, _ := NewSessionManager(context.Background(), SessionManagerOptions{})
	router := chi.NewRouter()
	router.Use(sm.LoadAndSave)
	router.Get("/test-set-data-to-session", func(w http.ResponseWriter, r *http.Request) {
		userID := "userID"
		sm.Put(r.Context(), "user_id", userID)
		w.WriteHeader(http.StatusOK)
	})
	router.Get("/test-saved-data-in-session", func(w http.ResponseWriter, r *http.Request) {
		if data := sm.GetString(r.Context(), "user_id"); len(data) == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(data))
		}
	})
	router.Get("/test-destroy-session", func(w http.ResponseWriter, r *http.Request) {
		if len(r.CookiesNamed("session")) == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		sm.Destroy(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	r1 := httptest.NewRequest(http.MethodGet, "/test-set-data-to-session", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, r1)
	var sessionCookie *http.Cookie
	for _, cookie := range w1.Result().Cookies() {
		if cookie.Name == "session" && len(cookie.Value) > 0 {
			sessionCookie = cookie
		}
	}

	if sessionCookie == nil {
		t.Fatal("Wrong! Session cookie is not set!\n")
	}

	r2 := httptest.NewRequest(http.MethodGet, "/test-saved-data-in-session", nil)
	r2.AddCookie(sessionCookie)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, r2)
	var sessionCookie2 *http.Cookie
	for _, cookie := range w2.Result().Cookies() {
		if cookie.Name == "session" && len(cookie.Value) > 0 {
			sessionCookie2 = cookie
		}
	}

	if got, want := w2.Code, http.StatusOK; got != want {
		t.Fatalf("Wrong! Data is not set in session!\n\tExpected code: %d\n\tActual code: %d", want, got)
	}
	if got, want := w2.Body.String(), "userID"; got != want {
		t.Fatalf("Wrong! Incorrect data in session!\n\tExpected: %s\n\tActual: %s", want, got)
	}
	if sessionCookie.Value == sessionCookie2.Value {
		t.Fatal("Wrong! Value of cookies must be different!")
	}

	r3 := httptest.NewRequest(http.MethodGet, "/test-destroy-session", nil)
	r3.AddCookie(sessionCookie2)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, r3)
	var sessionCookie3 *http.Cookie
	for _, cookie := range w3.Result().Cookies() {
		if cookie.Name == "session" && len(cookie.Value) > 0 {
			sessionCookie3 = cookie
		}
	}

	if got, want := w3.Code, http.StatusOK; got != want {
		t.Fatalf("Wrong! Session is not set!\n\tExpected code: %d\n\tActual code: %d", want, got)
	}
	if sessionCookie3 != nil {
		t.Fatalf("Wrong! Session not destroyed!")
	}
}

