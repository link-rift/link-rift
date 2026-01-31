package redirect

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"go.uber.org/zap"
)

// ruleCondition represents the JSON structure of a link rule's conditions.
type ruleCondition struct {
	Value string `json:"value"`
}

// RuleEngine evaluates conditional redirect rules for a link.
type RuleEngine struct {
	queries *sqlc.Queries
	logger  *zap.Logger
}

func NewRuleEngine(queries *sqlc.Queries, logger *zap.Logger) *RuleEngine {
	return &RuleEngine{queries: queries, logger: logger}
}

// Evaluate checks all active rules for a link and returns the destination URL
// if a rule matches, or empty string if no rules match.
func (re *RuleEngine) Evaluate(ctx context.Context, linkID uuid.UUID, r *http.Request) (string, bool) {
	rules, err := re.queries.GetActiveRulesForLink(ctx, linkID)
	if err != nil {
		re.logger.Warn("failed to fetch rules for link", zap.Error(err), zap.String("link_id", linkID.String()))
		return "", false
	}

	if len(rules) == 0 {
		return "", false
	}

	ua := r.UserAgent()

	for _, rule := range rules {
		if re.matchRule(rule, ua, r) {
			return rule.DestinationUrl, true
		}
	}

	return "", false
}

func (re *RuleEngine) parseCondition(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	var cond ruleCondition
	if err := json.Unmarshal(raw, &cond); err != nil {
		// Try parsing as a plain string
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return ""
		}
		return s
	}
	return cond.Value
}

func (re *RuleEngine) matchRule(rule sqlc.LinkRule, ua string, r *http.Request) bool {
	switch rule.RuleType {
	case "device":
		return re.matchDevice(rule, ua)
	case "browser":
		return re.matchBrowser(rule, ua)
	case "os":
		return re.matchOS(rule, ua)
	default:
		return false
	}
}

func (re *RuleEngine) matchDevice(rule sqlc.LinkRule, ua string) bool {
	condValue := re.parseCondition(rule.Conditions)
	if condValue == "" {
		return false
	}
	uaLower := strings.ToLower(ua)
	condLower := strings.ToLower(condValue)

	switch condLower {
	case "mobile":
		return strings.Contains(uaLower, "mobile") || strings.Contains(uaLower, "android") || strings.Contains(uaLower, "iphone")
	case "tablet":
		return strings.Contains(uaLower, "tablet") || strings.Contains(uaLower, "ipad")
	case "desktop":
		return !strings.Contains(uaLower, "mobile") && !strings.Contains(uaLower, "android") && !strings.Contains(uaLower, "iphone") && !strings.Contains(uaLower, "tablet") && !strings.Contains(uaLower, "ipad")
	default:
		return false
	}
}

func (re *RuleEngine) matchBrowser(rule sqlc.LinkRule, ua string) bool {
	condValue := re.parseCondition(rule.Conditions)
	if condValue == "" {
		return false
	}
	uaLower := strings.ToLower(ua)
	condLower := strings.ToLower(condValue)

	switch condLower {
	case "chrome":
		return strings.Contains(uaLower, "chrome") && !strings.Contains(uaLower, "edg")
	case "firefox":
		return strings.Contains(uaLower, "firefox")
	case "safari":
		return strings.Contains(uaLower, "safari") && !strings.Contains(uaLower, "chrome")
	case "edge":
		return strings.Contains(uaLower, "edg")
	default:
		return false
	}
}

func (re *RuleEngine) matchOS(rule sqlc.LinkRule, ua string) bool {
	condValue := re.parseCondition(rule.Conditions)
	if condValue == "" {
		return false
	}
	uaLower := strings.ToLower(ua)
	condLower := strings.ToLower(condValue)

	switch condLower {
	case "windows":
		return strings.Contains(uaLower, "windows")
	case "macos", "mac":
		return strings.Contains(uaLower, "macintosh") || strings.Contains(uaLower, "mac os")
	case "linux":
		return strings.Contains(uaLower, "linux") && !strings.Contains(uaLower, "android")
	case "ios":
		return strings.Contains(uaLower, "iphone") || strings.Contains(uaLower, "ipad")
	case "android":
		return strings.Contains(uaLower, "android")
	default:
		return false
	}
}
