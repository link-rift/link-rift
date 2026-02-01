package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/config"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/pkg/crypto"
	"github.com/link-rift/link-rift/pkg/httputil"
	"github.com/link-rift/link-rift/pkg/paseto"
	"go.uber.org/zap"
)

// --- Mock UserRepository ---

type mockUserRepo struct {
	createFn          func(ctx context.Context, params sqlc.CreateUserParams) (*models.User, error)
	getByIDFn         func(ctx context.Context, id uuid.UUID) (*models.User, error)
	getByEmailFn      func(ctx context.Context, email string) (*models.User, error)
	updateFn          func(ctx context.Context, params sqlc.UpdateUserParams) (*models.User, error)
	updatePasswordFn  func(ctx context.Context, id uuid.UUID, passwordHash string) error
	setEmailVerifiedFn func(ctx context.Context, id uuid.UUID) error
	softDeleteFn      func(ctx context.Context, id uuid.UUID) error
}

func (m *mockUserRepo) Create(ctx context.Context, params sqlc.CreateUserParams) (*models.User, error) {
	if m.createFn != nil {
		return m.createFn(ctx, params)
	}
	return nil, nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return nil, nil
}

func (m *mockUserRepo) Update(ctx context.Context, params sqlc.UpdateUserParams) (*models.User, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, params)
	}
	return nil, nil
}

func (m *mockUserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	if m.updatePasswordFn != nil {
		return m.updatePasswordFn(ctx, id, passwordHash)
	}
	return nil
}

func (m *mockUserRepo) SetEmailVerified(ctx context.Context, id uuid.UUID) error {
	if m.setEmailVerifiedFn != nil {
		return m.setEmailVerifiedFn(ctx, id)
	}
	return nil
}

func (m *mockUserRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if m.softDeleteFn != nil {
		return m.softDeleteFn(ctx, id)
	}
	return nil
}

// --- Mock SessionRepository ---

type mockSessionRepo struct {
	createFn                 func(ctx context.Context, params sqlc.CreateSessionParams) (*models.Session, error)
	getByRefreshTokenHashFn  func(ctx context.Context, tokenHash string) (*models.Session, error)
	listByUserIDFn           func(ctx context.Context, userID uuid.UUID) ([]*models.Session, error)
	revokeFn                 func(ctx context.Context, id uuid.UUID) error
	revokeAllForUserFn       func(ctx context.Context, userID uuid.UUID) error
	deleteExpiredFn          func(ctx context.Context) error
}

func (m *mockSessionRepo) Create(ctx context.Context, params sqlc.CreateSessionParams) (*models.Session, error) {
	if m.createFn != nil {
		return m.createFn(ctx, params)
	}
	return &models.Session{
		ID:        uuid.New(),
		UserID:    params.UserID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil
}

func (m *mockSessionRepo) GetByRefreshTokenHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	if m.getByRefreshTokenHashFn != nil {
		return m.getByRefreshTokenHashFn(ctx, tokenHash)
	}
	return nil, httputil.NotFound("session")
}

func (m *mockSessionRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Session, error) {
	if m.listByUserIDFn != nil {
		return m.listByUserIDFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockSessionRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	if m.revokeFn != nil {
		return m.revokeFn(ctx, id)
	}
	return nil
}

func (m *mockSessionRepo) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	if m.revokeAllForUserFn != nil {
		return m.revokeAllForUserFn(ctx, userID)
	}
	return nil
}

func (m *mockSessionRepo) DeleteExpired(ctx context.Context) error {
	if m.deleteExpiredFn != nil {
		return m.deleteExpiredFn(ctx)
	}
	return nil
}

// --- Mock PasswordResetRepository ---

type mockResetRepo struct {
	createFn        func(ctx context.Context, params sqlc.CreatePasswordResetParams) (sqlc.PasswordReset, error)
	getByTokenHashFn func(ctx context.Context, tokenHash string) (sqlc.PasswordReset, error)
	markUsedFn      func(ctx context.Context, id uuid.UUID) error
	deleteExpiredFn func(ctx context.Context) error
}

func (m *mockResetRepo) Create(ctx context.Context, params sqlc.CreatePasswordResetParams) (sqlc.PasswordReset, error) {
	if m.createFn != nil {
		return m.createFn(ctx, params)
	}
	return sqlc.PasswordReset{ID: uuid.New()}, nil
}

func (m *mockResetRepo) GetByTokenHash(ctx context.Context, tokenHash string) (sqlc.PasswordReset, error) {
	if m.getByTokenHashFn != nil {
		return m.getByTokenHashFn(ctx, tokenHash)
	}
	return sqlc.PasswordReset{}, httputil.NotFound("password_reset")
}

