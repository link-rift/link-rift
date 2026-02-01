package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/link-rift/link-rift/internal/ee/scim"
	"github.com/link-rift/link-rift/internal/license"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/pkg/crypto"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

// SCIMService provides SCIM 2.0 provisioning operations.
type SCIMService interface {
	// Token management (admin)
	CreateToken(ctx context.Context, workspaceID uuid.UUID, input models.CreateSCIMTokenInput) (*models.CreateSCIMTokenResponse, error)
	ListTokens(ctx context.Context, workspaceID uuid.UUID) ([]*models.SCIMToken, error)
	RevokeToken(ctx context.Context, workspaceID uuid.UUID, tokenID uuid.UUID) error
	ListSyncLogs(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]*models.SCIMSyncLog, int64, error)

	// Token validation (middleware)
	ValidateToken(tokenHash string) (uuid.UUID, uuid.UUID, error)
	RecordTokenUsage(tokenID uuid.UUID)

	// SCIM 2.0 User operations
	ListUsers(ctx context.Context, workspaceID uuid.UUID, filter string, startIndex, count int) (*scim.SCIMListResponse, error)
	GetUser(ctx context.Context, workspaceID uuid.UUID, userID string) (*scim.SCIMUser, error)
	CreateUser(ctx context.Context, workspaceID uuid.UUID, user scim.SCIMUser) (*scim.SCIMUser, error)
	UpdateUser(ctx context.Context, workspaceID uuid.UUID, userID string, user scim.SCIMUser) (*scim.SCIMUser, error)
	PatchUser(ctx context.Context, workspaceID uuid.UUID, userID string, patch scim.SCIMPatchOp) (*scim.SCIMUser, error)
	DeleteUser(ctx context.Context, workspaceID uuid.UUID, userID string) error

	SetAuditLogger(logger AuditLogger)
}

type scimService struct {
	scimRepo    repository.SCIMRepository
	userRepo    repository.UserRepository
	memberRepo  repository.WorkspaceMemberRepository
	licManager  *license.Manager
	auditLogger AuditLogger
	logger      *zap.Logger
}

// NewSCIMService creates a new SCIM service.
func NewSCIMService(
	scimRepo repository.SCIMRepository,
	userRepo repository.UserRepository,
	memberRepo repository.WorkspaceMemberRepository,
	licManager *license.Manager,
	logger *zap.Logger,
) SCIMService {
	return &scimService{
		scimRepo:    scimRepo,
		userRepo:    userRepo,
		memberRepo:  memberRepo,
		licManager:  licManager,
		auditLogger: noopAuditLogger{},
		logger:      logger,
	}
}

func (s *scimService) SetAuditLogger(logger AuditLogger) {
	if logger != nil {
		s.auditLogger = logger
	}
}

func (s *scimService) requireFeature() error {
	if !s.licManager.HasFeature(license.FeatureSCIM) {
		return httputil.PaymentRequiredWithDetails(string(license.FeatureSCIM), "enterprise")
	}
	return nil
}

// CreateToken generates a new SCIM bearer token for a workspace.
func (s *scimService) CreateToken(ctx context.Context, workspaceID uuid.UUID, input models.CreateSCIMTokenInput) (*models.CreateSCIMTokenResponse, error) {
	if err := s.requireFeature(); err != nil {
		return nil, err
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, httputil.Wrap(err, "failed to generate token")
	}
	rawToken := "scim_" + hex.EncodeToString(tokenBytes)
	tokenHash := scim.HashToken(rawToken)
	tokenPrefix := rawToken[:12]

	var expiresAt pgtype.Timestamptz
	if input.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *input.ExpiresAt)
		if err != nil {
			return nil, httputil.Validation("expires_at", "must be a valid RFC3339 date")
		}
		expiresAt = pgtype.Timestamptz{Time: t, Valid: true}
	}

	token, err := s.scimRepo.CreateToken(ctx, sqlc.CreateSCIMTokenParams{
		WorkspaceID: workspaceID,
		TokenHash:   tokenHash,
		TokenPrefix: tokenPrefix,
		Name:        input.Name,
		ExpiresAt:   expiresAt,
	})
	if err != nil {
		return nil, err
	}

	s.auditLogger.Log(ctx, AuditEntry{
		WorkspaceID:  workspaceID,
		Action:       "create",
		ResourceType: "scim_token",
		ResourceID:   token.ID,
		NewValues:    map[string]interface{}{"name": input.Name, "prefix": tokenPrefix},
	})

	return &models.CreateSCIMTokenResponse{
		Token:     rawToken,
		SCIMToken: token,
	}, nil
}

