package routes

import (
	"os"

	"github.com/go-chi/chi"
	"github.com/magnm/lcm/config"
	"github.com/magnm/lcm/pkg/routes/google"
	"github.com/magnm/lcm/pkg/routes/webhook"
	"golang.org/x/exp/slog"
)

func MainRouter(cfg config.Config) *chi.Mux {
	r := chi.NewRouter()

	r.Mount("/webhook", webhook.Routes())

	switch cfg.Type {
	case config.GoogleMetadata:
		r.Mount("/", google.Routes())
	default:
		slog.Error("unknown metadata type", "type", cfg.Type)
		os.Exit(1)
	}

	return r
}
