package pubsub

import (
	"flag"
	"net/http"
	"nexus-engine/internal/core"
	"nexus-engine/internal/pkg/logger"
)

var _ core.Module = (*Module)(nil)

type Module struct {
	hub       *Hub
	tickets   *TicketStore
	log       *logger.Logger
	ticketTTL *int
}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string { return "PubSub" }

func (m *Module) RegisterFlags(fs *flag.FlagSet) {
	m.ticketTTL = fs.Int("ws-ticket-ttl", 15, "Ticket TTL in seconds")
}

func (m *Module) Init(log *logger.Logger) error {
	m.log = log
	m.hub = NewHub(log)
	m.tickets = NewTicketStore()

	// Запускаем Hub в отдельной горутине
	go m.hub.Run()

	return nil
}

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	// Internal API (для SDK)
	mux.HandleFunc("/pubsub/ticket", m.handleCreateTicket)
	mux.HandleFunc("/pubsub/publish", m.handlePublish)

	// Public WebSocket (для Клиентов)
	mux.HandleFunc("/ws", m.handleWebSocket)
}

func (m *Module) Shutdown() {
	// Можно добавить graceful shutdown для сокетов, но пока не обязательно
}
