import { NexusClient } from "../../core/client";
import type { JsonValue } from "../../types";

export interface SetOptions {
  ttl?: number;
}

export class KVModule {
  constructor(private readonly client: NexusClient) {}

  /**
   * get возвращает типизированный ответ или null.
   * Мы используем Generic <T>, чтобы пользователь указал структуру данных.
   */
  async get<T extends JsonValue>(key: string): Promise<T | null> {
    try {
      // Здесь мы явно говорим клиенту, что ожидаем T
      return await this.client.request<T>("GET", `/kv/get?key=${key}`);
    } catch (e: any) {
      // Обработка 404
      if (e.message && e.message.includes("404")) {
        return null;
      }
      throw e;
    }
  }

  /**
   * set принимает только валидный JsonValue.
   * Если пользователь попытается сунуть функцию () => {}, TS выдаст ошибку компиляции.
   */
  async set(
    key: string,
    value: JsonValue,
    options?: SetOptions
  ): Promise<void> {
    await this.client.request("POST", "/kv/set", {
      key,
      value,
      ttl: options?.ttl || 0,
    });
  }
}
