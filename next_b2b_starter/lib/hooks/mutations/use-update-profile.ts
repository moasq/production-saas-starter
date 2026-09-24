/**
 * Update Profile Mutation
 *
 * Handles profile updates with optimistic UI updates and automatic cache invalidation.
 */

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { memberRepository } from "@/lib/api/api/repositories/member-repository";
import { queryKeys } from "../queries/query-keys";
import type { UpdateProfileRequest, UserProfile } from "@/lib/models/member.model";

export function useUpdateProfile() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: UpdateProfileRequest) =>
      memberRepository.updateProfile(request),

    // Optimistic update: Update UI immediately before API call completes
    onMutate: async (newProfile) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({
        queryKey: queryKeys.profile.detail(),
      });

      // Snapshot the previous value
      const previousProfile = queryClient.getQueryData<UserProfile>(
        queryKeys.profile.detail()
      );

      // Optimistically update the cache
      if (previousProfile) {
        queryClient.setQueryData<UserProfile>(
          queryKeys.profile.detail(),
          (old) =>
            old
              ? {
                  ...old,
                  name: newProfile.name ?? old.name,
                }
              : old
        );
      }

      // Return context with previous value for rollback
      return { previousProfile };
    },

    // On success, update cache with fresh data from server
    onSuccess: (data) => {
      queryClient.setQueryData(queryKeys.profile.detail(), data);
    },

    // On error, rollback to previous value
    onError: (error, _variables, context) => {
      if (context?.previousProfile) {
        queryClient.setQueryData(
          queryKeys.profile.detail(),
          context.previousProfile
        );
      }
      console.error("[Mutation] Profile update failed:", error);
    },

    // Always refetch after error or success to ensure consistency
    onSettled: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.profile.detail(),
      });
    },
  });
}
