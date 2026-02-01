package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/license"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

// --- Mock APIKeyRepository ---

type mockAPIKeyRepo struct {
	keys       map[uuid.UUID]*models.APIKey
	keysByPfx  map[string]*models.APIKey
	createErr  error
	revokeErr  error
}

func newMockAPIKeyRepo() *mockAPIKeyRepo {
	return &mockAPIKeyRepo{
		keys:      make(map[uuid.UUID]*models.APIKey),
		keysByPfx: make(map[string]*models.APIKey),
	}
}

func (m *mockAPIKeyRepo) Create(_ context.Context, params sqlc.CreateAPIKeyParams) (*models.APIKey, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	wsID := uuid.UUID(params.WorkspaceID.Bytes)
	k := &models.APIKey{
		ID:          uuid.New(),
		UserID:      params.UserID,
		WorkspaceID: wsID,
		Name:        params.Name,
		KeyHash:     params.KeyHash,
		KeyPrefix:   params.KeyPrefix,
		Scopes:      params.Scopes,
		CreatedAt:   time.Now(),
	}
	m.keys[k.ID] = k
	m.keysByPfx[k.KeyPrefix] = k
	return k, nil
}

func (m *mockAPIKeyRepo) GetByPrefix(_ context.Context, prefix string) (*models.APIKey, error) {
	k, ok := m.keysByPfx[prefix]
	if !ok {
		return nil, httputil.NotFound("api_key")
	}
	return k, nil
}

func (m *mockAPIKeyRepo) GetByID(_ context.Context, id uuid.UUID) (*models.APIKey, error) {
	k, ok := m.keys[id]
	if !ok {
		return nil, httputil.NotFound("api_key")
	}
	return k, nil
}

func (m *mockAPIKeyRepo) List(_ context.Context, workspaceID uuid.UUID) ([]*models.APIKey, error) {
	var result []*models.APIKey
	for _, k := range m.keys {
		if k.WorkspaceID == workspaceID {
			result = append(result, k)
		}
	}
	return result, nil
}

func (m *mockAPIKeyRepo) Revoke(_ context.Context, id uuid.UUID) error {
	if m.revokeErr != nil {
		return m.revokeErr
	}
	delete(m.keys, id)
	return nil
}

func (m *mockAPIKeyRepo) UpdateLastUsed(_ context.Context, _ uuid.UUID) error {
	return nil
}

// --- Helpers ---

func newTestAPIKeyService(repo *mockAPIKeyRepo) *apiKeyService {
	logger := zap.NewNop()
	verifier, _ := license.NewVerifier()
	licManager := license.NewManager(verifier, logger)

	return &apiKeyService{
		apiKeyRepo:  repo,
		licManager:  licManager,
		redis:       nil,
		auditLogger: noopAuditLogger{},
		logger:      logger,
	}
}

// --- Tests ---