func (m *mockResetRepo) MarkUsed(ctx context.Context, id uuid.UUID) error {
	if m.markUsedFn != nil {
		return m.markUsedFn(ctx, id)
	}
	return nil
}

func (m *mockResetRepo) DeleteExpired(ctx context.Context) error {
	if m.deleteExpiredFn != nil {
		return m.deleteExpiredFn(ctx)
	}
	return nil
}

// --- Mock TokenMaker ---

type mockTokenMaker struct {
	createTokenFn func(userID uuid.UUID, email string, sessionID uuid.UUID, duration time.Duration) (string, *paseto.Claims, error)
	verifyTokenFn func(token string) (*paseto.Claims, error)
}

func (m *mockTokenMaker) CreateToken(userID uuid.UUID, email string, sessionID uuid.UUID, duration time.Duration) (string, *paseto.Claims, error) {
	if m.createTokenFn != nil {
		return m.createTokenFn(userID, email, sessionID, duration)
	}
	return "test-access-token", &paseto.Claims{
		UserID:    userID,
		Email:     email,
		SessionID: sessionID,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(duration),
	}, nil
}

func (m *mockTokenMaker) VerifyToken(token string) (*paseto.Claims, error) {
	if m.verifyTokenFn != nil {
		return m.verifyTokenFn(token)
	}
	return nil, nil
}

// --- Helpers ---

func newTestAuthService(userRepo *mockUserRepo, sessionRepo *mockSessionRepo, resetRepo *mockResetRepo, tokenMaker *mockTokenMaker) *authService {
	logger, _ := zap.NewDevelopment()
	return &authService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		resetRepo:   resetRepo,
		tokenMaker:  tokenMaker,
		cfg: &config.Config{
			App: config.AppConfig{
				FrontendURL: "http://localhost:3000",
			},
			Auth: config.AuthConfig{
				AccessTokenExpiry:  15 * time.Minute,
				RefreshTokenExpiry: 7 * 24 * time.Hour,
			},
		},
		logger: logger,
	}
}

func makeUser(id uuid.UUID, email, name string) *models.User {
	hash, _ := crypto.HashPassword("password123")
	return &models.User{
		ID:           id,
		Email:        email,
		PasswordHash: hash,
		Name:         name,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// --- Login Tests ---

func TestLogin_Success(t *testing.T) {
	userID := uuid.New()
	user := makeUser(userID, "test@example.com", "Test User")

	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, email string) (*models.User, error) {
			if email != "test@example.com" {
				t.Errorf("expected email test@example.com, got %s", email)
			}
			return user, nil
		},
	}

	svc := newTestAuthService(userRepo, &mockSessionRepo{}, &mockResetRepo{}, &mockTokenMaker{})

	resp, err := svc.Login(context.Background(), models.LoginInput{
		Email:    "test@example.com",
		Password: "password123",
	}, "127.0.0.1", "TestAgent")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if resp.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if resp.User.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", resp.User.Email)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*models.User, error) {
			return nil, httputil.NotFound("user")
		},
	}

	svc := newTestAuthService(userRepo, &mockSessionRepo{}, &mockResetRepo{}, &mockTokenMaker{})

	_, err := svc.Login(context.Background(), models.LoginInput{
		Email:    "missing@example.com",
		Password: "password123",
	}, "", "")

	if err == nil {
		t.Fatal("expected error for missing user")
	}

	var appErr *httputil.AppError
	if !errors.As(err, &appErr) || appErr.Code != "UNAUTHORIZED" {
		t.Errorf("expected UNAUTHORIZED error, got %v", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	user := makeUser(uuid.New(), "test@example.com", "Test")

	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*models.User, error) {
			return user, nil
		},
	}

	svc := newTestAuthService(userRepo, &mockSessionRepo{}, &mockResetRepo{}, &mockTokenMaker{})

	_, err := svc.Login(context.Background(), models.LoginInput{
		Email:    "test@example.com",
		Password: "wrong-password",
	}, "", "")

	if err == nil {
		t.Fatal("expected error for wrong password")
	}

	var appErr *httputil.AppError
	if !errors.As(err, &appErr) || appErr.Code != "UNAUTHORIZED" {
		t.Errorf("expected UNAUTHORIZED error, got %v", err)
	}
}

func TestLogin_EmailNormalization(t *testing.T) {
	user := makeUser(uuid.New(), "test@example.com", "Test")
	var capturedEmail string

	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, email string) (*models.User, error) {
			capturedEmail = email
			return user, nil
		},
	}

	svc := newTestAuthService(userRepo, &mockSessionRepo{}, &mockResetRepo{}, &mockTokenMaker{})

	_, _ = svc.Login(context.Background(), models.LoginInput{
		Email:    "  TEST@Example.COM  ",
		Password: "password123",
	}, "", "")

	if capturedEmail != "test@example.com" {
		t.Errorf("expected normalized email 'test@example.com', got %q", capturedEmail)
	}
}

