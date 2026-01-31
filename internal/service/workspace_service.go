package service

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/link-rift/link-rift/internal/license"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/pkg/httputil"
	"go.uber.org/zap"
)

type WorkspaceService interface {
	CreateWorkspace(ctx context.Context, userID uuid.UUID, input models.CreateWorkspaceInput) (*models.Workspace, error)
	GetWorkspace(ctx context.Context, id uuid.UUID) (*models.Workspace, error)
	ListWorkspaces(ctx context.Context, userID uuid.UUID) ([]*models.Workspace, error)
	UpdateWorkspace(ctx context.Context, id uuid.UUID, input models.UpdateWorkspaceInput) (*models.Workspace, error)
	DeleteWorkspace(ctx context.Context, id uuid.UUID, actorID uuid.UUID) error

	InviteMember(ctx context.Context, workspaceID, inviterID uuid.UUID, input models.InviteMemberInput) (*models.WorkspaceMember, error)
	RemoveMember(ctx context.Context, workspaceID, actorID, targetUserID uuid.UUID) error
	UpdateMemberRole(ctx context.Context, workspaceID, actorID, targetUserID uuid.UUID, input models.UpdateMemberRoleInput) (*models.WorkspaceMember, error)
	TransferOwnership(ctx context.Context, workspaceID, actorID uuid.UUID, input models.TransferOwnershipInput) error
	ListMembers(ctx context.Context, workspaceID uuid.UUID) ([]*models.WorkspaceMemberResponse, error)
	GetMember(ctx context.Context, workspaceID, userID uuid.UUID) (*models.WorkspaceMember, error)
	GetMemberCount(ctx context.Context, workspaceID uuid.UUID) (int64, error)
}

type workspaceService struct {
	wsRepo     repository.WorkspaceRepository
	memberRepo repository.WorkspaceMemberRepository
	userRepo   repository.UserRepository
	licManager *license.Manager
	events     EventPublisher
	pool       *pgxpool.Pool
	logger     *zap.Logger
}

func NewWorkspaceService(
	wsRepo repository.WorkspaceRepository,
	memberRepo repository.WorkspaceMemberRepository,
	userRepo repository.UserRepository,
	licManager *license.Manager,
	events EventPublisher,
	pool *pgxpool.Pool,
	logger *zap.Logger,
) WorkspaceService {
	return &workspaceService{
		wsRepo:     wsRepo,
		memberRepo: memberRepo,
		userRepo:   userRepo,
		licManager: licManager,
		events:     events,
		pool:       pool,
		logger:     logger,
	}
}

func (s *workspaceService) CreateWorkspace(ctx context.Context, userID uuid.UUID, input models.CreateWorkspaceInput) (*models.Workspace, error) {
	// Check workspace limit
	count, err := s.wsRepo.GetCountForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !s.licManager.CheckLimit(license.LimitMaxWorkspaces, count) {
		return nil, httputil.PaymentRequired("workspace limit reached, upgrade your plan")
	}

	slug := strings.ToLower(strings.TrimSpace(input.Slug))

	// Use a transaction: create workspace + add owner as member
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback(ctx)

	qtx := sqlc.New(tx)
	txWsRepo := repository.NewWorkspaceRepository(qtx, s.logger)
	txMemberRepo := repository.NewWorkspaceMemberRepository(qtx, s.logger)

	ws, err := txWsRepo.Create(ctx, sqlc.CreateWorkspaceParams{
		Name:     strings.TrimSpace(input.Name),
		Slug:     slug,
		OwnerID:  userID,
		Plan:     string(s.licManager.GetTier()),
		Settings: json.RawMessage(`{}`),
	})
	if err != nil {
		return nil, err
	}

	_, err = txMemberRepo.Add(ctx, sqlc.AddWorkspaceMemberParams{
		WorkspaceID: ws.ID,
		UserID:      userID,
		Role:        string(models.RoleOwner),
		InvitedBy:   pgtype.UUID{},
	})
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, httputil.Wrap(err, "failed to commit transaction")
	}

	return ws, nil
}

func (s *workspaceService) GetWorkspace(ctx context.Context, id uuid.UUID) (*models.Workspace, error) {
	return s.wsRepo.GetByID(ctx, id)
}

func (s *workspaceService) ListWorkspaces(ctx context.Context, userID uuid.UUID) ([]*models.Workspace, error) {
	return s.wsRepo.ListForUser(ctx, userID)
}

func (s *workspaceService) UpdateWorkspace(ctx context.Context, id uuid.UUID, input models.UpdateWorkspaceInput) (*models.Workspace, error) {
	params := sqlc.UpdateWorkspaceParams{
		ID: id,
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		params.Name = pgtype.Text{String: name, Valid: true}
	}
	if input.Slug != nil {
		slug := strings.ToLower(strings.TrimSpace(*input.Slug))
		params.Slug = pgtype.Text{String: slug, Valid: true}
	}

	return s.wsRepo.Update(ctx, params)
}

func (s *workspaceService) DeleteWorkspace(ctx context.Context, id uuid.UUID, actorID uuid.UUID) error {
	ws, err := s.wsRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if ws.OwnerID != actorID {
		return httputil.Forbidden("only the workspace owner can delete the workspace")
	}

	return s.wsRepo.SoftDelete(ctx, id)
}

