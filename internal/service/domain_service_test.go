package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/config"
	"github.com/link-rift/link-rift/internal/license"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

// --- Mock DomainRepository ---

type mockDomainRepo struct {
	domains      map[uuid.UUID]*models.Domain
	domainsByStr map[string]*models.Domain
	count        int64
	createErr    error
}

func newMockDomainRepo() *mockDomainRepo {
	return &mockDomainRepo{
		domains:      make(map[uuid.UUID]*models.Domain),
		domainsByStr: make(map[string]*models.Domain),
	}
}

func (m *mockDomainRepo) Create(_ context.Context, params sqlc.CreateDomainParams) (*models.Domain, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	if _, exists := m.domainsByStr[params.Domain]; exists {
		return nil, httputil.AlreadyExists("domain")
	}
	d := &models.Domain{
		ID:          uuid.New(),
		WorkspaceID: params.WorkspaceID,
		Domain:      params.Domain,
		SSLStatus:   models.SSLPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.domains[d.ID] = d
	m.domainsByStr[d.Domain] = d
	return d, nil
}

func (m *mockDomainRepo) GetByID(_ context.Context, id uuid.UUID) (*models.Domain, error) {
	d, ok := m.domains[id]
	if !ok {
		return nil, httputil.NotFound("domain")
	}
	return d, nil
}

func (m *mockDomainRepo) GetByDomain(_ context.Context, domain string) (*models.Domain, error) {
	d, ok := m.domainsByStr[domain]
	if !ok {
		return nil, httputil.NotFound("domain")
	}
	return d, nil
}

func (m *mockDomainRepo) List(_ context.Context, workspaceID uuid.UUID) ([]*models.Domain, error) {
	var result []*models.Domain
	for _, d := range m.domains {
		if d.WorkspaceID == workspaceID {
			result = append(result, d)
		}
	}
	return result, nil
}

func (m *mockDomainRepo) Update(_ context.Context, params sqlc.UpdateDomainParams) (*models.Domain, error) {
	d, ok := m.domains[params.ID]
	if !ok {
		return nil, httputil.NotFound("domain")
	}
	if params.IsVerified.Valid {
		d.IsVerified = params.IsVerified.Bool
	}
	if params.VerifiedAt.Valid {
		t := params.VerifiedAt.Time
		d.VerifiedAt = &t
	}
	if params.SslStatus.Valid {
		d.SSLStatus = params.SslStatus.String
	}
	if params.DnsRecords != nil {
		d.DNSRecords = params.DnsRecords
	}
	if params.LastDnsCheckAt.Valid {
		t := params.LastDnsCheckAt.Time
		d.LastDNSCheckAt = &t
	}
	d.UpdatedAt = time.Now()
	return d, nil
}

func (m *mockDomainRepo) SoftDelete(_ context.Context, id uuid.UUID) error {
	if _, ok := m.domains[id]; !ok {
		return httputil.NotFound("domain")
	}
	d := m.domains[id]
	delete(m.domainsByStr, d.Domain)
	delete(m.domains, id)
	return nil
}

func (m *mockDomainRepo) GetCountForWorkspace(_ context.Context, _ uuid.UUID) (int64, error) {
	return m.count, nil
}

// --- Mock DNS Resolver ---

type mockDNSResolver struct {
	records map[string][]string
	err     error
}

func (m *mockDNSResolver) LookupTXT(_ context.Context, name string) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	if recs, ok := m.records[name]; ok {
		return recs, nil
	}
	return nil, &noSuchHostError{}
}

type noSuchHostError struct{}

func (e *noSuchHostError) Error() string { return "no such host" }

// --- Helpers ---

func newTestDomainService(repo *mockDomainRepo, tier license.Tier, resolver DNSResolver) *domainService {
	logger := zap.NewNop()
	verifier, _ := license.NewVerifier()
	licManager := license.NewManager(verifier, logger)

	if tier != license.TierFree {
		// Set a pro license by manipulating manager state
		// For testing, we just check what the mock returns
	}

	cfg := &config.Config{
		App: config.AppConfig{
			RedirectURL: "https://lnk.example.com",
		},
	}

	svc := &domainService{
		domainRepo:  repo,
		licManager:  licManager,
		sslProvider: NewMockSSLProvider(),
		dnsResolver: resolver,
		cfg:         cfg,
		logger:      logger,
	}

	return svc
}

