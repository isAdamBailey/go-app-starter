package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

// ClientIPFallback fills in the client IP from the raw TCP peer
// (r.RemoteAddr) when nothing upstream has already resolved one via one of
// chi's middleware.ClientIPFrom* middlewares. Use this after
// middleware.ClientIPFromXFFTrustedProxies (or similar) in the chain: behind
// the production nginx proxy (see docs/DEPLOY.md) that middleware already
// resolves the real client IP from X-Forwarded-For and this is a no-op; with
// no reverse proxy in front of this service (e.g. local Docker Compose,
// where the frontend calls the API directly), this falls back to the actual
// socket peer.
func ClientIPFallback(next http.Handler) http.Handler {
	fallback := middleware.ClientIPFromRemoteAddr(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if middleware.GetClientIP(r.Context()) != "" {
			next.ServeHTTP(w, r)
			return
		}
		fallback.ServeHTTP(w, r)
	})
}

// clientIPKey is an httprate.KeyFunc keyed off the client IP resolved by
// chi's middleware.ClientIPFrom* middlewares (see ClientIPFallback), rather
// than httprate's own KeyByIP/KeyByRealIP, which either ignore
// X-Forwarded-For entirely or trust it unconditionally.
func clientIPKey(r *http.Request) (string, error) {
	if ip := middleware.GetClientIP(r.Context()); ip != "" {
		return ip, nil
	}
	return httprate.KeyByIP(r)
}
