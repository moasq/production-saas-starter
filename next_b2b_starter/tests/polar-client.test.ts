import { test } from "node:test";
import assert from "node:assert/strict";
import { createBillingProvider } from "../lib/polar/provider.ts";

const productId = "00000000-0000-4000-8000-000000000001";
const product = {
  id: productId, name: "Workspace", description: null, is_archived: false, is_recurring: true,
  recurring_interval: "month", recurring_interval_count: 1,
  prices: [{ amount_type: "fixed", price_amount: 2500, price_currency: "usd", is_archived: false, product_id: productId }],
};
const checkout = { productId, organizationId: "org-a", email: "admin@example.test", name: "Synthetic Admin",
  successUrl: "https://starter.example.test/dashboard?checkout_id={CHECKOUT_ID}",
  returnUrl: "https://starter.example.test/dashboard/settings?view=subscription" };

for (const server of ["sandbox", "production"] as const) {
  test(`Polar ${server} success contracts keep the API version and organization identity`, async (t) => {
    const requests: Request[] = [];
    t.mock.method(globalThis, "fetch", async (input: string | URL | Request, init?: RequestInit) => {
      const request = new Request(input, init);
      requests.push(request);
      switch (new URL(request.url).pathname) {
        case `/v1/products/${productId}`: return Response.json(product);
        case "/v1/checkouts/": return Response.json({ url: "https://sandbox.polar.sh/checkout/synthetic" }, { status: 201 });
        case "/v1/customer-sessions/": return Response.json({ customer_portal_url: "https://sandbox.polar.sh/portal/synthetic" }, { status: 201 });
        default: throw new Error("Unexpected fixture request");
      }
    });
    const client = createBillingProvider({ accessToken: "test-only-token", server });
    assert.deepEqual(await client.getPlan(productId), { id: productId, productId, name: "Workspace", description: null,
      price: 25, currency: "usd", interval: "month" });
    assert.deepEqual(await client.createCheckout(checkout), { url: "https://sandbox.polar.sh/checkout/synthetic" });
    assert.deepEqual(await client.createPortal("org-a"), { url: "https://sandbox.polar.sh/portal/synthetic" });
    assert.deepEqual(requests.map((request) => request.method), ["GET", "POST", "POST"]);
    for (const request of requests) {
      assert.equal(new URL(request.url).origin, server === "sandbox" ? "https://sandbox-api.polar.sh" : "https://api.polar.sh");
      assert.equal(request.headers.get("Polar-Version"), "2026-04");
      assert.equal(request.headers.get("Authorization"), "Bearer test-only-token");
      // Application tenant identity is not the merchant's Polar-Organization header.
      assert.equal(request.headers.get("Polar-Organization"), null);
    }
    assert.deepEqual(await requests[1].json(), { products: [productId], external_customer_id: "org-a",
      customer_email: checkout.email, customer_name: checkout.name, success_url: checkout.successUrl, return_url: checkout.returnUrl });
    assert.deepEqual(await requests[2].json(), { external_customer_id: "org-a" });
  });
}

test("unsupported product pricing cannot be offered as the starter's fixed recurring plan", async (t) => {
  const bodies = [
    { ...product, is_archived: true }, { ...product, is_recurring: false },
    { ...product, recurring_interval_count: 3 }, { ...product, prices: [] },
    { ...product, prices: [product.prices[0], product.prices[0]] },
    { ...product, prices: [{ ...product.prices[0], amount_type: "metered_unit" }] },
    { ...product, prices: [{ ...product.prices[0], is_archived: true }] },
  ];
  const client = createBillingProvider({ accessToken: "test-only-token", server: "sandbox" });
  for (const body of bodies) {
    const fetch = t.mock.method(globalThis, "fetch", async () => Response.json(body));
    assert.equal(await client.getPlan(productId), null);
    fetch.mock.restore();
  }
});

test("malformed successful product responses fail closed instead of displaying an invented price", async (t) => {
  const bodies = [null, {}, { ...product, id: "other-product" }, { ...product, name: "" },
    { ...product, is_archived: undefined }, { ...product, prices: null },
    { ...product, prices: [{ ...product.prices[0], price_amount: "2500" }] },
    { ...product, prices: [{ ...product.prices[0], price_amount: -1 }] },
    { ...product, prices: [{ ...product.prices[0], price_currency: "" }] },
    { ...product, prices: [{ ...product.prices[0], product_id: "other-product" }] },
    { ...product, recurring_interval: null }];
  const client = createBillingProvider({ accessToken: "test-only-token", server: "sandbox" });
  for (const body of bodies) {
    const fetch = t.mock.method(globalThis, "fetch", async () => Response.json(body));
    await assert.rejects(client.getPlan(productId), /Invalid billing provider response/);
    fetch.mock.restore();
  }
});

test("checkout and portal success require a usable HTTPS URL", async (t) => {
  const client = createBillingProvider({ accessToken: "test-only-token", server: "sandbox" });
  for (const url of [undefined, null, "", "javascript:alert(1)", "http://polar.example.test", "https://user:password@example.test"]) {
    const fetch = t.mock.method(globalThis, "fetch", async () => Response.json({ url, customer_portal_url: url }));
    await assert.rejects(client.createCheckout(checkout));
    await assert.rejects(client.createPortal("org-a"));
    fetch.mock.restore();
  }
});

test("provider errors and ambiguous network failures never retry reads or mutations automatically", async (t) => {
  const client = createBillingProvider({ accessToken: "test-only-token", server: "sandbox" });
  for (const status of [404, 429, 500, "network"] as const) {
    const fetch = t.mock.method(globalThis, "fetch", async () => {
      if (status === "network") throw new TypeError("Synthetic connection reset after request");
      return Response.json({ error: "SyntheticProviderError" }, { status, headers: { "Retry-After": "0" } });
    });
    await assert.rejects(client.getPlan(productId));
    await assert.rejects(client.createCheckout(checkout));
    await assert.rejects(client.createPortal("org-a"));
    assert.equal(fetch.mock.callCount(), 3, `${status}: exactly one request per operation`);
    fetch.mock.restore();
  }
});

test("provider calls propagate a ten-second abort signal and do not retry after timeout", async (t) => {
  const controller = new AbortController();
  const timeout = t.mock.method(AbortSignal, "timeout", (milliseconds: number) => {
    assert.equal(milliseconds, 10_000);
    return controller.signal;
  });
  const fetch = t.mock.method(globalThis, "fetch", async (_input: unknown, init?: RequestInit) => {
    assert.equal(init?.signal, controller.signal);
    return new Promise<Response>((_resolve, reject) => {
      init!.signal!.addEventListener("abort", () => reject(init!.signal!.reason), { once: true });
    });
  });
  const client = createBillingProvider({ accessToken: "test-only-token", server: "sandbox" });
  const pending = client.createCheckout(checkout);
  controller.abort(new DOMException("Synthetic request timeout", "TimeoutError"));
  await assert.rejects(pending, /Synthetic request timeout/);
  assert.equal(timeout.mock.callCount(), 1);
  assert.equal(fetch.mock.callCount(), 1);
});
