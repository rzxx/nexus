package pubsub

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешаем все для Dev (в проде можно ужесточить)
	},
}

// --- INTERNAL API ---

func (m *Module) handleCreateTicket(w http.ResponseWriter, r *http.Request) {
	// Принимаем { userId, channels }
	var req struct {
		UserID   string   `json:"user_id"`
		Channels []string `json:"channels"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	ttl := time.Duration(*m.ticketTTL) * time.Second
	token := m.tickets.Create(req.UserID, req.Channels, ttl)

	json.NewEncoder(w).Encode(map[string]string{"ticket": token})
}

func (m *Module) handlePublish(w http.ResponseWriter, r *http.Request) {
	var req Message
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}

	// Отправляем в Hub (неблокирующе)
	m.hub.broadcast <- req

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}

// --- PUBLIC WEBSOCKET ---

func (m *Module) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	ticket := r.URL.Query().Get("ticket")
	if ticket == "" {
		http.Error(w, "Missing ticket", http.StatusUnauthorized)
		return
	}

	info, ok := m.tickets.Validate(ticket)
	if !ok {
		http.Error(w, "Invalid or expired ticket", http.StatusForbidden)
		return
	}

	// Upgrade HTTP -> WS
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		m.log.Error("WS Upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:      m.hub,
		conn:     conn,
		send:     make(chan Message, 256), // Буфер на 256 сообщений
		channels: info.Channels,
	}

	m.hub.register <- client

	// Запускаем процессы чтения/записи
	go client.writePump()
	go client.readPump()

	m.log.Debug("WS Connected: User=%s Channels=%v", info.UserID, info.Channels)
}
