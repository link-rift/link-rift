package realtime

import (
	"context"
	"encoding/json"

	"github.com/link-rift/link-rift/internal/models"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const realtimeChannel = "clicks:realtime"

// StartRedisSubscriber subscribes to the Redis Pub/Sub channel for real-time click
// notifications and broadcasts them to the WebSocket hub.
func StartRedisSubscriber(ctx context.Context, redisClient *redis.Client, hub *Hub, logger *zap.Logger) {
	pubsub := redisClient.Subscribe(ctx, realtimeChannel)
	ch := pubsub.Channel()

	go func() {
		defer pubsub.Close()

		logger.Info("realtime Redis subscriber started", zap.String("channel", realtimeChannel))

		for {
			select {
			case <-ctx.Done():
				logger.Info("realtime Redis subscriber stopped")
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}

				var notification models.ClickNotification
				if err := json.Unmarshal([]byte(msg.Payload), &notification); err != nil {
					logger.Warn("failed to unmarshal click notification", zap.Error(err))
					continue
				}

				hub.BroadcastToWorkspace(notification.WorkspaceID, &notification)
				hub.BroadcastToLink(notification.LinkID, &notification)
			}
		}
	}()
}
