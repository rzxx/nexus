import { spawn } from "bun";
import type { NexusConfig } from "../config";
import { logger } from "../logger";

export async function startDev(config: NexusConfig) {
  logger.info("Nexus", "Starting development server...");

  // --- 1. ENGINE (GO) ---

  const engineProc = spawn({
    cmd: [
      "go",
      "run",
      "cmd/nexus/main.go",
      "--port",
      config.engine.port.toString(),
      "--kv-data-dir",
      "data",
      "--log-level",
      "2",
    ],
    // Указываем явно, что cwd — строка (хотя она и так строка в конфиге)
    cwd: config.engine.path,
    stdout: "pipe",
    stderr: "pipe",
    // Явно приводим env к нужному типу, если нужно, или оставляем пустым для наследования
    env: { ...process.env } as Record<string, string>,
  });

  readStream(engineProc.stdout, "Engine");
  readStream(engineProc.stderr, "Engine");

  // Ждем старта (в будущем заменим на healthcheck)
  await new Promise((r) => setTimeout(r, 1000));

  // --- 2. APP (HONO) ---

  // Разбиваем команду и гарантируем, что первый элемент — строка
  const commandParts = config.app.command.split(" ");
  const cmd = commandParts[0];
  const args = commandParts.slice(1);

  if (!cmd) {
    logger.error("Nexus", "App command is empty!");
    engineProc.kill();
    return;
  }

  const appProc = spawn({
    cmd: [cmd, ...args],
    env: {
      ...(process.env as Record<string, string>),
      NEXUS_ENGINE_URL: `http://localhost:${config.engine.port}`,
      PORT: config.app.port.toString(),
    },
    stdout: "pipe",
    stderr: "pipe",
  });

  readStream(appProc.stdout, "App");
  readStream(appProc.stderr, "App");

  // --- 3. FRONTEND (Vite) ---
  let frontendProc: any = null;

  if (config.frontend) {
    const cmdParts = config.frontend.command.split(" ");
    const cmd = cmdParts[0];
    const args = cmdParts.slice(1);

    if (cmd) {
      frontendProc = spawn({
        cmd: [cmd, ...args],
        env: { ...process.env } as Record<string, string>,
        stdout: "pipe",
        stderr: "pipe",
      });
      readStream(frontendProc.stdout, "Vite");
      readStream(frontendProc.stderr, "Vite");
    }
  }

  // --- 4. SHUTDOWN ---

  // Обрабатываем Ctrl+C
  process.on("SIGINT", () => {
    logger.info("Nexus", "Stopping services...");
    engineProc.kill();
    appProc.kill();
    if (frontendProc) frontendProc.kill(); // Убиваем Vite
    process.exit(0);
  });
}

// Хелпер для чтения потока
async function readStream(
  stream: ReadableStream | null,
  name: "Engine" | "App" | "Vite"
) {
  if (!stream) return; // Защита от null

  const reader = stream.getReader();
  const decoder = new TextDecoder();

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;
    // value может прийти куском, содержащим несколько строк или половину строки
    // decoder.decode делает базовую работу, но logger.stream разбивает по \n
    logger.stream(name, decoder.decode(value));
  }
}