func (s *scimService) ListTokens(ctx context.Context, workspaceID uuid.UUID) ([]*models.SCIMToken, error) {
	if err := s.requireFeature(); err != nil {
		return nil, err
	}
	return s.scimRepo.ListTokens(ctx, workspaceID)
}

func (s *scimService) RevokeToken(ctx context.Context, workspaceID uuid.UUID, tokenID uuid.UUID) error {
	if err := s.requireFeature(); err != nil {
		return err
	}
	if err := s.scimRepo.RevokeToken(ctx, tokenID); err != nil {
		return err
	}

	s.auditLogger.Log(ctx, AuditEntry{
		WorkspaceID:  workspaceID,
		Action:       "revoke",
		ResourceType: "scim_token",
		ResourceID:   tokenID,
	})
	return nil
}

func (s *scimService) ListSyncLogs(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]*models.SCIMSyncLog, int64, error) {
	if err := s.requireFeature(); err != nil {
		return nil, 0, err
	}
	return s.scimRepo.ListSyncLogs(ctx, sqlc.ListSCIMSyncLogsParams{
		WorkspaceID: workspaceID,
		Limit:       int32(limit),
		Offset:      int32(offset),
	})
}

// ValidateToken validates a SCIM bearer token hash and returns workspace/token IDs.
func (s *scimService) ValidateToken(tokenHash string) (uuid.UUID, uuid.UUID, error) {
	token, err := s.scimRepo.GetTokenByHash(context.Background(), tokenHash)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	if !token.IsActive {
		return uuid.Nil, uuid.Nil, fmt.Errorf("token is inactive")
	}
	if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now()) {
		return uuid.Nil, uuid.Nil, fmt.Errorf("token has expired")
	}
	return token.WorkspaceID, token.ID, nil
}

// RecordTokenUsage updates the last_used_at timestamp (best-effort).
func (s *scimService) RecordTokenUsage(tokenID uuid.UUID) {
	if err := s.scimRepo.UpdateTokenLastUsed(context.Background(), tokenID); err != nil {
		s.logger.Warn("failed to update SCIM token last used", zap.Error(err))
	}
}

func (s *scimService) logSync(ctx context.Context, wsID uuid.UUID, operation, resourceType, externalID, status string, details interface{}) {
	detailsJSON, _ := json.Marshal(details)
	_ = s.scimRepo.CreateSyncLog(ctx, sqlc.CreateSCIMSyncLogParams{
		WorkspaceID:  wsID,
		Operation:    operation,
		ResourceType: resourceType,
		ExternalID:   pgtype.Text{String: externalID, Valid: externalID != ""},
		Status:       status,
		Details:      detailsJSON,
	})
}

