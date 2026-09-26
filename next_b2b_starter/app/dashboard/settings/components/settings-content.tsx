"use client";

import { useEffect, useMemo, useState, useRef } from "react";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import {
  User,
  Users,
  CreditCard,
  ArrowLeft,
  ChevronRight,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { toast } from "sonner";

import { ProfileSection } from "./profile-section";
import { MemberList } from "./member-list";
import { InviteMember } from "./invite-member";
import { MemberHelpers } from "@/lib/models/member.model";
import { usePermissions } from "@/lib/hooks/use-permissions";
import { PERMISSIONS } from "@/lib/auth/permissions";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import type { SubscriptionGateState } from "@/lib/polar/current-subscription";
import { SubscriptionTab } from "./subscription-tab";

// Query hooks - Component depends ONLY on these hooks
import { useProfileQuery } from "@/lib/hooks/queries/use-profile-query";
import { useMembersQuery } from "@/lib/hooks/queries/use-members-query";
import { useSubscriptionQuery } from "@/lib/hooks/queries/use-subscription-query";
import { useInviteMember } from "@/lib/hooks/mutations/use-invite-member";
import type { InviteMemberRequest } from "@/lib/models/member.model";
import { usePathname, useRouter, useSearchParams } from "next/navigation";

interface SettingsContentProps {
  billingEnabled: boolean;
  // No props required - component fetches its own data
}

type SettingsView = "overview" | "profile" | "members" | "subscription";

interface OverviewSection {
  key: Exclude<SettingsView, "overview">;
  title: string;
  description: string;
  value: string;
  helper: string;
  icon: LucideIcon;
  disabled?: boolean;
}

const DETAIL_META: Record<Exclude<SettingsView, "overview">, { title: string; description: string }> = {
  profile: {
    title: "Account & workspace",
    description: "Update your profile details and workspace metadata.",
  },
  members: {
    title: "Team access",
    description: "Invite new teammates, manage existing members, and adjust permissions.",
  },
  subscription: {
    title: "Subscription & billing",
    description: "Review your subscription status, usage limits, and cancellation controls.",
  },
};

function parseViewParam(raw: string | null): SettingsView | null {
  if (!raw) return null;
  const normalized = raw.toLowerCase();
  if (normalized === "profile" || normalized === "members" || normalized === "subscription") {
    return normalized as SettingsView;
  }
  return null;
}

function getSubscriptionQuickStatus(
  state: SubscriptionGateState | null,
  isLoading: boolean
) {
  if (isLoading && !state) {
    return {
      title: "Loading…",
      helper: "Fetching your Polar subscription.",
    };
  }

  if (!state) {
    return {
      title: "No active plan",
      helper: "Choose a plan below to unlock automations.",
    };
  }

  if (state.reason === "POLAR_UNCONFIGURED") {
    return {
      title: "Setup required",
      helper: "Add Polar credentials in the environment to enable billing.",
    };
  }

  if (state.reason === "BACKEND_UNAVAILABLE") {
    return {
      title: "Temporarily unavailable",
      helper: "We're still connecting to Polar. Try refreshing shortly.",
    };
  }

  if (!state.isActive || state.reason === "NO_ACTIVE_SUBSCRIPTION") {
    return {
      title: "No active plan",
      helper: "Select a subscription plan below.",
    };
  }

  if (state.subscription?.cancelAtPeriodEnd) {
    const cancellationDate = state.subscription.currentPeriodEnd
      ? new Date(state.subscription.currentPeriodEnd).toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" })
      : null;

    return {
      title: "Cancels soon",
      helper: cancellationDate
        ? `Ends on ${cancellationDate}. Update your plan below to stay active.`
        : "Scheduled to cancel at period end.",
      };
  }

  const planLabel = state.subscription?.productName || "Active plan";
  const renewalDate = state.subscription?.currentPeriodEnd
    ? new Date(state.subscription.currentPeriodEnd).toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" })
    : null;

  return {
    title: planLabel,
    helper: renewalDate ? `Renews on ${renewalDate}.` : "Billing is managed through Polar.",
  };
}

export function SettingsContent({ billingEnabled }: SettingsContentProps) {
  const {
    hasPermission,
    isInitialized: permissionsReady,
  } = usePermissions();
  const canManageMembers = hasPermission(PERMISSIONS.ORG_MANAGE);
  const hasSubscriptionPermission = billingEnabled && hasPermission(PERMISSIONS.ORG_MANAGE);
  const shouldLoadSubscription = permissionsReady && hasSubscriptionPermission;

  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const viewParam = searchParams.get("view");
  const paymentVerified = searchParams.get("payment_verified");
  const paymentError = searchParams.get("payment_error");

  const [viewStack, setViewStack] = useState<SettingsView[]>(["overview"]);
  const currentView = viewStack[viewStack.length - 1];
  const [isInviteModalOpen, setInviteModalOpen] = useState(false);
  const inviteTrigger = useRef<HTMLButtonElement>(null);

  // Track if we've shown the payment toast to prevent duplicates
  const paymentToastShownRef = useRef(false);

  // Handle payment verification feedback
  useEffect(() => {
    if (paymentToastShownRef.current) return;

    if (paymentVerified === "true") {
      paymentToastShownRef.current = true;
      toast.success("Subscription activated successfully!", {
        description: "Your payment has been verified and your subscription is now active.",
      });

      // Clean up the URL params
      const params = new URLSearchParams(searchParams.toString());
      params.delete("payment_verified");
      const queryString = params.toString();
      router.replace(queryString ? `${pathname}?${queryString}` : pathname, { scroll: false });
    } else if (paymentError === "true") {
      paymentToastShownRef.current = true;
      toast.error("Payment verification issue", {
        description: "We couldn't verify your payment immediately. Your subscription should activate shortly.",
      });

      // Clean up the URL params
      const params = new URLSearchParams(searchParams.toString());
      params.delete("payment_error");
      const queryString = params.toString();
      router.replace(queryString ? `${pathname}?${queryString}` : pathname, { scroll: false });
    }
  }, [paymentVerified, paymentError, router, pathname, searchParams]);

  useEffect(() => {
    if (!permissionsReady) return;
    const requested = parseViewParam(viewParam);
    if (!requested) return;
    if (requested === "members" && !canManageMembers) return;
    if (requested === "subscription" && !hasSubscriptionPermission) return;

    setViewStack((stack) => {
      if (stack[stack.length - 1] === requested) {
        return stack;
      }
      return ["overview", requested];
    });
  }, [permissionsReady, canManageMembers, hasSubscriptionPermission, viewParam]);

  const pushView = (view: Exclude<SettingsView, "overview">) => {
    setViewStack((stack) => {
      if (stack[stack.length - 1] === view) {
        return stack;
      }
      return [...stack, view];
    });
  };

  const goBack = () => {
    setViewStack((stack) => (stack.length > 1 ? stack.slice(0, -1) : stack));
  };

  useEffect(() => {
    const desiredParam = currentView === "overview" ? null : currentView;
    if (desiredParam === viewParam) {
      return;
    }
    const params = new URLSearchParams(searchParams.toString());
    if (desiredParam) {
      params.set("view", desiredParam);
    } else {
      params.delete("view");
    }
    const queryString = params.toString();
    router.replace(queryString ? `${pathname}?${queryString}` : pathname, { scroll: false });
  }, [currentView, viewParam, router, pathname, searchParams]);

  // Use query hooks - data is cached and reused globally
  const {
    data: profile,
    isLoading: isProfileLoading,
    error: profileError,
    refetch: refetchProfile,
  } = useProfileQuery({
    enabled: permissionsReady,
  });

  const {
    data: membersData,
    isLoading: isMembersLoading,
    isFetching: isMembersFetching,
    error: membersError,
    refetch: refetchMembers,
  } = useMembersQuery(
    {
      organizationId: profile?.organizationId,
      page: 1,
      pageSize: 50,
      enabled: canManageMembers && Boolean(profile?.organizationId),
    }
  );

  const {
    data: subscriptionState,
    isLoading: isSubscriptionLoading,
    error: subscriptionError,
    refetch: refetchSubscription,
  } = useSubscriptionQuery({
    enabled: shouldLoadSubscription,
  });

  // Mutations
  const inviteMemberMutation = useInviteMember();

  // Memoized values - MUST be before any early returns (React Hooks rules)
  const members = membersData?.members ?? [];
  const organizationId = profile?.organizationId ?? "";
  const hasOrganization = Boolean(organizationId);
  const canInviteMembers = canManageMembers && hasOrganization;
  const canViewMembers = canManageMembers && hasOrganization;

  useEffect(() => {
    if (currentView === "subscription" && !hasSubscriptionPermission) {
      setViewStack(["overview"]);
    }
  }, [currentView, hasSubscriptionPermission]);

  useEffect(() => {
    if (currentView === "members" && (!canManageMembers || !hasOrganization)) {
      setViewStack(["overview"]);
    }
  }, [currentView, canManageMembers, hasOrganization]);

  const subscriptionQuick = useMemo(() => {
    if (!hasSubscriptionPermission) {
      return null;
    }

    return getSubscriptionQuickStatus(
      subscriptionState ?? null,
      isSubscriptionLoading
    );
  }, [hasSubscriptionPermission, subscriptionState, isSubscriptionLoading]);

  const roleConfig = useMemo(
    () => profile ? MemberHelpers.getRoleConfig(profile.role) : MemberHelpers.getRoleConfig("member"),
    [profile]
  );

  const membersErrorMessage = membersError?.message ?? null;
  const subscriptionErrorMessage = subscriptionError?.message ?? null;

  const overviewSections = useMemo<OverviewSection[]>(() => {
    if (!profile) {
      return [];
    }

    const sections: OverviewSection[] = [];

    const accountLabel =
      profile.name?.trim().length
        ? profile.name
        : profile.email ?? "Account";
    const workspaceLabel =
      profile.organizationName?.trim().length
        ? profile.organizationName
        : "No workspace assigned";

    sections.push({
      key: "profile",
      title: "Account & workspace",
      description: "Profile identity, workspace label, and contact details.",
      value: accountLabel,
      helper: `${workspaceLabel} • ${roleConfig.label}`,
      icon: User,
    });

    if (canManageMembers) {
      const disabled = !hasOrganization;

      let value = "Invite teammates";
      let helper = "Bring collaborators into the workflow.";

      if (disabled) {
        value = "No organization";
        helper = "Join or create an organization to manage team access.";
      } else if (membersErrorMessage) {
        value = "Needs attention";
        helper = membersErrorMessage;
      } else if (isMembersLoading && members.length === 0) {
        value = "Loading…";
        helper = "Fetching your team roster.";
      } else if (members.length > 0) {
        value = `${members.length} ${members.length === 1 ? "member" : "members"}`;
        helper = "Manage roles, invitations, and permissions.";
      }

      sections.push({
        key: "members",
        title: "Team access",
        description: "Invite teammates and fine-tune their permissions.",
        value,
        helper,
        icon: Users,
        disabled,
      });
    }

    if (hasSubscriptionPermission) {
      let value = "Open details";
      let helper = "Review your subscription and payment history.";

      if (subscriptionErrorMessage) {
        value = "Needs attention";
        helper = subscriptionErrorMessage;
      } else if (subscriptionQuick) {
        value = subscriptionQuick.title;
        helper = subscriptionQuick.helper;
      } else if (isSubscriptionLoading) {
        value = "Loading…";
        helper = "Fetching your Polar subscription.";
      }

      sections.push({
        key: "subscription",
        title: "Subscription & billing",
        description: "Manage your plan and billing history.",
        value,
        helper,
        icon: CreditCard,
      });
    }

    return sections;
  }, [
    profile,
    roleConfig.label,
    canManageMembers,
    hasOrganization,
    members.length,
    membersErrorMessage,
    isMembersLoading,
    hasSubscriptionPermission,
    subscriptionQuick,
    isSubscriptionLoading,
    subscriptionErrorMessage,
  ]);

  // Handle invite member
  const handleInvite = async (request: InviteMemberRequest) => {
    if (!profile?.organizationId) {
      return;
    }

    const result = await inviteMemberMutation.mutateAsync({
      request,
      organizationId: profile.organizationId,
    });
    setInviteModalOpen(false);
    return result;
    // Members list automatically refetches due to invalidation in mutation
  };

  // Handle refresh members (manual refetch)
  const handleRefreshMembers = () => {
    refetchMembers();
  };

  // Handle refresh subscription
  const handleRefreshSubscription = async () => {
    await refetchSubscription();
  };

  const isOverview = currentView === "overview";
  const activeDetailMeta =
    currentView === "overview"
      ? null
      : DETAIL_META[currentView as Exclude<SettingsView, "overview">];
  const activeSectionSummary = overviewSections.find(
    (section) => section.key === currentView
  );

  const renderDetailContent = () => {
    if (!profile) {
      return null;
    }

    switch (currentView) {
      case "profile":
        return (
          <div className="space-y-6">
            <ProfileSection profile={profile} />
          </div>
        );
      case "members":
        if (!canManageMembers || !hasOrganization) {
          return (
            <Alert className="border border-amber-200 bg-amber-50">
              <AlertTitle>Organization required</AlertTitle>
              <AlertDescription>
                Join or create an organization to manage team members.
              </AlertDescription>
            </Alert>
          );
        }
        return (
          <>
            <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
              <div className="space-y-1">
                <h3 className="text-xl font-semibold text-gray-900">Team roster</h3>
                <p className="text-sm text-gray-600">
                  Review every teammate in your workspace and keep roles current.
                </p>
              </div>
              <Button
                ref={inviteTrigger}
                onClick={() => setInviteModalOpen(true)}
                disabled={!canInviteMembers}
                className="w-full bg-gray-900 text-white hover:bg-gray-800 sm:w-auto"
              >
                Add member
              </Button>
            </div>

            {membersError && (
              <Alert
                variant="destructive"
                className="border border-red-200 bg-red-50"
              >
                <AlertTitle>Error</AlertTitle>
                <AlertDescription>
                  {membersError.message || "Failed to load team members"}
                </AlertDescription>
              </Alert>
            )}

            {isMembersFetching && members.length > 0 && (
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <span className="inline-flex h-3 w-3 animate-spin rounded-full border-2 border-gray-200 border-t-primary-500" />
                Refreshing team roster…
              </div>
            )}

            {canViewMembers ? (
              <MemberList
                members={members}
                canManage={canManageMembers}
                currentUserId={profile.id}
                organizationId={organizationId}
                isFetching={isMembersFetching}
                onMemberUpdate={handleRefreshMembers}
              />
            ) : (
              <p className="text-sm text-muted-foreground">
                Join or switch into an organization to manage members.
              </p>
            )}

            <Dialog
              open={isInviteModalOpen}
              onOpenChange={(open) => {
                if (inviteMemberMutation.isPending) {
                  return;
                }
                setInviteModalOpen(open);
              }}
            >
              <DialogContent id="invite-member-dialog" className="sm:max-w-lg"
                onCloseAutoFocus={(event) => { event.preventDefault(); inviteTrigger.current?.focus(); }}>
                <DialogHeader className="space-y-2 text-left">
                  <DialogTitle className="text-xl font-semibold text-gray-900">
                    Add a teammate
                  </DialogTitle>
                  <DialogDescription className="text-sm text-gray-600">
                    Send a secure invitation and assign the right access before they join.
                  </DialogDescription>
                </DialogHeader>
                <div className="pt-4">
                  <InviteMember
                    canInvite={canInviteMembers}
                    onInvite={handleInvite}
                  />
                </div>
              </DialogContent>
            </Dialog>
          </>
        );
      case "subscription":
        if (!hasSubscriptionPermission) {
          return (
            <Alert variant="destructive" className="border border-red-200 bg-red-50">
              <AlertTitle>Access restricted</AlertTitle>
              <AlertDescription>
                You don&apos;t have permission to manage subscription or billing settings.
              </AlertDescription>
            </Alert>
          );
        }

        return (
          <SubscriptionTab
            state={shouldLoadSubscription ? subscriptionState ?? null : null}
            isLoading={shouldLoadSubscription ? isSubscriptionLoading : false}
            error={shouldLoadSubscription ? subscriptionErrorMessage : null}
            onRefresh={handleRefreshSubscription}
          />
        );
      default:
        return null;
    }
  };

  // Loading state
  const isLoading = isProfileLoading || !permissionsReady;

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-10 w-64" />
        <Skeleton className="h-[600px] rounded-xl" />
      </div>
    );
  }

  // Error state
  if (profileError || !profile) {
    return (
      <div className="flex min-h-[400px] flex-col items-center justify-center space-y-4 rounded-xl border border-red-200 bg-red-50 px-6 py-12 text-center">
        <p className="text-sm font-medium text-red-900">
          {profileError?.message || "Failed to load settings"}
        </p>
        <Button variant="outline" onClick={() => refetchProfile()}>
          Try again
        </Button>
      </div>
    );
  }

  if (!isOverview && activeDetailMeta) {
    const SummaryIcon = activeSectionSummary?.icon ?? null;

    return (
      <div className="space-y-8">
        <div>
          <Button
            variant="ghost"
            size="sm"
            onClick={goBack}
            className="group inline-flex items-center gap-2 px-0 text-gray-600 hover:text-gray-900"
          >
            <ArrowLeft className="h-4 w-4 transition-transform group-hover:-translate-x-0.5" />
            Back
          </Button>
        </div>

        <div className="space-y-2">
          <h1 className="text-3xl font-semibold text-gray-900">
            {activeDetailMeta.title}
          </h1>
          <p className="text-sm text-gray-600">
            {activeDetailMeta.description}
          </p>
        </div>

        {activeSectionSummary && SummaryIcon ? (
          <div className="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm">
            <div className="flex min-w-0 items-start gap-4">
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-gray-100">
                <SummaryIcon className="h-5 w-5 text-gray-600" />
              </div>
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.2em] text-gray-500">
                  {activeSectionSummary.title}
                </p>
                <p className="mt-2 text-lg font-semibold text-gray-900">
                  {activeSectionSummary.value}
                </p>
                <p className="mt-1 text-sm text-gray-600">
                  {activeSectionSummary.helper}
                </p>
              </div>
            </div>
          </div>
        ) : null}

        <div className="space-y-6">
          {renderDetailContent()}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold text-gray-900">Workspace settings</h1>
        <p className="text-sm text-gray-600">
          Open a section below to review the full details without the clutter.
        </p>
      </div>

      <div className="overflow-hidden rounded-3xl border border-gray-200 bg-white shadow-sm">
        <ul className="divide-y divide-gray-100">
          {overviewSections.map((section) => {
            const SectionIcon = section.icon;
            const isDisabled = Boolean(section.disabled);

            return (
              <li key={section.key}>
                <button
                  type="button"
                  onClick={() => {
                    if (!isDisabled) {
                      pushView(section.key);
                    }
                  }}
                  disabled={isDisabled}
                  className={`flex w-full flex-col items-start justify-between gap-4 px-6 py-5 text-left transition sm:flex-row sm:gap-6 ${
                    isDisabled
                      ? "cursor-not-allowed opacity-60"
                      : "hover:bg-gray-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-gray-900/10"
                  }`}
                >
                  <div className="flex min-w-0 items-start gap-4">
                    <div className="flex h-10 w-10 items-center justify-center rounded-full bg-gray-100">
                      <SectionIcon className="h-5 w-5 text-gray-600" />
                    </div>
                    <div className="min-w-0 space-y-1">
                      <p className="text-sm font-semibold text-gray-900">
                        {section.title}
                      </p>
                      <p className="text-sm text-gray-600">
                        {section.description}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-4">
                    <div className="min-w-0 break-words text-left sm:text-right">
                      <p className="text-base font-semibold text-gray-900">
                        {section.value}
                      </p>
                      <p className="mt-1 text-xs text-gray-500">
                        {section.helper}
                      </p>
                    </div>
                    <ChevronRight className="h-4 w-4 text-gray-400" aria-hidden="true" />
                  </div>
                </button>
              </li>
            );
          })}
        </ul>
      </div>
    </div>
  );
}
