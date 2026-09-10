// Package httpapi wires up HTTP routes and handlers for the API.
package httpapi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"

	"github.com/isAdamBailey/go-app-starter/backend/internal/auth"
	"github.com/isAdamBailey/go-app-starter/backend/internal/users"
)

// magicLinkRateLimit caps how many magic-link requests a single IP address
// may make per minute. This is in addition to auth.Service's per-email hourly
// cap; the two limits protect against different abuse shapes (one IP
// hammering many emails vs. one email being spammed from many IPs).
const magicLinkRateLimit = 5

// verifyRateLimit caps how many token-verification attempts a single IP may
// make per minute, as defense in depth against token-guessing even though
// tokens are 256 bits of entropy.
const verifyRateLimit = 20

// Handler holds the dependencies needed to serve the API.
type Handler struct {
	auth  *auth.Service
	users users.Repository
}

// NewHandler constructs a Handler.
func NewHandler(authSvc *auth.Service, userRepo users.Repository) *Handler {
	return &Handler{
		auth:  authSvc,
		users: userRepo,
	}
}

// Register attaches all API routes to the given router.
func (h *Handler) Register(r chi.Router) {
	r.Get("/healthz", healthz)

	r.Route("/api", func(r chi.Router) {
		r.Use(securityHeaders)

		r.With(httprate.Limit(magicLinkRateLimit, time.Minute, httprate.WithKeyFuncs(clientIPKey))).
			Post("/auth/magic-link", h.requestMagicLink)
		r.With(httprate.Limit(verifyRateLimit, time.Minute, httprate.WithKeyFuncs(clientIPKey))).
			Post("/auth/verify", h.verifyMagicLink)

		r.Group(func(r chi.Router) {
			r.Use(h.requireAuth)
			r.Get("/me", h.me)
			r.With(h.requireCSRF).Post("/auth/logout", h.logout)

			// TODO: register your app's authenticated routes here, e.g.:
			// r.Get("/widgets", h.listWidgets)
			// r.With(h.requireCSRF).Post("/widgets", h.createWidget)
		})
	})
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
