package paseto

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCreateAndVerifyToken(t *testing.T) {
	maker, err := NewPasetoMaker("test-secret-key-that-is-at-least-32-characters-long")
	if err != nil {
		t.Fatalf("failed to create maker: %v", err)
	}

	userID := uuid.New()
	email := "test@example.com"
	sessionID := uuid.New()
	duration := 15 * time.Minute

	tokenStr, claims, err := maker.CreateToken(userID, email, sessionID, duration)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	if tokenStr == "" {
		t.Fatal("token string is empty")
	}
	if claims.UserID != userID {
		t.Errorf("expected userID %v, got %v", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("expected email %v, got %v", email, claims.Email)
	}
	if claims.SessionID != sessionID {
		t.Errorf("expected sessionID %v, got %v", sessionID, claims.SessionID)
	}

	verified, err := maker.VerifyToken(tokenStr)
	if err != nil {
		t.Fatalf("failed to verify token: %v", err)
	}

	if verified.UserID != userID {
		t.Errorf("verified userID mismatch: got %v, want %v", verified.UserID, userID)
	}
	if verified.Email != email {
		t.Errorf("verified email mismatch: got %v, want %v", verified.Email, email)
	}
	if verified.SessionID != sessionID {
		t.Errorf("verified sessionID mismatch: got %v, want %v", verified.SessionID, sessionID)
	}
}

func TestExpiredToken(t *testing.T) {
	maker, err := NewPasetoMaker("test-secret-key-that-is-at-least-32-characters-long")
	if err != nil {
		t.Fatalf("failed to create maker: %v", err)
	}

	tokenStr, _, err := maker.CreateToken(uuid.New(), "test@example.com", uuid.New(), -time.Minute)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	_, err = maker.VerifyToken(tokenStr)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestInvalidToken(t *testing.T) {
	maker, err := NewPasetoMaker("test-secret-key-that-is-at-least-32-characters-long")
	if err != nil {
		t.Fatalf("failed to create maker: %v", err)
	}

	_, err = maker.VerifyToken("v4.local.invalid-token-data")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func TestDifferentKeyCannotVerify(t *testing.T) {
	maker1, _ := NewPasetoMaker("first-secret-key-that-is-at-least-32-characters")
	maker2, _ := NewPasetoMaker("second-secret-key-that-is-at-least-32-chars")

	tokenStr, _, err := maker1.CreateToken(uuid.New(), "test@example.com", uuid.New(), 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	_, err = maker2.VerifyToken(tokenStr)
	if err == nil {
		t.Fatal("expected error when verifying with different key, got nil")
	}
}

func TestShortSecret(t *testing.T) {
	_, err := NewPasetoMaker("short")
	if err == nil {
		t.Fatal("expected error for short secret, got nil")
	}
}
