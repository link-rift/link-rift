package service

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/link-rift/link-rift/internal/config"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/pkg/crypto"
	"github.com/link-rift/link-rift/pkg/httputil"
	"github.com/link-rift/link-rift/pkg/shortcode"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const maxShortCodeRetries = 5

type LinkService interface {
	CreateLink(ctx context.Context, userID, workspaceID uuid.UUID, input models.CreateLinkInput) (*models.Link, error)
	UpdateLink(ctx context.Context, id, workspaceID uuid.UUID, input models.UpdateLinkInput) (*models.Link, error)
	DeleteLink(ctx context.Context, id, workspaceID uuid.UUID) error
	GetLink(ctx context.Context, id uuid.UUID) (*models.Link, error)
	ListLinks(ctx context.Context, workspaceID uuid.UUID, filter models.LinkFilter, pagination models.Pagination) (*models.LinkListResult, error)
	BulkCreateLinks(ctx context.Context, userID, workspaceID uuid.UUID, input models.BulkCreateLinkInput) ([]*models.Link, error)
	GetQuickStats(ctx context.Context, id uuid.UUID) (*models.LinkQuickStats, error)
	CheckShortCodeAvailable(ctx context.Context, code string) (bool, error)
	VerifyLinkPassword(ctx context.Context, shortCode, password string) (bool, error)
}

type linkService struct {
	linkRepo  repository.LinkRepository
	clickRepo repository.ClickRepository
	pool      *pgxpool.Pool
	redis     *redis.Client
	cfg       *config.Config
	codeGen   shortcode.Generator
	events    EventPublisher
	logger    *zap.Logger
}

func NewLinkService(
	linkRepo repository.LinkRepository,
	clickRepo repository.ClickRepository,
	pool *pgxpool.Pool,
	redisClient *redis.Client,
	cfg *config.Config,
	events EventPublisher,
	logger *zap.Logger,
) LinkService {
	return &linkService{
		linkRepo:  linkRepo,
		clickRepo: clickRepo,
		pool:      pool,
		redis:     redisClient,
		cfg:       cfg,
		codeGen:   shortcode.NewGenerator(),
		events:    events,
		logger:    logger,
	}
}

