# Security Documentation

**Last Updated: 2025-01-24**

This document covers Linkrift's security architecture, authentication mechanisms, data protection, and security best practices.

---

## Table of Contents

- [Overview](#overview)
- [Security Architecture](#security-architecture)
  - [Defense in Depth](#defense-in-depth)
  - [Network Security](#network-security)
  - [Application Security Layers](#application-security-layers)
- [PASETO Token Security](#paseto-token-security)
  - [Token Structure](#token-structure)
  - [Token Generation](#token-generation)
  - [Token Validation](#token-validation)
  - [Token Rotation](#token-rotation)
- [CORS Configuration](#cors-configuration)
  - [CORS Middleware](#cors-middleware)
  - [Preflight Handling](#preflight-handling)
  - [Environment-Specific Configuration](#environment-specific-configuration)
- [Data Encryption](#data-encryption)
  - [Encryption at Rest](#encryption-at-rest)
  - [Encryption in Transit](#encryption-in-transit)
  - [Field-Level Encryption](#field-level-encryption)
- [Input Validation](#input-validation)
  - [Struct Validation with go-playground/validator](#struct-validation-with-go-playgroundvalidator)
  - [Custom Validators](#custom-validators)
  - [Request Validation Middleware](#request-validation-middleware)
- [SQL Injection Prevention](#sql-injection-prevention)
  - [Type-Safe Queries with sqlc](#type-safe-queries-with-sqlc)
  - [Parameterized Queries](#parameterized-queries)
  - [Query Safety Examples](#query-safety-examples)
- [Rate Limiting](#rate-limiting)
  - [Rate Limiting Strategy](#rate-limiting-strategy)
  - [Implementation](#implementation)
  - [Bypass Prevention](#bypass-prevention)
- [Secure Headers](#secure-headers)
  - [Security Headers Middleware](#security-headers-middleware)
  - [Content Security Policy](#content-security-policy)
- [gosec Integration](#gosec-integration)
  - [Configuration](#configuration)
  - [CI/CD Integration](#cicd-integration)
  - [Handling Findings](#handling-findings)
- [Additional Security Measures](#additional-security-measures)
- [Security Checklist](#security-checklist)

---

## Overview

Linkrift implements a comprehensive security model based on industry best practices:

| Layer | Protection |
|-------|------------|
| Network | TLS 1.3, WAF, DDoS protection |
| Transport | HTTPS only, HSTS |
| Application | Input validation, CSRF protection |
| Authentication | PASETO tokens, secure sessions |
| Authorization | RBAC, resource-level permissions |
| Data | Encryption at rest and in transit |

---

## Security Architecture

### Defense in Depth

```
                    ┌─────────────────────────────────────────┐
                    │              Internet                    │
                    └────────────────┬────────────────────────┘
                                     │
                    ┌────────────────▼────────────────────────┐
                    │         Cloudflare WAF/DDoS             │
                    │    - Rate limiting (edge)                │
                    │    - Bot protection                      │
                    │    - SSL termination                     │
                    └────────────────┬────────────────────────┘
                                     │
                    ┌────────────────▼────────────────────────┐
                    │           Load Balancer                  │
                    │    - Health checks                       │
                    │    - SSL re-encryption                   │
                    └────────────────┬────────────────────────┘
                                     │
                    ┌────────────────▼────────────────────────┐
                    │        Application Layer                 │
                    │    - Input validation                    │
                    │    - Authentication                      │
                    │    - Authorization                       │
                    │    - Rate limiting (app)                 │
                    └────────────────┬────────────────────────┘
                                     │
                    ┌────────────────▼────────────────────────┐
                    │          Data Layer                      │
                    │    - Encryption at rest                  │
                    │    - Access controls                     │
                    │    - Audit logging                       │
                    └─────────────────────────────────────────┘
```

### Network Security

```yaml
# kubernetes/network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: linkrift-api-policy
spec:
  podSelector:
    matchLabels:
      app: linkrift-api
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app: nginx-ingress
      ports:
        - protocol: TCP
          port: 8080
  egress:
    - to:
        - podSelector:
            matchLabels:
              app: postgresql
      ports:
        - protocol: TCP
          port: 5432
    - to:
        - podSelector:
            matchLabels:
              app: redis
      ports:
        - protocol: TCP
          port: 6379
```

### Application Security Layers

```go
// internal/server/middleware.go
package server

import (
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "linkrift/internal/auth"
    "linkrift/internal/security"
)

func (s *Server) setupMiddleware(r chi.Router) {
    // 1. Request ID for tracing
    r.Use(middleware.RequestID)

    // 2. Real IP extraction (for rate limiting)
    r.Use(middleware.RealIP)

    // 3. Panic recovery
    r.Use(security.RecoveryMiddleware)

    // 4. Security headers
    r.Use(security.SecureHeaders)

    // 5. CORS handling
    r.Use(security.CORSMiddleware(s.config.AllowedOrigins))

    // 6. Rate limiting
    r.Use(security.RateLimitMiddleware(s.rateLimiter))

    // 7. Request logging
    r.Use(middleware.Logger)

    // 8. Request timeout
    r.Use(middleware.Timeout(30 * time.Second))
}

func (s *Server) setupProtectedRoutes(r chi.Router) {
    r.Group(func(r chi.Router) {
        // Authentication required
        r.Use(auth.RequireAuth(s.tokenValidator))

        // CSRF protection for state-changing operations
        r.Use(security.CSRFMiddleware)

        r.Route("/api/v1", func(r chi.Router) {
            r.Route("/links", s.linksRoutes)
            r.Route("/account", s.accountRoutes)
        })
    })
}
```

---

## PASETO Token Security

### Token Structure

PASETO (Platform-Agnostic Security Tokens) provides a more secure alternative to JWT with sensible defaults.

```go
// internal/auth/token.go
package auth

import (
    "time"

    "aidanwoods.dev/go-paseto"
)

// TokenClaims represents the claims stored in a PASETO token
type TokenClaims struct {
    UserID    string   `json:"sub"`
    Email     string   `json:"email"`
    Roles     []string `json:"roles"`
    SessionID string   `json:"sid"`
    IssuedAt  time.Time `json:"iat"`
    ExpiresAt time.Time `json:"exp"`
    NotBefore time.Time `json:"nbf"`
    Issuer    string   `json:"iss"`
    Audience  string   `json:"aud"`
}
```

### Token Generation

```go
// internal/auth/token_service.go
package auth

import (
    "crypto/rand"
    "encoding/hex"
    "time"

    "aidanwoods.dev/go-paseto"
)

type TokenService struct {
    symmetricKey paseto.V4SymmetricKey
    issuer       string
    audience     string
    tokenTTL     time.Duration
    refreshTTL   time.Duration
}

func NewTokenService(keyHex string, issuer, audience string) (*TokenService, error) {
    keyBytes, err := hex.DecodeString(keyHex)
    if err != nil {
        return nil, err
    }

    key, err := paseto.V4SymmetricKeyFromBytes(keyBytes)
    if err != nil {
        return nil, err
    }

    return &TokenService{
        symmetricKey: key,
        issuer:       issuer,
        audience:     audience,
        tokenTTL:     15 * time.Minute,  // Short-lived access tokens
        refreshTTL:   7 * 24 * time.Hour, // Longer-lived refresh tokens
    }, nil
}

// GenerateAccessToken creates a new access token
func (s *TokenService) GenerateAccessToken(userID, email string, roles []string) (string, error) {
    now := time.Now()

    token := paseto.NewToken()
    token.SetIssuedAt(now)
    token.SetNotBefore(now)
    token.SetExpiration(now.Add(s.tokenTTL))
    token.SetIssuer(s.issuer)
    token.SetAudience(s.audience)
    token.SetSubject(userID)

    // Custom claims
    token.SetString("email", email)
    token.Set("roles", roles)
    token.SetString("sid", generateSessionID())
    token.SetString("type", "access")

    return token.V4Encrypt(s.symmetricKey, nil), nil
}

// GenerateRefreshToken creates a new refresh token
func (s *TokenService) GenerateRefreshToken(userID string, sessionID string) (string, error) {
    now := time.Now()

    token := paseto.NewToken()
    token.SetIssuedAt(now)
    token.SetNotBefore(now)
    token.SetExpiration(now.Add(s.refreshTTL))
    token.SetIssuer(s.issuer)
    token.SetSubject(userID)
    token.SetString("sid", sessionID)
    token.SetString("type", "refresh")

    return token.V4Encrypt(s.symmetricKey, nil), nil
}

func generateSessionID() string {
    bytes := make([]byte, 16)
    rand.Read(bytes)
    return hex.EncodeToString(bytes)
}
```

### Token Validation

```go
// internal/auth/validator.go
package auth

import (
    "context"
    "errors"
    "time"

    "aidanwoods.dev/go-paseto"
)

var (
    ErrInvalidToken  = errors.New("invalid token")
    ErrExpiredToken  = errors.New("token has expired")
    ErrRevokedToken  = errors.New("token has been revoked")
    ErrInvalidIssuer = errors.New("invalid token issuer")
)

type TokenValidator struct {
    symmetricKey  paseto.V4SymmetricKey
    issuer        string
    audience      string
    revokedTokens TokenRevoker
}

type TokenRevoker interface {
    IsRevoked(ctx context.Context, tokenID string) (bool, error)
}

func (v *TokenValidator) ValidateAccessToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
    parser := paseto.NewParser()

    // Set validation rules
    parser.AddRule(paseto.IssuedBy(v.issuer))
    parser.AddRule(paseto.ForAudience(v.audience))
    parser.AddRule(paseto.NotExpired())
    parser.AddRule(paseto.ValidAt(time.Now()))

    token, err := parser.ParseV4Local(v.symmetricKey, tokenString, nil)
    if err != nil {
        if errors.Is(err, paseto.RuleError{}) {
            return nil, ErrExpiredToken
        }
        return nil, ErrInvalidToken
    }

    // Verify token type
    tokenType, err := token.GetString("type")
    if err != nil || tokenType != "access" {
        return nil, ErrInvalidToken
    }

    // Check if token is revoked
    sessionID, _ := token.GetString("sid")
    if revoked, _ := v.revokedTokens.IsRevoked(ctx, sessionID); revoked {
        return nil, ErrRevokedToken
    }

    // Extract claims
    claims := &TokenClaims{
        SessionID: sessionID,
    }

    claims.UserID, _ = token.GetSubject()
    claims.Email, _ = token.GetString("email")
    claims.Issuer, _ = token.GetIssuer()
    claims.ExpiresAt, _ = token.GetExpiration()
    claims.IssuedAt, _ = token.GetIssuedAt()

    var roles []string
    token.Get("roles", &roles)
    claims.Roles = roles

    return claims, nil
}
```

### Token Rotation

```go
// internal/auth/rotation.go
package auth

import (
    "context"
    "time"
)

type KeyRotation struct {
    currentKey  paseto.V4SymmetricKey
    previousKey paseto.V4SymmetricKey
    rotatedAt   time.Time
}

type TokenRotationService struct {
    tokenService *TokenService
    sessionStore SessionStore
}

// RefreshTokens handles token refresh with rotation
func (s *TokenRotationService) RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error) {
    // Validate refresh token
    claims, err := s.tokenService.ValidateRefreshToken(ctx, refreshToken)
    if err != nil {
        return nil, err
    }

    // Get session
    session, err := s.sessionStore.Get(ctx, claims.SessionID)
    if err != nil {
        return nil, ErrRevokedToken
    }

    // Verify refresh token matches session
    if session.RefreshTokenHash != hashToken(refreshToken) {
        // Potential token theft - revoke all sessions for user
        s.sessionStore.RevokeAllForUser(ctx, claims.UserID)
        return nil, ErrRevokedToken
    }

    // Generate new token pair
    accessToken, err := s.tokenService.GenerateAccessToken(
        claims.UserID,
        session.Email,
        session.Roles,
    )
    if err != nil {
        return nil, err
    }

    // Rotate refresh token
    newRefreshToken, err := s.tokenService.GenerateRefreshToken(
        claims.UserID,
        claims.SessionID,
    )
    if err != nil {
        return nil, err
    }

    // Update session with new refresh token hash
    session.RefreshTokenHash = hashToken(newRefreshToken)
    session.RefreshedAt = time.Now()
    s.sessionStore.Update(ctx, session)

    return &TokenPair{
        AccessToken:  accessToken,
        RefreshToken: newRefreshToken,
        ExpiresIn:    int(s.tokenService.tokenTTL.Seconds()),
    }, nil
}
```

---

## CORS Configuration

### CORS Middleware

```go
// internal/security/cors.go
package security

import (
    "net/http"
    "strings"
)

type CORSConfig struct {
    AllowedOrigins   []string
    AllowedMethods   []string
    AllowedHeaders   []string
    ExposedHeaders   []string
    AllowCredentials bool
    MaxAge           int
}

func NewCORSConfig(origins []string) *CORSConfig {
    return &CORSConfig{
        AllowedOrigins: origins,
        AllowedMethods: []string{
            http.MethodGet,
            http.MethodPost,
            http.MethodPut,
            http.MethodPatch,
            http.MethodDelete,
            http.MethodOptions,
        },
        AllowedHeaders: []string{
            "Accept",
            "Authorization",
            "Content-Type",
            "X-Request-ID",
            "X-CSRF-Token",
        },
        ExposedHeaders: []string{
            "X-Request-ID",
            "X-RateLimit-Limit",
            "X-RateLimit-Remaining",
            "X-RateLimit-Reset",
        },
        AllowCredentials: true,
        MaxAge:           86400, // 24 hours
    }
}

func CORSMiddleware(config *CORSConfig) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")

            // Check if origin is allowed
            if isOriginAllowed(origin, config.AllowedOrigins) {
                w.Header().Set("Access-Control-Allow-Origin", origin)
                w.Header().Set("Vary", "Origin")

                if config.AllowCredentials {
                    w.Header().Set("Access-Control-Allow-Credentials", "true")
                }
            }

            // Handle preflight
            if r.Method == http.MethodOptions {
                w.Header().Set("Access-Control-Allow-Methods",
                    strings.Join(config.AllowedMethods, ", "))
                w.Header().Set("Access-Control-Allow-Headers",
                    strings.Join(config.AllowedHeaders, ", "))
                w.Header().Set("Access-Control-Max-Age",
                    strconv.Itoa(config.MaxAge))
                w.WriteHeader(http.StatusNoContent)
                return
            }

            // Set exposed headers for actual requests
            w.Header().Set("Access-Control-Expose-Headers",
                strings.Join(config.ExposedHeaders, ", "))

            next.ServeHTTP(w, r)
        })
    }
}

func isOriginAllowed(origin string, allowed []string) bool {
    if origin == "" {
        return false
    }

    for _, a := range allowed {
        if a == "*" {
            return true
        }
        if a == origin {
            return true
        }
        // Support wildcard subdomains
        if strings.HasPrefix(a, "*.") {
            domain := strings.TrimPrefix(a, "*")
            if strings.HasSuffix(origin, domain) {
                return true
            }
        }
    }
    return false
}
```

### Preflight Handling

```go
// internal/security/preflight.go
package security

import (
    "net/http"
    "time"
)

// PreflightCache caches preflight responses
type PreflightCache struct {
    maxAge time.Duration
}

func (c *PreflightCache) HandlePreflight(w http.ResponseWriter, r *http.Request) {
    // Validate requested method
    requestedMethod := r.Header.Get("Access-Control-Request-Method")
    if !isMethodAllowed(requestedMethod) {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    // Validate requested headers
    requestedHeaders := r.Header.Get("Access-Control-Request-Headers")
    if !areHeadersAllowed(requestedHeaders) {
        w.WriteHeader(http.StatusForbidden)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}
```

### Environment-Specific Configuration

```go
// internal/config/cors.go
package config

func GetCORSConfig(env string) *security.CORSConfig {
    switch env {
    case "production":
        return &security.CORSConfig{
            AllowedOrigins: []string{
                "https://linkrift.io",
                "https://app.linkrift.io",
                "https://*.linkrift.io",
            },
            AllowCredentials: true,
            MaxAge:           86400,
        }
    case "staging":
        return &security.CORSConfig{
            AllowedOrigins: []string{
                "https://staging.linkrift.io",
                "https://app.staging.linkrift.io",
            },
            AllowCredentials: true,
            MaxAge:           3600,
        }
    default: // development
        return &security.CORSConfig{
            AllowedOrigins: []string{
                "http://localhost:3000",
                "http://localhost:5173",
                "http://127.0.0.1:3000",
            },
            AllowCredentials: true,
            MaxAge:           0, // No caching in dev
        }
    }
}
```

---

## Data Encryption

### Encryption at Rest

```go
// internal/security/encryption.go
package security

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "errors"
    "io"
)

type Encryptor struct {
    gcm cipher.AEAD
}

func NewEncryptor(key []byte) (*Encryptor, error) {
    if len(key) != 32 {
        return nil, errors.New("key must be 32 bytes for AES-256")
    }

    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    return &Encryptor{gcm: gcm}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM
func (e *Encryptor) Encrypt(plaintext []byte) (string, error) {
    nonce := make([]byte, e.gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }

    ciphertext := e.gcm.Seal(nonce, nonce, plaintext, nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext using AES-256-GCM
func (e *Encryptor) Decrypt(encoded string) ([]byte, error) {
    ciphertext, err := base64.StdEncoding.DecodeString(encoded)
    if err != nil {
        return nil, err
    }

    if len(ciphertext) < e.gcm.NonceSize() {
        return nil, errors.New("ciphertext too short")
    }

    nonce, ciphertext := ciphertext[:e.gcm.NonceSize()], ciphertext[e.gcm.NonceSize():]
    return e.gcm.Open(nil, nonce, ciphertext, nil)
}
```

### Encryption in Transit

```go
// internal/server/tls.go
package server

import (
    "crypto/tls"
    "net/http"
)

func NewTLSConfig() *tls.Config {
    return &tls.Config{
        MinVersion: tls.VersionTLS12,
        MaxVersion: tls.VersionTLS13,
        CipherSuites: []uint16{
            // TLS 1.3 cipher suites (automatically selected)
            tls.TLS_AES_128_GCM_SHA256,
            tls.TLS_AES_256_GCM_SHA384,
            tls.TLS_CHACHA20_POLY1305_SHA256,
            // TLS 1.2 cipher suites
            tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
            tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
            tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
            tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
        },
        PreferServerCipherSuites: true,
        CurvePreferences: []tls.CurveID{
            tls.X25519,
            tls.CurveP256,
        },
    }
}
```

### Field-Level Encryption

```go
// internal/repository/encrypted_fields.go
package repository

import (
    "context"
    "linkrift/internal/security"
)

type EncryptedUserRepository struct {
    db        *sql.DB
    encryptor *security.Encryptor
}

// CreateUser stores user with encrypted sensitive fields
func (r *EncryptedUserRepository) CreateUser(ctx context.Context, user *User) error {
    // Encrypt sensitive fields
    encryptedEmail, err := r.encryptor.Encrypt([]byte(user.Email))
    if err != nil {
        return err
    }

    encryptedPhone, err := r.encryptor.Encrypt([]byte(user.Phone))
    if err != nil {
        return err
    }

    // Store with encrypted values
    _, err = r.db.ExecContext(ctx, `
        INSERT INTO users (id, email_encrypted, email_hash, phone_encrypted, name, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `,
        user.ID,
        encryptedEmail,
        hashForLookup(user.Email), // For searching
        encryptedPhone,
        user.Name,
        user.CreatedAt,
    )

    return err
}

// GetUser retrieves and decrypts user data
func (r *EncryptedUserRepository) GetUser(ctx context.Context, id string) (*User, error) {
    var user User
    var encryptedEmail, encryptedPhone string

    err := r.db.QueryRowContext(ctx, `
        SELECT id, email_encrypted, phone_encrypted, name, created_at
        FROM users WHERE id = $1
    `, id).Scan(&user.ID, &encryptedEmail, &encryptedPhone, &user.Name, &user.CreatedAt)

    if err != nil {
        return nil, err
    }

    // Decrypt sensitive fields
    emailBytes, err := r.encryptor.Decrypt(encryptedEmail)
    if err != nil {
        return nil, err
    }
    user.Email = string(emailBytes)

    phoneBytes, err := r.encryptor.Decrypt(encryptedPhone)
    if err != nil {
        return nil, err
    }
    user.Phone = string(phoneBytes)

    return &user, nil
}
```

---

## Input Validation

### Struct Validation with go-playground/validator

```go
// internal/validation/validator.go
package validation

import (
    "regexp"
    "strings"

    "github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
    validate = validator.New(validator.WithRequiredStructEnabled())

    // Register custom validators
    validate.RegisterValidation("shortcode", validateShortCode)
    validate.RegisterValidation("safurl", validateSafeURL)
    validate.RegisterValidation("noscript", validateNoScript)
}

// CreateLinkRequest represents the request to create a new link
type CreateLinkRequest struct {
    OriginalURL  string `json:"original_url" validate:"required,url,safurl,max=2048"`
    CustomCode   string `json:"custom_code,omitempty" validate:"omitempty,shortcode,min=4,max=20"`
    ExpiresAt    string `json:"expires_at,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
    Password     string `json:"password,omitempty" validate:"omitempty,min=6,max=128"`
    MaxClicks    int    `json:"max_clicks,omitempty" validate:"omitempty,min=1,max=1000000"`
    Tags         []string `json:"tags,omitempty" validate:"omitempty,max=10,dive,min=1,max=50,noscript"`
}

// UpdateLinkRequest represents the request to update a link
type UpdateLinkRequest struct {
    OriginalURL string `json:"original_url,omitempty" validate:"omitempty,url,safurl,max=2048"`
    IsActive    *bool  `json:"is_active,omitempty"`
    ExpiresAt   string `json:"expires_at,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

// Validate validates a struct and returns user-friendly errors
func Validate(s interface{}) error {
    return validate.Struct(s)
}

// ValidationErrors extracts field-specific error messages
func ValidationErrors(err error) map[string]string {
    errors := make(map[string]string)

    if validationErrors, ok := err.(validator.ValidationErrors); ok {
        for _, e := range validationErrors {
            field := strings.ToLower(e.Field())
            switch e.Tag() {
            case "required":
                errors[field] = "This field is required"
            case "url":
                errors[field] = "Must be a valid URL"
            case "safurl":
                errors[field] = "URL contains disallowed content"
            case "shortcode":
                errors[field] = "Must contain only letters, numbers, and hyphens"
            case "min":
                errors[field] = "Value is too short"
            case "max":
                errors[field] = "Value is too long"
            default:
                errors[field] = "Invalid value"
            }
        }
    }

    return errors
}
```

### Custom Validators

```go
// internal/validation/custom.go
package validation

import (
    "net/url"
    "regexp"
    "strings"

    "github.com/go-playground/validator/v10"
)

var (
    shortCodeRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]$|^[a-zA-Z0-9]$`)
    scriptRegex    = regexp.MustCompile(`(?i)<script|javascript:|data:|vbscript:`)
)

// Blocked URL schemes and domains
var (
    blockedSchemes = []string{"javascript", "data", "vbscript", "file"}
    blockedDomains = []string{
        "bit.ly", "tinyurl.com", "t.co", // Prevent link shortener chains
        "localhost", "127.0.0.1", "0.0.0.0", // Prevent internal access
    }
)

// validateShortCode ensures short codes are URL-safe
func validateShortCode(fl validator.FieldLevel) bool {
    code := fl.Field().String()
    if code == "" {
        return true
    }
    return shortCodeRegex.MatchString(code)
}

// validateSafeURL prevents malicious URLs
func validateSafeURL(fl validator.FieldLevel) bool {
    rawURL := fl.Field().String()
    if rawURL == "" {
        return true
    }

    parsed, err := url.Parse(rawURL)
    if err != nil {
        return false
    }

    // Check scheme
    scheme := strings.ToLower(parsed.Scheme)
    for _, blocked := range blockedSchemes {
        if scheme == blocked {
            return false
        }
    }

    // Only allow http and https
    if scheme != "http" && scheme != "https" {
        return false
    }

    // Check domain
    host := strings.ToLower(parsed.Host)
    for _, blocked := range blockedDomains {
        if host == blocked || strings.HasSuffix(host, "."+blocked) {
            return false
        }
    }

    // Check for IP addresses in private ranges
    if isPrivateIP(host) {
        return false
    }

    return true
}

// validateNoScript prevents script injection
func validateNoScript(fl validator.FieldLevel) bool {
    value := fl.Field().String()
    return !scriptRegex.MatchString(value)
}

func isPrivateIP(host string) bool {
    // Remove port if present
    if idx := strings.LastIndex(host, ":"); idx != -1 {
        host = host[:idx]
    }

    // Check common private ranges
    privateRanges := []string{
        "10.", "192.168.", "172.16.", "172.17.", "172.18.", "172.19.",
        "172.20.", "172.21.", "172.22.", "172.23.", "172.24.", "172.25.",
        "172.26.", "172.27.", "172.28.", "172.29.", "172.30.", "172.31.",
    }

    for _, prefix := range privateRanges {
        if strings.HasPrefix(host, prefix) {
            return true
        }
    }

    return false
}
```

### Request Validation Middleware

```go
// internal/middleware/validation.go
package middleware

import (
    "encoding/json"
    "net/http"

    "linkrift/internal/validation"
)

// ValidateRequest validates the request body against a struct
func ValidateRequest[T any](next func(http.ResponseWriter, *http.Request, T)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req T

        // Limit body size to prevent DoS
        r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB

        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            respondError(w, http.StatusBadRequest, "Invalid JSON")
            return
        }

        if err := validation.Validate(req); err != nil {
            errors := validation.ValidationErrors(err)
            respondValidationError(w, errors)
            return
        }

        next(w, r, req)
    }
}

func respondValidationError(w http.ResponseWriter, errors map[string]string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "error":   "Validation failed",
        "details": errors,
    })
}
```

---

## SQL Injection Prevention

### Type-Safe Queries with sqlc

```yaml
# sqlc.yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "db/queries/"
    schema: "db/migrations/"
    gen:
      go:
        package: "db"
        out: "internal/db"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_prepared_queries: true
        emit_interface: true
        emit_exact_table_names: false
```

```sql
-- db/queries/links.sql
-- name: GetLinkByShortCode :one
SELECT id, short_code, original_url, user_id, created_at, expires_at, is_active
FROM links
WHERE short_code = $1 AND is_active = true AND deleted_at IS NULL
LIMIT 1;

-- name: CreateLink :one
INSERT INTO links (id, short_code, original_url, user_id, expires_at, password_hash)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListLinksByUser :many
SELECT id, short_code, original_url, created_at, click_count
FROM links
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: SearchLinks :many
SELECT id, short_code, original_url, created_at
FROM links
WHERE user_id = $1
  AND deleted_at IS NULL
  AND (
    short_code ILIKE '%' || $2 || '%'
    OR original_url ILIKE '%' || $2 || '%'
  )
ORDER BY created_at DESC
LIMIT $3;
```

### Parameterized Queries

```go
// internal/db/links.sql.go (generated by sqlc)
package db

import (
    "context"
)

const getLinkByShortCode = `-- name: GetLinkByShortCode :one
SELECT id, short_code, original_url, user_id, created_at, expires_at, is_active
FROM links
WHERE short_code = $1 AND is_active = true AND deleted_at IS NULL
LIMIT 1
`

func (q *Queries) GetLinkByShortCode(ctx context.Context, shortCode string) (Link, error) {
    row := q.db.QueryRow(ctx, getLinkByShortCode, shortCode)
    var i Link
    err := row.Scan(
        &i.ID,
        &i.ShortCode,
        &i.OriginalURL,
        &i.UserID,
        &i.CreatedAt,
        &i.ExpiresAt,
        &i.IsActive,
    )
    return i, err
}
```

### Query Safety Examples

```go
// internal/repository/links.go
package repository

import (
    "context"

    "linkrift/internal/db"
)

type LinkRepository struct {
    queries *db.Queries
}

// GetByShortCode - SAFE: Uses parameterized query
func (r *LinkRepository) GetByShortCode(ctx context.Context, code string) (*Link, error) {
    // sqlc generates type-safe parameterized query
    dbLink, err := r.queries.GetLinkByShortCode(ctx, code)
    if err != nil {
        return nil, err
    }
    return toLink(dbLink), nil
}

// Search - SAFE: sqlc handles parameter escaping
func (r *LinkRepository) Search(ctx context.Context, userID, query string, limit int32) ([]*Link, error) {
    // The search term is safely parameterized by sqlc
    dbLinks, err := r.queries.SearchLinks(ctx, db.SearchLinksParams{
        UserID: userID,
        Column2: query, // ILIKE pattern parameter
        Limit:  limit,
    })
    if err != nil {
        return nil, err
    }
    return toLinks(dbLinks), nil
}

// UNSAFE - Never do this!
// func (r *LinkRepository) UnsafeSearch(ctx context.Context, query string) ([]*Link, error) {
//     sql := fmt.Sprintf("SELECT * FROM links WHERE short_code = '%s'", query)
//     // This is vulnerable to SQL injection!
// }
```

---

## Rate Limiting

### Rate Limiting Strategy

```go
// internal/security/ratelimit.go
package security

import (
    "context"
    "time"
)

type RateLimitConfig struct {
    // Global rate limits
    GlobalRPS     int           // Requests per second globally
    GlobalBurst   int           // Burst allowance

    // Per-IP rate limits
    IPRateLimit   int           // Requests per window
    IPWindow      time.Duration // Window duration

    // Per-user rate limits
    UserRateLimit int
    UserWindow    time.Duration

    // Endpoint-specific limits
    EndpointLimits map[string]EndpointLimit
}

type EndpointLimit struct {
    RateLimit int
    Window    time.Duration
    ByUser    bool // If true, limit per user; if false, per IP
}

func DefaultRateLimitConfig() *RateLimitConfig {
    return &RateLimitConfig{
        GlobalRPS:     10000,
        GlobalBurst:   1000,
        IPRateLimit:   100,
        IPWindow:      time.Minute,
        UserRateLimit: 1000,
        UserWindow:    time.Minute,
        EndpointLimits: map[string]EndpointLimit{
            "POST /api/v1/links": {
                RateLimit: 60,
                Window:    time.Minute,
                ByUser:    true,
            },
            "POST /api/v1/auth/login": {
                RateLimit: 5,
                Window:    time.Minute,
                ByUser:    false, // By IP to prevent brute force
            },
            "GET /{shortcode}": {
                RateLimit: 1000,
                Window:    time.Minute,
                ByUser:    false,
            },
        },
    }
}
```

### Implementation

```go
// internal/security/ratelimiter.go
package security

import (
    "context"
    "fmt"
    "net/http"
    "strconv"
    "time"

    "github.com/redis/go-redis/v9"
)

type RateLimiter struct {
    redis  *redis.Client
    config *RateLimitConfig
}

func NewRateLimiter(redis *redis.Client, config *RateLimitConfig) *RateLimiter {
    return &RateLimiter{
        redis:  redis,
        config: config,
    }
}

type RateLimitResult struct {
    Allowed   bool
    Limit     int
    Remaining int
    ResetAt   time.Time
}

// CheckLimit uses sliding window rate limiting
func (rl *RateLimiter) CheckLimit(ctx context.Context, key string, limit int, window time.Duration) (*RateLimitResult, error) {
    now := time.Now()
    windowStart := now.Add(-window)

    pipe := rl.redis.Pipeline()

    // Remove old entries
    pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixNano()))

    // Count current entries
    countCmd := pipe.ZCard(ctx, key)

    // Add current request
    pipe.ZAdd(ctx, key, redis.Z{
        Score:  float64(now.UnixNano()),
        Member: now.UnixNano(),
    })

    // Set expiration
    pipe.Expire(ctx, key, window)

    _, err := pipe.Exec(ctx)
    if err != nil {
        return nil, err
    }

    count := countCmd.Val()
    remaining := limit - int(count) - 1
    if remaining < 0 {
        remaining = 0
    }

    return &RateLimitResult{
        Allowed:   count < int64(limit),
        Limit:     limit,
        Remaining: remaining,
        ResetAt:   now.Add(window),
    }, nil
}

// RateLimitMiddleware applies rate limiting to requests
func RateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx := r.Context()

            // Determine rate limit key and config
            ip := getRealIP(r)
            key := fmt.Sprintf("ratelimit:ip:%s", ip)
            limit := rl.config.IPRateLimit
            window := rl.config.IPWindow

            // Check for endpoint-specific limits
            endpoint := r.Method + " " + r.URL.Path
            if epLimit, ok := rl.config.EndpointLimits[endpoint]; ok {
                limit = epLimit.RateLimit
                window = epLimit.Window

                if epLimit.ByUser {
                    if userID := getUserID(ctx); userID != "" {
                        key = fmt.Sprintf("ratelimit:user:%s:%s", userID, endpoint)
                    }
                } else {
                    key = fmt.Sprintf("ratelimit:ip:%s:%s", ip, endpoint)
                }
            }

            result, err := rl.CheckLimit(ctx, key, limit, window)
            if err != nil {
                // Fail open on Redis errors (but log)
                next.ServeHTTP(w, r)
                return
            }

            // Set rate limit headers
            w.Header().Set("X-RateLimit-Limit", strconv.Itoa(result.Limit))
            w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
            w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))

            if !result.Allowed {
                w.Header().Set("Retry-After", strconv.FormatInt(int64(window.Seconds()), 10))
                http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

### Bypass Prevention

```go
// internal/security/ip.go
package security

import (
    "net"
    "net/http"
    "strings"
)

var trustedProxies = []string{
    "10.0.0.0/8",
    "172.16.0.0/12",
    "192.168.0.0/16",
}

// getRealIP extracts the real client IP, preventing header spoofing
func getRealIP(r *http.Request) string {
    // Get the immediate connection IP
    remoteIP, _, _ := net.SplitHostPort(r.RemoteAddr)

    // Only trust X-Forwarded-For if request came from trusted proxy
    if isTrustedProxy(remoteIP) {
        if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
            // Take the first IP that isn't a trusted proxy
            ips := strings.Split(xff, ",")
            for i := len(ips) - 1; i >= 0; i-- {
                ip := strings.TrimSpace(ips[i])
                if !isTrustedProxy(ip) {
                    return ip
                }
            }
        }

        if xri := r.Header.Get("X-Real-IP"); xri != "" {
            return xri
        }
    }

    return remoteIP
}

func isTrustedProxy(ip string) bool {
    parsedIP := net.ParseIP(ip)
    if parsedIP == nil {
        return false
    }

    for _, cidr := range trustedProxies {
        _, network, _ := net.ParseCIDR(cidr)
        if network.Contains(parsedIP) {
            return true
        }
    }

    return false
}
```

---

## Secure Headers

### Security Headers Middleware

```go
// internal/security/headers.go
package security

import (
    "net/http"
)

func SecureHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Prevent clickjacking
        w.Header().Set("X-Frame-Options", "DENY")

        // Prevent MIME type sniffing
        w.Header().Set("X-Content-Type-Options", "nosniff")

        // Enable XSS filter
        w.Header().Set("X-XSS-Protection", "1; mode=block")

        // Referrer policy
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

        // Permissions policy (formerly Feature-Policy)
        w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

        // HSTS (only in production with HTTPS)
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

        // Content Security Policy
        w.Header().Set("Content-Security-Policy", buildCSP())

        next.ServeHTTP(w, r)
    })
}

func buildCSP() string {
    directives := []string{
        "default-src 'self'",
        "script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net",
        "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
        "font-src 'self' https://fonts.gstatic.com",
        "img-src 'self' data: https:",
        "connect-src 'self' https://api.linkrift.io",
        "frame-ancestors 'none'",
        "form-action 'self'",
        "base-uri 'self'",
        "upgrade-insecure-requests",
    }

    return strings.Join(directives, "; ")
}
```

### Content Security Policy

```go
// internal/security/csp.go
package security

import (
    "crypto/rand"
    "encoding/base64"
    "net/http"
)

type CSPConfig struct {
    DefaultSrc     []string
    ScriptSrc      []string
    StyleSrc       []string
    ImgSrc         []string
    ConnectSrc     []string
    FontSrc        []string
    FrameAncestors []string
    ReportURI      string
    UseNonce       bool
}

func DefaultCSPConfig() *CSPConfig {
    return &CSPConfig{
        DefaultSrc:     []string{"'self'"},
        ScriptSrc:      []string{"'self'"},
        StyleSrc:       []string{"'self'", "'unsafe-inline'"},
        ImgSrc:         []string{"'self'", "data:", "https:"},
        ConnectSrc:     []string{"'self'"},
        FontSrc:        []string{"'self'"},
        FrameAncestors: []string{"'none'"},
        ReportURI:      "/api/csp-report",
        UseNonce:       true,
    }
}

type CSPMiddleware struct {
    config *CSPConfig
}

func (m *CSPMiddleware) Handler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        nonce := ""
        if m.config.UseNonce {
            nonce = generateNonce()
            // Store nonce in context for templates
            ctx := context.WithValue(r.Context(), "csp-nonce", nonce)
            r = r.WithContext(ctx)
        }

        csp := m.buildPolicy(nonce)
        w.Header().Set("Content-Security-Policy", csp)

        next.ServeHTTP(w, r)
    })
}

func generateNonce() string {
    bytes := make([]byte, 16)
    rand.Read(bytes)
    return base64.StdEncoding.EncodeToString(bytes)
}

func (m *CSPMiddleware) buildPolicy(nonce string) string {
    var builder strings.Builder

    // Default source
    builder.WriteString("default-src ")
    builder.WriteString(strings.Join(m.config.DefaultSrc, " "))
    builder.WriteString("; ")

    // Script source with nonce
    builder.WriteString("script-src ")
    sources := m.config.ScriptSrc
    if nonce != "" {
        sources = append(sources, "'nonce-"+nonce+"'")
    }
    builder.WriteString(strings.Join(sources, " "))
    builder.WriteString("; ")

    // Add other directives...

    if m.config.ReportURI != "" {
        builder.WriteString("report-uri ")
        builder.WriteString(m.config.ReportURI)
    }

    return builder.String()
}
```

---

## gosec Integration

### Configuration

```yaml
# .gosec.yaml
global:
  audit: enabled
  nosec: false
  exclude-generated: true

rules:
  # Include all rules by default
  includes:
    - G101 # Look for hardcoded credentials
    - G102 # Bind to all interfaces
    - G103 # Audit the use of unsafe block
    - G104 # Audit errors not checked
    - G106 # Audit the use of ssh.InsecureIgnoreHostKey
    - G107 # Url provided to HTTP request as taint input
    - G108 # Profiling endpoint automatically exposed
    - G109 # Potential Integer overflow
    - G110 # Potential DoS via decompression bomb
    - G201 # SQL query construction using format string
    - G202 # SQL query construction using string concatenation
    - G203 # Use of unescaped data in HTML templates
    - G204 # Audit use of command execution
    - G301 # Poor file permissions used when creating a directory
    - G302 # Poor file permissions used with chmod
    - G303 # Creating tempfile using a predictable path
    - G304 # File path provided as taint input
    - G305 # File traversal when extracting zip archive
    - G306 # Poor file permissions used when writing to a new file
    - G307 # Poor file permissions used when creating a file with os.Create
    - G401 # Detect the usage of DES, RC4, MD5 or SHA1
    - G402 # Look for bad TLS connection settings
    - G403 # Ensure minimum RSA key length of 2048 bits
    - G404 # Insecure random number source (rand)
    - G501 # Import blocklist: crypto/md5
    - G502 # Import blocklist: crypto/des
    - G503 # Import blocklist: crypto/rc4
    - G504 # Import blocklist: net/http/cgi
    - G505 # Import blocklist: crypto/sha1
    - G601 # Implicit memory aliasing in for loop

severity: medium
confidence: medium
```

### CI/CD Integration

```yaml
# .github/workflows/security.yml
name: Security Scan

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  gosec:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Install gosec
        run: go install github.com/securego/gosec/v2/cmd/gosec@latest

      - name: Run gosec
        run: gosec -fmt=sarif -out=gosec.sarif ./...

      - name: Upload SARIF file
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: gosec.sarif

  govulncheck:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Install govulncheck
        run: go install golang.org/x/vuln/cmd/govulncheck@latest

      - name: Run govulncheck
        run: govulncheck ./...

  trivy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          severity: 'CRITICAL,HIGH'
```

### Handling Findings

```go
// Example: Fixing a G104 (error not checked) finding

// BEFORE (flagged by gosec)
func processFile(path string) {
    file, _ := os.Open(path)  // G104: Error not checked
    defer file.Close()
    // process file
}

// AFTER (fixed)
func processFile(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return fmt.Errorf("failed to open file: %w", err)
    }
    defer file.Close()
    // process file
    return nil
}

// Example: Using #nosec when risk is acceptable and documented
func generateShortCode() string {
    // Using math/rand is acceptable here as we don't need
    // cryptographic randomness for URL short codes.
    // The collision resistance comes from the retry logic.
    // #nosec G404 -- Not used for security-sensitive operations
    return randomString(6)
}
```

---

## Additional Security Measures

### API Key Management

```go
// internal/auth/apikey.go
package auth

import (
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
)

const (
    APIKeyPrefix = "lr_"
    APIKeyLength = 32
)

// GenerateAPIKey creates a new API key
func GenerateAPIKey() (plaintext string, hash string, err error) {
    bytes := make([]byte, APIKeyLength)
    if _, err := rand.Read(bytes); err != nil {
        return "", "", err
    }

    plaintext = APIKeyPrefix + hex.EncodeToString(bytes)
    hash = hashAPIKey(plaintext)

    return plaintext, hash, nil
}

// hashAPIKey creates a secure hash of the API key for storage
func hashAPIKey(key string) string {
    hash := sha256.Sum256([]byte(key))
    return hex.EncodeToString(hash[:])
}

// ValidateAPIKey checks if an API key is valid
func (s *Service) ValidateAPIKey(ctx context.Context, key string) (*APIKeyInfo, error) {
    if !strings.HasPrefix(key, APIKeyPrefix) {
        return nil, ErrInvalidAPIKey
    }

    hash := hashAPIKey(key)
    info, err := s.repo.GetAPIKeyByHash(ctx, hash)
    if err != nil {
        return nil, ErrInvalidAPIKey
    }

    if info.RevokedAt != nil {
        return nil, ErrRevokedAPIKey
    }

    if info.ExpiresAt != nil && time.Now().After(*info.ExpiresAt) {
        return nil, ErrExpiredAPIKey
    }

    return info, nil
}
```

### Audit Logging

```go
// internal/audit/logger.go
package audit

import (
    "context"
    "encoding/json"
    "time"

    "go.uber.org/zap"
)

type AuditEvent struct {
    Timestamp   time.Time              `json:"timestamp"`
    EventType   string                 `json:"event_type"`
    UserID      string                 `json:"user_id,omitempty"`
    IP          string                 `json:"ip"`
    UserAgent   string                 `json:"user_agent"`
    Resource    string                 `json:"resource"`
    Action      string                 `json:"action"`
    Status      string                 `json:"status"`
    Details     map[string]interface{} `json:"details,omitempty"`
    RequestID   string                 `json:"request_id"`
}

type AuditLogger struct {
    logger *zap.Logger
}

func (a *AuditLogger) Log(ctx context.Context, event AuditEvent) {
    event.Timestamp = time.Now()

    data, _ := json.Marshal(event)
    a.logger.Info("audit",
        zap.String("event", string(data)),
    )
}

// LogSecurityEvent logs security-relevant events
func (a *AuditLogger) LogSecurityEvent(ctx context.Context, eventType string, details map[string]interface{}) {
    a.Log(ctx, AuditEvent{
        EventType: eventType,
        UserID:    getUserIDFromContext(ctx),
        IP:        getIPFromContext(ctx),
        RequestID: getRequestIDFromContext(ctx),
        Details:   details,
        Status:    "logged",
    })
}
```

---

## Security Checklist

### Pre-Deployment Checklist

- [ ] All secrets stored in environment variables or secret manager
- [ ] TLS 1.2+ enforced for all connections
- [ ] HSTS header configured with appropriate max-age
- [ ] CORS configured with specific allowed origins (not *)
- [ ] Rate limiting enabled on all endpoints
- [ ] Input validation on all user inputs
- [ ] SQL injection prevention verified (using sqlc/parameterized queries)
- [ ] XSS prevention headers configured
- [ ] CSRF protection enabled for state-changing operations
- [ ] Sensitive data encrypted at rest
- [ ] Audit logging enabled for security events
- [ ] Error messages don't leak sensitive information
- [ ] Dependencies scanned for vulnerabilities
- [ ] gosec scan passing with no high-severity findings

### Ongoing Security Tasks

- [ ] Weekly dependency vulnerability scans
- [ ] Monthly security log review
- [ ] Quarterly penetration testing
- [ ] Annual security audit
- [ ] Regular backup testing
- [ ] Incident response plan review

---

## Related Documentation

- [LEGAL_COMPLIANCE.md](./LEGAL_COMPLIANCE.md) - Compliance requirements
- [../operations/MONITORING_LOGGING.md](../operations/MONITORING_LOGGING.md) - Monitoring and alerting
- [../operations/TROUBLESHOOTING.md](../operations/TROUBLESHOOTING.md) - Security incident response
