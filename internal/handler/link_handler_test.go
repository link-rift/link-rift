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

// --- Mock LinkService ---

type mockLinkService struct {
	createLinkFn           func(ctx context.Context, userID, workspaceID uuid.UUID, input models.CreateLinkInput) (*models.Link, error)
	updateLinkFn           func(ctx context.Context, id, userID uuid.UUID, input models.UpdateLinkInput) (*models.Link, error)
	deleteLinkFn           func(ctx context.Context, id, userID uuid.UUID) error
	getLinkFn              func(ctx context.Context, id uuid.UUID) (*models.Link, error)
	listLinksFn            func(ctx context.Context, workspaceID uuid.UUID, filter models.LinkFilter, pagination models.Pagination) (*models.LinkListResult, error)
	bulkCreateLinksFn      func(ctx context.Context, userID, workspaceID uuid.UUID, input models.BulkCreateLinkInput) ([]*models.Link, error)
	getQuickStatsFn        func(ctx context.Context, id uuid.UUID) (*models.LinkQuickStats, error)
	checkShortCodeFn       func(ctx context.Context, code string) (bool, error)
	verifyLinkPasswordFn   func(ctx context.Context, shortCode, password string) (bool, error)
}

func (m *mockLinkService) CreateLink(ctx context.Context, userID, workspaceID uuid.UUID, input models.CreateLinkInput) (*models.Link, error) {
	if m.createLinkFn != nil {
		return m.createLinkFn(ctx, userID, workspaceID, input)
	}
	return nil, nil
}

func (m *mockLinkService) UpdateLink(ctx context.Context, id, userID uuid.UUID, input models.UpdateLinkInput) (*models.Link, error) {
	if m.updateLinkFn != nil {
		return m.updateLinkFn(ctx, id, userID, input)
	}
	return nil, nil
}

func (m *mockLinkService) DeleteLink(ctx context.Context, id, userID uuid.UUID) error {
	if m.deleteLinkFn != nil {
		return m.deleteLinkFn(ctx, id, userID)
	}
	return nil
}

func (m *mockLinkService) GetLink(ctx context.Context, id uuid.UUID) (*models.Link, error) {
	if m.getLinkFn != nil {
		return m.getLinkFn(ctx, id)
	}
	return nil, nil
}

func (m *mockLinkService) ListLinks(ctx context.Context, workspaceID uuid.UUID, filter models.LinkFilter, pagination models.Pagination) (*models.LinkListResult, error) {
	if m.listLinksFn != nil {
		return m.listLinksFn(ctx, workspaceID, filter, pagination)
	}
	return nil, nil
}

func (m *mockLinkService) BulkCreateLinks(ctx context.Context, userID, workspaceID uuid.UUID, input models.BulkCreateLinkInput) ([]*models.Link, error) {
	if m.bulkCreateLinksFn != nil {
		return m.bulkCreateLinksFn(ctx, userID, workspaceID, input)
	}
	return nil, nil
}

func (m *mockLinkService) GetQuickStats(ctx context.Context, id uuid.UUID) (*models.LinkQuickStats, error) {
	if m.getQuickStatsFn != nil {
		return m.getQuickStatsFn(ctx, id)
	}
	return nil, nil
}

func (m *mockLinkService) CheckShortCodeAvailable(ctx context.Context, code string) (bool, error) {
	if m.checkShortCodeFn != nil {
		return m.checkShortCodeFn(ctx, code)
	}
	return false, nil
}

func (m *mockLinkService) VerifyLinkPassword(ctx context.Context, shortCode, password string) (bool, error) {
	if m.verifyLinkPasswordFn != nil {
		return m.verifyLinkPasswordFn(ctx, shortCode, password)
	}
	return false, nil
}

// --- Test Router Setup ---

