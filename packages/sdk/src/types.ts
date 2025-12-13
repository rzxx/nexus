// Строгая типизация JSON.
// Мы запрещаем any, Function, undefined, Symbol и прочее.

export type JsonPrimitive = string | number | boolean | null;

export type JsonObject = { [key: string]: JsonValue };

export type JsonArray = JsonValue[];

export type JsonValue = JsonPrimitive | JsonObject | JsonArray;

// Вспомогательный тип для HTTP методов
export type HttpMethod = "GET" | "POST" | "PUT" | "DELETE";
