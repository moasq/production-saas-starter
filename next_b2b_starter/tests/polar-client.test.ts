import { test } from "node:test";
import assert from "node:assert/strict";
import { Polar } from "@polar-sh/sdk";
import { createPolarHttpClient } from "../lib/polar/http-client.ts";

test("Polar product and checkout requests pin the same API version as Go", async () => {
  const requests: Request[] = [];
  const client = new Polar({
    accessToken: "test-token",
    server: "sandbox",
    httpClient: createPolarHttpClient({
      fetcher: async (request) => {
        requests.push(request as Request);
        return Response.json({ detail: "Not found" }, { status: 404 });
      },
    }),
  });

  const productId = "00000000-0000-4000-8000-000000000001";
  await assert.rejects(client.products.get({ id: productId }));
  await assert.rejects(client.checkouts.create({ products: [productId], externalCustomerId: "org-a" }));

  assert.equal(requests.length, 2);
  assert.deepEqual(requests.map((request) => request.method), ["GET", "POST"]);
  for (const request of requests) {
    assert.equal(new URL(request.url).origin, "https://sandbox-api.polar.sh");
    assert.equal(request.headers.get("Polar-Version"), "2026-04");
    assert.equal(request.headers.get("Authorization"), "Bearer test-token");
  }
  const checkout = await requests[1].json();
  assert.equal(checkout.external_customer_id, "org-a");
  assert.deepEqual(checkout.products, [productId]);
});
