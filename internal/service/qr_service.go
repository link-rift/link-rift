package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/link-rift/link-rift/internal/config"
	"github.com/link-rift/link-rift/internal/license"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/qrcode"
	"github.com/link-rift/link-rift/internal/repository"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/pkg/httputil"
	"github.com/link-rift/link-rift/pkg/storage"
	"go.uber.org/zap"
)

type QRCodeService interface {
	CreateQRCode(ctx context.Context, linkID, workspaceID uuid.UUID, input models.CreateQRCodeInput) (*models.QRCode, error)
	GetQRCode(ctx context.Context, id uuid.UUID) (*models.QRCode, error)
	GetQRCodeForLink(ctx context.Context, linkID uuid.UUID) (*models.QRCode, error)
	DownloadQRCode(ctx context.Context, linkID uuid.UUID, format string) ([]byte, string, error)
	DeleteQRCode(ctx context.Context, id uuid.UUID) error
	BulkGenerateQRCodes(ctx context.Context, workspaceID uuid.UUID, input models.BulkQRCodeInput) (*qrcode.BatchResult, error)
	GetStyleTemplates() map[string]qrcode.StyleTemplate
}

type qrCodeService struct {
	qrRepo     repository.QRCodeRepository
	linkRepo   repository.LinkRepository
	generator  *qrcode.Generator
	batchGen   *qrcode.BatchGenerator
	store      storage.ObjectStorage
	licManager *license.Manager
	cfg        *config.Config
	logger     *zap.Logger
}

func NewQRCodeService(
	qrRepo repository.QRCodeRepository,
	linkRepo repository.LinkRepository,
	generator *qrcode.Generator,
	batchGen *qrcode.BatchGenerator,
	store storage.ObjectStorage,
	licManager *license.Manager,
	cfg *config.Config,
	logger *zap.Logger,
) QRCodeService {
	return &qrCodeService{
		qrRepo:     qrRepo,
		linkRepo:   linkRepo,
		generator:  generator,
		batchGen:   batchGen,
		store:      store,
		licManager: licManager,
		cfg:        cfg,
		logger:     logger,
	}
}

func (s *qrCodeService) CreateQRCode(ctx context.Context, linkID, workspaceID uuid.UUID, input models.CreateQRCodeInput) (*models.QRCode, error) {
	// Verify link exists and belongs to workspace
	link, err := s.linkRepo.GetByID(ctx, linkID)
	if err != nil {
		return nil, err
	}
	if link.WorkspaceID != workspaceID {
		return nil, httputil.Forbidden("link does not belong to this workspace")
	}

	// Check if customization requires Pro tier
	if isCustomized(input) {
		if !s.licManager.HasFeature(license.FeatureQRCustomization) {
			return nil, httputil.PaymentRequiredWithDetails("qr_customization", "pro")
		}
	}

	// Set defaults
	if input.QRType == "" {
		input.QRType = "dynamic"
	}
	if input.ErrorCorrection == "" {
		input.ErrorCorrection = "M"
	}
	if input.ForegroundColor == "" {
		input.ForegroundColor = "#000000"
	}
	if input.BackgroundColor == "" {
		input.BackgroundColor = "#FFFFFF"
	}
	if input.DotStyle == "" {
		input.DotStyle = "square"
	}
	if input.CornerStyle == "" {
		input.CornerStyle = "square"
	}
	size := int32(512)
	if input.Size != nil {
		size = *input.Size
	}
	margin := int32(4)
	if input.Margin != nil {
		margin = *input.Margin
	}

	// Build URL for QR code
	var targetURL string
	if input.QRType == "dynamic" {
		targetURL = s.cfg.App.RedirectURL + "/" + link.ShortCode
	} else {
		targetURL = link.URL
	}

	// Generate QR PNG
	opts := qrcode.Options{
		Size:            int(size),
		ErrorCorrection: input.ErrorCorrection,
		ForegroundColor: input.ForegroundColor,
		BackgroundColor: input.BackgroundColor,
		LogoURL:         stringFromPtr(input.LogoURL),
		DotStyle:        input.DotStyle,
		CornerStyle:     input.CornerStyle,
		Margin:          int(margin),
	}

	qrID := uuid.New()
	storageKey := fmt.Sprintf("qr/%s/%s.png", linkID.String(), qrID.String())

	pngURL, err := s.generator.GenerateAndUpload(ctx, targetURL, storageKey, opts)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to generate QR code")
	}

	// Create DB record
	params := sqlc.CreateQRCodeParams{
		LinkID:          linkID,
		QrType:          input.QRType,
		ErrorCorrection: input.ErrorCorrection,
		ForegroundColor: input.ForegroundColor,
		BackgroundColor: input.BackgroundColor,
		LogoUrl:         models.OptionalText(input.LogoURL),
		PngUrl:          pgtype.Text{String: pngURL, Valid: true},
		DotStyle:        input.DotStyle,
		CornerStyle:     input.CornerStyle,
		Size:            size,
		Margin:          margin,
	}

	qr, err := s.qrRepo.Create(ctx, params)
	if err != nil {
		return nil, err
	}

	return qr, nil
}

