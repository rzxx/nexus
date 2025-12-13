import { Hono } from "hono";
import { z } from "zod";
import { zValidator } from "@hono/zod-validator";
import { createNexus } from "@nexus/sdk";

// 1. Определяем схему данных (Zod)
// Это валидирует входящий JSON от пользователя
const userSchema = z.object({
  username: z.string().min(3),
  email: z.email(),
  isAdmin: z.boolean().default(false),
  tags: z.array(z.string()).optional(),
});

// 2. Выводим TypeScript тип из схемы
// Тип User автоматически совместим с JsonObject из SDK
type User = z.infer<typeof userSchema>;

const app = new Hono();
const nexus = createNexus(); // Читает NEXUS_ENGINE_URL из env

// --- POST: Создание пользователя ---
app.post(
  "/users/:id",
  zValidator("json", userSchema), // Hono проверит тело запроса
  async (c) => {
    const id = c.req.param("id");
    const userData = c.req.valid("json"); // userData строго типизирована как User

    // ✅ SDK.set:
    // Если мы попытаемся добавить поле, которого нет в JsonValue (напр. функцию),
    // TypeScript подчеркнет это красным еще до запуска.
    await nexus.kv.set(`user:${id}`, userData, { ttl: 3600 });

    return c.json({
      success: true,
      message: `User ${userData.username} saved`,
    });
  }
);

// --- GET: Получение пользователя ---
app.get("/users/:id", async (c) => {
  const id = c.req.param("id");

  // ✅ SDK.get:
  // Мы ОБЯЗАНЫ передать дженерик <User>.
  // Теперь переменная 'user' имеет тип User | null.
  const user = await nexus.kv.get<User>(`user:${id}`);

  if (!user) {
    return c.json({ error: "Not found" }, 404);
  }

  // TypeScript знает, что у user есть поле .email
  return c.json({
    id,
    email: user.email, // Автодополнение работает!
    username: user.username,
  });
});

export default app;
