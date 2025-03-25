// Package client implements the echo client
// This file is kept for compatibility but the functionality has been moved to the main package.
package client

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"
)

// MainClient runs the echo client
func MainClient() {
	// Parse command-line flags
	var (
		httpURL   = flag.String("http", "http://localhost:8080", "HTTP server URL")
		transport = flag.String("transport", "stdio", "Transport type: 'stdio' or 'http'")
		message   = flag.String("message", "Hello, world!", "Message to echo")
		timeout   = flag.Duration("timeout", 5*time.Second, "Timeout for the request")
	)
	flag.Parse()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Create the echo client
	client := NewEchoClient()

	// Configure the transport
	switch *transport {
	case "stdio":
		client.UseStdioTransport()
		fmt.Fprintf(os.Stderr, "Using stdio transport\n")
	case "http":
		client.UseHTTPTransport(*httpURL)
		fmt.Fprintf(os.Stderr, "Using HTTP transport with URL: %s\n", *httpURL)
	default:
		fmt.Fprintf(os.Stderr, "Error: unsupported transport type: %s\n", *transport)
		fmt.Fprintf(os.Stderr, "Supported types: stdio, http\n")
		os.Exit(1)
	}

	// Send the echo request
	fmt.Fprintf(os.Stderr, "Sending message: %s\n", *message)
	response, err := client.Echo(ctx, *message)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Print the response
	fmt.Fprintf(os.Stderr, "Response: %s\n", response)
}
