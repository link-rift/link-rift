# Authentication

> Last Updated: 2026-02-01

Linkrift implements a secure authentication system using PASETO tokens, Argon2id password hashing, and database-backed session management.

## Table of Contents

- [Overview](#overview)
- [Token Management](#token-management)
  - [PASETO Tokens](#paseto-tokens)
- [Password Security](#password-security)
  - [Argon2id Hashing](#argon2id-hashing)
- [Session Management](#session-management)
- [Auth Middleware](#auth-middleware)
- [Auth Service](#auth-service)
- [React Authentication Patterns](#react-authentication-patterns)
- [SSO/SAML (Enterprise)](#ssosaml-enterprise)
- [API Reference](#api-reference)

---

## Overview

The authentication system in Linkrift is designed with security-first principles:

- **PASETO V4 Local tokens** as the primary token format (symmetric encryption, more secure than JWT)
- **Argon2id** for password hashing (winner of Password Hashing Competition)
- **Database-backed sessions** via PostgreSQL with refresh token rotation
- **SSO/SAML** for enterprise single sign-on (gated behind Enterprise license)

### Authentication Flow

```
                          LOGIN FLOW
                          ==========

  Client                    API Server                  Database
    │                          │                           │
    │  POST /auth/login        │                           │
    │  {email, password}       │                           │
    │─────────────────────────>│                           │
    │                          │  SELECT user by email     │
    │                          │──────────────────────────>│
    │                          │         user row          │
    │                          │<──────────────────────────│
    │                          │                           │
    │                          │  Verify password          │
    │                          │  (Argon2id)               │
    │                          │                           │
    │                          │  INSERT session           │
    │                          │──────────────────────────>│
    │                          │        session_id         │
    │                          │<──────────────────────────│
    │                          │                           │
    │                          │  Create PASETO token      │
    │                          │  (V4 Local, symmetric)    │
    │                          │                           │
    │  {access_token,          │                           │
    │   refresh_token, user}   │                           │
    │<─────────────────────────│                           │
    │                          │                           │


                    AUTHENTICATED REQUEST
                    =====================

  Client                    API Server                  Database
    │                          │                           │
    │  GET /api/v1/...         │                           │
    │  Authorization: Bearer   │                           │
    │  <paseto_token>          │                           │
    │─────────────────────────>│                           │
    │                          │  Decrypt & verify token   │
    │                          │  (PASETO V4 Local)        │
    │                          │                           │
    │                          │  Load user from DB        │
    │                          │──────────────────────────>│
    │                          │         user row          │
    │                          │<──────────────────────────│
    │                          │                           │
    │                          │  Set user in context      │
    │                          │  -> Handler processes     │
    │                          │                           │
    │  {success: true,         │                           │
    │   data: {...}}           │                           │
    │<─────────────────────────│                           │
    │                          │                           │


                      TOKEN REFRESH
                      =============

  Client                    API Server                  Database
    │                          │                           │
    │  POST /auth/refresh      │                           │
    │  {refresh_token}         │                           │
    │─────────────────────────>│                           │
    │                          │  Hash refresh token       │
    │                          │  Find session by hash     │
    │                          │──────────────────────────>│
    │                          │        session            │
    │                          │<──────────────────────────│
    │                          │                           │
    │                          │  Revoke old session       │
    │                          │  Create new session       │
    │                          │──────────────────────────>│
    │                          │        new session_id     │
    │                          │<──────────────────────────│
    │                          │                           │
    │  {access_token,          │                           │
    │   refresh_token, user}   │                           │
    │<─────────────────────────│                           │
    │                          │                           │
```

---

## Token Management

### PASETO Tokens

PASETO (Platform-Agnostic Security Tokens) is the primary token format, using V4 Local (symmetric) encryption. This eliminates algorithm confusion attacks inherent in JWT.

```go
// pkg/paseto/paseto.go
package paseto

import (
	"fmt"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/google/uuid"
)

// Claims holds the token payload
type Claims struct {
	UserID    uuid.UUID
	Email     string
	SessionID uuid.UUID
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// Maker defines the token creation/verification interface
type Maker interface {
	CreateToken(userID uuid.UUID, email string, sessionID uuid.UUID, duration time.Duration) (string, *Claims, error)
	VerifyToken(token string) (*Claims, error)
}

type pasetoMaker struct {
	symmetricKey paseto.V4SymmetricKey
}

// NewPasetoMaker creates a new PASETO maker from a secret (min 32 chars)
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

// CreateToken generates a new encrypted PASETO V4 Local token
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

// VerifyToken parses and validates a PASETO V4 Local token
func (m *pasetoMaker) VerifyToken(tokenString string) (*Claims, error) {
	parser := paseto.NewParser()
	parser.AddRule(paseto.NotExpired())
	parser.AddRule(paseto.ValidAt(time.Now()))

	token, err := parser.ParseV4Local(m.symmetricKey, tokenString, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	// ... extract claims from token fields
}
```

**Token lifecycle:**
- Access tokens are short-lived (configured via `AUTH_ACCESS_TOKEN_DURATION`)
- Refresh tokens are longer-lived (configured via `AUTH_REFRESH_TOKEN_DURATION`)
- Refresh tokens are hashed (SHA-256) and stored in the `sessions` table
- Token refresh rotates the refresh token and creates a new session entry

---

## Password Security

### Argon2id Hashing

Argon2id provides resistance against both side-channel and GPU brute-force attacks.

```go
// pkg/crypto/hash.go
package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/argon2"
)

// Default parameters: 64 MB memory, 3 iterations, 2 threads
var defaultParams = &argon2Params{
	memory:      64 * 1024, // 64 MB
	iterations:  3,
	parallelism: 2,
	saltLength:  16,
	keyLength:   32,
}

// HashPassword generates a PHC-format Argon2id hash
func HashPassword(password string) (string, error) {
	salt := make([]byte, defaultParams.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password), salt,
		defaultParams.iterations, defaultParams.memory,
		defaultParams.parallelism, defaultParams.keyLength,
	)

	// Output: $argon2id$v=19$m=65536,t=3,p=2$<base64-salt>$<base64-hash>
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, defaultParams.memory,
		defaultParams.iterations, defaultParams.parallelism,
		b64Salt, b64Hash,
	), nil
}

// VerifyPassword checks a password against a PHC-format hash
func VerifyPassword(password, encodedHash string) (bool, error) {
	p, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}
	otherHash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)
	// Constant-time comparison prevents timing attacks
	return subtle.ConstantTimeCompare(hash, otherHash) == 1, nil
}
```

---

## Session Management

Sessions are stored in PostgreSQL (not Redis) using the `sessions` table. Each session tracks the refresh token hash, IP address, user agent, and expiration.

```go
// internal/models/session.go
type Session struct {
	ID               uuid.UUID `json:"id"`
	UserID           uuid.UUID `json:"user_id"`
	RefreshTokenHash string    `json:"-"`
	IPAddress        string    `json:"ip_address,omitempty"`
	UserAgent        *string   `json:"user_agent,omitempty"`
	DeviceName       *string   `json:"device_name,omitempty"`
	IsRevoked        bool      `json:"is_revoked"`
	LastActiveAt     time.Time `json:"last_active_at"`
	CreatedAt        time.Time `json:"created_at"`
	ExpiresAt        time.Time `json:"expires_at"`
}
```

Session lifecycle:
1. **Login** creates a new session with a hashed refresh token
2. **Token refresh** validates the refresh token hash, creates a new session, and revokes the old one
3. **Logout** revokes the session by ID

---

## Auth Middleware

The auth middleware extracts and validates PASETO tokens from the `Authorization: Bearer <token>` header, then injects the authenticated user into the Gin context.

```go
// internal/middleware/auth.go
func RequireAuth(tokenMaker paseto.Maker, userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, httputil.Response{
				Success: false,
				Error: &httputil.ErrorBody{Code: "UNAUTHORIZED", Message: "missing or invalid authorization header"},
			})
			return
		}

		claims, err := tokenMaker.VerifyToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, httputil.Response{...})
			return
		}

		user, err := userRepo.GetByID(c.Request.Context(), claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, httputil.Response{...})
			return
		}

		c.Set("user", user)
		c.Set("session_id", claims.SessionID)
		c.Next()
	}
}
```

Two middleware variants:
- `RequireAuth()` — aborts with 401 if no valid token
- `OptionalAuth()` — sets user context if token present, continues regardless

Helper functions for downstream handlers:
- `GetUserFromContext(c *gin.Context) *models.User`
- `GetSessionIDFromContext(c *gin.Context) uuid.UUID`

---

## Auth Service

The `AuthService` interface defines the complete auth contract:

```go
// internal/service/auth_service.go
type AuthService interface {
	Register(ctx context.Context, input models.RegisterInput) (*models.AuthResponse, error)
	Login(ctx context.Context, input models.LoginInput, ip, userAgent string) (*models.AuthResponse, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
	RefreshToken(ctx context.Context, refreshToken, ip, userAgent string) (*models.AuthResponse, error)
	GetCurrentUser(ctx context.Context, userID uuid.UUID) (*models.UserResponse, error)
	ForgotPassword(ctx context.Context, input models.ForgotPasswordInput) error
	ResetPassword(ctx context.Context, input models.ResetPasswordInput) error
	VerifyEmail(ctx context.Context, input models.VerifyEmailInput) error
}
```

Dependencies injected via constructor:
- `UserRepository` — user CRUD
- `SessionRepository` — session management
- `PasswordResetRepository` — password reset tokens
- `paseto.Maker` — token creation/verification
- `*pgxpool.Pool` — database transactions
- `*redis.Client` — caching
- `*config.Config` — configuration
- `*zap.Logger` — structured logging

---

## React Authentication Patterns

The frontend uses **Zustand** for auth state and **TanStack Query** for server state synchronization.

### Auth Store (Zustand)

```typescript
// web/src/stores/authStore.ts
import { create } from "zustand"
import type { User } from "@/types/auth"

interface AuthState {
  user: User | null
  accessToken: string | null
  refreshToken: string | null
  isAuthenticated: boolean
  isLoading: boolean
  setAuth: (user: User, accessToken: string, refreshToken: string) => void
  clearAuth: () => void
  setUser: (user: User) => void
  setLoading: (loading: boolean) => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  accessToken: localStorage.getItem("access_token"),
  refreshToken: localStorage.getItem("refresh_token"),
  isAuthenticated: !!localStorage.getItem("access_token"),
  isLoading: true,

  setAuth: (user, accessToken, refreshToken) => {
    localStorage.setItem("access_token", accessToken)
    localStorage.setItem("refresh_token", refreshToken)
    set({ user, accessToken, refreshToken, isAuthenticated: true, isLoading: false })
  },

  clearAuth: () => {
    localStorage.removeItem("access_token")
    localStorage.removeItem("refresh_token")
    set({ user: null, accessToken: null, refreshToken: null, isAuthenticated: false, isLoading: false })
  },
  // ...
}))
```

### Auth Hooks (TanStack Query)

```typescript
// web/src/hooks/useAuth.ts
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useAuthStore } from "@/stores/authStore"
import * as authService from "@/services/auth"

export function useCurrentUser() {
  const { isAuthenticated, setUser, clearAuth } = useAuthStore()
  return useQuery({
    queryKey: ["currentUser"],
    queryFn: async () => {
      try {
        const user = await authService.getMe()
        setUser(user)
        return user
      } catch {
        clearAuth()
        return null
      }
    },
    enabled: isAuthenticated,
    retry: false,
    staleTime: 5 * 60 * 1000,
  })
}

export function useLogin() {
  const { setAuth } = useAuthStore()
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: LoginRequest) => authService.login(data),
    onSuccess: (response) => {
      setAuth(response.user, response.access_token, response.refresh_token)
      queryClient.setQueryData(["currentUser"], response.user)
    },
  })
}

export function useRegister() { /* similar pattern */ }
export function useLogout() { /* clears auth + query cache */ }
export function useForgotPassword() { /* mutation wrapper */ }
export function useResetPassword() { /* mutation wrapper */ }
```

### Auth API Client

```typescript
// web/src/services/auth.ts
import { apiRequest, setTokens, clearTokens } from "./api"
import type { AuthResponse, LoginRequest, RegisterRequest, User } from "@/types/auth"

export async function login(data: LoginRequest): Promise<AuthResponse> {
  const res = await apiRequest<AuthResponse>("/auth/login", {
    method: "POST",
    body: JSON.stringify(data),
  })
  if (!res.success || !res.data) throw new Error(res.error?.message || "Login failed")
  setTokens(res.data.access_token, res.data.refresh_token)
  return res.data
}

export async function getMe(): Promise<User> {
  const res = await apiRequest<User>("/auth/me")
  if (!res.success || !res.data) throw new Error(res.error?.message || "Failed to get user")
  return res.data
}

// register, logout, forgotPassword, resetPassword, verifyEmail follow the same pattern
```

---

## SSO/SAML (Enterprise)

Enterprise-licensed workspaces can configure SAML-based SSO. See the Enterprise Features documentation for details.

- Configuration stored in `sso_configs` table
- Identity mapping stored in `sso_identities` table
- Gated by `RequireFeature("saml")` middleware
- SSO flow endpoints: `/api/v1/auth/sso/login/:workspaceId`, `/api/v1/auth/sso/callback/:workspaceId`

---

## API Reference

### Authentication Endpoints

All endpoints are prefixed with `/api/v1`.

| Method | Endpoint | Auth Required | Description |
|--------|----------|:---:|-------------|
| POST | `/auth/register` | No | Register a new user |
| POST | `/auth/login` | No | Login with email/password |
| POST | `/auth/logout` | Yes | Logout and revoke session |
| POST | `/auth/refresh` | No | Refresh access token |
| GET | `/auth/me` | Yes | Get current user info |
| POST | `/auth/forgot-password` | No | Request password reset email |
| POST | `/auth/reset-password` | No | Reset password with token |
| POST | `/auth/verify-email` | No | Verify email address |

### Request/Response Examples

**Register Request:**
```json
POST /api/v1/auth/register
{
  "email": "user@example.com",
  "password": "securepassword123",
  "name": "John Doe"
}
```

**Login Request:**
```json
POST /api/v1/auth/login
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Auth Response (Success):**
```json
{
  "success": true,
  "data": {
    "access_token": "v4.local.encrypted-token-string...",
    "refresh_token": "random-refresh-token-string",
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "user@example.com",
      "name": "John Doe",
      "avatar_url": null,
      "email_verified_at": null,
      "two_factor_enabled": false,
      "created_at": "2026-01-15T10:30:00Z",
      "updated_at": "2026-01-15T10:30:00Z"
    }
  }
}
```

**Error Response:**
```json
{
  "success": false,
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "invalid email or password"
  }
}
```

**Refresh Token Request:**
```json
POST /api/v1/auth/refresh
{
  "refresh_token": "the-refresh-token-from-login"
}
```

**Get Current User:**
```bash
curl -H "Authorization: Bearer v4.local.encrypted-token..." \
  http://localhost:8080/api/v1/auth/me
```
