# Custom Domains

> Last Updated: 2025-01-24

Linkrift allows users to use their own custom domains for shortened URLs, providing white-label branding capabilities with automatic SSL provisioning and health monitoring.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Domain Verification](#domain-verification)
  - [DNS TXT Record Verification](#dns-txt-record-verification)
  - [Verification Flow](#verification-flow)
- [SSL Provisioning](#ssl-provisioning)
  - [Cloudflare Integration](#cloudflare-integration)
  - [Certificate Management](#certificate-management)
- [Domain Health Monitoring](#domain-health-monitoring)
- [API Endpoints](#api-endpoints)
- [React Components](#react-components)

---

## Overview

Custom domains enable users to:

- **Brand their short links** with their own domain (e.g., `link.yourcompany.com`)
- **Automatic SSL** provisioning via Cloudflare
- **Health monitoring** to ensure domains remain properly configured
- **Seamless failover** to default domain if custom domain fails

## Architecture

```
                                    ┌─────────────────────────────────────────┐
                                    │           User's DNS Provider           │
                                    │  (Cloudflare, Route53, GoDaddy, etc.)  │
                                    └──────────────────┬──────────────────────┘
                                                       │
                                                       │ CNAME/A Record
                                                       ▼
                                    ┌─────────────────────────────────────────┐
                                    │           Cloudflare Proxy              │
                                    │        (SSL Termination, CDN)           │
                                    └──────────────────┬──────────────────────┘
                                                       │
                                                       ▼
                                    ┌─────────────────────────────────────────┐
                                    │         Linkrift Load Balancer          │
                                    └──────────────────┬──────────────────────┘
                                                       │
                              ┌────────────────────────┼────────────────────────┐
                              │                        │                        │
                              ▼                        ▼                        ▼
               ┌──────────────────────┐  ┌──────────────────────┐  ┌──────────────────────┐
               │   Redirect Service   │  │   Redirect Service   │  │   Redirect Service   │
               │   (Domain Router)    │  │   (Domain Router)    │  │   (Domain Router)    │
               └──────────────────────┘  └──────────────────────┘  └──────────────────────┘
                              │
                              ▼
               ┌──────────────────────┐
               │   Domain Registry    │
               │   (PostgreSQL)       │
               └──────────────────────┘
```

---

## Domain Verification

### DNS TXT Record Verification

```go
// internal/domains/verification.go
package domains

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"
)

// VerificationStatus represents the domain verification state
type VerificationStatus string

const (
	StatusPending    VerificationStatus = "pending"
	StatusVerifying  VerificationStatus = "verifying"
	StatusVerified   VerificationStatus = "verified"
	StatusFailed     VerificationStatus = "failed"
)

// Domain represents a custom domain
type Domain struct {
	ID                string             `json:"id" db:"id"`
	WorkspaceID       string             `json:"workspace_id" db:"workspace_id"`
	Domain            string             `json:"domain" db:"domain"`
	VerificationToken string             `json:"-" db:"verification_token"`
	VerificationStatus VerificationStatus `json:"verification_status" db:"verification_status"`
	SSLStatus         string             `json:"ssl_status" db:"ssl_status"`
	CloudflareZoneID  string             `json:"-" db:"cloudflare_zone_id"`
	LastCheckedAt     *time.Time         `json:"last_checked_at" db:"last_checked_at"`
	VerifiedAt        *time.Time         `json:"verified_at" db:"verified_at"`
	CreatedAt         time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at" db:"updated_at"`
}

// DomainVerifier handles domain verification
type DomainVerifier struct {
	resolver    *net.Resolver
	expectedIP  string
	expectedCNAME string
}

// NewDomainVerifier creates a new domain verifier
func NewDomainVerifier(expectedIP, expectedCNAME string) *DomainVerifier {
	return &DomainVerifier{
		resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: 5 * time.Second}
				// Use multiple DNS servers for reliability
				return d.DialContext(ctx, network, "8.8.8.8:53")
			},
		},
		expectedIP:    expectedIP,
		expectedCNAME: expectedCNAME,
	}
}

// GenerateVerificationToken creates a unique verification token
func GenerateVerificationToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// VerifyTXTRecord checks if the domain has the correct TXT record
func (dv *DomainVerifier) VerifyTXTRecord(ctx context.Context, domain, expectedToken string) (bool, error) {
	txtRecords, err := dv.resolver.LookupTXT(ctx, fmt.Sprintf("_linkrift.%s", domain))
	if err != nil {
		return false, fmt.Errorf("TXT record lookup failed: %w", err)
	}

	expectedValue := fmt.Sprintf("linkrift-verification=%s", expectedToken)

	for _, record := range txtRecords {
		if strings.TrimSpace(record) == expectedValue {
			return true, nil
		}
	}

	return false, nil
}

// VerifyDNSConfiguration checks if the domain is properly configured
func (dv *DomainVerifier) VerifyDNSConfiguration(ctx context.Context, domain string) (*DNSVerificationResult, error) {
	result := &DNSVerificationResult{
		Domain: domain,
	}

	// Check CNAME record
	cnames, err := dv.resolver.LookupCNAME(ctx, domain)
	if err == nil && cnames != "" {
		result.HasCNAME = true
		result.CNAMETarget = strings.TrimSuffix(cnames, ".")
		result.CNAMEValid = strings.HasSuffix(result.CNAMETarget, dv.expectedCNAME)
	}

	// Check A record
	ips, err := dv.resolver.LookupIP(ctx, "ip4", domain)
	if err == nil && len(ips) > 0 {
		result.HasARecord = true
		for _, ip := range ips {
			result.ARecords = append(result.ARecords, ip.String())
			if ip.String() == dv.expectedIP {
				result.ARecordValid = true
			}
		}
	}

	// Domain is valid if either CNAME or A record is correct
	result.IsValid = result.CNAMEValid || result.ARecordValid

	return result, nil
}

// DNSVerificationResult contains the results of DNS verification
type DNSVerificationResult struct {
	Domain       string   `json:"domain"`
	HasCNAME     bool     `json:"has_cname"`
	CNAMETarget  string   `json:"cname_target,omitempty"`
	CNAMEValid   bool     `json:"cname_valid"`
	HasARecord   bool     `json:"has_a_record"`
	ARecords     []string `json:"a_records,omitempty"`
	ARecordValid bool     `json:"a_record_valid"`
	IsValid      bool     `json:"is_valid"`
}
```

### Verification Flow

```go
// internal/domains/service.go
package domains

import (
	"context"
	"errors"
	"time"

	"github.com/link-rift/link-rift/internal/db"
)

var (
	ErrDomainNotFound     = errors.New("domain not found")
	ErrDomainAlreadyExists = errors.New("domain already exists")
	ErrVerificationFailed = errors.New("domain verification failed")
	ErrDomainNotVerified  = errors.New("domain not verified")
)

// DomainService handles domain operations
type DomainService struct {
	repo       *db.DomainRepository
	verifier   *DomainVerifier
	sslManager *SSLManager
}

// NewDomainService creates a new domain service
func NewDomainService(
	repo *db.DomainRepository,
	verifier *DomainVerifier,
	sslManager *SSLManager,
) *DomainService {
	return &DomainService{
		repo:       repo,
		verifier:   verifier,
		sslManager: sslManager,
	}
}

// AddDomain adds a new custom domain
func (ds *DomainService) AddDomain(ctx context.Context, workspaceID, domainName string) (*Domain, error) {
	// Normalize domain
	domainName = normalizeDomain(domainName)

	// Check if domain already exists
	existing, _ := ds.repo.GetByDomain(ctx, domainName)
	if existing != nil {
		return nil, ErrDomainAlreadyExists
	}

	// Generate verification token
	token, err := GenerateVerificationToken()
	if err != nil {
		return nil, err
	}

	domain := &Domain{
		WorkspaceID:        workspaceID,
		Domain:             domainName,
		VerificationToken:  token,
		VerificationStatus: StatusPending,
		SSLStatus:          "pending",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := ds.repo.Create(ctx, domain); err != nil {
		return nil, err
	}

	return domain, nil
}

// VerifyDomain attempts to verify a domain
func (ds *DomainService) VerifyDomain(ctx context.Context, domainID string) (*Domain, error) {
	domain, err := ds.repo.GetByID(ctx, domainID)
	if err != nil {
		return nil, ErrDomainNotFound
	}

	// Update status to verifying
	domain.VerificationStatus = StatusVerifying
	ds.repo.Update(ctx, domain)

	// Verify TXT record
	verified, err := ds.verifier.VerifyTXTRecord(ctx, domain.Domain, domain.VerificationToken)
	if err != nil || !verified {
		domain.VerificationStatus = StatusFailed
		domain.LastCheckedAt = timePtr(time.Now())
		ds.repo.Update(ctx, domain)
		return domain, ErrVerificationFailed
	}

	// Verify DNS configuration
	dnsResult, err := ds.verifier.VerifyDNSConfiguration(ctx, domain.Domain)
	if err != nil || !dnsResult.IsValid {
		domain.VerificationStatus = StatusFailed
		domain.LastCheckedAt = timePtr(time.Now())
		ds.repo.Update(ctx, domain)
		return domain, ErrVerificationFailed
	}

	// Mark as verified
	now := time.Now()
	domain.VerificationStatus = StatusVerified
	domain.VerifiedAt = &now
	domain.LastCheckedAt = &now

	if err := ds.repo.Update(ctx, domain); err != nil {
		return nil, err
	}

	// Trigger SSL provisioning asynchronously
	go ds.sslManager.ProvisionSSL(context.Background(), domain)

	return domain, nil
}

// GetVerificationInstructions returns DNS setup instructions
func (ds *DomainService) GetVerificationInstructions(domain *Domain) *VerificationInstructions {
	return &VerificationInstructions{
		Domain: domain.Domain,
		TXTRecord: TXTRecordInstruction{
			Host:  fmt.Sprintf("_linkrift.%s", domain.Domain),
			Type:  "TXT",
			Value: fmt.Sprintf("linkrift-verification=%s", domain.VerificationToken),
			TTL:   300,
		},
		CNAMERecord: CNAMERecordInstruction{
			Host:  domain.Domain,
			Type:  "CNAME",
			Value: "links.linkrift.io",
			TTL:   300,
		},
		AlternativeARecord: ARecordInstruction{
			Host:  domain.Domain,
			Type:  "A",
			Value: "YOUR_LINKRIFT_IP", // Replace with actual IP
			TTL:   300,
		},
	}
}

type VerificationInstructions struct {
	Domain             string               `json:"domain"`
	TXTRecord          TXTRecordInstruction `json:"txt_record"`
	CNAMERecord        CNAMERecordInstruction `json:"cname_record"`
	AlternativeARecord ARecordInstruction   `json:"alternative_a_record"`
}

type TXTRecordInstruction struct {
	Host  string `json:"host"`
	Type  string `json:"type"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
}

type CNAMERecordInstruction struct {
	Host  string `json:"host"`
	Type  string `json:"type"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
}

type ARecordInstruction struct {
	Host  string `json:"host"`
	Type  string `json:"type"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
}

func normalizeDomain(domain string) string {
	domain = strings.ToLower(domain)
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimSuffix(domain, "/")
	return domain
}

func timePtr(t time.Time) *time.Time {
	return &t
}
```

---

## SSL Provisioning

### Cloudflare Integration

```go
// internal/domains/ssl.go
package domains

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudflare/cloudflare-go"
)

// SSLStatus represents SSL certificate status
type SSLStatus string

const (
	SSLStatusPending      SSLStatus = "pending"
	SSLStatusProvisioning SSLStatus = "provisioning"
	SSLStatusActive       SSLStatus = "active"
	SSLStatusFailed       SSLStatus = "failed"
	SSLStatusExpiring     SSLStatus = "expiring"
)

// SSLManager handles SSL certificate provisioning
type SSLManager struct {
	cf     *cloudflare.API
	zoneID string
	repo   *db.DomainRepository
}

// NewSSLManager creates a new SSL manager
func NewSSLManager(apiToken, zoneID string, repo *db.DomainRepository) (*SSLManager, error) {
	cf, err := cloudflare.NewWithAPIToken(apiToken)
	if err != nil {
		return nil, err
	}

	return &SSLManager{
		cf:     cf,
		zoneID: zoneID,
		repo:   repo,
	}, nil
}

// ProvisionSSL provisions an SSL certificate for a domain
func (sm *SSLManager) ProvisionSSL(ctx context.Context, domain *Domain) error {
	// Update status to provisioning
	domain.SSLStatus = string(SSLStatusProvisioning)
	sm.repo.Update(ctx, domain)

	// Create custom hostname in Cloudflare
	hostname, err := sm.cf.CreateCustomHostname(ctx, sm.zoneID, cloudflare.CustomHostname{
		Hostname: domain.Domain,
		SSL: &cloudflare.CustomHostnameSSL{
			Method: "http",
			Type:   "dv",
			Settings: cloudflare.CustomHostnameSSLSettings{
				HTTP2:         "on",
				MinTLSVersion: "1.2",
				TLS13:         "on",
				EarlyHints:    "on",
			},
		},
	})
	if err != nil {
		domain.SSLStatus = string(SSLStatusFailed)
		sm.repo.Update(ctx, domain)
		return fmt.Errorf("failed to create custom hostname: %w", err)
	}

	// Store Cloudflare hostname ID for future reference
	domain.CloudflareZoneID = hostname.ID
	sm.repo.Update(ctx, domain)

	// Poll for SSL certificate status
	go sm.pollSSLStatus(ctx, domain)

	return nil
}

func (sm *SSLManager) pollSSLStatus(ctx context.Context, domain *Domain) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	timeout := time.After(30 * time.Minute)

	for {
		select {
		case <-timeout:
			domain.SSLStatus = string(SSLStatusFailed)
			sm.repo.Update(ctx, domain)
			return
		case <-ticker.C:
			hostname, err := sm.cf.CustomHostname(ctx, sm.zoneID, domain.CloudflareZoneID)
			if err != nil {
				continue
			}

			if hostname.SSL != nil {
				switch hostname.SSL.Status {
				case "active":
					domain.SSLStatus = string(SSLStatusActive)
					sm.repo.Update(ctx, domain)
					return
				case "pending_validation", "pending_issuance", "pending_deployment":
					// Still provisioning, continue polling
				default:
					// Failed or unknown status
					domain.SSLStatus = string(SSLStatusFailed)
					sm.repo.Update(ctx, domain)
					return
				}
			}
		}
	}
}

// GetSSLStatus retrieves the current SSL status from Cloudflare
func (sm *SSLManager) GetSSLStatus(ctx context.Context, domain *Domain) (*SSLInfo, error) {
	if domain.CloudflareZoneID == "" {
		return nil, fmt.Errorf("domain not registered with Cloudflare")
	}

	hostname, err := sm.cf.CustomHostname(ctx, sm.zoneID, domain.CloudflareZoneID)
	if err != nil {
		return nil, err
	}

	info := &SSLInfo{
		Status: hostname.SSL.Status,
		Method: hostname.SSL.Method,
	}

	if hostname.SSL.CertificateAuthority != "" {
		info.Issuer = hostname.SSL.CertificateAuthority
	}

	if !hostname.SSL.ExpiresOn.IsZero() {
		info.ExpiresAt = &hostname.SSL.ExpiresOn
	}

	return info, nil
}

// RenewSSL triggers SSL certificate renewal
func (sm *SSLManager) RenewSSL(ctx context.Context, domain *Domain) error {
	_, err := sm.cf.UpdateCustomHostnameSSL(ctx, sm.zoneID, domain.CloudflareZoneID, &cloudflare.CustomHostnameSSL{
		Method: "http",
		Type:   "dv",
	})
	return err
}

// RemoveSSL removes the custom hostname from Cloudflare
func (sm *SSLManager) RemoveSSL(ctx context.Context, domain *Domain) error {
	if domain.CloudflareZoneID == "" {
		return nil
	}

	return sm.cf.DeleteCustomHostname(ctx, sm.zoneID, domain.CloudflareZoneID)
}

// SSLInfo contains SSL certificate information
type SSLInfo struct {
	Status    string     `json:"status"`
	Method    string     `json:"method"`
	Issuer    string     `json:"issuer,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}
```

### Certificate Management

```go
// internal/domains/certificate_monitor.go
package domains

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/link-rift/link-rift/internal/db"
)

// CertificateMonitor monitors SSL certificates for expiration
type CertificateMonitor struct {
	repo       *db.DomainRepository
	sslManager *SSLManager
	cron       *cron.Cron
}

// NewCertificateMonitor creates a new certificate monitor
func NewCertificateMonitor(repo *db.DomainRepository, ssl *SSLManager) *CertificateMonitor {
	return &CertificateMonitor{
		repo:       repo,
		sslManager: ssl,
		cron:       cron.New(),
	}
}

// Start begins the certificate monitoring schedule
func (cm *CertificateMonitor) Start() error {
	// Check certificates daily at 2 AM
	_, err := cm.cron.AddFunc("0 2 * * *", cm.checkCertificates)
	if err != nil {
		return err
	}

	cm.cron.Start()
	return nil
}

func (cm *CertificateMonitor) checkCertificates() {
	ctx := context.Background()

	// Get all verified domains with active SSL
	domains, err := cm.repo.GetDomainsWithActiveSSL(ctx)
	if err != nil {
		return
	}

	for _, domain := range domains {
		sslInfo, err := cm.sslManager.GetSSLStatus(ctx, domain)
		if err != nil {
			continue
		}

		// Check if certificate expires within 30 days
		if sslInfo.ExpiresAt != nil {
			daysUntilExpiry := time.Until(*sslInfo.ExpiresAt).Hours() / 24

			if daysUntilExpiry <= 30 {
				// Mark as expiring
				domain.SSLStatus = string(SSLStatusExpiring)
				cm.repo.Update(ctx, domain)

				// Trigger renewal
				cm.sslManager.RenewSSL(ctx, domain)
			}
		}
	}
}

// Stop shuts down the monitor
func (cm *CertificateMonitor) Stop() {
	cm.cron.Stop()
}
```

---

## Domain Health Monitoring

```go
// internal/domains/health.go
package domains

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/robfig/cron/v3"
)

// HealthStatus represents domain health state
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// DomainHealth contains health check results
type DomainHealth struct {
	DomainID      string       `json:"domain_id"`
	Domain        string       `json:"domain"`
	Status        HealthStatus `json:"status"`
	DNSResolves   bool         `json:"dns_resolves"`
	SSLValid      bool         `json:"ssl_valid"`
	SSLExpiresIn  int          `json:"ssl_expires_in_days,omitempty"`
	ResponseTime  int64        `json:"response_time_ms"`
	HTTPStatus    int          `json:"http_status,omitempty"`
	LastCheckedAt time.Time    `json:"last_checked_at"`
	Errors        []string     `json:"errors,omitempty"`
}

// HealthMonitor performs periodic health checks on domains
type HealthMonitor struct {
	repo       *db.DomainRepository
	httpClient *http.Client
	cron       *cron.Cron
	verifier   *DomainVerifier
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(repo *db.DomainRepository, verifier *DomainVerifier) *HealthMonitor {
	return &HealthMonitor{
		repo: repo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: false,
				},
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		cron:     cron.New(),
		verifier: verifier,
	}
}

// Start begins the health monitoring schedule
func (hm *HealthMonitor) Start() error {
	// Check health every 5 minutes
	_, err := hm.cron.AddFunc("*/5 * * * *", hm.checkAllDomains)
	if err != nil {
		return err
	}

	hm.cron.Start()
	return nil
}

func (hm *HealthMonitor) checkAllDomains() {
	ctx := context.Background()

	domains, err := hm.repo.GetVerifiedDomains(ctx)
	if err != nil {
		return
	}

	for _, domain := range domains {
		health := hm.CheckDomain(ctx, domain)
		hm.repo.UpdateHealth(ctx, domain.ID, health)
	}
}

// CheckDomain performs a comprehensive health check
func (hm *HealthMonitor) CheckDomain(ctx context.Context, domain *Domain) *DomainHealth {
	health := &DomainHealth{
		DomainID:      domain.ID,
		Domain:        domain.Domain,
		LastCheckedAt: time.Now(),
		Errors:        []string{},
	}

	// Check DNS resolution
	dnsResult, err := hm.verifier.VerifyDNSConfiguration(ctx, domain.Domain)
	if err != nil {
		health.Errors = append(health.Errors, fmt.Sprintf("DNS check failed: %v", err))
	} else {
		health.DNSResolves = dnsResult.IsValid
		if !dnsResult.IsValid {
			health.Errors = append(health.Errors, "DNS not pointing to Linkrift")
		}
	}

	// Check SSL certificate
	sslValid, expiresIn, sslErr := hm.checkSSL(domain.Domain)
	health.SSLValid = sslValid
	health.SSLExpiresIn = expiresIn
	if sslErr != nil {
		health.Errors = append(health.Errors, fmt.Sprintf("SSL check failed: %v", sslErr))
	}

	// Check HTTP response
	start := time.Now()
	resp, err := hm.httpClient.Get(fmt.Sprintf("https://%s/health", domain.Domain))
	health.ResponseTime = time.Since(start).Milliseconds()

	if err != nil {
		health.Errors = append(health.Errors, fmt.Sprintf("HTTP check failed: %v", err))
	} else {
		health.HTTPStatus = resp.StatusCode
		resp.Body.Close()
	}

	// Determine overall status
	health.Status = hm.determineStatus(health)

	return health
}

func (hm *HealthMonitor) checkSSL(domain string) (bool, int, error) {
	conn, err := tls.Dial("tcp", domain+":443", &tls.Config{
		ServerName: domain,
	})
	if err != nil {
		return false, 0, err
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return false, 0, fmt.Errorf("no certificates found")
	}

	cert := certs[0]
	expiresIn := int(time.Until(cert.NotAfter).Hours() / 24)

	// Verify the certificate is valid for this domain
	if err := cert.VerifyHostname(domain); err != nil {
		return false, expiresIn, err
	}

	return true, expiresIn, nil
}

func (hm *HealthMonitor) determineStatus(health *DomainHealth) HealthStatus {
	if !health.DNSResolves {
		return HealthStatusUnhealthy
	}

	if !health.SSLValid {
		return HealthStatusUnhealthy
	}

	if health.HTTPStatus >= 500 {
		return HealthStatusUnhealthy
	}

	if health.SSLExpiresIn <= 7 {
		return HealthStatusDegraded
	}

	if health.ResponseTime > 5000 {
		return HealthStatusDegraded
	}

	if len(health.Errors) > 0 {
		return HealthStatusDegraded
	}

	return HealthStatusHealthy
}

// Stop shuts down the monitor
func (hm *HealthMonitor) Stop() {
	hm.cron.Stop()
}
```

---

## API Endpoints

```go
// internal/api/handlers/domains.go
package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/link-rift/link-rift/internal/domains"
)

// DomainHandler handles domain API requests
type DomainHandler struct {
	service *domains.DomainService
}

// NewDomainHandler creates a new domain handler
func NewDomainHandler(service *domains.DomainService) *DomainHandler {
	return &DomainHandler{service: service}
}

// RegisterRoutes registers domain routes
func (h *DomainHandler) RegisterRoutes(app *fiber.App) {
	domains := app.Group("/api/v1/domains")

	domains.Get("/", h.ListDomains)
	domains.Post("/", h.AddDomain)
	domains.Get("/:id", h.GetDomain)
	domains.Delete("/:id", h.DeleteDomain)
	domains.Post("/:id/verify", h.VerifyDomain)
	domains.Get("/:id/health", h.GetDomainHealth)
	domains.Get("/:id/instructions", h.GetVerificationInstructions)
}

// AddDomain handles domain creation
// @Summary Add a custom domain
// @Tags Domains
// @Accept json
// @Produce json
// @Param body body AddDomainRequest true "Domain details"
// @Success 201 {object} DomainResponse
// @Router /api/v1/domains [post]
func (h *DomainHandler) AddDomain(c *fiber.Ctx) error {
	var req AddDomainRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	workspaceID := c.Locals("workspaceID").(string)

	domain, err := h.service.AddDomain(c.Context(), workspaceID, req.Domain)
	if err != nil {
		if err == domains.ErrDomainAlreadyExists {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Domain already exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add domain",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(toDomainResponse(domain))
}

// VerifyDomain handles domain verification
// @Summary Verify domain ownership
// @Tags Domains
// @Produce json
// @Param id path string true "Domain ID"
// @Success 200 {object} DomainResponse
// @Router /api/v1/domains/{id}/verify [post]
func (h *DomainHandler) VerifyDomain(c *fiber.Ctx) error {
	domainID := c.Params("id")

	domain, err := h.service.VerifyDomain(c.Context(), domainID)
	if err != nil {
		if err == domains.ErrDomainNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Domain not found",
			})
		}
		if err == domains.ErrVerificationFailed {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":             "Verification failed",
				"verification_status": domain.VerificationStatus,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Verification failed",
		})
	}

	return c.JSON(toDomainResponse(domain))
}

// GetVerificationInstructions returns DNS setup instructions
// @Summary Get DNS setup instructions
// @Tags Domains
// @Produce json
// @Param id path string true "Domain ID"
// @Success 200 {object} VerificationInstructions
// @Router /api/v1/domains/{id}/instructions [get]
func (h *DomainHandler) GetVerificationInstructions(c *fiber.Ctx) error {
	domainID := c.Params("id")

	domain, err := h.service.GetDomain(c.Context(), domainID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Domain not found",
		})
	}

	instructions := h.service.GetVerificationInstructions(domain)
	return c.JSON(instructions)
}

