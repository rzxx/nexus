import { Hono } from "hono";
import { cors } from "hono/cors";
import { z } from "zod";
import { zValidator } from "@hono/zod-validator";
import { createNexus } from "@nexus/sdk";

const app = new Hono();
const nexus = createNexus();

app.use("/*", cors());

// Zod схема (Source of Truth)
const userSchema = z.object({
  username: z.string(),
  email: z.email(),
  role: z.enum(["admin", "user", "guest"]), // Добавим Enum для проверки типов
});

// Экспорт типа User для использования на фронтенде
export type User = z.infer<typeof userSchema>;

// --- API ROUTES ---
// Мы чейним роуты, чтобы TS мог вывести их общий тип
const routes = app
  .basePath("/api") // Все API теперь начинаются с /api
  .post("/users", zValidator("json", userSchema), async (c) => {
    const data = c.req.valid("json");
    await nexus.kv.set(`user:${data.email}`, data);
    return c.json({ success: true, user: data });
  })
  .get("/users/:email", async (c) => {
    const email = c.req.param("email");
    const user = await nexus.kv.get<z.infer<typeof userSchema>>(
      `user:${email}`
    );

    if (!user) return c.json({ error: "Not found" }, 404);

    return c.json({ user });
  });

// ⚠️ ГЛАВНАЯ МАГИЯ: Экспорт типа роутера
export type AppType = typeof routes;

export default routes;
