import { redirect } from "next/navigation";
import { verifyPayment } from "@/lib/actions/billing/verify-payment";

interface DashboardPageProps {
  searchParams: Promise<{ checkout_id?: string }>;
}

export default async function DashboardPage({ searchParams }: DashboardPageProps) {
  const params = await searchParams;
  const checkoutId = params.checkout_id;

  if (checkoutId) {
    // Call Server Action directly from Server Component
    const result = await verifyPayment(checkoutId);

    if (result.success) {
      console.info("[Dashboard] Payment verified successfully", {
        sessionId: checkoutId,
        hasActiveSubscription: result.data.has_active_subscription,
      });
      redirect("/dashboard/settings?view=subscription&payment_verified=true");
    } else {
      console.error("[Dashboard] Payment verification failed", {
        sessionId: checkoutId,
        error: result.error,
      });
      redirect(`/dashboard/settings?view=subscription&payment_error=true`);
    }
  }

  redirect("/dashboard/settings");
}
