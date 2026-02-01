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

// --- Mock DomainService ---

type mockDomainService struct {
	addDomainFn    func(ctx context.Context, workspaceID uuid.UUID, input models.CreateDomainInput) (*models.Domain, error)
	getDomainFn    func(ctx context.Context, id uuid.UUID) (*models.Domain, error)
	listDomainsFn  func(ctx context.Context, workspaceID uuid.UUID) ([]*models.Domain, error)
	verifyDomainFn func(ctx context.Context, id, workspaceID uuid.UUID) (*models.Domain, error)
	removeDomainFn func(ctx context.Context, id, workspaceID uuid.UUID) error
	getDNSRecordsFn func(ctx context.Context, id uuid.UUID) (*models.VerificationInstructions, error)
}

func (m *mockDomainService) AddDomain(ctx context.Context, workspaceID uuid.UUID, input models.CreateDomainInput) (*models.Domain, error) {
	if m.addDomainFn != nil {
		return m.addDomainFn(ctx, workspaceID, input)
	}
	return nil, nil
}

func (m *mockDomainService) GetDomain(ctx context.Context, id uuid.UUID) (*models.Domain, error) {
	if m.getDomainFn != nil {
		return m.getDomainFn(ctx, id)
	}
	return nil, nil
}

func (m *mockDomainService) ListDomains(ctx context.Context, workspaceID uuid.UUID) ([]*models.Domain, error) {
	if m.listDomainsFn != nil {
		return m.listDomainsFn(ctx, workspaceID)
	}
	return nil, nil
}

func (m *mockDomainService) VerifyDomain(ctx context.Context, id, workspaceID uuid.UUID) (*models.Domain, error) {
	if m.verifyDomainFn != nil {
		return m.verifyDomainFn(ctx, id, workspaceID)
	}
	return nil, nil
}

func (m *mockDomainService) RemoveDomain(ctx context.Context, id, workspaceID uuid.UUID) error {
	if m.removeDomainFn != nil {
		return m.removeDomainFn(ctx, id, workspaceID)
	}
	return nil
}

func (m *mockDomainService) GetDNSRecords(ctx context.Context, id uuid.UUID) (*models.VerificationInstructions, error) {
	if m.getDNSRecordsFn != nil {
		return m.getDNSRecordsFn(ctx, id)
	}
	return nil, nil
}

func (m *mockDomainService) SetAuditLogger(_ service.AuditLogger) {}

// --- Helpers ---

func setupDomainTestRouter(svc *mockDomainService, withAuth bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	logger, _ := zap.NewDevelopment()
	handler := NewDomainHandler(svc, logger)

	authAndWsMw := func(c *gin.Context) {
		if withAuth {
			user := &models.User{
				ID:    uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				Email: "test@example.com",
				Name:  "Test User",
			}
			c.Set("user", user)
			ws := &models.Workspace{
				ID:      testWorkspaceID,
				Name:    "Test Workspace",
				Slug:    "test-workspace",
				OwnerID: user.ID,
			}
			c.Set("workspace", ws)
			member := &models.WorkspaceMember{
				ID:          uuid.New(),
				WorkspaceID: testWorkspaceID,
				UserID:      user.ID,
				Role:        models.RoleOwner,
			}
			c.Set("workspace_member", member)
		}
		c.Next()
	}

	editorMw := func(c *gin.Context) { c.Next() }

	wsScoped := r.Group("/api/v1/workspaces/:workspaceId", authAndWsMw)
	handler.RegisterRoutes(wsScoped, editorMw)

	return r
}

func domainURL(path string) string {
	return "/api/v1/workspaces/" + testWorkspaceID.String() + "/domains" + path
}

// --- Tests ---

