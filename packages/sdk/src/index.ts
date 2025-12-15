import { NexusClient } from "./core/client";
import type { NexusConfig } from "./core/client";
import { KVModule } from "./modules/kv";
import { WSModule } from "./modules/ws";

export class Nexus {
  public readonly kv: KVModule;
  public readonly ws: WSModule;

  private client: NexusClient;

  constructor(config: NexusConfig) {
    this.client = new NexusClient(config);
    this.kv = new KVModule(this.client);
    this.ws = new WSModule(this.client);
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
