// Command server runs the HTTP API.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"github.com/isAdamBailey/go-app-starter/backend/internal/auth"
	"github.com/isAdamBailey/go-app-starter/backend/internal/config"
	"github.com/isAdamBailey/go-app-starter/backend/internal/db"
	"github.com/isAdamBailey/go-app-starter/backend/internal/httpapi"
	"github.com/isAdamBailey/go-app-starter/backend/internal/mailer"
	"github.com/isAdamBailey/go-app-starter/backend/internal/users"
)

func main() {
	// Best-effort: outside Docker/Forge, env vars usually come from a local
	// .env file. In Docker Compose and Forge, real env vars are already set
	// and this is a silent no-op (no .env file present in the image/site).
	_ = godotenv.Load()

	if err := run(); err != nil {
		slog.Error("server exited", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if err := db.Migrate(cfg.DatabaseURL); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	queries := db.New(pool)
	userRepo := users.NewPostgresRepository(queries)

	if err := userRepo.SyncAllowlist(ctx, cfg.AllowedEmails); err != nil {
		return err
	}

	mailSvc, err := mailer.New(cfg.Mailer)
	if err != nil {
		return err
	}

	authSvc := auth.NewService(queries, userRepo, mailSvc, cfg.CookieSigningSecret, cfg.CookieSecure, cfg.AppBaseURL)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	// Behind the production nginx proxy (docs/DEPLOY.md) there is exactly one
	// trusted hop; ClientIPFallback covers the no-proxy case (local Docker
	// Compose, where the frontend calls this service directly).
	r.Use(middleware.ClientIPFromXFFTrustedProxies(1))
	r.Use(httpapi.ClientIPFallback)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.AppBaseURL},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	httpapi.NewHandler(authSvc, userRepo).Register(r)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		slog.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
