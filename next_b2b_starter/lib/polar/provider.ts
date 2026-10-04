import { createPolar } from "@polar-sh/sdk/2026-04";
import type { PolarConfig } from "./environment.ts";
import type { PolarPlan } from "./plans.ts";

type EnabledConfig = Extract<PolarConfig, { enabled: true }>;
type CheckoutInput = {
  productId: string;
  organizationId: string;
  email: string;
  name: string;
  successUrl: string;
  returnUrl: string;
};

function hostedUrl(value: unknown): string {
  if (typeof value !== "string") throw new Error("Invalid billing provider response");
  const url = new URL(value);
  if (url.protocol !== "https:" || url.username || url.password) {
    throw new Error("Invalid billing provider response");
  }
  return value;
}

// Server callers obtain this adapter through client.ts. SDK DTOs stay at this boundary.
export function createBillingProvider(config: Pick<EnabledConfig, "accessToken" | "server">) {
  // The versioned export supplies Polar-Version: 2026-04, matching Go's reads.
  // SDK 1.x performs one fetch per call; never wrap mutations in automatic retries.
  const client = createPolar({ accessToken: config.accessToken, environment: config.server, timeout: 10 });
  return {
    async getPlan(productId: string): Promise<PolarPlan | null> {
      const product = await client.products.get(productId);
      // SDK 1.x returns JSON without runtime schema validation. Check consumed fields.
      if (!product || product.id !== productId || typeof product.name !== "string" || !product.name.trim() ||
          (product.description !== null && typeof product.description !== "string") ||
          typeof product.is_archived !== "boolean" || typeof product.is_recurring !== "boolean" ||
          !Array.isArray(product.prices)) {
        throw new Error("Invalid billing provider response");
      }
      // The starter displays one fixed price per billing interval, with no metered or multi-interval pricing.
      if (product.is_archived || !product.is_recurring || product.prices.length !== 1 || product.recurring_interval_count !== 1) return null;
      const price = product.prices[0];
      if (!price || price.amount_type !== "fixed" || price.is_archived) return null;
      if (typeof price.is_archived !== "boolean" || price.product_id !== productId ||
          !Number.isSafeInteger(price.price_amount) || price.price_amount < 0 ||
          typeof price.price_currency !== "string" || !/^[a-z]{3}$/i.test(price.price_currency) ||
          !["day", "week", "month", "year"].includes(product.recurring_interval ?? "")) {
        throw new Error("Invalid billing provider response");
      }
      return { id: product.id, productId: product.id, name: product.name, description: product.description,
        price: price.price_amount / 100, currency: price.price_currency, interval: product.recurring_interval! };
    },
    async createCheckout(input: CheckoutInput): Promise<{ url: string }> {
      const checkout = await client.checkouts.create({ products: [input.productId],
        external_customer_id: input.organizationId, customer_email: input.email, customer_name: input.name,
        success_url: input.successUrl, return_url: input.returnUrl });
      return { url: hostedUrl(checkout?.url) };
    },
    async createPortal(organizationId: string): Promise<{ url: string }> {
      const portal = await client.customerSessions.create({ external_customer_id: organizationId });
      return { url: hostedUrl(portal?.customer_portal_url) };
    },
  };
}
