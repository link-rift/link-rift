package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/link-rift/link-rift/internal/license"
	"github.com/link-rift/link-rift/pkg/httputil"
)

const contextKeyLicense = "license"

// InjectLicense sets the current license in the Gin context.
func InjectLicense(manager *license.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(contextKeyLicense, manager.GetLicense())
		c.Next()
	}
}

// GetLicenseFromContext returns the license stored in the Gin context.
func GetLicenseFromContext(c *gin.Context) *license.License {
	val, exists := c.Get(contextKeyLicense)
	if !exists {
		return nil
	}
	lic, ok := val.(*license.License)
	if !ok {
		return nil
	}
	return lic
}

// RequireFeature returns 402 if the feature is not available on the current license.
func RequireFeature(manager *license.Manager, feature license.Feature) gin.HandlerFunc {
	return func(c *gin.Context) {
		if manager.HasFeature(feature) {
			c.Next()
			return
		}

		def, _ := license.GetFeatureDefinition(feature)
		appErr := httputil.PaymentRequiredWithDetails(string(feature), string(def.MinTier))
		c.AbortWithStatusJSON(http.StatusPaymentRequired, httputil.Response{
			Success: false,
			Error: &httputil.ErrorBody{
				Code:    appErr.Code,
				Message: appErr.Message,
				Details: appErr.Details,
			},
		})
	}
}

// RequirePlan returns 402 if the current tier is below the minimum required tier.
func RequirePlan(manager *license.Manager, minTier license.Tier) gin.HandlerFunc {
	return func(c *gin.Context) {
		if manager.GetTier().IncludesTier(minTier) {
			c.Next()
			return
		}

		appErr := httputil.PaymentRequired("this feature requires " + string(minTier) + " plan or higher")
		c.AbortWithStatusJSON(http.StatusPaymentRequired, httputil.Response{
			Success: false,
			Error: &httputil.ErrorBody{
				Code:    appErr.Code,
				Message: appErr.Message,
				Details: map[string]any{
					"current_tier":  string(manager.GetTier()),
					"required_tier": string(minTier),
				},
			},
		})
	}
}

// CheckLimit returns 402 if the current usage exceeds the license limit.
// usageFunc is called to get the current usage count.
func CheckLimit(manager *license.Manager, limitType license.LimitType, usageFunc func(c *gin.Context) (int64, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		current, err := usageFunc(c)
		if err != nil {
			httputil.RespondError(c, httputil.Wrap(err, "failed to check usage"))
			c.Abort()
			return
		}

		if manager.CheckLimit(limitType, current) {
			c.Next()
			return
		}

		limit := manager.GetLimits().GetLimit(limitType)
		appErr := httputil.PaymentRequired("usage limit reached")
		c.AbortWithStatusJSON(http.StatusPaymentRequired, httputil.Response{
			Success: false,
			Error: &httputil.ErrorBody{
				Code:    appErr.Code,
				Message: appErr.Message,
				Details: map[string]any{
					"limit_type": string(limitType),
					"current":    current,
					"limit":      limit,
				},
			},
		})
	}
}
