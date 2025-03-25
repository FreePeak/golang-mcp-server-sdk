package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FreePeak/golang-mcp-server-sdk/examples/echo/client"
	"github.com/FreePeak/golang-mcp-server-sdk/examples/echo/server"
)

func main() {
	// Parse command-line flags
	var (
		mode      = flag.String("mode", "", "Mode: 'server' or 'client'")
		transport = flag.String("transport", "stdio", "Transport type: 'stdio' or 'http'")
		httpAddr  = flag.String("http", ":8080", "HTTP server address (default :8080)")
		message   = flag.String("message", "Hello, world!", "Message to echo")
		timeout   = flag.Duration("timeout", 5*time.Second, "Timeout for the request")
	)

	// Define custom usage
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -mode=server [server flags]\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -mode=client [client flags]\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "Server flags:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  -transport=stdio|http   Transport type (default: stdio)\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  -http=:8080            HTTP server address (only for http transport)\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Client flags:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  -transport=stdio|http   Transport type (default: stdio)\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  -http=URL              HTTP server URL (default: http://localhost:8080)\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  -message=TEXT          Message to echo (default: Hello, world!)\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  -timeout=DURATION      Request timeout (default: 5s)\n")
	}

	flag.Parse()

	// Check mode
	switch *mode {
	case "server":
		runServer(*transport, *httpAddr)
	case "client":
		runClient(*transport, *httpAddr, *message, *timeout)
	default:
		flag.Usage()
		os.Exit(1)
	}
}

func runServer(transportType string, httpAddr string) {
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
	srv := server.NewEchoServer()

	// Configure the transport
	var serverTransportType server.TransportType
	switch transportType {
	case "stdio":
		serverTransportType = server.StdioTransport
	case "http":
		serverTransportType = server.HTTPTransport
	default:
		fmt.Fprintf(os.Stderr, "Error: unsupported transport type: %s\n", transportType)
		fmt.Fprintf(os.Stderr, "Supported types: stdio, http\n")
		os.Exit(1)
	}

	// Connect with the appropriate transport
	err := srv.UseTransport(serverTransportType, httpAddr)
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

func runClient(transportType string, httpURL string, message string, timeout time.Duration) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create the echo client
	client := client.NewEchoClient()

	// Configure the transport
	switch transportType {
	case "stdio":
		client.UseStdioTransport()
		fmt.Fprintf(os.Stderr, "Using stdio transport\n")
	case "http":
		client.UseHTTPTransport(httpURL)
		fmt.Fprintf(os.Stderr, "Using HTTP transport with URL: %s\n", httpURL)
	default:
		fmt.Fprintf(os.Stderr, "Error: unsupported transport type: %s\n", transportType)
		fmt.Fprintf(os.Stderr, "Supported types: stdio, http\n")
		os.Exit(1)
	}

	// Send the echo request
	fmt.Fprintf(os.Stderr, "Sending message: %s\n", message)
	response, err := client.Echo(ctx, message)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Print the response
	fmt.Fprintf(os.Stderr, "Response: %s\n", response)
}
