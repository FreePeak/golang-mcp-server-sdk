// Test script for SSE OPTIONS request
const http = require('http');

// Function to test OPTIONS request to /sse endpoint
function testSSEOptions() {
  console.log("Testing OPTIONS request to /sse endpoint...");
  
  const req = http.request({
    hostname: 'localhost',
    port: 8080,
    path: '/sse',
    method: 'OPTIONS',
    headers: {
      'Origin': 'http://example.com',
      'Access-Control-Request-Method': 'GET',
      'Access-Control-Request-Headers': 'Content-Type'
    }
  }, (res) => {
    console.log("SSE OPTIONS Response status:", res.statusCode);
    console.log("SSE OPTIONS Response headers:", res.headers);
    
    // Read any response body
    let data = '';
    res.on('data', (chunk) => {
      data += chunk.toString();
    });
    
    res.on('end', () => {
      if (data) {
        console.log("Response body:", data);
      }
      
      if (res.statusCode === 200) {
        console.log("SUCCESS: OPTIONS request to /sse endpoint returned 200 OK");
        console.log("The Cursor client should now be able to connect properly");
      } else {
        console.error("ERROR: OPTIONS request to /sse endpoint returned", res.statusCode);
        console.error("The Cursor client might still have issues connecting");
      }
    });
  });
  
  req.on('error', (e) => {
    console.error("SSE OPTIONS Error:", e.message);
  });
  
  req.end();
}

// Run the test
testSSEOptions(); 