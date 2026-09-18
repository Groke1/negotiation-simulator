package arena

import (
	"sync"

	"github.com/gorilla/websocket"
)

type wsClient struct {
	sessionID string
	userID    string
	conn      *websocket.Conn
	writeMu   sync.Mutex
}
type hub struct {
	mu       sync.RWMutex
	sessions map[string]map[*wsClient]struct{}
}

func newHub() *hub { return &hub{sessions: make(map[string]map[*wsClient]struct{})} }
func (h *hub) add(sessionID, userID string, conn *websocket.Conn) *wsClient {
	client := &wsClient{sessionID: sessionID, userID: userID, conn: conn}
	h.mu.Lock()
	if h.sessions[sessionID] == nil {
		h.sessions[sessionID] = make(map[*wsClient]struct{})
	}
	h.sessions[sessionID][client] = struct{}{}
	h.mu.Unlock()
	return client
}
func (h *hub) remove(client *wsClient) {
	h.mu.Lock()
	if clients := h.sessions[client.sessionID]; clients != nil {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.sessions, client.sessionID)
		}
	}
	h.mu.Unlock()
	_ = client.conn.Close()
}
func (h *hub) broadcast(sessionID string, event any) {
	h.mu.RLock()
	clients := make([]*wsClient, 0, len(h.sessions[sessionID]))
	for client := range h.sessions[sessionID] {
		clients = append(clients, client)
	}
	h.mu.RUnlock()
	for _, client := range clients {
		h.send(client, event)
	}
}
func (h *hub) send(client *wsClient, event any) {
	client.writeMu.Lock()
	defer client.writeMu.Unlock()
	_ = client.conn.WriteJSON(event)
}
