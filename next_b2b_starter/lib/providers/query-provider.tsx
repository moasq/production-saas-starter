"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useState, type ReactNode } from "react";

/**
 * Query Client Provider
 *
 * Configures TanStack Query with sensible defaults for the starter:
 * - 5 minute stale time (data stays fresh for 5 minutes)
 * - 10 minute garbage collection (cache persists for 10 minutes)
 * - Single retry on failure
 * - No aggressive refetching on window focus
 */

function makeQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        // Data is considered fresh for 5 minutes
        staleTime: 5 * 60 * 1000,

        // Cache data for 10 minutes after last use
        gcTime: 10 * 60 * 1000,

        // Retry failed requests once
        retry: 1,

        // Don't refetch on window focus (prevents aggressive refetching)
        refetchOnWindowFocus: false,

        // Don't refetch on mount - respect staleTime instead
        refetchOnMount: false,

        // Don't refetch on reconnect by default
        refetchOnReconnect: false,
      },
      mutations: {
        // Mutations may already have succeeded when a response is lost.
        retry: false,
      },
    },
  });
}

interface QueryProviderProps {
  children: ReactNode;
}

export function QueryProvider({ children }: QueryProviderProps) {
  const [queryClient] = useState(makeQueryClient);

  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
}
