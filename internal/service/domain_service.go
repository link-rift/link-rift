package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/link-rift/link-rift/internal/config"
	"github.com/link-rift/link-rift/internal/license"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

// DNSResolver abstracts DNS lookups for testability.
type DNSResolver interface {
	LookupTXT(ctx context.Context, name string) ([]string, error)
}

// netResolver wraps net.Resolver to satisfy DNSResolver.
type netResolver struct {
	resolver *net.Resolver
}

func (r *netResolver) LookupTXT(ctx context.Context, name string) ([]string, error) {
	return r.resolver.LookupTXT(ctx, name)
}

type DomainService interface {
	AddDomain(ctx context.Context, workspaceID uuid.UUID, input models.CreateDomainInput) (*models.Domain, error)
	GetDomain(ctx context.Context, id uuid.UUID) (*models.Domain, error)
	ListDomains(ctx context.Context, workspaceID uuid.UUID) ([]*models.Domain, error)
	VerifyDomain(ctx context.Context, id, workspaceID uuid.UUID) (*models.Domain, error)
	RemoveDomain(ctx context.Context, id, workspaceID uuid.UUID) error
	GetDNSRecords(ctx context.Context, id uuid.UUID) (*models.VerificationInstructions, error)
}

type domainService struct {
	domainRepo  repository.DomainRepository
	licManager  *license.Manager
	sslProvider SSLProvider
	dnsResolver DNSResolver
	events      EventPublisher
	cfg         *config.Config
	logger      *zap.Logger
}

func NewDomainService(
	domainRepo repository.DomainRepository,
	licManager *license.Manager,
	sslProvider SSLProvider,
	cfg *config.Config,
	events EventPublisher,
	logger *zap.Logger,
) DomainService {
	return &domainService{
		domainRepo:  domainRepo,
		licManager:  licManager,
		sslProvider: sslProvider,
		dnsResolver: &netResolver{resolver: net.DefaultResolver},
		events:      events,
		cfg:         cfg,
		logger:      logger,
	}
}

func (s *domainService) AddDomain(ctx context.Context, workspaceID uuid.UUID, input models.CreateDomainInput) (*models.Domain, error) {
	// Validate domain format
	domain := strings.TrimSpace(strings.ToLower(input.Domain))
	if !isValidDomainName(domain) {
		return nil, httputil.Validation("domain", "invalid domain format")
	}

	// Check license feature
	if !s.licManager.HasFeature(license.FeatureCustomDomains) {
		return nil, httputil.PaymentRequiredWithDetails("custom_domains", "pro")
	}

	// Check domain count limit
	count, err := s.domainRepo.GetCountForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	if !s.licManager.CheckLimit(license.LimitMaxDomains, count) {
		return nil, httputil.PaymentRequired("domain limit reached, upgrade your plan for more domains")
	}

	// Generate verification token
	token := uuid.New().String()
	dnsData, err := json.Marshal(models.DNSRecordsData{
		VerificationToken: token,
	})
	if err != nil {
		return nil, httputil.Wrap(err, "failed to encode DNS records")
	}

	// Create domain
	d, err := s.domainRepo.Create(ctx, sqlc.CreateDomainParams{
		WorkspaceID: workspaceID,
		Domain:      domain,
	})
	if err != nil {
		return nil, err
	}

	// Store verification token in dns_records JSONB
	d, err = s.domainRepo.Update(ctx, sqlc.UpdateDomainParams{
		ID:         d.ID,
		DnsRecords: dnsData,
	})
	if err != nil {
		return nil, err
	}

	// Publish webhook event (best-effort)
	if err := s.events.Publish(ctx, "domain.added", workspaceID, d); err != nil {
		s.logger.Warn("failed to publish domain.added event", zap.Error(err))
	}

	return d, nil
}

func (s *domainService) GetDomain(ctx context.Context, id uuid.UUID) (*models.Domain, error) {
	return s.domainRepo.GetByID(ctx, id)
}

func (s *domainService) ListDomains(ctx context.Context, workspaceID uuid.UUID) ([]*models.Domain, error) {
	return s.domainRepo.List(ctx, workspaceID)
}

