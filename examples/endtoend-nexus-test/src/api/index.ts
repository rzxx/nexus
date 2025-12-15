import { Hono } from "hono";
import { cors } from "hono/cors";
import { z } from "zod";
import { zValidator } from "@hono/zod-validator";
import { createNexus } from "@nexus/sdk";

const app = new Hono();
const nexus = createNexus();

app.use("/*", cors());

// Zod схемы (Source of Truth)
const userSchema = z.object({
  username: z.string(),
  email: z.email(),
  role: z.enum(["admin", "user", "guest"]), // Добавим Enum для проверки типов
});

const messageSchema = z.object({
  id: z.uuid(),
  text: z.string().min(1),
  username: z.string(), // В реальном приложении это берется из Session/JWT
});

// Экспорт типов для использования на фронтенде
export type User = z.infer<typeof userSchema>;
export type messageSchema = z.infer<typeof messageSchema>;

// --- API ROUTES ---

// 1. Чат (Sub-App)
const chatApp = new Hono()
  .post(
    "/auth",
    zValidator("json", z.object({ username: z.string() })),
    async (c) => {
      const { username } = c.req.valid("json");
      const ticket = await nexus.ws.createTicket({
        userId: username,
        channels: ["global-chat"],
      });
      return c.json({ ticket, wsUrl: "ws://localhost:4000/ws" });
    }
  )
  .post("/send", zValidator("json", messageSchema), async (c) => {
    const msg = c.req.valid("json");
    await nexus.ws.publish("global-chat", {
      ...msg,
      timestamp: Date.now(),
    });
    return c.json({ success: true });
  });

// 2. Пользователи (Sub-App)
const userApp = new Hono()
  .post("/", zValidator("json", userSchema), async (c) => {
    const data = c.req.valid("json");
    await nexus.kv.set(`user:${data.email}`, data);
    return c.json({ success: true, user: data });
  })
  .get("/:email", async (c) => {
    const email = c.req.param("email");
    const user = await nexus.kv.get<z.infer<typeof userSchema>>(
      `user:${email}`
    );
    if (!user) return c.json({ error: "Not found" }, 404);
    return c.json({ user });
  });

// --- MOUNTING (Сборка) ---

// Монтируем саб-аппы по нужным путям
// Это создает четкую структуру типов для клиента
const routes = app
  .route("/api/users", userApp) // Доступно как client.api.users
  .route("/api/chat", chatApp); // Доступно как client.api.chat

// Экспорт типа роутера для фронтенда
export type AppType = typeof routes;

export default routes;
