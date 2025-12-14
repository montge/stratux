package main

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestHandleDownloadAHRSLogsRequest_EdgeCasesForCoverage tests additional edge cases
// to improve coverage of handleDownloadAHRSLogsRequest from 72.2% to higher
func TestHandleDownloadAHRSLogsRequest_EdgeCasesForCoverage(t *testing.T) {
	// Save original varLogDirPath
	origVarLogDirPath := varLogDirPath
	defer func() {
		varLogDirPath = origVarLogDirPath
	}()

	t.Run("successful_zip_with_io_copy", func(t *testing.T) {
		// Create temp directory with test files
		tmpDir, err := os.MkdirTemp("", "ahrs_download_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Set varLogDirPath to our temp directory
		varLogDirPath = tmpDir

		// Create test files with actual content to ensure io.Copy is tested
		testFiles := map[string]string{
			"sensors_test1.csv": "timestamp,pitch,roll,heading\n1234567890,10.5,5.2,180.0\n",
			"sensors_test2.csv": "timestamp,pitch,roll,heading\n1234567891,11.5,6.2,181.0\n",
			"stratux.log":       "2025/12/14 Test log entry\nAnother log line\n",
		}

		for filename, content := range testFiles {
			err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file %s: %v", filename, err)
			}
		}

		// Make request
		req := httptest.NewRequest("GET", "/downloadahrslogs", nil)
		w := httptest.NewRecorder()

		handleDownloadAHRSLogsRequest(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
		}

		// Verify headers
		contentType := w.Header().Get("Content-Type")
		if contentType != "application/zip" {
			t.Errorf("Expected Content-Type 'application/zip', got %s", contentType)
		}

		contentDisp := w.Header().Get("Content-Disposition")
		if contentDisp != "attachment; filename=ahrs_logs.zip" {
			t.Errorf("Expected Content-Disposition 'attachment; filename=ahrs_logs.zip', got %s", contentDisp)
		}

		// Verify the zip contains our files
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		if len(body) == 0 {
			t.Error("Expected non-empty zip file")
			return
		}

		// Read the zip file
		zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		if err != nil {
			t.Fatalf("Failed to read zip file: %v", err)
		}

		// Verify all files are in the zip
		foundFiles := make(map[string]bool)
		for _, file := range zipReader.File {
			foundFiles[file.Name] = true

			// Verify file content
			rc, err := file.Open()
			if err != nil {
				t.Errorf("Failed to open file %s in zip: %v", file.Name, err)
				continue
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Errorf("Failed to read file %s in zip: %v", file.Name, err)
				continue
			}

			expectedContent := testFiles[file.Name]
			if string(content) != expectedContent {
				t.Errorf("Content mismatch for %s. Expected %q, got %q", file.Name, expectedContent, string(content))
			}
		}

		// Check all expected files are present
		for filename := range testFiles {
			if !foundFiles[filename] {
				t.Errorf("Expected file %s not found in zip", filename)
			}
		}

		t.Logf("Successfully created zip with %d files", len(foundFiles))
	})

	t.Run("empty_directory", func(t *testing.T) {
		// Create empty temp directory
		tmpDir, err := os.MkdirTemp("", "ahrs_download_empty_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		varLogDirPath = tmpDir

		req := httptest.NewRequest("GET", "/downloadahrslogs", nil)
		w := httptest.NewRecorder()

		handleDownloadAHRSLogsRequest(w, req)

		resp := w.Result()
		// Should succeed with empty zip
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for empty directory, got %d", resp.StatusCode)
		}
	})

	t.Run("directory_with_non_matching_files", func(t *testing.T) {
		// Create temp directory with files that don't match the pattern
		tmpDir, err := os.MkdirTemp("", "ahrs_download_nomatch_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		varLogDirPath = tmpDir

		// Create files that don't match sensors_*.csv or stratux.log
		err = os.WriteFile(filepath.Join(tmpDir, "other.log"), []byte("other"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		req := httptest.NewRequest("GET", "/downloadahrslogs", nil)
		w := httptest.NewRecorder()

		handleDownloadAHRSLogsRequest(w, req)

		resp := w.Result()
		// Should succeed with empty zip (no matching files)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("nonexistent_directory", func(t *testing.T) {
		// Set to a directory that doesn't exist
		varLogDirPath = "/nonexistent/directory/path/that/does/not/exist"

		req := httptest.NewRequest("GET", "/downloadahrslogs", nil)
		w := httptest.NewRecorder()

		handleDownloadAHRSLogsRequest(w, req)

		resp := w.Result()
		// Should return 404 error due to ReadDir failure
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if len(bodyStr) == 0 {
			t.Error("Expected error message in response body")
		}
		t.Logf("Error response: %s", bodyStr)
	})
}
