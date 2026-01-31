package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/repository"
	"github.com/link-rift/link-rift/internal/service"
	"github.com/link-rift/link-rift/pkg/httputil"
)

const contextKeyAPIKey = "api_key"

// APIKeyAuth authenticates requests using the X-API-Key header.
// On success, it loads the user and workspace into the gin context
// using the same keys as the session-based auth middleware.
func APIKeyAuth(
	apiKeyService service.APIKeyService,
	userRepo repository.UserRepository,
	wsRepo repository.WorkspaceRepository,
	memberRepo repository.WorkspaceMemberRepository,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawKey := c.GetHeader("X-API-Key")
		if rawKey == "" {
			c.Next()
			return
		}

		apiKey, err := apiKeyService.ValidateAPIKey(c.Request.Context(), rawKey)
		if err != nil {
			status := httputil.MapToHTTPStatus(err)
			c.AbortWithStatusJSON(status, httputil.Response{
				Success: false,
				Error: &httputil.ErrorBody{
					Code:    "UNAUTHORIZED",
					Message: "invalid API key",
				},
			})
			return
		}

		// Rate limit check
		remaining, err := apiKeyService.CheckRateLimit(c.Request.Context(), apiKey.ID)
		if err != nil {
			var appErr *httputil.AppError
			if ok := err.(*httputil.AppError); ok != nil {
				appErr = ok
			}
			if appErr != nil && appErr.Code == "RATE_LIMITED" {
				c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", *apiKey.RateLimit))
				c.Header("X-RateLimit-Remaining", "0")
				c.Header("X-RateLimit-Reset-After", "60")
				c.AbortWithStatusJSON(http.StatusTooManyRequests, httputil.Response{
					Success: false,
					Error: &httputil.ErrorBody{
						Code:    "RATE_LIMITED",
						Message: "API rate limit exceeded",
					},
				})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, httputil.Response{
				Success: false,
				Error: &httputil.ErrorBody{
					Code:    "INTERNAL_ERROR",
					Message: "failed to check rate limit",
				},
			})
			return
		}

		// Set rate limit headers
		if apiKey.RateLimit != nil {
			c.Header("X-RateLimit-Limit", strconv.Itoa(int(*apiKey.RateLimit)))
			if remaining >= 0 {
				c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
			}
			c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(60*time.Second).Unix(), 10))
			c.Header("X-RateLimit-Reset-After", "60")
		}

		// Load user
		user, err := userRepo.GetByID(c.Request.Context(), apiKey.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, httputil.Response{
				Success: false,
				Error: &httputil.ErrorBody{
					Code:    "UNAUTHORIZED",
					Message: "API key user not found",
				},
			})
			return
		}

		// Load workspace
		ws, err := wsRepo.GetByID(c.Request.Context(), apiKey.WorkspaceID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, httputil.Response{
				Success: false,
				Error: &httputil.ErrorBody{
					Code:    "FORBIDDEN",
					Message: "workspace not found",
				},
			})
			return
		}

		// Load membership
		member, err := memberRepo.Get(c.Request.Context(), apiKey.WorkspaceID, apiKey.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, httputil.Response{
				Success: false,
				Error: &httputil.ErrorBody{
					Code:    "FORBIDDEN",
					Message: "API key user is not a member of this workspace",
				},
			})
			return
		}

		c.Set(contextKeyUser, user)
		c.Set(contextKeyWorkspace, ws)
		c.Set(contextKeyWorkspaceMember, member)
		c.Set(contextKeyAPIKey, apiKey)
		c.Next()
	}
}

// RequireAPIKeyScope checks that the API key in context has the required scope.
func RequireAPIKeyScope(scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := GetAPIKeyFromContext(c)
		if apiKey == nil {
			// Not an API key request, skip scope check
			c.Next()
			return
		}

		if !apiKey.HasScope(scope) {
			c.AbortWithStatusJSON(http.StatusForbidden, httputil.Response{
				Success: false,
				Error: &httputil.ErrorBody{
					Code:    "FORBIDDEN",
					Message: fmt.Sprintf("API key missing required scope: %s", scope),
				},
			})
			return
		}

		c.Next()
	}
}

// GetAPIKeyFromContext returns the API key from the gin context, if present.
func GetAPIKeyFromContext(c *gin.Context) *models.APIKey {
	val, exists := c.Get(contextKeyAPIKey)
	if !exists {
		return nil
	}
	key, ok := val.(*models.APIKey)
	if !ok {
		return nil
	}
	return key
}
