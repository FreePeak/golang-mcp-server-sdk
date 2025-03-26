# Fix for MCP Cursor Client Connection Error

This document explains how to resolve the SSE connection error with Cursor client:

```
2025-03-26 02:08:05.901 [info] rver: Creating SSE transport
2025-03-26 02:08:05.904 [error] rver: Client error for command SSE error: Non-200 status code (405)
2025-03-26 02:08:05.904 [error] rver: Error in MCP: SSE error: Non-200 status code (405)
2025-03-26 02:08:05.904 [error] rver: Failed to reload client: SSE error: Non-200 status code (405)
2025-03-26 02:08:05.916 [info] rver: Handling ListOfferings action
2025-03-26 02:08:05.916 [error] rver: No server info found
```

## Problem Identified

The Cursor client is receiving a 405 (Method Not Allowed) error when attempting to connect to the SSE endpoint. Our tests confirmed:

1. HTTP POST requests to the main API endpoint work correctly
2. The SSE endpoint at `/sse` works correctly for direct HTTP GET requests
3. The issue is with CORS preflight handling - the OPTIONS request to `/sse` returns 405 Method Not Allowed
4. However, OPTIONS requests to the root path `/` work correctly

It appears the Cursor client is making an OPTIONS request specifically to the `/sse` path, but our server implementation only handles OPTIONS at the root path.

## Recommended Solutions

### Option 1: Configure Cursor to Use STDIO Transport

The simplest solution is to use STDIO transport instead of HTTP transport for the MCP server connection in Cursor. This avoids the CORS preflight issue entirely:

```
./run_stdio_server.sh
```

In Cursor, you would configure the MCP server connection to use STDIO transport rather than HTTP/SSE.

### Option 2: Use a Proxy Server

You can set up a proxy server (e.g., nginx) in front of your MCP server that correctly handles OPTIONS requests for all paths:

```nginx
server {
    listen 8080;
    
    location / {
        if ($request_method = 'OPTIONS') {
            add_header 'Access-Control-Allow-Origin' '*';
            add_header 'Access-Control-Allow-Methods' 'GET, POST, OPTIONS';
            add_header 'Access-Control-Allow-Headers' 'Content-Type';
            add_header 'Access-Control-Allow-Credentials' 'true';
            return 200;
        }
        
        proxy_pass http://localhost:8081; # Your actual MCP server
        proxy_http_version 1.1;
        proxy_set_header Connection "";
    }
}
```

### Option 3: Modify Cursor's MCP Client Implementation

If you have access to Cursor's source code, modify its MCP client implementation to:

1. Either send OPTIONS requests to the root path `/` instead of `/sse`
2. Or skip the OPTIONS preflight request entirely and directly make the GET request to `/sse`

### Option 4: Use WebSockets Instead of SSE

SSE inherently requires a preflight OPTIONS request which is causing this issue. WebSockets might be more compatible with Cursor's client implementation.

## Further Troubleshooting

We attempted multiple approaches to fix the server-side handling of OPTIONS requests:

1. Added CORS headers and OPTIONS handling in the handleSSE function
2. Centralized routing with OPTIONS handling in the root handler
3. Created a completely new implementation (FixedHTTPTransport)

Despite these attempts, the server continues to return 405 for OPTIONS requests to `/sse`. This suggests that either:

1. There's a deeper issue in the server's HTTP handling that we haven't identified
2. Cursor's client implementation is making the OPTIONS request in a way that our server can't handle
3. There might be a network or proxy issue between Cursor and our server

## Next Steps

1. Try using STDIO transport as the most reliable solution
2. Reach out to the Cursor team for assistance with their MCP client implementation
3. Consider using other MCP server implementations that might be more compatible with Cursor

You can test the server's API functionality using these Node.js scripts:
- `test_simple.js` - Tests basic JSON-RPC functionality
- `test_sse.js` - Tests the SSE endpoint for GET requests
- `test_sse_options.js` - Tests the SSE endpoint for OPTIONS requests 