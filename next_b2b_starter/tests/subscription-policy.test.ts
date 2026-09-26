import assert from "node:assert/strict";
import { test } from "node:test";
import type { SubscriptionGateState } from "../lib/polar/current-subscription.ts";
import { checkoutUnavailableReason, subscriptionSummary } from "../lib/polar/subscription-policy.ts";

const verifiedEmpty: SubscriptionGateState = {
  isAuthenticated: true, isActive: false, canStartCheckout: true,
  productId: null, planId: null, subscription: null, backendAvailable: true,
};
const active: SubscriptionGateState = {
  ...verifiedEmpty, isActive: true, canStartCheckout: false, productId: "configured-plan",
  subscription: { id: "subscription", status: "active", productId: "configured-plan", productName: null,
    currentPeriodEnd: "2099-01-01T00:00:00Z", cancelAtPeriodEnd: false },
};

test("checkout requires a positively verified empty subscription state", () => {
  assert.equal(checkoutUnavailableReason(verifiedEmpty), null);
  for (const state of [
    null,
    { ...verifiedEmpty, canStartCheckout: false },
    { ...verifiedEmpty, isAuthenticated: false },
    { ...verifiedEmpty, reason: "INSUFFICIENT_PERMISSIONS" },
    { ...verifiedEmpty, reason: "BILLING_DISABLED" },
    { ...verifiedEmpty, backendAvailable: false },
    { ...verifiedEmpty, subscription: active.subscription },
    { ...active, canStartCheckout: true },
  ]) {
    assert.ok(checkoutUnavailableReason(state), `checkout must be blocked for ${JSON.stringify(state)}`);
  }
});

test("scheduled cancellation still blocks a replacement checkout", () => {
  const scheduled = { ...active, subscription: { ...active.subscription!, cancelAtPeriodEnd: true } };
  assert.match(checkoutUnavailableReason(scheduled)!, /existing subscription/);
  assert.equal(subscriptionSummary(scheduled).title, "Cancellation scheduled");
  assert.match(subscriptionSummary(scheduled).description, /remains active until/);
  assert.equal(subscriptionSummary(active).title, "Active subscription");
});

test("unknown, disabled and revoked billing states are not presented as an empty paid plan", () => {
  assert.equal(subscriptionSummary(null).title, "Billing status unavailable");
  assert.equal(subscriptionSummary({ ...verifiedEmpty, backendAvailable: false }).title, "Billing status unavailable");
  assert.equal(subscriptionSummary({ ...verifiedEmpty, canStartCheckout: false }).title, "Billing status unavailable");
  assert.equal(subscriptionSummary({ ...verifiedEmpty, reason: "BILLING_DISABLED" }).title, "Billing is disabled");
  assert.equal(subscriptionSummary({ ...verifiedEmpty, reason: "INSUFFICIENT_PERMISSIONS" }).title, "Billing access unavailable");
  assert.equal(subscriptionSummary(verifiedEmpty).title, "No active subscription");
  assert.match(subscriptionSummary(verifiedEmpty).description, /workspace remains available/);
});
