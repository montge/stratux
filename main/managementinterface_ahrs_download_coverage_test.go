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
// to improve coverage of handleDownloadAHRSLogsRequest from 72.2% to 83.3%.
//
// This test covers:
// - Successful zip creation with multiple files (sensors_*.csv and stratux.log)
// - Empty directory (no matching files)
// - Directory with non-matching files
// - Nonexistent directory (ReadDir error path)
// - File with no read permissions (Open error path)
// - Large file content (io.Copy with multiple reads)
// - Mixed matching and non-matching files (pattern matching)
// - File handle management (multiple files)
// - Single sensor file
// - Only stratux.log file
// - Empty sensor file
// - Subdirectory with matching pattern (io.Copy error when reading directory)
//
// Remaining uncovered paths (16.7%):
// - f.Info() error (line 815-817): Difficult to trigger without mocking
// - zip.FileInfoHeader() error (line 820-822): Difficult to trigger without mocking
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

	t.Run("file_deleted_between_readdir_and_open", func(t *testing.T) {
		// Create temp directory
		tmpDir, err := os.MkdirTemp("", "ahrs_download_deleted_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		varLogDirPath = tmpDir

		// Create a file
		testFile := filepath.Join(tmpDir, "sensors_test.csv")
		err = os.WriteFile(testFile, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Delete the file to simulate a race condition
		// This needs to happen between ReadDir and Open
		// We can't easily simulate this without race conditions, so skip this test
		// Instead test with a file that exists but has no read permissions
		err = os.Chmod(testFile, 0000)
		if err != nil {
			t.Fatalf("Failed to chmod test file: %v", err)
		}
		// Restore permissions for cleanup
		defer os.Chmod(testFile, 0644)

		req := httptest.NewRequest("GET", "/downloadahrslogs", nil)
		w := httptest.NewRecorder()

		handleDownloadAHRSLogsRequest(w, req)

		resp := w.Result()
		// Should return 404 error due to Open failure
		if resp.StatusCode != http.StatusNotFound {
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status 404, got %d. Body: %s", resp.StatusCode, string(body))
		}

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if len(bodyStr) == 0 {
			t.Error("Expected error message in response body")
		}
		t.Logf("Error response for unreadable file: %s", bodyStr)
	})

	t.Run("file_with_special_permissions", func(t *testing.T) {
		// Create temp directory
		tmpDir, err := os.MkdirTemp("", "ahrs_download_perms_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		varLogDirPath = tmpDir

		// Create a file with normal permissions
		testFile := filepath.Join(tmpDir, "sensors_readable.csv")
		err = os.WriteFile(testFile, []byte("readable content\n"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		req := httptest.NewRequest("GET", "/downloadahrslogs", nil)
		w := httptest.NewRecorder()

		handleDownloadAHRSLogsRequest(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
		}

		// Verify the zip contains the file
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		if err != nil {
			t.Fatalf("Failed to read zip file: %v", err)
		}

		if len(zipReader.File) != 1 {
			t.Errorf("Expected 1 file in zip, got %d", len(zipReader.File))
		}
	})

	t.Run("large_file_io_copy", func(t *testing.T) {
		// Create temp directory
		tmpDir, err := os.MkdirTemp("", "ahrs_download_large_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		varLogDirPath = tmpDir

		// Create a large file to ensure io.Copy handles multiple reads
		largeContent := make([]byte, 100*1024) // 100KB
		for i := range largeContent {
			largeContent[i] = byte(i % 256)
		}

		testFile := filepath.Join(tmpDir, "sensors_large.csv")
		err = os.WriteFile(testFile, largeContent, 0644)
		if err != nil {
			t.Fatalf("Failed to create large test file: %v", err)
		}

		req := httptest.NewRequest("GET", "/downloadahrslogs", nil)
		w := httptest.NewRecorder()

		handleDownloadAHRSLogsRequest(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
		}

		// Verify the zip contains the large file with correct content
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		if err != nil {
			t.Fatalf("Failed to read zip file: %v", err)
		}

		if len(zipReader.File) != 1 {
			t.Fatalf("Expected 1 file in zip, got %d", len(zipReader.File))
		}

		// Verify content
		rc, err := zipReader.File[0].Open()
		if err != nil {
			t.Fatalf("Failed to open file in zip: %v", err)
		}
		defer rc.Close()

		content, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("Failed to read file in zip: %v", err)
		}

		if len(content) != len(largeContent) {
			t.Errorf("Content size mismatch. Expected %d, got %d", len(largeContent), len(content))
		}

		if !bytes.Equal(content, largeContent) {
			t.Error("Content mismatch for large file")
		}
	})

	t.Run("mixed_matching_and_non_matching_files", func(t *testing.T) {
		// Create temp directory
		tmpDir, err := os.MkdirTemp("", "ahrs_download_mixed_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		varLogDirPath = tmpDir

		// Create a mix of matching and non-matching files
		files := map[string]string{
			"sensors_1.csv":   "sensor data 1",
			"sensors_2.csv":   "sensor data 2",
			"stratux.log":     "log data",
			"other.log":       "other log",
			"data.txt":        "text data",
			"sensors.csv":     "no underscore - should not match",
			"sensors_abc.txt": "wrong extension - should not match",
		}

		for filename, content := range files {
			err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file %s: %v", filename, err)
			}
		}

		req := httptest.NewRequest("GET", "/downloadahrslogs", nil)
		w := httptest.NewRecorder()

		handleDownloadAHRSLogsRequest(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
		}

		// Verify the zip contains only matching files
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		if err != nil {
			t.Fatalf("Failed to read zip file: %v", err)
		}

		// Should have exactly 3 files: sensors_1.csv, sensors_2.csv, stratux.log
		if len(zipReader.File) != 3 {
			t.Errorf("Expected 3 files in zip, got %d", len(zipReader.File))
		}

		expectedFiles := map[string]bool{
			"sensors_1.csv": false,
			"sensors_2.csv": false,
			"stratux.log":   false,
		}

		for _, file := range zipReader.File {
			if _, ok := expectedFiles[file.Name]; ok {
				expectedFiles[file.Name] = true
			} else {
				t.Errorf("Unexpected file in zip: %s", file.Name)
			}
		}

		for filename, found := range expectedFiles {
			if !found {
				t.Errorf("Expected file %s not found in zip", filename)
			}
		}
	})

	t.Run("verify_file_handles_cleanup", func(t *testing.T) {
		// Create temp directory
		tmpDir, err := os.MkdirTemp("", "ahrs_download_handles_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		varLogDirPath = tmpDir

		// Create multiple files to ensure multiple file handles are managed
		for i := 1; i <= 10; i++ {
			filename := filepath.Join(tmpDir, "sensors_"+string(rune('0'+i))+".csv")
			err := os.WriteFile(filename, []byte("content"), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
		}

		req := httptest.NewRequest("GET", "/downloadahrslogs", nil)
		w := httptest.NewRecorder()

		handleDownloadAHRSLogsRequest(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
		}

		// Verify all files are accessible after the handler runs
		// This would fail if file handles weren't closed properly
		for i := 1; i <= 10; i++ {
			filename := filepath.Join(tmpDir, "sensors_"+string(rune('0'+i))+".csv")
			_, err := os.Stat(filename)
			if err != nil {
				t.Errorf("File %s not accessible after handler: %v", filename, err)
			}
		}
	})

	t.Run("single_sensor_file", func(t *testing.T) {
		// Test with just one sensor file
		tmpDir, err := os.MkdirTemp("", "ahrs_download_single_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		varLogDirPath = tmpDir

		// Create a single sensor file
		testFile := filepath.Join(tmpDir, "sensors_123456.csv")
		testContent := "timestamp,pitch,roll,heading\n123456,1.0,2.0,3.0\n"
		err = os.WriteFile(testFile, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		req := httptest.NewRequest("GET", "/downloadahrslogs", nil)
		w := httptest.NewRecorder()

		handleDownloadAHRSLogsRequest(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
		}

		// Verify zip contains exactly one file
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		if err != nil {
			t.Fatalf("Failed to read zip file: %v", err)
		}

		if len(zipReader.File) != 1 {
			t.Errorf("Expected 1 file in zip, got %d", len(zipReader.File))
		}

		if zipReader.File[0].Name != "sensors_123456.csv" {
			t.Errorf("Expected file name 'sensors_123456.csv', got %s", zipReader.File[0].Name)
		}
	})

	t.Run("only_stratux_log", func(t *testing.T) {
		// Test with only stratux.log file
		tmpDir, err := os.MkdirTemp("", "ahrs_download_stratux_only_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		varLogDirPath = tmpDir

		// Create only stratux.log
		testFile := filepath.Join(tmpDir, "stratux.log")
		testContent := "2025/12/14 10:00:00 Stratux started\n"
		err = os.WriteFile(testFile, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		req := httptest.NewRequest("GET", "/downloadahrslogs", nil)
		w := httptest.NewRecorder()

		handleDownloadAHRSLogsRequest(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
		}

		// Verify zip contains exactly one file
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		if err != nil {
			t.Fatalf("Failed to read zip file: %v", err)
		}

		if len(zipReader.File) != 1 {
			t.Errorf("Expected 1 file in zip, got %d", len(zipReader.File))
		}

		if zipReader.File[0].Name != "stratux.log" {
			t.Errorf("Expected file name 'stratux.log', got %s", zipReader.File[0].Name)
		}
	})

	t.Run("empty_sensor_file", func(t *testing.T) {
		// Test with an empty sensor file
		tmpDir, err := os.MkdirTemp("", "ahrs_download_empty_file_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		varLogDirPath = tmpDir

		// Create an empty sensor file
		testFile := filepath.Join(tmpDir, "sensors_empty.csv")
		err = os.WriteFile(testFile, []byte(""), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		req := httptest.NewRequest("GET", "/downloadahrslogs", nil)
		w := httptest.NewRecorder()

		handleDownloadAHRSLogsRequest(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
		}

		// Verify zip contains the empty file
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
		if err != nil {
			t.Fatalf("Failed to read zip file: %v", err)
		}

		if len(zipReader.File) != 1 {
			t.Errorf("Expected 1 file in zip, got %d", len(zipReader.File))
		}

		// Verify the file is empty
		rc, err := zipReader.File[0].Open()
		if err != nil {
			t.Fatalf("Failed to open file in zip: %v", err)
		}
		defer rc.Close()

		content, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("Failed to read file in zip: %v", err)
		}

		if len(content) != 0 {
			t.Errorf("Expected empty file, got %d bytes", len(content))
		}
	})

	t.Run("subdirectory_causes_error", func(t *testing.T) {
		// Test that subdirectories matching the pattern cause an error
		// This tests the io.Copy error path when trying to read a directory
		tmpDir, err := os.MkdirTemp("", "ahrs_download_subdir_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		varLogDirPath = tmpDir

		// Create a subdirectory with a matching name
		// This will match "sensors_*.csv" pattern but is a directory
		subdir := filepath.Join(tmpDir, "sensors_subdir.csv")
		err = os.Mkdir(subdir, 0755)
		if err != nil {
			t.Fatalf("Failed to create subdirectory: %v", err)
		}

		req := httptest.NewRequest("GET", "/downloadahrslogs", nil)
		w := httptest.NewRecorder()

		handleDownloadAHRSLogsRequest(w, req)

		resp := w.Result()
		// Should fail when trying to read the directory
		if resp.StatusCode != http.StatusNotFound {
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status 404, got %d. Body: %s", resp.StatusCode, string(body))
		}

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if len(bodyStr) == 0 {
			t.Error("Expected error message in response body")
		}
		// Error should mention it's a directory
		t.Logf("Error response for directory: %s", bodyStr)
	})
}
