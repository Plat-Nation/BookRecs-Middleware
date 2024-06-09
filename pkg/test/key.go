package test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v4"
)

var jwtKey ed25519.PrivateKey


func init() {
	privateKeyHex := os.Getenv("JWT_KEY")
	if privateKeyHex == "" {
			fmt.Println("Error: JWT_KEY environment variable is not set")
			os.Exit(1)
	}

	privateKeyBytes, err := hex.DecodeString(os.Getenv("JWT_KEY"))
	if err != nil {
			panic("Invalid private key")
	}
	jwtKey = ed25519.PrivateKey(privateKeyBytes)
}

// Utility function for generating keys, only used manually to initially create keys
func GenerateKeys() {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	fmt.Printf("Public Key: %x\n", pub)
  fmt.Printf("Private Key: %x\n", priv)
}

func GenerateJWT() string {
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, jwt.RegisteredClaims{})

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
			panic(err)
	}

	fmt.Println(tokenString)
	return tokenString
}