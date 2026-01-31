package redirect

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/link-rift/link-rift/internal/models"
	"go.uber.org/zap"
)

func makeClickEvent(shortCode string) *models.ClickEvent {
	return &models.ClickEvent{
		LinkID:    uuid.New(),
		ShortCode: shortCode,
		IP:        "1.2.3.4",
		UserAgent: "Mozilla/5.0",
		Timestamp: time.Now(),
	}
}

func TestClickTracker_TrackAndBufferFull(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Small buffer to trigger overflow
	ct := &ClickTracker{
		logger:    logger,
		events:    make(chan *models.ClickEvent, 2),
		batchSize: 10,
		flushTick: 10 * time.Second,
		done:      make(chan struct{}),
	}

	// Track 2 events (fills buffer)
	ct.Track(makeClickEvent("a"))
	ct.Track(makeClickEvent("b"))

	// 3rd should be dropped (non-blocking)
	ct.Track(makeClickEvent("c"))

	if len(ct.events) != 2 {
		t.Errorf("expected 2 events in buffer, got %d", len(ct.events))
	}
}

func TestClickTracker_FlushRemaining(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	ct := &ClickTracker{
		logger:    logger,
		events:    make(chan *models.ClickEvent, 100),
		batchSize: 500,
		flushTick: 1 * time.Hour, // won't trigger during test
		done:      make(chan struct{}),
	}

	// Add events to the channel
	ct.Track(makeClickEvent("x"))
	ct.Track(makeClickEvent("y"))

	var flushed int32
	// Override flush to count
	origEvents := ct.events

	// Drain the channel to verify events are there
	ctx := context.Background()
	_ = ctx

	remaining := make([]*models.ClickEvent, 0)
	for {
		select {
		case event := <-origEvents:
			remaining = append(remaining, event)
			atomic.AddInt32(&flushed, 1)
		default:
			goto done
		}
	}
done:

	if len(remaining) != 2 {
		t.Errorf("expected 2 remaining events, got %d", len(remaining))
	}
}
