# Legal and Compliance Documentation

**Last Updated: 2025-01-24**

This document covers Linkrift's compliance with data protection regulations including GDPR, CCPA, and related legal requirements.

---

## Table of Contents

- [Overview](#overview)
- [GDPR Compliance](#gdpr-compliance)
  - [GDPR Compliance Checklist](#gdpr-compliance-checklist)
  - [Legal Basis for Processing](#legal-basis-for-processing)
  - [Data Subject Rights Implementation](#data-subject-rights-implementation)
  - [Data Protection Impact Assessment](#data-protection-impact-assessment)
- [CCPA Compliance](#ccpa-compliance)
  - [CCPA Compliance Checklist](#ccpa-compliance-checklist)
  - [Consumer Rights Implementation](#consumer-rights-implementation)
  - [Do Not Sell Implementation](#do-not-sell-implementation)
- [Data Retention Policies](#data-retention-policies)
  - [Retention Schedule](#retention-schedule)
  - [Automated Data Cleanup](#automated-data-cleanup)
  - [Retention Policy Implementation](#retention-policy-implementation)
- [User Data Export](#user-data-export)
  - [Export Format](#export-format)
  - [Export API Implementation](#export-api-implementation)
  - [Automated Export Generation](#automated-export-generation)
- [Account Deletion](#account-deletion)
  - [Deletion Process](#deletion-process)
  - [Data Anonymization](#data-anonymization)
  - [Deletion Implementation](#deletion-implementation)
- [Cookie Policy](#cookie-policy)
  - [Cookie Categories](#cookie-categories)
  - [Cookie Consent Implementation](#cookie-consent-implementation)
  - [Cookie Management](#cookie-management)
- [Terms of Service Considerations](#terms-of-service-considerations)
  - [Key Terms](#key-terms)
  - [Acceptable Use Policy](#acceptable-use-policy)
  - [Service Level Agreement](#service-level-agreement)
- [Compliance Monitoring](#compliance-monitoring)
- [Audit Trail](#audit-trail)

---

## Overview

Linkrift is committed to protecting user privacy and complying with applicable data protection regulations:

| Regulation | Applicability | Key Requirements |
|------------|---------------|------------------|
| GDPR | EU/EEA users | Consent, data rights, DPO |
| CCPA | California residents | Disclosure, opt-out, deletion |
| LGPD | Brazilian users | Similar to GDPR |
| PIPEDA | Canadian users | Consent, access, correction |

---

## GDPR Compliance

### GDPR Compliance Checklist

#### Lawfulness, Fairness, and Transparency

- [x] Privacy policy clearly explains data processing
- [x] Legal basis documented for each processing activity
- [x] Consent obtained where required and is freely given
- [x] Users informed about their rights
- [x] Contact information for DPO provided

#### Purpose Limitation

- [x] Data collected for specified, explicit purposes
- [x] Further processing compatible with original purposes
- [x] Purpose documented in privacy policy

#### Data Minimization

- [x] Only necessary data collected
- [x] Regular review of data collected
- [x] Unnecessary fields removed from forms

#### Accuracy

- [x] Users can update their information
- [x] Processes in place to keep data accurate
- [x] Inaccurate data corrected or deleted promptly

#### Storage Limitation

- [x] Retention periods defined for all data types
- [x] Automated deletion of expired data
- [x] Annual review of retention policies

#### Integrity and Confidentiality

- [x] Data encrypted at rest and in transit
- [x] Access controls implemented
- [x] Security measures documented
- [x] Regular security assessments

#### Accountability

- [x] Data processing activities documented
- [x] DPO appointed (if required)
- [x] Staff trained on data protection
- [x] Contracts with processors include required clauses

### Legal Basis for Processing

```go
// internal/compliance/legal_basis.go
package compliance

type LegalBasis string

const (
    LegalBasisConsent            LegalBasis = "consent"
    LegalBasisContract           LegalBasis = "contract"
    LegalBasisLegalObligation    LegalBasis = "legal_obligation"
    LegalBasisVitalInterests     LegalBasis = "vital_interests"
    LegalBasisPublicTask         LegalBasis = "public_task"
    LegalBasisLegitimateInterest LegalBasis = "legitimate_interest"
)

// ProcessingActivity documents a data processing activity
type ProcessingActivity struct {
    Name           string
    Purpose        string
    LegalBasis     LegalBasis
    DataCategories []string
    DataSubjects   []string
    Recipients     []string
    Retention      string
    SecurityMeasures []string
}

// GetProcessingActivities returns all documented processing activities
func GetProcessingActivities() []ProcessingActivity {
    return []ProcessingActivity{
        {
            Name:           "User Account Management",
            Purpose:        "Provide URL shortening service to registered users",
            LegalBasis:     LegalBasisContract,
            DataCategories: []string{"Email", "Name", "Password hash"},
            DataSubjects:   []string{"Registered users"},
            Recipients:     []string{"Internal systems"},
            Retention:      "Account lifetime + 30 days",
            SecurityMeasures: []string{
                "Encryption at rest",
                "Access controls",
                "Audit logging",
            },
        },
        {
            Name:           "Link Analytics",
            Purpose:        "Provide click statistics to link owners",
            LegalBasis:     LegalBasisLegitimateInterest,
            DataCategories: []string{"IP address (anonymized)", "Country", "Device type", "Referrer"},
            DataSubjects:   []string{"Link visitors"},
            Recipients:     []string{"Link owners (aggregated only)"},
            Retention:      "2 years",
            SecurityMeasures: []string{
                "IP anonymization",
                "Aggregated reporting",
                "Access controls",
            },
        },
        {
            Name:           "Marketing Communications",
            Purpose:        "Send product updates and promotional content",
            LegalBasis:     LegalBasisConsent,
            DataCategories: []string{"Email", "Name", "Preferences"},
            DataSubjects:   []string{"Users who opted in"},
            Recipients:     []string{"Email service provider"},
            Retention:      "Until consent withdrawn",
            SecurityMeasures: []string{
                "Consent records",
                "Unsubscribe functionality",
                "Preference center",
            },
        },
    }
}
```

### Data Subject Rights Implementation

```go
// internal/compliance/rights.go
package compliance

import (
    "context"
    "time"
)

type DataSubjectRight string

const (
    RightAccess       DataSubjectRight = "access"
    RightRectification DataSubjectRight = "rectification"
    RightErasure      DataSubjectRight = "erasure"
    RightRestriction  DataSubjectRight = "restriction"
    RightPortability  DataSubjectRight = "portability"
    RightObjection    DataSubjectRight = "objection"
)

// DataSubjectRequest represents a GDPR rights request
type DataSubjectRequest struct {
    ID          string
    UserID      string
    Email       string
    Right       DataSubjectRight
    Status      string
    RequestedAt time.Time
    CompletedAt *time.Time
    Notes       string
}

type RightsService struct {
    repo       RightsRepository
    userRepo   UserRepository
    exportSvc  *ExportService
    deleteSvc  *DeletionService
    auditLog   AuditLogger
}

// ProcessRequest handles a data subject rights request
func (s *RightsService) ProcessRequest(ctx context.Context, req *DataSubjectRequest) error {
    // Log the request
    s.auditLog.LogDataSubjectRequest(ctx, req)

    // Verify identity (important step before processing)
    if err := s.verifyIdentity(ctx, req); err != nil {
        return fmt.Errorf("identity verification failed: %w", err)
    }

    switch req.Right {
    case RightAccess:
        return s.handleAccessRequest(ctx, req)
    case RightRectification:
        return s.handleRectificationRequest(ctx, req)
    case RightErasure:
        return s.handleErasureRequest(ctx, req)
    case RightPortability:
        return s.handlePortabilityRequest(ctx, req)
    case RightObjection:
        return s.handleObjectionRequest(ctx, req)
    default:
        return fmt.Errorf("unknown right: %s", req.Right)
    }
}

func (s *RightsService) handleAccessRequest(ctx context.Context, req *DataSubjectRequest) error {
    // Generate data export
    export, err := s.exportSvc.GenerateExport(ctx, req.UserID)
    if err != nil {
        return err
    }

    // Notify user that export is ready
    return s.notifyUser(ctx, req.UserID, "data_access_ready", map[string]string{
        "download_link": export.DownloadURL,
        "expires_at":    export.ExpiresAt.Format(time.RFC3339),
    })
}

func (s *RightsService) handleErasureRequest(ctx context.Context, req *DataSubjectRequest) error {
    // Check if erasure can be fulfilled
    canDelete, reason := s.canFulfillErasure(ctx, req.UserID)
    if !canDelete {
        req.Notes = "Erasure denied: " + reason
        return s.repo.Update(ctx, req)
    }

    // Process deletion
    return s.deleteSvc.DeleteUserData(ctx, req.UserID)
}

// canFulfillErasure checks if data can be erased
func (s *RightsService) canFulfillErasure(ctx context.Context, userID string) (bool, string) {
    // Check for legal holds
    if s.hasLegalHold(ctx, userID) {
        return false, "Data subject to legal hold"
    }

    // Check for ongoing legal obligations
    if s.hasLegalObligation(ctx, userID) {
        return false, "Data required for legal compliance"
    }

    // Check for outstanding financial obligations
    if s.hasOutstandingPayment(ctx, userID) {
        return false, "Outstanding financial obligations"
    }

    return true, ""
}
```

### Data Protection Impact Assessment

```go
// internal/compliance/dpia.go
package compliance

// DPIA documents a Data Protection Impact Assessment
type DPIA struct {
    ID                 string
    ProjectName        string
    Description        string
    DataFlows          []DataFlow
    Risks              []Risk
    Mitigations        []Mitigation
    DPOApproval        bool
    DPOApprovalDate    *time.Time
    ReviewDate         time.Time
}

type DataFlow struct {
    Source      string
    Destination string
    DataTypes   []string
    Purpose     string
    LegalBasis  LegalBasis
}

type Risk struct {
    Description string
    Likelihood  string // low, medium, high
    Impact      string // low, medium, high
    Category    string // confidentiality, integrity, availability
}

type Mitigation struct {
    RiskID      string
    Description string
    Status      string
    Owner       string
}

// Example DPIA for click analytics
var ClickAnalyticsDPIA = DPIA{
    ProjectName: "Click Analytics System",
    Description: "System to track and analyze link clicks for user analytics",
    DataFlows: []DataFlow{
        {
            Source:      "Website visitor",
            Destination: "ClickHouse analytics database",
            DataTypes:   []string{"IP address", "User agent", "Timestamp", "Referrer"},
            Purpose:     "Provide click statistics to link owners",
            LegalBasis:  LegalBasisLegitimateInterest,
        },
    },
    Risks: []Risk{
        {
            Description: "Re-identification of visitors through IP correlation",
            Likelihood:  "medium",
            Impact:      "high",
            Category:    "confidentiality",
        },
        {
            Description: "Unauthorized access to analytics data",
            Likelihood:  "low",
            Impact:      "medium",
            Category:    "confidentiality",
        },
    },
    Mitigations: []Mitigation{
        {
            Description: "Anonymize IP addresses by removing last octet",
            Status:      "implemented",
            Owner:       "Engineering",
        },
        {
            Description: "Implement role-based access control",
            Status:      "implemented",
            Owner:       "Security",
        },
        {
            Description: "Encrypt analytics data at rest",
            Status:      "implemented",
            Owner:       "Infrastructure",
        },
    },
}
```

---

## CCPA Compliance

### CCPA Compliance Checklist

#### Notice Requirements

- [x] Privacy policy discloses categories of personal information collected
- [x] Privacy policy discloses purposes for collection
- [x] Privacy policy discloses consumer rights
- [x] "Do Not Sell My Personal Information" link on homepage
- [x] Notice at collection provided

#### Consumer Rights

- [x] Right to know what personal information is collected
- [x] Right to know what personal information is sold/disclosed
- [x] Right to delete personal information
- [x] Right to opt-out of sale of personal information
- [x] Right to non-discrimination

#### Operational Requirements

- [x] Methods for submitting requests (web form, email, toll-free number)
- [x] Identity verification process
- [x] Response within 45 days (extendable to 90 days)
- [x] Free response for requests twice per year
- [x] Staff training on CCPA requirements

### Consumer Rights Implementation

```go
// internal/compliance/ccpa.go
package compliance

import (
    "context"
    "time"
)

type CCPACategory string

const (
    CCPAIdentifiers          CCPACategory = "identifiers"
    CCPACommercialInfo       CCPACategory = "commercial_info"
    CCPAInternetActivity     CCPACategory = "internet_activity"
    CCPAGeolocation          CCPACategory = "geolocation"
    CCPAProfessionalInfo     CCPACategory = "professional_info"
    CCPAInferences           CCPACategory = "inferences"
)

// CCPADisclosure provides information about collected data categories
type CCPADisclosure struct {
    Category    CCPACategory
    Examples    []string
    Purpose     string
    Sold        bool
    Disclosed   bool
    Sources     []string
    Recipients  []string
}

// GetCCPADisclosures returns required CCPA disclosures
func GetCCPADisclosures() []CCPADisclosure {
    return []CCPADisclosure{
        {
            Category: CCPAIdentifiers,
            Examples: []string{
                "Email address",
                "Name",
                "Account ID",
                "IP address",
            },
            Purpose:   "Account management and service delivery",
            Sold:      false,
            Disclosed: true,
            Sources:   []string{"Directly from consumer", "Automatically collected"},
            Recipients: []string{
                "Service providers (hosting, email)",
            },
        },
        {
            Category: CCPAInternetActivity,
            Examples: []string{
                "Links created",
                "Click data",
                "Browser type",
                "Referring URLs",
            },
            Purpose:   "Service delivery and analytics",
            Sold:      false,
            Disclosed: true,
            Sources:   []string{"Automatically collected"},
            Recipients: []string{
                "Analytics providers",
                "Service providers",
            },
        },
    }
}

// CCPARequest represents a consumer rights request
type CCPARequest struct {
    ID           string
    ConsumerID   string
    Email        string
    RequestType  string // know, delete, optout, access
    Status       string
    SubmittedAt  time.Time
    VerifiedAt   *time.Time
    CompletedAt  *time.Time
    ResponseDue  time.Time // 45 days from submission
    ExtendedDue  *time.Time // Up to 90 days if extended
}

type CCPAService struct {
    repo     CCPARepository
    userSvc  UserService
    auditLog AuditLogger
}

// SubmitRequest creates a new CCPA request
func (s *CCPAService) SubmitRequest(ctx context.Context, req *CCPARequest) error {
    req.ID = generateID()
    req.SubmittedAt = time.Now()
    req.ResponseDue = req.SubmittedAt.Add(45 * 24 * time.Hour)
    req.Status = "pending_verification"

    if err := s.repo.Create(ctx, req); err != nil {
        return err
    }

    // Send verification email
    return s.sendVerificationEmail(ctx, req)
}

// ProcessRequest handles a verified CCPA request
func (s *CCPAService) ProcessRequest(ctx context.Context, requestID string) error {
    req, err := s.repo.Get(ctx, requestID)
    if err != nil {
        return err
    }

    if req.VerifiedAt == nil {
        return fmt.Errorf("request not verified")
    }

    switch req.RequestType {
    case "know":
        return s.handleKnowRequest(ctx, req)
    case "access":
        return s.handleAccessRequest(ctx, req)
    case "delete":
        return s.handleDeleteRequest(ctx, req)
    case "optout":
        return s.handleOptOutRequest(ctx, req)
    default:
        return fmt.Errorf("unknown request type: %s", req.RequestType)
    }
}

func (s *CCPAService) handleKnowRequest(ctx context.Context, req *CCPARequest) error {
    // Generate disclosure report
    disclosures := GetCCPADisclosures()

    // Get user-specific data collected
    userData, err := s.userSvc.GetUserData(ctx, req.ConsumerID)
    if err != nil {
        return err
    }

    // Create response document
    response := CCPAResponse{
        RequestID:    req.ID,
        Categories:   disclosures,
        SpecificData: userData,
        GeneratedAt:  time.Now(),
    }

    // Send to consumer
    return s.sendResponse(ctx, req, response)
}
```

### Do Not Sell Implementation

```go
// internal/compliance/donotsell.go
package compliance

import (
    "context"
    "time"
)

type DoNotSellService struct {
    repo     DoNotSellRepository
    auditLog AuditLogger
}

// OptOut records a do-not-sell preference
func (s *DoNotSellService) OptOut(ctx context.Context, consumerID string) error {
    preference := &DoNotSellPreference{
        ConsumerID: consumerID,
        OptedOut:   true,
        OptOutDate: time.Now(),
    }

    if err := s.repo.Save(ctx, preference); err != nil {
        return err
    }

    s.auditLog.Log(ctx, AuditEvent{
        EventType: "ccpa_optout",
        UserID:    consumerID,
        Action:    "do_not_sell",
        Details: map[string]interface{}{
            "opted_out": true,
        },
    })

    return nil
}

// CheckOptOut checks if a consumer has opted out
func (s *DoNotSellService) CheckOptOut(ctx context.Context, consumerID string) (bool, error) {
    pref, err := s.repo.Get(ctx, consumerID)
    if err != nil {
        return false, nil // Default to not opted out
    }
    return pref.OptedOut, nil
}

// Global Privacy Control support
func (s *DoNotSellService) ProcessGPC(ctx context.Context, consumerID string, gpcSignal bool) error {
    if gpcSignal {
        // GPC signal is treated as a valid opt-out request
        return s.OptOut(ctx, consumerID)
    }
    return nil
}
```

---

## Data Retention Policies

### Retention Schedule

| Data Category | Retention Period | Legal Basis | Deletion Method |
|---------------|------------------|-------------|-----------------|
| User accounts | Account lifetime + 30 days | Contract | Hard delete |
| Link data | Account lifetime + 30 days | Contract | Hard delete |
| Click analytics | 2 years | Legitimate interest | Automated purge |
| Access logs | 90 days | Security | Log rotation |
| Audit logs | 7 years | Legal obligation | Archive then delete |
| Support tickets | 3 years after resolution | Contract | Archive then delete |
| Marketing consent | Until withdrawn + 30 days | Consent | Soft delete |
| Payment records | 7 years | Legal obligation | Archive |

### Automated Data Cleanup

```go
// internal/compliance/retention.go
package compliance

import (
    "context"
    "time"

    "go.uber.org/zap"
)

type RetentionPolicy struct {
    DataType   string
    Retention  time.Duration
    DeleteFunc func(ctx context.Context, before time.Time) (int64, error)
}

type RetentionService struct {
    policies []RetentionPolicy
    logger   *zap.Logger
    auditLog AuditLogger
}

func NewRetentionService(logger *zap.Logger, repos *Repositories) *RetentionService {
    return &RetentionService{
        logger: logger,
        policies: []RetentionPolicy{
            {
                DataType:   "click_analytics",
                Retention:  2 * 365 * 24 * time.Hour, // 2 years
                DeleteFunc: repos.Analytics.DeleteBefore,
            },
            {
                DataType:   "access_logs",
                Retention:  90 * 24 * time.Hour, // 90 days
                DeleteFunc: repos.Logs.DeleteBefore,
            },
            {
                DataType:   "expired_links",
                Retention:  30 * 24 * time.Hour, // 30 days after expiry
                DeleteFunc: repos.Links.DeleteExpiredBefore,
            },
            {
                DataType:   "deleted_accounts",
                Retention:  30 * 24 * time.Hour, // 30 days after deletion
                DeleteFunc: repos.Users.PurgeDeletedBefore,
            },
            {
                DataType:   "password_reset_tokens",
                Retention:  24 * time.Hour,
                DeleteFunc: repos.Tokens.DeleteExpiredBefore,
            },
        },
    }
}

// RunCleanup executes retention policies
func (s *RetentionService) RunCleanup(ctx context.Context) error {
    s.logger.Info("Starting retention cleanup")

    for _, policy := range s.policies {
        cutoff := time.Now().Add(-policy.Retention)

        deleted, err := policy.DeleteFunc(ctx, cutoff)
        if err != nil {
            s.logger.Error("Retention cleanup failed",
                zap.String("data_type", policy.DataType),
                zap.Error(err),
            )
            continue
        }

        s.logger.Info("Retention cleanup completed",
            zap.String("data_type", policy.DataType),
            zap.Int64("records_deleted", deleted),
            zap.Time("cutoff", cutoff),
        )

        s.auditLog.Log(ctx, AuditEvent{
            EventType: "retention_cleanup",
            Action:    "delete",
            Resource:  policy.DataType,
            Details: map[string]interface{}{
                "records_deleted": deleted,
                "cutoff_date":     cutoff,
            },
        })
    }

    return nil
}
```

### Retention Policy Implementation

```sql
-- db/migrations/003_retention_policies.sql

-- Add soft delete columns
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE links ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;

-- Create indexes for efficient cleanup queries
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_links_deleted_at ON links(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_clicks_created_at ON clicks(created_at);

-- Function to purge old data
CREATE OR REPLACE FUNCTION purge_old_data()
RETURNS void AS $$
BEGIN
    -- Delete clicks older than 2 years
    DELETE FROM clicks WHERE created_at < NOW() - INTERVAL '2 years';

    -- Hard delete accounts that were soft-deleted more than 30 days ago
    DELETE FROM users WHERE deleted_at < NOW() - INTERVAL '30 days';

    -- Hard delete links for deleted accounts
    DELETE FROM links
    WHERE user_id NOT IN (SELECT id FROM users)
    AND deleted_at < NOW() - INTERVAL '30 days';
END;
$$ LANGUAGE plpgsql;
```

---

## User Data Export

### Export Format

```go
// internal/compliance/export.go
package compliance

import (
    "archive/zip"
    "encoding/json"
    "io"
    "time"
)

// DataExport represents a complete user data export
type DataExport struct {
    ExportID    string         `json:"export_id"`
    UserID      string         `json:"user_id"`
    GeneratedAt time.Time      `json:"generated_at"`
    ExpiresAt   time.Time      `json:"expires_at"`
    Format      string         `json:"format"` // json, csv
    Data        ExportData     `json:"data"`
}

type ExportData struct {
    Profile     ProfileData      `json:"profile"`
    Links       []LinkData       `json:"links"`
    Analytics   []AnalyticsData  `json:"analytics"`
    Sessions    []SessionData    `json:"sessions"`
    APIKeys     []APIKeyData     `json:"api_keys"`
    AuditLog    []AuditLogEntry  `json:"audit_log"`
    Preferences PreferencesData  `json:"preferences"`
}

type ProfileData struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    Plan      string    `json:"plan"`
}

type LinkData struct {
    ID          string     `json:"id"`
    ShortCode   string     `json:"short_code"`
    OriginalURL string     `json:"original_url"`
    CreatedAt   time.Time  `json:"created_at"`
    ExpiresAt   *time.Time `json:"expires_at,omitempty"`
    ClickCount  int64      `json:"click_count"`
    IsActive    bool       `json:"is_active"`
    Tags        []string   `json:"tags"`
}

type AnalyticsData struct {
    LinkID     string    `json:"link_id"`
    Date       string    `json:"date"`
    Clicks     int64     `json:"clicks"`
    Countries  map[string]int64 `json:"countries"`
    Devices    map[string]int64 `json:"devices"`
    Referrers  map[string]int64 `json:"referrers"`
}
```

### Export API Implementation

```go
// internal/api/export.go
package api

import (
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"
)

type ExportHandler struct {
    exportSvc *compliance.ExportService
}

// RequestExport initiates a data export
func (h *ExportHandler) RequestExport(w http.ResponseWriter, r *http.Request) {
    userID := auth.GetUserID(r.Context())

    export, err := h.exportSvc.InitiateExport(r.Context(), userID)
    if err != nil {
        respondError(w, http.StatusInternalServerError, "Failed to initiate export")
        return
    }

    respondJSON(w, http.StatusAccepted, map[string]interface{}{
        "export_id": export.ID,
        "status":    "processing",
        "message":   "Your data export is being prepared. You will receive an email when it's ready.",
    })
}

// GetExportStatus checks export progress
func (h *ExportHandler) GetExportStatus(w http.ResponseWriter, r *http.Request) {
    exportID := chi.URLParam(r, "exportID")
    userID := auth.GetUserID(r.Context())

    export, err := h.exportSvc.GetExport(r.Context(), exportID, userID)
    if err != nil {
        respondError(w, http.StatusNotFound, "Export not found")
        return
    }

    response := map[string]interface{}{
        "export_id": export.ID,
        "status":    export.Status,
        "created_at": export.CreatedAt,
    }

    if export.Status == "completed" {
        response["download_url"] = export.DownloadURL
        response["expires_at"] = export.ExpiresAt
    }

    respondJSON(w, http.StatusOK, response)
}

// DownloadExport serves the export file
func (h *ExportHandler) DownloadExport(w http.ResponseWriter, r *http.Request) {
    exportID := chi.URLParam(r, "exportID")
    token := r.URL.Query().Get("token")

    export, err := h.exportSvc.ValidateDownload(r.Context(), exportID, token)
    if err != nil {
        respondError(w, http.StatusForbidden, "Invalid or expired download link")
        return
    }

    w.Header().Set("Content-Type", "application/zip")
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"linkrift-export-%s.zip\"", export.ID))

    if err := h.exportSvc.StreamExport(r.Context(), export.ID, w); err != nil {
        // Error occurred after headers sent, can't change response
        return
    }
}
```

### Automated Export Generation

```go
// internal/compliance/export_service.go
package compliance

import (
    "archive/zip"
    "bytes"
    "context"
    "encoding/csv"
    "encoding/json"
    "time"
)

type ExportService struct {
    userRepo      UserRepository
    linksRepo     LinksRepository
    analyticsRepo AnalyticsRepository
    storage       StorageService
    notifier      NotificationService
}

// GenerateExport creates a complete data export
func (s *ExportService) GenerateExport(ctx context.Context, userID string) (*DataExport, error) {
    // Gather all user data
    profile, err := s.userRepo.GetProfile(ctx, userID)
    if err != nil {
        return nil, err
    }

    links, err := s.linksRepo.GetAllByUser(ctx, userID)
    if err != nil {
        return nil, err
    }

    analytics, err := s.analyticsRepo.GetByUser(ctx, userID)
    if err != nil {
        return nil, err
    }

    export := &DataExport{
        ExportID:    generateID(),
        UserID:      userID,
        GeneratedAt: time.Now(),
        ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
        Format:      "json",
        Data: ExportData{
            Profile:   toProfileData(profile),
            Links:     toLinksData(links),
            Analytics: toAnalyticsData(analytics),
        },
    }

    // Create ZIP file
    zipBuffer, err := s.createZipExport(export)
    if err != nil {
        return nil, err
    }

    // Upload to storage
    path := fmt.Sprintf("exports/%s/%s.zip", userID, export.ExportID)
    downloadURL, err := s.storage.Upload(ctx, path, zipBuffer)
    if err != nil {
        return nil, err
    }

    export.DownloadURL = downloadURL

    return export, nil
}

func (s *ExportService) createZipExport(export *DataExport) (*bytes.Buffer, error) {
    buf := new(bytes.Buffer)
    zw := zip.NewWriter(buf)

    // Add JSON export
    jsonData, err := json.MarshalIndent(export.Data, "", "  ")
    if err != nil {
        return nil, err
    }

    jw, err := zw.Create("data.json")
    if err != nil {
        return nil, err
    }
    jw.Write(jsonData)

    // Add CSV exports for links
    cw, err := zw.Create("links.csv")
    if err != nil {
        return nil, err
    }

    csvWriter := csv.NewWriter(cw)
    csvWriter.Write([]string{"ID", "Short Code", "Original URL", "Created At", "Clicks"})
    for _, link := range export.Data.Links {
        csvWriter.Write([]string{
            link.ID,
            link.ShortCode,
            link.OriginalURL,
            link.CreatedAt.Format(time.RFC3339),
            fmt.Sprintf("%d", link.ClickCount),
        })
    }
    csvWriter.Flush()

    // Add README
    readme, _ := zw.Create("README.txt")
    readme.Write([]byte(fmt.Sprintf(`Linkrift Data Export
=====================
Export ID: %s
Generated: %s
User ID: %s

This archive contains all data associated with your Linkrift account.

Files:
- data.json: Complete data export in JSON format
- links.csv: Your shortened links in CSV format

For questions, contact: privacy@linkrift.io
`, export.ExportID, export.GeneratedAt.Format(time.RFC3339), export.UserID)))

    zw.Close()
    return buf, nil
}
```

---

## Account Deletion

### Deletion Process

```
User Request → Verification → Grace Period → Data Deletion → Confirmation

1. User requests deletion via settings or support
2. Identity verification (email confirmation)
3. 14-day grace period (can cancel)
4. Data deletion process begins
5. Confirmation email sent
```

### Data Anonymization

```go
// internal/compliance/anonymize.go
package compliance

import (
    "crypto/sha256"
    "encoding/hex"
)

// AnonymizationConfig defines what to anonymize vs delete
type AnonymizationConfig struct {
    // Data to hard delete
    HardDelete []string

    // Data to anonymize (for aggregate analytics)
    Anonymize []string

    // Data to retain (legal requirements)
    Retain []string
}

var DefaultAnonymizationConfig = AnonymizationConfig{
    HardDelete: []string{
        "email",
        "name",
        "api_keys",
        "sessions",
        "password_hash",
        "profile_picture",
    },
    Anonymize: []string{
        "links",          // Keep links but remove user association
        "click_data",     // Anonymize visitor IP
    },
    Retain: []string{
        "payment_records", // Required for tax purposes
        "audit_logs",      // Required for security
    },
}

// Anonymizer handles data anonymization
type Anonymizer struct {
    salt []byte
}

// AnonymizeEmail creates a consistent but irreversible hash
func (a *Anonymizer) AnonymizeEmail(email string) string {
    hash := sha256.Sum256(append(a.salt, []byte(email)...))
    return "deleted_" + hex.EncodeToString(hash[:8])
}

// AnonymizeIP removes identifying portion of IP
func (a *Anonymizer) AnonymizeIP(ip string) string {
    // IPv4: 192.168.1.100 -> 192.168.1.0
    // IPv6: 2001:db8::1 -> 2001:db8::
    parts := strings.Split(ip, ".")
    if len(parts) == 4 {
        return parts[0] + "." + parts[1] + "." + parts[2] + ".0"
    }
    // Handle IPv6
    if idx := strings.LastIndex(ip, ":"); idx != -1 {
        return ip[:idx+1] + ":"
    }
    return "0.0.0.0"
}
```

### Deletion Implementation

```go
// internal/compliance/deletion.go
package compliance

import (
    "context"
    "time"
)

type DeletionService struct {
    userRepo      UserRepository
    linksRepo     LinksRepository
    analyticsRepo AnalyticsRepository
    paymentsRepo  PaymentsRepository
    anonymizer    *Anonymizer
    notifier      NotificationService
    auditLog      AuditLogger
}

// ScheduleDeletion initiates the deletion process
func (s *DeletionService) ScheduleDeletion(ctx context.Context, userID string) error {
    // Set deletion date (14 days from now)
    deletionDate := time.Now().Add(14 * 24 * time.Hour)

    if err := s.userRepo.ScheduleDeletion(ctx, userID, deletionDate); err != nil {
        return err
    }

    // Send confirmation email
    return s.notifier.SendDeletionScheduled(ctx, userID, deletionDate)
}

// CancelDeletion cancels a scheduled deletion
func (s *DeletionService) CancelDeletion(ctx context.Context, userID string) error {
    return s.userRepo.CancelDeletion(ctx, userID)
}

// ExecuteDeletion performs the actual deletion
func (s *DeletionService) ExecuteDeletion(ctx context.Context, userID string) error {
    s.auditLog.Log(ctx, AuditEvent{
        EventType: "account_deletion_started",
        UserID:    userID,
    })

    // 1. Export final data (for audit purposes)
    _, err := s.createFinalExport(ctx, userID)
    if err != nil {
        return err
    }

    // 2. Anonymize analytics data
    if err := s.anonymizeAnalytics(ctx, userID); err != nil {
        return err
    }

    // 3. Delete or anonymize links
    if err := s.processLinks(ctx, userID); err != nil {
        return err
    }

    // 4. Delete user data
    if err := s.deleteUserData(ctx, userID); err != nil {
        return err
    }

    // 5. Log completion
    s.auditLog.Log(ctx, AuditEvent{
        EventType: "account_deletion_completed",
        UserID:    s.anonymizer.AnonymizeEmail(userID),
    })

    return nil
}

func (s *DeletionService) processLinks(ctx context.Context, userID string) error {
    links, err := s.linksRepo.GetAllByUser(ctx, userID)
    if err != nil {
        return err
    }

    for _, link := range links {
        // Option 1: Delete the link entirely
        // s.linksRepo.Delete(ctx, link.ID)

        // Option 2: Anonymize the link (keep for redirect but remove user)
        if err := s.linksRepo.Anonymize(ctx, link.ID); err != nil {
            return err
        }
    }

    return nil
}

func (s *DeletionService) deleteUserData(ctx context.Context, userID string) error {
    // Delete in order (respect foreign keys)

    // API keys
    if err := s.userRepo.DeleteAPIKeys(ctx, userID); err != nil {
        return err
    }

    // Sessions
    if err := s.userRepo.DeleteSessions(ctx, userID); err != nil {
        return err
    }

    // User profile (this is the final step)
    return s.userRepo.Delete(ctx, userID)
}
```

---

## Cookie Policy

### Cookie Categories

```go
// internal/compliance/cookies.go
package compliance

type CookieCategory string

const (
    CookieStrictlyNecessary CookieCategory = "strictly_necessary"
    CookiePerformance       CookieCategory = "performance"
    CookieFunctional        CookieCategory = "functional"
    CookieTargeting         CookieCategory = "targeting"
)

type Cookie struct {
    Name        string
    Category    CookieCategory
    Purpose     string
    Duration    string
    Provider    string
    Required    bool
}

// GetCookieInventory returns all cookies used
func GetCookieInventory() []Cookie {
    return []Cookie{
        // Strictly Necessary
        {
            Name:     "__session",
            Category: CookieStrictlyNecessary,
            Purpose:  "Maintains user session state",
            Duration: "Session",
            Provider: "Linkrift",
            Required: true,
        },
        {
            Name:     "__csrf",
            Category: CookieStrictlyNecessary,
            Purpose:  "Prevents cross-site request forgery",
            Duration: "Session",
            Provider: "Linkrift",
            Required: true,
        },
        {
            Name:     "cookie_consent",
            Category: CookieStrictlyNecessary,
            Purpose:  "Stores cookie consent preferences",
            Duration: "1 year",
            Provider: "Linkrift",
            Required: true,
        },

        // Performance
        {
            Name:     "_ga",
            Category: CookiePerformance,
            Purpose:  "Google Analytics - distinguishes users",
            Duration: "2 years",
            Provider: "Google",
            Required: false,
        },
        {
            Name:     "_gid",
            Category: CookiePerformance,
            Purpose:  "Google Analytics - distinguishes users",
            Duration: "24 hours",
            Provider: "Google",
            Required: false,
        },

        // Functional
        {
            Name:     "theme",
            Category: CookieFunctional,
            Purpose:  "Stores user's theme preference",
            Duration: "1 year",
            Provider: "Linkrift",
            Required: false,
        },
        {
            Name:     "locale",
            Category: CookieFunctional,
            Purpose:  "Stores user's language preference",
            Duration: "1 year",
            Provider: "Linkrift",
            Required: false,
        },
    }
}
```

### Cookie Consent Implementation

```go
// internal/compliance/consent.go
package compliance

import (
    "encoding/json"
    "net/http"
    "time"
)

type CookieConsent struct {
    StrictlyNecessary bool      `json:"strictly_necessary"` // Always true
    Performance       bool      `json:"performance"`
    Functional        bool      `json:"functional"`
    Targeting         bool      `json:"targeting"`
    ConsentedAt       time.Time `json:"consented_at"`
    Version           string    `json:"version"` // Consent version for re-consent
}

const ConsentCookieName = "cookie_consent"
const ConsentVersion = "2025-01-24"

type ConsentService struct {
    auditLog AuditLogger
}

// GetConsent retrieves current consent from cookie
func (s *ConsentService) GetConsent(r *http.Request) *CookieConsent {
    cookie, err := r.Cookie(ConsentCookieName)
    if err != nil {
        return nil
    }

    var consent CookieConsent
    if err := json.Unmarshal([]byte(cookie.Value), &consent); err != nil {
        return nil
    }

    // Check if re-consent needed due to version change
    if consent.Version != ConsentVersion {
        return nil
    }

    return &consent
}

// SetConsent saves consent preferences
func (s *ConsentService) SetConsent(w http.ResponseWriter, r *http.Request, consent *CookieConsent) {
    consent.StrictlyNecessary = true // Always required
    consent.ConsentedAt = time.Now()
    consent.Version = ConsentVersion

    data, _ := json.Marshal(consent)

    http.SetCookie(w, &http.Cookie{
        Name:     ConsentCookieName,
        Value:    string(data),
        Path:     "/",
        MaxAge:   365 * 24 * 60 * 60, // 1 year
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteLaxMode,
    })

    // Log consent
    s.auditLog.Log(r.Context(), AuditEvent{
        EventType: "cookie_consent",
        IP:        getIP(r),
        Details: map[string]interface{}{
            "performance": consent.Performance,
            "functional":  consent.Functional,
            "targeting":   consent.Targeting,
        },
    })
}

// CookieConsentMiddleware blocks non-essential cookies without consent
func CookieConsentMiddleware(consentSvc *ConsentService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            consent := consentSvc.GetConsent(r)

            // Add consent status to context
            ctx := context.WithValue(r.Context(), "cookie_consent", consent)

            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### Cookie Management

```typescript
// frontend/src/components/CookieBanner.tsx
import React, { useState, useEffect } from 'react';

interface CookieConsent {
  strictly_necessary: boolean;
  performance: boolean;
  functional: boolean;
  targeting: boolean;
}

export const CookieBanner: React.FC = () => {
  const [showBanner, setShowBanner] = useState(false);
  const [consent, setConsent] = useState<CookieConsent>({
    strictly_necessary: true,
    performance: false,
    functional: false,
    targeting: false,
  });

  useEffect(() => {
    const existingConsent = getCookieConsent();
    if (!existingConsent) {
      setShowBanner(true);
    }
  }, []);

  const acceptAll = () => {
    const fullConsent = {
      strictly_necessary: true,
      performance: true,
      functional: true,
      targeting: true,
    };
    saveConsent(fullConsent);
    setShowBanner(false);
  };

  const acceptSelected = () => {
    saveConsent(consent);
    setShowBanner(false);
  };

  const rejectAll = () => {
    const minimalConsent = {
      strictly_necessary: true,
      performance: false,
      functional: false,
      targeting: false,
    };
    saveConsent(minimalConsent);
    setShowBanner(false);
  };

  if (!showBanner) return null;

  return (
    <div className="cookie-banner">
      <h3>Cookie Preferences</h3>
      <p>
        We use cookies to enhance your experience. Some are essential,
        while others help us improve our service.
      </p>

      <div className="cookie-options">
        <label>
          <input type="checkbox" checked disabled />
          Strictly Necessary (Required)
        </label>
        <label>
          <input
            type="checkbox"
            checked={consent.performance}
            onChange={(e) => setConsent({...consent, performance: e.target.checked})}
          />
          Performance & Analytics
        </label>
        <label>
          <input
            type="checkbox"
            checked={consent.functional}
            onChange={(e) => setConsent({...consent, functional: e.target.checked})}
          />
          Functional
        </label>
      </div>

      <div className="cookie-actions">
        <button onClick={rejectAll}>Reject All</button>
        <button onClick={acceptSelected}>Save Preferences</button>
        <button onClick={acceptAll}>Accept All</button>
      </div>

      <a href="/privacy#cookies">Learn more about our cookies</a>
    </div>
  );
};
```

---

## Terms of Service Considerations

### Key Terms

```markdown
## Terms of Service - Key Provisions

### 1. Service Description
- URL shortening and link management
- Analytics and tracking (with user consent)
- API access for automated link creation

### 2. User Responsibilities
- Maintain account security
- Comply with Acceptable Use Policy
- Accurate registration information

### 3. Prohibited Content
- Malware or phishing links
- Illegal content
- Spam or bulk unsolicited messages
- Content violating third-party rights

### 4. Data and Privacy
- Reference to Privacy Policy
- User data ownership
- Data portability rights

### 5. Service Availability
- Best-effort availability
- Scheduled maintenance windows
- Force majeure provisions

### 6. Limitation of Liability
- Service provided "as is"
- Limitation on damages
- Indemnification

### 7. Termination
- User right to terminate
- Company right to terminate for violations
- Data retention after termination

### 8. Governing Law
- Jurisdiction
- Dispute resolution
```

### Acceptable Use Policy

```go
// internal/compliance/aup.go
package compliance

type ViolationType string

const (
    ViolationMalware      ViolationType = "malware"
    ViolationPhishing     ViolationType = "phishing"
    ViolationSpam         ViolationType = "spam"
    ViolationIllegal      ViolationType = "illegal_content"
    ViolationCopyright    ViolationType = "copyright"
    ViolationHarassment   ViolationType = "harassment"
    ViolationFraud        ViolationType = "fraud"
)

type AUPViolation struct {
    ID          string
    LinkID      string
    UserID      string
    Type        ViolationType
    ReportedBy  string
    ReportedAt  time.Time
    Status      string // pending, confirmed, dismissed
    Action      string // warning, disabled, terminated
    Notes       string
}

type AUPService struct {
    repo       AUPRepository
    linksRepo  LinksRepository
    userRepo   UserRepository
    scanner    URLScanner
    notifier   NotificationService
}

// CheckURL validates a URL against AUP
func (s *AUPService) CheckURL(ctx context.Context, url string) error {
    // Check against known malware/phishing databases
    if s.scanner.IsMalicious(url) {
        return fmt.Errorf("URL flagged as malicious")
    }

    // Check against blocklist
    if s.isBlockedDomain(url) {
        return fmt.Errorf("domain is blocked")
    }

    return nil
}

// ReportViolation handles violation reports
func (s *AUPService) ReportViolation(ctx context.Context, report *AUPViolation) error {
    report.ID = generateID()
    report.ReportedAt = time.Now()
    report.Status = "pending"

    if err := s.repo.Create(ctx, report); err != nil {
        return err
    }

    // Auto-disable if high-confidence violation
    if s.isHighConfidenceViolation(report) {
        return s.autoDisable(ctx, report)
    }

    return nil
}

// ProcessViolation handles confirmed violations
func (s *AUPService) ProcessViolation(ctx context.Context, violationID string, action string) error {
    violation, err := s.repo.Get(ctx, violationID)
    if err != nil {
        return err
    }

    switch action {
    case "warning":
        return s.issueWarning(ctx, violation)
    case "disable_link":
        return s.disableLink(ctx, violation)
    case "suspend_account":
        return s.suspendAccount(ctx, violation)
    case "terminate_account":
        return s.terminateAccount(ctx, violation)
    default:
        return fmt.Errorf("unknown action: %s", action)
    }
}
```

### Service Level Agreement

```go
// internal/compliance/sla.go
package compliance

type SLAMetrics struct {
    // Uptime
    TargetUptime     float64 // 99.9%
    MeasuredUptime   float64
    DowntimeMinutes  int

    // Response Time
    TargetP99Latency time.Duration // 100ms
    MeasuredP99      time.Duration

    // Error Rate
    TargetErrorRate  float64 // 0.1%
    MeasuredErrorRate float64
}

func (m *SLAMetrics) IsMeetingSLA() bool {
    return m.MeasuredUptime >= m.TargetUptime &&
           m.MeasuredP99 <= m.TargetP99Latency &&
           m.MeasuredErrorRate <= m.TargetErrorRate
}

func (m *SLAMetrics) CalculateCredit() float64 {
    // Credit calculation based on downtime
    // 99.9% = 43.2 minutes/month allowed
    // Below 99.9%: 10% credit
    // Below 99.5%: 25% credit
    // Below 99.0%: 50% credit

    if m.MeasuredUptime >= 99.9 {
        return 0
    } else if m.MeasuredUptime >= 99.5 {
        return 10
    } else if m.MeasuredUptime >= 99.0 {
        return 25
    }
    return 50
}
```

---

## Compliance Monitoring

```go
// internal/compliance/monitoring.go
package compliance

import (
    "context"
    "time"
)

type ComplianceMonitor struct {
    ccpaSvc     *CCPAService
    gdprSvc     *GDPRService
    retentionSvc *RetentionService
    auditLog    AuditLogger
}

// RunComplianceChecks performs regular compliance checks
func (m *ComplianceMonitor) RunComplianceChecks(ctx context.Context) *ComplianceReport {
    report := &ComplianceReport{
        GeneratedAt: time.Now(),
        Checks:      make([]ComplianceCheck, 0),
    }

    // Check pending GDPR requests
    pendingGDPR, _ := m.gdprSvc.GetPendingRequests(ctx)
    report.Checks = append(report.Checks, ComplianceCheck{
        Name:   "GDPR Pending Requests",
        Status: len(pendingGDPR) == 0,
        Count:  len(pendingGDPR),
        Note:   fmt.Sprintf("%d pending requests", len(pendingGDPR)),
    })

    // Check overdue GDPR requests (>30 days)
    overdueGDPR := filterOverdue(pendingGDPR, 30*24*time.Hour)
    report.Checks = append(report.Checks, ComplianceCheck{
        Name:   "GDPR Overdue Requests",
        Status: len(overdueGDPR) == 0,
        Count:  len(overdueGDPR),
        Note:   fmt.Sprintf("%d overdue requests", len(overdueGDPR)),
    })

    // Check pending CCPA requests
    pendingCCPA, _ := m.ccpaSvc.GetPendingRequests(ctx)
    report.Checks = append(report.Checks, ComplianceCheck{
        Name:   "CCPA Pending Requests",
        Status: len(pendingCCPA) == 0,
        Count:  len(pendingCCPA),
    })

    // Check data retention compliance
    retentionStatus := m.retentionSvc.CheckCompliance(ctx)
    report.Checks = append(report.Checks, ComplianceCheck{
        Name:   "Data Retention Compliance",
        Status: retentionStatus.Compliant,
        Note:   retentionStatus.Message,
    })

    return report
}

type ComplianceReport struct {
    GeneratedAt time.Time
    Checks      []ComplianceCheck
    OverallStatus bool
}

type ComplianceCheck struct {
    Name   string
    Status bool
    Count  int
    Note   string
}
```

---

## Audit Trail

```go
// internal/compliance/audit.go
package compliance

import (
    "context"
    "time"
)

type ComplianceAuditLog struct {
    ID          string                 `json:"id"`
    Timestamp   time.Time              `json:"timestamp"`
    EventType   string                 `json:"event_type"`
    Actor       string                 `json:"actor"` // user_id or "system"
    Subject     string                 `json:"subject"` // affected user_id
    Action      string                 `json:"action"`
    Resource    string                 `json:"resource"`
    Details     map[string]interface{} `json:"details"`
    IP          string                 `json:"ip,omitempty"`
    UserAgent   string                 `json:"user_agent,omitempty"`
}

// Compliance-specific event types
const (
    AuditConsentGiven     = "consent_given"
    AuditConsentWithdrawn = "consent_withdrawn"
    AuditDataExport       = "data_export"
    AuditDataDeletion     = "data_deletion"
    AuditAccessRequest    = "access_request"
    AuditRectification    = "rectification"
    AuditDataBreach       = "data_breach"
    AuditPolicyChange     = "policy_change"
)

type ComplianceAuditLogger struct {
    repo AuditRepository
}

func (l *ComplianceAuditLogger) LogConsentChange(ctx context.Context, userID string, consentType string, granted bool) {
    eventType := AuditConsentGiven
    if !granted {
        eventType = AuditConsentWithdrawn
    }

    l.repo.Create(ctx, &ComplianceAuditLog{
        ID:        generateID(),
        Timestamp: time.Now(),
        EventType: eventType,
        Actor:     userID,
        Subject:   userID,
        Action:    "consent_change",
        Details: map[string]interface{}{
            "consent_type": consentType,
            "granted":      granted,
        },
    })
}

func (l *ComplianceAuditLogger) LogDataSubjectRequest(ctx context.Context, req *DataSubjectRequest) {
    l.repo.Create(ctx, &ComplianceAuditLog{
        ID:        generateID(),
        Timestamp: time.Now(),
        EventType: AuditAccessRequest,
        Actor:     req.UserID,
        Subject:   req.UserID,
        Action:    string(req.Right),
        Details: map[string]interface{}{
            "request_id": req.ID,
            "status":     req.Status,
        },
    })
}

func (l *ComplianceAuditLogger) LogDataBreach(ctx context.Context, breach *DataBreach) {
    l.repo.Create(ctx, &ComplianceAuditLog{
        ID:        generateID(),
        Timestamp: time.Now(),
        EventType: AuditDataBreach,
        Actor:     "system",
        Action:    "breach_detected",
        Details: map[string]interface{}{
            "breach_id":       breach.ID,
            "affected_count":  breach.AffectedUsers,
            "data_categories": breach.DataCategories,
            "severity":        breach.Severity,
        },
    })
}
```

---

## Related Documentation

- [SECURITY.md](./SECURITY.md) - Security architecture and implementation
- [../operations/MONITORING_LOGGING.md](../operations/MONITORING_LOGGING.md) - Audit logging infrastructure
- [../architecture/DATA_MODEL.md](../architecture/DATA_MODEL.md) - Data schema and storage
