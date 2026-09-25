"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { AlertCircle, CheckCircle2, Loader2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import { completeAuthentication } from "@/lib/actions/auth/complete-login";


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
  const searchParams = useSearchParams();
  const [status, setStatus] = useState<StatusState>(INITIAL_STATUS);

  const hasAttemptedAuthRef = useRef(false);

  const signupId = searchParams.get("signup") || undefined;
  const linkError = searchParams.get("error");

  const exchangeMagicLink = useCallback(async () => {
    if (linkError) {
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
      const result = await completeAuthentication(signupId);

      if (!result.success) {
        throw new Error(result.error || "Failed to verify magic link.");
      }

      setStatus({
        state: "success",
        headline: "Magic link verified",
        message: "You're all set. Redirecting you to your workspace…",
      });

      window.location.assign(result.data.destination);
    } catch (error) {
      setStatus({
        state: "error",
        headline: "We couldn't verify your link",
        message: extractErrorMessage(error),
      });
    }
  }, [signupId, linkError]);

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
