package auth

import (
	"crypto/ed25519"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	log "github.com/Plat-Nation/BookRecs-Middleware/log"
	"github.com/golang-jwt/jwt/v4"
)

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

// Test the verify function
func TestVerify(t *testing.T) {
    validToken := createValidJWT()
    invalidToken := "invalidtoken"

    tests := []struct {
        name       string
        authHeader string
        expected       bool
    }{
        {"No Auth Header", "", false},
        {"Invalid Token", "Bearer " + invalidToken, false},
        {"Valid Token", "Bearer " + validToken, true},
        {"No Bearer Prefix", validToken, false},
    }

    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            if got := verify(test.authHeader); got != test.expected {
                t.Errorf("verify() = %v, expected %v", got, test.expected)
            }
        })
    }
}

// Test the Auth middleware
func TestAuthMiddleware(t *testing.T) {
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

						logger, err := log.Init()
						if err != nil {
							panic(err)
						}
            handler := Auth(logger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

// Helper function to create a valid JWT for testing
func createValidJWT() string {
    token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, jwt.RegisteredClaims{})

    tokenString, err := token.SignedString(jwtKey)
    if err != nil {
        panic(err)
    }

    return tokenString
}