func (s *domainService) VerifyDomain(ctx context.Context, id, workspaceID uuid.UUID) (*models.Domain, error) {
	d, err := s.domainRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if d.WorkspaceID != workspaceID {
		return nil, httputil.Forbidden("domain does not belong to this workspace")
	}

	if d.IsVerified {
		return d, nil
	}

	// Get verification token
	token := d.GetVerificationToken()
	if token == "" {
		return nil, httputil.Wrap(fmt.Errorf("missing verification token"), "domain has no verification token")
	}

	// Lookup DNS TXT record: _linkrift.<domain>
	txtHost := fmt.Sprintf("_linkrift.%s", d.Domain)
	records, err := s.dnsResolver.LookupTXT(ctx, txtHost)
	if err != nil {
		s.logger.Debug("DNS TXT lookup failed", zap.String("host", txtHost), zap.Error(err))
		// Update last check time even on failure
		now := time.Now()
		_, _ = s.domainRepo.Update(ctx, sqlc.UpdateDomainParams{
			ID:             d.ID,
			LastDnsCheckAt: pgtype.Timestamptz{Time: now, Valid: true},
		})
		return nil, httputil.Validation("dns", "DNS TXT record not found. Please add the required TXT record and try again.")
	}

	// Check for matching verification record
	expectedValue := fmt.Sprintf("linkrift-verification=%s", token)
	found := false
	for _, record := range records {
		if strings.TrimSpace(record) == expectedValue {
			found = true
			break
		}
	}

	now := time.Now()
	if !found {
		_, _ = s.domainRepo.Update(ctx, sqlc.UpdateDomainParams{
			ID:             d.ID,
			LastDnsCheckAt: pgtype.Timestamptz{Time: now, Valid: true},
		})
		return nil, httputil.Validation("dns", "DNS TXT record found but does not match. Expected: "+expectedValue)
	}

	// Verification successful - provision SSL
	sslStatus, err := s.sslProvider.ProvisionSSL(ctx, d.Domain)
	if err != nil {
		s.logger.Warn("SSL provisioning failed", zap.String("domain", d.Domain), zap.Error(err))
		sslStatus = models.SSLPending
	}

	d, err = s.domainRepo.Update(ctx, sqlc.UpdateDomainParams{
		ID:             d.ID,
		IsVerified:     pgtype.Bool{Bool: true, Valid: true},
		VerifiedAt:     pgtype.Timestamptz{Time: now, Valid: true},
		SslStatus:      pgtype.Text{String: sslStatus, Valid: true},
		LastDnsCheckAt: pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	// Publish webhook event (best-effort)
	if err := s.events.Publish(ctx, "domain.verified", workspaceID, d); err != nil {
		s.logger.Warn("failed to publish domain.verified event", zap.Error(err))
	}

	return d, nil
}

func (s *domainService) RemoveDomain(ctx context.Context, id, workspaceID uuid.UUID) error {
	d, err := s.domainRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if d.WorkspaceID != workspaceID {
		return httputil.Forbidden("domain does not belong to this workspace")
	}

	// Remove SSL certificate if domain was verified
	if d.IsVerified {
		if err := s.sslProvider.RemoveSSL(ctx, d.Domain); err != nil {
			s.logger.Warn("failed to remove SSL certificate", zap.String("domain", d.Domain), zap.Error(err))
		}
	}

	if err := s.domainRepo.SoftDelete(ctx, id); err != nil {
		return err
	}

	// Publish webhook event (best-effort)
	if err := s.events.Publish(ctx, "domain.removed", d.WorkspaceID, d); err != nil {
		s.logger.Warn("failed to publish domain.removed event", zap.Error(err))
	}

	return nil
}

func (s *domainService) GetDNSRecords(ctx context.Context, id uuid.UUID) (*models.VerificationInstructions, error) {
	d, err := s.domainRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	token := d.GetVerificationToken()

	instructions := &models.VerificationInstructions{
		Records: []models.DNSRecordInstruction{
			{
				Type:  "TXT",
				Host:  fmt.Sprintf("_linkrift.%s", d.Domain),
				Value: fmt.Sprintf("linkrift-verification=%s", token),
			},
			{
				Type:  "CNAME",
				Host:  d.Domain,
				Value: s.cfg.App.RedirectURL,
			},
		},
	}

	return instructions, nil
}

// isValidDomainName validates a domain name format.
func isValidDomainName(domain string) bool {
	if len(domain) == 0 || len(domain) > 253 {
		return false
	}

	// Must contain at least one dot
	if !strings.Contains(domain, ".") {
		return false
	}

	// No leading/trailing dots or hyphens
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}
	if strings.HasPrefix(domain, "-") || strings.HasSuffix(domain, "-") {
		return false
	}

	labels := strings.Split(domain, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return false
		}
		for _, c := range label {
			if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
				return false
			}
		}
	}

	return true
}
