package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/pkg/httputil"
	"github.com/link-rift/link-rift/pkg/paseto"
)

const testSecret = "12345678901234567890123456789012" // 32 chars

// --- Mock UserRepository ---

type mockUserRepo struct {
	getByIDFn    func(ctx context.Context, id uuid.UUID) (*models.User, error)
	getByEmailFn func(ctx context.Context, email string) (*models.User, error)
}

func (m *mockUserRepo) Create(_ context.Context, _ sqlc.CreateUserParams) (*models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, httputil.NotFound("user")
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return nil, httputil.NotFound("user")
}
func (m *mockUserRepo) Update(_ context.Context, _ sqlc.UpdateUserParams) (*models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) UpdatePassword(_ context.Context, _ uuid.UUID, _ string) error { return nil }
func (m *mockUserRepo) SetEmailVerified(_ context.Context, _ uuid.UUID) error               { return nil }
func (m *mockUserRepo) SoftDelete(_ context.Context, _ uuid.UUID) error                     { return nil }

// --- Mock WorkspaceRepository ---

type mockWsRepo struct {
	getByIDFn func(ctx context.Context, id uuid.UUID) (*models.Workspace, error)
}

func (m *mockWsRepo) Create(_ context.Context, _ sqlc.CreateWorkspaceParams) (*models.Workspace, error) {
	return nil, nil
}
func (m *mockWsRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Workspace, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, httputil.NotFound("workspace")
}
func (m *mockWsRepo) GetBySlug(_ context.Context, _ string) (*models.Workspace, error) {
	return nil, nil
}
func (m *mockWsRepo) ListForUser(_ context.Context, _ uuid.UUID) ([]*models.Workspace, error) {
	return nil, nil
}
func (m *mockWsRepo) Update(_ context.Context, _ sqlc.UpdateWorkspaceParams) (*models.Workspace, error) {
	return nil, nil
}
func (m *mockWsRepo) UpdateOwner(_ context.Context, _ sqlc.UpdateWorkspaceOwnerParams) (*models.Workspace, error) {
	return nil, nil
}
func (m *mockWsRepo) SoftDelete(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockWsRepo) GetCountForUser(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}

// --- Mock WorkspaceMemberRepository ---

type mockMemberRepo struct {
	getFn func(ctx context.Context, workspaceID, userID uuid.UUID) (*models.WorkspaceMember, error)
}

func (m *mockMemberRepo) Add(_ context.Context, _ sqlc.AddWorkspaceMemberParams) (*models.WorkspaceMember, error) {
	return nil, nil
}
func (m *mockMemberRepo) Get(ctx context.Context, workspaceID, userID uuid.UUID) (*models.WorkspaceMember, error) {
	if m.getFn != nil {
		return m.getFn(ctx, workspaceID, userID)
	}
	return nil, httputil.NotFound("member")
}
func (m *mockMemberRepo) List(_ context.Context, _ uuid.UUID) ([]*models.WorkspaceMemberResponse, error) {
	return nil, nil
}
func (m *mockMemberRepo) UpdateRole(_ context.Context, _ sqlc.UpdateMemberRoleParams) (*models.WorkspaceMember, error) {
	return nil, nil
}
func (m *mockMemberRepo) Remove(_ context.Context, _, _ uuid.UUID) error { return nil }
func (m *mockMemberRepo) GetCount(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}

// --- Tests ---

func init() {
	gin.SetMode(gin.TestMode)
}

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{"valid bearer", "Bearer abc123", "abc123"},
		{"bearer lowercase", "bearer abc123", "abc123"},
		{"no bearer prefix", "abc123", ""},
		{"empty header", "", ""},
		{"only bearer", "Bearer ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)
			if tt.header != "" {
				c.Request.Header.Set("Authorization", tt.header)
			}

			got := extractBearerToken(c)
			if got != tt.expected {
				t.Errorf("extractBearerToken(%q) = %q, want %q", tt.header, got, tt.expected)
			}
		})
	}
}

