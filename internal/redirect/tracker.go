package redirect

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/link-rift/link-rift/internal/models"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	clickQueueKey  = "clicks:queue"
	defaultBatch   = 500
)

// ClickTracker provides non-blocking, async click event tracking.
// Events are buffered in-memory and flushed to a Redis list for downstream processing.
type ClickTracker struct {
	redis     *redis.Client
	logger    *zap.Logger
	events    chan *models.ClickEvent
	batchSize int
	flushTick time.Duration
	wg        sync.WaitGroup
	done      chan struct{}
}

func NewClickTracker(redisClient *redis.Client, bufferSize int, flushInterval time.Duration, logger *zap.Logger) *ClickTracker {
	ct := &ClickTracker{
		redis:     redisClient,
		logger:    logger,
		events:    make(chan *models.ClickEvent, bufferSize),
		batchSize: defaultBatch,
		flushTick: flushInterval,
		done:      make(chan struct{}),
	}
	ct.wg.Add(1)
	go ct.processLoop()
	return ct
}

// Track enqueues a click event for async processing. Non-blocking — drops events if buffer is full.
func (ct *ClickTracker) Track(event *models.ClickEvent) {
	select {
	case ct.events <- event:
	default:
		ct.logger.Warn("click tracker buffer full, dropping event",
			zap.String("short_code", event.ShortCode),
		)
	}
}

// Shutdown gracefully stops the tracker and flushes remaining events.
func (ct *ClickTracker) Shutdown(ctx context.Context) {
	close(ct.done)
	ct.wg.Wait()

	// Flush remaining events in the channel
	ct.flushRemaining(ctx)
}

func (ct *ClickTracker) processLoop() {
	defer ct.wg.Done()

	ticker := time.NewTicker(ct.flushTick)
	defer ticker.Stop()

	batch := make([]*models.ClickEvent, 0, ct.batchSize)

	for {
		select {
		case event := <-ct.events:
			batch = append(batch, event)
			if len(batch) >= ct.batchSize {
				ct.flush(context.Background(), batch)
				batch = make([]*models.ClickEvent, 0, ct.batchSize)
			}
		case <-ticker.C:
			if len(batch) > 0 {
				ct.flush(context.Background(), batch)
				batch = make([]*models.ClickEvent, 0, ct.batchSize)
			}
		case <-ct.done:
			// Flush remaining batch
			if len(batch) > 0 {
				ct.flush(context.Background(), batch)
			}
			return
		}
	}
}

func (ct *ClickTracker) flush(ctx context.Context, batch []*models.ClickEvent) {
	if len(batch) == 0 {
		return
	}

	vals := make([]interface{}, 0, len(batch))
	for _, event := range batch {
		data, err := json.Marshal(event)
		if err != nil {
			ct.logger.Warn("failed to marshal click event", zap.Error(err))
			continue
		}
		vals = append(vals, data)
	}

	if len(vals) == 0 {
		return
	}

	if err := ct.redis.RPush(ctx, clickQueueKey, vals...).Err(); err != nil {
		ct.logger.Error("failed to push click events to Redis",
			zap.Error(err),
			zap.Int("count", len(vals)),
		)
	}
}

func (ct *ClickTracker) flushRemaining(ctx context.Context) {
	batch := make([]*models.ClickEvent, 0, ct.batchSize)
	for {
		select {
		case event := <-ct.events:
			batch = append(batch, event)
		default:
			if len(batch) > 0 {
				ct.flush(ctx, batch)
			}
			return
		}
	}
}
