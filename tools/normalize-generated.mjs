import { readdir, readFile, writeFile } from "node:fs/promises";

const descriptorPaths = [
  new URL(
    "../packages/manager-contracts-dart/lib/gen/google/protobuf/descriptor.pb.dart",
    import.meta.url,
  ),
];
const unescapedMapExample = "///     map<KeyType, ValueType> map_field = 1;";
const escapedMapExample = "///     `map<KeyType, ValueType> map_field = 1;`";

for (const descriptorPath of descriptorPaths) {
  const source = await readFile(descriptorPath, "utf8");
  if (!source.includes(unescapedMapExample)) {
    throw new Error(`The generated descriptor map example changed: ${descriptorPath.pathname}`);
  }

  await writeFile(descriptorPath, source.replace(unescapedMapExample, escapedMapExample));
}

const typescriptRoots = [
  new URL("../packages/auth-contracts-web/src/gen/", import.meta.url),
  new URL("../packages/library-contracts-web/src/gen/", import.meta.url),
  new URL("../packages/manager-contracts-web/src/gen/", import.meta.url),
  new URL("../packages/status-contracts-web/src/gen/", import.meta.url),
  new URL("../services/manager/src/gen/", import.meta.url),
];

async function* files(directory) {
  for (const entry of await readdir(directory, { withFileTypes: true })) {
    const child = new URL(entry.name, directory);
    if (entry.isDirectory()) {
      yield* files(new URL(`${entry.name}/`, directory));
    } else {
      yield child;
    }
  }
}

for (const root of typescriptRoots) {
  for await (const file of files(root)) {
    if (!file.pathname.endsWith(".ts")) {
      continue;
    }

    const source = await readFile(file, "utf8");
    await writeFile(file, `${source.trimEnd()}\n`);
  }
}
