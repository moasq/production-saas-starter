import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import Ajv from "ajv/dist/2020.js";
import addFormats from "ajv-formats";

const spec = JSON.parse(readFileSync(new URL("../../go-b2b-starter/apicontract/openapi.json", import.meta.url), "utf8"));
const ajv = new Ajv({ strict: false, allErrors: true });
addFormats(ajv);
ajv.addSchema({ ...spec, $id: "starter-api" });
const checked = new Set();
const exercised = new Set();
const validators = new Map();
const pointer = (value) => value.replaceAll("~", "~0").replaceAll("/", "~1");
function validate(pointerPath, body) {
  let validator = validators.get(pointerPath);
  if (!validator) { validator = ajv.compile({ $ref: `starter-api#${pointerPath}` }); validators.set(pointerPath, validator); }
  assert.ok(validator(body), `API contract mismatch at ${pointerPath}: ${ajv.errorsText(validator.errors)}`);
}

// Checks actual responses from the mounted local stack; never prints response data.
export function assertApiResponse(url, method, status, data, requestBody) {
  const pathname = new URL(url, "http://localhost").pathname.replace(/^\/api/, "");
  const entry = Object.entries(spec.paths).find(([path]) => {
    const pattern = path.split("/").map(part => part.startsWith("{") ? "[^/]+" : part.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")).join("/");
    return new RegExp(`^${pattern}$`).test(pathname);
  });
  if (!entry) return; // Identity/workspace routes are explicitly owned by Next.js.
  const [path, operations] = entry;
  const operation = operations[method.toLowerCase()];
  assert.ok(operation, `Undocumented method ${method} ${path}`);
  const responseKey = String(status) in operation.responses ? String(status) : "default";
  const response = operation.responses[responseKey];
  assert.ok(response, `Undocumented response ${status} for ${method} ${path}`);
  exercised.add(`${method} ${path}`);
  const base = `/paths/${pointer(path)}/${method.toLowerCase()}`;
  if (response["x-empty-body"] && data === "") return;
  if (response.content?.["application/json"]) validate(`${base}/responses/${responseKey}/content/application~1json/schema`, data);
  else assert.ok(data === "" || data === null || data === undefined, "No-content response carried a body");
  if (status >= 200 && status < 300) {
    checked.add(`${method} ${path}`);
    if (operation.requestBody) validate(`${base}/requestBody/content/application~1json/schema`, requestBody);
  }
}

export function assertApiCoverage() {
  const expected = Object.entries(spec.paths).flatMap(([path, operations]) => Object.keys(operations).map(method => `${method.toUpperCase()} ${path}`));
  // Disabled billing deliberately rejects verification; its error and permission
  // paths are exercised without a live payment or fabricated success response.
  assert.deepEqual(expected.filter(key => !exercised.has(key)), [], "Every mounted operation must be exercised");
  const optional = "POST /subscriptions/verify-payment";
  assert.deepEqual(expected.filter(key => key !== optional && !checked.has(key)), [], "Every supported core operation needs a successful mounted response");
  console.log(`API contract checked against ${checked.size} successful mounted operations; live payment verification remains separate.`);
}
