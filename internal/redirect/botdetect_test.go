package redirect

import "testing"

func TestBotDetector_KnownBots(t *testing.T) {
	d := NewBotDetector()

	bots := []struct {
		ua   string
		name string
	}{
		{"Googlebot/2.1 (+http://www.google.com/bot.html)", "Googlebot"},
		{"Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)", "Bingbot"},
		{"facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)", "Facebook"},
		{"Twitterbot/1.0", "Twitterbot"},
		{"LinkedInBot/1.0 (compatible; Mozilla/5.0)", "LinkedInBot"},
		{"Slackbot-LinkExpanding 1.0 (+https://api.slack.com/robots)", "Slackbot"},
		{"Discordbot/2.0", "Discordbot"},
		{"curl/7.68.0", "curl"},
		{"wget/1.20.3", "wget"},
		{"python-requests/2.25.1", "python-requests"},
		{"go-http-client/1.1", "go-http-client"},
		{"Apache-HttpClient/4.5.13", "Apache HttpClient"},
		{"okhttp/4.9.3", "OkHttp"},
		{"Mozilla/5.0 HeadlessChrome/90.0.4430.212", "HeadlessChrome"},
		{"PhantomJS/2.1.1", "PhantomJS"},
		{"UptimeRobot/2.0", "UptimeRobot"},
		{"SemrushBot/7~bl", "SemrushBot"},
		{"AhrefsBot/7.0", "AhrefsBot"},
		{"", "empty UA"},
	}

	for _, tt := range bots {
		t.Run(tt.name, func(t *testing.T) {
			if !d.IsBot(tt.ua) {
				t.Errorf("expected %q to be detected as bot", tt.ua)
			}
		})
	}
}

func TestBotDetector_HumanBrowsers(t *testing.T) {
	d := NewBotDetector()

	humans := []struct {
		ua   string
		name string
	}{
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36", "Chrome Windows"},
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15", "Safari Mac"},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0", "Firefox Windows"},
		{"Mozilla/5.0 (Linux; Android 11; SM-G998B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.120 Mobile Safari/537.36", "Chrome Android"},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Mobile/15E148 Safari/604.1", "Safari iOS"},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.59", "Edge"},
	}

	for _, tt := range humans {
		t.Run(tt.name, func(t *testing.T) {
			if d.IsBot(tt.ua) {
				t.Errorf("expected %q to NOT be detected as bot", tt.ua)
			}
		})
	}
}

// --- Benchmarks ---

func BenchmarkBotDetectorIsBot_Human(b *testing.B) {
	d := NewBotDetector()
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.IsBot(ua)
	}
}

func BenchmarkBotDetectorIsBot_Bot(b *testing.B) {
	d := NewBotDetector()
	ua := "Googlebot/2.1 (+http://www.google.com/bot.html)"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.IsBot(ua)
	}
}

func BenchmarkBotDetectorIsBot_Empty(b *testing.B) {
	d := NewBotDetector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.IsBot("")
	}
}
