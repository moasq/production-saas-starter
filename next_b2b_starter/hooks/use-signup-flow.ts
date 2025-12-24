import { useCallback, useMemo, useRef, useState } from "react";
import { signupRepository } from "@/lib/api/api/repositories/signup-repository";
import {
  SignupOrganization,
  SignupOwner,
  SignupResult,
} from "@/lib/models/signup.model";

export type SignupStep = "account" | "organization";

interface UseSignupFlowState {
  step: SignupStep;
  owner: SignupOwner;
  organization: SignupOrganization;
  isLoading: boolean;
  error: string | null;
  emailSent: boolean;
  result: SignupResult | null;
  stepIndex: number;
  canContinueAccount: boolean;
  canContinueOrganization: boolean;
  goBack: () => void;
  goNext: () => void;
  sendMagicLink: () => Promise<void>;
  updateOwner: (updates: Partial<SignupOwner>) => void;
  updateOrganization: (updates: Partial<SignupOrganization>) => void;
  reset: () => void;
}

const defaultOwner: SignupOwner = {
  fullName: "",
  email: "",
};

const defaultOrganization: SignupOrganization = {
  displayName: "",
  industry: "Technology",
};

export function useSignupFlow(): UseSignupFlowState {
  const [step, setStep] = useState<SignupStep>("account");
  const [owner, setOwner] = useState<SignupOwner>(defaultOwner);
  const [organization, setOrganization] = useState<SignupOrganization>(
    defaultOrganization
  );
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [emailSent, setEmailSent] = useState(false);
  const [result, setResult] = useState<SignupResult | null>(null);

  // Ref to prevent duplicate submissions
  const isSubmittingRef = useRef(false);

  const stepIndex = useMemo(() => {
    switch (step) {
      case "account":
        return 0;
      case "organization":
        return 1;
      default:
        return 0;
    }
  }, [step]);

  const canContinueAccount = useMemo(() => {
    return (
      owner.fullName.trim().length >= 2 &&
      /.+@.+\..+/.test(owner.email)
    );
  }, [owner]);

  const canContinueOrganization = useMemo(() => {
    return (
      organization.displayName.trim().length >= 2 &&
      organization.industry.trim().length > 0
    );
  }, [organization]);

  const goBack = useCallback(() => {
    setError(null);
    if (step === "organization") {
      setStep("account");
    }
  }, [step]);

  const goNext = useCallback(() => {
    setError(null);
    if (step === "account" && canContinueAccount) {
      setStep("organization");
    }
  }, [step, canContinueAccount]);

  const updateOwner = useCallback((updates: Partial<SignupOwner>) => {
    setOwner((prev) => ({ ...prev, ...updates }));
    setError(null); // Clear error when user types
  }, []);

  const updateOrganization = useCallback((updates: Partial<SignupOrganization>) => {
    setOrganization((prev) => ({ ...prev, ...updates }));
    setError(null); // Clear error when user types
  }, []);

  const reset = useCallback(() => {
    setOwner(defaultOwner);
    setOrganization(defaultOrganization);
    setStep("account");
    setError(null);
    setEmailSent(false);
    setResult(null);
  }, []);

  const sendMagicLink = useCallback(async () => {
    // Prevent duplicate submissions
    if (isSubmittingRef.current) return;

    if (!canContinueOrganization) {
      setError("Please fill in all required fields correctly");
      return;
    }

    isSubmittingRef.current = true;
    setIsLoading(true);
    setError(null);

    try {
      // Backend signup endpoint already sends magic link via Stytch
      // No need to call sendMagicLink() Server Action separately
      const signupResult = await signupRepository.createOrganizationWithMagicLink(
        owner,
        organization
      );

      setResult(signupResult);
      setEmailSent(true);
    } catch (signupError) {
      const message =
        signupError instanceof Error
          ? signupError.message
          : "Failed to create account. Please try again.";
      setError(message);
      setEmailSent(false);
    } finally {
      setIsLoading(false);
      isSubmittingRef.current = false;
    }
  }, [owner, organization, canContinueOrganization]);

  return {
    step,
    owner,
    organization,
    isLoading,
    error,
    emailSent,
    result,
    stepIndex,
    canContinueAccount,
    canContinueOrganization,
    goBack,
    goNext,
    sendMagicLink,
    updateOwner,
    updateOrganization,
    reset,
  };
}