// ListUsers returns SCIM users (workspace members).
func (s *scimService) ListUsers(ctx context.Context, workspaceID uuid.UUID, filter string, startIndex, count int) (*scim.SCIMListResponse, error) {
	if err := s.requireFeature(); err != nil {
		return nil, err
	}

	members, err := s.memberRepo.List(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Apply filter if present
	var filtered []*models.WorkspaceMemberResponse
	if filter != "" {
		attr, op, value, parseErr := scim.ParseFilter(filter)
		if parseErr != nil {
			return nil, httputil.Validation("filter", parseErr.Error())
		}
		if op != "eq" {
			return nil, httputil.Validation("filter", "only 'eq' operator is supported")
		}
		for _, m := range members {
			match := false
			switch attr {
			case "userName":
				match = m.Email == value
			case "externalId":
				match = m.UserID.String() == value
			}
			if match {
				filtered = append(filtered, m)
			}
		}
	} else {
		filtered = members
	}

	// Pagination (SCIM is 1-indexed)
	total := len(filtered)
	if startIndex < 1 {
		startIndex = 1
	}
	if count < 1 {
		count = 100
	}
	start := startIndex - 1
	end := start + count
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	page := filtered[start:end]

	users := make([]scim.SCIMUser, 0, len(page))
	for _, m := range page {
		users = append(users, memberResponseToSCIMUser(m))
	}

	usersJSON, _ := json.Marshal(users)
	return &scim.SCIMListResponse{
		Schemas:      []string{scim.SchemaListResponse},
		TotalResults: total,
		ItemsPerPage: count,
		StartIndex:   startIndex,
		Resources:    usersJSON,
	}, nil
}

// GetUser returns a single SCIM user by user ID.
func (s *scimService) GetUser(ctx context.Context, workspaceID uuid.UUID, userID string) (*scim.SCIMUser, error) {
	if err := s.requireFeature(); err != nil {
		return nil, err
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, httputil.NotFound("user")
	}

	user, err := s.userRepo.GetByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	// Verify user is a member of this workspace
	_, err = s.memberRepo.Get(ctx, workspaceID, uid)
	if err != nil {
		return nil, httputil.NotFound("user")
	}

	return userToSCIMUser(user), nil
}

// CreateUser provisions a new user via SCIM.
func (s *scimService) CreateUser(ctx context.Context, workspaceID uuid.UUID, scimUser scim.SCIMUser) (*scim.SCIMUser, error) {
	if err := s.requireFeature(); err != nil {
		return nil, err
	}

	email := scimUser.UserName
	if len(scimUser.Emails) > 0 {
		for _, e := range scimUser.Emails {
			if e.Primary {
				email = e.Value
				break
			}
		}
		if email == "" {
			email = scimUser.Emails[0].Value
		}
	}
	if email == "" {
		return nil, httputil.Validation("userName", "email is required")
	}

	name := email
	if scimUser.Name != nil {
		if scimUser.Name.Formatted != "" {
			name = scimUser.Name.Formatted
		} else if scimUser.Name.GivenName != "" || scimUser.Name.FamilyName != "" {
			name = scimUser.Name.GivenName
			if scimUser.Name.FamilyName != "" {
				if name != "" {
					name += " "
				}
				name += scimUser.Name.FamilyName
			}
		}
	}

	// Check if user already exists
	existingUser, _ := s.userRepo.GetByEmail(ctx, email)
	if existingUser != nil {
		// Add existing user to workspace if not already a member
		_, memberErr := s.memberRepo.Get(ctx, workspaceID, existingUser.ID)
		if memberErr != nil {
			_, addErr := s.memberRepo.Add(ctx, sqlc.AddWorkspaceMemberParams{
				WorkspaceID: workspaceID,
				UserID:      existingUser.ID,
				Role:        "viewer",
			})
			if addErr != nil {
				s.logSync(ctx, workspaceID, "create", "User", scimUser.ExternalID, "error", map[string]string{"error": addErr.Error()})
				return nil, addErr
			}
		}

		result := userToSCIMUser(existingUser)
		result.ExternalID = scimUser.ExternalID
		s.logSync(ctx, workspaceID, "create", "User", scimUser.ExternalID, "success", map[string]string{"user_id": existingUser.ID.String(), "action": "added_to_workspace"})
		return result, nil
	}

	// Create new user with random password (they'll use SSO to authenticate)
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	passwordHash, err := crypto.HashPassword(hex.EncodeToString(randomBytes))
	if err != nil {
		return nil, httputil.Wrap(err, "failed to hash password")
	}

	newUser, err := s.userRepo.Create(ctx, sqlc.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
	})
	if err != nil {
		s.logSync(ctx, workspaceID, "create", "User", scimUser.ExternalID, "error", map[string]string{"error": err.Error()})
		return nil, err
	}

	// Add to workspace
	_, err = s.memberRepo.Add(ctx, sqlc.AddWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      newUser.ID,
		Role:        "viewer",
	})
	if err != nil {
		s.logSync(ctx, workspaceID, "create", "User", scimUser.ExternalID, "error", map[string]string{"error": err.Error()})
		return nil, err
	}

	result := userToSCIMUser(newUser)
	result.ExternalID = scimUser.ExternalID
	s.logSync(ctx, workspaceID, "create", "User", scimUser.ExternalID, "success", map[string]string{"user_id": newUser.ID.String(), "action": "created"})

	s.auditLogger.Log(ctx, AuditEntry{
		WorkspaceID:  workspaceID,
		Action:       "scim_create",
		ResourceType: "user",
		ResourceID:   newUser.ID,
		NewValues:    map[string]interface{}{"email": email, "name": name, "external_id": scimUser.ExternalID},
	})

	return result, nil
}