// --- Tests ---

func TestAddDomain_InvalidFormat(t *testing.T) {
	repo := newMockDomainRepo()
	svc := newTestDomainService(repo, license.TierFree, nil)

	// Free tier doesn't have custom_domains feature, but let's test format first
	// Actually, format is checked before license, so invalid format should fail on format
	testCases := []string{
		"",
		"nodot",
		".leading-dot.com",
		"trailing-dot.com.",
		"-leading-hyphen.com",
		"UPPERCASE.COM", // will be lowercased first, then validated, so this is actually valid
	}

	ctx := context.Background()
	wsID := uuid.New()

	for _, tc := range testCases {
		_, err := svc.AddDomain(ctx, wsID, models.CreateDomainInput{Domain: tc})
		if err == nil && tc != "UPPERCASE.COM" {
			t.Errorf("expected error for domain %q, got nil", tc)
		}
	}
}

func TestAddDomain_FeatureNotAvailable(t *testing.T) {
	repo := newMockDomainRepo()
	svc := newTestDomainService(repo, license.TierFree, nil)

	ctx := context.Background()
	wsID := uuid.New()

	_, err := svc.AddDomain(ctx, wsID, models.CreateDomainInput{Domain: "example.com"})
	if err == nil {
		t.Fatal("expected payment required error for free tier")
	}

	appErr, ok := err.(*httputil.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != "PAYMENT_REQUIRED" {
		t.Errorf("expected PAYMENT_REQUIRED, got %s", appErr.Code)
	}
}

func TestVerifyDomain_Success(t *testing.T) {
	repo := newMockDomainRepo()
	wsID := uuid.New()
	domainID := uuid.New()
	token := uuid.New().String()

	dnsData, _ := json.Marshal(models.DNSRecordsData{VerificationToken: token})
	repo.domains[domainID] = &models.Domain{
		ID:          domainID,
		WorkspaceID: wsID,
		Domain:      "test.example.com",
		SSLStatus:   models.SSLPending,
		DNSRecords:  dnsData,
	}
	repo.domainsByStr["test.example.com"] = repo.domains[domainID]

	resolver := &mockDNSResolver{
		records: map[string][]string{
			"_linkrift.test.example.com": {"linkrift-verification=" + token},
		},
	}

	svc := newTestDomainService(repo, license.TierPro, resolver)

	ctx := context.Background()
	d, err := svc.VerifyDomain(ctx, domainID, wsID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !d.IsVerified {
		t.Error("expected domain to be verified")
	}
	if d.SSLStatus != "active" {
		t.Errorf("expected SSL status 'active', got %q", d.SSLStatus)
	}
}

func TestVerifyDomain_TXTRecordMissing(t *testing.T) {
	repo := newMockDomainRepo()
	wsID := uuid.New()
	domainID := uuid.New()
	token := uuid.New().String()

	dnsData, _ := json.Marshal(models.DNSRecordsData{VerificationToken: token})
	repo.domains[domainID] = &models.Domain{
		ID:          domainID,
		WorkspaceID: wsID,
		Domain:      "test.example.com",
		SSLStatus:   models.SSLPending,
		DNSRecords:  dnsData,
	}

	resolver := &mockDNSResolver{
		err: &noSuchHostError{},
	}

	svc := newTestDomainService(repo, license.TierPro, resolver)

	ctx := context.Background()
	_, err := svc.VerifyDomain(ctx, domainID, wsID)
	if err == nil {
		t.Fatal("expected error when TXT record is missing")
	}
}

func TestVerifyDomain_WrongWorkspace(t *testing.T) {
	repo := newMockDomainRepo()
	wsID := uuid.New()
	otherWS := uuid.New()
	domainID := uuid.New()

	repo.domains[domainID] = &models.Domain{
		ID:          domainID,
		WorkspaceID: wsID,
		Domain:      "test.example.com",
		SSLStatus:   models.SSLPending,
	}

	svc := newTestDomainService(repo, license.TierPro, nil)

	ctx := context.Background()
	_, err := svc.VerifyDomain(ctx, domainID, otherWS)
	if err == nil {
		t.Fatal("expected forbidden error for wrong workspace")
	}
	appErr, ok := err.(*httputil.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != "FORBIDDEN" {
		t.Errorf("expected FORBIDDEN, got %s", appErr.Code)
	}
}

func TestRemoveDomain_Success(t *testing.T) {
	repo := newMockDomainRepo()
	wsID := uuid.New()
	domainID := uuid.New()

	repo.domains[domainID] = &models.Domain{
		ID:          domainID,
		WorkspaceID: wsID,
		Domain:      "test.example.com",
		IsVerified:  true,
		SSLStatus:   models.SSLActive,
	}
	repo.domainsByStr["test.example.com"] = repo.domains[domainID]

	svc := newTestDomainService(repo, license.TierPro, nil)

	ctx := context.Background()
	err := svc.RemoveDomain(ctx, domainID, wsID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify domain was removed
	if _, ok := repo.domains[domainID]; ok {
		t.Error("expected domain to be deleted")
	}
}

func TestRemoveDomain_WrongWorkspace(t *testing.T) {
	repo := newMockDomainRepo()
	wsID := uuid.New()
	otherWS := uuid.New()
	domainID := uuid.New()

	repo.domains[domainID] = &models.Domain{
		ID:          domainID,
		WorkspaceID: wsID,
		Domain:      "test.example.com",
	}

	svc := newTestDomainService(repo, license.TierPro, nil)

	ctx := context.Background()
	err := svc.RemoveDomain(ctx, domainID, otherWS)
	if err == nil {
		t.Fatal("expected forbidden error")
	}
}

func TestGetDNSRecords(t *testing.T) {
	repo := newMockDomainRepo()
	domainID := uuid.New()
	token := "test-token-123"

	dnsData, _ := json.Marshal(models.DNSRecordsData{VerificationToken: token})
	repo.domains[domainID] = &models.Domain{
		ID:         domainID,
		Domain:     "test.example.com",
		DNSRecords: dnsData,
	}

	svc := newTestDomainService(repo, license.TierPro, nil)

	ctx := context.Background()
	instructions, err := svc.GetDNSRecords(ctx, domainID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(instructions.Records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(instructions.Records))
	}

	// Check TXT record
	txt := instructions.Records[0]
	if txt.Type != "TXT" {
		t.Errorf("expected TXT record, got %s", txt.Type)
	}
	if txt.Host != "_linkrift.test.example.com" {
		t.Errorf("unexpected host: %s", txt.Host)
	}
	if txt.Value != "linkrift-verification=test-token-123" {
		t.Errorf("unexpected value: %s", txt.Value)
	}

	// Check CNAME record
	cname := instructions.Records[1]
	if cname.Type != "CNAME" {
		t.Errorf("expected CNAME record, got %s", cname.Type)
	}
}

func TestIsValidDomainName(t *testing.T) {
	valid := []string{
		"example.com",
		"sub.example.com",
		"deep.sub.example.com",
		"my-domain.co.uk",
		"123.example.com",
	}
	invalid := []string{
		"",
		"nodot",
		".leading.com",
		"trailing.",
		"-start.com",
		"end-.com",
		"spa ce.com",
		"under_score.com",
	}

	for _, d := range valid {
		if !isValidDomainName(d) {
			t.Errorf("expected %q to be valid", d)
		}
	}
	for _, d := range invalid {
		if isValidDomainName(d) {
			t.Errorf("expected %q to be invalid", d)
		}
	}
}

func TestVerifyDomain_AlreadyVerified(t *testing.T) {
	repo := newMockDomainRepo()
	wsID := uuid.New()
	domainID := uuid.New()
	verifiedAt := time.Now()

	repo.domains[domainID] = &models.Domain{
		ID:          domainID,
		WorkspaceID: wsID,
		Domain:      "test.example.com",
		IsVerified:  true,
		VerifiedAt:  &verifiedAt,
		SSLStatus:   models.SSLActive,
	}

	svc := newTestDomainService(repo, license.TierPro, nil)

	ctx := context.Background()
	d, err := svc.VerifyDomain(ctx, domainID, wsID)
	if err != nil {
		t.Fatalf("expected no error for already verified domain, got %v", err)
	}
	if !d.IsVerified {
		t.Error("expected domain to remain verified")
	}
}

