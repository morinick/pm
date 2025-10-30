package infra

import (
	"context"
	"encoding/json"
	"net/http"
)

type ISessionManager interface {
	Keys(context.Context) []string
}

func AuthMiddleware(sm ISessionManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if keys := sm.Keys(r.Context()); len(keys) == 0 {
				ErrorHandler(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func ResponseJSON(w http.ResponseWriter, data any, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(data)
}

func ErrorHandler(w http.ResponseWriter, code int, msg string) {
	ResponseJSON(w, struct {
		Error string `json:"error"`
	}{Error: msg}, code)
}