// GetDomainHealth returns domain health status
// @Summary Get domain health status
// @Tags Domains
// @Produce json
// @Param id path string true "Domain ID"
// @Success 200 {object} DomainHealth
// @Router /api/v1/domains/{id}/health [get]
func (h *DomainHandler) GetDomainHealth(c *fiber.Ctx) error {
	domainID := c.Params("id")

	health, err := h.service.GetDomainHealth(c.Context(), domainID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Domain not found",
		})
	}

	return c.JSON(health)
}

// Request/Response types
type AddDomainRequest struct {
	Domain string `json:"domain" validate:"required,fqdn"`
}

type DomainResponse struct {
	ID                 string `json:"id"`
	Domain             string `json:"domain"`
	VerificationStatus string `json:"verification_status"`
	SSLStatus          string `json:"ssl_status"`
	CreatedAt          string `json:"created_at"`
}

func toDomainResponse(d *domains.Domain) *DomainResponse {
	return &DomainResponse{
		ID:                 d.ID,
		Domain:             d.Domain,
		VerificationStatus: string(d.VerificationStatus),
		SSLStatus:          d.SSLStatus,
		CreatedAt:          d.CreatedAt.Format(time.RFC3339),
	}
}
```

---

## React Components

### Domain Management Component

```typescript
// src/components/domains/DomainManager.tsx
import React, { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { domainsApi, Domain, VerificationInstructions } from '@/api/domains';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { AlertCircle, CheckCircle, Clock, Shield } from 'lucide-react';

export const DomainManager: React.FC = () => {
  const [newDomain, setNewDomain] = useState('');
  const queryClient = useQueryClient();

  const { data: domains, isLoading } = useQuery({
    queryKey: ['domains'],
    queryFn: domainsApi.list,
  });

  const addDomainMutation = useMutation({
    mutationFn: domainsApi.add,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['domains'] });
      setNewDomain('');
    },
  });

  const handleAddDomain = (e: React.FormEvent) => {
    e.preventDefault();
    if (newDomain) {
      addDomainMutation.mutate({ domain: newDomain });
    }
  };

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Add Custom Domain</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleAddDomain} className="flex gap-4">
            <Input
              placeholder="links.yourdomain.com"
              value={newDomain}
              onChange={(e) => setNewDomain(e.target.value)}
              className="flex-1"
            />
            <Button type="submit" disabled={addDomainMutation.isPending}>
              {addDomainMutation.isPending ? 'Adding...' : 'Add Domain'}
            </Button>
          </form>
        </CardContent>
      </Card>

      <div className="space-y-4">
        {domains?.map((domain) => (
          <DomainCard key={domain.id} domain={domain} />
        ))}
      </div>
    </div>
  );
};

