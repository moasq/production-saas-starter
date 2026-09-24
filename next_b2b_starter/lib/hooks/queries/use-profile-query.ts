/**
 * Profile Query Hook
 *
 * Fetches and caches the current user's profile.
 * Data is cached globally and reused across components.
 */

import { useQuery, type UseQueryOptions } from "@tanstack/react-query";
import { memberRepository } from "@/lib/api/api/repositories/member-repository";
import { queryKeys } from "./query-keys";
import type { UserProfile } from "@/lib/models/member.model";

export function useProfileQuery(
  options?: Omit<
    UseQueryOptions<UserProfile, Error>,
    "queryKey" | "queryFn"
  >
) {
  return useQuery({
    queryKey: queryKeys.profile.detail(),
    queryFn: () => memberRepository.getProfile(),

    // Keep profile fresh for 10 minutes (longer than default)
    staleTime: 10 * 60 * 1000,

    // Cache for 30 minutes (profile doesn't change often)
    gcTime: 30 * 60 * 1000,

    // Retry failed requests once
    retry: 1,

    // Don't refetch on window focus (profile rarely changes)
    refetchOnWindowFocus: false,

    ...options,
  });
}
