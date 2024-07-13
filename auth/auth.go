package auth

import (
	"crypto/ed25519"
	"encoding/hex"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

var jwtPublicKey ed25519.PublicKey

func init() {
	publicKeyBytes, err := hex.DecodeString("51d83030a2a6796341950abbfb7516772d0c184583dc76db22f8c8ada3994e4a")
	if err != nil {
		panic("Invalid public key")
	}
	jwtPublicKey = ed25519.PublicKey(publicKeyBytes)
}

// This function verifies the JWT in the authorization header and returns true if it's valid
func Verify(authHeader string) bool {
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
	println("Public Key: ",jwtPublicKey)
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