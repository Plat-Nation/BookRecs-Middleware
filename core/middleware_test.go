package core

import (
	"bytes"
	"crypto/ed25519"
	"crypto/tls"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

/*
Logging tests
*/

func TestLogAll(t *testing.T) {
	os.Setenv("APP_ENV", "development")

	// Save the original stderr
	origStderr := os.Stderr
	defer func() { os.Stderr = origStderr }()

	// Create a pipe to capture stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Initialize middleware
	mw, err := Init(false, true)
	if err != nil {
		panic(err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	loggedHandler := mw.LogAll(handler)

	// Create a sample request
	req := httptest.NewRequest("GET", "http://example.com/foo", bytes.NewBufferString("test body"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
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

	// Normalize the output so tests don't break whether they receive "method: GET" or "method:GET"
	re := regexp.MustCompile(`:\s`)
  normalizedOutput := re.ReplaceAllString(loggedOutput, ":")

	if !bytes.Contains([]byte(normalizedOutput), []byte(`"method":"GET"`)) {
		t.Errorf("Expected log to contain method: GET, but it didn't. Log: %s", normalizedOutput)
	}
	if !bytes.Contains([]byte(normalizedOutput), []byte(`"url":"http://example.com/foo"`)) {
		t.Errorf("Expected log to contain URL: http://example.com/foo, but it didn't. Log: %s", normalizedOutput)
	}
	if !bytes.Contains([]byte(normalizedOutput), []byte(`"headers":["Content-Type:application/json"]`)) {
		t.Errorf("Expected log to contain headers, but it didn't. Log: %s", normalizedOutput)
	}
	if !bytes.Contains([]byte(normalizedOutput), []byte(`"body":"test body"`)) {
		t.Errorf("Expected log to contain body: test body, but it didn't. Log: %s", normalizedOutput)
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

	// Initialize middleware
	mw, err := Init(false, true)
	if err != nil {
		panic(err)
	}

	req := httptest.NewRequest("GET", "http://example.com/foo", bytes.NewBufferString("test body"))
	req.Header.Set("Content-Type", "application/json")

	mw.Log(req, "Test log message", zap.String("customField", "customValue"))
	mw.Logger.Sync()

	// Close the writer end of the pipe and read the captured stderr
	w.Close()
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		t.Fatal(err)
	}

	// Check the logs
	loggedOutput := buf.String()
	t.Log(buf.String())

	// Normalize the output so tests don't break whether they receive "method: GET" or "method:GET"
	re := regexp.MustCompile(`:\s`)
  normalizedOutput := re.ReplaceAllString(loggedOutput, ":")

	if !strings.Contains(normalizedOutput, `"method":"GET"`) {
		t.Errorf("Expected log to contain method: GET, but it didn't. Log: %s", normalizedOutput)
	}
	if !strings.Contains(normalizedOutput, `"url":"http://example.com/foo"`) {
		t.Errorf("Expected log to contain URL: http://example.com/foo, but it didn't. Log: %s", normalizedOutput)
	}
	if !strings.Contains(normalizedOutput, `"headers":["Content-Type:application/json"]`) {
		t.Errorf("Expected log to contain headers, but it didn't. Log: %s", normalizedOutput)
	}
	if !strings.Contains(normalizedOutput, `"body":"test body"`) {
		t.Errorf("Expected log to contain body: test body, but it didn't. Log: %s", normalizedOutput)
	}
	if !strings.Contains(normalizedOutput, `"customField":"customValue"`) {
		t.Errorf("Expected log to contain customField: customValue, but it didn't. Log: %s", normalizedOutput)
	}
}

/*
Auth tests
*/

var jwtKey ed25519.PrivateKey

func init() {
	privateKeyHex := os.Getenv("JWT_KEY")
	if privateKeyHex == "" {
		panic("invalid private key, please input the private key as the JWT_KEY environment variable")
	}
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
			panic("invalid private key")
	}
	jwtKey = ed25519.PrivateKey(privateKeyBytes)
}

// Helper function to create a valid JWT for testing
func createValidJWT() string {
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, jwt.RegisteredClaims{})

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
			panic(err)
	}

	return tokenString
}

// Test the Auth middleware
func TestAuthMiddleware(t *testing.T) {
	// Initialize middleware
	mw, err := Init(false, true)
	if err != nil {
		panic(err)
	}

	validToken := createValidJWT()

	tests := []struct {
			name          string
			authHeader    string
			expectedCode  int
			expectedBody  string
	}{
			{"No Auth Header", "", http.StatusForbidden, "Forbidden\n"},
			{"Invalid Token", "Bearer invalidtoken", http.StatusForbidden, "Forbidden\n"},
			{"Valid Token", "Bearer " + validToken, http.StatusOK, "Hello, World!"},
	}

	for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
					req := httptest.NewRequest("GET", "/", nil)
					req.Header.Set("Authorization", test.authHeader)
					rr := httptest.NewRecorder()

					handler := mw.Auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							w.Write([]byte("Hello, World!"))
					}))

					handler.ServeHTTP(rr, req)

					if status := rr.Code; status != test.expectedCode {
							t.Errorf("handler returned wrong status code: got %v want %v", status, test.expectedCode)
					}

					if body := rr.Body.String(); body != test.expectedBody {
							t.Errorf("handler returned unexpected body: got %v want %v", body, test.expectedBody)
					}
			})
	}
}

func TestHttpsMiddleware(t *testing.T) {
	// Initialize middleware
	mw, err := Init(false, true)
	if err != nil {
		panic(err)
	}

	rawHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := mw.Https(rawHandler)

	tests := []struct {
		name         string
		tls          *tls.ConnectionState
		header       http.Header
		expectedCode int
		expectedURL  string
	}{
			{
					name:         "Insecure Request",
					tls:          nil,
					header:       http.Header{"X-Forwarded-Proto": []string{"http"}},
					expectedCode: http.StatusPermanentRedirect,
					expectedURL:  "https://example.com/foo",
			},
			{
					name:         "Secure Request with TLS",
					tls:          &tls.ConnectionState{},
					header:       http.Header{},
					expectedCode: http.StatusOK,
			},
			{
					name:         "Secure Request with Header",
					tls:          nil,
					header:       http.Header{"X-Forwarded-Proto": []string{"https"}},
					expectedCode: http.StatusOK,
			},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
				// Create an http.Request object
				req := httptest.NewRequest("GET", "http://example.com/foo", nil)
				req.Header = test.header
				req.TLS = test.tls

				// Use httptest.ResponseRecorder to record the response
				w := httptest.NewRecorder()

				// Serve the request to the middleware
				handler.ServeHTTP(w, req)

				// Check if the status code is what we expect
				res := w.Result()
				if res.StatusCode != test.expectedCode {
						t.Errorf("Expected status code %d, got %d", test.expectedCode, res.StatusCode)
				}

				// If a redirect is expected, check the URL
				if test.expectedCode == http.StatusPermanentRedirect {
						location, err := res.Location()
						if err != nil {
								t.Fatal("Expected a location header for redirect, but none was found")
						}
						if location.String() != test.expectedURL {
								t.Errorf("Expected redirect URL %s, got %s", test.expectedURL, location.String())
						}
				}
		})
}
}