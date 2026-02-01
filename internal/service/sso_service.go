package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/link-rift/link-rift/internal/config"
	"github.com/link-rift/link-rift/internal/ee/sso"
	"github.com/link-rift/link-rift/internal/license"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

type SSOService interface {
	// Config management (admin)
	GetConfig(ctx context.Context, workspaceID uuid.UUID) (*models.SSOConfig, error)
	CreateConfig(ctx context.Context, workspaceID uuid.UUID, input models.CreateSSOConfigInput) (*models.SSOConfig, error)
	UpdateConfig(ctx context.Context, workspaceID uuid.UUID, input models.UpdateSSOConfigInput) (*models.SSOConfig, error)
	DeleteConfig(ctx context.Context, workspaceID uuid.UUID) error

	// SSO flow (public)
	GetSPMetadata(ctx context.Context, workspaceID uuid.UUID) (string, error)
	InitiateSSO(ctx context.Context, workspaceID uuid.UUID) (string, error)
	HandleCallback(ctx context.Context, workspaceID uuid.UUID, samlResponse string) (*models.SSOIdentity, error)

	// Identity management
	ListIdentities(ctx context.Context, workspaceID uuid.UUID) ([]*models.SSOIdentity, error)

	SetAuditLogger(logger AuditLogger)
}

type ssoService struct {
	ssoRepo     repository.SSORepository
	licManager  *license.Manager
	cfg         *config.Config
	auditLogger AuditLogger
	logger      *zap.Logger
}

func NewSSOService(
	ssoRepo repository.SSORepository,
	licManager *license.Manager,
	cfg *config.Config,
	logger *zap.Logger,
) SSOService {
	return &ssoService{
		ssoRepo:     ssoRepo,
		licManager:  licManager,
		cfg:         cfg,
		auditLogger: noopAuditLogger{},
		logger:      logger,
	}
}

func (s *ssoService) SetAuditLogger(logger AuditLogger) {
	s.auditLogger = logger
}

func (s *ssoService) GetConfig(ctx context.Context, workspaceID uuid.UUID) (*models.SSOConfig, error) {
	if !s.licManager.HasFeature(license.FeatureSAML) {
		return nil, httputil.PaymentRequiredWithDetails(string(license.FeatureSAML), string(license.TierEnterprise))
	}
	return s.ssoRepo.GetConfig(ctx, workspaceID)
}

func (s *ssoService) CreateConfig(ctx context.Context, workspaceID uuid.UUID, input models.CreateSSOConfigInput) (*models.SSOConfig, error) {
	if !s.licManager.HasFeature(license.FeatureSAML) {
		return nil, httputil.PaymentRequiredWithDetails(string(license.FeatureSAML), string(license.TierEnterprise))
	}

	// Validate certificate
	if _, err := sso.ParseCertificate(input.Certificate); err != nil {
		return nil, httputil.Validation("certificate", "invalid X.509 certificate")
	}

	attrMapping := input.AttributeMapping
	if len(attrMapping) == 0 {
		attrMapping, _ = json.Marshal(sso.DefaultAttributeMapping())
	}

	params := sqlc.CreateSSOConfigParams{
		WorkspaceID:      workspaceID,
		Provider:         input.Provider,
		EntityID:         input.EntityID,
		SsoUrl:           input.SSOURL,
		SloUrl:           pgtype.Text{String: input.SLOURL, Valid: input.SLOURL != ""},
		Certificate:      input.Certificate,
		IdpMetadataUrl:   pgtype.Text{String: input.IDPMetadataURL, Valid: input.IDPMetadataURL != ""},
		IdpMetadataXml:   pgtype.Text{String: input.IDPMetadataXML, Valid: input.IDPMetadataXML != ""},
		AttributeMapping: attrMapping,
		IsEnabled:        input.IsEnabled,
		EnforceSso:       input.EnforceSSO,
		AllowedDomains:   input.AllowedDomains,
	}
	if params.AllowedDomains == nil {
		params.AllowedDomains = []string{}
	}

	cfg, err := s.ssoRepo.CreateConfig(ctx, params)
	if err != nil {
		return nil, err
	}

	s.auditLogger.Log(ctx, AuditEntry{
		WorkspaceID:  workspaceID,
		Action:       "create",
		ResourceType: "sso_config",
		ResourceID:   cfg.ID,
	})

	return cfg, nil
}

