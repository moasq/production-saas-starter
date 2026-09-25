import { test } from "node:test";
import assert from "node:assert/strict";
import { sanitizeReturnTo } from "../lib/auth/urls.ts";
test("authentication redirects stay on the application origin", () => {
 for (const path of ["https://evil.test", "//evil.test", "/\\evil.test", "/\n/evil.test"]) assert.equal(sanitizeReturnTo(path), undefined);
 assert.equal(sanitizeReturnTo("/dashboard/settings?view=profile"), "/dashboard/settings?view=profile");
});