interface DomainCardProps {
  domain: Domain;
}

const DomainCard: React.FC<DomainCardProps> = ({ domain }) => {
  const [showInstructions, setShowInstructions] = useState(false);
  const queryClient = useQueryClient();

  const { data: instructions } = useQuery({
    queryKey: ['domain-instructions', domain.id],
    queryFn: () => domainsApi.getInstructions(domain.id),
    enabled: showInstructions && domain.verification_status !== 'verified',
  });

  const verifyMutation = useMutation({
    mutationFn: () => domainsApi.verify(domain.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['domains'] });
    },
  });

  const getStatusBadge = () => {
    switch (domain.verification_status) {
      case 'verified':
        return <Badge variant="success"><CheckCircle className="w-3 h-3 mr-1" /> Verified</Badge>;
      case 'pending':
        return <Badge variant="warning"><Clock className="w-3 h-3 mr-1" /> Pending</Badge>;
      case 'failed':
        return <Badge variant="destructive"><AlertCircle className="w-3 h-3 mr-1" /> Failed</Badge>;
      default:
        return <Badge variant="secondary">{domain.verification_status}</Badge>;
    }
  };

  const getSSLBadge = () => {
    switch (domain.ssl_status) {
      case 'active':
        return <Badge variant="success"><Shield className="w-3 h-3 mr-1" /> SSL Active</Badge>;
      case 'provisioning':
        return <Badge variant="warning"><Clock className="w-3 h-3 mr-1" /> SSL Provisioning</Badge>;
      case 'pending':
        return <Badge variant="secondary">SSL Pending</Badge>;
      default:
        return <Badge variant="destructive">SSL {domain.ssl_status}</Badge>;
    }
  };

  return (
    <Card>
      <CardContent className="pt-6">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-semibold">{domain.domain}</h3>
            <div className="flex gap-2 mt-2">
              {getStatusBadge()}
              {domain.verification_status === 'verified' && getSSLBadge()}
            </div>
          </div>
          <div className="flex gap-2">
            {domain.verification_status !== 'verified' && (
              <>
                <Button
                  variant="outline"
                  onClick={() => setShowInstructions(!showInstructions)}
                >
                  {showInstructions ? 'Hide' : 'Show'} Instructions
                </Button>
                <Button
                  onClick={() => verifyMutation.mutate()}
                  disabled={verifyMutation.isPending}
                >
                  {verifyMutation.isPending ? 'Verifying...' : 'Verify'}
                </Button>
              </>
            )}
          </div>
        </div>

        {showInstructions && instructions && (
          <DNSInstructions instructions={instructions} />
        )}
      </CardContent>
    </Card>
  );
};

