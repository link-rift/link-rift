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
	"github.com/link-rift/link-rift/internal/service"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

// --- Mock WorkspaceService ---

type mockWorkspaceService struct {
	createWorkspaceFn   func(ctx context.Context, userID uuid.UUID, input models.CreateWorkspaceInput) (*models.Workspace, error)
	getWorkspaceFn      func(ctx context.Context, id uuid.UUID) (*models.Workspace, error)
	listWorkspacesFn    func(ctx context.Context, userID uuid.UUID) ([]*models.Workspace, error)
	updateWorkspaceFn   func(ctx context.Context, id uuid.UUID, input models.UpdateWorkspaceInput) (*models.Workspace, error)
	deleteWorkspaceFn   func(ctx context.Context, id uuid.UUID, actorID uuid.UUID) error
	inviteMemberFn      func(ctx context.Context, workspaceID, inviterID uuid.UUID, input models.InviteMemberInput) (*models.WorkspaceMember, error)
	removeMemberFn      func(ctx context.Context, workspaceID, actorID, targetUserID uuid.UUID) error
	updateMemberRoleFn  func(ctx context.Context, workspaceID, actorID, targetUserID uuid.UUID, input models.UpdateMemberRoleInput) (*models.WorkspaceMember, error)
	transferOwnershipFn func(ctx context.Context, workspaceID, actorID uuid.UUID, input models.TransferOwnershipInput) error
	listMembersFn       func(ctx context.Context, workspaceID uuid.UUID) ([]*models.WorkspaceMemberResponse, error)
	getMemberFn         func(ctx context.Context, workspaceID, userID uuid.UUID) (*models.WorkspaceMember, error)
	getMemberCountFn    func(ctx context.Context, workspaceID uuid.UUID) (int64, error)
}

func (m *mockWorkspaceService) CreateWorkspace(ctx context.Context, userID uuid.UUID, input models.CreateWorkspaceInput) (*models.Workspace, error) {
	if m.createWorkspaceFn != nil {
		return m.createWorkspaceFn(ctx, userID, input)
	}
	return nil, nil
}

func (m *mockWorkspaceService) GetWorkspace(ctx context.Context, id uuid.UUID) (*models.Workspace, error) {
	if m.getWorkspaceFn != nil {
		return m.getWorkspaceFn(ctx, id)
	}
	return nil, nil
}

