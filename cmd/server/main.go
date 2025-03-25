package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/infrastructure/server"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/usecases/calculator"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/usecases/docs"
)

func main() {
	// Parse command-line flags
	var (
		httpAddr = flag.String("http", "", "HTTP server address (e.g., :8080)")
		useStdio = flag.Bool("stdio", true, "Use stdio transport")
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

	// Create the MCP server
	srv := server.NewServer("calculator-server", "1.0.0")

	// Add handlers
	srv.WithToolHandler(calculator.NewCalculatorHandler())
	srv.WithResourceHandler(docs.NewDocsHandler())

	// Connect with the appropriate transport
	var err error
	if *useStdio {
		// Use stdio transport
		err = srv.Connect(server.NewStdioTransport())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting transport: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "MCP server running on stdio\n")
	} else if *httpAddr != "" {
		// Use HTTP transport
		err = srv.Connect(server.NewHTTPTransport(*httpAddr))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting transport: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "MCP server running on http://%s\n", *httpAddr)
	} else {
		fmt.Fprintf(os.Stderr, "Error: must specify either -stdio or -http\n")
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
