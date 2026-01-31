package crypto

import (
	"strings"
	"testing"
)

func TestHashAndVerify(t *testing.T) {
	password := "correct-horse-battery-staple"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	match, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword() error: %v", err)
	}
	if !match {
		t.Error("VerifyPassword() returned false for correct password")
	}
}

func TestVerifyWrongPassword(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	match, err := VerifyPassword("wrong-password", hash)
	if err != nil {
		t.Fatalf("VerifyPassword() error: %v", err)
	}
	if match {
		t.Error("VerifyPassword() returned true for wrong password")
	}
}

func TestHashFormat(t *testing.T) {
	hash, err := HashPassword("test")
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	if !strings.HasPrefix(hash, "$argon2id$v=19$") {
		t.Errorf("hash should start with $argon2id$v=19$, got %s", hash)
	}

	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		t.Errorf("expected 6 parts in PHC format, got %d", len(parts))
	}
}

func TestHashUniqueness(t *testing.T) {
	hash1, _ := HashPassword("same-password")
	hash2, _ := HashPassword("same-password")

	if hash1 == hash2 {
		t.Error("hashing same password twice should produce different hashes (different salts)")
	}
}

func TestVerifyInvalidHash(t *testing.T) {
	_, err := VerifyPassword("test", "not-a-valid-hash")
	if err == nil {
		t.Error("expected error for invalid hash format")
	}
}

func TestEmptyPassword(t *testing.T) {
	hash, err := HashPassword("")
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	match, err := VerifyPassword("", hash)
	if err != nil {
		t.Fatalf("VerifyPassword() error: %v", err)
	}
	if !match {
		t.Error("empty password should verify against its own hash")
	}

	match, err = VerifyPassword("not-empty", hash)
	if err != nil {
		t.Fatalf("VerifyPassword() error: %v", err)
	}
	if match {
		t.Error("non-empty password should not match empty password hash")
	}
}
