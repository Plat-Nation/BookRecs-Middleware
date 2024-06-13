package log

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestLogAll(t *testing.T) {
	// Save the original stderr
	origStderr := os.Stderr
	defer func() { os.Stderr = origStderr }()

	// Create a pipe to capture stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	loggedHandler := LogAll(handler)

	// Create a sample request
	req := httptest.NewRequest("GET", "http://example.com/foo", bytes.NewBufferString("test body"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	os.Setenv("APP_ENV", "development")
	loggedHandler.ServeHTTP(rr, req)

	// Close the writer end of the pipe and read the captured stderr
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Check the response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the logs
	loggedOutput := buf.String()
	if !bytes.Contains([]byte(loggedOutput), []byte(`"method": "GET"`)) {
		t.Errorf("Expected log to contain method: GET, but it didn't. Log: %s", loggedOutput)
	}
	if !bytes.Contains([]byte(loggedOutput), []byte(`"url": "http://example.com/foo"`)) {
		t.Errorf("Expected log to contain URL: http://example.com/foo, but it didn't. Log: %s", loggedOutput)
	}
	if !bytes.Contains([]byte(loggedOutput), []byte(`"headers": ["Content-Type: application/json"]`)) {
		t.Errorf("Expected log to contain headers, but it didn't. Log: %s", loggedOutput)
	}
	if !bytes.Contains([]byte(loggedOutput), []byte(`"body": "test body"`)) {
		t.Errorf("Expected log to contain body: test body, but it didn't. Log: %s", loggedOutput)
	}
}

func TestLog(t *testing.T) {
	// Save the original stderr
	origStderr := os.Stderr
	defer func() { os.Stderr = origStderr }()

	// Create a pipe to capture stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	os.Setenv("APP_ENV", "development")

	req := httptest.NewRequest("GET", "http://example.com/foo", bytes.NewBufferString("test body"))
	req.Header.Set("Content-Type", "application/json")

	Log(req, "Test log message", zap.String("customField", "customValue"))

	// Close the writer end of the pipe and read the captured stderr
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)

	// Check the logs
	loggedOutput := buf.String()
	if !strings.Contains(loggedOutput, `"method": "GET"`) {
		t.Errorf("Expected log to contain method: GET, but it didn't. Log: %s", loggedOutput)
	}
	if !strings.Contains(loggedOutput, `"url": "http://example.com/foo"`) {
		t.Errorf("Expected log to contain URL: http://example.com/foo, but it didn't. Log: %s", loggedOutput)
	}
	if !strings.Contains(loggedOutput, `"headers": ["Content-Type: application/json"]`) {
		t.Errorf("Expected log to contain headers, but it didn't. Log: %s", loggedOutput)
	}
	if !strings.Contains(loggedOutput, `"body": "test body"`) {
		t.Errorf("Expected log to contain body: test body, but it didn't. Log: %s", loggedOutput)
	}
	if !strings.Contains(loggedOutput, `"customField": "customValue"`) {
		t.Errorf("Expected log to contain customField: customValue, but it didn't. Log: %s", loggedOutput)
	}
}
