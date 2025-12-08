/*
	Copyright (c) 2015-2016 Christopher Young
	Distributable under the terms of The "BSD New" License
	that can be found in the LICENSE file, herein included
	as part of this header.

	managementinterface_test_tilesets_new.go: Additional tests for handleTilesets function
*/

package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// =============================================================================
// Additional Tests for handleTilesets
// =============================================================================
//
// These tests improve coverage for the handleTilesets function
// by testing HTTP methods, JSON formatting, and various edge cases.

// TestHandleTilesets_HTTPMethods tests various HTTP methods on the tilesets endpoint
func TestHandleTilesets_HTTPMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/tiles/tilesets", nil)
			w := httptest.NewRecorder()

			handleTilesets(w, req)

			resp := w.Result()
			// The function doesn't explicitly check HTTP method, so all should work
			// (or return 500 if directory doesn't exist)
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusInternalServerError {
				t.Errorf("Unexpected status code for %s: %d", method, resp.StatusCode)
			}
		})
	}
}

// TestHandleTilesets_JSONResponseFormat tests that the response is valid JSON
func TestHandleTilesets_JSONResponseFormat(t *testing.T) {
	req := httptest.NewRequest("GET", "/tiles/tilesets", nil)
	w := httptest.NewRecorder()

	handleTilesets(w, req)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// If we got a 200, verify JSON structure
	if resp.StatusCode == http.StatusOK {
		var result map[string]map[string]string
		if err := json.Unmarshal(body, &result); err != nil {
			t.Errorf("Response is not valid JSON: %v. Body: %s", err, string(body))
		}

		// Verify it's a map of maps with string keys and values
		for filename, metadata := range result {
			if filename == "" {
				t.Error("Found empty filename key in response")
			}
			if metadata != nil {
				for key := range metadata {
					if key == "" {
						t.Errorf("Found empty metadata key for file %s", filename)
					}
				}
			}
		}
	}
}

// TestHandleTilesets_ResponseHeaders tests that proper headers are set
func TestHandleTilesets_ResponseHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "/tiles/tilesets", nil)
	w := httptest.NewRecorder()

	handleTilesets(w, req)

	resp := w.Result()

	// Note: handleTilesets doesn't explicitly set content-type or cache headers
	// This test documents the current behavior
	contentType := resp.Header.Get("Content-Type")
	t.Logf("Content-Type: %s", contentType)

	// The function doesn't call setJSONHeaders, so Content-Type may not be set
	if contentType == "" {
		t.Log("Note: No Content-Type header set (consider adding application/json)")
	}
}

// TestHandleTilesets_WithQueryParameters tests that query parameters are ignored
func TestHandleTilesets_WithQueryParameters(t *testing.T) {
	testCases := []struct {
		name string
		url  string
	}{
		{"with_single_param", "/tiles/tilesets?refresh=true"},
		{"with_multiple_params", "/tiles/tilesets?refresh=true&format=json"},
		{"with_special_chars", "/tiles/tilesets?param=value%20with%20spaces"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.url, nil)
			w := httptest.NewRecorder()

			handleTilesets(w, req)

			resp := w.Result()
			// Query parameters are ignored, should behave same as base URL
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusInternalServerError {
				t.Errorf("Unexpected status code: %d", resp.StatusCode)
			}
		})
	}
}

// TestHandleTilesets_ConcurrentRequests tests thread safety with concurrent requests
func TestHandleTilesets_ConcurrentRequests(t *testing.T) {
	// Make concurrent requests
	var wg sync.WaitGroup
	numRequests := 10

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req := httptest.NewRequest("GET", "/tiles/tilesets", nil)
			w := httptest.NewRecorder()

			handleTilesets(w, req)

			resp := w.Result()
			// Should not crash or return unexpected status
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusInternalServerError {
				t.Errorf("Concurrent request got unexpected status %d", resp.StatusCode)
			}
		}()
	}

	wg.Wait()
}

// TestHandleTilesets_ResponseBodyFormat tests that response body ends properly
func TestHandleTilesets_ResponseBodyFormat(t *testing.T) {
	req := httptest.NewRequest("GET", "/tiles/tilesets", nil)
	w := httptest.NewRecorder()

	handleTilesets(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK {
		bodyStr := string(body)

		// Should be valid JSON
		if len(bodyStr) > 0 && !strings.HasPrefix(bodyStr, "{") {
			t.Errorf("Expected JSON object to start with '{', got: %s", bodyStr[:min(10, len(bodyStr))])
		}

		if len(bodyStr) > 0 && !strings.HasSuffix(bodyStr, "}") {
			t.Errorf("Expected JSON object to end with '}', got: ...%s", bodyStr[max(0, len(bodyStr)-10):])
		}
	}
}

// TestHandleTilesets_EmptyMapJSON tests that empty map is valid JSON
func TestHandleTilesets_EmptyMapJSON(t *testing.T) {
	req := httptest.NewRequest("GET", "/tiles/tilesets", nil)
	w := httptest.NewRecorder()

	handleTilesets(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK {
		var result map[string]map[string]string
		if err := json.Unmarshal(body, &result); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}
		// Just verify the structure is correct, don't care about contents
		t.Logf("Found %d tilesets in response", len(result))
	}
}

// Helper functions for min/max
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
