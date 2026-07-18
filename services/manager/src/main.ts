import { assertSafeProcess, loadConfig } from "./infrastructure/config.js";
import { createServer } from "./server.js";

const config = loadConfig();
assertSafeProcess(config);
const application = await createServer(config);

await application.server.listen({ host: config.host, port: config.port });

let stopping = false;
async function stop(signal: string): Promise<void> {
  if (stopping) {
    return;
  }
  stopping = true;
  application.server.log.info({ signal }, "shutting down");
  try {
    await application.close();
    process.exitCode = 0;
  } catch (error) {
    application.server.log.error({ err: error }, "shutdown failed");
    process.exitCode = 1;
  }
}

process.once("SIGINT", () => void stop("SIGINT"));
process.once("SIGTERM", () => void stop("SIGTERM"));
