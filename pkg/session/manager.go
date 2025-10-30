package session

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
)

type SessionManagerOptions struct {
	Lifetime       time.Duration
	CookieName     string
	CookieHTTPOnly bool
	CookieSameSite http.SameSite
}

type SessionManager struct {
	sm *scs.SessionManager
}

func NewSessionManager(ctx context.Context, opts SessionManagerOptions) (*SessionManager, error) {
	opts = prepareOptions(opts)

	sm := scs.New()

	sm.Lifetime = opts.Lifetime
	sm.Cookie.Name = opts.CookieName
	sm.Cookie.HttpOnly = opts.CookieHTTPOnly
	sm.Cookie.SameSite = opts.CookieSameSite

	return &SessionManager{sm: sm}, nil
}

func (smw *SessionManager) LoadAndSave(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookieName := smw.sm.Cookie.Name

		var token string
		if cookie, err := r.Cookie(cookieName); err == nil {
			token = cookie.Value
		}

		ctx, err := smw.sm.Load(r.Context(), token)
		if err != nil {
			smw.sm.ErrorFunc(w, r, err)
			return
		}

		rs := r.WithContext(ctx)
		wb := &bufferedResponseWriter{ResponseWriter: w}

		next.ServeHTTP(wb, rs)

		if smw.sm.Status(ctx) == scs.Destroyed {
			smw.sm.WriteSessionCookie(ctx, w, "", time.Time{})
		} else {
			if err := smw.sm.RenewToken(ctx); err != nil {
				smw.sm.ErrorFunc(w, r, err)
				return
			}

			newToken, expiry, err := smw.sm.Commit(ctx)
			if err != nil {
				smw.sm.ErrorFunc(w, r, err)
				return
			}

			smw.sm.WriteSessionCookie(ctx, w, newToken, expiry)
		}

		if wb.code != 0 {
			w.WriteHeader(wb.code)
		}
		w.Write(wb.buf.Bytes())
	})
}

func (smw *SessionManager) GetString(ctx context.Context, key string) string {
	return smw.sm.GetString(ctx, key)
}

func (smw *SessionManager) Put(ctx context.Context, key string, value any) {
	smw.sm.Put(ctx, key, value)
}

func (smw *SessionManager) Destroy(cxt context.Context) error {
	return smw.sm.Destroy(cxt)
}

func (smw *SessionManager) Keys(ctx context.Context) []string {
	return smw.sm.Keys(ctx)
}

func prepareOptions(opts SessionManagerOptions) SessionManagerOptions {
	if opts.Lifetime == 0 {
		opts.Lifetime = 24 * time.Hour
	}

	if len(opts.CookieName) == 0 {
		opts.CookieName = "session"
	}

	if opts.CookieSameSite == 0 {
		opts.CookieSameSite = http.SameSiteLaxMode
	}

	return opts
}

type bufferedResponseWriter struct {
	http.ResponseWriter
	buf  bytes.Buffer
	code int
}

func (bw *bufferedResponseWriter) Write(b []byte) (int, error) {
	return bw.buf.Write(b)
}

func (bw *bufferedResponseWriter) WriteHeader(code int) {
	bw.code = code
}
