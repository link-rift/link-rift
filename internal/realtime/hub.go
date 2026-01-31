package realtime

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/link-rift/link-rift/internal/models"
	"go.uber.org/zap"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingInterval   = 30 * time.Second
	maxMessageSize = 512
)

// Hub manages WebSocket client connections and broadcasts.
type Hub struct {
	mu                sync.RWMutex
	workspaceClients  map[uuid.UUID]map[*Client]bool
	linkClients       map[uuid.UUID]map[*Client]bool
	register          chan *Client
	unregister        chan *Client
	logger            *zap.Logger
}

func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		workspaceClients: make(map[uuid.UUID]map[*Client]bool),
		linkClients:      make(map[uuid.UUID]map[*Client]bool),
		register:         make(chan *Client, 64),
		unregister:       make(chan *Client, 64),
		logger:           logger,
	}
}

// Run starts the hub event loop. Call in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.addClient(client)
		case client := <-h.unregister:
			h.removeClient(client)
		}
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) addClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.workspaceClients[c.WorkspaceID] == nil {
		h.workspaceClients[c.WorkspaceID] = make(map[*Client]bool)
	}
	h.workspaceClients[c.WorkspaceID][c] = true
}

func (h *Hub) removeClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.workspaceClients[c.WorkspaceID]; ok {
		delete(clients, c)
		if len(clients) == 0 {
			delete(h.workspaceClients, c.WorkspaceID)
		}
	}

	for linkID := range c.subscribedLinks {
		if clients, ok := h.linkClients[linkID]; ok {
			delete(clients, c)
			if len(clients) == 0 {
				delete(h.linkClients, linkID)
			}
		}
	}

	close(c.send)
}

// SubscribeLink subscribes a client to link-level broadcasts.
func (h *Hub) SubscribeLink(c *Client, linkID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	c.subscribedLinks[linkID] = true
	if h.linkClients[linkID] == nil {
		h.linkClients[linkID] = make(map[*Client]bool)
	}
	h.linkClients[linkID][c] = true
}

// UnsubscribeLink removes a client from link-level broadcasts.
func (h *Hub) UnsubscribeLink(c *Client, linkID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(c.subscribedLinks, linkID)
	if clients, ok := h.linkClients[linkID]; ok {
		delete(clients, c)
		if len(clients) == 0 {
			delete(h.linkClients, linkID)
		}
	}
}

// BroadcastToWorkspace sends a message to all clients in a workspace.
func (h *Hub) BroadcastToWorkspace(workspaceID uuid.UUID, notification *models.ClickNotification) {
	data, err := json.Marshal(map[string]any{
		"type": "click",
		"data": notification,
	})
	if err != nil {
		h.logger.Warn("failed to marshal notification", zap.Error(err))
		return
	}

	h.mu.RLock()
	clients := h.workspaceClients[workspaceID]
	h.mu.RUnlock()

	for c := range clients {
		select {
		case c.send <- data:
		default:
			// Client buffer full, skip
		}
	}
}

// BroadcastToLink sends a message to all clients subscribed to a specific link.
func (h *Hub) BroadcastToLink(linkID uuid.UUID, notification *models.ClickNotification) {
	data, err := json.Marshal(map[string]any{
		"type": "click",
		"data": notification,
	})
	if err != nil {
		h.logger.Warn("failed to marshal notification", zap.Error(err))
		return
	}

	h.mu.RLock()
	clients := h.linkClients[linkID]
	h.mu.RUnlock()

	for c := range clients {
		select {
		case c.send <- data:
		default:
		}
	}
}

// Client represents a single WebSocket connection.
type Client struct {
	hub             *Hub
	conn            *websocket.Conn
	WorkspaceID     uuid.UUID
	subscribedLinks map[uuid.UUID]bool
	send            chan []byte
}

func NewClient(hub *Hub, conn *websocket.Conn, workspaceID uuid.UUID) *Client {
	return &Client{
		hub:             hub,
		conn:            conn,
		WorkspaceID:     workspaceID,
		subscribedLinks: make(map[uuid.UUID]bool),
		send:            make(chan []byte, 256),
	}
}

// ReadPump reads messages from the WebSocket connection.
// It handles subscribe/unsubscribe commands and ping/pong.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var msg struct {
			Action string `json:"action"`
			LinkID string `json:"link_id"`
		}
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		linkID, err := uuid.Parse(msg.LinkID)
		if err != nil {
			continue
		}

		switch msg.Action {
		case "subscribe":
			c.hub.SubscribeLink(c, linkID)
		case "unsubscribe":
			c.hub.UnsubscribeLink(c, linkID)
		}
	}
}

// WritePump sends messages from the send channel to the WebSocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
