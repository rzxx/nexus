import type { HttpMethod, JsonValue } from "../types";

export interface NexusConfig {
  engineUrl: string;
}

export class NexusClient {
  private baseUrl: string;

  constructor(config: NexusConfig) {
    // Убираем trailing slash, если есть
    this.baseUrl = config.engineUrl.replace(/\/$/, "");
  }

  /**
   * request делает запрос к Engine.
   * T = ожидаемый тип ответа.
   * body = строго типизирован как JsonValue (никаких функций или circular objects).
   */
  async request<T>(
    method: HttpMethod,
    path: string,
    body?: JsonValue
  ): Promise<T> {
    const headers: HeadersInit = {
      "Content-Type": "application/json",
    };

    const response = await fetch(`${this.baseUrl}${path}`, {
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined,
    });

    if (!response.ok) {
      const text = await response.text();
      // Выбрасываем типизированную ошибку, которую можно ловить
      throw new Error(`[Nexus Engine] ${response.status}: ${text}`);
    }

    const text = await response.text();
    if (!text) {
      // Если тело пустое, но мы ожидали T, приходится возвращать null as T
      // Это единственное "узкое место", где мы доверяем Engine.
      return null as unknown as T;
    }

    try {
      return JSON.parse(text) as T;
    } catch (e) {
      throw new Error(`[Nexus SDK] Failed to parse response: ${text}`);
    }
  }
}