// --- Logout Tests ---

func TestLogout_Success(t *testing.T) {
	sessionID := uuid.New()
	revoked := false

	sessionRepo := &mockSessionRepo{
		revokeFn: func(_ context.Context, id uuid.UUID) error {
			if id != sessionID {
				t.Errorf("expected session ID %s, got %s", sessionID, id)
			}
			revoked = true
			return nil
		},
	}

	svc := newTestAuthService(&mockUserRepo{}, sessionRepo, &mockResetRepo{}, &mockTokenMaker{})

	err := svc.Logout(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !revoked {
		t.Error("session was not revoked")
	}
}

// --- GetCurrentUser Tests ---

func TestGetCurrentUser_Success(t *testing.T) {
	userID := uuid.New()
	user := makeUser(userID, "test@example.com", "Test User")

	userRepo := &mockUserRepo{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*models.User, error) {
			if id != userID {
				t.Errorf("expected user ID %s, got %s", userID, id)
			}
			return user, nil
		},
	}

	svc := newTestAuthService(userRepo, &mockSessionRepo{}, &mockResetRepo{}, &mockTokenMaker{})

	resp, err := svc.GetCurrentUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", resp.Email)
	}
	if resp.Name != "Test User" {
		t.Errorf("expected name 'Test User', got %s", resp.Name)
	}
}

func TestGetCurrentUser_NotFound(t *testing.T) {
	userRepo := &mockUserRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return nil, httputil.NotFound("user")
		},
	}

	svc := newTestAuthService(userRepo, &mockSessionRepo{}, &mockResetRepo{}, &mockTokenMaker{})

	_, err := svc.GetCurrentUser(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error for missing user")
	}
}

// --- ForgotPassword Tests ---

func TestForgotPassword_UserExists(t *testing.T) {
	user := makeUser(uuid.New(), "test@example.com", "Test")
	resetCreated := false

	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*models.User, error) {
			return user, nil
		},
	}

	resetRepo := &mockResetRepo{
		createFn: func(_ context.Context, params sqlc.CreatePasswordResetParams) (sqlc.PasswordReset, error) {
			if params.UserID != user.ID {
				t.Errorf("expected user ID %s, got %s", user.ID, params.UserID)
			}
			resetCreated = true
			return sqlc.PasswordReset{ID: uuid.New()}, nil
		},
	}

	svc := newTestAuthService(userRepo, &mockSessionRepo{}, resetRepo, &mockTokenMaker{})

	err := svc.ForgotPassword(context.Background(), models.ForgotPasswordInput{Email: "test@example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resetCreated {
		t.Error("password reset was not created")
	}
}

func TestForgotPassword_UserNotFound(t *testing.T) {
	userRepo := &mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*models.User, error) {
			return nil, httputil.NotFound("user")
		},
	}

	svc := newTestAuthService(userRepo, &mockSessionRepo{}, &mockResetRepo{}, &mockTokenMaker{})

	// Should return nil even if user not found (prevent email enumeration)
	err := svc.ForgotPassword(context.Background(), models.ForgotPasswordInput{Email: "missing@example.com"})
	if err != nil {
		t.Fatalf("expected nil error for non-existent user (prevent enumeration), got %v", err)
	}
}

// --- ResetPassword Tests ---

