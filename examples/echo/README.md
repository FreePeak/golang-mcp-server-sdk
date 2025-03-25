# Echo Example

This example demonstrates how to use the MCP Server SDK with both stdio and SSE (HTTP) transport methods. The example includes:

1. A simple echo tool server
2. A compatible client
3. Support for both transport types

## Building

To build the example:

```bash
cd examples/echo
go build
```

## Running the Server

### Stdio Transport

```bash
./echo -mode=server -transport=stdio
```

### HTTP Transport

```bash
./echo -mode=server -transport=http -http=:8080
```

## Running the Client

### Stdio Transport

This command will run the client in stdio mode, which will connect to a stdio server:

```bash
./echo -mode=client -transport=stdio -message="Hello, MCP!"
```

### HTTP Transport

This command will run the client in HTTP mode, which will connect to the server using HTTP:

```bash
./echo -mode=client -transport=http -http=http://localhost:8080 -message="Hello, MCP!"
```

## Two-way Communication

To demonstrate two-way communication with stdio, you can pipe the client and server together:

```bash
# In terminal 1, start the server
./echo -mode=server -transport=stdio

# In terminal 2, run the client and pipe to/from the server
./echo -mode=client -transport=stdio -message="Hello, MCP!" | ./echo -mode=server -transport=stdio
```

## Testing Both Transport Methods

1. Start an HTTP server:
```bash
./echo -mode=server -transport=http -http=:8080
```

2. In another terminal, connect with the HTTP client:
```bash
./echo -mode=client -transport=http -http=http://localhost:8080 -message="Testing HTTP transport"
```

3. Then test stdio transport:
```bash
./echo -mode=server -transport=stdio & 
./echo -mode=client -transport=stdio -message="Testing stdio transport"
```

## Notes

- The HTTP transport uses the Server-Sent Events (SSE) protocol for server-to-client communication
- The stdio transport uses newline-delimited JSON for message framing
- Both methods implement the JSON-RPC 2.0 protocol for consistent API interactions 