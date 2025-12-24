import { NextResponse } from "next/server";
import { Webhooks } from "@polar-sh/nextjs";

const webhookSecret = process.env.POLAR_WEBHOOK_SECRET;

async function handleSubscriptionEvent(_eventType: string, _payload: unknown) {
  // TODO: Forward webhook events to Go backend for persistence.
  // Options:
  // 1. Call backend API: POST /api/webhooks/polar with { eventType, payload }
  // 2. Configure Polar.sh to send webhooks directly to Go backend
  //
  // Backend already has ProcessWebhookEvent service ready in:
  // src/app/billing/app/services/process_webhook_event_service.go
}

export const POST = webhookSecret
  ? Webhooks({
      webhookSecret,
      onSubscriptionCreated: async (subscription) => {
        await handleSubscriptionEvent("subscription.created", subscription);
      },
      onSubscriptionUpdated: async (subscription) => {
        await handleSubscriptionEvent("subscription.updated", subscription);
      },
      onSubscriptionCanceled: async (subscription) => {
        await handleSubscriptionEvent("subscription.canceled", subscription);
      },
      onOrderPaid: async (order) => {
        await handleSubscriptionEvent("order.paid", order);
      },
    })
  : async () =>
      NextResponse.json(
        { error: "Polar webhook secret not configured." },
        { status: 503 }
      );