func (s *workspaceService) InviteMember(ctx context.Context, workspaceID, inviterID uuid.UUID, input models.InviteMemberInput) (*models.WorkspaceMember, error) {
	if !input.Role.IsValid() || input.Role == models.RoleOwner {
		return nil, httputil.Validation("role", "invalid role; must be admin, editor, or viewer")
	}

	// Check member limit
	memberCount, err := s.memberRepo.GetCount(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	if !s.licManager.CheckLimit(license.LimitMaxUsers, memberCount) {
		return nil, httputil.PaymentRequired("team member limit reached, upgrade your plan")
	}

	// Look up user by email
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, httputil.Validation("email", "user with this email not found")
	}

	member, err := s.memberRepo.Add(ctx, sqlc.AddWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      user.ID,
		Role:        string(input.Role),
		InvitedBy:   pgtype.UUID{Bytes: inviterID, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	// Publish webhook event (best-effort)
	if err := s.events.Publish(ctx, "team.member_invited", workspaceID, member); err != nil {
		s.logger.Warn("failed to publish team.member_invited event", zap.Error(err))
	}

	return member, nil
}

func (s *workspaceService) RemoveMember(ctx context.Context, workspaceID, actorID, targetUserID uuid.UUID) error {
	target, err := s.memberRepo.Get(ctx, workspaceID, targetUserID)
	if err != nil {
		return err
	}

	// Cannot remove the owner
	if target.Role == models.RoleOwner {
		return httputil.Forbidden("cannot remove the workspace owner; transfer ownership first")
	}

	// Admins cannot remove other admins (only owner can)
	if target.Role == models.RoleAdmin {
		actor, err := s.memberRepo.Get(ctx, workspaceID, actorID)
		if err != nil {
			return err
		}
		if actor.Role != models.RoleOwner {
			return httputil.Forbidden("only the workspace owner can remove admins")
		}
	}

	if err := s.memberRepo.Remove(ctx, workspaceID, targetUserID); err != nil {
		return err
	}

	// Publish webhook event (best-effort)
	if err := s.events.Publish(ctx, "team.member_removed", workspaceID, target); err != nil {
		s.logger.Warn("failed to publish team.member_removed event", zap.Error(err))
	}

	return nil
}

func (s *workspaceService) UpdateMemberRole(ctx context.Context, workspaceID, actorID, targetUserID uuid.UUID, input models.UpdateMemberRoleInput) (*models.WorkspaceMember, error) {
	if !input.Role.IsValid() {
		return nil, httputil.Validation("role", "invalid role")
	}

	// Cannot change to owner via this endpoint
	if input.Role == models.RoleOwner {
		return nil, httputil.Validation("role", "use transfer ownership to assign owner role")
	}

	target, err := s.memberRepo.Get(ctx, workspaceID, targetUserID)
	if err != nil {
		return nil, err
	}

	// Cannot change the owner's role
	if target.Role == models.RoleOwner {
		return nil, httputil.Forbidden("cannot change the owner's role; transfer ownership first")
	}

	// Admin can only assign roles up to admin level
	actor, err := s.memberRepo.Get(ctx, workspaceID, actorID)
	if err != nil {
		return nil, err
	}
	if actor.Role == models.RoleAdmin && input.Role == models.RoleAdmin {
		// Admin can promote to admin — this is fine
	}
	if actor.Role != models.RoleOwner && target.Role == models.RoleAdmin {
		return nil, httputil.Forbidden("only the workspace owner can change admin roles")
	}

	return s.memberRepo.UpdateRole(ctx, sqlc.UpdateMemberRoleParams{
		WorkspaceID: workspaceID,
		UserID:      targetUserID,
		Role:        string(input.Role),
	})
}

func (s *workspaceService) TransferOwnership(ctx context.Context, workspaceID, actorID uuid.UUID, input models.TransferOwnershipInput) error {
	ws, err := s.wsRepo.GetByID(ctx, workspaceID)
	if err != nil {
		return err
	}

	if ws.OwnerID != actorID {
		return httputil.Forbidden("only the current owner can transfer ownership")
	}

	if input.NewOwnerID == actorID {
		return httputil.Validation("new_owner_id", "cannot transfer ownership to yourself")
	}

	// Verify new owner is a member
	_, err = s.memberRepo.Get(ctx, workspaceID, input.NewOwnerID)
	if err != nil {
		return httputil.Validation("new_owner_id", "new owner must be a workspace member")
	}

	// Transaction: update workspace owner_id, old owner -> admin, new owner -> owner
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return httputil.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback(ctx)

	qtx := sqlc.New(tx)
	txWsRepo := repository.NewWorkspaceRepository(qtx, s.logger)
	txMemberRepo := repository.NewWorkspaceMemberRepository(qtx, s.logger)

	// Update workspace owner
	_, err = txWsRepo.UpdateOwner(ctx, sqlc.UpdateWorkspaceOwnerParams{
		ID:      workspaceID,
		OwnerID: input.NewOwnerID,
	})
	if err != nil {
		return err
	}

	// Old owner becomes admin
	_, err = txMemberRepo.UpdateRole(ctx, sqlc.UpdateMemberRoleParams{
		WorkspaceID: workspaceID,
		UserID:      actorID,
		Role:        string(models.RoleAdmin),
	})
	if err != nil {
		return err
	}

	// New owner gets owner role
	_, err = txMemberRepo.UpdateRole(ctx, sqlc.UpdateMemberRoleParams{
		WorkspaceID: workspaceID,
		UserID:      input.NewOwnerID,
		Role:        string(models.RoleOwner),
	})
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return httputil.Wrap(err, "failed to commit transaction")
	}

	return nil
}

func (s *workspaceService) ListMembers(ctx context.Context, workspaceID uuid.UUID) ([]*models.WorkspaceMemberResponse, error) {
	return s.memberRepo.List(ctx, workspaceID)
}

func (s *workspaceService) GetMember(ctx context.Context, workspaceID, userID uuid.UUID) (*models.WorkspaceMember, error) {
	return s.memberRepo.Get(ctx, workspaceID, userID)
}

func (s *workspaceService) GetMemberCount(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	return s.memberRepo.GetCount(ctx, workspaceID)
}