func (m *mockWorkspaceService) ListWorkspaces(ctx context.Context, userID uuid.UUID) ([]*models.Workspace, error) {
	if m.listWorkspacesFn != nil {
		return m.listWorkspacesFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockWorkspaceService) UpdateWorkspace(ctx context.Context, id uuid.UUID, input models.UpdateWorkspaceInput) (*models.Workspace, error) {
	if m.updateWorkspaceFn != nil {
		return m.updateWorkspaceFn(ctx, id, input)
	}
	return nil, nil
}

func (m *mockWorkspaceService) DeleteWorkspace(ctx context.Context, id uuid.UUID, actorID uuid.UUID) error {
	if m.deleteWorkspaceFn != nil {
		return m.deleteWorkspaceFn(ctx, id, actorID)
	}
	return nil
}

func (m *mockWorkspaceService) InviteMember(ctx context.Context, workspaceID, inviterID uuid.UUID, input models.InviteMemberInput) (*models.WorkspaceMember, error) {
	if m.inviteMemberFn != nil {
		return m.inviteMemberFn(ctx, workspaceID, inviterID, input)
	}
	return nil, nil
}

func (m *mockWorkspaceService) RemoveMember(ctx context.Context, workspaceID, actorID, targetUserID uuid.UUID) error {
	if m.removeMemberFn != nil {
		return m.removeMemberFn(ctx, workspaceID, actorID, targetUserID)
	}
	return nil
}

func (m *mockWorkspaceService) UpdateMemberRole(ctx context.Context, workspaceID, actorID, targetUserID uuid.UUID, input models.UpdateMemberRoleInput) (*models.WorkspaceMember, error) {
	if m.updateMemberRoleFn != nil {
		return m.updateMemberRoleFn(ctx, workspaceID, actorID, targetUserID, input)
	}
	return nil, nil
}

func (m *mockWorkspaceService) TransferOwnership(ctx context.Context, workspaceID, actorID uuid.UUID, input models.TransferOwnershipInput) error {
	if m.transferOwnershipFn != nil {
		return m.transferOwnershipFn(ctx, workspaceID, actorID, input)
	}
	return nil
}

func (m *mockWorkspaceService) ListMembers(ctx context.Context, workspaceID uuid.UUID) ([]*models.WorkspaceMemberResponse, error) {
	if m.listMembersFn != nil {
		return m.listMembersFn(ctx, workspaceID)
	}
	return nil, nil
}

func (m *mockWorkspaceService) GetMember(ctx context.Context, workspaceID, userID uuid.UUID) (*models.WorkspaceMember, error) {
	if m.getMemberFn != nil {
		return m.getMemberFn(ctx, workspaceID, userID)
	}
	return nil, nil
}

func (m *mockWorkspaceService) GetMemberCount(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	if m.getMemberCountFn != nil {
		return m.getMemberCountFn(ctx, workspaceID)
	}
	return 1, nil
}

func (m *mockWorkspaceService) SetAuditLogger(_ service.AuditLogger) {}

// --- Helpers ---

var testUserID = uuid.MustParse("11111111-1111-1111-1111-111111111111")

func setupWorkspaceTestRouter(svc *mockWorkspaceService, withAuth bool, withWorkspace bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	logger, _ := zap.NewDevelopment()
	handler := NewWorkspaceHandler(svc, logger)

	authMw := func(c *gin.Context) {
		if withAuth {
			user := &models.User{
				ID:    testUserID,
				Email: "test@example.com",
				Name:  "Test User",
			}
			c.Set("user", user)
		}
		c.Next()
	}

	wsAccessMw := func(c *gin.Context) {
		if withWorkspace {
			ws := &models.Workspace{
				ID:      testWorkspaceID,
				Name:    "Test Workspace",
				Slug:    "test-workspace",
				OwnerID: testUserID,
			}
			c.Set("workspace", ws)
			member := &models.WorkspaceMember{
				ID:          uuid.New(),
				WorkspaceID: testWorkspaceID,
				UserID:      testUserID,
				Role:        models.RoleOwner,
			}
			c.Set("workspace_member", member)
		}
		c.Next()
	}

	handler.RegisterRoutes(r.Group("/api/v1"), authMw, wsAccessMw)

	return r
}

func wsURL(path string) string {
	return "/api/v1/workspaces" + path
}

// --- Tests ---

func TestCreateWorkspace_Success(t *testing.T) {
	svc := &mockWorkspaceService{
		createWorkspaceFn: func(_ context.Context, userID uuid.UUID, input models.CreateWorkspaceInput) (*models.Workspace, error) {
			return &models.Workspace{
				ID:      uuid.New(),
				Name:    input.Name,
				Slug:    input.Slug,
				OwnerID: userID,
				Plan:    "free",
			}, nil
		},
	}

	r := setupWorkspaceTestRouter(svc, true, false)

	body, _ := json.Marshal(map[string]string{
		"name": "My Workspace",
		"slug": "myworkspace",
	})

	req := httptest.NewRequest("POST", wsURL(""), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d (body: %s)", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestCreateWorkspace_Unauthenticated(t *testing.T) {
	svc := &mockWorkspaceService{}
	r := setupWorkspaceTestRouter(svc, false, false)

	body, _ := json.Marshal(map[string]string{
		"name": "My Workspace",
		"slug": "myworkspace",
	})

	req := httptest.NewRequest("POST", wsURL(""), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestCreateWorkspace_InvalidBody(t *testing.T) {
	svc := &mockWorkspaceService{}
	r := setupWorkspaceTestRouter(svc, true, false)

	req := httptest.NewRequest("POST", wsURL(""), bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateWorkspace_LimitReached(t *testing.T) {
	svc := &mockWorkspaceService{
		createWorkspaceFn: func(_ context.Context, _ uuid.UUID, _ models.CreateWorkspaceInput) (*models.Workspace, error) {
			return nil, httputil.PaymentRequired("workspace limit reached")
		},
	}

	r := setupWorkspaceTestRouter(svc, true, false)

	body, _ := json.Marshal(map[string]string{
		"name": "Another Workspace",
		"slug": "anotherworkspace",
	})

	req := httptest.NewRequest("POST", wsURL(""), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusPaymentRequired {
		t.Errorf("expected status %d, got %d", http.StatusPaymentRequired, w.Code)
	}
}

func TestListWorkspaces_Success(t *testing.T) {
	svc := &mockWorkspaceService{
		listWorkspacesFn: func(_ context.Context, _ uuid.UUID) ([]*models.Workspace, error) {
			return []*models.Workspace{
				{ID: testWorkspaceID, Name: "Workspace 1", Slug: "ws1", OwnerID: testUserID, Plan: "free"},
			}, nil
		},
		getMemberCountFn: func(_ context.Context, _ uuid.UUID) (int64, error) {
			return 3, nil
		},
		getMemberFn: func(_ context.Context, _, _ uuid.UUID) (*models.WorkspaceMember, error) {
			return &models.WorkspaceMember{Role: models.RoleOwner}, nil
		},
	}

	r := setupWorkspaceTestRouter(svc, true, false)

	req := httptest.NewRequest("GET", wsURL(""), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d (body: %s)", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestGetWorkspace_Success(t *testing.T) {
	svc := &mockWorkspaceService{
		getMemberCountFn: func(_ context.Context, _ uuid.UUID) (int64, error) {
			return 5, nil
		},
	}

	r := setupWorkspaceTestRouter(svc, true, true)

	req := httptest.NewRequest("GET", wsURL("/"+testWorkspaceID.String()), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d (body: %s)", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestGetWorkspace_NoAccess(t *testing.T) {
	svc := &mockWorkspaceService{}
	r := setupWorkspaceTestRouter(svc, true, false) // no workspace access

	req := httptest.NewRequest("GET", wsURL("/"+testWorkspaceID.String()), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestDeleteWorkspace_Success(t *testing.T) {
	svc := &mockWorkspaceService{
		deleteWorkspaceFn: func(_ context.Context, _, _ uuid.UUID) error {
			return nil
		},
	}

	r := setupWorkspaceTestRouter(svc, true, true)

	req := httptest.NewRequest("DELETE", wsURL("/"+testWorkspaceID.String()), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d (body: %s)", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestDeleteWorkspace_NotOwner(t *testing.T) {
	svc := &mockWorkspaceService{
		deleteWorkspaceFn: func(_ context.Context, _, _ uuid.UUID) error {
			return httputil.Forbidden("only the workspace owner can delete the workspace")
		},
	}

	r := setupWorkspaceTestRouter(svc, true, true)

	req := httptest.NewRequest("DELETE", wsURL("/"+testWorkspaceID.String()), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestListMembers_Success(t *testing.T) {
	svc := &mockWorkspaceService{
		listMembersFn: func(_ context.Context, _ uuid.UUID) ([]*models.WorkspaceMemberResponse, error) {
			return []*models.WorkspaceMemberResponse{
				{ID: uuid.New(), UserID: testUserID, Role: models.RoleOwner, Email: "test@example.com", Name: "Test User"},
			}, nil
		},
	}

	r := setupWorkspaceTestRouter(svc, true, true)

	req := httptest.NewRequest("GET", wsURL("/"+testWorkspaceID.String()+"/members"), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d (body: %s)", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestInviteMember_Success(t *testing.T) {
	svc := &mockWorkspaceService{
		inviteMemberFn: func(_ context.Context, wsID, inviterID uuid.UUID, input models.InviteMemberInput) (*models.WorkspaceMember, error) {
			return &models.WorkspaceMember{
				ID:          uuid.New(),
				WorkspaceID: wsID,
				UserID:      uuid.New(),
				Role:        input.Role,
			}, nil
		},
	}

	r := setupWorkspaceTestRouter(svc, true, true)

	body, _ := json.Marshal(map[string]string{
		"email": "newmember@example.com",
		"role":  "editor",
	})

	req := httptest.NewRequest("POST", wsURL("/"+testWorkspaceID.String()+"/members"), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d (body: %s)", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestInviteMember_InvalidBody(t *testing.T) {
	svc := &mockWorkspaceService{}
	r := setupWorkspaceTestRouter(svc, true, true)

	req := httptest.NewRequest("POST", wsURL("/"+testWorkspaceID.String()+"/members"), bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestRemoveMember_Success(t *testing.T) {
	targetUserID := uuid.New()
	svc := &mockWorkspaceService{
		removeMemberFn: func(_ context.Context, _, _, _ uuid.UUID) error {
			return nil
		},
	}

	r := setupWorkspaceTestRouter(svc, true, true)

	req := httptest.NewRequest("DELETE", wsURL("/"+testWorkspaceID.String()+"/members/"+targetUserID.String()), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d (body: %s)", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestRemoveMember_InvalidUserID(t *testing.T) {
	svc := &mockWorkspaceService{}
	r := setupWorkspaceTestRouter(svc, true, true)

	req := httptest.NewRequest("DELETE", wsURL("/"+testWorkspaceID.String()+"/members/not-a-uuid"), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
