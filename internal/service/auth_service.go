package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/link-rift/link-rift/internal/config"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/link-rift/link-rift/pkg/crypto"
	"github.com/link-rift/link-rift/pkg/httputil"
	"github.com/link-rift/link-rift/pkg/paseto"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type AuthService interface {
	Register(ctx context.Context, input models.RegisterInput) (*models.AuthResponse, error)
	Login(ctx context.Context, input models.LoginInput, ip, userAgent string) (*models.AuthResponse, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
	RefreshToken(ctx context.Context, refreshToken, ip, userAgent string) (*models.AuthResponse, error)
	GetCurrentUser(ctx context.Context, userID uuid.UUID) (*models.UserResponse, error)
	ForgotPassword(ctx context.Context, input models.ForgotPasswordInput) error
	ResetPassword(ctx context.Context, input models.ResetPasswordInput) error
	VerifyEmail(ctx context.Context, input models.VerifyEmailInput) error
}

type authService struct {
	userRepo     repository.UserRepository
	sessionRepo  repository.SessionRepository
	resetRepo    repository.PasswordResetRepository
	tokenMaker   paseto.Maker
	pool         *pgxpool.Pool
	redis        *redis.Client
	cfg          *config.Config
	logger       *zap.Logger
}

func NewAuthService(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	resetRepo repository.PasswordResetRepository,
	tokenMaker paseto.Maker,
	pool *pgxpool.Pool,
	redisClient *redis.Client,
	cfg *config.Config,
	logger *zap.Logger,
) AuthService {
	return &authService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		resetRepo:   resetRepo,
		tokenMaker:  tokenMaker,
		pool:        pool,
		redis:       redisClient,
		cfg:         cfg,
		logger:      logger,
	}
}

func (s *authService) Register(ctx context.Context, input models.RegisterInput) (*models.AuthResponse, error) {
	passwordHash, err := crypto.HashPassword(input.Password)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to hash password")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback(ctx)

	qtx := sqlc.New(tx)

	user, err := qtx.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        strings.ToLower(strings.TrimSpace(input.Email)),
		PasswordHash: passwordHash,
		Name:         strings.TrimSpace(input.Name),
	})
	if err != nil {
		return nil, mapCreateUserError(err)
	}

	slug, err := generateWorkspaceSlug(input.Name)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to generate workspace slug")
	}
	workspace, err := qtx.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		Name:    fmt.Sprintf("%s's Workspace", strings.TrimSpace(input.Name)),
		Slug:    slug,
		OwnerID: user.ID,
		Plan:    "free",
		Settings: json.RawMessage(`{}`),
	})
	if err != nil {
		return nil, httputil.Wrap(err, "failed to create workspace")
	}

	_, err = qtx.AddWorkspaceMember(ctx, sqlc.AddWorkspaceMemberParams{
		WorkspaceID: workspace.ID,
		UserID:      user.ID,
		Role:        "owner",
	})
	if err != nil {
		return nil, httputil.Wrap(err, "failed to add workspace member")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, httputil.Wrap(err, "failed to commit transaction")
	}

	domainUser := models.UserFromSqlc(user)

	refreshToken, refreshTokenHash, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	session, err := s.sessionRepo.Create(ctx, sqlc.CreateSessionParams{
		UserID:           domainUser.ID,
		RefreshTokenHash: refreshTokenHash,
		IpAddress:        "",
		DeviceName:       pgtype.Text{},
		ExpiresAt:        pgtype.Timestamptz{Time: time.Now().Add(s.cfg.Auth.RefreshTokenExpiry), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	accessToken, _, err := s.tokenMaker.CreateToken(
		domainUser.ID,
		domainUser.Email,
		session.ID,
		s.cfg.Auth.AccessTokenExpiry,
	)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to create access token")
	}

	// Generate email verification token and store in Redis
	verifyToken, verifyTokenHash, err := generateRefreshToken()
	if err != nil {
		s.logger.Error("failed to generate email verification token", zap.Error(err))
	} else {
		key := fmt.Sprintf("email_verify:%s", verifyTokenHash)
		s.redis.Set(ctx, key, domainUser.ID.String(), 24*time.Hour)
		s.logger.Info("email verification link",
			zap.String("url", fmt.Sprintf("%s/auth/verify-email?token=%s", s.cfg.App.FrontendURL, verifyToken)),
		)
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         domainUser.ToResponse(),
	}, nil
}

func (s *authService) Login(ctx context.Context, input models.LoginInput, ip, userAgent string) (*models.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(input.Email)))
	if err != nil {
		if errors.Is(err, httputil.ErrNotFound) {
			return nil, httputil.Unauthorized("invalid email or password")
		}
		return nil, err
	}

	match, err := crypto.VerifyPassword(input.Password, user.PasswordHash)
	if err != nil || !match {
		return nil, httputil.Unauthorized("invalid email or password")
	}

	refreshToken, refreshTokenHash, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	session, err := s.sessionRepo.Create(ctx, sqlc.CreateSessionParams{
		UserID:           user.ID,
		RefreshTokenHash: refreshTokenHash,
		IpAddress:        ip,
		UserAgent:        pgtype.Text{String: userAgent, Valid: userAgent != ""},
		DeviceName:       pgtype.Text{},
		ExpiresAt:        pgtype.Timestamptz{Time: time.Now().Add(s.cfg.Auth.RefreshTokenExpiry), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	accessToken, _, err := s.tokenMaker.CreateToken(
		user.ID,
		user.Email,
		session.ID,
		s.cfg.Auth.AccessTokenExpiry,
	)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to create access token")
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.ToResponse(),
	}, nil
}

func (s *authService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return s.sessionRepo.Revoke(ctx, sessionID)
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken, ip, userAgent string) (*models.AuthResponse, error) {
	tokenHash := hashToken(refreshToken)

	session, err := s.sessionRepo.GetByRefreshTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, httputil.ErrNotFound) {
			return nil, httputil.Unauthorized("invalid refresh token")
		}
		return nil, err
	}

	// Revoke old session
	if err := s.sessionRepo.Revoke(ctx, session.ID); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}

	// Create new session with new refresh token
	newRefreshToken, newRefreshTokenHash, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	newSession, err := s.sessionRepo.Create(ctx, sqlc.CreateSessionParams{
		UserID:           user.ID,
		RefreshTokenHash: newRefreshTokenHash,
		IpAddress:        ip,
		UserAgent:        pgtype.Text{String: userAgent, Valid: userAgent != ""},
		DeviceName:       pgtype.Text{},
		ExpiresAt:        pgtype.Timestamptz{Time: time.Now().Add(s.cfg.Auth.RefreshTokenExpiry), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	accessToken, _, err := s.tokenMaker.CreateToken(
		user.ID,
		user.Email,
		newSession.ID,
		s.cfg.Auth.AccessTokenExpiry,
	)
	if err != nil {
		return nil, httputil.Wrap(err, "failed to create access token")
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         user.ToResponse(),
	}, nil
}

func (s *authService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*models.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user.ToResponse(), nil
}

func (s *authService) ForgotPassword(ctx context.Context, input models.ForgotPasswordInput) error {
	user, err := s.userRepo.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(input.Email)))
	if err != nil {
		// Return success even if user not found (prevent email enumeration)
		if errors.Is(err, httputil.ErrNotFound) {
			return nil
		}
		return err
	}

	token, tokenHash, err := generateRefreshToken()
	if err != nil {
		return err
	}

	_, err = s.resetRepo.Create(ctx, sqlc.CreatePasswordResetParams{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(1 * time.Hour), Valid: true},
	})
	if err != nil {
		return err
	}

	s.logger.Info("password reset link",
		zap.String("url", fmt.Sprintf("%s/auth/reset-password?token=%s", s.cfg.App.FrontendURL, token)),
		zap.String("email", user.Email),
	)

	return nil
}

