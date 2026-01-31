package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/link-rift/link-rift/internal/license"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/pkg/httputil"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const apiKeyPrefix = "lr_live_sk_"

type APIKeyService interface {
	CreateAPIKey(ctx context.Context, userID, workspaceID uuid.UUID, input models.CreateAPIKeyInput) (*models.CreateAPIKeyResponse, error)
	ListAPIKeys(ctx context.Context, workspaceID uuid.UUID) ([]*models.APIKey, error)
	RevokeAPIKey(ctx context.Context, id, workspaceID uuid.UUID) error
	ValidateAPIKey(ctx context.Context, rawKey string) (*models.APIKey, error)
	CheckRateLimit(ctx context.Context, keyID uuid.UUID) (remaining int64, err error)
}

type apiKeyService struct {
	apiKeyRepo repository.APIKeyRepository
	licManager *license.Manager
	redis      *redis.Client
	logger     *zap.Logger
}

func NewAPIKeyService(
	apiKeyRepo repository.APIKeyRepository,
	licManager *license.Manager,
	redisClient *redis.Client,
	logger *zap.Logger,
) APIKeyService {
	return &apiKeyService{
		apiKeyRepo: apiKeyRepo,
		licManager: licManager,
		redis:      redisClient,
		logger:     logger,
	}
}

func (s *apiKeyService) CreateAPIKey(ctx context.Context, userID, workspaceID uuid.UUID, input models.CreateAPIKeyInput) (*models.CreateAPIKeyResponse, error) {
	if !s.licManager.HasFeature(license.FeatureAPIAccess) {
		return nil, httputil.PaymentRequiredWithDetails(string(license.FeatureAPIAccess), "pro")
	}

	// Validate scopes
	for _, scope := range input.Scopes {
		if !models.IsValidScope(scope) {
			return nil, httputil.Validation("scopes", fmt.Sprintf("invalid scope: %s", scope))
		}
	}

	// Generate key: lr_live_sk_ + 32 random hex bytes
	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		return nil, httputil.Wrap(err, "failed to generate API key")
	}
	rawKey := apiKeyPrefix + hex.EncodeToString(rawBytes)

	// SHA-256 hash for storage
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	// 12-char prefix for lookup
	keyPrefixStr := rawKey[:len(apiKeyPrefix)+12]

	// Parse optional expiry
	var expiresAt pgtype.Timestamptz
	if input.ExpiresAt != nil && *input.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, *input.ExpiresAt)
		if err != nil {
			return nil, httputil.Validation("expires_at", "invalid date format, use RFC3339")
		}
		if t.Before(time.Now()) {
			return nil, httputil.Validation("expires_at", "expiration date must be in the future")
		}
		expiresAt = pgtype.Timestamptz{Time: t, Valid: true}
	}

	// Get rate limit from license
	limits := s.licManager.GetLimits()
	var rateLimit pgtype.Int4
	if limits.MaxAPIRequestsPerMin > 0 {
		rateLimit = pgtype.Int4{Int32: int32(limits.MaxAPIRequestsPerMin), Valid: true}
	}

	params := sqlc.CreateAPIKeyParams{
		UserID:      userID,
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		Name:        input.Name,
		KeyHash:     keyHash,
		KeyPrefix:   keyPrefixStr,
		Scopes:      input.Scopes,
		RateLimit:   rateLimit,
		ExpiresAt:   expiresAt,
	}

	key, err := s.apiKeyRepo.Create(ctx, params)
	if err != nil {
		return nil, err
	}

	return &models.CreateAPIKeyResponse{
		APIKey: key,
		Key:    rawKey,
	}, nil
}

func (s *apiKeyService) ListAPIKeys(ctx context.Context, workspaceID uuid.UUID) ([]*models.APIKey, error) {
	return s.apiKeyRepo.List(ctx, workspaceID)
}

func (s *apiKeyService) RevokeAPIKey(ctx context.Context, id, workspaceID uuid.UUID) error {
	key, err := s.apiKeyRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if key.WorkspaceID != workspaceID {
		return httputil.Forbidden("API key does not belong to this workspace")
	}
	return s.apiKeyRepo.Revoke(ctx, id)
}

func (s *apiKeyService) ValidateAPIKey(ctx context.Context, rawKey string) (*models.APIKey, error) {
	if len(rawKey) < len(apiKeyPrefix)+12 {
		return nil, httputil.Unauthorized("invalid API key format")
	}

	prefix := rawKey[:len(apiKeyPrefix)+12]

	key, err := s.apiKeyRepo.GetByPrefix(ctx, prefix)
	if err != nil {
		return nil, httputil.Unauthorized("invalid API key")
	}

	// Constant-time compare hash
	hash := sha256.Sum256([]byte(rawKey))
	providedHash := hex.EncodeToString(hash[:])
	if subtle.ConstantTimeCompare([]byte(providedHash), []byte(key.KeyHash)) != 1 {
		return nil, httputil.Unauthorized("invalid API key")
	}

	// Check expiration
	if key.IsExpired() {
		return nil, httputil.Unauthorized("API key has expired")
	}

	// Update last used (best-effort)
	go func() {
		bgCtx := context.Background()
		if err := s.apiKeyRepo.UpdateLastUsed(bgCtx, key.ID); err != nil {
			s.logger.Warn("failed to update API key last used", zap.Error(err))
		}
	}()

	return key, nil
}

func (s *apiKeyService) CheckRateLimit(ctx context.Context, keyID uuid.UUID) (int64, error) {
	key, err := s.apiKeyRepo.GetByID(ctx, keyID)
	if err != nil {
		return 0, err
	}

	if key.RateLimit == nil {
		return -1, nil // unlimited
	}

	limit := int64(*key.RateLimit)
	redisKey := fmt.Sprintf("api_rate:%s", keyID.String())

	count, err := s.redis.Incr(ctx, redisKey).Result()
	if err != nil {
		return 0, httputil.Wrap(err, "failed to check rate limit")
	}

	if count == 1 {
		s.redis.Expire(ctx, redisKey, 60*time.Second)
	}

	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	if count > limit {
		return 0, httputil.RateLimited()
	}

	return remaining, nil
}