func TestAddDomain_Success(t *testing.T) {
	domainID := uuid.New()
	svc := &mockDomainService{
		addDomainFn: func(_ context.Context, workspaceID uuid.UUID, input models.CreateDomainInput) (*models.Domain, error) {
			return &models.Domain{
				ID:          domainID,
				WorkspaceID: workspaceID,
				Domain:      input.Domain,
				SSLStatus:   "pending",
			}, nil
		},
	}

	r := setupDomainTestRouter(svc, true)

	body := `{"domain":"example.com"}`
	req := httptest.NewRequest("POST", domainURL(""), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d (body: %s)", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestAddDomain_InvalidBody(t *testing.T) {
	svc := &mockDomainService{}
	r := setupDomainTestRouter(svc, true)

	req := httptest.NewRequest("POST", domainURL(""), bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAddDomain_NoAuth(t *testing.T) {
	svc := &mockDomainService{}
	r := setupDomainTestRouter(svc, false)

	body := `{"domain":"example.com"}`
	req := httptest.NewRequest("POST", domainURL(""), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestListDomains_Success(t *testing.T) {
	svc := &mockDomainService{
		listDomainsFn: func(_ context.Context, _ uuid.UUID) ([]*models.Domain, error) {
			return []*models.Domain{
				{ID: uuid.New(), Domain: "example.com"},
			}, nil
		},
	}

	r := setupDomainTestRouter(svc, true)

	req := httptest.NewRequest("GET", domainURL(""), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d (body: %s)", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestGetDomain_InvalidUUID(t *testing.T) {
	svc := &mockDomainService{}
	r := setupDomainTestRouter(svc, true)

	req := httptest.NewRequest("GET", domainURL("/not-a-uuid"), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetDomain_NotFound(t *testing.T) {
	svc := &mockDomainService{
		getDomainFn: func(_ context.Context, _ uuid.UUID) (*models.Domain, error) {
			return nil, httputil.NotFound("domain")
		},
	}

	r := setupDomainTestRouter(svc, true)

	req := httptest.NewRequest("GET", domainURL("/"+uuid.New().String()), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestVerifyDomain_Success(t *testing.T) {
	domainID := uuid.New()
	svc := &mockDomainService{
		verifyDomainFn: func(_ context.Context, id, _ uuid.UUID) (*models.Domain, error) {
			return &models.Domain{
				ID:         id,
				Domain:     "example.com",
				IsVerified: true,
				SSLStatus:  "active",
			}, nil
		},
	}

	r := setupDomainTestRouter(svc, true)

	req := httptest.NewRequest("POST", domainURL("/"+domainID.String()+"/verify"), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d (body: %s)", http.StatusOK, w.Code, w.Body.String())
	}

	var resp httputil.Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestRemoveDomain_Success(t *testing.T) {
	domainID := uuid.New()
	svc := &mockDomainService{
		removeDomainFn: func(_ context.Context, _, _ uuid.UUID) error {
			return nil
		},
	}

	r := setupDomainTestRouter(svc, true)

	req := httptest.NewRequest("DELETE", domainURL("/"+domainID.String()), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d (body: %s)", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestRemoveDomain_InvalidUUID(t *testing.T) {
	svc := &mockDomainService{}
	r := setupDomainTestRouter(svc, true)

	req := httptest.NewRequest("DELETE", domainURL("/not-a-uuid"), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetDNSRecords_Success(t *testing.T) {
	domainID := uuid.New()
	svc := &mockDomainService{
		getDNSRecordsFn: func(_ context.Context, _ uuid.UUID) (*models.VerificationInstructions, error) {
			return &models.VerificationInstructions{
				Records: []models.DNSRecordInstruction{
					{Type: "TXT", Host: "_linkrift.example.com", Value: "linkrift-verification=test"},
					{Type: "CNAME", Host: "example.com", Value: "proxy.example.com"},
				},
			}, nil
		},
	}

	r := setupDomainTestRouter(svc, true)

	req := httptest.NewRequest("GET", domainURL("/"+domainID.String()+"/dns-records"), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d (body: %s)", http.StatusOK, w.Code, w.Body.String())
	}
}
