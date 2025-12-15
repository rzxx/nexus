import { NexusClient } from "../../core/client";
import type { JsonValue } from "../../types";

export interface TicketOptions {
  userId: string;
  channels: string[];
}

export interface PublishOptions<T> {
  channel: string;
  data: T;
}

export class WSModule {
  constructor(private readonly client: NexusClient) {}

  /**
   * Создать одноразовый тикет для подключения по WebSocket
   */
  async createTicket(opts: TicketOptions): Promise<string> {
    const res = await this.client.request<{ ticket: string }>(
      "POST",
      "/pubsub/ticket",
      {
        user_id: opts.userId,
        channels: opts.channels,
      }
    );
    return res.ticket;
  }

  /**
   * Опубликовать сообщение в канал
   */
  async publish<T extends JsonValue>(channel: string, data: T): Promise<void> {
    await this.client.request("POST", "/pubsub/publish", {
      channel,
      data,
    });
  }
}
