package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v9"
	"github.com/go-chi/chi"
	"github.com/magnm/lcm/config"
	"github.com/magnm/lcm/pkg/routes/google"
	"golang.org/x/exp/slog"
)

// Run starts a chi http server
func Run() {
	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		slog.Error("failed to parse config", "err", err)
		os.Exit(1)
	}
	config.Current = cfg

	var r *chi.Mux
	switch cfg.Type {
	case config.GoogleMetadata:
		r = google.Routes()
	default:
		slog.Error("unknown metadata type", "type", cfg.Type)
		os.Exit(1)
	}

	// Create a new http server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Create a channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start the http server
	go func() {
		fmt.Printf("Server listening on port %s\n", cfg.Port)
		if err := srv.ListenAndServe(); err != nil {
			fmt.Printf("Server error: %s\n", err)
		}
	}()

	// Wait for the OS signal to stop the server
	<-stop
	fmt.Println("Server stopped")

	// Give the server 5 seconds to shutdown gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