interface DNSInstructionsProps {
  instructions: VerificationInstructions;
}

const DNSInstructions: React.FC<DNSInstructionsProps> = ({ instructions }) => {
  return (
    <div className="mt-6 space-y-4 p-4 bg-muted rounded-lg">
      <h4 className="font-semibold">DNS Configuration Instructions</h4>

      <div className="space-y-3">
        <div>
          <p className="text-sm font-medium text-muted-foreground">
            Step 1: Add TXT Record for Verification
          </p>
          <div className="mt-1 p-3 bg-background rounded border font-mono text-sm">
            <div>Host: <code>{instructions.txt_record.host}</code></div>
            <div>Type: <code>{instructions.txt_record.type}</code></div>
            <div>Value: <code className="break-all">{instructions.txt_record.value}</code></div>
            <div>TTL: <code>{instructions.txt_record.ttl}</code></div>
          </div>
        </div>

        <div>
          <p className="text-sm font-medium text-muted-foreground">
            Step 2: Add CNAME Record (Recommended)
          </p>
          <div className="mt-1 p-3 bg-background rounded border font-mono text-sm">
            <div>Host: <code>{instructions.cname_record.host}</code></div>
            <div>Type: <code>{instructions.cname_record.type}</code></div>
            <div>Value: <code>{instructions.cname_record.value}</code></div>
            <div>TTL: <code>{instructions.cname_record.ttl}</code></div>
          </div>
        </div>

        <div>
          <p className="text-sm font-medium text-muted-foreground">
            Alternative: Add A Record (if CNAME not supported)
          </p>
          <div className="mt-1 p-3 bg-background rounded border font-mono text-sm">
            <div>Host: <code>{instructions.alternative_a_record.host}</code></div>
            <div>Type: <code>{instructions.alternative_a_record.type}</code></div>
            <div>Value: <code>{instructions.alternative_a_record.value}</code></div>
            <div>TTL: <code>{instructions.alternative_a_record.ttl}</code></div>
          </div>
        </div>
      </div>

      <p className="text-sm text-muted-foreground">
        DNS changes can take up to 48 hours to propagate. Click "Verify" once you have added the records.
      </p>
    </div>
  );
};
```

### Domain API Client

```typescript
// src/api/domains.ts
import { apiClient } from './client';

