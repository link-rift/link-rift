package scim

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TokenValidator is the interface for validating SCIM bearer tokens.
type TokenValidator interface {
	ValidateToken(tokenHash string) (workspaceID uuid.UUID, tokenID uuid.UUID, err error)
	RecordTokenUsage(tokenID uuid.UUID)
}

// AuthMiddleware returns a Gin middleware that authenticates SCIM requests
// using Bearer token from the Authorization header.
func AuthMiddleware(validator TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, NewSCIMError(401, "Authorization header required"))
			c.Abort()
			return
		}

		if !strings.HasPrefix(auth, "Bearer ") {
			c.JSON(http.StatusUnauthorized, NewSCIMError(401, "Bearer token required"))
			c.Abort()
			return
		}

		token := strings.TrimPrefix(auth, "Bearer ")
		tokenHash := HashToken(token)

		workspaceID, tokenID, err := validator.ValidateToken(tokenHash)
		if err != nil {
			c.JSON(http.StatusUnauthorized, NewSCIMError(401, "Invalid or expired token"))
			c.Abort()
			return
		}

		// Set workspace ID in context for downstream handlers
		c.Set("scim_workspace_id", workspaceID)
		c.Set("scim_token_id", tokenID)

		// Record token usage asynchronously (best-effort)
		go validator.RecordTokenUsage(tokenID)

		c.Next()
	}
}

// HashToken creates a SHA-256 hash of a SCIM token.
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// GetWorkspaceIDFromContext extracts the workspace ID set by SCIM auth middleware.
func GetWorkspaceIDFromContext(c *gin.Context) uuid.UUID {
	val, _ := c.Get("scim_workspace_id")
	wsID, _ := val.(uuid.UUID)
	return wsID
}
