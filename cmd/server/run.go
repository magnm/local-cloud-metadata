package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/magnm/lcm/pkg/routes/google"
)

// Run starts a chi http server
func Run() {
	// Create a new chi router
	r := google.Routes()

	// Create a new http server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Create a channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start the http server
	go func() {
		fmt.Printf("Server listening on port %s\n", port)
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
