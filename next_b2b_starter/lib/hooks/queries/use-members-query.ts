/**
 * Members Query Hook
 *
 * Fetches and caches organization members list.
 * Only fetches when organizationId is provided (lazy loading).
 */

import { useQuery, type UseQueryOptions } from "@tanstack/react-query";
import { memberRepository } from "@/lib/api/api/repositories/member-repository";
import { queryKeys } from "./query-keys";
import type { MemberListResponse } from "@/lib/models/member.model";

interface UseMembersQueryOptions {
  organizationId?: string;
  page?: number;
  pageSize?: number;
  enabled?: boolean;
}

export function useMembersQuery(
  options: UseMembersQueryOptions = {},
  queryOptions?: Omit<
    UseQueryOptions<MemberListResponse, Error>,
    "queryKey" | "queryFn" | "enabled"
  >
) {
  const {
    organizationId,
    page = 1,
    pageSize = 50,
    enabled = true,
  } = options;

  return useQuery({
    queryKey: queryKeys.members.list({ organizationId, page, pageSize }),
    queryFn: () =>
      memberRepository.getMembers(),

    // Only fetch if organizationId is provided and enabled is true
    enabled: Boolean(organizationId) && enabled,

    // Members data is fresh for 5 minutes
    staleTime: 5 * 60 * 1000,

    // Cache for 10 minutes
    gcTime: 10 * 60 * 1000,

    // Retry once on failure
    retry: 1,

    // Don't refetch on window focus
    refetchOnWindowFocus: false,

    ...queryOptions,
  });
}
