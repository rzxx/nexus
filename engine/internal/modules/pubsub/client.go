package pubsub

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Время на запись сообщения клиенту
	writeWait = 10 * time.Second
	// Время жизни без Ping-сообщений
	pongWait = 60 * time.Second
	// Как часто слать Ping (должно быть меньше pongWait)
	pingPeriod = (pongWait * 9) / 10
)

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan Message // Буферизированный канал для исходящих
	channels []string     // Список каналов, доступных клиенту
}

// readPump слушает входящие сообщения от клиента (Ping/Pong/Close)
// В Ticket Pattern мы обычно не ожидаем данных от клиента в сокет,
// но должны читать, чтобы обрабатывать Control Frames.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512) // Макс размер сообщения (нам нужны только пинги)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		// Читаем, но игнорируем контент (если мы не реализуем чат через сокеты)
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// writePump отправляет сообщения из канала send в сокет
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub закрыл канал
				c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}

			// Пишем JSON
			if err := c.conn.WriteJSON(msg); err != nil {
				return
			}

		case <-ticker.C:
			// Heartbeat
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
