package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository"
	"github.com/link-rift/link-rift/pkg/httputil"
	"github.com/link-rift/link-rift/pkg/paseto"
)

const (
	contextKeyUser      = "user"
	contextKeySessionID = "session_id"
)

func RequireAuth(tokenMaker paseto.Maker, userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, httputil.Response{
				Success: false,
				Error: &httputil.ErrorBody{
					Code:    "UNAUTHORIZED",
					Message: "missing or invalid authorization header",
				},
			})
			return
		}

		claims, err := tokenMaker.VerifyToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, httputil.Response{
				Success: false,
				Error: &httputil.ErrorBody{
					Code:    "UNAUTHORIZED",
					Message: "invalid or expired token",
				},
			})
			return
		}

		user, err := userRepo.GetByID(c.Request.Context(), claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, httputil.Response{
				Success: false,
				Error: &httputil.ErrorBody{
					Code:    "UNAUTHORIZED",
					Message: "user not found",
				},
			})
			return
		}

		c.Set(contextKeyUser, user)
		c.Set(contextKeySessionID, claims.SessionID)
		c.Next()
	}
}

func OptionalAuth(tokenMaker paseto.Maker, userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token == "" {
			c.Next()
			return
		}

		claims, err := tokenMaker.VerifyToken(token)
		if err != nil {
			c.Next()
			return
		}

		user, err := userRepo.GetByID(c.Request.Context(), claims.UserID)
		if err != nil {
			c.Next()
			return
		}

		c.Set(contextKeyUser, user)
		c.Set(contextKeySessionID, claims.SessionID)
		c.Next()
	}
}

func GetUserFromContext(c *gin.Context) *models.User {
	val, exists := c.Get(contextKeyUser)
	if !exists {
		return nil
	}
	user, ok := val.(*models.User)
	if !ok {
		return nil
	}
	return user
}

func GetSessionIDFromContext(c *gin.Context) uuid.UUID {
	val, exists := c.Get(contextKeySessionID)
	if !exists {
		return uuid.Nil
	}
	id, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return id
}

func extractBearerToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
