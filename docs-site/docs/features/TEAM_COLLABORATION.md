# Team Collaboration

> Last Updated: 2025-01-24

Linkrift provides robust team collaboration features enabling organizations to work together efficiently with fine-grained access control, workspace management, and comprehensive audit logging.

## Table of Contents

- [Overview](#overview)
- [Workspace Architecture](#workspace-architecture)
- [Role-Based Access Control (RBAC)](#role-based-access-control-rbac)
  - [Roles](#roles)
  - [Permission Matrix](#permission-matrix)
- [Invitation Flow](#invitation-flow)
- [Audit Logging](#audit-logging)
- [API Endpoints](#api-endpoints)
- [React Components](#react-components)

---

## Overview

Team collaboration in Linkrift includes:

- **Workspace-based organization** for isolating team resources
- **Role-based access control** with four predefined roles
- **Granular permissions** for different actions and resources
- **Invitation system** with email-based onboarding
- **Comprehensive audit logging** for compliance and security

## Workspace Architecture

```go
// internal/workspace/models.go
package workspace

import (
	"time"
)

// Workspace represents a team workspace
type Workspace struct {
	ID          string          `json:"id" db:"id"`
	Name        string          `json:"name" db:"name"`
	Slug        string          `json:"slug" db:"slug"`
	Description string          `json:"description,omitempty" db:"description"`
	LogoURL     string          `json:"logo_url,omitempty" db:"logo_url"`
	OwnerID     string          `json:"owner_id" db:"owner_id"`
	Settings    WorkspaceSettings `json:"settings" db:"settings"`
	PlanID      string          `json:"plan_id" db:"plan_id"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

// WorkspaceSettings contains workspace configuration
type WorkspaceSettings struct {
	DefaultDomain      string `json:"default_domain"`
	DefaultRedirectType string `json:"default_redirect_type"`
	RequireApproval    bool   `json:"require_approval"`
	AllowPublicLinks   bool   `json:"allow_public_links"`
	EnforceSSO         bool   `json:"enforce_sso"`
	AllowedEmailDomains []string `json:"allowed_email_domains,omitempty"`
}

// Member represents a workspace member
type Member struct {
	ID          string    `json:"id" db:"id"`
	WorkspaceID string    `json:"workspace_id" db:"workspace_id"`
	UserID      string    `json:"user_id" db:"user_id"`
	Role        Role      `json:"role" db:"role"`
	JoinedAt    time.Time `json:"joined_at" db:"joined_at"`
	InvitedBy   string    `json:"invited_by,omitempty" db:"invited_by"`

	// Populated from User table
	User *User `json:"user,omitempty" db:"-"`
}

// User represents basic user information
type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// WorkspaceService handles workspace operations
type WorkspaceService struct {
	repo       *db.WorkspaceRepository
	memberRepo *db.MemberRepository
	auditLog   *AuditLogger
}

// NewWorkspaceService creates a new workspace service
func NewWorkspaceService(
	repo *db.WorkspaceRepository,
	memberRepo *db.MemberRepository,
	auditLog *AuditLogger,
) *WorkspaceService {
	return &WorkspaceService{
		repo:       repo,
		memberRepo: memberRepo,
		auditLog:   auditLog,
	}
}

// Create creates a new workspace
func (ws *WorkspaceService) Create(ctx context.Context, input *CreateWorkspaceInput) (*Workspace, error) {
	workspace := &Workspace{
		Name:        input.Name,
		Slug:        generateSlug(input.Name),
		Description: input.Description,
		OwnerID:     input.OwnerID,
		PlanID:      "free",
		Settings: WorkspaceSettings{
			DefaultRedirectType: "permanent",
			AllowPublicLinks:    true,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := ws.repo.Create(ctx, workspace); err != nil {
		return nil, err
	}

	// Add owner as first member
	member := &Member{
		WorkspaceID: workspace.ID,
		UserID:      input.OwnerID,
		Role:        RoleOwner,
		JoinedAt:    time.Now(),
	}

	if err := ws.memberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	ws.auditLog.Log(ctx, AuditEvent{
		WorkspaceID: workspace.ID,
		ActorID:     input.OwnerID,
		Action:      "workspace.created",
		ResourceType: "workspace",
		ResourceID:  workspace.ID,
	})

	return workspace, nil
}

// CreateWorkspaceInput represents workspace creation input
type CreateWorkspaceInput struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description"`
	OwnerID     string `json:"-"`
}
```

---

## Role-Based Access Control (RBAC)

### Roles

```go
// internal/workspace/rbac.go
package workspace

// Role defines user roles within a workspace
type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

// Permission defines granular permissions
type Permission string

const (
	// Workspace permissions
	PermWorkspaceRead     Permission = "workspace:read"
	PermWorkspaceUpdate   Permission = "workspace:update"
	PermWorkspaceDelete   Permission = "workspace:delete"
	PermWorkspaceSettings Permission = "workspace:settings"

	// Member permissions
	PermMemberList    Permission = "member:list"
	PermMemberInvite  Permission = "member:invite"
	PermMemberRemove  Permission = "member:remove"
	PermMemberRole    Permission = "member:role"

	// Link permissions
	PermLinkCreate    Permission = "link:create"
	PermLinkRead      Permission = "link:read"
	PermLinkUpdate    Permission = "link:update"
	PermLinkDelete    Permission = "link:delete"
	PermLinkBulk      Permission = "link:bulk"

	// Analytics permissions
	PermAnalyticsView     Permission = "analytics:view"
	PermAnalyticsExport   Permission = "analytics:export"
	PermAnalyticsAdvanced Permission = "analytics:advanced"

	// Domain permissions
	PermDomainList    Permission = "domain:list"
	PermDomainAdd     Permission = "domain:add"
	PermDomainRemove  Permission = "domain:remove"
	PermDomainVerify  Permission = "domain:verify"

	// Bio page permissions
	PermBioPageCreate Permission = "biopage:create"
	PermBioPageRead   Permission = "biopage:read"
	PermBioPageUpdate Permission = "biopage:update"
	PermBioPageDelete Permission = "biopage:delete"

	// QR code permissions
	PermQRCreate Permission = "qr:create"
	PermQRRead   Permission = "qr:read"
	PermQRDelete Permission = "qr:delete"

	// Webhook permissions
	PermWebhookList   Permission = "webhook:list"
	PermWebhookCreate Permission = "webhook:create"
	PermWebhookUpdate Permission = "webhook:update"
	PermWebhookDelete Permission = "webhook:delete"

	// Billing permissions
	PermBillingView   Permission = "billing:view"
	PermBillingManage Permission = "billing:manage"

	// Audit log permissions
	PermAuditView Permission = "audit:view"

	// API permissions
	PermAPIKeyCreate Permission = "apikey:create"
	PermAPIKeyRevoke Permission = "apikey:revoke"
)

// RolePermissions maps roles to their permissions
var RolePermissions = map[Role][]Permission{
	RoleOwner: {
		// All permissions
		PermWorkspaceRead, PermWorkspaceUpdate, PermWorkspaceDelete, PermWorkspaceSettings,
		PermMemberList, PermMemberInvite, PermMemberRemove, PermMemberRole,
		PermLinkCreate, PermLinkRead, PermLinkUpdate, PermLinkDelete, PermLinkBulk,
		PermAnalyticsView, PermAnalyticsExport, PermAnalyticsAdvanced,
		PermDomainList, PermDomainAdd, PermDomainRemove, PermDomainVerify,
		PermBioPageCreate, PermBioPageRead, PermBioPageUpdate, PermBioPageDelete,
		PermQRCreate, PermQRRead, PermQRDelete,
		PermWebhookList, PermWebhookCreate, PermWebhookUpdate, PermWebhookDelete,
		PermBillingView, PermBillingManage,
		PermAuditView,
		PermAPIKeyCreate, PermAPIKeyRevoke,
	},
	RoleAdmin: {
		PermWorkspaceRead, PermWorkspaceUpdate, PermWorkspaceSettings,
		PermMemberList, PermMemberInvite, PermMemberRemove, PermMemberRole,
		PermLinkCreate, PermLinkRead, PermLinkUpdate, PermLinkDelete, PermLinkBulk,
		PermAnalyticsView, PermAnalyticsExport, PermAnalyticsAdvanced,
		PermDomainList, PermDomainAdd, PermDomainRemove, PermDomainVerify,
		PermBioPageCreate, PermBioPageRead, PermBioPageUpdate, PermBioPageDelete,
		PermQRCreate, PermQRRead, PermQRDelete,
		PermWebhookList, PermWebhookCreate, PermWebhookUpdate, PermWebhookDelete,
		PermBillingView,
		PermAuditView,
		PermAPIKeyCreate, PermAPIKeyRevoke,
	},
	RoleEditor: {
		PermWorkspaceRead,
		PermMemberList,
		PermLinkCreate, PermLinkRead, PermLinkUpdate, PermLinkDelete,
		PermAnalyticsView,
		PermDomainList,
		PermBioPageCreate, PermBioPageRead, PermBioPageUpdate,
		PermQRCreate, PermQRRead, PermQRDelete,
		PermWebhookList,
	},
	RoleViewer: {
		PermWorkspaceRead,
		PermMemberList,
		PermLinkRead,
		PermAnalyticsView,
		PermDomainList,
		PermBioPageRead,
		PermQRRead,
	},
}

// RoleHierarchy defines role hierarchy for promotion/demotion
var RoleHierarchy = map[Role]int{
	RoleOwner:  4,
	RoleAdmin:  3,
	RoleEditor: 2,
	RoleViewer: 1,
}

// RBAC provides role-based access control
type RBAC struct{}

// NewRBAC creates a new RBAC instance
func NewRBAC() *RBAC {
	return &RBAC{}
}

// HasPermission checks if a role has a specific permission
func (r *RBAC) HasPermission(role Role, permission Permission) bool {
	permissions, ok := RolePermissions[role]
	if !ok {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}

	return false
}

// CanManageRole checks if a role can manage another role
func (r *RBAC) CanManageRole(actorRole, targetRole Role) bool {
	actorLevel, ok1 := RoleHierarchy[actorRole]
	targetLevel, ok2 := RoleHierarchy[targetRole]

	if !ok1 || !ok2 {
		return false
	}

	// Can only manage roles below your level
	// Owner can manage everyone, Admin can manage Editor and Viewer, etc.
	return actorLevel > targetLevel
}

// GetAssignableRoles returns roles that can be assigned by a given role
func (r *RBAC) GetAssignableRoles(role Role) []Role {
	level := RoleHierarchy[role]
	var assignable []Role

	for r, l := range RoleHierarchy {
		if l < level && r != RoleOwner { // Can't assign owner role
			assignable = append(assignable, r)
		}
	}

	return assignable
}
```

### Permission Matrix

| Permission | Owner | Admin | Editor | Viewer |
|------------|-------|-------|--------|--------|
| **Workspace** |
| Read workspace | Yes | Yes | Yes | Yes |
| Update workspace | Yes | Yes | No | No |
| Delete workspace | Yes | No | No | No |
| Manage settings | Yes | Yes | No | No |
| **Members** |
| List members | Yes | Yes | Yes | Yes |
| Invite members | Yes | Yes | No | No |
| Remove members | Yes | Yes | No | No |
| Change roles | Yes | Yes | No | No |
| **Links** |
| Create links | Yes | Yes | Yes | No |
| Read links | Yes | Yes | Yes | Yes |
| Update links | Yes | Yes | Yes | No |
| Delete links | Yes | Yes | Yes | No |
| Bulk operations | Yes | Yes | No | No |
| **Analytics** |
| View analytics | Yes | Yes | Yes | Yes |
| Export data | Yes | Yes | No | No |
| Advanced analytics | Yes | Yes | No | No |
| **Domains** |
| List domains | Yes | Yes | Yes | Yes |
| Add domains | Yes | Yes | No | No |
| Remove domains | Yes | Yes | No | No |
| Verify domains | Yes | Yes | No | No |
| **Bio Pages** |
| Create pages | Yes | Yes | Yes | No |
| Read pages | Yes | Yes | Yes | Yes |
| Update pages | Yes | Yes | Yes | No |
| Delete pages | Yes | Yes | No | No |
| **Webhooks** |
| List webhooks | Yes | Yes | Yes | No |
| Create webhooks | Yes | Yes | No | No |
| Update webhooks | Yes | Yes | No | No |
| Delete webhooks | Yes | Yes | No | No |
| **Billing** |
| View billing | Yes | Yes | No | No |
| Manage billing | Yes | No | No | No |
| **Audit** |
| View audit logs | Yes | Yes | No | No |

---

## Invitation Flow

```go
// internal/workspace/invitation.go
package workspace

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"
)

// Invitation represents a workspace invitation
type Invitation struct {
	ID          string    `json:"id" db:"id"`
	WorkspaceID string    `json:"workspace_id" db:"workspace_id"`
	Email       string    `json:"email" db:"email"`
	Role        Role      `json:"role" db:"role"`
	Token       string    `json:"-" db:"token"`
	InvitedBy   string    `json:"invited_by" db:"invited_by"`
	ExpiresAt   time.Time `json:"expires_at" db:"expires_at"`
	AcceptedAt  *time.Time `json:"accepted_at,omitempty" db:"accepted_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`

	// Populated fields
	InvitedByUser *User `json:"invited_by_user,omitempty" db:"-"`
}

// InvitationService handles invitation operations
type InvitationService struct {
	repo       *db.InvitationRepository
	memberRepo *db.MemberRepository
	userRepo   *db.UserRepository
	rbac       *RBAC
	mailer     EmailSender
	auditLog   *AuditLogger
}

// NewInvitationService creates a new invitation service
func NewInvitationService(
	repo *db.InvitationRepository,
	memberRepo *db.MemberRepository,
	userRepo *db.UserRepository,
	rbac *RBAC,
	mailer EmailSender,
	auditLog *AuditLogger,
) *InvitationService {
	return &InvitationService{
		repo:       repo,
		memberRepo: memberRepo,
		userRepo:   userRepo,
		rbac:       rbac,
		mailer:     mailer,
		auditLog:   auditLog,
	}
}

// CreateInvitation sends a workspace invitation
func (is *InvitationService) CreateInvitation(
	ctx context.Context,
	workspaceID string,
	inviterID string,
	inviterRole Role,
	email string,
	role Role,
) (*Invitation, error) {
	// Validate inviter can assign this role
	if !is.rbac.CanManageRole(inviterRole, role) {
		return nil, ErrInsufficientPermissions
	}

	// Check if user is already a member
	existing, _ := is.memberRepo.GetByEmail(ctx, workspaceID, email)
	if existing != nil {
		return nil, ErrAlreadyMember
	}

	// Check for existing pending invitation
	pendingInv, _ := is.repo.GetPendingByEmail(ctx, workspaceID, email)
	if pendingInv != nil {
		// Resend existing invitation
		return is.resendInvitation(ctx, pendingInv)
	}

	// Generate invitation token
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	invitation := &Invitation{
		WorkspaceID: workspaceID,
		Email:       email,
		Role:        role,
		Token:       token,
		InvitedBy:   inviterID,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour), // 7 days
		CreatedAt:   time.Now(),
	}

	if err := is.repo.Create(ctx, invitation); err != nil {
		return nil, err
	}

	// Send invitation email
	if err := is.sendInvitationEmail(ctx, invitation); err != nil {
		// Log but don't fail
	}

	is.auditLog.Log(ctx, AuditEvent{
		WorkspaceID:  workspaceID,
		ActorID:      inviterID,
		Action:       "member.invited",
		ResourceType: "invitation",
		ResourceID:   invitation.ID,
		Metadata: map[string]interface{}{
			"email": email,
			"role":  role,
		},
	})

	return invitation, nil
}

// AcceptInvitation accepts a workspace invitation
func (is *InvitationService) AcceptInvitation(ctx context.Context, token string, userID string) (*Member, error) {
	invitation, err := is.repo.GetByToken(ctx, token)
	if err != nil {
		return nil, ErrInvitationNotFound
	}

	if invitation.AcceptedAt != nil {
		return nil, ErrInvitationAlreadyAccepted
	}

	if time.Now().After(invitation.ExpiresAt) {
		return nil, ErrInvitationExpired
	}

	// Verify user email matches invitation
	user, err := is.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.Email != invitation.Email {
		return nil, ErrEmailMismatch
	}

	// Create member
	member := &Member{
		WorkspaceID: invitation.WorkspaceID,
		UserID:      userID,
		Role:        invitation.Role,
		JoinedAt:    time.Now(),
		InvitedBy:   invitation.InvitedBy,
	}

	if err := is.memberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	// Mark invitation as accepted
	now := time.Now()
	invitation.AcceptedAt = &now
	is.repo.Update(ctx, invitation)

	is.auditLog.Log(ctx, AuditEvent{
		WorkspaceID:  invitation.WorkspaceID,
		ActorID:      userID,
		Action:       "member.joined",
		ResourceType: "member",
		ResourceID:   member.ID,
		Metadata: map[string]interface{}{
			"role":         invitation.Role,
			"invitation_id": invitation.ID,
		},
	})

	return member, nil
}

// RevokeInvitation cancels a pending invitation
func (is *InvitationService) RevokeInvitation(ctx context.Context, invitationID string, actorID string) error {
	invitation, err := is.repo.GetByID(ctx, invitationID)
	if err != nil {
		return ErrInvitationNotFound
	}

	if invitation.AcceptedAt != nil {
		return ErrInvitationAlreadyAccepted
	}

	if err := is.repo.Delete(ctx, invitationID); err != nil {
		return err
	}

	is.auditLog.Log(ctx, AuditEvent{
		WorkspaceID:  invitation.WorkspaceID,
		ActorID:      actorID,
		Action:       "invitation.revoked",
		ResourceType: "invitation",
		ResourceID:   invitationID,
	})

	return nil
}

func (is *InvitationService) sendInvitationEmail(ctx context.Context, invitation *Invitation) error {
	inviter, _ := is.userRepo.GetByID(ctx, invitation.InvitedBy)
	workspace, _ := is.repo.GetWorkspace(ctx, invitation.WorkspaceID)

	return is.mailer.Send(ctx, &Email{
		To:       invitation.Email,
		Subject:  fmt.Sprintf("You've been invited to join %s on Linkrift", workspace.Name),
		Template: "invitation",
		Data: map[string]interface{}{
			"workspace_name": workspace.Name,
			"inviter_name":   inviter.Name,
			"role":           invitation.Role,
			"invite_url":     fmt.Sprintf("https://app.linkrift.io/invite/%s", invitation.Token),
			"expires_at":     invitation.ExpiresAt,
		},
	})
}

func (is *InvitationService) resendInvitation(ctx context.Context, invitation *Invitation) (*Invitation, error) {
	// Extend expiration
	invitation.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	is.repo.Update(ctx, invitation)

	// Resend email
	is.sendInvitationEmail(ctx, invitation)

	return invitation, nil
}

func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
```

---

## Audit Logging

```go
// internal/workspace/audit.go
package workspace

import (
	"context"
	"encoding/json"
	"time"
)

// AuditEvent represents an audit log entry
type AuditEvent struct {
	ID           string                 `json:"id" db:"id"`
	WorkspaceID  string                 `json:"workspace_id" db:"workspace_id"`
	ActorID      string                 `json:"actor_id" db:"actor_id"`
	ActorEmail   string                 `json:"actor_email" db:"actor_email"`
	Action       string                 `json:"action" db:"action"`
	ResourceType string                 `json:"resource_type" db:"resource_type"`
	ResourceID   string                 `json:"resource_id" db:"resource_id"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	IPAddress    string                 `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    string                 `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

// AuditAction defines standard audit actions
type AuditAction string

const (
	// Workspace actions
	ActionWorkspaceCreated  AuditAction = "workspace.created"
	ActionWorkspaceUpdated  AuditAction = "workspace.updated"
	ActionWorkspaceDeleted  AuditAction = "workspace.deleted"
	ActionSettingsUpdated   AuditAction = "workspace.settings_updated"

	// Member actions
	ActionMemberInvited     AuditAction = "member.invited"
	ActionMemberJoined      AuditAction = "member.joined"
	ActionMemberRemoved     AuditAction = "member.removed"
	ActionMemberRoleChanged AuditAction = "member.role_changed"
	ActionInvitationRevoked AuditAction = "invitation.revoked"

	// Link actions
	ActionLinkCreated AuditAction = "link.created"
	ActionLinkUpdated AuditAction = "link.updated"
	ActionLinkDeleted AuditAction = "link.deleted"
	ActionLinkBulkCreated AuditAction = "link.bulk_created"
	ActionLinkBulkDeleted AuditAction = "link.bulk_deleted"

	// Domain actions
	ActionDomainAdded    AuditAction = "domain.added"
	ActionDomainVerified AuditAction = "domain.verified"
	ActionDomainRemoved  AuditAction = "domain.removed"

	// Bio page actions
	ActionBioPageCreated   AuditAction = "biopage.created"
	ActionBioPageUpdated   AuditAction = "biopage.updated"
	ActionBioPageDeleted   AuditAction = "biopage.deleted"
	ActionBioPagePublished AuditAction = "biopage.published"

	// Webhook actions
	ActionWebhookCreated AuditAction = "webhook.created"
	ActionWebhookUpdated AuditAction = "webhook.updated"
	ActionWebhookDeleted AuditAction = "webhook.deleted"

	// Billing actions
	ActionSubscriptionCreated  AuditAction = "subscription.created"
	ActionSubscriptionCanceled AuditAction = "subscription.canceled"
	ActionSubscriptionChanged  AuditAction = "subscription.changed"
	ActionPaymentMethodAdded   AuditAction = "payment.method_added"

	// API actions
	ActionAPIKeyCreated AuditAction = "apikey.created"
	ActionAPIKeyRevoked AuditAction = "apikey.revoked"

	// Auth actions
	ActionLoginSuccess AuditAction = "auth.login_success"
	ActionLoginFailed  AuditAction = "auth.login_failed"
	ActionLogout       AuditAction = "auth.logout"
	ActionPasswordChanged AuditAction = "auth.password_changed"
	Action2FAEnabled   AuditAction = "auth.2fa_enabled"
	Action2FADisabled  AuditAction = "auth.2fa_disabled"
)

// AuditLogger handles audit logging
type AuditLogger struct {
	repo     *db.AuditRepository
	userRepo *db.UserRepository
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(repo *db.AuditRepository, userRepo *db.UserRepository) *AuditLogger {
	return &AuditLogger{
		repo:     repo,
		userRepo: userRepo,
	}
}

// Log creates an audit log entry
func (al *AuditLogger) Log(ctx context.Context, event AuditEvent) error {
	// Get actor email
	if event.ActorID != "" && event.ActorEmail == "" {
		if user, err := al.userRepo.GetByID(ctx, event.ActorID); err == nil {
			event.ActorEmail = user.Email
		}
	}

	// Get IP and User-Agent from context if available
	if ip, ok := ctx.Value("ip_address").(string); ok {
		event.IPAddress = ip
	}
	if ua, ok := ctx.Value("user_agent").(string); ok {
		event.UserAgent = ua
	}

	event.CreatedAt = time.Now()

	return al.repo.Create(ctx, &event)
}

// Query retrieves audit logs with filtering
func (al *AuditLogger) Query(ctx context.Context, filter AuditFilter) (*AuditQueryResult, error) {
	return al.repo.Query(ctx, filter)
}

// AuditFilter defines filters for querying audit logs
type AuditFilter struct {
	WorkspaceID  string    `json:"workspace_id"`
	ActorID      string    `json:"actor_id,omitempty"`
	Action       string    `json:"action,omitempty"`
	ResourceType string    `json:"resource_type,omitempty"`
	ResourceID   string    `json:"resource_id,omitempty"`
	StartDate    time.Time `json:"start_date,omitempty"`
	EndDate      time.Time `json:"end_date,omitempty"`
	Limit        int       `json:"limit"`
	Offset       int       `json:"offset"`
}

// AuditQueryResult contains paginated audit results
type AuditQueryResult struct {
	Events []AuditEvent `json:"events"`
	Total  int64        `json:"total"`
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
}

// AuditExporter exports audit logs
type AuditExporter struct {
	logger *AuditLogger
}

// ExportCSV exports audit logs to CSV format
func (ae *AuditExporter) ExportCSV(ctx context.Context, filter AuditFilter) ([]byte, error) {
	// Get all matching events
	filter.Limit = 10000 // Max export size
	result, err := ae.logger.Query(ctx, filter)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Header
	writer.Write([]string{
		"Timestamp",
		"Actor",
		"Action",
		"Resource Type",
		"Resource ID",
		"IP Address",
		"Details",
	})

	// Data rows
	for _, event := range result.Events {
		metadata, _ := json.Marshal(event.Metadata)
		writer.Write([]string{
			event.CreatedAt.Format(time.RFC3339),
			event.ActorEmail,
			event.Action,
			event.ResourceType,
			event.ResourceID,
			event.IPAddress,
			string(metadata),
		})
	}

	writer.Flush()
	return buf.Bytes(), nil
}
```

---

## API Endpoints

```go
// internal/api/handlers/workspace.go
package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/link-rift/link-rift/internal/workspace"
)

// WorkspaceHandler handles workspace API requests
type WorkspaceHandler struct {
	workspaceSvc  *workspace.WorkspaceService
	invitationSvc *workspace.InvitationService
	auditLogger   *workspace.AuditLogger
	rbac          *workspace.RBAC
}

// RegisterRoutes registers workspace routes
func (h *WorkspaceHandler) RegisterRoutes(app *fiber.App) {
	ws := app.Group("/api/v1/workspace")

	// Workspace
	ws.Get("/", h.GetWorkspace)
	ws.Put("/", h.UpdateWorkspace)
	ws.Delete("/", h.DeleteWorkspace)
	ws.Put("/settings", h.UpdateSettings)

	// Members
	ws.Get("/members", h.ListMembers)
	ws.Delete("/members/:id", h.RemoveMember)
	ws.Put("/members/:id/role", h.UpdateMemberRole)

	// Invitations
	ws.Get("/invitations", h.ListInvitations)
	ws.Post("/invitations", h.CreateInvitation)
	ws.Delete("/invitations/:id", h.RevokeInvitation)

	// Audit logs
	ws.Get("/audit", h.GetAuditLogs)
	ws.Get("/audit/export", h.ExportAuditLogs)

	// Accept invitation (public route)
	app.Get("/invite/:token", h.GetInvitationDetails)
	app.Post("/invite/:token/accept", h.AcceptInvitation)
}

// CreateInvitation sends a workspace invitation
// @Summary Invite a member
// @Tags Workspace
// @Accept json
// @Produce json
// @Param body body CreateInvitationRequest true "Invitation details"
// @Success 201 {object} InvitationResponse
// @Router /api/v1/workspace/invitations [post]
func (h *WorkspaceHandler) CreateInvitation(c *fiber.Ctx) error {
	var req CreateInvitationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	workspaceID := c.Locals("workspaceID").(string)
	userID := c.Locals("userID").(string)
	userRole := c.Locals("userRole").(workspace.Role)

	// Check permission
	if !h.rbac.HasPermission(userRole, workspace.PermMemberInvite) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Insufficient permissions",
		})
	}

	invitation, err := h.invitationSvc.CreateInvitation(
		c.Context(),
		workspaceID,
		userID,
		userRole,
		req.Email,
		workspace.Role(req.Role),
	)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(invitation)
}

// UpdateMemberRole changes a member's role
// @Summary Update member role
// @Tags Workspace
// @Accept json
// @Produce json
// @Param id path string true "Member ID"
// @Param body body UpdateRoleRequest true "New role"
// @Success 200 {object} MemberResponse
// @Router /api/v1/workspace/members/{id}/role [put]
func (h *WorkspaceHandler) UpdateMemberRole(c *fiber.Ctx) error {
	memberID := c.Params("id")

	var req UpdateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	workspaceID := c.Locals("workspaceID").(string)
	actorID := c.Locals("userID").(string)
	actorRole := c.Locals("userRole").(workspace.Role)

	// Check permission
	if !h.rbac.HasPermission(actorRole, workspace.PermMemberRole) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Insufficient permissions",
		})
	}

	newRole := workspace.Role(req.Role)

	// Can't change own role
	member, _ := h.workspaceSvc.GetMember(c.Context(), memberID)
	if member.UserID == actorID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot change your own role",
		})
	}

	// Check if actor can manage this role
	if !h.rbac.CanManageRole(actorRole, member.Role) || !h.rbac.CanManageRole(actorRole, newRole) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Cannot assign this role",
		})
	}

	member, err := h.workspaceSvc.UpdateMemberRole(c.Context(), workspaceID, memberID, newRole, actorID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(member)
}

// GetAuditLogs retrieves audit logs
// @Summary Get audit logs
// @Tags Workspace
// @Produce json
// @Param action query string false "Filter by action"
// @Param resource_type query string false "Filter by resource type"
// @Param start_date query string false "Start date (ISO 8601)"
// @Param end_date query string false "End date (ISO 8601)"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} AuditQueryResult
// @Router /api/v1/workspace/audit [get]
func (h *WorkspaceHandler) GetAuditLogs(c *fiber.Ctx) error {
	workspaceID := c.Locals("workspaceID").(string)
	userRole := c.Locals("userRole").(workspace.Role)

	// Check permission
	if !h.rbac.HasPermission(userRole, workspace.PermAuditView) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Insufficient permissions",
		})
	}

	filter := workspace.AuditFilter{
		WorkspaceID:  workspaceID,
		Action:       c.Query("action"),
		ResourceType: c.Query("resource_type"),
		Limit:        c.QueryInt("limit", 50),
		Offset:       c.QueryInt("offset", 0),
	}

	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			filter.StartDate = t
		}
	}

	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			filter.EndDate = t
		}
	}

	result, err := h.auditLogger.Query(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve audit logs",
		})
	}

	return c.JSON(result)
}

// Request types
type CreateInvitationRequest struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=admin editor viewer"`
}

type UpdateRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin editor viewer"`
}
```

---

## React Components

### Team Management

```typescript
// src/components/workspace/TeamManager.tsx
import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { workspaceApi, Member, Invitation, Role } from '@/api/workspace';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { MoreVertical, Mail, UserPlus, Crown, Shield, Pencil, Eye } from 'lucide-react';
import { useAuth } from '@/contexts/AuthContext';

export const TeamManager: React.FC = () => {
  const { user } = useAuth();
  const queryClient = useQueryClient();
  const [isInviteOpen, setIsInviteOpen] = useState(false);

  const { data: members } = useQuery({
    queryKey: ['members'],
    queryFn: workspaceApi.listMembers,
  });

  const { data: invitations } = useQuery({
    queryKey: ['invitations'],
    queryFn: workspaceApi.listInvitations,
  });

  const currentUserRole = members?.find((m) => m.user_id === user?.id)?.role || 'viewer';
  const canManageMembers = ['owner', 'admin'].includes(currentUserRole);

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold">Team Members</h2>
        {canManageMembers && (
          <Dialog open={isInviteOpen} onOpenChange={setIsInviteOpen}>
            <DialogTrigger asChild>
              <Button>
                <UserPlus className="w-4 h-4 mr-2" />
                Invite Member
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Invite Team Member</DialogTitle>
              </DialogHeader>
              <InviteForm
                onSuccess={() => {
                  setIsInviteOpen(false);
                  queryClient.invalidateQueries({ queryKey: ['invitations'] });
                }}
                currentUserRole={currentUserRole as Role}
              />
            </DialogContent>
          </Dialog>
        )}
      </div>

      {/* Active Members */}
      <Card>
        <CardHeader>
          <CardTitle>Active Members ({members?.length || 0})</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {members?.map((member) => (
              <MemberRow
                key={member.id}
                member={member}
                currentUserId={user?.id || ''}
                currentUserRole={currentUserRole as Role}
                canManage={canManageMembers}
              />
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Pending Invitations */}
      {invitations && invitations.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Pending Invitations ({invitations.length})</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {invitations.map((invitation) => (
                <InvitationRow
                  key={invitation.id}
                  invitation={invitation}
                  canManage={canManageMembers}
                />
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
};

// Member Row Component
interface MemberRowProps {
  member: Member;
  currentUserId: string;
  currentUserRole: Role;
  canManage: boolean;
}

const MemberRow: React.FC<MemberRowProps> = ({
  member,
  currentUserId,
  currentUserRole,
  canManage,
}) => {
  const queryClient = useQueryClient();

  const updateRoleMutation = useMutation({
    mutationFn: (newRole: Role) => workspaceApi.updateMemberRole(member.id, newRole),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['members'] }),
  });

  const removeMutation = useMutation({
    mutationFn: () => workspaceApi.removeMember(member.id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['members'] }),
  });

  const isCurrentUser = member.user_id === currentUserId;
  const canEdit = canManage && !isCurrentUser && canManageRole(currentUserRole, member.role);

  const roleIcons: Record<Role, JSX.Element> = {
    owner: <Crown className="w-4 h-4 text-yellow-500" />,
    admin: <Shield className="w-4 h-4 text-blue-500" />,
    editor: <Pencil className="w-4 h-4 text-green-500" />,
    viewer: <Eye className="w-4 h-4 text-gray-500" />,
  };

  return (
    <div className="flex items-center justify-between p-2 hover:bg-muted rounded-lg">
      <div className="flex items-center gap-3">
        <Avatar>
          <AvatarImage src={member.user?.avatar_url} />
          <AvatarFallback>
            {member.user?.name?.charAt(0) || member.user?.email?.charAt(0)}
          </AvatarFallback>
        </Avatar>
        <div>
          <div className="flex items-center gap-2">
            <span className="font-medium">{member.user?.name || 'Unknown'}</span>
            {isCurrentUser && (
              <Badge variant="outline" className="text-xs">
                You
              </Badge>
            )}
          </div>
          <span className="text-sm text-muted-foreground">{member.user?.email}</span>
        </div>
      </div>

      <div className="flex items-center gap-2">
        <div className="flex items-center gap-1">
          {roleIcons[member.role]}
          <span className="text-sm capitalize">{member.role}</span>
        </div>

        {canEdit && (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="sm">
                <MoreVertical className="w-4 h-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem disabled className="text-xs text-muted-foreground">
                Change Role
              </DropdownMenuItem>
              {getAssignableRoles(currentUserRole).map((role) => (
                <DropdownMenuItem
                  key={role}
                  onClick={() => updateRoleMutation.mutate(role)}
                  disabled={role === member.role}
                >
                  <span className="capitalize">{role}</span>
                </DropdownMenuItem>
              ))}
              <DropdownMenuSeparator />
              <DropdownMenuItem
                className="text-destructive"
                onClick={() => {
                  if (confirm('Are you sure you want to remove this member?')) {
                    removeMutation.mutate();
                  }
                }}
              >
                Remove from team
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        )}
      </div>
    </div>
  );
};

// Invitation Row Component
interface InvitationRowProps {
  invitation: Invitation;
  canManage: boolean;
}

const InvitationRow: React.FC<InvitationRowProps> = ({ invitation, canManage }) => {
  const queryClient = useQueryClient();

  const revokeMutation = useMutation({
    mutationFn: () => workspaceApi.revokeInvitation(invitation.id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['invitations'] }),
  });

  const resendMutation = useMutation({
    mutationFn: () => workspaceApi.resendInvitation(invitation.id),
  });

  return (
    <div className="flex items-center justify-between p-2 bg-muted/50 rounded-lg">
      <div className="flex items-center gap-3">
        <div className="w-10 h-10 rounded-full bg-muted flex items-center justify-center">
          <Mail className="w-5 h-5 text-muted-foreground" />
        </div>
        <div>
          <span className="font-medium">{invitation.email}</span>
          <div className="flex items-center gap-2">
            <Badge variant="outline" className="text-xs capitalize">
              {invitation.role}
            </Badge>
            <span className="text-xs text-muted-foreground">
              Expires {new Date(invitation.expires_at).toLocaleDateString()}
            </span>
          </div>
        </div>
      </div>

      {canManage && (
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => resendMutation.mutate()}
            disabled={resendMutation.isPending}
          >
            Resend
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => revokeMutation.mutate()}
            disabled={revokeMutation.isPending}
          >
            Revoke
          </Button>
        </div>
      )}
    </div>
  );
};

// Invite Form Component
interface InviteFormProps {
  onSuccess: () => void;
  currentUserRole: Role;
}

const InviteForm: React.FC<InviteFormProps> = ({ onSuccess, currentUserRole }) => {
  const [email, setEmail] = useState('');
  const [role, setRole] = useState<Role>('editor');

  const inviteMutation = useMutation({
    mutationFn: workspaceApi.createInvitation,
    onSuccess,
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    inviteMutation.mutate({ email, role });
  };

  const assignableRoles = getAssignableRoles(currentUserRole);

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <label className="text-sm font-medium">Email Address</label>
        <Input
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="colleague@company.com"
          required
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium">Role</label>
        <Select value={role} onValueChange={(v) => setRole(v as Role)}>
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {assignableRoles.map((r) => (
              <SelectItem key={r} value={r}>
                <div className="flex flex-col">
                  <span className="capitalize">{r}</span>
                  <span className="text-xs text-muted-foreground">
                    {getRoleDescription(r)}
                  </span>
                </div>
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {inviteMutation.error && (
        <p className="text-sm text-destructive">
          {(inviteMutation.error as Error).message}
        </p>
      )}

      <Button
        type="submit"
        className="w-full"
        disabled={inviteMutation.isPending}
      >
        {inviteMutation.isPending ? 'Sending...' : 'Send Invitation'}
      </Button>
    </form>
  );
};

// Helper functions
function canManageRole(actorRole: Role, targetRole: Role): boolean {
  const hierarchy: Record<Role, number> = {
    owner: 4,
    admin: 3,
    editor: 2,
    viewer: 1,
  };
  return hierarchy[actorRole] > hierarchy[targetRole];
}

function getAssignableRoles(role: Role): Role[] {
  switch (role) {
    case 'owner':
      return ['admin', 'editor', 'viewer'];
    case 'admin':
      return ['editor', 'viewer'];
    default:
      return [];
  }
}

function getRoleDescription(role: Role): string {
  switch (role) {
    case 'admin':
      return 'Can manage team, settings, and all resources';
    case 'editor':
      return 'Can create and edit links, pages, and QR codes';
    case 'viewer':
      return 'Can view links and analytics only';
    default:
      return '';
  }
}
```

### Audit Log Viewer

```typescript
// src/components/workspace/AuditLog.tsx
import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { workspaceApi, AuditEvent, AuditFilter } from '@/api/workspace';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import { Download, Filter, ChevronLeft, ChevronRight } from 'lucide-react';

export const AuditLog: React.FC = () => {
  const [filter, setFilter] = useState<AuditFilter>({
    limit: 50,
    offset: 0,
  });

  const { data, isLoading } = useQuery({
    queryKey: ['audit-logs', filter],
    queryFn: () => workspaceApi.getAuditLogs(filter),
  });

  const handleExport = async () => {
    const blob = await workspaceApi.exportAuditLogs(filter);
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `audit-log-${new Date().toISOString().split('T')[0]}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const totalPages = Math.ceil((data?.total || 0) / filter.limit);
  const currentPage = Math.floor(filter.offset / filter.limit) + 1;

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold">Audit Log</h2>
        <Button variant="outline" onClick={handleExport}>
          <Download className="w-4 h-4 mr-2" />
          Export CSV
        </Button>
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex flex-wrap gap-4">
            <Select
              value={filter.action || 'all'}
              onValueChange={(v) =>
                setFilter({ ...filter, action: v === 'all' ? undefined : v, offset: 0 })
              }
            >
              <SelectTrigger className="w-48">
                <SelectValue placeholder="Filter by action" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Actions</SelectItem>
                <SelectItem value="link.created">Link Created</SelectItem>
                <SelectItem value="link.updated">Link Updated</SelectItem>
                <SelectItem value="link.deleted">Link Deleted</SelectItem>
                <SelectItem value="member.invited">Member Invited</SelectItem>
                <SelectItem value="member.joined">Member Joined</SelectItem>
                <SelectItem value="member.removed">Member Removed</SelectItem>
                <SelectItem value="domain.added">Domain Added</SelectItem>
                <SelectItem value="domain.verified">Domain Verified</SelectItem>
              </SelectContent>
            </Select>

            <Select
              value={filter.resource_type || 'all'}
              onValueChange={(v) =>
                setFilter({ ...filter, resource_type: v === 'all' ? undefined : v, offset: 0 })
              }
            >
              <SelectTrigger className="w-48">
                <SelectValue placeholder="Filter by resource" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Resources</SelectItem>
                <SelectItem value="link">Links</SelectItem>
                <SelectItem value="member">Members</SelectItem>
                <SelectItem value="invitation">Invitations</SelectItem>
                <SelectItem value="domain">Domains</SelectItem>
                <SelectItem value="biopage">Bio Pages</SelectItem>
                <SelectItem value="webhook">Webhooks</SelectItem>
              </SelectContent>
            </Select>

            <Input
              type="date"
              placeholder="Start date"
              onChange={(e) =>
                setFilter({
                  ...filter,
                  start_date: e.target.value ? new Date(e.target.value).toISOString() : undefined,
                  offset: 0,
                })
              }
              className="w-40"
            />

            <Input
              type="date"
              placeholder="End date"
              onChange={(e) =>
                setFilter({
                  ...filter,
                  end_date: e.target.value ? new Date(e.target.value).toISOString() : undefined,
                  offset: 0,
                })
              }
              className="w-40"
            />
          </div>
        </CardContent>
      </Card>

      {/* Audit Events */}
      <Card>
        <CardContent className="pt-6">
          {isLoading ? (
            <div>Loading...</div>
          ) : (
            <div className="space-y-2">
              {data?.events.map((event) => (
                <AuditEventRow key={event.id} event={event} />
              ))}

              {data?.events.length === 0 && (
                <div className="text-center py-8 text-muted-foreground">
                  No audit events found
                </div>
              )}
            </div>
          )}

          {/* Pagination */}
          {data && data.total > filter.limit && (
            <div className="flex items-center justify-between mt-4 pt-4 border-t">
              <span className="text-sm text-muted-foreground">
                Showing {filter.offset + 1} - {Math.min(filter.offset + filter.limit, data.total)} of{' '}
                {data.total}
              </span>
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setFilter({ ...filter, offset: filter.offset - filter.limit })}
                  disabled={filter.offset === 0}
                >
                  <ChevronLeft className="w-4 h-4" />
                </Button>
                <span className="text-sm">
                  Page {currentPage} of {totalPages}
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setFilter({ ...filter, offset: filter.offset + filter.limit })}
                  disabled={filter.offset + filter.limit >= data.total}
                >
                  <ChevronRight className="w-4 h-4" />
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
};

// Audit Event Row
interface AuditEventRowProps {
  event: AuditEvent;
}

const AuditEventRow: React.FC<AuditEventRowProps> = ({ event }) => {
  const [expanded, setExpanded] = useState(false);

  const actionColors: Record<string, string> = {
    created: 'bg-green-100 text-green-800',
    updated: 'bg-blue-100 text-blue-800',
    deleted: 'bg-red-100 text-red-800',
    invited: 'bg-purple-100 text-purple-800',
    joined: 'bg-green-100 text-green-800',
    removed: 'bg-red-100 text-red-800',
    verified: 'bg-green-100 text-green-800',
  };

  const getActionColor = (action: string) => {
    for (const [key, value] of Object.entries(actionColors)) {
      if (action.includes(key)) return value;
    }
    return 'bg-gray-100 text-gray-800';
  };

  return (
    <div
      className="p-3 border rounded-lg hover:bg-muted/50 cursor-pointer"
      onClick={() => setExpanded(!expanded)}
    >
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Badge className={getActionColor(event.action)}>
            {event.action.replace('.', ' ')}
          </Badge>
          <span className="text-sm">
            <span className="font-medium">{event.actor_email}</span>
            {' performed action on '}
            <span className="font-medium">{event.resource_type}</span>
          </span>
        </div>
        <span className="text-sm text-muted-foreground">
          {new Date(event.created_at).toLocaleString()}
        </span>
      </div>

      {expanded && event.metadata && (
        <div className="mt-3 p-2 bg-muted rounded text-xs font-mono">
          <pre>{JSON.stringify(event.metadata, null, 2)}</pre>
        </div>
      )}

      {expanded && event.ip_address && (
        <div className="mt-2 text-xs text-muted-foreground">
          IP: {event.ip_address}
        </div>
      )}
    </div>
  );
};
```

### Workspace API Client

```typescript
// src/api/workspace.ts
import { apiClient } from './client';

export type Role = 'owner' | 'admin' | 'editor' | 'viewer';

export interface Member {
  id: string;
  workspace_id: string;
  user_id: string;
  role: Role;
  joined_at: string;
  user?: {
    id: string;
    email: string;
    name: string;
    avatar_url?: string;
  };
}

export interface Invitation {
  id: string;
  email: string;
  role: Role;
  invited_by: string;
  expires_at: string;
  created_at: string;
}

export interface AuditEvent {
  id: string;
  actor_id: string;
  actor_email: string;
  action: string;
  resource_type: string;
  resource_id: string;
  metadata?: Record<string, any>;
  ip_address?: string;
  created_at: string;
}

export interface AuditFilter {
  action?: string;
  resource_type?: string;
  start_date?: string;
  end_date?: string;
  limit: number;
  offset: number;
}

export const workspaceApi = {
  // Members
  listMembers: async (): Promise<Member[]> => {
    const response = await apiClient.get<Member[]>('/api/v1/workspace/members');
    return response.data;
  },

  updateMemberRole: async (memberId: string, role: Role): Promise<Member> => {
    const response = await apiClient.put<Member>(
      `/api/v1/workspace/members/${memberId}/role`,
      { role }
    );
    return response.data;
  },

  removeMember: async (memberId: string): Promise<void> => {
    await apiClient.delete(`/api/v1/workspace/members/${memberId}`);
  },

  // Invitations
  listInvitations: async (): Promise<Invitation[]> => {
    const response = await apiClient.get<Invitation[]>('/api/v1/workspace/invitations');
    return response.data;
  },

  createInvitation: async (data: { email: string; role: Role }): Promise<Invitation> => {
    const response = await apiClient.post<Invitation>(
      '/api/v1/workspace/invitations',
      data
    );
    return response.data;
  },

  revokeInvitation: async (invitationId: string): Promise<void> => {
    await apiClient.delete(`/api/v1/workspace/invitations/${invitationId}`);
  },

  resendInvitation: async (invitationId: string): Promise<void> => {
    await apiClient.post(`/api/v1/workspace/invitations/${invitationId}/resend`);
  },

  // Audit logs
  getAuditLogs: async (
    filter: AuditFilter
  ): Promise<{ events: AuditEvent[]; total: number }> => {
    const params = new URLSearchParams();
    if (filter.action) params.set('action', filter.action);
    if (filter.resource_type) params.set('resource_type', filter.resource_type);
    if (filter.start_date) params.set('start_date', filter.start_date);
    if (filter.end_date) params.set('end_date', filter.end_date);
    params.set('limit', filter.limit.toString());
    params.set('offset', filter.offset.toString());

    const response = await apiClient.get(`/api/v1/workspace/audit?${params}`);
    return response.data;
  },

  exportAuditLogs: async (filter: Omit<AuditFilter, 'limit' | 'offset'>): Promise<Blob> => {
    const params = new URLSearchParams();
    if (filter.action) params.set('action', filter.action);
    if (filter.resource_type) params.set('resource_type', filter.resource_type);
    if (filter.start_date) params.set('start_date', filter.start_date);
    if (filter.end_date) params.set('end_date', filter.end_date);

    const response = await apiClient.get(`/api/v1/workspace/audit/export?${params}`, {
      responseType: 'blob',
    });
    return response.data;
  },
};
```
