import { useState, useEffect, useRef } from "react";
import { hc } from "hono/client";
import type { AppType } from "./api";

const client = hc<AppType>("/");

interface Message {
  id: string;
  username: string;
  text: string;
  timestamp: number;
}

export function Chat({ username }: { username: string }) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputValue, setInputValue] = useState("");
  const [status, setStatus] = useState<
    "disconnected" | "connecting" | "connected"
  >("disconnected");

  // Ref не нужен для закрытия, если делать правильно,
  // но может пригодиться для дебага
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    // 1. Флаг, который скажет нам, жив ли компонент
    let isMounted = true;
    let ws: WebSocket | null = null;

    const connect = async () => {
      setStatus("connecting");

      try {
        const res = await client.api.chat.auth.$post({
          json: { username },
        });

        if (!res.ok) throw new Error("Failed to auth");

        // 2. ВАЖНАЯ ПРОВЕРКА:
        // Если пока мы ждали HTTP, компонент умер (StrictMode unmount),
        // то мы останавливаемся и ничего не открываем.
        if (!isMounted) return;

        const { ticket, wsUrl } = await res.json();

        // Создаем сокет
        ws = new WebSocket(`${wsUrl}?ticket=${ticket}`);
        wsRef.current = ws;

        ws.onopen = () => {
          if (isMounted) setStatus("connected");
        };

        ws.onmessage = (event) => {
          if (!isMounted) return;
          const msg = JSON.parse(event.data);
          const newMsg = msg.data as Message;

          setMessages((prev) => {
            // Защита от дублей по ID
            if (prev.some((m) => m.id === newMsg.id)) return prev;
            return [...prev, newMsg];
          });
        };

        ws.onclose = () => {
          if (isMounted) setStatus("disconnected");
        };
      } catch (e) {
        console.error(e);
        if (isMounted) setStatus("disconnected");
      }
    };

    connect();

    // 3. CLEANUP FUNCTION
    return () => {
      isMounted = false; // Отменяем все будущие действия
      if (ws) {
        ws.close(); // Закрываем сокет, если он успел создаться
      }
    };
  }, [username]);

  const sendMessage = async () => {
    if (!inputValue) return;
    const messageId = crypto.randomUUID();

    setMessages((prev) => [
      ...prev,
      {
        id: messageId,
        text: inputValue,
        username: username,
        timestamp: Date.now(),
      },
    ]);

    await client.api.chat.send.$post({
      json: {
        id: messageId,
        username,
        text: inputValue,
      },
    });

    setInputValue("");
  };

  return (
    <div
      style={{
        border: "1px solid #ccc",
        padding: 20,
        marginTop: 20,
        maxWidth: 400,
      }}>
      <h3>Chat (User: {username})</h3>
      <div
        style={{
          color: status === "connected" ? "green" : "red",
          marginBottom: 10,
        }}>
        Status: {status}
      </div>

      <div
        style={{
          height: 200,
          overflowY: "auto",
          background: "#f9f9f9",
          border: "1px solid #eee",
          marginBottom: 10,
          padding: 10,
        }}>
        {messages.map((m, i) => (
          <div key={i} style={{ marginBottom: 5 }}>
            <strong>{m.username}: </strong> {m.text}
          </div>
        ))}
      </div>

      <div style={{ display: "flex", gap: 5 }}>
        <input
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && sendMessage()}
          placeholder="Type a message..."
          style={{ flex: 1 }}
        />
        <button onClick={sendMessage}>Send</button>
      </div>
    </div>
  );
}