func (s *authService) ResetPassword(ctx context.Context, input models.ResetPasswordInput) error {
	tokenHash := hashToken(input.Token)

	reset, err := s.resetRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, httputil.ErrNotFound) {
			return httputil.Validation("token", "invalid or expired reset token")
		}
		return err
	}

	passwordHash, err := crypto.HashPassword(input.NewPassword)
	if err != nil {
		return httputil.Wrap(err, "failed to hash password")
	}

	if err := s.userRepo.UpdatePassword(ctx, reset.UserID, passwordHash); err != nil {
		return err
	}

	if err := s.resetRepo.MarkUsed(ctx, reset.ID); err != nil {
		return err
	}

	if err := s.sessionRepo.RevokeAllForUser(ctx, reset.UserID); err != nil {
		return err
	}

	return nil
}

func (s *authService) VerifyEmail(ctx context.Context, input models.VerifyEmailInput) error {
	tokenHash := hashToken(input.Token)
	key := fmt.Sprintf("email_verify:%s", tokenHash)

	userIDStr, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return httputil.Validation("token", "invalid or expired verification token")
		}
		return httputil.Wrap(err, "failed to get verification token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return httputil.Wrap(err, "invalid user ID in verification token")
	}

	if err := s.userRepo.SetEmailVerified(ctx, userID); err != nil {
		return err
	}

	s.redis.Del(ctx, key)

	return nil
}

func generateRefreshToken() (token, hash string, err error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", httputil.Wrap(err, "failed to generate random token")
	}
	token = hex.EncodeToString(bytes)
	hash = hashToken(token)
	return token, hash, nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func generateWorkspaceSlug(name string) (string, error) {
	slug := strings.ToLower(strings.TrimSpace(name))
	slug = strings.ReplaceAll(slug, " ", "-")
	// Append short random suffix to avoid collisions
	suffix := make([]byte, 4)
	if _, err := rand.Read(suffix); err != nil {
		return "", fmt.Errorf("failed to generate random suffix: %w", err)
	}
	return fmt.Sprintf("%s-%s", slug, hex.EncodeToString(suffix)), nil
}

func mapCreateUserError(err error) error {
	if err == nil {
		return nil
	}
	errMsg := err.Error()
	if strings.Contains(errMsg, "23505") || strings.Contains(errMsg, "unique") || strings.Contains(errMsg, "duplicate") {
		return httputil.AlreadyExists("user")
	}
	return httputil.Wrap(err, "failed to create user")
}
