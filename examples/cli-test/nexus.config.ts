export default {
  app: {
    port: 3000,
    command: "bun run index.ts",
  },
  engine: {
    port: 4000,
    // Путь относительно места запуска CLI (корня my-app)
    // Предполагаем монорепозиторий
    path: "../../engine",
  },
};
