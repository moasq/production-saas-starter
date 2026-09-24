"use client";

import { useSignupFlow } from "@/hooks/use-signup-flow";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ArrowRight, Home, Inbox } from "lucide-react";
import Link from "next/link";

export default function SignupPage() {
  const {
    step,
    owner,
    organization,
    isLoading,
    error,
    emailSent,
    canContinueAccount,
    canContinueOrganization,
    goNext,
    goBack,
    sendMagicLink,
    updateOwner,
    updateOrganization,
  } = useSignupFlow();

  // Success view after email sent
  if (emailSent) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 px-6">
        <div className="w-full max-w-md text-center space-y-6">
          <div className="mx-auto h-14 w-14 bg-primary-50 rounded-full flex items-center justify-center">
            <Inbox className="h-7 w-7 text-primary-600" />
          </div>
          <h1 className="text-2xl font-semibold text-gray-900">Check your email</h1>
          <p className="text-gray-600">
            We sent a verification link to <strong>{owner.email}</strong>.
            Click the link to complete your signup.
          </p>
          <Link href="/auth">
            <Button variant="outline">Back to Sign In</Button>
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 px-6">
      <div className="w-full max-w-md">
        <div className="bg-white p-8 rounded-2xl shadow-lg border border-gray-200">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h1 className="text-2xl font-semibold text-gray-900">
                Create your account
              </h1>
              <p className="text-sm text-gray-600 mt-1">
                Get started with Your App
              </p>
            </div>
            <Link href="/" className="flex items-center gap-1.5 text-xs text-gray-500 hover:text-primary-600 transition-colors">
              <Home className="h-3.5 w-3.5" />
              <span>Home</span>
            </Link>
          </div>

          {error && (
            <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          {/* Step 1: Account */}
          {step === "account" && (
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Full Name
                </label>
                <Input
                  type="text"
                  placeholder="John Doe"
                  value={owner.fullName}
                  onChange={(e) => updateOwner({ fullName: e.target.value })}
                  disabled={isLoading}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Email
                </label>
                <Input
                  type="email"
                  placeholder="you@company.com"
                  value={owner.email}
                  onChange={(e) => updateOwner({ email: e.target.value })}
                  disabled={isLoading}
                />
              </div>
              <Button
                onClick={goNext}
                disabled={!canContinueAccount || isLoading}
                className="w-full"
              >
                Continue <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
            </div>
          )}

          {/* Step 2: Organization */}
          {step === "organization" && (
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Organization Name
                </label>
                <Input
                  type="text"
                  placeholder="Acme Inc"
                  value={organization.displayName}
                  onChange={(e) => updateOrganization({ displayName: e.target.value })}
                  disabled={isLoading}
                />
              </div>

              <div className="flex gap-3">
                <Button
                  variant="outline"
                  onClick={goBack}
                  disabled={isLoading}
                  className="flex-1"
                >
                  Back
                </Button>
                <Button
                  onClick={sendMagicLink}
                  disabled={!canContinueOrganization || isLoading}
                  className="flex-1"
                >
                  {isLoading ? "Creating..." : "Create Account"}
                </Button>
              </div>
            </div>
          )}

          {!emailSent && (
            <p className="mt-6 text-center text-sm text-gray-600">
              Already have an account?{" "}
              <Link href="/auth" className="text-primary-600 hover:underline font-medium">
                Sign in
              </Link>
            </p>
          )}
        </div>
      </div>
    </div>
  );
}
