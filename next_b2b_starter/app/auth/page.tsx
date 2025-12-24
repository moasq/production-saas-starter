"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { useStytchMember } from "@stytch/nextjs/b2b";
import { ArrowRight, CheckCircle2, Home, Inbox, Mail } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { sendMagicLink } from "@/lib/actions/auth/send-magic-link";


const highlights = [
  "Single workspace to review invoices, approvals, and exports.",
  "Ready-made controls that plug into your existing banking stack.",
  "Sessions scoped to your organization with role-aware access.",
];

const emailProviders = [
  {
    label: "Open Gmail",
    href: "https://mail.google.com/",
  },
  {
    label: "Open Outlook",
    href: "https://outlook.office.com/mail/",
  },
  {
    label: "Open iCloud Mail",
    href: "https://www.icloud.com/mail",
  },
  {
    label: "Open Yahoo Mail",
    href: "https://mail.yahoo.com/",
  },
];

export default function AuthPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { member, isInitialized } = useStytchMember();
  const [email, setEmail] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [status, setStatus] = useState<{
    type: "info" | "error" | "success";
    message: string;
  } | null>(null);
  const [isRedirecting, setIsRedirecting] = useState(false);
  const [view, setView] = useState<"form" | "success">("form");
  const [lastSubmittedEmail, setLastSubmittedEmail] = useState("");
  const hasRedirectedRef = useRef(false);
  const redirectTimeoutRef = useRef<number | null>(null);

  const targetAfterLogin = useMemo(() => {
    const returnTo = searchParams.get("returnTo") || "/dashboard";
    // Validate returnTo is a safe relative path:
    // - Must start with single /
    // - Cannot be protocol-relative (//), contain backslash (\), or have : before first /
    const isSafeRelativePath =
      returnTo.startsWith("/") &&
      !returnTo.startsWith("//") &&
      !returnTo.includes("\\") &&
      !returnTo.slice(1).includes(":");
    return isSafeRelativePath ? returnTo : "/dashboard";
  }, [searchParams]);

  const handleAuthSuccess = useCallback(() => {
    if (hasRedirectedRef.current) return;
    hasRedirectedRef.current = true;
    setIsRedirecting(true);
    setStatus({
      type: "info",
      message: "You’re signed in. Redirecting to your workspace…",
    });
    router.replace(targetAfterLogin);
    setTimeout(() => {
      router.refresh();
    }, 150);
    if (typeof window !== "undefined") {
      if (redirectTimeoutRef.current !== null) {
        window.clearTimeout(redirectTimeoutRef.current);
      }
      redirectTimeoutRef.current = window.setTimeout(() => {
        window.location.assign(targetAfterLogin);
      }, 1500);
    }
  }, [router, targetAfterLogin]);

  useEffect(() => {
    if (!isInitialized) return;
    if (member) {
      handleAuthSuccess();
    }
  }, [isInitialized, member, handleAuthSuccess]);

  useEffect(() => {
    const hasMagicLinkParams =
      searchParams.has("stytch_token") ||
      searchParams.has("token") ||
      searchParams.has("stytch_token_type");

    if (hasMagicLinkParams) {
      setStatus({
        type: "info",
        message: "We’re verifying your sign-in link. This usually takes just a moment.",
      });
    }
  }, [searchParams]);

  const submitEmail = useCallback(
    async (
      rawEmail: string,
      options: { resetField?: boolean; stayOnSuccessView?: boolean } = {}
    ) => {
      const { resetField = true, stayOnSuccessView = false } = options;
      const trimmedEmail = rawEmail.trim().toLowerCase();

      if (!trimmedEmail) {
        setStatus({
          type: "error",
          message: "Please enter a valid email address.",
        });
        return;
      }

      setIsSubmitting(true);
      setStatus({
        type: "info",
        message: "Checking your workspace access…",
      });
      if (!stayOnSuccessView) {
        setView("form");
      }

      try {
        const query = new URLSearchParams({ email: trimmedEmail }).toString();
        const apiBaseUrl = (process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080/api").replace(/\/$/, "");

        const emailCheckResponse = await fetch(`${apiBaseUrl}/auth/check-email?${query}`);

        if (emailCheckResponse.status === 404) {
          const errorBody = await emailCheckResponse.json().catch(() => null);
          setStatus({
            type: "error",
            message:
              (errorBody && errorBody.message) ||
              "We couldn't find an account with that email. Try a different email or ask your admin to invite you.",
          });
          return;
        }

        if (!emailCheckResponse.ok) {
          const errorBody = await emailCheckResponse.json().catch(() => null);
          throw new Error(
            (errorBody && errorBody.message) ||
              "We couldn't verify that email right now. Please try again in a moment.",
          );
        }

        setStatus({
          type: "info",
          message: "Sending your secure sign-in link…",
        });

        // Call Server Action instead of API route
        const result = await sendMagicLink(trimmedEmail);

        if (!result.success) {
          throw new Error(result.error || "Failed to send sign-in link.");
        }

        setLastSubmittedEmail(trimmedEmail);
        setStatus({
          type: "success",
          message: "Check your email for a secure link to sign in.",
        });
        setView("success");

        if (resetField) {
          setEmail("");
        }
      } catch (error: any) {
        setStatus({
          type: "error",
          message:
            error?.message ||
            "Something went wrong while sending your sign-in link. Please try again.",
        });
      } finally {
        setIsSubmitting(false);
      }
    },
    []
  );

  const handleSendMagicLink = async (e: React.FormEvent) => {
    e.preventDefault();
    await submitEmail(email, { resetField: true });
  };

  const handleResend = async () => {
    if (!lastSubmittedEmail) return;
    await submitEmail(lastSubmittedEmail, {
      resetField: false,
      stayOnSuccessView: true,
    });
  };

  useEffect(() => {
    router.prefetch(targetAfterLogin);
  }, [router, targetAfterLogin]);

  useEffect(() => {
    return () => {
      if (typeof window !== "undefined" && redirectTimeoutRef.current !== null) {
        window.clearTimeout(redirectTimeoutRef.current);
        redirectTimeoutRef.current = null;
      }
    };
  }, []);

  if (!isInitialized) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gray-50">
        <div className="flex flex-col items-center gap-3">
          <div className="h-10 w-10 animate-spin rounded-full border-4 border-gray-200 border-t-primary-500" />
          <p className="text-sm text-gray-600">
            Checking your workspace session…
          </p>
        </div>
      </div>
    );
  }

  if (isRedirecting || member) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gray-50">
        <div className="flex flex-col items-center gap-3 rounded-xl border border-gray-200 bg-white px-8 py-10 shadow-lg">
          <div className="h-10 w-10 animate-spin rounded-full border-4 border-gray-200 border-t-primary-500" />
          <div className="text-center">
            <p className="text-sm font-medium text-gray-900">
              Redirecting to your dashboard
            </p>
            <p className="text-sm text-gray-600">
              {(status && status.message) ||
                "We’re setting up your workspace now."}
            </p>
            <p className="mt-4 text-xs text-gray-500">
              Taking longer than expected?{" "}
              <a
                href={targetAfterLogin}
                className="font-medium text-primary-600 hover:underline"
              >
                Open your workspace
              </a>
              .
            </p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-b from-gray-50 via-white to-gray-50 py-16">
      <div className="mx-auto flex w-full max-w-5xl flex-col gap-10 px-6 lg:flex-row lg:items-start">
        <section className="flex-1 space-y-8">
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <Badge
                variant="outline"
                className="w-fit items-center gap-2 border-primary/30 bg-primary/5 text-xs font-medium text-primary-700"
              >
                <Mail className="h-3.5 w-3.5" aria-hidden />
                Secure email sign-in
              </Badge>
              <Link href="/" className="flex items-center gap-2 text-sm text-gray-600 hover:text-primary-600 transition-colors">
                <Home className="h-4 w-4" />
                <span>Home</span>
              </Link>
            </div>
            <div className="space-y-4">
              <h1 className="text-3xl font-semibold text-gray-900 sm:text-4xl">
                Welcome back to Your App
              </h1>
              <p className="max-w-xl text-base text-gray-600">
                Use your work email to receive a one-time, organization-aware
                sign-in link. We’ll land you back where you left off as soon as
                you’re authenticated.
              </p>
            </div>
          </div>
          <dl className="space-y-3">
            {highlights.map((item) => (
              <div
                key={item}
                className="flex items-start gap-3 rounded-xl border border-gray-200 bg-white p-4 shadow-sm"
              >
                <CheckCircle2
                  className="mt-1 h-5 w-5 flex-none text-primary-600"
                  aria-hidden
                />
                <p className="text-sm text-gray-600">{item}</p>
              </div>
            ))}
          </dl>
          <div className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
            <h2 className="text-sm font-semibold text-gray-900">
              Need a hand?
            </h2>
            <p className="mt-2 text-sm text-gray-600">
              Reach out at{" "}
              <a
                className="font-medium text-primary-600 hover:underline"
                href="mailto:support@yourapp.com"
              >
                support@yourapp.com
              </a>{" "}
              for support, or check the documentation in your workspace.
            </p>
          </div>
        </section>

        <aside className="w-full max-w-md lg:sticky lg:top-24">
          <div className="rounded-2xl border border-gray-200 bg-white p-8 shadow-lg">
            <div className="mb-6 space-y-2 text-center">
              <h2 className="text-2xl font-semibold text-gray-900">
                Sign in to Your App
              </h2>
              <p className="text-sm text-gray-600">
                Enter your work email to receive a secure sign-in link. We’ll
                handle redirects automatically.
              </p>
            </div>
            {view === "form" ? (
              <>
                {status && status.type !== "success" && (
                  <Alert
                    variant={status.type === "error" ? "destructive" : "default"}
                    className="mb-6 text-left"
                  >
                    <AlertTitle>
                      {status.type === "error" ? "We hit a snag" : "Hang tight"}
                    </AlertTitle>
                    <AlertDescription>{status.message}</AlertDescription>
                  </Alert>
                )}
                <form onSubmit={handleSendMagicLink} className="space-y-4">
                  <div className="space-y-2 text-left">
                    <label htmlFor="email" className="text-sm font-medium text-gray-700">
                      Work email address
                    </label>
                    <Input
                      id="email"
                      type="email"
                      placeholder="you@company.com"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      disabled={isSubmitting}
                      required
                      className="w-full"
                      autoComplete="email"
                      inputMode="email"
                    />
                  </div>
                  <Button
                    type="submit"
                    disabled={isSubmitting || !email.trim()}
                    className="w-full"
                  >
                    {isSubmitting ? "Working…" : "Email me a sign-in link"}
                  </Button>
                </form>
                <p className="mt-6 text-center text-xs text-gray-400">
                  By continuing you agree to the terms of service and
                  acknowledge the privacy notice.
                </p>
                <p className="mt-4 text-center text-sm text-gray-600">
                  Don't have an account?{" "}
                  <Link href="/signup" className="text-primary-600 hover:underline font-medium">
                    Sign up
                  </Link>
                </p>
              </>
            ) : (
              <div className="space-y-6 text-center">
                <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-full bg-primary-50 text-primary-600">
                  <Inbox className="h-7 w-7" aria-hidden />
                </div>
                <div className="space-y-2">
                  <h3 className="text-xl font-semibold text-gray-900">
                    Check your email
                  </h3>
                  <p className="text-sm text-gray-600">
                    We sent a secure link to {" "}
                    <span className="font-medium text-gray-900">
                      {lastSubmittedEmail}
                    </span>
                    . Open it on any device to finish signing in.
                  </p>
                </div>
                <div className="space-y-2">
                  {emailProviders.map((provider) => (
                    <Button key={provider.href} variant="secondary" className="w-full justify-between" asChild>
                      <a href={provider.href} target="_blank" rel="noreferrer">
                        <span>{provider.label}</span>
                        <ArrowRight className="h-4 w-4" aria-hidden />
                      </a>
                    </Button>
                  ))}
                  <p className="text-xs text-gray-500">
                    Prefer another inbox? Open your mail app and look for an
                    email from Your App security.
                  </p>
                </div>
                {status && status.type === "error" && (
                  <Alert variant="destructive" className="text-left">
                    <AlertTitle>We couldn’t send the link</AlertTitle>
                    <AlertDescription>{status.message}</AlertDescription>
                  </Alert>
                )}
                <div className="flex flex-col gap-3 sm:flex-row sm:justify-center">
                  <Button
                    variant="outline"
                    onClick={() => {
                      setView("form");
                      setStatus(null);
                      setEmail("");
                    }}
                    disabled={isSubmitting}
                  >
                    Use a different email
                  </Button>
                  <Button onClick={handleResend} disabled={isSubmitting}>
                    Resend link
                  </Button>
                </div>
                {status && status.type !== "error" && status.message && (
                  <p className="text-xs text-gray-500">{status.message}</p>
                )}
              </div>
            )}
            <div className="mt-8 flex items-center justify-center gap-2 text-xs text-gray-400">
              <span>Powered by</span>
              <span className="font-semibold text-gray-600">Stytch</span>
            </div>
          </div>
        </aside>
      </div>
    </div>
  );
}