func (s *qrCodeService) GetQRCode(ctx context.Context, id uuid.UUID) (*models.QRCode, error) {
	return s.qrRepo.GetByID(ctx, id)
}

func (s *qrCodeService) GetQRCodeForLink(ctx context.Context, linkID uuid.UUID) (*models.QRCode, error) {
	return s.qrRepo.GetByLinkID(ctx, linkID)
}

func (s *qrCodeService) DownloadQRCode(ctx context.Context, linkID uuid.UUID, format string) ([]byte, string, error) {
	qr, err := s.qrRepo.GetByLinkID(ctx, linkID)
	if err != nil {
		return nil, "", err
	}

	// Get the link to build URL
	link, err := s.linkRepo.GetByID(ctx, linkID)
	if err != nil {
		return nil, "", err
	}

	var targetURL string
	if qr.QRType == "dynamic" {
		targetURL = s.cfg.App.RedirectURL + "/" + link.ShortCode
	} else {
		targetURL = link.URL
	}

	opts := qrcode.Options{
		Size:            int(qr.Size),
		ErrorCorrection: qr.ErrorCorrection,
		ForegroundColor: qr.ForegroundColor,
		BackgroundColor: qr.BackgroundColor,
		DotStyle:        qr.DotStyle,
		CornerStyle:     qr.CornerStyle,
		Margin:          int(qr.Margin),
	}

	if format == "svg" {
		data, err := s.generator.GenerateSVG(targetURL, opts)
		if err != nil {
			return nil, "", httputil.Wrap(err, "failed to generate SVG")
		}
		return data, "image/svg+xml", nil
	}

	// Default: PNG
	data, err := s.generator.Generate(targetURL, opts)
	if err != nil {
		return nil, "", httputil.Wrap(err, "failed to generate PNG")
	}
	return data, "image/png", nil
}

func (s *qrCodeService) DeleteQRCode(ctx context.Context, id uuid.UUID) error {
	qr, err := s.qrRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete from storage if we have a PNG URL
	if qr.PngURL != nil {
		storageKey := fmt.Sprintf("qr/%s/%s.png", qr.LinkID.String(), qr.ID.String())
		if delErr := s.store.Delete(ctx, storageKey); delErr != nil {
			s.logger.Warn("failed to delete QR code from storage", zap.Error(delErr))
		}
	}

	return s.qrRepo.Delete(ctx, id)
}

func (s *qrCodeService) BulkGenerateQRCodes(ctx context.Context, workspaceID uuid.UUID, input models.BulkQRCodeInput) (*qrcode.BatchResult, error) {
	items := make([]qrcode.BatchItem, 0, len(input.LinkIDs))

	for _, linkID := range input.LinkIDs {
		link, err := s.linkRepo.GetByID(ctx, linkID)
		if err != nil {
			continue
		}
		if link.WorkspaceID != workspaceID {
			continue
		}

		var targetURL string
		if input.Options.QRType == "static" {
			targetURL = link.URL
		} else {
			targetURL = s.cfg.App.RedirectURL + "/" + link.ShortCode
		}

		items = append(items, qrcode.BatchItem{
			LinkID: linkID,
			URL:    targetURL,
		})
	}

	if len(items) == 0 {
		return nil, httputil.Validation("link_ids", "no valid links found")
	}

	opts := qrcode.Options{
		Size:            512,
		ErrorCorrection: input.Options.ErrorCorrection,
		ForegroundColor: input.Options.ForegroundColor,
		BackgroundColor: input.Options.BackgroundColor,
		DotStyle:        input.Options.DotStyle,
		CornerStyle:     input.Options.CornerStyle,
		Margin:          4,
	}
	if input.Options.Size != nil {
		opts.Size = int(*input.Options.Size)
	}
	if input.Options.Margin != nil {
		opts.Margin = int(*input.Options.Margin)
	}

	return s.batchGen.GenerateBatch(ctx, items, opts)
}

func (s *qrCodeService) GetStyleTemplates() map[string]qrcode.StyleTemplate {
	return qrcode.StyleTemplates
}

// isCustomized returns true if any non-default customization is set.
func isCustomized(input models.CreateQRCodeInput) bool {
	if input.ForegroundColor != "" && input.ForegroundColor != "#000000" {
		return true
	}
	if input.BackgroundColor != "" && input.BackgroundColor != "#FFFFFF" {
		return true
	}
	if input.LogoURL != nil {
		return true
	}
	if input.DotStyle != "" && input.DotStyle != "square" {
		return true
	}
	if input.CornerStyle != "" && input.CornerStyle != "square" {
		return true
	}
	return false
}

func stringFromPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
