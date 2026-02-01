package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/link-rift/link-rift/internal/ee/whitelabel"
	"github.com/link-rift/link-rift/internal/license"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

type BrandingService interface {
	GetBranding(ctx context.Context, workspaceID uuid.UUID) (*models.WorkspaceBranding, error)
	UpdateBranding(ctx context.Context, workspaceID uuid.UUID, input models.UpdateBrandingInput) (*models.WorkspaceBranding, error)
	DeleteBranding(ctx context.Context, workspaceID uuid.UUID) error
	SetAuditLogger(logger AuditLogger)
}

type brandingService struct {
	brandingRepo repository.BrandingRepository
	licManager   *license.Manager
	auditLogger  AuditLogger
	logger       *zap.Logger
}

func NewBrandingService(
	brandingRepo repository.BrandingRepository,
	licManager *license.Manager,
	logger *zap.Logger,
) BrandingService {
	return &brandingService{
		brandingRepo: brandingRepo,
		licManager:   licManager,
		auditLogger:  noopAuditLogger{},
		logger:       logger,
	}
}

func (s *brandingService) SetAuditLogger(logger AuditLogger) {
	s.auditLogger = logger
}

func (s *brandingService) GetBranding(ctx context.Context, workspaceID uuid.UUID) (*models.WorkspaceBranding, error) {
	if !s.licManager.HasFeature(license.FeatureWhiteLabel) {
		return nil, httputil.PaymentRequiredWithDetails(string(license.FeatureWhiteLabel), string(license.TierEnterprise))
	}

	return s.brandingRepo.Get(ctx, workspaceID)
}

func (s *brandingService) UpdateBranding(ctx context.Context, workspaceID uuid.UUID, input models.UpdateBrandingInput) (*models.WorkspaceBranding, error) {
	if !s.licManager.HasFeature(license.FeatureWhiteLabel) {
		return nil, httputil.PaymentRequiredWithDetails(string(license.FeatureWhiteLabel), string(license.TierEnterprise))
	}

	// Validate colors
	colors := map[string]*string{
		"primary_color":   input.PrimaryColor,
		"secondary_color": input.SecondaryColor,
		"accent_color":    input.AccentColor,
	}
	for field, color := range colors {
		if color != nil && !whitelabel.ValidateHexColor(*color) {
			return nil, httputil.Validation(field, "invalid hex color format")
		}
	}

	// Validate and sanitize CSS (requires custom_css feature)
	if input.CustomCSS != nil && *input.CustomCSS != "" {
		if !s.licManager.HasFeature(license.FeatureCustomCSS) {
			return nil, httputil.PaymentRequiredWithDetails(string(license.FeatureCustomCSS), string(license.TierEnterprise))
		}
		sanitized, ok := whitelabel.SanitizeCSS(*input.CustomCSS)
		if !ok {
			return nil, httputil.Validation("custom_css", "CSS contains disallowed patterns")
		}
		input.CustomCSS = &sanitized
	}

	params := sqlc.UpsertWorkspaceBrandingParams{
		WorkspaceID:      workspaceID,
		LogoUrl:          optionalTextToSqlc(input.LogoURL),
		LogoDarkUrl:      optionalTextToSqlc(input.LogoDarkURL),
		FaviconUrl:       optionalTextToSqlc(input.FaviconURL),
		PrimaryColor:     optionalTextToSqlc(input.PrimaryColor),
		SecondaryColor:   optionalTextToSqlc(input.SecondaryColor),
		AccentColor:      optionalTextToSqlc(input.AccentColor),
		CustomCss:        optionalTextToSqlc(input.CustomCSS),
		HidePoweredBy:    input.HidePoweredBy != nil && *input.HidePoweredBy,
		CustomFooterText: optionalTextToSqlc(input.CustomFooterText),
		CustomFooterUrl:  optionalTextToSqlc(input.CustomFooterURL),
	}

	branding, err := s.brandingRepo.Upsert(ctx, params)
	if err != nil {
		return nil, err
	}

	s.auditLogger.Log(ctx, AuditEntry{
		WorkspaceID:  workspaceID,
		Action:       "update",
		ResourceType: "workspace_branding",
		ResourceID:   branding.ID,
		NewValues:    input,
	})

	return branding, nil
}

func (s *brandingService) DeleteBranding(ctx context.Context, workspaceID uuid.UUID) error {
	if !s.licManager.HasFeature(license.FeatureWhiteLabel) {
		return httputil.PaymentRequiredWithDetails(string(license.FeatureWhiteLabel), string(license.TierEnterprise))
	}

	if err := s.brandingRepo.Delete(ctx, workspaceID); err != nil {
		return err
	}

	s.auditLogger.Log(ctx, AuditEntry{
		WorkspaceID:  workspaceID,
		Action:       "delete",
		ResourceType: "workspace_branding",
	})

	return nil
}

// optionalTextToSqlc converts a *string to pgtype.Text.
func optionalTextToSqlc(s *string) pgtype.Text {
	if s == nil || *s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}
