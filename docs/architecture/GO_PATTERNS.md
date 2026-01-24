# Go Patterns

> Last Updated: 2025-01-24

Go-specific patterns, idioms, and best practices used in Linkrift.

## Table of Contents

- [Project Layout](#project-layout)
- [Dependency Injection](#dependency-injection)
- [Error Handling](#error-handling)
- [Repository Pattern](#repository-pattern)
- [Service Pattern](#service-pattern)
- [Handler Pattern](#handler-pattern)
- [Middleware Pattern](#middleware-pattern)
- [Configuration](#configuration)
- [Graceful Shutdown](#graceful-shutdown)
- [Concurrency Patterns](#concurrency-patterns)
- [Testing Patterns](#testing-patterns)

---

## Project Layout

Following the [Standard Go Project Layout](https://github.com/golang-standards/project-layout):

```
├── cmd/                    # Main applications
│   ├── api/main.go         # API server entry point
│   ├── redirect/main.go    # Redirect service entry point
│   └── worker/main.go      # Worker entry point
│
├── internal/               # Private application code
│   ├── config/             # Configuration
│   ├── database/           # Database connections
│   ├── handler/            # HTTP handlers
│   ├── middleware/         # HTTP middleware
│   ├── models/             # Domain models
│   ├── repository/         # Data access
│   ├── service/            # Business logic
│   └── worker/             # Background jobs
│
├── pkg/                    # Public libraries
│   ├── httputil/           # HTTP utilities
│   ├── paseto/             # Token handling
│   ├── shortcode/          # Short code generation
│   └── validator/          # Custom validators
│
└── go.mod
```

### Key Principles

1. **cmd/** — Minimal main packages, just wire dependencies
2. **internal/** — Cannot be imported by external packages
3. **pkg/** — Stable APIs, can be imported externally

---

## Dependency Injection

### Constructor Injection

```go
// internal/service/link_service.go
package service

type LinkService interface {
    Create(ctx context.Context, input CreateLinkInput) (*models.Link, error)
    GetByCode(ctx context.Context, code string) (*models.Link, error)
    List(ctx context.Context, opts ListOptions) ([]*models.Link, error)
}

type linkService struct {
    repo   repository.LinkRepository
    cache  cache.Cache
    logger *zap.Logger
}

// Constructor with explicit dependencies
func NewLinkService(
    repo repository.LinkRepository,
    cache cache.Cache,
    logger *zap.Logger,
) LinkService {
    return &linkService{
        repo:   repo,
        cache:  cache,
        logger: logger,
    }
}
```

### Wire-style DI (Manual)

```go
// cmd/api/main.go
package main

func main() {
    // Load configuration
    cfg := config.Load()

    // Initialize logger
    logger := zap.Must(zap.NewProduction())
    defer logger.Sync()

    // Initialize database
    db, err := database.NewPostgres(cfg.DatabaseURL)
    if err != nil {
        logger.Fatal("failed to connect to database", zap.Error(err))
    }
    defer db.Close()

    // Initialize Redis
    redis := database.NewRedis(cfg.RedisURL)
    defer redis.Close()

    // Initialize repositories
    userRepo := repository.NewUserRepository(db, logger)
    linkRepo := repository.NewLinkRepository(db, logger)

    // Initialize cache
    cache := cache.NewRedisCache(redis)

    // Initialize services
    authService := service.NewAuthService(userRepo, cache, cfg, logger)
    linkService := service.NewLinkService(linkRepo, cache, logger)

    // Initialize handlers
    authHandler := handler.NewAuthHandler(authService, logger)
    linkHandler := handler.NewLinkHandler(linkService, logger)

    // Setup router
    router := setupRouter(authHandler, linkHandler, cfg, logger)

    // Start server
    server := &http.Server{
        Addr:    ":" + cfg.Port,
        Handler: router,
    }

    // Graceful shutdown...
}
```

### Options Pattern (for complex initialization)

```go
// pkg/httputil/client.go
package httputil

type ClientOption func(*Client)

func WithTimeout(d time.Duration) ClientOption {
    return func(c *Client) {
        c.timeout = d
    }
}

func WithRetries(n int) ClientOption {
    return func(c *Client) {
        c.maxRetries = n
    }
}

func WithLogger(l *zap.Logger) ClientOption {
    return func(c *Client) {
        c.logger = l
    }
}

type Client struct {
    httpClient *http.Client
    timeout    time.Duration
    maxRetries int
    logger     *zap.Logger
}

func NewClient(opts ...ClientOption) *Client {
    c := &Client{
        httpClient: &http.Client{},
        timeout:    30 * time.Second,
        maxRetries: 3,
        logger:     zap.NewNop(),
    }

    for _, opt := range opts {
        opt(c)
    }

    c.httpClient.Timeout = c.timeout
    return c
}
```

---

## Error Handling

### Custom Error Types

```go
// internal/errors/errors.go
package errors

import (
    "errors"
    "fmt"
)

// Sentinel errors
var (
    ErrNotFound        = errors.New("not found")
    ErrAlreadyExists   = errors.New("already exists")
    ErrUnauthorized    = errors.New("unauthorized")
    ErrForbidden       = errors.New("forbidden")
    ErrValidation      = errors.New("validation error")
    ErrRateLimited     = errors.New("rate limited")
)

// Structured error with context
type AppError struct {
    Err     error
    Message string
    Code    string
    Details map[string]any
}

func (e *AppError) Error() string {
    if e.Message != "" {
        return e.Message
    }
    return e.Err.Error()
}

func (e *AppError) Unwrap() error {
    return e.Err
}

// Constructors
func NotFound(resource string) *AppError {
    return &AppError{
        Err:     ErrNotFound,
        Message: fmt.Sprintf("%s not found", resource),
        Code:    "NOT_FOUND",
    }
}

func Validation(field, message string) *AppError {
    return &AppError{
        Err:     ErrValidation,
        Message: message,
        Code:    "VALIDATION_ERROR",
        Details: map[string]any{"field": field},
    }
}

func Wrap(err error, message string) *AppError {
    return &AppError{
        Err:     err,
        Message: message,
    }
}
```

### Error Wrapping

```go
// internal/repository/link_repo.go
func (r *linkRepository) GetByCode(ctx context.Context, code string) (*models.Link, error) {
    link, err := r.db.GetLinkByCode(ctx, code)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, errors.NotFound("link")
        }
        return nil, fmt.Errorf("failed to get link by code: %w", err)
    }
    return link, nil
}
```

### Error Handling in Handlers

```go
// internal/handler/error.go
func (h *Handler) handleError(c *gin.Context, err error) {
    var appErr *errors.AppError
    if errors.As(err, &appErr) {
        status := mapErrorToStatus(appErr.Err)
        c.JSON(status, gin.H{
            "error":   appErr.Code,
            "message": appErr.Message,
            "details": appErr.Details,
        })
        return
    }

    // Log unexpected errors
    h.logger.Error("unexpected error", zap.Error(err))

    c.JSON(http.StatusInternalServerError, gin.H{
        "error":   "INTERNAL_ERROR",
        "message": "An unexpected error occurred",
    })
}

func mapErrorToStatus(err error) int {
    switch {
    case errors.Is(err, errors.ErrNotFound):
        return http.StatusNotFound
    case errors.Is(err, errors.ErrAlreadyExists):
        return http.StatusConflict
    case errors.Is(err, errors.ErrUnauthorized):
        return http.StatusUnauthorized
    case errors.Is(err, errors.ErrForbidden):
        return http.StatusForbidden
    case errors.Is(err, errors.ErrValidation):
        return http.StatusBadRequest
    case errors.Is(err, errors.ErrRateLimited):
        return http.StatusTooManyRequests
    default:
        return http.StatusInternalServerError
    }
}
```

---

## Repository Pattern

### Interface Definition

```go
// internal/repository/interfaces.go
package repository

type LinkRepository interface {
    Create(ctx context.Context, link *models.Link) error
    GetByID(ctx context.Context, id uuid.UUID) (*models.Link, error)
    GetByCode(ctx context.Context, code string) (*models.Link, error)
    List(ctx context.Context, opts ListOptions) ([]*models.Link, int64, error)
    Update(ctx context.Context, link *models.Link) error
    Delete(ctx context.Context, id uuid.UUID) error
    IncrementClicks(ctx context.Context, id uuid.UUID) error
}

type ListOptions struct {
    UserID      uuid.UUID
    WorkspaceID uuid.UUID
    Search      string
    DomainID    *uuid.UUID
    IsActive    *bool
    Limit       int
    Offset      int
    OrderBy     string
    OrderDir    string
}
```

### Implementation with sqlc

```go
// internal/repository/link_repo.go
package repository

type linkRepository struct {
    db     *sqlc.Queries
    logger *zap.Logger
}

func NewLinkRepository(db *sqlc.Queries, logger *zap.Logger) LinkRepository {
    return &linkRepository{
        db:     db,
        logger: logger,
    }
}

func (r *linkRepository) Create(ctx context.Context, link *models.Link) error {
    params := sqlc.CreateLinkParams{
        ID:          link.ID,
        UserID:      link.UserID,
        WorkspaceID: link.WorkspaceID,
        URL:         link.URL,
        ShortCode:   link.ShortCode,
        Title:       sqlNullString(link.Title),
    }

    result, err := r.db.CreateLink(ctx, params)
    if err != nil {
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) && pgErr.Code == "23505" {
            return errors.AlreadyExists("short code")
        }
        return fmt.Errorf("failed to create link: %w", err)
    }

    // Update link with generated fields
    link.CreatedAt = result.CreatedAt
    link.UpdatedAt = result.UpdatedAt

    return nil
}

func (r *linkRepository) List(ctx context.Context, opts ListOptions) ([]*models.Link, int64, error) {
    // Build dynamic query with sqlc
    params := sqlc.ListLinksParams{
        WorkspaceID: opts.WorkspaceID,
        Search:      sqlNullString(opts.Search),
        DomainID:    sqlNullUUID(opts.DomainID),
        Limit:       int32(opts.Limit),
        Offset:      int32(opts.Offset),
    }

    rows, err := r.db.ListLinks(ctx, params)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to list links: %w", err)
    }

    links := make([]*models.Link, len(rows))
    var total int64
    for i, row := range rows {
        links[i] = rowToLink(row)
        total = row.TotalCount
    }

    return links, total, nil
}
```

---

## Service Pattern

### Business Logic Layer

```go
// internal/service/link_service.go
package service

type linkService struct {
    repo      repository.LinkRepository
    cache     cache.Cache
    shortcode shortcode.Generator
    logger    *zap.Logger
}

func NewLinkService(
    repo repository.LinkRepository,
    cache cache.Cache,
    logger *zap.Logger,
) LinkService {
    return &linkService{
        repo:      repo,
        cache:     cache,
        shortcode: shortcode.NewBase62Generator(),
        logger:    logger,
    }
}

type CreateLinkInput struct {
    URL         string
    CustomCode  string
    Title       string
    Description string
    ExpiresAt   *time.Time
}

func (s *linkService) Create(ctx context.Context, userID, workspaceID uuid.UUID, input CreateLinkInput) (*models.Link, error) {
    // Validate URL
    if err := validateURL(input.URL); err != nil {
        return nil, errors.Validation("url", err.Error())
    }

    // Generate or validate short code
    var code string
    if input.CustomCode != "" {
        if err := validateShortCode(input.CustomCode); err != nil {
            return nil, errors.Validation("custom_code", err.Error())
        }
        // Check availability
        existing, err := s.repo.GetByCode(ctx, input.CustomCode)
        if err != nil && !errors.Is(err, errors.ErrNotFound) {
            return nil, err
        }
        if existing != nil {
            return nil, errors.AlreadyExists("short code")
        }
        code = input.CustomCode
    } else {
        code = s.shortcode.Generate()
    }

    // Create link
    link := &models.Link{
        ID:          uuid.New(),
        UserID:      userID,
        WorkspaceID: workspaceID,
        URL:         input.URL,
        ShortCode:   code,
        Title:       input.Title,
        Description: input.Description,
        ExpiresAt:   input.ExpiresAt,
        IsActive:    true,
    }

    if err := s.repo.Create(ctx, link); err != nil {
        return nil, err
    }

    // Cache the link
    s.cacheLink(ctx, link)

    s.logger.Info("link created",
        zap.String("id", link.ID.String()),
        zap.String("code", link.ShortCode),
    )

    return link, nil
}

func (s *linkService) GetByCode(ctx context.Context, code string) (*models.Link, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("link:%s", code)
    var link models.Link
    if err := s.cache.Get(ctx, cacheKey, &link); err == nil {
        return &link, nil
    }

    // Get from database
    dbLink, err := s.repo.GetByCode(ctx, code)
    if err != nil {
        return nil, err
    }

    // Cache for next time
    s.cacheLink(ctx, dbLink)

    return dbLink, nil
}

func (s *linkService) cacheLink(ctx context.Context, link *models.Link) {
    cacheKey := fmt.Sprintf("link:%s", link.ShortCode)
    if err := s.cache.Set(ctx, cacheKey, link, 5*time.Minute); err != nil {
        s.logger.Warn("failed to cache link", zap.Error(err))
    }
}
```

---

## Handler Pattern

### HTTP Handler Structure

```go
// internal/handler/link_handler.go
package handler

type LinkHandler struct {
    linkService service.LinkService
    logger      *zap.Logger
}

func NewLinkHandler(ls service.LinkService, logger *zap.Logger) *LinkHandler {
    return &LinkHandler{
        linkService: ls,
        logger:      logger,
    }
}

// Request/Response DTOs
type CreateLinkRequest struct {
    URL        string     `json:"url" binding:"required,url"`
    CustomCode string     `json:"custom_code" binding:"omitempty,min=3,max=50,alphanum"`
    Title      string     `json:"title" binding:"omitempty,max=500"`
    ExpiresAt  *time.Time `json:"expires_at" binding:"omitempty"`
}

type LinkResponse struct {
    ID          string     `json:"id"`
    URL         string     `json:"url"`
    ShortCode   string     `json:"short_code"`
    ShortURL    string     `json:"short_url"`
    Title       string     `json:"title,omitempty"`
    TotalClicks int64      `json:"total_clicks"`
    IsActive    bool       `json:"is_active"`
    ExpiresAt   *time.Time `json:"expires_at,omitempty"`
    CreatedAt   time.Time  `json:"created_at"`
}

func (r LinkResponse) FromModel(link *models.Link, baseURL string) LinkResponse {
    return LinkResponse{
        ID:          link.ID.String(),
        URL:         link.URL,
        ShortCode:   link.ShortCode,
        ShortURL:    fmt.Sprintf("%s/%s", baseURL, link.ShortCode),
        Title:       link.Title,
        TotalClicks: link.TotalClicks,
        IsActive:    link.IsActive,
        ExpiresAt:   link.ExpiresAt,
        CreatedAt:   link.CreatedAt,
    }
}

// Handler methods
func (h *LinkHandler) Create(c *gin.Context) {
    var req CreateLinkRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Get user from context (set by auth middleware)
    userID := middleware.GetUserID(c)
    workspaceID := middleware.GetWorkspaceID(c)

    input := service.CreateLinkInput{
        URL:        req.URL,
        CustomCode: req.CustomCode,
        Title:      req.Title,
        ExpiresAt:  req.ExpiresAt,
    }

    link, err := h.linkService.Create(c.Request.Context(), userID, workspaceID, input)
    if err != nil {
        h.handleError(c, err)
        return
    }

    baseURL := h.getBaseURL(c)
    c.JSON(http.StatusCreated, LinkResponse{}.FromModel(link, baseURL))
}

func (h *LinkHandler) List(c *gin.Context) {
    workspaceID := middleware.GetWorkspaceID(c)

    opts := service.ListOptions{
        WorkspaceID: workspaceID,
        Search:      c.Query("search"),
        Limit:       getIntQuery(c, "limit", 20),
        Offset:      getIntQuery(c, "offset", 0),
    }

    links, total, err := h.linkService.List(c.Request.Context(), opts)
    if err != nil {
        h.handleError(c, err)
        return
    }

    baseURL := h.getBaseURL(c)
    response := make([]LinkResponse, len(links))
    for i, link := range links {
        response[i] = LinkResponse{}.FromModel(link, baseURL)
    }

    c.JSON(http.StatusOK, gin.H{
        "data":  response,
        "total": total,
    })
}
```

---

## Middleware Pattern

### Auth Middleware

```go
// internal/middleware/auth.go
package middleware

type AuthMiddleware struct {
    authService service.AuthService
    logger      *zap.Logger
}

func NewAuthMiddleware(as service.AuthService, logger *zap.Logger) *AuthMiddleware {
    return &AuthMiddleware{
        authService: as,
        logger:      logger,
    }
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract token
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "missing authorization header",
            })
            return
        }

        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "invalid authorization header format",
            })
            return
        }

        // Verify token
        claims, err := m.authService.VerifyToken(c.Request.Context(), parts[1])
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "invalid or expired token",
            })
            return
        }

        // Set user info in context
        c.Set("user_id", claims.UserID)
        c.Set("workspace_id", claims.WorkspaceID)
        c.Set("role", claims.Role)

        c.Next()
    }
}

// Helper functions to get values from context
func GetUserID(c *gin.Context) uuid.UUID {
    id, _ := c.Get("user_id")
    return id.(uuid.UUID)
}

func GetWorkspaceID(c *gin.Context) uuid.UUID {
    id, _ := c.Get("workspace_id")
    return id.(uuid.UUID)
}
```

### Rate Limiting Middleware

```go
// internal/middleware/ratelimit.go
package middleware

type RateLimiter struct {
    redis  *redis.Client
    logger *zap.Logger
}

func NewRateLimiter(redis *redis.Client, logger *zap.Logger) *RateLimiter {
    return &RateLimiter{
        redis:  redis,
        logger: logger,
    }
}

func (r *RateLimiter) Limit(limit int, window time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        key := fmt.Sprintf("rate:%s:%s", getClientIP(c), c.Request.URL.Path)

        // Sliding window rate limiting
        now := time.Now().UnixNano()
        windowStart := now - int64(window)

        pipe := r.redis.Pipeline()

        // Remove old entries
        pipe.ZRemRangeByScore(c.Request.Context(), key, "0", fmt.Sprint(windowStart))

        // Add current request
        pipe.ZAdd(c.Request.Context(), key, redis.Z{
            Score:  float64(now),
            Member: now,
        })

        // Count requests in window
        countCmd := pipe.ZCard(c.Request.Context(), key)

        // Set expiry
        pipe.Expire(c.Request.Context(), key, window)

        _, err := pipe.Exec(c.Request.Context())
        if err != nil {
            r.logger.Error("rate limiter error", zap.Error(err))
            c.Next()
            return
        }

        count := countCmd.Val()
        remaining := limit - int(count)
        if remaining < 0 {
            remaining = 0
        }

        // Set rate limit headers
        c.Header("X-RateLimit-Limit", fmt.Sprint(limit))
        c.Header("X-RateLimit-Remaining", fmt.Sprint(remaining))
        c.Header("X-RateLimit-Reset", fmt.Sprint(time.Now().Add(window).Unix()))

        if count > int64(limit) {
            c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
                "error":       "rate_limit_exceeded",
                "retry_after": window.Seconds(),
            })
            return
        }

        c.Next()
    }
}
```

---

## Configuration

### Viper-based Configuration

```go
// internal/config/config.go
package config

import (
    "github.com/spf13/viper"
)

type Config struct {
    App      AppConfig
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
    Auth     AuthConfig
}

type AppConfig struct {
    Env   string `mapstructure:"env"`
    Debug bool   `mapstructure:"debug"`
}

type ServerConfig struct {
    Port         string        `mapstructure:"port"`
    ReadTimeout  time.Duration `mapstructure:"read_timeout"`
    WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
    URL             string `mapstructure:"url"`
    MaxConnections  int    `mapstructure:"max_connections"`
    MaxIdleConns    int    `mapstructure:"max_idle_connections"`
    ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
    URL      string `mapstructure:"url"`
    PoolSize int    `mapstructure:"pool_size"`
}

type AuthConfig struct {
    JWTSecret        string        `mapstructure:"jwt_secret"`
    PASETOKey        string        `mapstructure:"paseto_key"`
    AccessTokenTTL   time.Duration `mapstructure:"access_token_ttl"`
    RefreshTokenTTL  time.Duration `mapstructure:"refresh_token_ttl"`
}

func Load() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")
    viper.AddConfigPath("./config")

    // Environment variable overrides
    viper.AutomaticEnv()
    viper.SetEnvPrefix("LINKRIFT")

    // Defaults
    viper.SetDefault("server.port", "8080")
    viper.SetDefault("server.read_timeout", "30s")
    viper.SetDefault("server.write_timeout", "30s")
    viper.SetDefault("database.max_connections", 25)
    viper.SetDefault("auth.access_token_ttl", "15m")
    viper.SetDefault("auth.refresh_token_ttl", "7d")

    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, fmt.Errorf("failed to read config: %w", err)
        }
    }

    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }

    return &cfg, nil
}
```

---

## Graceful Shutdown

```go
// cmd/api/main.go
func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    logger := setupLogger(cfg)
    defer logger.Sync()

    // Initialize dependencies...
    db := setupDatabase(cfg, logger)
    defer db.Close()

    router := setupRouter(cfg, logger)

    server := &http.Server{
        Addr:         ":" + cfg.Server.Port,
        Handler:      router,
        ReadTimeout:  cfg.Server.ReadTimeout,
        WriteTimeout: cfg.Server.WriteTimeout,
    }

    // Channel to listen for errors from server
    serverErrors := make(chan error, 1)

    // Start server in goroutine
    go func() {
        logger.Info("server starting", zap.String("port", cfg.Server.Port))
        serverErrors <- server.ListenAndServe()
    }()

    // Channel to listen for interrupt signals
    shutdown := make(chan os.Signal, 1)
    signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

    // Block until signal or error
    select {
    case err := <-serverErrors:
        logger.Fatal("server error", zap.Error(err))

    case sig := <-shutdown:
        logger.Info("shutdown signal received", zap.String("signal", sig.String()))

        // Create context with timeout for graceful shutdown
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        // Attempt graceful shutdown
        if err := server.Shutdown(ctx); err != nil {
            logger.Error("graceful shutdown failed", zap.Error(err))
            server.Close()
        }

        logger.Info("server stopped")
    }
}
```

---

## Concurrency Patterns

### Worker Pool

```go
// internal/worker/pool.go
package worker

type Task func(ctx context.Context) error

type Pool struct {
    tasks   chan Task
    workers int
    wg      sync.WaitGroup
    logger  *zap.Logger
}

func NewPool(workers int, logger *zap.Logger) *Pool {
    return &Pool{
        tasks:   make(chan Task, workers*10),
        workers: workers,
        logger:  logger,
    }
}

func (p *Pool) Start(ctx context.Context) {
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go p.worker(ctx, i)
    }
}

func (p *Pool) worker(ctx context.Context, id int) {
    defer p.wg.Done()

    for {
        select {
        case <-ctx.Done():
            return
        case task, ok := <-p.tasks:
            if !ok {
                return
            }
            if err := task(ctx); err != nil {
                p.logger.Error("task failed",
                    zap.Int("worker", id),
                    zap.Error(err),
                )
            }
        }
    }
}

func (p *Pool) Submit(task Task) {
    p.tasks <- task
}

func (p *Pool) Stop() {
    close(p.tasks)
    p.wg.Wait()
}
```

### sync.Pool for Object Reuse

```go
// internal/redirect/resolver.go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func (r *Resolver) buildRedirectURL(link *models.Link, req *http.Request) string {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()

    buf.WriteString(link.URL)

    // Append query params if needed
    if link.PassQueryParams && req.URL.RawQuery != "" {
        if strings.Contains(link.URL, "?") {
            buf.WriteByte('&')
        } else {
            buf.WriteByte('?')
        }
        buf.WriteString(req.URL.RawQuery)
    }

    return buf.String()
}
```

---

## Testing Patterns

### Table-Driven Tests

```go
func TestLinkService_Create(t *testing.T) {
    tests := []struct {
        name      string
        input     service.CreateLinkInput
        mockSetup func(*mocks.MockLinkRepository, *mocks.MockCache)
        wantErr   bool
        errType   error
    }{
        {
            name: "successful creation with custom code",
            input: service.CreateLinkInput{
                URL:        "https://example.com",
                CustomCode: "test123",
            },
            mockSetup: func(repo *mocks.MockLinkRepository, cache *mocks.MockCache) {
                repo.EXPECT().
                    GetByCode(gomock.Any(), "test123").
                    Return(nil, errors.ErrNotFound)
                repo.EXPECT().
                    Create(gomock.Any(), gomock.Any()).
                    Return(nil)
                cache.EXPECT().
                    Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
                    Return(nil)
            },
            wantErr: false,
        },
        {
            name: "duplicate code error",
            input: service.CreateLinkInput{
                URL:        "https://example.com",
                CustomCode: "existing",
            },
            mockSetup: func(repo *mocks.MockLinkRepository, cache *mocks.MockCache) {
                repo.EXPECT().
                    GetByCode(gomock.Any(), "existing").
                    Return(&models.Link{}, nil)
            },
            wantErr: true,
            errType: errors.ErrAlreadyExists,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            mockRepo := mocks.NewMockLinkRepository(ctrl)
            mockCache := mocks.NewMockCache(ctrl)
            tt.mockSetup(mockRepo, mockCache)

            svc := service.NewLinkService(mockRepo, mockCache, zap.NewNop())

            _, err := svc.Create(context.Background(), uuid.New(), uuid.New(), tt.input)

            if tt.wantErr {
                assert.Error(t, err)
                if tt.errType != nil {
                    assert.True(t, errors.Is(err, tt.errType))
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Integration Tests with testcontainers

```go
func TestLinkRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    ctx := context.Background()

    // Start PostgreSQL container
    pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "postgres:16-alpine",
            ExposedPorts: []string{"5432/tcp"},
            Env: map[string]string{
                "POSTGRES_USER":     "test",
                "POSTGRES_PASSWORD": "test",
                "POSTGRES_DB":       "linkrift_test",
            },
            WaitingFor: wait.ForLog("database system is ready to accept connections"),
        },
        Started: true,
    })
    require.NoError(t, err)
    defer pgContainer.Terminate(ctx)

    // Get connection string
    host, _ := pgContainer.Host(ctx)
    port, _ := pgContainer.MappedPort(ctx, "5432")
    dsn := fmt.Sprintf("postgres://test:test@%s:%s/linkrift_test?sslmode=disable", host, port.Port())

    // Run migrations
    m, err := migrate.New("file://../../migrations/postgres", dsn)
    require.NoError(t, err)
    require.NoError(t, m.Up())

    // Create repository
    db, err := pgxpool.New(ctx, dsn)
    require.NoError(t, err)
    defer db.Close()

    queries := sqlc.New(db)
    repo := repository.NewLinkRepository(queries, zap.NewNop())

    // Run tests
    t.Run("Create and GetByCode", func(t *testing.T) {
        link := &models.Link{
            ID:          uuid.New(),
            UserID:      uuid.New(),
            WorkspaceID: uuid.New(),
            URL:         "https://example.com",
            ShortCode:   "abc123",
        }

        err := repo.Create(ctx, link)
        require.NoError(t, err)

        found, err := repo.GetByCode(ctx, "abc123")
        require.NoError(t, err)
        assert.Equal(t, link.URL, found.URL)
    })
}
```

---

## Related Documentation

- [Architecture](ARCHITECTURE.md) — System design
- [Database Schema](DATABASE_SCHEMA.md) — Data model
- [Testing](../testing/TESTING.md) — Comprehensive testing
- [Development Guide](../getting-started/DEVELOPMENT_GUIDE.md) — Development workflow
