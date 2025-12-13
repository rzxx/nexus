import { Hono } from "hono";
import { createNexus } from "@nexus/sdk";

const app = new Hono();
const nexus = createNexus(); // Магия: само находит Engine через ENV

// Бизнес-логика (Stateless)
app.post("/users", async (c) => {
  const body = await c.req.json();
  const { id, name } = body;

  // Логика (Stateful вызов через SDK)
  // TypeScript знает типы аргументов, потому что SDK типизирован
  await nexus.kv.set(`user:${id}`, { name, role: "user" }, { ttl: 3600 });

  return c.json({ message: "User created" });
});

app.get("/users/:id", async (c) => {
  const id = c.req.param("id");

  // Чтение
  const user = await nexus.kv.get<{ name: string; role: string }>(`user:${id}`);

  if (!user) {
    return c.json({ error: "User not found" }, 404);
  }

  return c.json(user);
});

export default app;
