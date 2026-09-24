"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { type ReactNode } from "react";

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

// Browser: Create query client once
let browserQueryClient: QueryClient | undefined = undefined;

function getQueryClient() {
  if (typeof window === "undefined") {
    // Server: Always create a new query client
    return makeQueryClient();
  } else {
    // Browser: Create query client if it doesn't exist
    if (!browserQueryClient) {
      browserQueryClient = makeQueryClient();
    }
    return browserQueryClient;
  }
}

interface QueryProviderProps {
  children: ReactNode;
}

export function QueryProvider({ children }: QueryProviderProps) {
  // NOTE: Avoid useState when initializing the query client if you don't
  // have a suspense boundary between this and the code that may
  // suspend because React will throw away the client on the initial
  // render if it suspends and there is no boundary
  const queryClient = getQueryClient();

  return (
    <QueryClientProvider client={queryClient}>
      {children}
      {process.env.NODE_ENV === "development" && (
        <ReactQueryDevtools
          initialIsOpen={false}
          buttonPosition="bottom-right"
        />
      )}
    </QueryClientProvider>
  );
}
