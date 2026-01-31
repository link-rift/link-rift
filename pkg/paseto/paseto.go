package paseto

import (
	"fmt"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/google/uuid"
)

type Claims struct {
	UserID    uuid.UUID
	Email     string
	SessionID uuid.UUID
	IssuedAt  time.Time
	ExpiresAt time.Time
}

type Maker interface {
	CreateToken(userID uuid.UUID, email string, sessionID uuid.UUID, duration time.Duration) (string, *Claims, error)
	VerifyToken(token string) (*Claims, error)
}

type pasetoMaker struct {
	symmetricKey paseto.V4SymmetricKey
}

func NewPasetoMaker(secret string) (Maker, error) {
	if len(secret) < 32 {
		return nil, fmt.Errorf("token secret must be at least 32 characters")
	}

	key, err := paseto.V4SymmetricKeyFromBytes([]byte(secret)[:32])
	if err != nil {
		return nil, fmt.Errorf("failed to create symmetric key: %w", err)
	}

	return &pasetoMaker{symmetricKey: key}, nil
}

func (m *pasetoMaker) CreateToken(userID uuid.UUID, email string, sessionID uuid.UUID, duration time.Duration) (string, *Claims, error) {
	now := time.Now()
	claims := &Claims{
		UserID:    userID,
		Email:     email,
		SessionID: sessionID,
		IssuedAt:  now,
		ExpiresAt: now.Add(duration),
	}

	token := paseto.NewToken()
	token.SetIssuedAt(claims.IssuedAt)
	token.SetExpiration(claims.ExpiresAt)
	token.SetNotBefore(claims.IssuedAt)
	token.SetString("user_id", claims.UserID.String())
	token.SetString("email", claims.Email)
	token.SetString("session_id", claims.SessionID.String())

	encrypted := token.V4Encrypt(m.symmetricKey, nil)
	return encrypted, claims, nil
}

func (m *pasetoMaker) VerifyToken(tokenString string) (*Claims, error) {
	parser := paseto.NewParser()
	parser.AddRule(paseto.NotExpired())
	parser.AddRule(paseto.ValidAt(time.Now()))

	token, err := parser.ParseV4Local(m.symmetricKey, tokenString, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	userIDStr, err := token.GetString("user_id")
	if err != nil {
		return nil, fmt.Errorf("missing user_id claim: %w", err)
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}

	email, err := token.GetString("email")
	if err != nil {
		return nil, fmt.Errorf("missing email claim: %w", err)
	}

	sessionIDStr, err := token.GetString("session_id")
	if err != nil {
		return nil, fmt.Errorf("missing session_id claim: %w", err)
	}
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid session_id: %w", err)
	}

	issuedAt, err := token.GetIssuedAt()
	if err != nil {
		return nil, fmt.Errorf("missing iat claim: %w", err)
	}

	expiresAt, err := token.GetExpiration()
	if err != nil {
		return nil, fmt.Errorf("missing exp claim: %w", err)
	}

	return &Claims{
		UserID:    userID,
		Email:     email,
		SessionID: sessionID,
		IssuedAt:  issuedAt,
		ExpiresAt: expiresAt,
	}, nil
}
