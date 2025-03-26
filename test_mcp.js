// Simple Node.js script to test MCP server functionality
const http = require('http');
const https = require('https');
const EventSourceModule = require('eventsource');
const EventSource = EventSourceModule.default || EventSourceModule;

// Global variable to store the SSE connection
let eventSource = null;

// Set up SSE connection
function setupSSE() {
  return new Promise((resolve, reject) => {
    // Close existing connection if any
    if (eventSource) {
      eventSource.close();
    }

    console.log("Setting up SSE connection to http://localhost:8080/sse");
    eventSource = new EventSource('http://localhost:8080/sse');

    eventSource.onopen = () => {
      console.log("SSE connection established");
      resolve(eventSource);
    };

    eventSource.onerror = (error) => {
      console.error("SSE connection error:", error);
      reject(error);
    };

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        console.log("SSE message received:", JSON.stringify(data, null, 2));
      } catch (e) {
        console.log("Raw SSE message:", event.data);
      }
    };
  });
}

// Send a JSON-RPC request
function sendJSONRPC(method, params, id = null) {
  return new Promise((resolve, reject) => {
    // Build the request object
    const request = {
      jsonrpc: '2.0',
      method: method
    };
    
    if (id) {
      request.id = id;
    }
    
    if (params) {
      request.params = params;
    }
    
    const data = JSON.stringify(request);
    
    const options = {
      hostname: 'localhost',
      port: 8080,
      path: '/',  // Root path for all JSON-RPC requests as per implementation
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': data.length
      }
    };
    
    const req = http.request(options, (res) => {
      let responseData = '';
      
      res.on('data', (chunk) => {
        responseData += chunk;
      });
      
      res.on('end', () => {
        console.log(`[${method}] Server response status:`, res.statusCode);
        
        // For requests with IDs (not notifications)
        if (id !== null && responseData) {
          try {
            const response = JSON.parse(responseData);
            console.log(`[${method}] Response:`, JSON.stringify(response, null, 2));
            resolve(response);
          } catch (e) {
            console.error(`[${method}] Error parsing response:`, e);
            console.log('Raw response:', responseData);
            reject(e);
          }
        } else {
          // For notifications, just indicate success
          console.log(`[${method}] Notification sent successfully`);
          resolve();
        }
      });
    });
    
    req.on('error', (e) => {
      console.error(`[${method}] Request error:`, e.message);
      reject(e);
    });
    
    console.log(`[${method}] Sending request:`, data);
    req.write(data);
    req.end();
  });
}

// Main test function
async function runTests() {
  try {
    // Set up SSE connection first
    console.log("\n=== SETTING UP SSE CONNECTION ===");
    await setupSSE();
    
    // Wait a moment for the SSE connection to stabilize
    await new Promise(resolve => setTimeout(resolve, 1000));

    // Step 1: Initialize the server
    console.log("\n=== STEP 1: INITIALIZE SERVER ===");
    const initResponse = await sendJSONRPC('initialize', {
      capabilities: {
        tools: {}
      }
    }, 'init-1');
    
    // Wait a moment to let the server process
    await new Promise(resolve => setTimeout(resolve, 500));
    
    // Step 2: Send initialized notification
    console.log("\n=== STEP 2: SEND INITIALIZED NOTIFICATION ===");
    await sendJSONRPC('initialized', {});
    
    // Wait for server processing
    await new Promise(resolve => setTimeout(resolve, 500));
    
    // Step 3: List tools
    console.log("\n=== STEP 3: LIST TOOLS ===");
    const toolsResponse = await sendJSONRPC('tools/list', {}, 'list-1');
    
    // Step 4: Call the calculator tool if available
    if (toolsResponse && toolsResponse.result && toolsResponse.result.tools && toolsResponse.result.tools.length > 0) {
      console.log("\n=== STEP 4: CALL CALCULATOR TOOL ===");
      const calcTool = toolsResponse.result.tools.find(tool => tool.name === 'calculator');
      
      if (calcTool) {
        console.log("Found calculator tool:", calcTool.name);
        const callResponse = await sendJSONRPC('tools/call', {
          name: 'calculator',
          arguments: { a: 5, b: 3, operation: "add" }
        }, 'call-1');
        
        if (callResponse && callResponse.result) {
          console.log("Calculator tool result:", callResponse.result);
        }
      } else {
        console.log("Calculator tool not found in available tools");
        
        // List all available tools
        console.log("Available tools:", 
          toolsResponse.result.tools.map(t => t.name).join(", "));
      }
    } else {
      console.log("No tools available or invalid response from tools/list");
    }
    
    console.log("\nAll tests completed successfully!");
  } catch (error) {
    console.error("Test sequence failed:", error);
  } finally {
    // Close the SSE connection when done
    if (eventSource) {
      eventSource.close();
      console.log("SSE connection closed");
    }
  }
}

// Execute the tests
runTests(); 