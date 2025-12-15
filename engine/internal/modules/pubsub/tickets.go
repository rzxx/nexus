package pubsub

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type TicketInfo struct {
	UserID    string
	Channels  []string
	ExpiresAt int64
}

type TicketStore struct {
	tickets map[string]TicketInfo
	mu      sync.RWMutex
}

func NewTicketStore() *TicketStore {
	return &TicketStore{
		tickets: make(map[string]TicketInfo),
	}
}

// Create генерирует случайный токен
func (ts *TicketStore) Create(userID string, channels []string, ttl time.Duration) string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	token := hex.EncodeToString(bytes)

	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.tickets[token] = TicketInfo{
		UserID:    userID,
		Channels:  channels,
		ExpiresAt: time.Now().Add(ttl).Unix(),
	}

	// В реальном проде здесь нужен cleanup горутина, но для MVP хватит проверки при чтении
	return token
}

// Validate проверяет и СЖИГАЕТ тикет (он одноразовый)
func (ts *TicketStore) Validate(token string) (TicketInfo, bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	info, ok := ts.tickets[token]
	if !ok {
		return TicketInfo{}, false
	}

	// Удаляем сразу (Ticket Pattern предполагает One-Time Use)
	delete(ts.tickets, token)

	if time.Now().Unix() > info.ExpiresAt {
		return TicketInfo{}, false
	}

	return info, true
}
