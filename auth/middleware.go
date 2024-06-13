package auth

import (
	"crypto/ed25519"
	"encoding/hex"
	"net/http"
	"strings"

	log "github.com/Plat-Nation/BookRecs-Middleware/log"
	"github.com/golang-jwt/jwt/v4"
)

var jwtPublicKey ed25519.PublicKey

func init() {
	publicKeyBytes, err := hex.DecodeString("f6ac3a793ccff33ad19999a612ca65db90a8db09f18dba6014e4df18a5992424")
	if err != nil {
		panic("Invalid public key")
	}
	jwtPublicKey = ed25519.PublicKey(publicKeyBytes)
}

// This function verifies the JWT in the authorization header and returns true if it's valid
func verify(authHeader string) bool {
	if authHeader == "" {
		return false
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		// If the auth header didn't have a Bearer prefix, block the request
		return false
	}

	// Parse and validate the JWT
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
					return nil, jwt.ErrSignatureInvalid
			}
			return jwtPublicKey, nil
	})

	if err != nil || !token.Valid {
			return false
	}

	return true
}

// The auth middleware takes in an HTTP handler that the traffic should be routed to on success.
// If the Authorization header is not included in the request or the JWT included in it is invalid,
// it responds with a 403, otherwise traffic is forwarded.
func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		// Pass the header to the verify function that will return true or false if the header is valid
		// Forward the traffic along if the header is valid, otherwise block it
		if verify(authHeader) {
			next.ServeHTTP(w, r)
		} else {
			log.Log(r, "Failed Login")
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	}
}