# Analytics Pipeline

> Last Updated: 2025-01-24

Linkrift's analytics pipeline processes millions of click events daily, providing both real-time and historical analytics through a robust data processing architecture.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Click Event Processing](#click-event-processing)
  - [Event Schema](#event-schema)
  - [Event Ingestion](#event-ingestion)
- [Real-Time vs Batch Processing](#real-time-vs-batch-processing)
  - [Real-Time Stream](#real-time-stream)
  - [Batch Processing](#batch-processing)
- [ClickHouse Schema](#clickhouse-schema)
  - [Main Events Table](#main-events-table)
  - [Materialized Views](#materialized-views)
- [Aggregation Jobs](#aggregation-jobs)
- [Geographic Enrichment](#geographic-enrichment)
- [Real-Time Dashboard WebSocket](#real-time-dashboard-websocket)
- [API Reference](#api-reference)

---

## Overview

The analytics pipeline is designed for:

- **High throughput**: 100,000+ events/second ingestion
- **Real-time analytics**: Sub-second updates to dashboards
- **Historical analysis**: Efficient queries over billions of events
- **Geographic enrichment**: MaxMind GeoIP2 integration
- **Cost efficiency**: ClickHouse columnar storage with 10x compression

## Architecture

```
                                    ┌─────────────────────────────────────────┐
                                    │           Redirect Service              │
                                    │         (Click Event Source)            │
                                    └──────────────────┬──────────────────────┘
                                                       │
                                                       ▼
                                    ┌─────────────────────────────────────────┐
                                    │              Kafka                       │
                                    │         (clicks.events topic)           │
                                    └──────────────────┬──────────────────────┘
                                                       │
                              ┌────────────────────────┼────────────────────────┐
                              │                        │                        │
                              ▼                        ▼                        ▼
               ┌──────────────────────┐  ┌──────────────────────┐  ┌──────────────────────┐
               │   Real-Time Worker   │  │   Real-Time Worker   │  │   Batch Aggregator   │
               │   (WebSocket Feed)   │  │   (ClickHouse Write) │  │   (Hourly/Daily)     │
               └──────────┬───────────┘  └──────────┬───────────┘  └──────────┬───────────┘
                          │                         │                         │
                          ▼                         │                         │
               ┌──────────────────────┐             │                         │
               │        Redis         │             │                         │
               │   (Real-Time Pub)    │             │                         │
               └──────────┬───────────┘             │                         │
                          │                         ▼                         ▼
                          │            ┌──────────────────────────────────────────┐
                          │            │              ClickHouse                   │
                          │            │    (Raw Events + Aggregated Tables)      │
                          │            └──────────────────────────────────────────┘
                          │
                          ▼
               ┌──────────────────────┐
               │     WebSocket Hub    │
               │   (Dashboard Feed)   │
               └──────────────────────┘
```

---

## Click Event Processing

### Event Schema

```go
// internal/analytics/events.go
package analytics

import (
	"time"
)

// ClickEvent represents a single click/redirect event
type ClickEvent struct {
	// Identifiers
	EventID     string `json:"event_id" ch:"event_id"`
	LinkID      string `json:"link_id" ch:"link_id"`
	ShortCode   string `json:"short_code" ch:"short_code"`
	WorkspaceID string `json:"workspace_id" ch:"workspace_id"`

	// Timing
	Timestamp   time.Time `json:"timestamp" ch:"timestamp"`

	// Request Data
	IPAddress   string `json:"ip_address" ch:"ip_address"`
	UserAgent   string `json:"user_agent" ch:"user_agent"`
	Referer     string `json:"referer" ch:"referer"`
	RefererHost string `json:"referer_host" ch:"referer_host"`

	// Geographic Data
	Country     string  `json:"country" ch:"country"`
	CountryCode string  `json:"country_code" ch:"country_code"`
	Region      string  `json:"region" ch:"region"`
	City        string  `json:"city" ch:"city"`
	Latitude    float64 `json:"latitude" ch:"latitude"`
	Longitude   float64 `json:"longitude" ch:"longitude"`
	Timezone    string  `json:"timezone" ch:"timezone"`

	// Device Data
	DeviceType  string `json:"device_type" ch:"device_type"`   // desktop, mobile, tablet
	Browser     string `json:"browser" ch:"browser"`
	BrowserVer  string `json:"browser_version" ch:"browser_version"`
	OS          string `json:"os" ch:"os"`
	OSVersion   string `json:"os_version" ch:"os_version"`

	// Classification
	IsBot       bool   `json:"is_bot" ch:"is_bot"`
	BotName     string `json:"bot_name" ch:"bot_name"`
	IsUnique    bool   `json:"is_unique" ch:"is_unique"`

	// UTM Parameters
	UTMSource   string `json:"utm_source" ch:"utm_source"`
	UTMMedium   string `json:"utm_medium" ch:"utm_medium"`
	UTMCampaign string `json:"utm_campaign" ch:"utm_campaign"`
	UTMTerm     string `json:"utm_term" ch:"utm_term"`
	UTMContent  string `json:"utm_content" ch:"utm_content"`
}

// ClickEventBatch represents a batch of events for bulk processing
type ClickEventBatch struct {
	Events    []*ClickEvent
	Partition int32
	Offset    int64
}
```

### Event Ingestion

```go
// internal/analytics/ingestion.go
package analytics

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/link-rift/link-rift/internal/config"
)

// EventIngester handles click event ingestion to Kafka
type EventIngester struct {
	writer      *kafka.Writer
	topic       string
	batchSize   int
	buffer      []*ClickEvent
	bufferMu    sync.Mutex
	flushTicker *time.Ticker
	quit        chan struct{}
}

// NewEventIngester creates a new event ingester
func NewEventIngester(cfg *config.KafkaConfig) *EventIngester {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.ClicksTopic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    cfg.BatchSize,
		BatchTimeout: cfg.BatchTimeout,
		Async:        true,
		Compression:  kafka.Lz4,
		RequiredAcks: kafka.RequireOne,
	}

	ei := &EventIngester{
		writer:    writer,
		topic:     cfg.ClicksTopic,
		batchSize: cfg.BatchSize,
		buffer:    make([]*ClickEvent, 0, cfg.BatchSize),
		quit:      make(chan struct{}),
	}

	ei.flushTicker = time.NewTicker(time.Second)
	go ei.flushLoop()

	return ei
}

// Ingest adds an event to the ingestion buffer
func (ei *EventIngester) Ingest(event *ClickEvent) error {
	ei.bufferMu.Lock()
	ei.buffer = append(ei.buffer, event)
	shouldFlush := len(ei.buffer) >= ei.batchSize
	ei.bufferMu.Unlock()

	if shouldFlush {
		return ei.Flush()
	}
	return nil
}

// Flush sends buffered events to Kafka
func (ei *EventIngester) Flush() error {
	ei.bufferMu.Lock()
	if len(ei.buffer) == 0 {
		ei.bufferMu.Unlock()
		return nil
	}

	events := ei.buffer
	ei.buffer = make([]*ClickEvent, 0, ei.batchSize)
	ei.bufferMu.Unlock()

	messages := make([]kafka.Message, len(events))
	for i, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			continue
		}

		messages[i] = kafka.Message{
			Key:   []byte(event.LinkID),
			Value: data,
			Time:  event.Timestamp,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return ei.writer.WriteMessages(ctx, messages...)
}

func (ei *EventIngester) flushLoop() {
	for {
		select {
		case <-ei.flushTicker.C:
			ei.Flush()
		case <-ei.quit:
			return
		}
	}
}

// Close shuts down the ingester
func (ei *EventIngester) Close() error {
	close(ei.quit)
	ei.Flush()
	return ei.writer.Close()
}
```

---

## Real-Time vs Batch Processing

### Real-Time Stream

```go
// internal/analytics/realtime/consumer.go
package realtime

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/link-rift/link-rift/internal/analytics"
)

// StreamConsumer processes events for real-time dashboards
type StreamConsumer struct {
	reader       *kafka.Reader
	redis        *redis.Client
	hub          *WebSocketHub
	counters     *RealtimeCounters
	quit         chan struct{}
	wg           sync.WaitGroup
}

// NewStreamConsumer creates a new real-time stream consumer
func NewStreamConsumer(
	brokers []string,
	topic string,
	groupID string,
	redis *redis.Client,
	hub *WebSocketHub,
) *StreamConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1e3,  // 1KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
	})

	return &StreamConsumer{
		reader:   reader,
		redis:    redis,
		hub:      hub,
		counters: NewRealtimeCounters(redis),
		quit:     make(chan struct{}),
	}
}

// Start begins consuming events
func (sc *StreamConsumer) Start(workers int) {
	for i := 0; i < workers; i++ {
		sc.wg.Add(1)
		go sc.consume()
	}
}

func (sc *StreamConsumer) consume() {
	defer sc.wg.Done()

	for {
		select {
		case <-sc.quit:
			return
		default:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			msg, err := sc.reader.FetchMessage(ctx)
			cancel()

			if err != nil {
				continue
			}

			var event analytics.ClickEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				sc.reader.CommitMessages(context.Background(), msg)
				continue
			}

			sc.processEvent(&event)
			sc.reader.CommitMessages(context.Background(), msg)
		}
	}
}

func (sc *StreamConsumer) processEvent(event *analytics.ClickEvent) {
	ctx := context.Background()

	// Update real-time counters
	sc.counters.IncrementClicks(ctx, event.WorkspaceID, event.LinkID)
	sc.counters.IncrementByCountry(ctx, event.WorkspaceID, event.CountryCode)
	sc.counters.IncrementByDevice(ctx, event.WorkspaceID, event.DeviceType)
	sc.counters.IncrementByReferer(ctx, event.WorkspaceID, event.RefererHost)

	// Publish to WebSocket subscribers
	sc.hub.BroadcastToWorkspace(event.WorkspaceID, &WebSocketMessage{
		Type: "click",
		Data: event,
	})

	// Publish to link-specific subscribers
	sc.hub.BroadcastToLink(event.LinkID, &WebSocketMessage{
		Type: "click",
		Data: event,
	})
}

// Stop gracefully shuts down the consumer
func (sc *StreamConsumer) Stop() {
	close(sc.quit)
	sc.wg.Wait()
	sc.reader.Close()
}

// RealtimeCounters manages real-time analytics counters in Redis
type RealtimeCounters struct {
	redis *redis.Client
}

func NewRealtimeCounters(redis *redis.Client) *RealtimeCounters {
	return &RealtimeCounters{redis: redis}
}

func (rc *RealtimeCounters) IncrementClicks(ctx context.Context, workspaceID, linkID string) {
	now := time.Now()
	minute := now.Truncate(time.Minute).Unix()
	hour := now.Truncate(time.Hour).Unix()

	pipe := rc.redis.Pipeline()

	// Minute-level counters (kept for 1 hour)
	minuteKey := fmt.Sprintf("clicks:minute:%s:%d", workspaceID, minute)
	pipe.Incr(ctx, minuteKey)
	pipe.Expire(ctx, minuteKey, time.Hour)

	// Hour-level counters (kept for 24 hours)
	hourKey := fmt.Sprintf("clicks:hour:%s:%d", workspaceID, hour)
	pipe.Incr(ctx, hourKey)
	pipe.Expire(ctx, hourKey, 24*time.Hour)

	// Link-specific counters
	linkMinuteKey := fmt.Sprintf("clicks:minute:%s:%s:%d", workspaceID, linkID, minute)
	pipe.Incr(ctx, linkMinuteKey)
	pipe.Expire(ctx, linkMinuteKey, time.Hour)

	pipe.Exec(ctx)
}

func (rc *RealtimeCounters) IncrementByCountry(ctx context.Context, workspaceID, countryCode string) {
	key := fmt.Sprintf("clicks:country:%s:%s", workspaceID, time.Now().Format("2006-01-02"))
	rc.redis.ZIncrBy(ctx, key, 1, countryCode)
	rc.redis.Expire(ctx, key, 48*time.Hour)
}

func (rc *RealtimeCounters) IncrementByDevice(ctx context.Context, workspaceID, deviceType string) {
	key := fmt.Sprintf("clicks:device:%s:%s", workspaceID, time.Now().Format("2006-01-02"))
	rc.redis.ZIncrBy(ctx, key, 1, deviceType)
	rc.redis.Expire(ctx, key, 48*time.Hour)
}

func (rc *RealtimeCounters) IncrementByReferer(ctx context.Context, workspaceID, refererHost string) {
	if refererHost == "" {
		refererHost = "direct"
	}
	key := fmt.Sprintf("clicks:referer:%s:%s", workspaceID, time.Now().Format("2006-01-02"))
	rc.redis.ZIncrBy(ctx, key, 1, refererHost)
	rc.redis.Expire(ctx, key, 48*time.Hour)
}
```

### Batch Processing

```go
// internal/analytics/batch/processor.go
package batch

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/segmentio/kafka-go"
	"github.com/link-rift/link-rift/internal/analytics"
)

// BatchProcessor handles batch ingestion to ClickHouse
type BatchProcessor struct {
	reader    *kafka.Reader
	clickhouse clickhouse.Conn
	batchSize  int
	quit       chan struct{}
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(
	brokers []string,
	topic string,
	groupID string,
	ch clickhouse.Conn,
	batchSize int,
) *BatchProcessor {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   brokers,
		Topic:     topic,
		GroupID:   groupID,
		MinBytes:  10e3,  // 10KB
		MaxBytes:  100e6, // 100MB
	})

	return &BatchProcessor{
		reader:     reader,
		clickhouse: ch,
		batchSize:  batchSize,
		quit:       make(chan struct{}),
	}
}

// Start begins batch processing
func (bp *BatchProcessor) Start() {
	go bp.process()
}

func (bp *BatchProcessor) process() {
	batch := make([]*analytics.ClickEvent, 0, bp.batchSize)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-bp.quit:
			if len(batch) > 0 {
				bp.writeBatch(batch)
			}
			return
		case <-ticker.C:
			if len(batch) > 0 {
				bp.writeBatch(batch)
				batch = make([]*analytics.ClickEvent, 0, bp.batchSize)
			}
		default:
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			msg, err := bp.reader.FetchMessage(ctx)
			cancel()

			if err != nil {
				continue
			}

			var event analytics.ClickEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				bp.reader.CommitMessages(context.Background(), msg)
				continue
			}

			batch = append(batch, &event)

			if len(batch) >= bp.batchSize {
				bp.writeBatch(batch)
				batch = make([]*analytics.ClickEvent, 0, bp.batchSize)
			}

			bp.reader.CommitMessages(context.Background(), msg)
		}
	}
}

func (bp *BatchProcessor) writeBatch(events []*analytics.ClickEvent) error {
	ctx := context.Background()

	batch, err := bp.clickhouse.PrepareBatch(ctx, `
		INSERT INTO click_events (
			event_id, link_id, short_code, workspace_id, timestamp,
			ip_address, user_agent, referer, referer_host,
			country, country_code, region, city, latitude, longitude, timezone,
			device_type, browser, browser_version, os, os_version,
			is_bot, bot_name, is_unique,
			utm_source, utm_medium, utm_campaign, utm_term, utm_content
		)
	`)
	if err != nil {
		return err
	}

	for _, event := range events {
		err := batch.Append(
			event.EventID,
			event.LinkID,
			event.ShortCode,
			event.WorkspaceID,
			event.Timestamp,
			event.IPAddress,
			event.UserAgent,
			event.Referer,
			event.RefererHost,
			event.Country,
			event.CountryCode,
			event.Region,
			event.City,
			event.Latitude,
			event.Longitude,
			event.Timezone,
			event.DeviceType,
			event.Browser,
			event.BrowserVer,
			event.OS,
			event.OSVersion,
			event.IsBot,
			event.BotName,
			event.IsUnique,
			event.UTMSource,
			event.UTMMedium,
			event.UTMCampaign,
			event.UTMTerm,
			event.UTMContent,
		)
		if err != nil {
			return err
		}
	}

	return batch.Send()
}

// Stop gracefully shuts down the processor
func (bp *BatchProcessor) Stop() {
	close(bp.quit)
	bp.reader.Close()
}
```

---

## ClickHouse Schema

### Main Events Table

```sql
-- migrations/clickhouse/001_click_events.sql

-- Main click events table with partitioning by month
CREATE TABLE IF NOT EXISTS click_events (
    event_id        UUID DEFAULT generateUUIDv4(),
    link_id         String,
    short_code      LowCardinality(String),
    workspace_id    String,
    timestamp       DateTime64(3),

    -- Request data
    ip_address      IPv6,
    user_agent      String,
    referer         String,
    referer_host    LowCardinality(String),

    -- Geographic data
    country         LowCardinality(String),
    country_code    LowCardinality(FixedString(2)),
    region          LowCardinality(String),
    city            LowCardinality(String),
    latitude        Float32,
    longitude       Float32,
    timezone        LowCardinality(String),

    -- Device data
    device_type     LowCardinality(String),  -- desktop, mobile, tablet
    browser         LowCardinality(String),
    browser_version LowCardinality(String),
    os              LowCardinality(String),
    os_version      LowCardinality(String),

    -- Classification
    is_bot          UInt8,
    bot_name        LowCardinality(String),
    is_unique       UInt8,

    -- UTM Parameters
    utm_source      LowCardinality(String),
    utm_medium      LowCardinality(String),
    utm_campaign    LowCardinality(String),
    utm_term        String,
    utm_content     String,

    -- Computed columns
    date            Date DEFAULT toDate(timestamp),
    hour            UInt8 DEFAULT toHour(timestamp)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (workspace_id, link_id, timestamp)
TTL timestamp + INTERVAL 2 YEAR
SETTINGS index_granularity = 8192;

-- Index for workspace-level queries
ALTER TABLE click_events ADD INDEX idx_workspace workspace_id TYPE bloom_filter GRANULARITY 1;

-- Index for link-level queries
ALTER TABLE click_events ADD INDEX idx_link link_id TYPE bloom_filter GRANULARITY 1;

-- Index for country queries
ALTER TABLE click_events ADD INDEX idx_country country_code TYPE set(100) GRANULARITY 1;
```

### Materialized Views

```sql
-- migrations/clickhouse/002_materialized_views.sql

-- Hourly aggregation by link
CREATE MATERIALIZED VIEW IF NOT EXISTS clicks_hourly_mv
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(hour)
ORDER BY (workspace_id, link_id, hour)
AS SELECT
    workspace_id,
    link_id,
    toStartOfHour(timestamp) AS hour,
    count() AS total_clicks,
    countIf(is_unique = 1) AS unique_clicks,
    countIf(is_bot = 0) AS human_clicks,
    countIf(device_type = 'mobile') AS mobile_clicks,
    countIf(device_type = 'desktop') AS desktop_clicks,
    countIf(device_type = 'tablet') AS tablet_clicks
FROM click_events
GROUP BY workspace_id, link_id, hour;

-- Daily aggregation by link
CREATE MATERIALIZED VIEW IF NOT EXISTS clicks_daily_mv
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (workspace_id, link_id, date)
AS SELECT
    workspace_id,
    link_id,
    toDate(timestamp) AS date,
    count() AS total_clicks,
    countIf(is_unique = 1) AS unique_clicks,
    countIf(is_bot = 0) AS human_clicks,
    uniqExact(country_code) AS unique_countries,
    uniqExact(referer_host) AS unique_referers
FROM click_events
GROUP BY workspace_id, link_id, date;

-- Country aggregation
CREATE MATERIALIZED VIEW IF NOT EXISTS clicks_by_country_mv
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (workspace_id, date, country_code)
AS SELECT
    workspace_id,
    toDate(timestamp) AS date,
    country_code,
    country,
    count() AS clicks,
    countIf(is_unique = 1) AS unique_clicks
FROM click_events
WHERE is_bot = 0
GROUP BY workspace_id, date, country_code, country;

-- Referer aggregation
CREATE MATERIALIZED VIEW IF NOT EXISTS clicks_by_referer_mv
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (workspace_id, link_id, date, referer_host)
AS SELECT
    workspace_id,
    link_id,
    toDate(timestamp) AS date,
    referer_host,
    count() AS clicks
FROM click_events
WHERE is_bot = 0
GROUP BY workspace_id, link_id, date, referer_host;

-- Device aggregation
CREATE MATERIALIZED VIEW IF NOT EXISTS clicks_by_device_mv
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (workspace_id, date, device_type, browser, os)
AS SELECT
    workspace_id,
    toDate(timestamp) AS date,
    device_type,
    browser,
    os,
    count() AS clicks
FROM click_events
WHERE is_bot = 0
GROUP BY workspace_id, date, device_type, browser, os;
```

---

## Aggregation Jobs

```go
// internal/analytics/jobs/aggregator.go
package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/robfig/cron/v3"
)

// Aggregator runs scheduled aggregation jobs
type Aggregator struct {
	clickhouse clickhouse.Conn
	cron       *cron.Cron
}

// NewAggregator creates a new aggregation job runner
func NewAggregator(ch clickhouse.Conn) *Aggregator {
	return &Aggregator{
		clickhouse: ch,
		cron:       cron.New(cron.WithSeconds()),
	}
}

// Start begins the aggregation scheduler
func (a *Aggregator) Start() error {
	// Run hourly roll-up at minute 5
	_, err := a.cron.AddFunc("0 5 * * * *", a.hourlyRollup)
	if err != nil {
		return err
	}

	// Run daily roll-up at 00:30
	_, err = a.cron.AddFunc("0 30 0 * * *", a.dailyRollup)
	if err != nil {
		return err
	}

	// Run monthly cleanup at 3 AM on the 1st
	_, err = a.cron.AddFunc("0 0 3 1 * *", a.monthlyCleanup)
	if err != nil {
		return err
	}

	a.cron.Start()
	return nil
}

func (a *Aggregator) hourlyRollup() {
	ctx := context.Background()
	lastHour := time.Now().Add(-time.Hour).Truncate(time.Hour)

	// Optimize materialized view tables
	err := a.clickhouse.Exec(ctx, fmt.Sprintf(`
		OPTIMIZE TABLE clicks_hourly_mv
		PARTITION %d
		FINAL
	`, lastHour.Year()*100+int(lastHour.Month())))

	if err != nil {
		// Log error
	}
}

func (a *Aggregator) dailyRollup() {
	ctx := context.Background()
	yesterday := time.Now().AddDate(0, 0, -1)

	// Create daily summary table
	err := a.clickhouse.Exec(ctx, fmt.Sprintf(`
		INSERT INTO daily_link_stats
		SELECT
			workspace_id,
			link_id,
			toDate('%s') as date,
			sum(total_clicks) as total_clicks,
			sum(unique_clicks) as unique_clicks,
			sum(human_clicks) as human_clicks,
			sum(mobile_clicks) as mobile_clicks,
			sum(desktop_clicks) as desktop_clicks,
			sum(tablet_clicks) as tablet_clicks
		FROM clicks_hourly_mv
		WHERE hour >= toDateTime('%s 00:00:00')
		AND hour < toDateTime('%s 00:00:00')
		GROUP BY workspace_id, link_id
	`, yesterday.Format("2006-01-02"),
		yesterday.Format("2006-01-02"),
		time.Now().Format("2006-01-02")))

	if err != nil {
		// Log error
	}

	// Optimize daily tables
	a.clickhouse.Exec(ctx, `OPTIMIZE TABLE clicks_daily_mv FINAL`)
	a.clickhouse.Exec(ctx, `OPTIMIZE TABLE clicks_by_country_mv FINAL`)
	a.clickhouse.Exec(ctx, `OPTIMIZE TABLE clicks_by_referer_mv FINAL`)
	a.clickhouse.Exec(ctx, `OPTIMIZE TABLE clicks_by_device_mv FINAL`)
}

func (a *Aggregator) monthlyCleanup() {
	ctx := context.Background()

	// Drop old partitions (older than 2 years)
	cutoff := time.Now().AddDate(-2, 0, 0)
	partition := cutoff.Year()*100 + int(cutoff.Month())

	a.clickhouse.Exec(ctx, fmt.Sprintf(`
		ALTER TABLE click_events
		DROP PARTITION %d
	`, partition))
}

// Stop shuts down the aggregator
func (a *Aggregator) Stop() {
	a.cron.Stop()
}
```

---

## Geographic Enrichment

```go
// internal/analytics/geo/locator.go
package geo

import (
	"net"
	"sync"

	"github.com/oschwald/geoip2-golang"
)

// GeoData represents geographic information
type GeoData struct {
	Country     string
	CountryCode string
	Region      string
	City        string
	Latitude    float64
	Longitude   float64
	Timezone    string
}

// GeoLocator provides IP geolocation using MaxMind GeoIP2
type GeoLocator struct {
	cityDB *geoip2.Reader
	asnDB  *geoip2.Reader
	cache  sync.Map // LRU cache for frequent IPs
}

// NewGeoLocator creates a new geo locator
func NewGeoLocator(cityDBPath, asnDBPath string) (*GeoLocator, error) {
	cityDB, err := geoip2.Open(cityDBPath)
	if err != nil {
		return nil, err
	}

	var asnDB *geoip2.Reader
	if asnDBPath != "" {
		asnDB, err = geoip2.Open(asnDBPath)
		if err != nil {
			cityDB.Close()
			return nil, err
		}
	}

	return &GeoLocator{
		cityDB: cityDB,
		asnDB:  asnDB,
	}, nil
}

// Lookup returns geographic data for an IP address
func (gl *GeoLocator) Lookup(ipStr string) *GeoData {
	// Check cache first
	if cached, ok := gl.cache.Load(ipStr); ok {
		return cached.(*GeoData)
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil
	}

	record, err := gl.cityDB.City(ip)
	if err != nil {
		return nil
	}

	geo := &GeoData{
		Country:     record.Country.Names["en"],
		CountryCode: record.Country.IsoCode,
		Latitude:    record.Location.Latitude,
		Longitude:   record.Location.Longitude,
		Timezone:    record.Location.TimeZone,
	}

	if len(record.Subdivisions) > 0 {
		geo.Region = record.Subdivisions[0].Names["en"]
	}

	if record.City.Names != nil {
		geo.City = record.City.Names["en"]
	}

	// Cache the result
	gl.cache.Store(ipStr, geo)

	return geo
}

// LookupASN returns ASN information for an IP
func (gl *GeoLocator) LookupASN(ipStr string) (uint, string) {
	if gl.asnDB == nil {
		return 0, ""
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 0, ""
	}

	record, err := gl.asnDB.ASN(ip)
	if err != nil {
		return 0, ""
	}

	return record.AutonomousSystemNumber, record.AutonomousSystemOrganization
}

// Close releases resources
func (gl *GeoLocator) Close() error {
	if gl.cityDB != nil {
		gl.cityDB.Close()
	}
	if gl.asnDB != nil {
		gl.asnDB.Close()
	}
	return nil
}
```

---

## Real-Time Dashboard WebSocket

```go
// internal/analytics/realtime/websocket.go
package realtime

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// WebSocketMessage represents a message sent to clients
type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Client represents a WebSocket client
type Client struct {
	conn        *websocket.Conn
	workspaceID string
	linkIDs     map[string]bool
	send        chan *WebSocketMessage
	hub         *WebSocketHub
}

// WebSocketHub manages WebSocket connections
type WebSocketHub struct {
	// Registered clients by workspace
	workspaceClients map[string]map[*Client]bool
	// Registered clients by link
	linkClients map[string]map[*Client]bool

	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage

	mu sync.RWMutex
}

type BroadcastMessage struct {
	WorkspaceID string
	LinkID      string
	Message     *WebSocketMessage
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub() *WebSocketHub {
	hub := &WebSocketHub{
		workspaceClients: make(map[string]map[*Client]bool),
		linkClients:      make(map[string]map[*Client]bool),
		register:         make(chan *Client),
		unregister:       make(chan *Client),
		broadcast:        make(chan *BroadcastMessage, 1000),
	}
	go hub.run()
	return hub
}

func (h *WebSocketHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			// Register for workspace
			if h.workspaceClients[client.workspaceID] == nil {
				h.workspaceClients[client.workspaceID] = make(map[*Client]bool)
			}
			h.workspaceClients[client.workspaceID][client] = true

			// Register for links
			for linkID := range client.linkIDs {
				if h.linkClients[linkID] == nil {
					h.linkClients[linkID] = make(map[*Client]bool)
				}
				h.linkClients[linkID][client] = true
			}
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			// Unregister from workspace
			if clients, ok := h.workspaceClients[client.workspaceID]; ok {
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.workspaceClients, client.workspaceID)
				}
			}

			// Unregister from links
			for linkID := range client.linkIDs {
				if clients, ok := h.linkClients[linkID]; ok {
					delete(clients, client)
					if len(clients) == 0 {
						delete(h.linkClients, linkID)
					}
				}
			}
			h.mu.Unlock()
			close(client.send)

		case msg := <-h.broadcast:
			h.mu.RLock()
			if msg.LinkID != "" {
				// Broadcast to link subscribers
				if clients, ok := h.linkClients[msg.LinkID]; ok {
					for client := range clients {
						select {
						case client.send <- msg.Message:
						default:
							// Client buffer full, skip
						}
					}
				}
			} else if msg.WorkspaceID != "" {
				// Broadcast to workspace subscribers
				if clients, ok := h.workspaceClients[msg.WorkspaceID]; ok {
					for client := range clients {
						select {
						case client.send <- msg.Message:
						default:
						}
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToWorkspace sends a message to all clients in a workspace
func (h *WebSocketHub) BroadcastToWorkspace(workspaceID string, msg *WebSocketMessage) {
	h.broadcast <- &BroadcastMessage{
		WorkspaceID: workspaceID,
		Message:     msg,
	}
}

// BroadcastToLink sends a message to all clients subscribed to a link
func (h *WebSocketHub) BroadcastToLink(linkID string, msg *WebSocketMessage) {
	h.broadcast <- &BroadcastMessage{
		LinkID:  linkID,
		Message: msg,
	}
}

// HandleWebSocket handles WebSocket connections
func (h *WebSocketHub) HandleWebSocket(c *websocket.Conn) {
	workspaceID := c.Locals("workspaceID").(string)

	client := &Client{
		conn:        c,
		workspaceID: workspaceID,
		linkIDs:     make(map[string]bool),
		send:        make(chan *WebSocketMessage, 256),
		hub:         h,
	}

	h.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var msg struct {
			Type   string   `json:"type"`
			LinkID string   `json:"link_id,omitempty"`
			Links  []string `json:"links,omitempty"`
		}

		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "subscribe_link":
			c.linkIDs[msg.LinkID] = true
			c.hub.mu.Lock()
			if c.hub.linkClients[msg.LinkID] == nil {
				c.hub.linkClients[msg.LinkID] = make(map[*Client]bool)
			}
			c.hub.linkClients[msg.LinkID][c] = true
			c.hub.mu.Unlock()

		case "unsubscribe_link":
			delete(c.linkIDs, msg.LinkID)
			c.hub.mu.Lock()
			if clients, ok := c.hub.linkClients[msg.LinkID]; ok {
				delete(clients, c)
			}
			c.hub.mu.Unlock()
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				continue
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
```

### React WebSocket Hook

```typescript
// src/hooks/useRealtimeAnalytics.ts
import { useEffect, useCallback, useState, useRef } from 'react';
import { useAuth } from '@/contexts/AuthContext';

interface ClickEvent {
  event_id: string;
  link_id: string;
  short_code: string;
  timestamp: string;
  country: string;
  country_code: string;
  city: string;
  device_type: string;
  browser: string;
  referer_host: string;
}

interface UseRealtimeAnalyticsOptions {
  workspaceId: string;
  linkIds?: string[];
  onClickEvent?: (event: ClickEvent) => void;
}

export function useRealtimeAnalytics({
  workspaceId,
  linkIds = [],
  onClickEvent,
}: UseRealtimeAnalyticsOptions) {
  const { accessToken } = useAuth();
  const [isConnected, setIsConnected] = useState(false);
  const [recentClicks, setRecentClicks] = useState<ClickEvent[]>([]);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>();

  const connect = useCallback(() => {
    if (!accessToken) return;

    const wsUrl = `${import.meta.env.VITE_WS_URL}/analytics/realtime?token=${accessToken}`;
    const ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      setIsConnected(true);

      // Subscribe to specific links
      if (linkIds.length > 0) {
        linkIds.forEach(linkId => {
          ws.send(JSON.stringify({ type: 'subscribe_link', link_id: linkId }));
        });
      }
    };

    ws.onmessage = (event) => {
      const message = JSON.parse(event.data);

      if (message.type === 'click') {
        const clickEvent = message.data as ClickEvent;

        setRecentClicks(prev => [clickEvent, ...prev].slice(0, 100));
        onClickEvent?.(clickEvent);
      }
    };

    ws.onclose = () => {
      setIsConnected(false);
      // Reconnect after 3 seconds
      reconnectTimeoutRef.current = setTimeout(connect, 3000);
    };

    ws.onerror = () => {
      ws.close();
    };

    wsRef.current = ws;
  }, [accessToken, linkIds, onClickEvent]);

  useEffect(() => {
    connect();

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      wsRef.current?.close();
    };
  }, [connect]);

  const subscribeToLink = useCallback((linkId: string) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: 'subscribe_link', link_id: linkId }));
    }
  }, []);

  const unsubscribeFromLink = useCallback((linkId: string) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: 'unsubscribe_link', link_id: linkId }));
    }
  }, []);

  return {
    isConnected,
    recentClicks,
    subscribeToLink,
    unsubscribeFromLink,
  };
}
```

---

## API Reference

### Analytics Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/analytics/overview` | Get workspace analytics overview |
| GET | `/api/v1/analytics/links/:id` | Get link-specific analytics |
| GET | `/api/v1/analytics/clicks` | Get click event details |
| GET | `/api/v1/analytics/countries` | Get geographic breakdown |
| GET | `/api/v1/analytics/devices` | Get device breakdown |
| GET | `/api/v1/analytics/referers` | Get referer breakdown |
| GET | `/api/v1/analytics/timeseries` | Get time-series data |
| WS | `/ws/analytics/realtime` | Real-time analytics WebSocket |

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `start_date` | string | Start date (ISO 8601) |
| `end_date` | string | End date (ISO 8601) |
| `link_id` | string | Filter by link ID |
| `interval` | string | Aggregation interval (hour, day, week, month) |
| `limit` | int | Maximum results to return |
| `offset` | int | Pagination offset |
