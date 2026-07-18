import { readFile, writeFile } from "node:fs/promises";

const descriptorPath = new URL("../app/lib/gen/google/protobuf/descriptor.pb.dart", import.meta.url);
const unescapedMapExample = "///     map<KeyType, ValueType> map_field = 1;";
const escapedMapExample = "///     `map<KeyType, ValueType> map_field = 1;`";
const source = await readFile(descriptorPath, "utf8");

if (!source.includes(unescapedMapExample)) {
  throw new Error("The generated descriptor map example changed; update the normalizer");
}

await writeFile(descriptorPath, source.replace(unescapedMapExample, escapedMapExample));
