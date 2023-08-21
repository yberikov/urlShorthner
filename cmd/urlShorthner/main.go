package main

import (
	"net/http"
	"os"
	"urlshorthner/internal/config"
	"urlshorthner/internal/http-server/handlers/redirect"
	"urlshorthner/internal/http-server/handlers/url/save"
	"urlshorthner/internal/lib/logger/sl"
	"urlshorthner/internal/storage/sqlite"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slog"
)

func main() {
	cfg := config.LoadConfig()

	log := setupLogger(cfg.Env)

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	router := chi.NewRouter()

	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/url", save.New(log, storage))
	router.Get("/{alias}", redirect.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}
	log.Error("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	// add dev and prod
	switch env {
	case "local":
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}

	return log
}
