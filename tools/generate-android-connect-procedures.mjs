import { execFileSync } from "node:child_process";
import { writeFile } from "node:fs/promises";

const repository = new URL("../", import.meta.url);
const protoPath = "proto/realtime/me/status/v1/ingest.proto";
const output = new URL(
  "../apps/mobile/android/app/src/main/kotlin/me/realtime/mobile/status/StatusGatewayProcedures.kt",
  import.meta.url,
);

const descriptorSet = JSON.parse(
  execFileSync(
    "buf",
    [
      "build",
      "--as-file-descriptor-set",
      "--exclude-imports",
      "--exclude-source-info",
      "--path",
      protoPath,
      "--output",
      "-#format=json",
    ],
    { cwd: repository, encoding: "utf8" },
  ),
);
const descriptor = descriptorSet.file?.find(
  (file) => file.name === protoPath.slice("proto/".length),
);
if (!descriptor?.package || !descriptor.service?.length) {
  throw new Error(`No services found in ${protoPath}`);
}

const screamingSnakeCase = (value) => value.replace(/([a-z0-9])([A-Z])/g, "$1_$2").toUpperCase();

const objects = descriptor.service.map((service) => {
  const methods = service.method
    .map(
      (method) =>
        `    const val ${screamingSnakeCase(method.name)} = "/${descriptor.package}.${service.name}/${method.name}"`,
    )
    .join("\n");
  return `internal object ${service.name}Procedures {\n${methods}\n}`;
});

const source = `// Code generated from ${protoPath} by tools/generate-android-connect-procedures.mjs. DO NOT EDIT.

package me.realtime.mobile.status

${objects.join("\n\n")}
`;
await writeFile(output, source);
