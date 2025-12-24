"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { AlertCircle, CheckCircle2, Loader2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import { consumeMagicLink } from "@/lib/actions/auth/consume-magic-link";

const SESSION_DURATION_MINUTES = Number(
  process.env.NEXT_PUBLIC_STYTCH_SESSION_DURATION_MINUTES ?? "60"
) || 60;

const DEFAULT_DESTINATION = "/dashboard";

type StatusState = {
  state: "verifying" | "success" | "error";
  headline: string;
  message: string;
};

const INITIAL_STATUS: StatusState = {
  state: "verifying",
  headline: "We're verifying your magic link",
  message: "Hang tight—this usually takes just a moment.",
};

function extractErrorMessage(error: unknown): string {
  if (error && typeof error === "object") {
    const typed = error as any;

    if (typed.error_message) {
      return typed.error_message;
    }

    if (typed.message) {
      return typed.message;
    }
  }

  return "We couldn't verify that link. Please request a new magic link from the login page.";
}

export default function AuthenticateRedirectPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [status, setStatus] = useState<StatusState>(INITIAL_STATUS);

  const hasAttemptedAuthRef = useRef(false);

  const magicLinkToken = searchParams.get("stytch_token") || searchParams.get("token");
  const tokenType = searchParams.get("stytch_token_type") || searchParams.get("token_type");
  const returnTo = searchParams.get("returnTo")?.trim() || DEFAULT_DESTINATION;

  const redirectToDestination = useCallback(() => {
    router.push(returnTo);
    router.refresh();
  }, [returnTo, router]);

  const exchangeMagicLink = useCallback(async () => {
    if (!magicLinkToken) {
      setStatus({
        state: "error",
        headline: "Magic link is missing or invalid",
        message: "This sign-in link is missing its token. Please request a new magic link.",
      });
      return;
    }

    hasAttemptedAuthRef.current = true;
    setStatus(INITIAL_STATUS);

    try {
      const result = await consumeMagicLink(
        magicLinkToken,
        SESSION_DURATION_MINUTES
      );

      if (!result.success) {
        throw new Error(result.error || "Failed to verify magic link.");
      }

      if (!result.data.memberAuthenticated) {
        setStatus({
          state: "error",
          headline: "Additional verification required",
          message: "We need a bit more information to finish signing you in. Please continue from the login page.",
        });
        return;
      }

      setStatus({
        state: "success",
        headline: "Magic link verified",
        message: "You're all set. Redirecting you to your workspace…",
      });

      redirectToDestination();
    } catch (error) {
      setStatus({
        state: "error",
        headline: "We couldn't verify your link",
        message: extractErrorMessage(error),
      });
    }
  }, [magicLinkToken, redirectToDestination]);

  useEffect(() => {
    if (hasAttemptedAuthRef.current) return;
    void exchangeMagicLink();
  }, [exchangeMagicLink]);

  const icon =
    status.state === "success" ? (
      <CheckCircle2 className="h-10 w-10 text-green-500" aria-hidden="true" />
    ) : status.state === "error" ? (
      <AlertCircle className="h-10 w-10 text-red-500" aria-hidden="true" />
    ) : (
      <Loader2
        className="h-10 w-10 animate-spin text-primary-500"
        aria-hidden="true"
      />
    );

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4">
      <div className="w-full max-w-md rounded-xl border border-gray-200 bg-white px-8 py-10 text-center shadow-lg">
        <div className="flex flex-col items-center gap-4">
          {icon}
          <h1 className="text-lg font-semibold text-gray-900" role="status">
            {status.headline}
          </h1>
          <p className="text-sm text-gray-600">{status.message}</p>
          {status.state === "error" ? (
            <div className="mt-6 flex flex-col items-center gap-2">
              <Button asChild className="w-full justify-center">
                <Link href="/auth">Back to login</Link>
              </Button>
              <p className="text-xs text-gray-500">
                Need help? Contact your workspace admin or request a new magic
                link from the login page.
              </p>
            </div>
          ) : null}
        </div>
      </div>
    </div>
  );
}