func (s *linkService) CreateLink(ctx context.Context, userID, workspaceID uuid.UUID, input models.CreateLinkInput) (*models.Link, error) {
	normalizedURL, err := normalizeURL(input.URL)
	if err != nil {
		return nil, httputil.Validation("url", "invalid URL format")
	}

	// Generate or validate short code
	var code string
	if input.ShortCode != nil && *input.ShortCode != "" {
		code = *input.ShortCode
		if !isValidShortCode(code) {
			return nil, httputil.Validation("short_code", "short code must be 3-50 alphanumeric characters, hyphens, or underscores")
		}
		exists, err := s.linkRepo.ShortCodeExists(ctx, code)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, httputil.AlreadyExists("short_code")
		}
	} else {
		code, err = s.generateUniqueShortCode(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Hash password if provided
	var passwordHash pgtype.Text
	if input.Password != nil && *input.Password != "" {
		hash, err := crypto.HashPassword(*input.Password)
		if err != nil {
			return nil, httputil.Wrap(err, "failed to hash password")
		}
		passwordHash = pgtype.Text{String: hash, Valid: true}
	}

	// Parse expires_at
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

	params := sqlc.CreateLinkParams{
		UserID:       userID,
		WorkspaceID:  workspaceID,
		Url:          normalizedURL,
		ShortCode:    code,
		Title:        models.OptionalText(input.Title),
		Description:  models.OptionalText(input.Description),
		IsActive:     true,
		PasswordHash: passwordHash,
		ExpiresAt:    expiresAt,
		MaxClicks:    models.OptionalInt4(input.MaxClicks),
		UtmSource:    models.OptionalText(input.UTMSource),
		UtmMedium:    models.OptionalText(input.UTMMedium),
		UtmCampaign:  models.OptionalText(input.UTMCampaign),
		UtmTerm:      models.OptionalText(input.UTMTerm),
		UtmContent:   models.OptionalText(input.UTMContent),
	}

	link, err := s.linkRepo.Create(ctx, params)
	if err != nil {
		return nil, err
	}

	// Publish webhook event (best-effort)
	if err := s.events.Publish(ctx, "link.created", workspaceID, link); err != nil {
		s.logger.Warn("failed to publish link.created event", zap.Error(err))
	}

	return link, nil
}

func (s *linkService) UpdateLink(ctx context.Context, id, workspaceID uuid.UUID, input models.UpdateLinkInput) (*models.Link, error) {
	existing, err := s.linkRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existing.WorkspaceID != workspaceID {
		return nil, httputil.Forbidden("link does not belong to this workspace")
	}

	// If URL is being updated, validate it
	var urlText pgtype.Text
	if input.URL != nil {
		normalizedURL, err := normalizeURL(*input.URL)
		if err != nil {
			return nil, httputil.Validation("url", "invalid URL format")
		}
		urlText = pgtype.Text{String: normalizedURL, Valid: true}
	}

	// Hash password if being updated
	var passwordHash pgtype.Text
	if input.Password != nil {
		if *input.Password == "" {
			// Empty string clears the password
			passwordHash = pgtype.Text{String: "", Valid: true}
		} else {
			hash, err := crypto.HashPassword(*input.Password)
			if err != nil {
				return nil, httputil.Wrap(err, "failed to hash password")
			}
			passwordHash = pgtype.Text{String: hash, Valid: true}
		}
	}

	// Parse expires_at
	var expiresAt pgtype.Timestamptz
	if input.ExpiresAt != nil {
		if *input.ExpiresAt == "" {
			// Empty string removes expiration - set to far future to clear via COALESCE
		} else {
			t, err := time.Parse(time.RFC3339, *input.ExpiresAt)
			if err != nil {
				return nil, httputil.Validation("expires_at", "invalid date format, use RFC3339")
			}
			expiresAt = pgtype.Timestamptz{Time: t, Valid: true}
		}
	}

	params := sqlc.UpdateLinkParams{
		ID:           id,
		Title:        models.OptionalText(input.Title),
		Description:  models.OptionalText(input.Description),
		Url:          urlText,
		IsActive:     models.OptionalBool(input.IsActive),
		PasswordHash: passwordHash,
		ExpiresAt:    expiresAt,
		MaxClicks:    models.OptionalInt4(input.MaxClicks),
	}

	link, err := s.linkRepo.Update(ctx, params)
	if err != nil {
		return nil, err
	}

	// Publish webhook event (best-effort)
	if err := s.events.Publish(ctx, "link.updated", workspaceID, link); err != nil {
		s.logger.Warn("failed to publish link.updated event", zap.Error(err))
	}

	return link, nil
}

func (s *linkService) DeleteLink(ctx context.Context, id, workspaceID uuid.UUID) error {
	existing, err := s.linkRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if existing.WorkspaceID != workspaceID {
		return httputil.Forbidden("link does not belong to this workspace")
	}

	if err := s.linkRepo.SoftDelete(ctx, id); err != nil {
		return err
	}

	// Publish webhook event (best-effort)
	if err := s.events.Publish(ctx, "link.deleted", workspaceID, existing); err != nil {
		s.logger.Warn("failed to publish link.deleted event", zap.Error(err))
	}

	return nil
}

func (s *linkService) GetLink(ctx context.Context, id uuid.UUID) (*models.Link, error) {
	return s.linkRepo.GetByID(ctx, id)
}

func (s *linkService) ListLinks(ctx context.Context, workspaceID uuid.UUID, filter models.LinkFilter, pagination models.Pagination) (*models.LinkListResult, error) {
	if pagination.Limit == 0 {
		pagination.Limit = 20
	}

	params := sqlc.ListLinksForWorkspaceParams{
		WorkspaceID: workspaceID,
		Limit:       int32(pagination.Limit),
		Offset:      int32(pagination.Offset),
		Search:      models.OptionalText(filter.Search),
	}

	links, total, err := s.linkRepo.List(ctx, params)
	if err != nil {
		return nil, err
	}

	redirectBaseURL := s.cfg.App.RedirectURL
	responses := make([]*models.LinkResponse, 0, len(links))
	for _, link := range links {
		responses = append(responses, link.ToResponse(redirectBaseURL))
	}

	return &models.LinkListResult{
		Links: responses,
		Total: total,
	}, nil
}

func (s *linkService) BulkCreateLinks(ctx context.Context, userID, workspaceID uuid.UUID, input models.BulkCreateLinkInput) ([]*models.Link, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback(ctx)

	qtx := sqlc.New(tx)
	txLinkRepo := repository.NewLinkRepository(qtx, s.logger)

	links := make([]*models.Link, 0, len(input.Links))
	for i, linkInput := range input.Links {
		normalizedURL, err := normalizeURL(linkInput.URL)
		if err != nil {
			return nil, httputil.Validation("url", "invalid URL at index "+string(rune('0'+i)))
		}

		var code string
		if linkInput.ShortCode != nil && *linkInput.ShortCode != "" {
			code = *linkInput.ShortCode
		} else {
			code, err = s.generateUniqueShortCode(ctx)
			if err != nil {
				return nil, err
			}
		}

		var passwordHash pgtype.Text
		if linkInput.Password != nil && *linkInput.Password != "" {
			hash, err := crypto.HashPassword(*linkInput.Password)
			if err != nil {
				return nil, httputil.Wrap(err, "failed to hash password")
			}
			passwordHash = pgtype.Text{String: hash, Valid: true}
		}

		var expiresAt pgtype.Timestamptz
		if linkInput.ExpiresAt != nil && *linkInput.ExpiresAt != "" {
			t, err := time.Parse(time.RFC3339, *linkInput.ExpiresAt)
			if err != nil {
				return nil, httputil.Validation("expires_at", "invalid date format at index "+string(rune('0'+i)))
			}
			expiresAt = pgtype.Timestamptz{Time: t, Valid: true}
		}

		params := sqlc.CreateLinkParams{
			UserID:       userID,
			WorkspaceID:  workspaceID,
			Url:          normalizedURL,
			ShortCode:    code,
			Title:        models.OptionalText(linkInput.Title),
			Description:  models.OptionalText(linkInput.Description),
			IsActive:     true,
			PasswordHash: passwordHash,
			ExpiresAt:    expiresAt,
			MaxClicks:    models.OptionalInt4(linkInput.MaxClicks),
			UtmSource:    models.OptionalText(linkInput.UTMSource),
			UtmMedium:    models.OptionalText(linkInput.UTMMedium),
			UtmCampaign:  models.OptionalText(linkInput.UTMCampaign),
			UtmTerm:      models.OptionalText(linkInput.UTMTerm),
			UtmContent:   models.OptionalText(linkInput.UTMContent),
		}

		link, err := txLinkRepo.Create(ctx, params)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, httputil.Wrap(err, "failed to commit transaction")
	}

	return links, nil
}

func (s *linkService) GetQuickStats(ctx context.Context, id uuid.UUID) (*models.LinkQuickStats, error) {
	return s.linkRepo.GetQuickStats(ctx, id)
}

func (s *linkService) CheckShortCodeAvailable(ctx context.Context, code string) (bool, error) {
	exists, err := s.linkRepo.ShortCodeExists(ctx, code)
	if err != nil {
		return false, err
	}
	return !exists, nil
}

func (s *linkService) VerifyLinkPassword(ctx context.Context, shortCode, password string) (bool, error) {
	link, err := s.linkRepo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return false, err
	}

	if link.PasswordHash == nil {
		return true, nil
	}

	match, err := crypto.VerifyPassword(password, *link.PasswordHash)
	if err != nil {
		return false, httputil.Wrap(err, "failed to verify password")
	}
	return match, nil
}

func (s *linkService) generateUniqueShortCode(ctx context.Context) (string, error) {
	for i := 0; i < maxShortCodeRetries; i++ {
		code := s.codeGen.Generate()
		exists, err := s.linkRepo.ShortCodeExists(ctx, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
	return "", httputil.Wrap(errors.New("short code generation failed"), "failed to generate unique short code after retries")
}

func normalizeURL(rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", errors.New("empty URL")
	}

	// Add scheme if missing
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	if parsed.Host == "" {
		return "", errors.New("missing host")
	}

	return parsed.String(), nil
}

func isValidShortCode(code string) bool {
	if len(code) < 3 || len(code) > 50 {
		return false
	}
	for _, c := range code {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}
