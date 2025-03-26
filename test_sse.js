// Test script for SSE endpoint
const http = require('http');

// Function to test the SSE endpoint
function testSSE() {
  console.log("Testing SSE endpoint...");
  
  const req = http.request({
    hostname: 'localhost',
    port: 8080,
    path: '/sse',
    method: 'GET',
    headers: {
      'Accept': 'text/event-stream'
    }
  }, (res) => {
    console.log("SSE Response status:", res.statusCode);
    console.log("SSE Response headers:", res.headers);
    
    if (res.statusCode !== 200) {
      console.error(`Error: Unexpected status code ${res.statusCode}`);
      return;
    }
    
    res.on('data', (chunk) => {
      const data = chunk.toString();
      console.log("SSE Data received:", data);
    });
    
    res.on('end', () => {
      console.log("SSE connection closed");
    });
  });
  
  req.on('error', (e) => {
    console.error("SSE Error:", e.message);
  });
  
  req.end();
}

// Function to make OPTIONS request (CORS preflight)
function testCORS() {
  console.log("Testing CORS preflight...");
  
  const req = http.request({
    hostname: 'localhost',
    port: 8080,
    path: '/',  // Test the root path
    method: 'OPTIONS',
    headers: {
      'Origin': 'http://example.com',
      'Access-Control-Request-Method': 'GET'
    }
  }, (res) => {
    console.log("CORS Response status:", res.statusCode);
    console.log("CORS Response headers:", res.headers);
  });
  
  req.on('error', (e) => {
    console.error("CORS Error:", e.message);
  });
  
  req.end();
}

// Run tests
testCORS();
setTimeout(() => {
  testSSE();
}, 1000); 