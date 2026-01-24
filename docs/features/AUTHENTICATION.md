# Authentication

> Last Updated: 2025-01-24

Linkrift implements a comprehensive authentication system supporting multiple authentication methods, secure token management, and modern security practices.

## Table of Contents

- [Overview](#overview)
- [Token Management](#token-management)
  - [PASETO Tokens](#paseto-tokens)
  - [JWT Support](#jwt-support)
- [OAuth2 Integration](#oauth2-integration)
  - [Google OAuth](#google-oauth)
  - [GitHub OAuth](#github-oauth)
- [Password Security](#password-security)
  - [Argon2id Hashing](#argon2id-hashing)
- [Two-Factor Authentication](#two-factor-authentication)
  - [TOTP Implementation](#totp-implementation)
- [Session Management](#session-management)
- [React Authentication Patterns](#react-authentication-patterns)
- [API Reference](#api-reference)

---

## Overview

The authentication system in Linkrift is designed with security-first principles:

- **PASETO tokens** as the primary token format (more secure than JWT)
- **JWT support** for legacy integrations and third-party services
- **OAuth2** for social login (Google, GitHub)
- **Argon2id** for password hashing (winner of Password Hashing Competition)
- **TOTP-based 2FA** for additional security
- **Secure session management** with Redis-backed storage

## Token Management

### PASETO Tokens

PASETO (Platform-Agnostic Security Tokens) is our primary token format, offering improved security over JWT by eliminating algorithm confusion attacks.

```go
// internal/auth/paseto.go
package auth

import (
	"crypto/ed25519"
	"time"

	"github.com/o1egl/paseto"
	"github.com/link-rift/link-rift/internal/config"
)

// PASETOClaims represents the token payload
type PASETOClaims struct {
	UserID      string    `json:"user_id"`
	Email       string    `json:"email"`
	WorkspaceID string    `json:"workspace_id,omitempty"`
	Role        string    `json:"role"`
	IssuedAt    time.Time `json:"iat"`
	ExpiresAt   time.Time `json:"exp"`
	TokenType   TokenType `json:"type"`
}

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// PASETOManager handles PASETO token operations
type PASETOManager struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	paseto     *paseto.V2
}

// NewPASETOManager creates a new PASETO manager
func NewPASETOManager(cfg *config.AuthConfig) (*PASETOManager, error) {
	privateKey := ed25519.NewKeyFromSeed([]byte(cfg.PASETOSecret))
	publicKey := privateKey.Public().(ed25519.PublicKey)

	return &PASETOManager{
		privateKey: privateKey,
		publicKey:  publicKey,
		paseto:     paseto.NewV2(),
	}, nil
}

// GenerateAccessToken creates a new access token
func (pm *PASETOManager) GenerateAccessToken(userID, email, workspaceID, role string) (string, error) {
	now := time.Now()
	claims := PASETOClaims{
		UserID:      userID,
		Email:       email,
		WorkspaceID: workspaceID,
		Role:        role,
		IssuedAt:    now,
		ExpiresAt:   now.Add(15 * time.Minute), // Short-lived access tokens
		TokenType:   TokenTypeAccess,
	}

	return pm.paseto.Sign(pm.privateKey, claims, nil)
}

// GenerateRefreshToken creates a new refresh token
func (pm *PASETOManager) GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()
	claims := PASETOClaims{
		UserID:    userID,
		IssuedAt:  now,
		ExpiresAt: now.Add(7 * 24 * time.Hour), // 7 days
		TokenType: TokenTypeRefresh,
	}

	return pm.paseto.Sign(pm.privateKey, claims, nil)
}

// VerifyToken validates and parses a PASETO token
func (pm *PASETOManager) VerifyToken(token string) (*PASETOClaims, error) {
	var claims PASETOClaims

	err := pm.paseto.Verify(token, pm.publicKey, &claims, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if time.Now().After(claims.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	return &claims, nil
}
```

### JWT Support

For backwards compatibility and third-party integrations:

```go
// internal/auth/jwt.go
package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/link-rift/link-rift/internal/config"
)

// JWTClaims represents JWT token claims
type JWTClaims struct {
	jwt.RegisteredClaims
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	WorkspaceID string `json:"workspace_id,omitempty"`
	Role        string `json:"role"`
}

// JWTManager handles JWT operations
type JWTManager struct {
	secretKey     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(cfg *config.AuthConfig) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(cfg.JWTSecret),
		accessExpiry:  15 * time.Minute,
		refreshExpiry: 7 * 24 * time.Hour,
	}
}

// GenerateToken creates a new JWT token
func (jm *JWTManager) GenerateToken(userID, email, workspaceID, role string) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "linkrift",
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(jm.accessExpiry)),
		},
		UserID:      userID,
		Email:       email,
		WorkspaceID: workspaceID,
		Role:        role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jm.secretKey)
}

// ValidateToken parses and validates a JWT token
func (jm *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSigningMethod
		}
		return jm.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
```

---

## OAuth2 Integration

### Google OAuth

```go
// internal/auth/oauth/google.go
package oauth

import (
	"context"
	"encoding/json"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"github.com/link-rift/link-rift/internal/config"
)

// GoogleUser represents user info from Google
type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// GoogleProvider handles Google OAuth2
type GoogleProvider struct {
	config *oauth2.Config
}

// NewGoogleProvider creates a new Google OAuth provider
func NewGoogleProvider(cfg *config.OAuthConfig) *GoogleProvider {
	return &GoogleProvider{
		config: &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}
}

// GetAuthURL generates the OAuth2 authorization URL
func (gp *GoogleProvider) GetAuthURL(state string) string {
	return gp.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges authorization code for tokens
func (gp *GoogleProvider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return gp.config.Exchange(ctx, code)
}

// GetUserInfo fetches user information from Google
func (gp *GoogleProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUser, error) {
	client := gp.config.Client(ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrOAuthFetchFailed
	}

	var user GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
```

### GitHub OAuth

```go
// internal/auth/oauth/github.go
package oauth

import (
	"context"
	"encoding/json"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"github.com/link-rift/link-rift/internal/config"
)

// GitHubUser represents user info from GitHub
type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// GitHubEmail represents a GitHub email
type GitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

// GitHubProvider handles GitHub OAuth2
type GitHubProvider struct {
	config *oauth2.Config
}

// NewGitHubProvider creates a new GitHub OAuth provider
func NewGitHubProvider(cfg *config.OAuthConfig) *GitHubProvider {
	return &GitHubProvider{
		config: &oauth2.Config{
			ClientID:     cfg.GitHubClientID,
			ClientSecret: cfg.GitHubClientSecret,
			RedirectURL:  cfg.GitHubRedirectURL,
			Scopes:       []string{"user:email", "read:user"},
			Endpoint:     github.Endpoint,
		},
	}
}

// GetAuthURL generates the OAuth2 authorization URL
func (gp *GitHubProvider) GetAuthURL(state string) string {
	return gp.config.AuthCodeURL(state)
}

// ExchangeCode exchanges authorization code for tokens
func (gp *GitHubProvider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return gp.config.Exchange(ctx, code)
}

// GetUserInfo fetches user information from GitHub
func (gp *GitHubProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*GitHubUser, error) {
	client := gp.config.Client(ctx, token)

	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	// Fetch primary email if not public
	if user.Email == "" {
		email, err := gp.fetchPrimaryEmail(ctx, client)
		if err == nil {
			user.Email = email
		}
	}

	return &user, nil
}

func (gp *GitHubProvider) fetchPrimaryEmail(ctx context.Context, client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emails []GitHubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}

	return "", ErrNoPrimaryEmail
}
```

---

## Password Security

### Argon2id Hashing

Argon2id is the recommended password hashing algorithm, combining resistance against both side-channel and GPU attacks.

```go
// internal/auth/password.go
package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2Params defines the Argon2id parameters
type Argon2Params struct {
	Memory      uint32 // Memory cost in KiB
	Iterations  uint32 // Time cost (number of iterations)
	Parallelism uint8  // Degree of parallelism
	SaltLength  uint32 // Salt length in bytes
	KeyLength   uint32 // Derived key length in bytes
}

// DefaultArgon2Params returns recommended parameters for Argon2id
func DefaultArgon2Params() *Argon2Params {
	return &Argon2Params{
		Memory:      64 * 1024, // 64 MB
		Iterations:  3,
		Parallelism: 4,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// PasswordHasher handles password hashing with Argon2id
type PasswordHasher struct {
	params *Argon2Params
}

// NewPasswordHasher creates a new password hasher
func NewPasswordHasher(params *Argon2Params) *PasswordHasher {
	if params == nil {
		params = DefaultArgon2Params()
	}
	return &PasswordHasher{params: params}
}

// Hash generates an Argon2id hash of the password
func (ph *PasswordHasher) Hash(password string) (string, error) {
	salt := make([]byte, ph.params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		ph.params.Iterations,
		ph.params.Memory,
		ph.params.Parallelism,
		ph.params.KeyLength,
	)

	// Encode to PHC string format
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		ph.params.Memory,
		ph.params.Iterations,
		ph.params.Parallelism,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

// Verify compares a password against an Argon2id hash
func (ph *PasswordHasher) Verify(password, encodedHash string) (bool, error) {
	params, salt, hash, err := ph.decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	otherHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	// Constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}

	return false, nil
}

func (ph *PasswordHasher) decodeHash(encodedHash string) (*Argon2Params, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, ErrInvalidHashFormat
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}

	params := &Argon2Params{}
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.Memory, &params.Iterations, &params.Parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, err
	}
	params.SaltLength = uint32(len(salt))

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, err
	}
	params.KeyLength = uint32(len(hash))

	return params, salt, hash, nil
}
```

---

## Two-Factor Authentication

### TOTP Implementation

```go
// internal/auth/totp.go
package auth

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TOTPConfig holds TOTP configuration
type TOTPConfig struct {
	Issuer      string
	AccountName string
	Period      uint
	SecretSize  uint
	Algorithm   otp.Algorithm
	Digits      otp.Digits
}

// TOTPManager handles TOTP operations
type TOTPManager struct {
	issuer string
}

// NewTOTPManager creates a new TOTP manager
func NewTOTPManager(issuer string) *TOTPManager {
	return &TOTPManager{issuer: issuer}
}

// GenerateSecret generates a new TOTP secret for a user
func (tm *TOTPManager) GenerateSecret(email string) (*otp.Key, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      tm.issuer,
		AccountName: email,
		Period:      30,
		SecretSize:  32,
		Algorithm:   otp.AlgorithmSHA256,
		Digits:      otp.DigitsSix,
	})
	if err != nil {
		return nil, err
	}

	return key, nil
}

// ValidateCode validates a TOTP code
func (tm *TOTPManager) ValidateCode(secret, code string) bool {
	return totp.Validate(code, secret)
}

// GenerateRecoveryCodes generates backup recovery codes
func (tm *TOTPManager) GenerateRecoveryCodes(count int) ([]string, error) {
	codes := make([]string, count)

	for i := 0; i < count; i++ {
		bytes := make([]byte, 5)
		if _, err := rand.Read(bytes); err != nil {
			return nil, err
		}
		// Format: XXXX-XXXX
		code := base32.StdEncoding.EncodeToString(bytes)[:8]
		codes[i] = fmt.Sprintf("%s-%s", code[:4], code[4:])
	}

	return codes, nil
}

// TOTPSetup represents the setup response
type TOTPSetup struct {
	Secret        string   `json:"secret"`
	QRCode        string   `json:"qr_code"` // Base64 encoded PNG
	RecoveryCodes []string `json:"recovery_codes"`
}
```

---

## Session Management

```go
// internal/auth/session.go
package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/link-rift/link-rift/internal/config"
)

// Session represents a user session
type Session struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	WorkspaceID  string    `json:"workspace_id,omitempty"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	CreatedAt    time.Time `json:"created_at"`
	LastActiveAt time.Time `json:"last_active_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// SessionManager handles session operations
type SessionManager struct {
	redis     *redis.Client
	ttl       time.Duration
	keyPrefix string
}

// NewSessionManager creates a new session manager
func NewSessionManager(rdb *redis.Client, cfg *config.SessionConfig) *SessionManager {
	return &SessionManager{
		redis:     rdb,
		ttl:       cfg.SessionTTL,
		keyPrefix: "session:",
	}
}

// Create creates a new session
func (sm *SessionManager) Create(ctx context.Context, userID, workspaceID, ip, userAgent string) (*Session, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := &Session{
		ID:           sessionID,
		UserID:       userID,
		WorkspaceID:  workspaceID,
		IPAddress:    ip,
		UserAgent:    userAgent,
		CreatedAt:    now,
		LastActiveAt: now,
		ExpiresAt:    now.Add(sm.ttl),
	}

	key := sm.keyPrefix + sessionID
	if err := sm.redis.HSet(ctx, key, session).Err(); err != nil {
		return nil, err
	}

	if err := sm.redis.Expire(ctx, key, sm.ttl).Err(); err != nil {
		return nil, err
	}

	// Track session in user's session set
	userSessionsKey := fmt.Sprintf("user_sessions:%s", userID)
	sm.redis.SAdd(ctx, userSessionsKey, sessionID)

	return session, nil
}

// Get retrieves a session by ID
func (sm *SessionManager) Get(ctx context.Context, sessionID string) (*Session, error) {
	key := sm.keyPrefix + sessionID

	var session Session
	if err := sm.redis.HGetAll(ctx, key).Scan(&session); err != nil {
		return nil, err
	}

	if session.ID == "" {
		return nil, ErrSessionNotFound
	}

	return &session, nil
}

// Touch updates the session's last active time
func (sm *SessionManager) Touch(ctx context.Context, sessionID string) error {
	key := sm.keyPrefix + sessionID

	pipe := sm.redis.Pipeline()
	pipe.HSet(ctx, key, "last_active_at", time.Now())
	pipe.Expire(ctx, key, sm.ttl)

	_, err := pipe.Exec(ctx)
	return err
}

// Revoke invalidates a session
func (sm *SessionManager) Revoke(ctx context.Context, sessionID string) error {
	session, err := sm.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	key := sm.keyPrefix + sessionID
	if err := sm.redis.Del(ctx, key).Err(); err != nil {
		return err
	}

	// Remove from user's session set
	userSessionsKey := fmt.Sprintf("user_sessions:%s", session.UserID)
	return sm.redis.SRem(ctx, userSessionsKey, sessionID).Err()
}

// RevokeAllForUser revokes all sessions for a user
func (sm *SessionManager) RevokeAllForUser(ctx context.Context, userID string) error {
	userSessionsKey := fmt.Sprintf("user_sessions:%s", userID)

	sessionIDs, err := sm.redis.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return err
	}

	pipe := sm.redis.Pipeline()
	for _, sessionID := range sessionIDs {
		pipe.Del(ctx, sm.keyPrefix+sessionID)
	}
	pipe.Del(ctx, userSessionsKey)

	_, err = pipe.Exec(ctx)
	return err
}

func generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
```

---

## React Authentication Patterns

### Auth Context Provider

```typescript
// src/contexts/AuthContext.tsx
import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { authApi, TokenResponse, User } from '@/api/auth';

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  loginWithGoogle: () => void;
  loginWithGitHub: () => void;
  logout: () => Promise<void>;
  refreshToken: () => Promise<void>;
  verify2FA: (code: string) => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const TOKEN_REFRESH_INTERVAL = 14 * 60 * 1000; // 14 minutes

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [pendingMFA, setPendingMFA] = useState<string | null>(null);

  const handleTokenResponse = useCallback((response: TokenResponse) => {
    localStorage.setItem('access_token', response.accessToken);
    localStorage.setItem('refresh_token', response.refreshToken);
    setUser(response.user);
  }, []);

  const refreshToken = useCallback(async () => {
    const refreshToken = localStorage.getItem('refresh_token');
    if (!refreshToken) return;

    try {
      const response = await authApi.refresh(refreshToken);
      handleTokenResponse(response);
    } catch (error) {
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
      setUser(null);
    }
  }, [handleTokenResponse]);

  // Initial auth check
  useEffect(() => {
    const initAuth = async () => {
      const token = localStorage.getItem('access_token');
      if (token) {
        try {
          const user = await authApi.me();
          setUser(user);
        } catch {
          await refreshToken();
        }
      }
      setIsLoading(false);
    };
    initAuth();
  }, [refreshToken]);

  // Token refresh interval
  useEffect(() => {
    if (!user) return;

    const interval = setInterval(refreshToken, TOKEN_REFRESH_INTERVAL);
    return () => clearInterval(interval);
  }, [user, refreshToken]);

  const login = async (email: string, password: string) => {
    const response = await authApi.login({ email, password });

    if (response.requiresMFA) {
      setPendingMFA(response.mfaToken);
      throw new Error('MFA_REQUIRED');
    }

    handleTokenResponse(response);
  };

  const verify2FA = async (code: string) => {
    if (!pendingMFA) throw new Error('No pending MFA');

    const response = await authApi.verifyMFA({
      mfaToken: pendingMFA,
      code
    });

    setPendingMFA(null);
    handleTokenResponse(response);
  };

  const loginWithGoogle = () => {
    window.location.href = `${import.meta.env.VITE_API_URL}/auth/google`;
  };

  const loginWithGitHub = () => {
    window.location.href = `${import.meta.env.VITE_API_URL}/auth/github`;
  };

  const logout = async () => {
    try {
      await authApi.logout();
    } finally {
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
      setUser(null);
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: !!user,
        isLoading,
        login,
        loginWithGoogle,
        loginWithGitHub,
        logout,
        refreshToken,
        verify2FA,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
```

### Protected Route Component

```typescript
// src/components/ProtectedRoute.tsx
import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '@/contexts/AuthContext';

interface ProtectedRouteProps {
  children: React.ReactNode;
  requiredRole?: string[];
}

export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({
  children,
  requiredRole
}) => {
  const { isAuthenticated, isLoading, user } = useAuth();
  const location = useLocation();

  if (isLoading) {
    return <div className="flex items-center justify-center h-screen">
      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
    </div>;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  if (requiredRole && user && !requiredRole.includes(user.role)) {
    return <Navigate to="/unauthorized" replace />;
  }

  return <>{children}</>;
};
```

### Auth API Client

```typescript
// src/api/auth.ts
import { apiClient } from './client';

export interface User {
  id: string;
  email: string;
  name: string;
  avatarUrl?: string;
  role: string;
  workspaceId?: string;
  mfaEnabled: boolean;
}

export interface TokenResponse {
  accessToken: string;
  refreshToken: string;
  user: User;
  requiresMFA?: boolean;
  mfaToken?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface MFAVerifyRequest {
  mfaToken: string;
  code: string;
}

export const authApi = {
  login: async (data: LoginRequest): Promise<TokenResponse> => {
    const response = await apiClient.post<TokenResponse>('/auth/login', data);
    return response.data;
  },

  register: async (data: { email: string; password: string; name: string }): Promise<TokenResponse> => {
    const response = await apiClient.post<TokenResponse>('/auth/register', data);
    return response.data;
  },

  logout: async (): Promise<void> => {
    await apiClient.post('/auth/logout');
  },

  me: async (): Promise<User> => {
    const response = await apiClient.get<User>('/auth/me');
    return response.data;
  },

  refresh: async (refreshToken: string): Promise<TokenResponse> => {
    const response = await apiClient.post<TokenResponse>('/auth/refresh', { refreshToken });
    return response.data;
  },

  verifyMFA: async (data: MFAVerifyRequest): Promise<TokenResponse> => {
    const response = await apiClient.post<TokenResponse>('/auth/mfa/verify', data);
    return response.data;
  },

  setupMFA: async (): Promise<{ secret: string; qrCode: string }> => {
    const response = await apiClient.post('/auth/mfa/setup');
    return response.data;
  },

  enableMFA: async (code: string): Promise<{ recoveryCodes: string[] }> => {
    const response = await apiClient.post('/auth/mfa/enable', { code });
    return response.data;
  },

  disableMFA: async (password: string): Promise<void> => {
    await apiClient.post('/auth/mfa/disable', { password });
  },
};
```

---

## API Reference

### Authentication Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/register` | Register a new user |
| POST | `/auth/login` | Login with email/password |
| POST | `/auth/logout` | Logout and invalidate session |
| POST | `/auth/refresh` | Refresh access token |
| GET | `/auth/me` | Get current user info |
| GET | `/auth/google` | Initiate Google OAuth |
| GET | `/auth/google/callback` | Google OAuth callback |
| GET | `/auth/github` | Initiate GitHub OAuth |
| GET | `/auth/github/callback` | GitHub OAuth callback |
| POST | `/auth/mfa/setup` | Initialize 2FA setup |
| POST | `/auth/mfa/enable` | Enable 2FA |
| POST | `/auth/mfa/verify` | Verify 2FA code |
| POST | `/auth/mfa/disable` | Disable 2FA |
| POST | `/auth/password/reset` | Request password reset |
| POST | `/auth/password/reset/confirm` | Confirm password reset |

### Request/Response Examples

**Login Request:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Login Response (Success):**
```json
{
  "access_token": "v2.public.eyJ1c2VyX2lkIjoiMTIzNDU...",
  "refresh_token": "v2.public.eyJ1c2VyX2lkIjoiMTIzNDU...",
  "user": {
    "id": "usr_123456789",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "admin",
    "mfa_enabled": false
  }
}
```

**Login Response (MFA Required):**
```json
{
  "requires_mfa": true,
  "mfa_token": "mfa_abc123xyz"
}
```
