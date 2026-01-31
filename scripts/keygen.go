// +build ignore

package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os"
)

func main() {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate key: %v\n", err)
		os.Exit(1)
	}

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pub,
	})

	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: priv,
	})

	if err := os.WriteFile("internal/license/keys/public.pem", pubPEM, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write public key: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile("internal/license/keys/private.pem", privPEM, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "write private key: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Ed25519 keypair generated:")
	fmt.Println("  Public:  internal/license/keys/public.pem")
	fmt.Println("  Private: internal/license/keys/private.pem")
	fmt.Println("")
	fmt.Println("WARNING: Never commit private.pem to source control!")
}