func TestRequireAuth_NoToken(t *testing.T) {
	tokenMaker, _ := paseto.NewPasetoMaker(testSecret)
	userRepo := &mockUserRepo{}

	r := gin.New()
	r.GET("/test", RequireAuth(tokenMaker, userRepo), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	tokenMaker, _ := paseto.NewPasetoMaker(testSecret)
	userRepo := &mockUserRepo{}

	r := gin.New()
	r.GET("/test", RequireAuth(tokenMaker, userRepo), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-here")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuth_ValidToken_UserNotFound(t *testing.T) {
	tokenMaker, _ := paseto.NewPasetoMaker(testSecret)
	userID := uuid.New()
	sessionID := uuid.New()

	token, _, err := tokenMaker.CreateToken(userID, "test@example.com", sessionID, 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	userRepo := &mockUserRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*models.User, error) {
			return nil, httputil.NotFound("user")
		},
	}

	r := gin.New()
	r.GET("/test", RequireAuth(tokenMaker, userRepo), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuth_ValidToken_UserFound(t *testing.T) {
	tokenMaker, _ := paseto.NewPasetoMaker(testSecret)
	userID := uuid.New()
	sessionID := uuid.New()

	token, _, err := tokenMaker.CreateToken(userID, "test@example.com", sessionID, 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	userRepo := &mockUserRepo{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*models.User, error) {
			return &models.User{ID: id, Email: "test@example.com", Name: "Test"}, nil
		},
	}

	var gotUser *models.User
	r := gin.New()
	r.GET("/test", RequireAuth(tokenMaker, userRepo), func(c *gin.Context) {
		gotUser = GetUserFromContext(c)
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if gotUser == nil {
		t.Fatal("expected user to be set in context")
	}
	if gotUser.ID != userID {
		t.Errorf("expected user ID %s, got %s", userID, gotUser.ID)
	}
}

func TestOptionalAuth_NoToken(t *testing.T) {
	tokenMaker, _ := paseto.NewPasetoMaker(testSecret)
	userRepo := &mockUserRepo{}

	var gotUser *models.User
	r := gin.New()
	r.GET("/test", OptionalAuth(tokenMaker, userRepo), func(c *gin.Context) {
		gotUser = GetUserFromContext(c)
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if gotUser != nil {
		t.Error("expected no user in context")
	}
}

func TestRequireWorkspaceRole_Owner(t *testing.T) {
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		c.Set("workspace_member", &models.WorkspaceMember{
			Role: models.RoleOwner,
		})
		c.Next()
	}, RequireWorkspaceRole(models.RoleAdmin), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for owner with admin requirement, got %d", w.Code)
	}
}

func TestRequireWorkspaceRole_InsufficientRole(t *testing.T) {
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		c.Set("workspace_member", &models.WorkspaceMember{
			Role: models.RoleViewer,
		})
		c.Next()
	}, RequireWorkspaceRole(models.RoleAdmin), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestRequireWorkspaceRole_NoMember(t *testing.T) {
	r := gin.New()
	r.GET("/test", RequireWorkspaceRole(models.RoleEditor), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestGetUserFromContext_NotSet(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	user := GetUserFromContext(c)
	if user != nil {
		t.Error("expected nil when user not in context")
	}
}

func TestGetWorkspaceFromContext_NotSet(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	ws := GetWorkspaceFromContext(c)
	if ws != nil {
		t.Error("expected nil when workspace not in context")
	}
}

func TestGetWorkspaceMemberFromContext_NotSet(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	member := GetWorkspaceMemberFromContext(c)
	if member != nil {
		t.Error("expected nil when member not in context")
	}
}

func TestGetSessionIDFromContext_NotSet(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	id := GetSessionIDFromContext(c)
	if id != uuid.Nil {
		t.Error("expected nil UUID when session not in context")
	}
}
