"use client";

import Link from "next/link";
import { useMemo, useCallback, useTransition } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { buildLoginUrl } from "@/lib/auth/stytch-client";
import { useStytchConfig } from "@/lib/contexts/stytch-config-context";
import { usePermissions } from "@/lib/hooks/use-permissions";
import { useAuthContext } from "@/lib/contexts/auth-context";
import { resetCachedToken } from "@/lib/api/api/client/api-client";
import { logout } from "@/lib/actions/auth/logout";

function getInitials(name?: string) {
  if (!name) return "?";
  const parts = name.trim().split(/\s+/);
  const first = parts[0]?.[0] || "";
  const last = parts.length > 1 ? parts[parts.length - 1][0] : "";
  return (first + last).toUpperCase();
}

export function UserMenu() {
  const { profile, isInitialized } = usePermissions();
  const authContext = useAuthContext();
  const queryClient = useQueryClient();
  const stytchConfig = useStytchConfig();
  const [isPending, startTransition] = useTransition();

  const loginHref = useMemo(() => {
    return buildLoginUrl(stytchConfig);
  }, [stytchConfig]);

  const handleLogout = useCallback(() => {
    startTransition(async () => {
      // Clear all client-side state
      authContext?.clearAuthState();
      queryClient.clear();
      resetCachedToken();

      // Call Server Action (will redirect to home page)
      await logout("/");
    });
  }, [authContext, queryClient]);

  if (!isInitialized) {
    return (
      <div className="h-9 w-24 animate-pulse rounded-md bg-gray-100" aria-label="Loading user" />
    );
  }

  if (!profile) {
    return (
      <Button asChild variant="default" className="h-9">
        <a href={loginHref}>Log in</a>
      </Button>
    );
  }

  const display = profile.name || profile.email || "Account";
  const initials = getInitials(profile.name || profile.email);

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" className="h-9 gap-2">
          <span className="flex h-6 w-6 items-center justify-center rounded-full bg-gray-900 text-xs font-semibold text-white">
            {initials}
          </span>
          <span className="hidden max-w-[160px] truncate text-sm md:inline">{display}</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-52">
        <DropdownMenuItem asChild>
          <Link href="/dashboard">Dashboard</Link>
        </DropdownMenuItem>
        <DropdownMenuItem onClick={handleLogout} disabled={isPending}>
          {isPending ? "Logging out..." : "Log out"}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
