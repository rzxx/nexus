import { join } from "path";

export interface NexusConfig {
  app: {
    port: number;
    command: string; // "bun run src/index.ts"
  };
  engine: {
    port: number;
    path: string; // Путь к папке engine (../../engine)
  };
  frontend?: {
    port: number;
    command: string;
  };
}

// Дефолтный конфиг
export const defaultConfig: NexusConfig = {
  app: {
    port: 3000,
    command: "bun run src/index.ts",
  },
  engine: {
    port: 4000,
    path: "./nexus/engine", // По умолчанию
  },
};

export async function loadConfig(): Promise<NexusConfig> {
  const configPath = join(process.cwd(), "nexus.config.ts");
  try {
    // Динамический импорт конфига пользователя
    const userConfig = await import(configPath);
    return { ...defaultConfig, ...userConfig.default };
  } catch (e) {
    console.warn("⚠️  nexus.config.ts not found, using defaults.");
    return defaultConfig;
  }
}
