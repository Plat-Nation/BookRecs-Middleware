package auth

import (
	"crypto/ed25519"
	"encoding/hex"
	"os"
	"testing"

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

// Helper function to create a valid JWT for testing
func createValidJWT() string {
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, jwt.RegisteredClaims{})

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
			panic(err)
	}

	return tokenString
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
            if got := Verify(test.authHeader); got != test.expected {
                t.Errorf("verify() = %v, expected %v", got, test.expected)
            }
        })
    }
}