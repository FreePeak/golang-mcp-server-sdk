// Package server implements the echo server
// This file is kept for compatibility but the functionality has been moved to the main package.
package server

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func MainServer() {
	// Parse command-line flags
	var (
		httpAddr  = flag.String("http", ":8080", "HTTP server address (default :8080)")
		transport = flag.String("transport", "stdio", "Transport type: 'stdio' or 'http'")
	)
	flag.Parse()

	// Create a context that cancels on SIGINT/SIGTERM
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		cancel()
	}()

	// Create the echo server
	srv := NewEchoServer()

	// Configure the transport
	var transportType TransportType
	switch *transport {
	case "stdio":
		transportType = StdioTransport
	case "http":
		transportType = HTTPTransport
	default:
		fmt.Fprintf(os.Stderr, "Error: unsupported transport type: %s\n", *transport)
		fmt.Fprintf(os.Stderr, "Supported types: stdio, http\n")
		os.Exit(1)
	}

	// Connect with the appropriate transport
	err := srv.UseTransport(transportType, *httpAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error configuring transport: %v\n", err)
		os.Exit(1)
	}

	// Start the server
	err = srv.Start(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}

	// Wait for context cancellation (from signal handler)
	<-ctx.Done()

	// Stop the server
	if err := srv.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "Error stopping server: %v\n", err)
	}
}
