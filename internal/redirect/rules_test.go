package redirect

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"go.uber.org/zap"
)

func makeTestRule(ruleType string, conditions json.RawMessage, destURL string) sqlc.LinkRule {
	return sqlc.LinkRule{
		ID:             uuid.New(),
		LinkID:         uuid.New(),
		RuleType:       ruleType,
		Priority:       1,
		IsActive:       true,
		Conditions:     conditions,
		DestinationUrl: destURL,
	}
}

func TestRuleEngine_ParseCondition(t *testing.T) {
	re := &RuleEngine{logger: zap.NewNop()}

	tests := []struct {
		name     string
		raw      json.RawMessage
		expected string
	}{
		{"object format", json.RawMessage(`{"value":"mobile"}`), "mobile"},
		{"string format", json.RawMessage(`"desktop"`), "desktop"},
		{"empty", json.RawMessage(``), ""},
		{"null", json.RawMessage(`null`), ""},
		{"invalid json", json.RawMessage(`{invalid`), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := re.parseCondition(tt.raw)
			if got != tt.expected {
				t.Errorf("parseCondition(%q) = %q, want %q", string(tt.raw), got, tt.expected)
			}
		})
	}
}

func TestRuleEngine_MatchDevice(t *testing.T) {
	re := &RuleEngine{logger: zap.NewNop()}

	tests := []struct {
		name      string
		condition string
		ua        string
		expected  bool
	}{
		{"mobile android", "mobile", "Mozilla/5.0 (Linux; Android 11; SM-G998B) Mobile Safari/537.36", true},
		{"mobile iphone", "mobile", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) Safari/604.1", true},
		{"tablet ipad", "tablet", "Mozilla/5.0 (iPad; CPU OS 14_6 like Mac OS X) Safari/604.1", true},
		{"desktop chrome", "desktop", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/91.0.4472.124 Safari/537.36", true},
		{"desktop not mobile", "desktop", "Mozilla/5.0 (Linux; Android 11) Mobile Chrome", false},
		{"empty condition", "", "any user agent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond, _ := json.Marshal(ruleCondition{Value: tt.condition})
			rule := makeTestRule("device", cond, "https://dest.com")
			got := re.matchDevice(rule, tt.ua)
			if got != tt.expected {
				t.Errorf("matchDevice(%q, %q) = %v, want %v", tt.condition, tt.ua, got, tt.expected)
			}
		})
	}
}

func TestRuleEngine_MatchBrowser(t *testing.T) {
	re := &RuleEngine{logger: zap.NewNop()}

	tests := []struct {
		name      string
		condition string
		ua        string
		expected  bool
	}{
		{"chrome", "chrome", "Mozilla/5.0 Chrome/91.0.4472.124 Safari/537.36", true},
		{"chrome not edge", "chrome", "Mozilla/5.0 Chrome/91.0 Safari/537.36 Edg/91.0.864.59", false},
		{"firefox", "firefox", "Mozilla/5.0 (Windows NT 10.0; rv:89.0) Gecko/20100101 Firefox/89.0", true},
		{"safari", "safari", "Mozilla/5.0 (Macintosh) AppleWebKit/605.1.15 Safari/605.1.15", true},
		{"safari not chrome", "safari", "Mozilla/5.0 Chrome/91.0.4472.124 Safari/537.36", false},
		{"edge", "edge", "Mozilla/5.0 Chrome/91.0 Safari/537.36 Edg/91.0.864.59", true},
		{"unknown browser", "opera", "Mozilla/5.0 Chrome/91.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond, _ := json.Marshal(ruleCondition{Value: tt.condition})
			rule := makeTestRule("browser", cond, "https://dest.com")
			got := re.matchBrowser(rule, tt.ua)
			if got != tt.expected {
				t.Errorf("matchBrowser(%q, %q) = %v, want %v", tt.condition, tt.ua, got, tt.expected)
			}
		})
	}
}

func TestRuleEngine_MatchOS(t *testing.T) {
	re := &RuleEngine{logger: zap.NewNop()}

	tests := []struct {
		name      string
		condition string
		ua        string
		expected  bool
	}{
		{"windows", "windows", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)", true},
		{"macos", "macos", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)", true},
		{"mac alias", "mac", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)", true},
		{"linux", "linux", "Mozilla/5.0 (X11; Linux x86_64)", true},
		{"linux not android", "linux", "Mozilla/5.0 (Linux; Android 11; SM-G998B)", false},
		{"ios iphone", "ios", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X)", true},
		{"ios ipad", "ios", "Mozilla/5.0 (iPad; CPU OS 14_6 like Mac OS X)", true},
		{"android", "android", "Mozilla/5.0 (Linux; Android 11; SM-G998B)", true},
		{"unknown os", "chromeos", "Mozilla/5.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond, _ := json.Marshal(ruleCondition{Value: tt.condition})
			rule := makeTestRule("os", cond, "https://dest.com")
			got := re.matchOS(rule, tt.ua)
			if got != tt.expected {
				t.Errorf("matchOS(%q, %q) = %v, want %v", tt.condition, tt.ua, got, tt.expected)
			}
		})
	}
}

// --- Benchmarks ---

func BenchmarkRuleEngineMatchDevice(b *testing.B) {
	re := &RuleEngine{logger: zap.NewNop()}
	cond, _ := json.Marshal(ruleCondition{Value: "mobile"})
	rule := makeTestRule("device", cond, "https://dest.com")
	ua := "Mozilla/5.0 (Linux; Android 11; SM-G998B) Mobile Safari/537.36"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re.matchDevice(rule, ua)
	}
}

func BenchmarkRuleEngineMatchBrowser(b *testing.B) {
	re := &RuleEngine{logger: zap.NewNop()}
	cond, _ := json.Marshal(ruleCondition{Value: "chrome"})
	rule := makeTestRule("browser", cond, "https://dest.com")
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/91.0.4472.124 Safari/537.36"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		re.matchBrowser(rule, ua)
	}
}