func setupTestRouter(svc *mockLinkService, withAuth bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	logger, _ := zap.NewDevelopment()
	handler := NewLinkHandler(svc, logger)

	authMw := func(c *gin.Context) {
		if withAuth {
			// Inject a test user
			user := &models.User{
				ID:    uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				Email: "test@example.com",
				Name:  "Test User",
			}
			c.Set("user", user)
		}
		c.Next()
	}

	api := r.Group("/api/v1")
	handler.RegisterRoutes(api, authMw)

	return r
}

func parseResponse(t *testing.T, w *httptest.ResponseRecorder) httputil.Response {
	t.Helper()
	var resp httputil.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v (body: %s)", err, w.Body.String())
	}
	return resp
}

// --- Tests ---

func TestCreateLink_Success(t *testing.T) {
	svc := &mockLinkService{
		createLinkFn: func(_ context.Context, userID, workspaceID uuid.UUID, input models.CreateLinkInput) (*models.Link, error) {
			return &models.Link{
				ID:          uuid.New(),
				UserID:      userID,
				WorkspaceID: workspaceID,
				URL:         input.URL,
				ShortCode:   "abc123",
				IsActive:    true,
			}, nil
		},
	}

	r := setupTestRouter(svc, true)

	body := `{"url":"https://example.com","title":"Test"}`
	wsID := uuid.New().String()
	req := httptest.NewRequest("POST", "/api/v1/links?workspace_id="+wsID, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d (body: %s)", http.StatusCreated, w.Code, w.Body.String())
	}

	resp := parseResponse(t, w)
	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestCreateLink_Unauthenticated(t *testing.T) {
	svc := &mockLinkService{}

	r := setupTestRouter(svc, false) // no auth

	body := `{"url":"https://example.com"}`
	wsID := uuid.New().String()
	req := httptest.NewRequest("POST", "/api/v1/links?workspace_id="+wsID, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestCreateLink_MissingWorkspaceID(t *testing.T) {
	svc := &mockLinkService{}

	r := setupTestRouter(svc, true)

	body := `{"url":"https://example.com"}`
	req := httptest.NewRequest("POST", "/api/v1/links", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d (body: %s)", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func TestCreateLink_InvalidBody(t *testing.T) {
	svc := &mockLinkService{}

	r := setupTestRouter(svc, true)

	wsID := uuid.New().String()
	req := httptest.NewRequest("POST", "/api/v1/links?workspace_id="+wsID, bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestListLinks_Success(t *testing.T) {
	wsID := uuid.New()

	svc := &mockLinkService{
		listLinksFn: func(_ context.Context, workspaceID uuid.UUID, _ models.LinkFilter, _ models.Pagination) (*models.LinkListResult, error) {
			if workspaceID != wsID {
				t.Errorf("expected workspace_id %s, got %s", wsID, workspaceID)
			}
			return &models.LinkListResult{
				Links: []*models.LinkResponse{},
				Total: 0,
			}, nil
		},
	}

	r := setupTestRouter(svc, true)

	req := httptest.NewRequest("GET", "/api/v1/links?workspace_id="+wsID.String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d (body: %s)", http.StatusOK, w.Code, w.Body.String())
	}

	resp := parseResponse(t, w)
	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.Meta == nil {
		t.Error("expected meta in response")
	}
}

func TestGetLink_Success(t *testing.T) {
	linkID := uuid.New()

	svc := &mockLinkService{
		getLinkFn: func(_ context.Context, id uuid.UUID) (*models.Link, error) {
			if id != linkID {
				t.Errorf("expected ID %s, got %s", linkID, id)
			}
			return &models.Link{
				ID:        linkID,
				URL:       "https://example.com",
				ShortCode: "abc123",
			}, nil
		},
	}

	r := setupTestRouter(svc, true)

	req := httptest.NewRequest("GET", "/api/v1/links/"+linkID.String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d (body: %s)", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestGetLink_InvalidUUID(t *testing.T) {
	svc := &mockLinkService{}

	r := setupTestRouter(svc, true)

	req := httptest.NewRequest("GET", "/api/v1/links/not-a-uuid", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateLink_Success(t *testing.T) {
	linkID := uuid.New()

	svc := &mockLinkService{
		updateLinkFn: func(_ context.Context, id, userID uuid.UUID, input models.UpdateLinkInput) (*models.Link, error) {
			return &models.Link{
				ID:        id,
				URL:       "https://updated.com",
				ShortCode: "abc123",
			}, nil
		},
	}

	r := setupTestRouter(svc, true)

	body := `{"title":"Updated"}`
	req := httptest.NewRequest("PUT", "/api/v1/links/"+linkID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d (body: %s)", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestUpdateLink_Unauthenticated(t *testing.T) {
	svc := &mockLinkService{}

	r := setupTestRouter(svc, false)

	body := `{"title":"Updated"}`
	req := httptest.NewRequest("PUT", "/api/v1/links/"+uuid.New().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestDeleteLink_Success(t *testing.T) {
	linkID := uuid.New()

	svc := &mockLinkService{
		deleteLinkFn: func(_ context.Context, id, userID uuid.UUID) error {
			if id != linkID {
				t.Errorf("expected ID %s, got %s", linkID, id)
			}
			return nil
		},
	}

	r := setupTestRouter(svc, true)

	req := httptest.NewRequest("DELETE", "/api/v1/links/"+linkID.String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d (body: %s)", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestDeleteLink_Unauthenticated(t *testing.T) {
	svc := &mockLinkService{}

	r := setupTestRouter(svc, false)

	req := httptest.NewRequest("DELETE", "/api/v1/links/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestBulkCreateLinks_Success(t *testing.T) {
	svc := &mockLinkService{
		bulkCreateLinksFn: func(_ context.Context, userID, workspaceID uuid.UUID, input models.BulkCreateLinkInput) ([]*models.Link, error) {
			links := make([]*models.Link, len(input.Links))
			for i := range input.Links {
				links[i] = &models.Link{
					ID:        uuid.New(),
					URL:       input.Links[i].URL,
					ShortCode: "bulk" + string(rune('0'+i)),
				}
			}
			return links, nil
		},
	}

	r := setupTestRouter(svc, true)

	body := `{"links":[{"url":"https://example.com"},{"url":"https://example.org"}]}`
	wsID := uuid.New().String()
	req := httptest.NewRequest("POST", "/api/v1/links/bulk?workspace_id="+wsID, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d (body: %s)", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestBulkCreateLinks_Unauthenticated(t *testing.T) {
	svc := &mockLinkService{}

	r := setupTestRouter(svc, false)

	body := `{"links":[{"url":"https://example.com"}]}`
	wsID := uuid.New().String()
	req := httptest.NewRequest("POST", "/api/v1/links/bulk?workspace_id="+wsID, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetQuickStats_Success(t *testing.T) {
	linkID := uuid.New()

	svc := &mockLinkService{
		getQuickStatsFn: func(_ context.Context, id uuid.UUID) (*models.LinkQuickStats, error) {
			return &models.LinkQuickStats{
				TotalClicks:  100,
				UniqueClicks: 80,
				Clicks24h:    10,
				Clicks7d:     50,
			}, nil
		},
	}

	r := setupTestRouter(svc, true)

	req := httptest.NewRequest("GET", "/api/v1/links/"+linkID.String()+"/stats", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d (body: %s)", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestGetQuickStats_InvalidID(t *testing.T) {
	svc := &mockLinkService{}

	r := setupTestRouter(svc, true)

	req := httptest.NewRequest("GET", "/api/v1/links/not-a-uuid/stats", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetQuickStats_NotFound(t *testing.T) {
	svc := &mockLinkService{
		getQuickStatsFn: func(_ context.Context, _ uuid.UUID) (*models.LinkQuickStats, error) {
			return nil, httputil.NotFound("link")
		},
	}

	r := setupTestRouter(svc, true)

	req := httptest.NewRequest("GET", "/api/v1/links/"+uuid.New().String()+"/stats", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d (body: %s)", http.StatusNotFound, w.Code, w.Body.String())
	}
}
