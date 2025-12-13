import { NexusClient } from "./core/client";
import type { NexusConfig } from "./core/client";
import { KVModule } from "./modules/kv";

// Главный класс, объединяющий все модули
export class Nexus {
  // Модули (Public API)
  public readonly kv: KVModule;
  // public readonly queue: QueueModule (в будущем)

  private client: NexusClient;

  constructor(config: NexusConfig) {
    this.client = new NexusClient(config);

    // Инициализируем модули
    this.kv = new KVModule(this.client);
  }
}

// Singleton helper (для удобства DX)
// Позволяет создать глобальный инстанс, читающий ENV
export const createNexus = () => {
  const url = process.env.NEXUS_ENGINE_URL || "http://localhost:4000";
  return new Nexus({ engineUrl: url });
};
