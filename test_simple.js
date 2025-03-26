// Simple Node.js script to test MCP server functionality - no SSE
const http = require('http');

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
      path: '/',  // Root path for all JSON-RPC requests
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
    
    // Step 4: Call a tool if available
    if (toolsResponse && toolsResponse.result && toolsResponse.result.tools && toolsResponse.result.tools.length > 0) {
      console.log("\n=== STEP 4: CALL A TOOL ===");
      console.log("Available tools:", toolsResponse.result.tools.map(t => t.name).join(", "));
      
      // Pick the first tool
      const firstTool = toolsResponse.result.tools[0];
      if (firstTool) {
        console.log(`Using tool: ${firstTool.name}`);
        console.log("Tool schema:", JSON.stringify(firstTool.inputSchema, null, 2));
        
        // Construct arguments based on tool schema
        const sampleArgs = {};
        if (firstTool.inputSchema && firstTool.inputSchema.properties) {
          Object.keys(firstTool.inputSchema.properties).forEach(key => {
            const prop = firstTool.inputSchema.properties[key];
            // Add sample values based on type
            if (prop.type === 'string') {
              sampleArgs[key] = "test_value";
            } else if (prop.type === 'number') {
              sampleArgs[key] = 42;
            } else if (prop.type === 'boolean') {
              sampleArgs[key] = true;
            }
          });
        }
        
        console.log("Calling with arguments:", sampleArgs);
        const callResponse = await sendJSONRPC('tools/call', {
          name: firstTool.name,
          arguments: sampleArgs
        }, 'call-1');
        
        if (callResponse && callResponse.result) {
          console.log(`${firstTool.name} tool result:`, callResponse.result);
        }
      } else {
        console.log("No tools available to call");
      }
    } else {
      console.log("No tools available or invalid response from tools/list");
    }
    
    console.log("\nAll tests completed successfully!");
  } catch (error) {
    console.error("Test sequence failed:", error);
  }
}

// Execute the tests
runTests(); 