import pc from "picocolors";

type ServiceName = "Nexus" | "Engine" | "App";

export const logger = {
  info: (service: ServiceName, msg: string) => {
    const prefix = getPrefix(service);
    console.log(`${prefix} ${msg}`);
  },
  error: (service: ServiceName, msg: string) => {
    const prefix = getPrefix(service);
    console.log(`${prefix} ${pc.red(msg)}`);
  },
  // Для перенаправления stdout процессов
  stream: (service: ServiceName, text: string) => {
    const prefix = getPrefix(service);
    // Убираем лишние переносы строк в конце
    const lines = text.trim().split("\n");
    lines.forEach((line) => {
      if (line) console.log(`${prefix} ${line}`);
    });
  },
};

function getPrefix(service: ServiceName) {
  switch (service) {
    case "Nexus":
      return pc.bgBlue(pc.black(" NX "));
    case "Engine":
      return pc.magenta(" [Engine] "); // Фиолетовый как Go
    case "App":
      return pc.cyan(" [App]    "); // Голубой как TS
  }
}
