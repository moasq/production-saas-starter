import { useQuery, type UseQueryOptions } from "@tanstack/react-query";

import { getProducts } from "@/lib/actions/billing/get-products";
import type { PolarPlan } from "@/lib/polar/plans";
import { queryKeys } from "./query-keys";

/**
 * Fetch products from Polar
 *
 * Returns all active products with pricing and metadata.
 * Uses Server Action instead of API route.
 */
export function useProductsQuery(
  options?: Omit<
    UseQueryOptions<PolarPlan[], Error>,
    "queryKey" | "queryFn"
  >
) {
  return useQuery({
    queryKey: queryKeys.products.list,
    queryFn: async (): Promise<PolarPlan[]> => {
      const result = await getProducts();

      if (!result.success) {
        throw new Error(result.error ?? "Failed to fetch products");
      }

      return result.data;
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes
    refetchOnWindowFocus: false,
    refetchOnMount: false,
    ...options,
  });
}
