/**
 * Subscription Query Hook
 *
 * Fetches and caches the current subscription status from Polar.
 * Uses Server Action instead of API route.
 */

import { useQuery, type UseQueryOptions } from "@tanstack/react-query";
import { queryKeys } from "./query-keys";
import { getSubscriptionStatus } from "@/lib/actions/billing/get-subscription-status";
import type { SubscriptionGateState } from "@/lib/polar/current-subscription";

async function fetchSubscriptionStatus(): Promise<SubscriptionGateState> {
  const result = await getSubscriptionStatus();

  if (!result.success) {
    throw new Error(result.error ?? "Unable to load subscription status");
  }

  return result.data;
}

export function useSubscriptionQuery(
  options?: Omit<
    UseQueryOptions<SubscriptionGateState, Error>,
    "queryKey" | "queryFn"
  >
) {
  return useQuery({
    queryKey: queryKeys.subscription.status(),
    queryFn: fetchSubscriptionStatus,

    // Subscription status is fresh for 5 minutes
    staleTime: 5 * 60 * 1000,

    // Cache for 15 minutes
    gcTime: 15 * 60 * 1000,

    // Retry once on failure
    retry: 1,

    // Don't refetch on window focus
    refetchOnWindowFocus: false,

    ...options,
  });
}

/**
 * Hook to get subscription state with safe defaults
 */
export function useSubscription() {
  const { data } = useSubscriptionQuery();
  return data ?? null;
}

/**
 * Hook to check if subscription is active
 */
export function useIsSubscriptionActive() {
  const subscription = useSubscription();
  return subscription?.isActive ?? false;
}
