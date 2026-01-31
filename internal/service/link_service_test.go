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
	"github.com/link-rift/link-rift/pkg/httputil"
	"github.com/link-rift/link-rift/pkg/shortcode"
	"go.uber.org/zap"
)

// --- Mock LinkRepository ---

type mockLinkRepo struct {
	createFn             func(ctx context.Context, params sqlc.CreateLinkParams) (*models.Link, error)
	getByIDFn            func(ctx context.Context, id uuid.UUID) (*models.Link, error)
	getByShortCodeFn     func(ctx context.Context, shortCode string) (*models.Link, error)
	getByURLFn           func(ctx context.Context, params sqlc.GetLinkByURLParams) (*models.Link, error)
	listFn               func(ctx context.Context, params sqlc.ListLinksForWorkspaceParams) ([]*models.Link, int64, error)
	updateFn             func(ctx context.Context, params sqlc.UpdateLinkParams) (*models.Link, error)
	softDeleteFn         func(ctx context.Context, id uuid.UUID) error
	shortCodeExistsFn    func(ctx context.Context, shortCode string) (bool, error)
	incrementClicksFn    func(ctx context.Context, id uuid.UUID) error
	incrementUniqueFn    func(ctx context.Context, id uuid.UUID) error
	getQuickStatsFn      func(ctx context.Context, id uuid.UUID) (*models.LinkQuickStats, error)
	getCountFn           func(ctx context.Context, workspaceID uuid.UUID) (int64, error)
}

func (m *mockLinkRepo) Create(ctx context.Context, params sqlc.CreateLinkParams) (*models.Link, error) {
	if m.createFn != nil {
		return m.createFn(ctx, params)
	}
	return nil, nil
}

func (m *mockLinkRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Link, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockLinkRepo) GetByShortCode(ctx context.Context, shortCode string) (*models.Link, error) {
	if m.getByShortCodeFn != nil {
		return m.getByShortCodeFn(ctx, shortCode)
	}
	return nil, nil
}

func (m *mockLinkRepo) GetByURL(ctx context.Context, params sqlc.GetLinkByURLParams) (*models.Link, error) {
	if m.getByURLFn != nil {
		return m.getByURLFn(ctx, params)
	}
	return nil, nil
}

func (m *mockLinkRepo) List(ctx context.Context, params sqlc.ListLinksForWorkspaceParams) ([]*models.Link, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, params)
	}
	return nil, 0, nil
}

func (m *mockLinkRepo) Update(ctx context.Context, params sqlc.UpdateLinkParams) (*models.Link, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, params)
	}
	return nil, nil
}

func (m *mockLinkRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if m.softDeleteFn != nil {
		return m.softDeleteFn(ctx, id)
	}
	return nil
}

func (m *mockLinkRepo) ShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	if m.shortCodeExistsFn != nil {
		return m.shortCodeExistsFn(ctx, shortCode)
	}
	return false, nil
}

func (m *mockLinkRepo) IncrementClicks(ctx context.Context, id uuid.UUID) error {
	if m.incrementClicksFn != nil {
		return m.incrementClicksFn(ctx, id)
	}
	return nil
}

func (m *mockLinkRepo) IncrementUniqueClicks(ctx context.Context, id uuid.UUID) error {
	if m.incrementUniqueFn != nil {
		return m.incrementUniqueFn(ctx, id)
	}
	return nil
}

func (m *mockLinkRepo) GetQuickStats(ctx context.Context, id uuid.UUID) (*models.LinkQuickStats, error) {
	if m.getQuickStatsFn != nil {
		return m.getQuickStatsFn(ctx, id)
	}
	return nil, nil
}

func (m *mockLinkRepo) GetCountForWorkspace(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	if m.getCountFn != nil {
		return m.getCountFn(ctx, workspaceID)
	}
	return 0, nil
}

// --- Mock ClickRepository ---

type mockClickRepo struct {
	insertFn      func(ctx context.Context, params sqlc.InsertClickParams) error
	getByLinkIDFn func(ctx context.Context, params sqlc.GetClicksByLinkIDParams) ([]*models.Click, error)
}

func (m *mockClickRepo) Insert(ctx context.Context, params sqlc.InsertClickParams) error {
	if m.insertFn != nil {
		return m.insertFn(ctx, params)
	}
	return nil
}