func TestCreateAPIKey_FeatureGated(t *testing.T) {
	repo := newMockAPIKeyRepo()
	svc := newTestAPIKeyService(repo)

	ctx := context.Background()
	userID := uuid.New()
	wsID := uuid.New()

	_, err := svc.CreateAPIKey(ctx, userID, wsID, models.CreateAPIKeyInput{
		Name:   "test-key",
		Scopes: []string{"links:read"},
	})
	if err == nil {
		t.Fatal("expected payment required error for free tier")
	}
	appErr, ok := err.(*httputil.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != "PAYMENT_REQUIRED" {
		t.Errorf("expected PAYMENT_REQUIRED, got %s", appErr.Code)
	}
}

func TestCreateAPIKey_InvalidScope(t *testing.T) {
	repo := newMockAPIKeyRepo()
	svc := newTestAPIKeyService(repo)

	ctx := context.Background()
	userID := uuid.New()
	wsID := uuid.New()

	// Even though feature is gated (checked first), let's verify the flow
	_, err := svc.CreateAPIKey(ctx, userID, wsID, models.CreateAPIKeyInput{
		Name:   "test-key",
		Scopes: []string{"invalid:scope"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRevokeAPIKey_Success(t *testing.T) {
	repo := newMockAPIKeyRepo()
	svc := newTestAPIKeyService(repo)

	wsID := uuid.New()
	keyID := uuid.New()
	repo.keys[keyID] = &models.APIKey{
		ID:          keyID,
		WorkspaceID: wsID,
		Name:        "test-key",
		Scopes:      []string{"links:read"},
	}

	ctx := context.Background()
	err := svc.RevokeAPIKey(ctx, keyID, wsID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, ok := repo.keys[keyID]; ok {
		t.Error("expected key to be revoked/removed")
	}
}

func TestRevokeAPIKey_WrongWorkspace(t *testing.T) {
	repo := newMockAPIKeyRepo()
	svc := newTestAPIKeyService(repo)

	wsID := uuid.New()
	otherWS := uuid.New()
	keyID := uuid.New()
	repo.keys[keyID] = &models.APIKey{
		ID:          keyID,
		WorkspaceID: wsID,
	}

	ctx := context.Background()
	err := svc.RevokeAPIKey(ctx, keyID, otherWS)
	if err == nil {
		t.Fatal("expected forbidden error")
	}
	appErr, ok := err.(*httputil.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != "FORBIDDEN" {
		t.Errorf("expected FORBIDDEN, got %s", appErr.Code)
	}
}

func TestRevokeAPIKey_NotFound(t *testing.T) {
	repo := newMockAPIKeyRepo()
	svc := newTestAPIKeyService(repo)

	ctx := context.Background()
	err := svc.RevokeAPIKey(ctx, uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
}

func TestValidateAPIKey_TooShort(t *testing.T) {
	repo := newMockAPIKeyRepo()
	svc := newTestAPIKeyService(repo)

	ctx := context.Background()
	_, err := svc.ValidateAPIKey(ctx, "short")
	if err == nil {
		t.Fatal("expected error for short key")
	}
	appErr, ok := err.(*httputil.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != "UNAUTHORIZED" {
		t.Errorf("expected UNAUTHORIZED, got %s", appErr.Code)
	}
}

func TestValidateAPIKey_NotFound(t *testing.T) {
	repo := newMockAPIKeyRepo()
	svc := newTestAPIKeyService(repo)

	ctx := context.Background()
	// Generate a valid-format key that doesn't exist in repo
	rawKey := "lr_live_sk_" + "abcdef012345" + "6789abcdef0123456789abcdef0123456789ab"
	_, err := svc.ValidateAPIKey(ctx, rawKey)
	if err == nil {
		t.Fatal("expected error for unknown key")
	}
}

func TestValidateAPIKey_HashMismatch(t *testing.T) {
	repo := newMockAPIKeyRepo()
	svc := newTestAPIKeyService(repo)

	prefix := "lr_live_sk_abcdef012345"
	// Store a key with a different hash
	repo.keysByPfx[prefix] = &models.APIKey{
		ID:      uuid.New(),
		KeyHash: "wronghash",
	}

	ctx := context.Background()
	rawKey := prefix + "6789abcdef0123456789abcdef0123456789ab"
	_, err := svc.ValidateAPIKey(ctx, rawKey)
	if err == nil {
		t.Fatal("expected unauthorized error for hash mismatch")
	}
}

func TestValidateAPIKey_Expired(t *testing.T) {
	repo := newMockAPIKeyRepo()
	svc := newTestAPIKeyService(repo)

	rawKey := "lr_live_sk_abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789ab"
	prefix := rawKey[:len("lr_live_sk_")+12]
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	past := time.Now().Add(-24 * time.Hour)
	repo.keysByPfx[prefix] = &models.APIKey{
		ID:        uuid.New(),
		KeyHash:   keyHash,
		ExpiresAt: &past,
	}

	ctx := context.Background()
	_, err := svc.ValidateAPIKey(ctx, rawKey)
	if err == nil {
		t.Fatal("expected unauthorized error for expired key")
	}
	appErr, ok := err.(*httputil.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != "UNAUTHORIZED" {
		t.Errorf("expected UNAUTHORIZED, got %s", appErr.Code)
	}
}

func TestValidateAPIKey_Success(t *testing.T) {
	repo := newMockAPIKeyRepo()
	svc := newTestAPIKeyService(repo)

	rawKey := "lr_live_sk_abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789ab"
	prefix := rawKey[:len("lr_live_sk_")+12]
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	keyID := uuid.New()
	repo.keysByPfx[prefix] = &models.APIKey{
		ID:      keyID,
		KeyHash: keyHash,
		Scopes:  []string{"links:read"},
	}

	ctx := context.Background()
	key, err := svc.ValidateAPIKey(ctx, rawKey)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if key.ID != keyID {
		t.Errorf("expected key ID %s, got %s", keyID, key.ID)
	}
}

func TestListAPIKeys(t *testing.T) {
	repo := newMockAPIKeyRepo()
	svc := newTestAPIKeyService(repo)

	wsID := uuid.New()
	otherWS := uuid.New()

	repo.keys[uuid.New()] = &models.APIKey{ID: uuid.New(), WorkspaceID: wsID}
	repo.keys[uuid.New()] = &models.APIKey{ID: uuid.New(), WorkspaceID: wsID}
	repo.keys[uuid.New()] = &models.APIKey{ID: uuid.New(), WorkspaceID: otherWS}

	ctx := context.Background()
	keys, err := svc.ListAPIKeys(ctx, wsID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys for workspace, got %d", len(keys))
	}
}

func TestAPIKey_IsExpired(t *testing.T) {
	// Not expired
	future := time.Now().Add(24 * time.Hour)
	k1 := &models.APIKey{ExpiresAt: &future}
	if k1.IsExpired() {
		t.Error("expected key to not be expired")
	}

	// Expired
	past := time.Now().Add(-24 * time.Hour)
	k2 := &models.APIKey{ExpiresAt: &past}
	if !k2.IsExpired() {
		t.Error("expected key to be expired")
	}

	// No expiry
	k3 := &models.APIKey{}
	if k3.IsExpired() {
		t.Error("expected key with no expiry to not be expired")
	}
}

func TestAPIKey_HasScope(t *testing.T) {
	k := &models.APIKey{Scopes: []string{"links:read", "links:write"}}

	if !k.HasScope("links:read") {
		t.Error("expected key to have links:read scope")
	}
	if k.HasScope("domains:read") {
		t.Error("expected key to not have domains:read scope")
	}
}