// UpdateUser replaces user attributes via SCIM PUT.
func (s *scimService) UpdateUser(ctx context.Context, workspaceID uuid.UUID, userID string, scimUser scim.SCIMUser) (*scim.SCIMUser, error) {
	if err := s.requireFeature(); err != nil {
		return nil, err
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, httputil.NotFound("user")
	}

	user, err := s.userRepo.GetByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	// Handle deactivation
	if !scimUser.Active {
		if err := s.memberRepo.Remove(ctx, workspaceID, uid); err != nil {
			s.logSync(ctx, workspaceID, "deactivate", "User", scimUser.ExternalID, "error", map[string]string{"error": err.Error()})
			return nil, err
		}
		s.logSync(ctx, workspaceID, "deactivate", "User", scimUser.ExternalID, "success", map[string]string{"user_id": userID})

		s.auditLogger.Log(ctx, AuditEntry{
			WorkspaceID:  workspaceID,
			Action:       "scim_deactivate",
			ResourceType: "user",
			ResourceID:   uid,
		})
	} else {
		s.logSync(ctx, workspaceID, "update", "User", scimUser.ExternalID, "success", map[string]string{"user_id": userID})
	}

	result := userToSCIMUser(user)
	result.ExternalID = scimUser.ExternalID
	result.Active = scimUser.Active
	return result, nil
}

// PatchUser applies a SCIM PATCH operation to a user.
func (s *scimService) PatchUser(ctx context.Context, workspaceID uuid.UUID, userID string, patch scim.SCIMPatchOp) (*scim.SCIMUser, error) {
	if err := s.requireFeature(); err != nil {
		return nil, err
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, httputil.NotFound("user")
	}

	user, err := s.userRepo.GetByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	active := true
	for _, op := range patch.Operations {
		switch op.Path {
		case "active":
			if v, ok := op.Value.(bool); ok {
				active = v
			}
		}
	}

	if !active {
		if err := s.memberRepo.Remove(ctx, workspaceID, uid); err != nil {
			s.logSync(ctx, workspaceID, "deactivate", "User", "", "error", map[string]string{"error": err.Error()})
			return nil, err
		}
		s.logSync(ctx, workspaceID, "deactivate", "User", "", "success", map[string]string{"user_id": userID})

		s.auditLogger.Log(ctx, AuditEntry{
			WorkspaceID:  workspaceID,
			Action:       "scim_deactivate",
			ResourceType: "user",
			ResourceID:   uid,
		})
	}

	result := userToSCIMUser(user)
	result.Active = active
	return result, nil
}

// DeleteUser removes a user from the workspace via SCIM.
func (s *scimService) DeleteUser(ctx context.Context, workspaceID uuid.UUID, userID string) error {
	if err := s.requireFeature(); err != nil {
		return err
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return httputil.NotFound("user")
	}

	if err := s.memberRepo.Remove(ctx, workspaceID, uid); err != nil {
		s.logSync(ctx, workspaceID, "delete", "User", "", "error", map[string]string{"error": err.Error()})
		return err
	}

	s.logSync(ctx, workspaceID, "delete", "User", "", "success", map[string]string{"user_id": userID})

	s.auditLogger.Log(ctx, AuditEntry{
		WorkspaceID:  workspaceID,
		Action:       "scim_delete",
		ResourceType: "user",
		ResourceID:   uid,
	})

	return nil
}

func memberResponseToSCIMUser(m *models.WorkspaceMemberResponse) scim.SCIMUser {
	return scim.SCIMUser{
		Schemas:  []string{scim.SchemaUser},
		ID:       m.UserID.String(),
		UserName: m.Email,
		Name: &scim.SCIMName{
			Formatted: m.Name,
		},
		Emails: []scim.SCIMEmail{{Value: m.Email, Primary: true}},
		Active: true,
		Meta: &scim.SCIMMeta{
			ResourceType: "User",
			Created:      scim.FormatTime(m.CreatedAt),
		},
	}
}

func userToSCIMUser(user *models.User) *scim.SCIMUser {
	return &scim.SCIMUser{
		Schemas:  []string{scim.SchemaUser},
		ID:       user.ID.String(),
		UserName: user.Email,
		Name: &scim.SCIMName{
			Formatted: user.Name,
		},
		Emails: []scim.SCIMEmail{{Value: user.Email, Primary: true}},
		Active: true,
		Meta: &scim.SCIMMeta{
			ResourceType: "User",
			Created:      scim.FormatTime(user.CreatedAt),
			LastModified: scim.FormatTime(user.UpdatedAt),
		},
	}
}
