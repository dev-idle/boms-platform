// Command genkey prints a base64-encoded Ed25519 seed for JWT_ED25519_PRIVATE_KEY.
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
)

func main() {
	seed := make([]byte, ed25519.SeedSize)
	if _, err := rand.Read(seed); err != nil {
		fmt.Fprintf(os.Stderr, "genkey: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(base64.StdEncoding.EncodeToString(seed))
}