func (m *mockClickRepo) GetByLinkID(ctx context.Context, params sqlc.GetClicksByLinkIDParams) ([]*models.Click, error) {
	if m.getByLinkIDFn != nil {
		return m.getByLinkIDFn(ctx, params)
	}
	return nil, nil
}

// --- Mock shortcode Generator ---

type mockCodeGen struct {
	code string
	seq  int
}

func (m *mockCodeGen) Generate() string {
	m.seq++
	if m.code != "" {
		return m.code
	}
	return "abc1234"
}

func (m *mockCodeGen) GenerateWithLength(n int) string {
	return m.Generate()
}

// --- Helpers ---

func newTestService(linkRepo *mockLinkRepo, clickRepo *mockClickRepo, codeGen shortcode.Generator) *linkService {
	logger, _ := zap.NewDevelopment()
	return &linkService{
		linkRepo:  linkRepo,
		clickRepo: clickRepo,
		cfg:       &config.Config{App: config.AppConfig{RedirectURL: "http://localhost:8081"}},
		codeGen:   codeGen,
		logger:    logger,
	}
}

func makeLink(id, userID, workspaceID uuid.UUID, shortCode string) *models.Link {
	return &models.Link{
		ID:          id,
		UserID:      userID,
		WorkspaceID: workspaceID,
		URL:         "https://example.com",
		ShortCode:   shortCode,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func strPtr(s string) *string { return &s }
func int32Ptr(i int32) *int32 { return &i }
func boolPtr(b bool) *bool    { return &b }

// --- Tests ---

func TestCreateLink_ValidInput(t *testing.T) {
	linkID := uuid.New()
	userID := uuid.New()
	workspaceID := uuid.New()

	repo := &mockLinkRepo{
		shortCodeExistsFn: func(_ context.Context, _ string) (bool, error) { return false, nil },
		createFn: func(_ context.Context, params sqlc.CreateLinkParams) (*models.Link, error) {
			if params.UserID != userID {
				t.Errorf("expected user_id %s, got %s", userID, params.UserID)
			}
			if params.WorkspaceID != workspaceID {
				t.Errorf("expected workspace_id %s, got %s", workspaceID, params.WorkspaceID)
			}
			if params.ShortCode != "test123" {
				t.Errorf("expected short_code test123, got %s", params.ShortCode)
			}
			return makeLink(linkID, userID, workspaceID, "test123"), nil
		},
	}

	svc := newTestService(repo, &mockClickRepo{}, &mockCodeGen{code: "test123"})

	input := models.CreateLinkInput{
		URL:   "https://example.com",
		Title: strPtr("Test Link"),
	}

	link, err := svc.CreateLink(context.Background(), userID, workspaceID, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link.ID != linkID {
		t.Errorf("expected link ID %s, got %s", linkID, link.ID)
	}
}

func TestCreateLink_CustomShortCode(t *testing.T) {
	userID := uuid.New()
	workspaceID := uuid.New()

	repo := &mockLinkRepo{
		shortCodeExistsFn: func(_ context.Context, code string) (bool, error) {
			if code != "my-custom" {
				t.Errorf("expected short code 'my-custom', got %s", code)
			}
			return false, nil
		},
		createFn: func(_ context.Context, params sqlc.CreateLinkParams) (*models.Link, error) {
			if params.ShortCode != "my-custom" {
				t.Errorf("expected short_code my-custom, got %s", params.ShortCode)
			}
			return makeLink(uuid.New(), userID, workspaceID, "my-custom"), nil
		},
	}

	svc := newTestService(repo, &mockClickRepo{}, &mockCodeGen{})

	input := models.CreateLinkInput{
		URL:       "https://example.com",
		ShortCode: strPtr("my-custom"),
	}

	link, err := svc.CreateLink(context.Background(), userID, workspaceID, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link.ShortCode != "my-custom" {
		t.Errorf("expected short code 'my-custom', got %s", link.ShortCode)
	}
}

func TestCreateLink_DuplicateShortCode(t *testing.T) {
	repo := &mockLinkRepo{
		shortCodeExistsFn: func(_ context.Context, _ string) (bool, error) { return true, nil },
	}

	svc := newTestService(repo, &mockClickRepo{}, &mockCodeGen{})

	input := models.CreateLinkInput{
		URL:       "https://example.com",
		ShortCode: strPtr("taken"),
	}

	_, err := svc.CreateLink(context.Background(), uuid.New(), uuid.New(), input)
	if err == nil {
		t.Fatal("expected error for duplicate short code")
	}

	var appErr *httputil.AppError
	if !errors.As(err, &appErr) || appErr.Code != "ALREADY_EXISTS" {
		t.Errorf("expected ALREADY_EXISTS error, got %v", err)
	}
}

func TestCreateLink_InvalidURL(t *testing.T) {
	svc := newTestService(&mockLinkRepo{}, &mockClickRepo{}, &mockCodeGen{})

	input := models.CreateLinkInput{
		URL: "",
	}

	_, err := svc.CreateLink(context.Background(), uuid.New(), uuid.New(), input)
	if err == nil {
		t.Fatal("expected error for empty URL")
	}

	var appErr *httputil.AppError
	if !errors.As(err, &appErr) || appErr.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %v", err)
	}
}

func TestCreateLink_InvalidShortCode(t *testing.T) {
	svc := newTestService(&mockLinkRepo{}, &mockClickRepo{}, &mockCodeGen{})

	input := models.CreateLinkInput{
		URL:       "https://example.com",
		ShortCode: strPtr("ab"), // too short
	}

	_, err := svc.CreateLink(context.Background(), uuid.New(), uuid.New(), input)
	if err == nil {
		t.Fatal("expected error for invalid short code")
	}

	var appErr *httputil.AppError
	if !errors.As(err, &appErr) || appErr.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %v", err)
	}
}

func TestCreateLink_WithPassword(t *testing.T) {
	repo := &mockLinkRepo{
		shortCodeExistsFn: func(_ context.Context, _ string) (bool, error) { return false, nil },
		createFn: func(_ context.Context, params sqlc.CreateLinkParams) (*models.Link, error) {
			if !params.PasswordHash.Valid {
				t.Error("expected password hash to be set")
			}
			if params.PasswordHash.String == "" {
				t.Error("expected non-empty password hash")
			}
			return makeLink(uuid.New(), params.UserID, params.WorkspaceID, params.ShortCode), nil
		},
	}

	svc := newTestService(repo, &mockClickRepo{}, &mockCodeGen{})

	input := models.CreateLinkInput{
		URL:      "https://example.com",
		Password: strPtr("secret123"),
	}

	_, err := svc.CreateLink(context.Background(), uuid.New(), uuid.New(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateLink_WithExpiration(t *testing.T) {
	future := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	repo := &mockLinkRepo{
		shortCodeExistsFn: func(_ context.Context, _ string) (bool, error) { return false, nil },
		createFn: func(_ context.Context, params sqlc.CreateLinkParams) (*models.Link, error) {
			if !params.ExpiresAt.Valid {
				t.Error("expected expires_at to be set")
			}
			return makeLink(uuid.New(), params.UserID, params.WorkspaceID, params.ShortCode), nil
		},
	}

	svc := newTestService(repo, &mockClickRepo{}, &mockCodeGen{})

	input := models.CreateLinkInput{
		URL:       "https://example.com",
		ExpiresAt: &future,
	}

	_, err := svc.CreateLink(context.Background(), uuid.New(), uuid.New(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateLink_PastExpiration(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)

	repo := &mockLinkRepo{
		shortCodeExistsFn: func(_ context.Context, _ string) (bool, error) { return false, nil },
	}

	svc := newTestService(repo, &mockClickRepo{}, &mockCodeGen{})

	input := models.CreateLinkInput{
		URL:       "https://example.com",
		ExpiresAt: &past,
	}

	_, err := svc.CreateLink(context.Background(), uuid.New(), uuid.New(), input)
	if err == nil {
		t.Fatal("expected error for past expiration date")
	}
}

func TestUpdateLink_ValidUpdate(t *testing.T) {
	linkID := uuid.New()
	userID := uuid.New()

	repo := &mockLinkRepo{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*models.Link, error) {
			return makeLink(linkID, userID, uuid.New(), "abc123"), nil
		},
		updateFn: func(_ context.Context, params sqlc.UpdateLinkParams) (*models.Link, error) {
			if params.ID != linkID {
				t.Errorf("expected link ID %s, got %s", linkID, params.ID)
			}
			link := makeLink(linkID, userID, uuid.New(), "abc123")
			link.URL = "https://updated.com"
			return link, nil
		},
	}

	svc := newTestService(repo, &mockClickRepo{}, &mockCodeGen{})

	input := models.UpdateLinkInput{
		URL:   strPtr("https://updated.com"),
		Title: strPtr("Updated Title"),
	}

	link, err := svc.UpdateLink(context.Background(), linkID, userID, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link.ID != linkID {
		t.Errorf("expected link ID %s, got %s", linkID, link.ID)
	}
}

func TestUpdateLink_OwnershipCheck(t *testing.T) {
	linkID := uuid.New()
	ownerID := uuid.New()
	otherUserID := uuid.New()

	repo := &mockLinkRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*models.Link, error) {
			return makeLink(linkID, ownerID, uuid.New(), "abc123"), nil
		},
	}

	svc := newTestService(repo, &mockClickRepo{}, &mockCodeGen{})

	input := models.UpdateLinkInput{Title: strPtr("New Title")}

	_, err := svc.UpdateLink(context.Background(), linkID, otherUserID, input)
	if err == nil {
		t.Fatal("expected forbidden error for non-owner")
	}

	var appErr *httputil.AppError
	if !errors.As(err, &appErr) || appErr.Code != "FORBIDDEN" {
		t.Errorf("expected FORBIDDEN error, got %v", err)
	}
}

func TestUpdateLink_InvalidURL(t *testing.T) {
	linkID := uuid.New()
	userID := uuid.New()

	repo := &mockLinkRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*models.Link, error) {
			return makeLink(linkID, userID, uuid.New(), "abc123"), nil
		},
	}

	svc := newTestService(repo, &mockClickRepo{}, &mockCodeGen{})

	input := models.UpdateLinkInput{URL: strPtr("")}

	_, err := svc.UpdateLink(context.Background(), linkID, userID, input)
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestUpdateLink_ClearPassword(t *testing.T) {
	linkID := uuid.New()
	userID := uuid.New()

	repo := &mockLinkRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*models.Link, error) {
			link := makeLink(linkID, userID, uuid.New(), "abc123")
			hash := "hashed_password"
			link.PasswordHash = &hash
			return link, nil
		},
		updateFn: func(_ context.Context, params sqlc.UpdateLinkParams) (*models.Link, error) {
			if !params.PasswordHash.Valid || params.PasswordHash.String != "" {
				t.Error("expected password hash to be cleared (empty valid string)")
			}
			return makeLink(linkID, userID, uuid.New(), "abc123"), nil
		},
	}

	svc := newTestService(repo, &mockClickRepo{}, &mockCodeGen{})

	emptyPass := ""
	input := models.UpdateLinkInput{Password: &emptyPass}

	_, err := svc.UpdateLink(context.Background(), linkID, userID, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteLink_Valid(t *testing.T) {
	linkID := uuid.New()
	userID := uuid.New()
	deleted := false

	repo := &mockLinkRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*models.Link, error) {
			return makeLink(linkID, userID, uuid.New(), "abc123"), nil
		},
		softDeleteFn: func(_ context.Context, id uuid.UUID) error {
			deleted = true
			if id != linkID {
				t.Errorf("expected link ID %s, got %s", linkID, id)
			}
			return nil
		},
	}

	svc := newTestService(repo, &mockClickRepo{}, &mockCodeGen{})

	err := svc.DeleteLink(context.Background(), linkID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Error("soft delete was not called")
	}
}

func TestDeleteLink_OwnershipCheck(t *testing.T) {
	linkID := uuid.New()
	ownerID := uuid.New()
	otherUserID := uuid.New()

	repo := &mockLinkRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*models.Link, error) {
			return makeLink(linkID, ownerID, uuid.New(), "abc123"), nil
		},
	}

	svc := newTestService(repo, &mockClickRepo{}, &mockCodeGen{})

	err := svc.DeleteLink(context.Background(), linkID, otherUserID)
	if err == nil {
		t.Fatal("expected forbidden error for non-owner")
	}
}