func (s *ssoService) UpdateConfig(ctx context.Context, workspaceID uuid.UUID, input models.UpdateSSOConfigInput) (*models.SSOConfig, error) {
	if !s.licManager.HasFeature(license.FeatureSAML) {
		return nil, httputil.PaymentRequiredWithDetails(string(license.FeatureSAML), string(license.TierEnterprise))
	}

	existing, err := s.ssoRepo.GetConfig(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Merge update fields
	provider := existing.Provider
	if input.Provider != nil {
		provider = *input.Provider
	}
	entityID := existing.EntityID
	if input.EntityID != nil {
		entityID = *input.EntityID
	}
	ssoURL := existing.SSOURL
	if input.SSOURL != nil {
		ssoURL = *input.SSOURL
	}
	slourl := existing.SLOURL
	if input.SLOURL != nil {
		slourl = *input.SLOURL
	}
	cert := existing.Certificate
	if input.Certificate != nil {
		if _, err := sso.ParseCertificate(*input.Certificate); err != nil {
			return nil, httputil.Validation("certificate", "invalid X.509 certificate")
		}
		cert = *input.Certificate
	}
	idpMetaURL := existing.IDPMetadataURL
	if input.IDPMetadataURL != nil {
		idpMetaURL = *input.IDPMetadataURL
	}
	idpMetaXML := existing.IDPMetadataXML
	if input.IDPMetadataXML != nil {
		idpMetaXML = *input.IDPMetadataXML
	}
	attrMapping := existing.AttributeMapping
	if input.AttributeMapping != nil {
		attrMapping = *input.AttributeMapping
	}
	isEnabled := existing.IsEnabled
	if input.IsEnabled != nil {
		isEnabled = *input.IsEnabled
	}
	enforceSSO := existing.EnforceSSO
	if input.EnforceSSO != nil {
		enforceSSO = *input.EnforceSSO
	}
	allowedDomains := existing.AllowedDomains
	if input.AllowedDomains != nil {
		allowedDomains = *input.AllowedDomains
	}
	if allowedDomains == nil {
		allowedDomains = []string{}
	}

	params := sqlc.UpdateSSOConfigParams{
		WorkspaceID:      workspaceID,
		Provider:         provider,
		EntityID:         entityID,
		SsoUrl:           ssoURL,
		SloUrl:           pgtype.Text{String: slourl, Valid: slourl != ""},
		Certificate:      cert,
		IdpMetadataUrl:   pgtype.Text{String: idpMetaURL, Valid: idpMetaURL != ""},
		IdpMetadataXml:   pgtype.Text{String: idpMetaXML, Valid: idpMetaXML != ""},
		AttributeMapping: attrMapping,
		IsEnabled:        isEnabled,
		EnforceSso:       enforceSSO,
		AllowedDomains:   allowedDomains,
	}

	cfg, err := s.ssoRepo.UpdateConfig(ctx, params)
	if err != nil {
		return nil, err
	}

	s.auditLogger.Log(ctx, AuditEntry{
		WorkspaceID:  workspaceID,
		Action:       "update",
		ResourceType: "sso_config",
		ResourceID:   cfg.ID,
	})

	return cfg, nil
}

func (s *ssoService) DeleteConfig(ctx context.Context, workspaceID uuid.UUID) error {
	if !s.licManager.HasFeature(license.FeatureSAML) {
		return httputil.PaymentRequiredWithDetails(string(license.FeatureSAML), string(license.TierEnterprise))
	}

	if err := s.ssoRepo.DeleteConfig(ctx, workspaceID); err != nil {
		return err
	}

	s.auditLogger.Log(ctx, AuditEntry{
		WorkspaceID:  workspaceID,
		Action:       "delete",
		ResourceType: "sso_config",
	})

	return nil
}

func (s *ssoService) GetSPMetadata(ctx context.Context, workspaceID uuid.UUID) (string, error) {
	cfg, err := s.ssoRepo.GetConfig(ctx, workspaceID)
	if err != nil {
		return "", err
	}

	acsURL := fmt.Sprintf("%s/api/v1/auth/sso/callback/%s", s.cfg.App.BaseURL, workspaceID.String())
	entityID := fmt.Sprintf("%s/api/v1/auth/sso/metadata/%s", s.cfg.App.BaseURL, workspaceID.String())

	provider, err := sso.NewSAMLProvider(sso.SAMLConfig{
		EntityID:    entityID,
		SSOURL:      cfg.SSOURL,
		SLOURL:      cfg.SLOURL,
		CertificatePEM: cfg.Certificate,
		ACSURL:      acsURL,
		MetadataURL: entityID,
	})
	if err != nil {
		return "", httputil.Wrap(err, "failed to create SAML provider")
	}

	return provider.GenerateMetadataXML()
}

func (s *ssoService) InitiateSSO(ctx context.Context, workspaceID uuid.UUID) (string, error) {
	cfg, err := s.ssoRepo.GetConfig(ctx, workspaceID)
	if err != nil {
		return "", err
	}

	if !cfg.IsEnabled {
		return "", httputil.Validation("sso", "SSO is not enabled for this workspace")
	}

	acsURL := fmt.Sprintf("%s/api/v1/auth/sso/callback/%s", s.cfg.App.BaseURL, workspaceID.String())
	entityID := fmt.Sprintf("%s/api/v1/auth/sso/metadata/%s", s.cfg.App.BaseURL, workspaceID.String())

	provider, err := sso.NewSAMLProvider(sso.SAMLConfig{
		EntityID:         entityID,
		SSOURL:           cfg.SSOURL,
		SLOURL:           cfg.SLOURL,
		CertificatePEM:   cfg.Certificate,
		ACSURL:           acsURL,
		MetadataURL:      entityID,
		AttributeMapping: sso.ParseAttributeMapping(cfg.AttributeMapping),
	})
	if err != nil {
		return "", httputil.Wrap(err, "failed to create SAML provider")
	}

	requestID := uuid.New().String()
	return provider.GenerateAuthURL(requestID)
}

func (s *ssoService) HandleCallback(ctx context.Context, workspaceID uuid.UUID, samlResponse string) (*models.SSOIdentity, error) {
	cfg, err := s.ssoRepo.GetConfig(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	if !cfg.IsEnabled {
		return nil, httputil.Validation("sso", "SSO is not enabled for this workspace")
	}

	acsURL := fmt.Sprintf("%s/api/v1/auth/sso/callback/%s", s.cfg.App.BaseURL, workspaceID.String())
	entityID := fmt.Sprintf("%s/api/v1/auth/sso/metadata/%s", s.cfg.App.BaseURL, workspaceID.String())

	provider, err := sso.NewSAMLProvider(sso.SAMLConfig{
		EntityID:         entityID,
		SSOURL:           cfg.SSOURL,
		SLOURL:           cfg.SLOURL,
		CertificatePEM:   cfg.Certificate,
		ACSURL:           acsURL,
		MetadataURL:      entityID,
		AttributeMapping: sso.ParseAttributeMapping(cfg.AttributeMapping),
	})
	if err != nil {
		return nil, httputil.Wrap(err, "failed to create SAML provider")
	}

	assertion, err := provider.ValidateResponse(samlResponse)
	if err != nil {
		return nil, httputil.Wrap(err, "SAML response validation failed")
	}

	// Look up or create the SSO identity
	identity, err := s.ssoRepo.GetIdentityByExternalID(ctx, workspaceID, assertion.NameID)
	if err == nil {
		// Update last login
		attrsJSON, _ := json.Marshal(assertion.Attributes)
		_ = s.ssoRepo.UpdateIdentityLastLogin(ctx, identity.UserID, workspaceID, assertion.Email, assertion.Name(), attrsJSON)
		return identity, nil
	}

	// Identity not found — this would require creating a user and identity
	// which is handled by the caller (auth service / handler)
	return nil, httputil.NotFound("sso_identity")
}

func (s *ssoService) ListIdentities(ctx context.Context, workspaceID uuid.UUID) ([]*models.SSOIdentity, error) {
	if !s.licManager.HasFeature(license.FeatureSAML) {
		return nil, httputil.PaymentRequiredWithDetails(string(license.FeatureSAML), string(license.TierEnterprise))
	}
	return s.ssoRepo.ListIdentities(ctx, workspaceID)
}
