// ABOUTME: Entry point for the mock BOSH Director server.
// ABOUTME: Parses CLI flags and starts the HTTP server.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/malston/bosh-mock-director/internal/mockbosh"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("mock-bosh-director %s\n", version)
		os.Exit(0)
	}

	config := mockbosh.DefaultServerConfig()

	flag.IntVar(&config.Port, "port", config.Port, "Port to listen on")
	flag.StringVar(&config.Username, "username", config.Username, "Basic auth username")
	flag.StringVar(&config.Password, "password", config.Password, "Basic auth password")
	flag.BoolVar(&config.UseTLS, "tls", config.UseTLS, "Enable TLS with self-signed cert")
	flag.Float64Var(&config.Speed, "speed", config.Speed, "Simulation speed multiplier (1.0 = normal)")
	flag.BoolVar(&config.Debug, "debug", config.Debug, "Enable debug logging")
	flag.Parse()

	server := mockbosh.NewServer(config)

	// Handle shutdown signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.Start()
	}()

	// Wait for shutdown signal or error
	select {
	case err := <-serverErr:
		if err != nil {
			log.Fatalf("Server error: %v", err)
		}
	case sig := <-shutdown:
		log.Printf("Received signal %v, shutting down...", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Shutdown error: %v", err)
		}
	}
}