func TestResetPassword_Success(t *testing.T) {
	userID := uuid.New()
	resetID := uuid.New()
	passwordUpdated := false
	sessionsRevoked := false

	resetRepo := &mockResetRepo{
		getByTokenHashFn: func(_ context.Context, _ string) (sqlc.PasswordReset, error) {
			return sqlc.PasswordReset{ID: resetID, UserID: userID}, nil
		},
	}

	userRepo := &mockUserRepo{
		updatePasswordFn: func(_ context.Context, id uuid.UUID, hash string) error {
			if id != userID {
				t.Errorf("expected user ID %s, got %s", userID, id)
			}
			if hash == "" {
				t.Error("expected non-empty password hash")
			}
			passwordUpdated = true
			return nil
		},
	}

	sessionRepo := &mockSessionRepo{
		revokeAllForUserFn: func(_ context.Context, id uuid.UUID) error {
			if id != userID {
				t.Errorf("expected user ID %s, got %s", userID, id)
			}
			sessionsRevoked = true
			return nil
		},
	}

	svc := newTestAuthService(userRepo, sessionRepo, resetRepo, &mockTokenMaker{})

	err := svc.ResetPassword(context.Background(), models.ResetPasswordInput{
		Token:       "valid-reset-token",
		NewPassword: "newPassword123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !passwordUpdated {
		t.Error("password was not updated")
	}
	if !sessionsRevoked {
		t.Error("sessions were not revoked")
	}
}

func TestResetPassword_InvalidToken(t *testing.T) {
	resetRepo := &mockResetRepo{
		getByTokenHashFn: func(_ context.Context, _ string) (sqlc.PasswordReset, error) {
			return sqlc.PasswordReset{}, httputil.NotFound("password_reset")
		},
	}

	svc := newTestAuthService(&mockUserRepo{}, &mockSessionRepo{}, resetRepo, &mockTokenMaker{})

	err := svc.ResetPassword(context.Background(), models.ResetPasswordInput{
		Token:       "invalid-token",
		NewPassword: "newPassword123",
	})
	if err == nil {
		t.Fatal("expected error for invalid token")
	}

	var appErr *httputil.AppError
	if !errors.As(err, &appErr) || appErr.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %v", err)
	}
}

// --- RefreshToken Tests ---

func TestRefreshToken_Success(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	user := makeUser(userID, "test@example.com", "Test")

	sessionRepo := &mockSessionRepo{
		getByRefreshTokenHashFn: func(_ context.Context, _ string) (*models.Session, error) {
			return &models.Session{
				ID:     sessionID,
				UserID: userID,
			}, nil
		},
	}

	userRepo := &mockUserRepo{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*models.User, error) {
			return user, nil
		},
	}

	svc := newTestAuthService(userRepo, sessionRepo, &mockResetRepo{}, &mockTokenMaker{})

	resp, err := svc.RefreshToken(context.Background(), "valid-refresh-token", "127.0.0.1", "TestAgent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if resp.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	sessionRepo := &mockSessionRepo{
		getByRefreshTokenHashFn: func(_ context.Context, _ string) (*models.Session, error) {
			return nil, httputil.NotFound("session")
		},
	}

	svc := newTestAuthService(&mockUserRepo{}, sessionRepo, &mockResetRepo{}, &mockTokenMaker{})

	_, err := svc.RefreshToken(context.Background(), "invalid-token", "", "")
	if err == nil {
		t.Fatal("expected error for invalid refresh token")
	}

	var appErr *httputil.AppError
	if !errors.As(err, &appErr) || appErr.Code != "UNAUTHORIZED" {
		t.Errorf("expected UNAUTHORIZED error, got %v", err)
	}
}

// --- Helper function tests ---

func TestHashToken(t *testing.T) {
	hash1 := hashToken("test-token")
	hash2 := hashToken("test-token")
	hash3 := hashToken("different-token")

	if hash1 != hash2 {
		t.Error("same input should produce same hash")
	}
	if hash1 == hash3 {
		t.Error("different input should produce different hash")
	}
	if len(hash1) != 64 { // SHA-256 hex encoded = 64 chars
		t.Errorf("expected hash length 64, got %d", len(hash1))
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	token, hash, err := generateRefreshToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
	if hash == "" {
		t.Error("expected non-empty hash")
	}
	if hashToken(token) != hash {
		t.Error("hash should match hashToken(token)")
	}

	// Verify uniqueness
	token2, _, _ := generateRefreshToken()
	if token == token2 {
		t.Error("tokens should be unique")
	}
}

func TestGenerateWorkspaceSlug(t *testing.T) {
	slug, err := generateWorkspaceSlug("Test User")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if slug == "" {
		t.Error("expected non-empty slug")
	}
	// Should be lowercase with hyphen
	if slug[:9] != "test-user" {
		t.Errorf("expected slug to start with 'test-user', got %s", slug)
	}
}

func TestMapCreateUserError_Duplicate(t *testing.T) {
	err := mapCreateUserError(errors.New("ERROR: duplicate key value violates unique constraint \"users_email_key\" (SQLSTATE 23505)"))

	var appErr *httputil.AppError
	if !errors.As(err, &appErr) || appErr.Code != "ALREADY_EXISTS" {
		t.Errorf("expected ALREADY_EXISTS error, got %v", err)
	}
}

func TestMapCreateUserError_Other(t *testing.T) {
	err := mapCreateUserError(errors.New("connection refused"))

	var appErr *httputil.AppError
	if !errors.As(err, &appErr) || appErr.Code != "INTERNAL_ERROR" {
		t.Errorf("expected INTERNAL_ERROR from Wrap, got %v", err)
	}
}

func TestMapCreateUserError_Nil(t *testing.T) {
	err := mapCreateUserError(nil)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}
