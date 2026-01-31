package service

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/link-rift/link-rift/internal/license"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

type BioPageService interface {
	// CRUD
	CreateBioPage(ctx context.Context, workspaceID uuid.UUID, input models.CreateBioPageInput) (*models.BioPage, error)
	GetBioPage(ctx context.Context, id uuid.UUID) (*models.BioPage, error)
	ListBioPages(ctx context.Context, workspaceID uuid.UUID) ([]*models.BioPage, error)
	UpdateBioPage(ctx context.Context, id, workspaceID uuid.UUID, input models.UpdateBioPageInput) (*models.BioPage, error)
	DeleteBioPage(ctx context.Context, id, workspaceID uuid.UUID) error

	// Publish
	PublishBioPage(ctx context.Context, id, workspaceID uuid.UUID) (*models.BioPage, error)
	UnpublishBioPage(ctx context.Context, id, workspaceID uuid.UUID) (*models.BioPage, error)

	// Links
	AddLink(ctx context.Context, pageID, workspaceID uuid.UUID, input models.CreateBioPageLinkInput) (*models.BioPageLink, error)
	UpdateLink(ctx context.Context, pageID, linkID, workspaceID uuid.UUID, input models.UpdateBioPageLinkInput) (*models.BioPageLink, error)
	DeleteLink(ctx context.Context, pageID, linkID, workspaceID uuid.UUID) error
	ListLinks(ctx context.Context, pageID uuid.UUID) ([]*models.BioPageLink, error)
	ReorderLinks(ctx context.Context, pageID, workspaceID uuid.UUID, input models.ReorderBioLinksInput) error
	TrackLinkClick(ctx context.Context, linkID uuid.UUID) error

	// Themes
	ListThemes() []models.BioPageTheme
	GetTheme(themeID string) (*models.BioPageTheme, error)

	// Public
	GetPublicPage(ctx context.Context, slug string) (*models.PublicBioPageResponse, error)
}

type bioPageService struct {
	bioPageRepo repository.BioPageRepository
	licManager  *license.Manager
	logger      *zap.Logger
}

func NewBioPageService(
	bioPageRepo repository.BioPageRepository,
	licManager *license.Manager,
	logger *zap.Logger,
) BioPageService {
	return &bioPageService{
		bioPageRepo: bioPageRepo,
		licManager:  licManager,
		logger:      logger,
	}
}

var slugRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$`)

func (s *bioPageService) CreateBioPage(ctx context.Context, workspaceID uuid.UUID, input models.CreateBioPageInput) (*models.BioPage, error) {
	// Check license
	if !s.licManager.HasFeature(license.FeatureBioPages) {
		return nil, httputil.PaymentRequiredWithDetails("bio_pages", "pro")
	}

	// Validate slug
	slug := strings.ToLower(strings.TrimSpace(input.Slug))
	if !isValidSlug(slug) {
		return nil, httputil.Validation("slug", "slug must be lowercase alphanumeric with hyphens, 1-100 characters")
	}

	// Build create params
	params := sqlc.CreateBioPageParams{
		WorkspaceID: workspaceID,
		Slug:        slug,
		Title:       strings.TrimSpace(input.Title),
	}

	if input.Bio != nil {
		params.Bio = pgtype.Text{String: *input.Bio, Valid: true}
	}
	if input.AvatarURL != nil {
		params.AvatarUrl = pgtype.Text{String: *input.AvatarURL, Valid: true}
	}
	if input.ThemeID != nil {
		themeUUID := models.ThemeIDToUUID(*input.ThemeID)
		params.ThemeID = pgtype.UUID{Bytes: themeUUID, Valid: true}
	}
	if input.MetaTitle != nil {
		params.MetaTitle = pgtype.Text{String: *input.MetaTitle, Valid: true}
	}
	if input.MetaDescription != nil {
		params.MetaDescription = pgtype.Text{String: *input.MetaDescription, Valid: true}
	}
	if input.OgImageURL != nil {
		params.OgImageUrl = pgtype.Text{String: *input.OgImageURL, Valid: true}
	}

	return s.bioPageRepo.Create(ctx, params)
}

func (s *bioPageService) GetBioPage(ctx context.Context, id uuid.UUID) (*models.BioPage, error) {
	page, err := s.bioPageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Attach link count
	links, err := s.bioPageRepo.ListLinks(ctx, id)
	if err == nil {
		page.LinkCount = len(links)
	}

	return page, nil
}

func (s *bioPageService) ListBioPages(ctx context.Context, workspaceID uuid.UUID) ([]*models.BioPage, error) {
	pages, err := s.bioPageRepo.List(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Attach link counts
	for _, page := range pages {
		links, err := s.bioPageRepo.ListLinks(ctx, page.ID)
		if err == nil {
			page.LinkCount = len(links)
		}
	}

	return pages, nil
}

func (s *bioPageService) UpdateBioPage(ctx context.Context, id, workspaceID uuid.UUID, input models.UpdateBioPageInput) (*models.BioPage, error) {
	// Verify ownership
	page, err := s.bioPageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if page.WorkspaceID != workspaceID {
		return nil, httputil.Forbidden("bio page does not belong to this workspace")
	}

	params := sqlc.UpdateBioPageParams{ID: id}

	if input.Slug != nil {
		slug := strings.ToLower(strings.TrimSpace(*input.Slug))
		if !isValidSlug(slug) {
			return nil, httputil.Validation("slug", "slug must be lowercase alphanumeric with hyphens, 1-100 characters")
		}
		params.Slug = pgtype.Text{String: slug, Valid: true}
	}
	if input.Title != nil {
		params.Title = pgtype.Text{String: strings.TrimSpace(*input.Title), Valid: true}
	}
	if input.Bio != nil {
		params.Bio = pgtype.Text{String: *input.Bio, Valid: true}
	}
	if input.AvatarURL != nil {
		params.AvatarUrl = pgtype.Text{String: *input.AvatarURL, Valid: true}
	}
	if input.ThemeID != nil {
		themeUUID := models.ThemeIDToUUID(*input.ThemeID)
		params.ThemeID = pgtype.UUID{Bytes: themeUUID, Valid: true}
	}
	if input.CustomCSS != nil {
		// Check license for custom CSS
		if !s.licManager.HasFeature(license.FeatureCustomCSS) {
			return nil, httputil.PaymentRequiredWithDetails("custom_css", "enterprise")
		}
		sanitized, err := sanitizeCSS(*input.CustomCSS)
		if err != nil {
			return nil, httputil.Validation("custom_css", err.Error())
		}
		params.CustomCss = pgtype.Text{String: sanitized, Valid: true}
	}
	if input.MetaTitle != nil {
		params.MetaTitle = pgtype.Text{String: *input.MetaTitle, Valid: true}
	}
	if input.MetaDescription != nil {
		params.MetaDescription = pgtype.Text{String: *input.MetaDescription, Valid: true}
	}
	if input.OgImageURL != nil {
		params.OgImageUrl = pgtype.Text{String: *input.OgImageURL, Valid: true}
	}

	return s.bioPageRepo.Update(ctx, params)
}

func (s *bioPageService) DeleteBioPage(ctx context.Context, id, workspaceID uuid.UUID) error {
	page, err := s.bioPageRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if page.WorkspaceID != workspaceID {
		return httputil.Forbidden("bio page does not belong to this workspace")
	}

	return s.bioPageRepo.SoftDelete(ctx, id)
}

func (s *bioPageService) PublishBioPage(ctx context.Context, id, workspaceID uuid.UUID) (*models.BioPage, error) {
	page, err := s.bioPageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if page.WorkspaceID != workspaceID {
		return nil, httputil.Forbidden("bio page does not belong to this workspace")
	}

	return s.bioPageRepo.Update(ctx, sqlc.UpdateBioPageParams{
		ID:          id,
		IsPublished: pgtype.Bool{Bool: true, Valid: true},
	})
}

func (s *bioPageService) UnpublishBioPage(ctx context.Context, id, workspaceID uuid.UUID) (*models.BioPage, error) {
	page, err := s.bioPageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if page.WorkspaceID != workspaceID {
		return nil, httputil.Forbidden("bio page does not belong to this workspace")
	}

	return s.bioPageRepo.Update(ctx, sqlc.UpdateBioPageParams{
		ID:          id,
		IsPublished: pgtype.Bool{Bool: false, Valid: true},
	})
}

// Links

func (s *bioPageService) AddLink(ctx context.Context, pageID, workspaceID uuid.UUID, input models.CreateBioPageLinkInput) (*models.BioPageLink, error) {
	// Verify page ownership
	page, err := s.bioPageRepo.GetByID(ctx, pageID)
	if err != nil {
		return nil, err
	}
	if page.WorkspaceID != workspaceID {
		return nil, httputil.Forbidden("bio page does not belong to this workspace")
	}

	// Get next position
	maxPos, err := s.bioPageRepo.GetMaxLinkPosition(ctx, pageID)
	if err != nil {
		return nil, err
	}

	isVisible := true
	if input.IsVisible != nil {
		isVisible = *input.IsVisible
	}

	params := sqlc.CreateBioPageLinkParams{
		BioPageID: pageID,
		Title:     strings.TrimSpace(input.Title),
		Url:       strings.TrimSpace(input.URL),
		Position:  maxPos + 1,
		IsVisible: isVisible,
	}

	if input.Icon != nil {
		params.Icon = pgtype.Text{String: *input.Icon, Valid: true}
	}
	if input.VisibleFrom != nil {
		t, err := time.Parse(time.RFC3339, *input.VisibleFrom)
		if err != nil {
			return nil, httputil.Validation("visible_from", "invalid datetime format, use RFC3339")
		}
		params.VisibleFrom = pgtype.Timestamptz{Time: t, Valid: true}
	}
	if input.VisibleUntil != nil {
		t, err := time.Parse(time.RFC3339, *input.VisibleUntil)
		if err != nil {
			return nil, httputil.Validation("visible_until", "invalid datetime format, use RFC3339")
		}
		params.VisibleUntil = pgtype.Timestamptz{Time: t, Valid: true}
	}

	return s.bioPageRepo.CreateLink(ctx, params)
}

func (s *bioPageService) UpdateLink(ctx context.Context, pageID, linkID, workspaceID uuid.UUID, input models.UpdateBioPageLinkInput) (*models.BioPageLink, error) {
	// Verify page ownership
	page, err := s.bioPageRepo.GetByID(ctx, pageID)
	if err != nil {
		return nil, err
	}
	if page.WorkspaceID != workspaceID {
		return nil, httputil.Forbidden("bio page does not belong to this workspace")
	}

	// Verify link belongs to page
	link, err := s.bioPageRepo.GetLinkByID(ctx, linkID)
	if err != nil {
		return nil, err
	}
	if link.BioPageID != pageID {
		return nil, httputil.Forbidden("link does not belong to this bio page")
	}

	params := sqlc.UpdateBioPageLinkParams{ID: linkID}

	if input.Title != nil {
		params.Title = pgtype.Text{String: strings.TrimSpace(*input.Title), Valid: true}
	}
	if input.URL != nil {
		params.Url = pgtype.Text{String: strings.TrimSpace(*input.URL), Valid: true}
	}
	if input.Icon != nil {
		params.Icon = pgtype.Text{String: *input.Icon, Valid: true}
	}
	if input.IsVisible != nil {
		params.IsVisible = pgtype.Bool{Bool: *input.IsVisible, Valid: true}
	}
	if input.VisibleFrom != nil {
		t, err := time.Parse(time.RFC3339, *input.VisibleFrom)
		if err != nil {
			return nil, httputil.Validation("visible_from", "invalid datetime format, use RFC3339")
		}
		params.VisibleFrom = pgtype.Timestamptz{Time: t, Valid: true}
	}
	if input.VisibleUntil != nil {
		t, err := time.Parse(time.RFC3339, *input.VisibleUntil)
		if err != nil {
			return nil, httputil.Validation("visible_until", "invalid datetime format, use RFC3339")
		}
		params.VisibleUntil = pgtype.Timestamptz{Time: t, Valid: true}
	}

	return s.bioPageRepo.UpdateLink(ctx, params)
}

func (s *bioPageService) DeleteLink(ctx context.Context, pageID, linkID, workspaceID uuid.UUID) error {
	// Verify page ownership
	page, err := s.bioPageRepo.GetByID(ctx, pageID)
	if err != nil {
		return err
	}
	if page.WorkspaceID != workspaceID {
		return httputil.Forbidden("bio page does not belong to this workspace")
	}

	// Verify link belongs to page
	link, err := s.bioPageRepo.GetLinkByID(ctx, linkID)
	if err != nil {
		return err
	}
	if link.BioPageID != pageID {
		return httputil.Forbidden("link does not belong to this bio page")
	}

	return s.bioPageRepo.DeleteLink(ctx, linkID)
}

func (s *bioPageService) ListLinks(ctx context.Context, pageID uuid.UUID) ([]*models.BioPageLink, error) {
	return s.bioPageRepo.ListLinks(ctx, pageID)
}

func (s *bioPageService) ReorderLinks(ctx context.Context, pageID, workspaceID uuid.UUID, input models.ReorderBioLinksInput) error {
	// Verify page ownership
	page, err := s.bioPageRepo.GetByID(ctx, pageID)
	if err != nil {
		return err
	}
	if page.WorkspaceID != workspaceID {
		return httputil.Forbidden("bio page does not belong to this workspace")
	}

	for i, linkIDStr := range input.LinkIDs {
		linkID, err := uuid.Parse(linkIDStr)
		if err != nil {
			return httputil.Validation("link_ids", "invalid link ID: "+linkIDStr)
		}

		err = s.bioPageRepo.UpdateLinkPosition(ctx, sqlc.UpdateBioPageLinkPositionParams{
			ID:       linkID,
			Position: int32(i),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *bioPageService) TrackLinkClick(ctx context.Context, linkID uuid.UUID) error {
	return s.bioPageRepo.IncrementLinkClickCount(ctx, linkID)
}

// Themes

func (s *bioPageService) ListThemes() []models.BioPageTheme {
	themes := make([]models.BioPageTheme, 0, len(models.PredefinedThemes))
	for _, theme := range models.PredefinedThemes {
		themes = append(themes, theme)
	}
	return themes
}

func (s *bioPageService) GetTheme(themeID string) (*models.BioPageTheme, error) {
	theme, ok := models.PredefinedThemes[themeID]
	if !ok {
		return nil, httputil.NotFound("theme")
	}
	return &theme, nil
}

// Public

func (s *bioPageService) GetPublicPage(ctx context.Context, slug string) (*models.PublicBioPageResponse, error) {
	page, err := s.bioPageRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if !page.IsPublished {
		return nil, httputil.NotFound("bio page")
	}

	// Get visible links
	allLinks, err := s.bioPageRepo.ListLinks(ctx, page.ID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	publicLinks := make([]models.PublicBioLink, 0)
	for _, link := range allLinks {
		if !link.IsVisible {
			continue
		}
		if link.VisibleFrom != nil && now.Before(*link.VisibleFrom) {
			continue
		}
		if link.VisibleUntil != nil && now.After(*link.VisibleUntil) {
			continue
		}
		publicLinks = append(publicLinks, models.PublicBioLink{
			ID:    link.ID,
			Title: link.Title,
			URL:   link.URL,
			Icon:  link.Icon,
		})
	}

	resp := &models.PublicBioPageResponse{
		Title:           page.Title,
		Bio:             page.Bio,
		AvatarURL:       page.AvatarURL,
		Slug:            page.Slug,
		CustomCSS:       page.CustomCSS,
		MetaTitle:       page.MetaTitle,
		MetaDescription: page.MetaDescription,
		OgImageURL:      page.OgImageURL,
		Links:           publicLinks,
	}

	// Resolve theme
	if page.ThemeID != nil {
		themeKey := models.ThemeUUIDToID(*page.ThemeID)
		if themeKey != "" {
			theme := models.PredefinedThemes[themeKey]
			resp.Theme = &theme
		}
	}

	return resp, nil
}

// Helpers

func isValidSlug(slug string) bool {
	if len(slug) == 0 || len(slug) > 100 {
		return false
	}
	return slugRegex.MatchString(slug)
}

// sanitizeCSS removes dangerous CSS patterns.
func sanitizeCSS(css string) (string, error) {
	lower := strings.ToLower(css)

	dangerousPatterns := []string{
		"javascript:",
		"expression(",
		"@import",
		"data:",
		"behavior:",
		"-moz-binding",
		"url(",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(lower, pattern) {
			return "", httputil.Validation("custom_css", "CSS contains forbidden pattern: "+pattern)
		}
	}

	return css, nil
}
