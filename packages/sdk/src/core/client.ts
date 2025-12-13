export interface NexusConfig {
  engineUrl: string; // Например, http://localhost:4000
}

export class NexusClient {
  private baseUrl: string;

  constructor(config: NexusConfig) {
    this.baseUrl = config.engineUrl.replace(/\/$/, ""); // Убираем слеш в конце
  }

  /**
   * Универсальный метод для общения с Go-ядром
   */
  async request<T>(
    method: "GET" | "POST",
    path: string,
    body?: any
  ): Promise<T | null> {
    const headers = { "Content-Type": "application/json" };

    const response = await fetch(`${this.baseUrl}${path}`, {
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined,
    });

    if (!response.ok) {
      // Пытаемся прочитать ошибку от Go
      const text = await response.text();
      throw new Error(`Nexus Engine Error [${response.status}]: ${text}`);
    }

    // Если тело пустое (например, при status 204), возвращаем null
    const text = await response.text();
    return text ? (JSON.parse(text) as T) : null;
  }
}
