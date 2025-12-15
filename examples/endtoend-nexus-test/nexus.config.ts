export default {
  app: {
    port: 3000,
    command: "bun run src/api/index.ts",
  },
  engine: {
    port: 4000,
    path: "../../engine", // Путь к Go Engine
  },
  frontend: {
    port: 5173,
    command: "bunx --bun vite", // Запускаем Vite
  },
};