export interface Domain {
  id: string;
  domain: string;
  verification_status: 'pending' | 'verifying' | 'verified' | 'failed';
  ssl_status: 'pending' | 'provisioning' | 'active' | 'failed' | 'expiring';
  created_at: string;
}

export interface VerificationInstructions {
  domain: string;
  txt_record: {
    host: string;
    type: string;
    value: string;
    ttl: number;
  };
  cname_record: {
    host: string;
    type: string;
    value: string;
    ttl: number;
  };
  alternative_a_record: {
    host: string;
    type: string;
    value: string;
    ttl: number;
  };
}

export interface DomainHealth {
  domain_id: string;
  domain: string;
  status: 'healthy' | 'degraded' | 'unhealthy';
  dns_resolves: boolean;
  ssl_valid: boolean;
  ssl_expires_in_days?: number;
  response_time_ms: number;
  http_status?: number;
  last_checked_at: string;
  errors?: string[];
}

export const domainsApi = {
  list: async (): Promise<Domain[]> => {
    const response = await apiClient.get<Domain[]>('/api/v1/domains');
    return response.data;
  },

  add: async (data: { domain: string }): Promise<Domain> => {
    const response = await apiClient.post<Domain>('/api/v1/domains', data);
    return response.data;
  },

  get: async (id: string): Promise<Domain> => {
    const response = await apiClient.get<Domain>(`/api/v1/domains/${id}`);
    return response.data;
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/api/v1/domains/${id}`);
  },

  verify: async (id: string): Promise<Domain> => {
    const response = await apiClient.post<Domain>(`/api/v1/domains/${id}/verify`);
    return response.data;
  },

  getInstructions: async (id: string): Promise<VerificationInstructions> => {
    const response = await apiClient.get<VerificationInstructions>(
      `/api/v1/domains/${id}/instructions`
    );
    return response.data;
  },

  getHealth: async (id: string): Promise<DomainHealth> => {
    const response = await apiClient.get<DomainHealth>(`/api/v1/domains/${id}/health`);
    return response.data;
  },
};
```
