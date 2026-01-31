package redirect

import (
	"regexp"
	"strings"
)

// BotDetector identifies bot/crawler traffic from User-Agent strings.
type BotDetector struct {
	patterns []*regexp.Regexp
}

func NewBotDetector() *BotDetector {
	rawPatterns := []string{
		// Search engine crawlers
		`(?i)googlebot`,
		`(?i)bingbot`,
		`(?i)yandexbot`,
		`(?i)baiduspider`,
		`(?i)duckduckbot`,
		`(?i)sogou`,
		`(?i)exabot`,
		`(?i)ia_archiver`,

		// Social media crawlers
		`(?i)facebookexternalhit`,
		`(?i)facebot`,
		`(?i)twitterbot`,
		`(?i)linkedinbot`,
		`(?i)pinterestbot`,
		`(?i)slackbot`,
		`(?i)telegrambot`,
		`(?i)whatsapp`,
		`(?i)discordbot`,

		// SEO and monitoring tools
		`(?i)semrushbot`,
		`(?i)ahrefsbot`,
		`(?i)mj12bot`,
		`(?i)dotbot`,
		`(?i)screaming frog`,
		`(?i)rogerbot`,

		// General bot patterns
		`(?i)bot[\s/;]`,
		`(?i)crawler`,
		`(?i)spider`,
		`(?i)headlesschrome`,
		`(?i)phantomjs`,

		// Monitoring/uptime tools
		`(?i)uptimerobot`,
		`(?i)pingdom`,
		`(?i)newrelic`,
		`(?i)datadog`,
		`(?i)site24x7`,
		`(?i)statuscake`,

		// HTTP clients/libraries
		`(?i)^curl/`,
		`(?i)^wget/`,
		`(?i)^python-requests`,
		`(?i)^go-http-client`,
		`(?i)^java/`,
		`(?i)^apache-httpclient`,
		`(?i)^okhttp`,
	}

	patterns := make([]*regexp.Regexp, 0, len(rawPatterns))
	for _, p := range rawPatterns {
		patterns = append(patterns, regexp.MustCompile(p))
	}

	return &BotDetector{patterns: patterns}
}

// IsBot returns true if the User-Agent string matches a known bot pattern.
func (d *BotDetector) IsBot(userAgent string) bool {
	if userAgent == "" {
		return true // No UA is likely a bot
	}

	ua := strings.TrimSpace(userAgent)
	for _, p := range d.patterns {
		if p.MatchString(ua) {
			return true
		}
	}

	return false
}
