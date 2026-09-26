import { readFile, writeFile, mkdir } from "node:fs/promises";
import openapiTS, { astToString } from "openapi-typescript";

const schema = new URL("../../go-b2b-starter/apicontract/openapi.json", import.meta.url);
const output = new URL("../lib/api/generated/schema.ts", import.meta.url);
const source = JSON.parse(await readFile(schema, "utf8"));
const generated = "// Generated from go-b2b-starter/apicontract/openapi.json; run pnpm api:generate.\n" + astToString(await openapiTS(source));
if (process.argv.length > 3 || (process.argv[2] && process.argv[2] !== "--check")) throw new Error("Usage: node scripts/generate-api.mjs [--check]");
if (process.argv[2] === "--check") {
  if (await readFile(output, "utf8") !== generated) throw new Error("API client schema is stale. Run pnpm api:generate and commit the output.");
  console.log("Generated API contract is current.");
} else {
  await mkdir(new URL(".", output), { recursive: true });
  await writeFile(output, generated);
}