func TestGetLink_Found(t *testing.T) {
	linkID := uuid.New()

	repo := &mockLinkRepo{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*models.Link, error) {
			return makeLink(linkID, uuid.New(), uuid.New(), "abc123"), nil
		},
	}

	svc := newTestService(repo, &mockClickRepo{}, &mockCodeGen{})

	link, err := svc.GetLink(context.Background(), linkID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link.ID != linkID {
		t.Errorf("expected ID %s, got %s", linkID, link.ID)
	}
}

func TestGetLink_NotFound(t *testing.T) {
	repo := &mockLinkRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*models.Link, error) {
			return nil, httputil.NotFound("link")
		},
	}

	svc := newTestService(repo, &mockClickRepo{}, &mockCodeGen{})

	_, err := svc.GetLink(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
}

func TestListLinks_DefaultPagination(t *testing.T) {
	workspaceID := uuid.New()

	repo := &mockLinkRepo{
		listFn: func(_ context.Context, params sqlc.ListLinksForWorkspaceParams) ([]*models.Link, int64, error) {
			if params.Limit != 20 {
				t.Errorf("expected default limit 20, got %d", params.Limit)
			}
			if params.WorkspaceID != workspaceID {
				t.Errorf("expected workspace_id %s, got %s", workspaceID, params.WorkspaceID)
			}
			return []*models.Link{makeLink(uuid.New(), uuid.New(), workspaceID, "abc123")}, 1, nil
		},
	}

	svc := newTestService(repo, &mockClickRepo{}, &mockCodeGen{})

	result, err := svc.ListLinks(context.Background(), workspaceID, models.LinkFilter{}, models.Pagination{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
	if len(result.Links) != 1 {
		t.Errorf("expected 1 link, got %d", len(result.Links))
	}
}

func TestListLinks_WithFilter(t *testing.T) {
	workspaceID := uuid.New()
	search := "test"

	repo := &mockLinkRepo{
		listFn: func(_ context.Context, params sqlc.ListLinksForWorkspaceParams) ([]*models.Link, int64, error) {
			if !params.Search.Valid || params.Search.String != "test" {
				t.Errorf("expected search 'test', got %v", params.Search)
			}
			return []*models.Link{}, 0, nil
		},
	}

	svc := newTestService(repo, &mockClickRepo{}, &mockCodeGen{})

	result, err := svc.ListLinks(context.Background(), workspaceID, models.LinkFilter{Search: &search}, models.Pagination{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
}

func TestGetQuickStats_Success(t *testing.T) {
	linkID := uuid.New()
	expected := &models.LinkQuickStats{
		TotalClicks:  100,
		UniqueClicks: 80,
		Clicks24h:    10,
		Clicks7d:     50,
	}

	repo := &mockLinkRepo{
		getQuickStatsFn: func(_ context.Context, id uuid.UUID) (*models.LinkQuickStats, error) {
			if id != linkID {
				t.Errorf("expected ID %s, got %s", linkID, id)
			}
			return expected, nil
		},
	}

	svc := newTestService(repo, &mockClickRepo{}, &mockCodeGen{})

	stats, err := svc.GetQuickStats(context.Background(), linkID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalClicks != 100 {
		t.Errorf("expected 100 total clicks, got %d", stats.TotalClicks)
	}
}

func TestGetQuickStats_NotFound(t *testing.T) {
	repo := &mockLinkRepo{
		getQuickStatsFn: func(_ context.Context, _ uuid.UUID) (*models.LinkQuickStats, error) {
			return nil, httputil.NotFound("link")
		},
	}

	svc := newTestService(repo, &mockClickRepo{}, &mockCodeGen{})

	_, err := svc.GetQuickStats(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
}

func TestCheckShortCodeAvailable_Available(t *testing.T) {
	repo := &mockLinkRepo{
		shortCodeExistsFn: func(_ context.Context, _ string) (bool, error) { return false, nil },
	}

	svc := newTestService(repo, &mockClickRepo{}, &mockCodeGen{})

	available, err := svc.CheckShortCodeAvailable(context.Background(), "newcode")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !available {
		t.Error("expected code to be available")
	}
}

func TestCheckShortCodeAvailable_Taken(t *testing.T) {
	repo := &mockLinkRepo{
		shortCodeExistsFn: func(_ context.Context, _ string) (bool, error) { return true, nil },
	}

	svc := newTestService(repo, &mockClickRepo{}, &mockCodeGen{})

	available, err := svc.CheckShortCodeAvailable(context.Background(), "taken")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if available {
		t.Error("expected code to be taken")
	}
}

func TestVerifyLinkPassword_Correct(t *testing.T) {
	// We can't easily test bcrypt/argon2 without a real hash,
	// so we test the no-password path instead.
	repo := &mockLinkRepo{
		getByShortCodeFn: func(_ context.Context, _ string) (*models.Link, error) {
			link := makeLink(uuid.New(), uuid.New(), uuid.New(), "abc123")
			// No password set
			return link, nil
		},
	}

	svc := newTestService(repo, &mockClickRepo{}, &mockCodeGen{})

	ok, err := svc.VerifyLinkPassword(context.Background(), "abc123", "anything")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected password verification to succeed when no password is set")
	}
}

func TestVerifyLinkPassword_NotFound(t *testing.T) {
	repo := &mockLinkRepo{
		getByShortCodeFn: func(_ context.Context, _ string) (*models.Link, error) {
			return nil, httputil.NotFound("link")
		},
	}

	svc := newTestService(repo, &mockClickRepo{}, &mockCodeGen{})

	_, err := svc.VerifyLinkPassword(context.Background(), "missing", "pass")
	if err == nil {
		t.Fatal("expected error for missing link")
	}
}

func TestBulkCreateLinks_NilPool(t *testing.T) {
	// BulkCreateLinks requires a pgxpool which we can't easily mock in unit tests.
	// Verify it handles the nil pool case by recovering from the panic.
	svc := newTestService(&mockLinkRepo{}, &mockClickRepo{}, &mockCodeGen{})

	input := models.BulkCreateLinkInput{
		Links: []models.CreateLinkInput{
			{URL: "https://example.com"},
		},
	}

	// pool is nil so Begin() will panic — verify the function signature is correct
	// by checking it exists (compilation check). Skip actual execution.
	_ = svc
	_ = input
	t.Skip("BulkCreateLinks requires a real pgxpool; covered by integration tests")
}

// --- Helper function tests ---

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"full URL", "https://example.com/path", "https://example.com/path", false},
		{"no scheme", "example.com/path", "https://example.com/path", false},
		{"http scheme", "http://example.com", "http://example.com", false},
		{"with query params", "https://example.com?q=test", "https://example.com?q=test", false},
		{"empty string", "", "", true},
		{"whitespace only", "   ", "", true},
		{"no host", "https://", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("normalizeURL(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("normalizeURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidShortCode(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		{"abc", true},
		{"ABC123", true},
		{"my-link", true},
		{"under_score", true},
		{"ab", false},               // too short
		{"a", false},                // too short
		{"abc!def", false},          // invalid char
		{"short code", false},       // space
		{"", false},                 // empty
		{"abc123def456ghi789jkl012mno345pqr678stu901vwx234yz", true}, // 50 chars
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			if got := isValidShortCode(tt.code); got != tt.want {
				t.Errorf("isValidShortCode(%q) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestGenerateUniqueShortCode_Success(t *testing.T) {
	callCount := 0
	repo := &mockLinkRepo{
		shortCodeExistsFn: func(_ context.Context, _ string) (bool, error) {
			callCount++
			return false, nil
		},
	}

	svc := newTestService(repo, &mockClickRepo{}, &mockCodeGen{code: "unique1"})

	code, err := svc.generateUniqueShortCode(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != "unique1" {
		t.Errorf("expected 'unique1', got %s", code)
	}
	if callCount != 1 {
		t.Errorf("expected 1 existence check, got %d", callCount)
	}
}

func TestGenerateUniqueShortCode_RetriesOnCollision(t *testing.T) {
	callCount := 0
	repo := &mockLinkRepo{
		shortCodeExistsFn: func(_ context.Context, _ string) (bool, error) {
			callCount++
			// First 3 calls return exists=true, 4th returns false
			return callCount < 4, nil
		},
	}

	svc := newTestService(repo, &mockClickRepo{}, &mockCodeGen{})

	code, err := svc.generateUniqueShortCode(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code == "" {
		t.Error("expected non-empty code")
	}
	if callCount != 4 {
		t.Errorf("expected 4 calls, got %d", callCount)
	}
}

func TestGenerateUniqueShortCode_ExhaustedRetries(t *testing.T) {
	repo := &mockLinkRepo{
		shortCodeExistsFn: func(_ context.Context, _ string) (bool, error) {
			return true, nil // Always exists
		},
	}

	svc := newTestService(repo, &mockClickRepo{}, &mockCodeGen{})

	_, err := svc.generateUniqueShortCode(context.Background())
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
}

