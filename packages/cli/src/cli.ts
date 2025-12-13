#!/usr/bin/env bun
import cac from "cac";
import { loadConfig } from "./config";
import { startDev } from "./commands/dev";
import { logger } from "./logger";

const cli = cac("nexus");

cli.command("dev", "Start the development environment").action(async () => {
  try {
    const config = await loadConfig();
    await startDev(config);
  } catch (e: any) {
    logger.error("Nexus", e.message);
    process.exit(1);
  }
});

cli.help();
cli.parse();
