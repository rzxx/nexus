import { NexusClient } from "./core/client";
import type { NexusConfig } from "./core/client";
import { KVModule } from "./modules/kv";

export class Nexus {
  public readonly kv: KVModule;
  private client: NexusClient;

  constructor(config: NexusConfig) {
    this.client = new NexusClient(config);
    this.kv = new KVModule(this.client);
  }
}

export const createNexus = (config?: Partial<NexusConfig>) => {
  const url =
    config?.engineUrl ||
    process.env.NEXUS_ENGINE_URL ||
    "http://localhost:4000";
  return new Nexus({ engineUrl: url });
};

// Экспортируем типы, чтобы пользователь мог их использовать
export * from "./types";
