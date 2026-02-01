package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

// --- Mock AuthService ---

type mockAuthService struct {
	registerFn       func(ctx context.Context, input models.RegisterInput) (*models.AuthResponse, error)
	loginFn          func(ctx context.Context, input models.LoginInput, ip, userAgent string) (*models.AuthResponse, error)
	logoutFn         func(ctx context.Context, sessionID uuid.UUID) error
	refreshTokenFn   func(ctx context.Context, refreshToken, ip, userAgent string) (*models.AuthResponse, error)
	getCurrentUserFn func(ctx context.Context, userID uuid.UUID) (*models.UserResponse, error)
	forgotPasswordFn func(ctx context.Context, input models.ForgotPasswordInput) error
	resetPasswordFn  func(ctx context.Context, input models.ResetPasswordInput) error
	verifyEmailFn    func(ctx context.Context, input models.VerifyEmailInput) error
}

func (m *mockAuthService) Register(ctx context.Context, input models.RegisterInput) (*models.AuthResponse, error) {
	if m.registerFn != nil {
		return m.registerFn(ctx, input)
	}
	return nil, nil
}

func (m *mockAuthService) Login(ctx context.Context, input models.LoginInput, ip, userAgent string) (*models.AuthResponse, error) {
	if m.loginFn != nil {
		return m.loginFn(ctx, input, ip, userAgent)
	}
	return nil, nil
}

func (m *mockAuthService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	if m.logoutFn != nil {
		return m.logoutFn(ctx, sessionID)
	}
	return nil
}

func (m *mockAuthService) RefreshToken(ctx context.Context, refreshToken, ip, userAgent string) (*models.AuthResponse, error) {
	if m.refreshTokenFn != nil {
		return m.refreshTokenFn(ctx, refreshToken, ip, userAgent)
	}
	return nil, nil
}

func (m *mockAuthService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*models.UserResponse, error) {
	if m.getCurrentUserFn != nil {
		return m.getCurrentUserFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockAuthService) ForgotPassword(ctx context.Context, input models.ForgotPasswordInput) error {
	if m.forgotPasswordFn != nil {
		return m.forgotPasswordFn(ctx, input)
	}
	return nil
}

func (m *mockAuthService) ResetPassword(ctx context.Context, input models.ResetPasswordInput) error {
	if m.resetPasswordFn != nil {
		return m.resetPasswordFn(ctx, input)
	}
	return nil
}

func (m *mockAuthService) VerifyEmail(ctx context.Context, input models.VerifyEmailInput) error {
	if m.verifyEmailFn != nil {
		return m.verifyEmailFn(ctx, input)
	}
	return nil
}

// --- Helpers ---

func newTestAuthHandler(svc *mockAuthService) *AuthHandler {
	logger, _ := zap.NewDevelopment()
	return NewAuthHandler(svc, logger)
}

func setupAuthRouter(h *AuthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	auth := r.Group("/api/v1/auth")
	auth.POST("/login", h.Login)
	auth.POST("/register", h.Register)
	auth.POST("/forgot-password", h.ForgotPassword)
	auth.POST("/reset-password", h.ResetPassword)
	return r
}

type apiResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *apiError       `json:"error,omitempty"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// --- Login Handler Tests ---

func TestLoginHandler_Success(t *testing.T) {
	svc := &mockAuthService{
		loginFn: func(_ context.Context, input models.LoginInput, _, _ string) (*models.AuthResponse, error) {
			return &models.AuthResponse{
				AccessToken:  "test-access",
				RefreshToken: "test-refresh",
				User: &models.UserResponse{
					ID:    uuid.New(),
					Email: input.Email,
					Name:  "Test",
				},
			}, nil
		},
	}

	h := newTestAuthHandler(svc)
	router := setupAuthRouter(h)

	body, _ := json.Marshal(map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp apiResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestLoginHandler_InvalidJSON(t *testing.T) {
	h := newTestAuthHandler(&mockAuthService{})
	router := setupAuthRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestLoginHandler_Unauthorized(t *testing.T) {
	svc := &mockAuthService{
		loginFn: func(_ context.Context, _ models.LoginInput, _, _ string) (*models.AuthResponse, error) {
			return nil, httputil.Unauthorized("invalid email or password")
		},
	}

	h := newTestAuthHandler(svc)
	router := setupAuthRouter(h)

	body, _ := json.Marshal(map[string]string{
		"email":    "test@example.com",
		"password": "wrong",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}

	var resp apiResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Error == nil || resp.Error.Code != "UNAUTHORIZED" {
		t.Errorf("expected UNAUTHORIZED error code")
	}
}

// --- Register Handler Tests ---

func TestRegisterHandler_Success(t *testing.T) {
	svc := &mockAuthService{
		registerFn: func(_ context.Context, input models.RegisterInput) (*models.AuthResponse, error) {
			return &models.AuthResponse{
				AccessToken:  "new-token",
				RefreshToken: "new-refresh",
				User: &models.UserResponse{
					ID:    uuid.New(),
					Email: input.Email,
					Name:  input.Name,
				},
			}, nil
		},
	}

	h := newTestAuthHandler(svc)
	router := setupAuthRouter(h)

	body, _ := json.Marshal(map[string]string{
		"email":    "new@example.com",
		"password": "password123",
		"name":     "New User",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}
}

func TestRegisterHandler_DuplicateEmail(t *testing.T) {
	svc := &mockAuthService{
		registerFn: func(_ context.Context, _ models.RegisterInput) (*models.AuthResponse, error) {
			return nil, httputil.AlreadyExists("user")
		},
	}

	h := newTestAuthHandler(svc)
	router := setupAuthRouter(h)

	body, _ := json.Marshal(map[string]string{
		"email":    "existing@example.com",
		"password": "password123",
		"name":     "Test",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", w.Code)
	}
}

// --- ForgotPassword Handler Tests ---

func TestForgotPasswordHandler_Success(t *testing.T) {
	svc := &mockAuthService{
		forgotPasswordFn: func(_ context.Context, _ models.ForgotPasswordInput) error {
			return nil
		},
	}

	h := newTestAuthHandler(svc)
	router := setupAuthRouter(h)

	body, _ := json.Marshal(map[string]string{"email": "test@example.com"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/forgot-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// --- ResetPassword Handler Tests ---

func TestResetPasswordHandler_Success(t *testing.T) {
	svc := &mockAuthService{
		resetPasswordFn: func(_ context.Context, _ models.ResetPasswordInput) error {
			return nil
		},
	}

	h := newTestAuthHandler(svc)
	router := setupAuthRouter(h)

	body, _ := json.Marshal(map[string]string{
		"token":        "valid-token",
		"new_password": "newPassword123",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/reset-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestResetPasswordHandler_InvalidToken(t *testing.T) {
	svc := &mockAuthService{
		resetPasswordFn: func(_ context.Context, _ models.ResetPasswordInput) error {
			return httputil.Validation("token", "invalid or expired reset token")
		},
	}

	h := newTestAuthHandler(svc)
	router := setupAuthRouter(h)

	body, _ := json.Marshal(map[string]string{
		"token":        "invalid",
		"new_password": "newPassword123",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/reset-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
