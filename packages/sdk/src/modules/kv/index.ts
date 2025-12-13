import { NexusClient } from "../../core/client";

// Типы данных (совпадают с Go JSON)
export interface SetOptions {
  ttl?: number;
}

export class KVModule {
  private client: NexusClient;

  constructor(client: NexusClient) {
    this.client = client;
  }

  /**
   * Получить значение по ключу
   */
  async get<T>(key: string): Promise<T | null> {
    try {
      // Вызываем ручку Go Engine
      return await this.client.request<T>("GET", `/kv/get?key=${key}`);
    } catch (e: any) {
      // Если 404, возвращаем null (это нормальное поведение для KV)
      if (e.message && e.message.includes("404")) {
        return null;
      }
      throw e;
    }
  }

  /**
   * Сохранить значение
   */
  async set(key: string, value: any, options?: SetOptions): Promise<void> {
    await this.client.request("POST", "/kv/set", {
      key,
      value,
      ttl: options?.ttl || 0,
    });
  }
}
