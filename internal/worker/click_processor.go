package worker

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/link-rift/link-rift/internal/models"
	"github.com/link-rift/link-rift/internal/redirect"
	"github.com/link-rift/link-rift/internal/repository"
	"github.com/link-rift/link-rift/internal/repository/sqlc"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	clickQueueKey = "clicks:queue"
	batchSize     = 100
	batchWindow   = 1 * time.Second
)

// ClickProcessor reads click events from the Redis queue and processes them into the database.
type ClickProcessor struct {
	redis       *redis.Client
	clickRepo   repository.ClickRepository
	linkRepo    repository.LinkRepository
	botDetector *redirect.BotDetector
	logger      *zap.Logger
	done        chan struct{}
}

func NewClickProcessor(
	redisClient *redis.Client,
	clickRepo repository.ClickRepository,
	linkRepo repository.LinkRepository,
	botDetector *redirect.BotDetector,
	logger *zap.Logger,
) *ClickProcessor {
	return &ClickProcessor{
		redis:       redisClient,
		clickRepo:   clickRepo,
		linkRepo:    linkRepo,
		botDetector: botDetector,
		logger:      logger,
		done:        make(chan struct{}),
	}
}

// Start begins processing click events from the Redis queue.
func (cp *ClickProcessor) Start(ctx context.Context) {
	cp.logger.Info("click processor started")

	for {
		select {
		case <-ctx.Done():
			cp.logger.Info("click processor shutting down")
			return
		case <-cp.done:
			return
		default:
			cp.processBatch(ctx)
		}
	}
}

// Stop signals the processor to stop.
func (cp *ClickProcessor) Stop() {
	close(cp.done)
}

func (cp *ClickProcessor) processBatch(ctx context.Context) {
	// BLPOP with a timeout so we don't block forever
	result, err := cp.redis.BLPop(ctx, 2*time.Second, clickQueueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return // Timeout, no events
		}
		if ctx.Err() != nil {
			return // Context cancelled
		}
		cp.logger.Error("failed to pop from click queue", zap.Error(err))
		time.Sleep(1 * time.Second)
		return
	}

	// result[0] is the key, result[1] is the value
	events := []*models.ClickEvent{}
	var firstEvent models.ClickEvent
	if err := json.Unmarshal([]byte(result[1]), &firstEvent); err != nil {
		cp.logger.Warn("failed to unmarshal click event", zap.Error(err))
		return
	}
	events = append(events, &firstEvent)

	// Try to collect more events within the batch window
	deadline := time.Now().Add(batchWindow)
	for len(events) < batchSize && time.Now().Before(deadline) {
		data, err := cp.redis.LPop(ctx, clickQueueKey).Bytes()
		if err != nil {
			break // No more events
		}
		var event models.ClickEvent
		if err := json.Unmarshal(data, &event); err != nil {
			cp.logger.Warn("failed to unmarshal click event", zap.Error(err))
			continue
		}
		events = append(events, &event)
	}

	cp.processEvents(ctx, events)
}

func (cp *ClickProcessor) processEvents(ctx context.Context, events []*models.ClickEvent) {
	for _, event := range events {
		isBot := cp.botDetector.IsBot(event.UserAgent)

		// Parse user agent
		browser, browserVersion := parseBrowser(event.UserAgent)
		osName, osVersion := parseOS(event.UserAgent)
		deviceType := parseDeviceType(event.UserAgent)

		params := sqlc.InsertClickParams{
			LinkID:         event.LinkID,
			ClickedAt:      pgtype.Timestamptz{Time: event.Timestamp, Valid: true},
			IpAddress:      event.IP,
			UserAgent:      pgtype.Text{String: event.UserAgent, Valid: event.UserAgent != ""},
			Referer:        pgtype.Text{String: event.Referer, Valid: event.Referer != ""},
			IsBot:          isBot,
			Browser:        pgtype.Text{String: browser, Valid: browser != ""},
			BrowserVersion: pgtype.Text{String: browserVersion, Valid: browserVersion != ""},
			Os:             pgtype.Text{String: osName, Valid: osName != ""},
			OsVersion:      pgtype.Text{String: osVersion, Valid: osVersion != ""},
			DeviceType:     pgtype.Text{String: deviceType, Valid: deviceType != ""},
		}

		if err := cp.clickRepo.Insert(ctx, params); err != nil {
			cp.logger.Error("failed to insert click",
				zap.Error(err),
				zap.String("link_id", event.LinkID.String()),
			)
			continue
		}

		// Increment link click counters
		if !isBot {
			if err := cp.linkRepo.IncrementClicks(ctx, event.LinkID); err != nil {
				cp.logger.Error("failed to increment click counter",
					zap.Error(err),
					zap.String("link_id", event.LinkID.String()),
				)
			}
		}
	}

	cp.logger.Debug("processed click batch", zap.Int("count", len(events)))
}

// Simple UA parsing functions

var (
	chromeRe  = regexp.MustCompile(`Chrome/(\d+[\.\d]*)`)
	firefoxRe = regexp.MustCompile(`Firefox/(\d+[\.\d]*)`)
	safariRe  = regexp.MustCompile(`Version/(\d+[\.\d]*).*Safari`)
	edgeRe    = regexp.MustCompile(`Edg/(\d+[\.\d]*)`)
	operaRe   = regexp.MustCompile(`OPR/(\d+[\.\d]*)`)

	windowsRe = regexp.MustCompile(`Windows NT (\d+[\.\d]*)`)
	macRe     = regexp.MustCompile(`Mac OS X (\d+[_\.\d]*)`)
	linuxRe   = regexp.MustCompile(`Linux`)
	androidRe = regexp.MustCompile(`Android (\d+[\.\d]*)`)
	iosRe     = regexp.MustCompile(`(?:iPhone|iPad) OS (\d+[_\.\d]*)`)
)

func parseBrowser(ua string) (name, version string) {
	if m := edgeRe.FindStringSubmatch(ua); len(m) > 1 {
		return "Edge", m[1]
	}
	if m := operaRe.FindStringSubmatch(ua); len(m) > 1 {
		return "Opera", m[1]
	}
	if m := chromeRe.FindStringSubmatch(ua); len(m) > 1 {
		return "Chrome", m[1]
	}
	if m := firefoxRe.FindStringSubmatch(ua); len(m) > 1 {
		return "Firefox", m[1]
	}
	if m := safariRe.FindStringSubmatch(ua); len(m) > 1 {
		return "Safari", m[1]
	}
	return "", ""
}

func parseOS(ua string) (name, version string) {
	if m := iosRe.FindStringSubmatch(ua); len(m) > 1 {
		return "iOS", strings.ReplaceAll(m[1], "_", ".")
	}
	if m := androidRe.FindStringSubmatch(ua); len(m) > 1 {
		return "Android", m[1]
	}
	if m := macRe.FindStringSubmatch(ua); len(m) > 1 {
		return "macOS", strings.ReplaceAll(m[1], "_", ".")
	}
	if m := windowsRe.FindStringSubmatch(ua); len(m) > 1 {
		return "Windows", m[1]
	}
	if linuxRe.MatchString(ua) {
		return "Linux", ""
	}
	return "", ""
}

func parseDeviceType(ua string) string {
	uaLower := strings.ToLower(ua)
	if strings.Contains(uaLower, "tablet") || strings.Contains(uaLower, "ipad") {
		return "tablet"
	}
	if strings.Contains(uaLower, "mobile") || strings.Contains(uaLower, "iphone") || strings.Contains(uaLower, "android") {
		return "mobile"
	}
	return "desktop"
}
