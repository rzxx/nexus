import { useState } from "react";
import { hc } from "hono/client";
// 👇 Импортируем ТИПЫ прямо из бэкенд-файла
import type { AppType, User } from "./api/index";

// Создаем типизированный клиент
// Указываем пустой URL, так как Vite проксирует /api на тот же домен
const client = hc<AppType>("/");

function App() {
  const [email, setEmail] = useState("");
  // Используем экспортированный тип `User` из бэкенда
  const [userData, setUserData] = useState<User | null>(null);

  const saveUser = async () => {
    // 🪄 МАГИЯ ЗДЕСЬ 🪄
    // TypeScript знает, что $post принимает json с полями username, email, role.
    // Если ты напишешь role: "super-admin", TS подчеркнет это красным,
    // так как в Zod схеме есть только admin | user | guest.

    const res = await client.api.users.$post({
      json: {
        username: "NexusDev",
        email: email,
        role: "admin", // Попробуй поменять на "god-mode" и увидишь ошибку
      },
    });

    if (res.ok) {
      const data = await res.json();
      alert(`Saved: ${data.user.username}`);
    }
  };

  const loadUser = async () => {
    // Клиент знает, что параметр :email обязателен
    const res = await client.api.users[":email"].$get({
      param: { email: email },
    });

    if (res.ok) {
      const data = await res.json();
      setUserData(data.user);
    }
  };

  return (
    <div style={{ padding: 20 }}>
      <h1>Nexus Fullstack RPC</h1>

      <div style={{ marginBottom: 20 }}>
        <input
          placeholder="Email key"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          style={{ padding: 5, marginRight: 10 }}
        />
        <button onClick={saveUser}>Save (RPC)</button>
        <button onClick={loadUser}>Load (RPC)</button>
      </div>

      <pre>{JSON.stringify(userData, null, 2)}</pre>
    </div>
  );
}

export default App;
