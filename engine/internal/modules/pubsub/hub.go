package pubsub

import (
	"nexus-engine/internal/pkg/logger"
)

// Message — внутренняя структура сообщения
type Message struct {
	Channel string `json:"channel"`
	Data    any    `json:"data"`
}

type Hub struct {
	// Зарегистрированные клиенты
	clients map[*Client]bool

	// Подписки: channel -> map[client]bool
	subscriptions map[string]map[*Client]bool

	// Каналы управления (Action Channels)
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client

	log *logger.Logger
}

func NewHub(log *logger.Logger) *Hub {
	return &Hub{
		broadcast:     make(chan Message),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		clients:       make(map[*Client]bool),
		subscriptions: make(map[string]map[*Client]bool),
		log:           log,
	}
}

// Run запускает главный цикл обработки событий (Event Loop)
func (h *Hub) Run() {
	for {
		select {
		// 1. Подключение нового клиента
		case client := <-h.register:
			h.clients[client] = true
			// Подписываем клиента на его разрешенные каналы
			for _, ch := range client.channels {
				if _, ok := h.subscriptions[ch]; !ok {
					h.subscriptions[ch] = make(map[*Client]bool)
				}
				h.subscriptions[ch][client] = true
				h.log.Debug("Client subscribed to '%s'", ch)
			}

		// 2. Отключение клиента
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.removeClient(client)
			}

		// 3. Рассылка сообщения
		case msg := <-h.broadcast:
			clients, ok := h.subscriptions[msg.Channel]
			if !ok || len(clients) == 0 {
				continue // Нет подписчиков
			}

			for client := range clients {
				select {
				case client.send <- msg:
				default:
					// Если буфер клиента переполнен (он завис), отключаем его,
					// чтобы не блокировать остальных
					h.removeClient(client)
				}
			}
		}
	}
}

func (h *Hub) removeClient(client *Client) {
	delete(h.clients, client)
	close(client.send) // Закрываем канал, чтобы остановить writePump

	// Удаляем из всех подписок
	for _, ch := range client.channels {
		delete(h.subscriptions[ch], client)
		if len(h.subscriptions[ch]) == 0 {
			delete(h.subscriptions, ch) // Чистим память, если канал пуст
		}
	}
}
