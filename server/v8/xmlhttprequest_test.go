package v8_test

import (
	v8 "github.com/lizc2003/vue-ssr-v8go/server/v8"
	"testing"
	"time"
)

func TestXhr(t *testing.T) {
	vmMgr, err := v8.NewVmMgr("dev", nil, &v8.VmConfig{}, &v8.XhrConfig{})
	if err != nil {
		t.Fatalf("create vm mgr err: %v", err)
		return
	}

	err, _ = vmMgr.Execute(testXhrJsContent, "test.js")
	if err != nil {
		t.Fatalf("test fail: %v", err)
		return
	}

	time.Sleep(10 * time.Second)
}

const testXhrJsContent = `
function assert(condition, message) {
  if (!condition) {
    throw new Error(message || "Assertion failed");
  }
  console.log(message);
}

function runTestSuite(name, tests) {
  console.log('Running test suite: ' + name + '.\n');
  try {
    tests();
    console.log('All tests passed in ' + name + '.\n');
  } catch (error) {
    console.error('Test failed: ' + error.message);
  }
}

runTestSuite("Sync Send", function() {
  const testUrl = "https://jsonplaceholder.typicode.com/posts/1";
  const xhr = new XMLHttpRequest();

  xhr.open("GET", testUrl, false);
  xhr.send();
});

runTestSuite("Basic GET Request", function() {
  const testUrl = "https://jsonplaceholder.typicode.com/posts/1";
  const xhr = new XMLHttpRequest();

  xhr.open("GET", testUrl);
  xhr.onload = function() {
    assert(xhr.status === 200, "Status should be 200");
    const response = JSON.parse(xhr.responseText);
    assert(response.id === 1, "Response should contain id=1");
    assert(typeof response.title === "string", "Response should contain title");
  };
  
  xhr.onerror = function() {
    assert(false, "Request should not fail");
  };
  
  xhr.send();
});

runTestSuite("POST Request", function() {
  const testUrl = "https://jsonplaceholder.typicode.com/posts";
  const xhr = new XMLHttpRequest();

  const postData = JSON.stringify({
    title: "foo",
    body: "bar",
    userId: 1
  });
  
  xhr.open("POST", testUrl);
  xhr.setRequestHeader("Content-Type", "application/json");
  
  xhr.onload = function() {
    assert(xhr.status === 201, "Status should be 201 for created resource");
    const response = JSON.parse(xhr.responseText);
    assert(response.id === 101, "Response should contain new id");
    assert(response.title === "foo", "Response should contain posted title");
  };
  
  xhr.onerror = function() {
    assert(false, "Request should not fail");
  };
  
  xhr.send(postData);
});

runTestSuite("Error Handling", function() {
  const invalidUrl = "https://invalid.url.that.does.not.exist";
  const xhr = new XMLHttpRequest();
  
  xhr.open("GET", invalidUrl);

  xhr.onload = function() {
    assert(false, "Request should not succeed");
  };
  
  xhr.onerror = function() {
    assert(true, "Error handler should be called for invalid URL");
  };
  
  xhr.send();
});

runTestSuite("Progress Events", function() {
  const testUrl = "https://www.baidu.com/";
  const xhr = new XMLHttpRequest();

  let progressEvents = 0;
  
  xhr.open("GET", testUrl);
  
  xhr.onprogress = function(event) {
    progressEvents++;
    assert(event.lengthComputable, "Length should be computable");
    assert(event.loaded <= event.total, "Loaded should not exceed total");
  };
  
  xhr.onload = function() {
    assert(progressEvents > 0, "Should receive at least one progress event");
    assert(xhr.status === 200, "Status should be 200");
  };
  
  xhr.send();
});

runTestSuite("Request Abort", function() {
  const testUrl = "https://jsonplaceholder.typicode.com/posts";
  const xhr = new XMLHttpRequest();
  let abortEventFired = false;
  
  xhr.open("GET", testUrl);
  
  xhr.onabort = function() {
    abortEventFired = true;
    assert(true, "Abort handler should be called");
  };
  
  xhr.onload = function() {
    assert(false, "Request should not complete after abort");
  };
  
  xhr.send();
  xhr.abort();
});

runTestSuite("Ready State Changes", function() {
  const testUrl = "https://jsonplaceholder.typicode.com/posts/1";
  const xhr = new XMLHttpRequest();
  const states = [];
  
  xhr.onreadystatechange = function() {
    states.push(xhr.readyState);
    
    if (xhr.readyState === XMLHttpRequest.DONE) {
      assert(
        states.includes(XMLHttpRequest.OPENED) && 
        states.includes(XMLHttpRequest.HEADERS_RECEIVED) &&
        states.includes(XMLHttpRequest.LOADING),
        "Should go through all ready states"
      );
    }
  };
  
  xhr.open("GET", testUrl);
  xhr.send();
});

runTestSuite("Request Headers", function() {
  const testUrl = "https://httpbin.org/headers";
  const xhr = new XMLHttpRequest();
  const customHeaderValue = "custom-value";
  
  xhr.open("GET", testUrl, true);
  xhr.setRequestHeader("X-Custom-Header", customHeaderValue);
  
  xhr.onload = function() {
    const response = JSON.parse(xhr.responseText);
    assert(
      response.headers["X-Custom-Header"] === customHeaderValue,
      "Custom header should be included in request"
    );
  };
  
  xhr.send();
});

runTestSuite("Comprehensive Test", function() {
  const testUrl = "https://jsonplaceholder.typicode.com/posts";
  const xhr = new XMLHttpRequest();
  let progressCount = 0;
  let stateChanges = 0;
  
  xhr.onreadystatechange = function() {
    stateChanges++;
  };
  
  xhr.onprogress = function(event) {
    progressCount++;
  };
  
  xhr.onload = function() {
    assert(xhr.status === 200, "Status should be 200");
    const response = JSON.parse(xhr.responseText);
    assert(Array.isArray(response), "Response should be an array");
    assert(response.length > 0, "Response array should not be empty");
    
    assert(stateChanges >= 4, "Should have at least 4 state changes");
    assert(progressCount > 0, "Should have at least one progress event");
  };
  
  xhr.open("GET", testUrl);
  xhr.send();
});

runTestSuite("Response Headers - Basic", function() {
  const testUrl = "https://httpbin.org/headers";
  const xhr = new XMLHttpRequest();

  xhr.open("GET", testUrl);

  xhr.onload = function() {
    assert(xhr.status === 200, "Status should be 200");
    
    const contentType = xhr.getResponseHeader("Content-Type");
    assert(
      contentType && contentType.includes("application/json"),
      "Content-Type should be application/json"
    );
    
    const server = xhr.getResponseHeader("Server");
    assert(server, "Server header should exist");

    const invalidHeader = xhr.getResponseHeader("X-This-Header-Does-Not-Exist");
    assert(invalidHeader === null, "Non-existent header should return null");
    
    const allHeaders = xhr.getAllResponseHeaders();
    assert(
      allHeaders.includes("Content-Type:") && allHeaders.includes("Server:"),
      "All headers should include Content-Type and Server"
    );
  };

  xhr.send();
});
`
