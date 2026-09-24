// lib/models/member.model.ts

/**
 * Member and User Management Models
 */

export type MemberRole = "admin" | "manager" | "member";
export type MemberStatus = "active" | "pending" | "inactive";

/**
 * Organization Member
 */
export interface OrganizationMember {
  id: string;
  email: string;
  name?: string;
  role: MemberRole;
  status: MemberStatus;
  joinedAt: Date;
}

/**
 * Current User Profile
 */
export interface UserProfile {
  id: string;
  email: string;
  name?: string;
  role: MemberRole;
  organizationId: string;
  organizationName: string;
}

/**
 * Member Invitation Request
 */
export interface InviteMemberRequest {
  email: string;
  name: string;
  role: MemberRole;
}

/**
 * Member Invitation Response
 */
export interface InviteMemberResponse {
  inviteSent: boolean;
  success: boolean;
  memberId?: string;
  message?: string;
}

/**
 * Update Profile Request
 */
export interface UpdateProfileRequest {
  name?: string;
}

/**
 * Member List Response
 */
export interface MemberListResponse {
  members: OrganizationMember[];
  totalCount: number;
}

/**
 * Helper functions for member management
 */
export const MemberHelpers = {
  /**
   * Get role display configuration
   */
  getRoleConfig: (role: MemberRole) => {
    const configs = {
      admin: {
        label: "Admin",
        color: "bg-blue-100 text-blue-700 border-blue-200",
        description: "Full system control and member management",
      },
      manager: {
        label: "Manager",
        color: "bg-emerald-100 text-emerald-700 border-emerald-200",
        description: "Workspace access; organization administration requires admin",
      },
      member: {
        label: "Member",
        color: "bg-gray-100 text-gray-700 border-gray-200",
        description: "Workspace access",
      },
    };
    return configs[role] || configs.member;
  },

  /**
   * Get status configuration
   */
  getStatusConfig: (status: MemberStatus) => {
    const configs = {
      active: {
        label: "Active",
        color: "bg-emerald-100 text-emerald-700",
        dotColor: "bg-emerald-500",
      },
      pending: {
        label: "Pending",
        color: "bg-amber-100 text-amber-700",
        dotColor: "bg-amber-500",
      },
      inactive: {
        label: "Inactive",
        color: "bg-gray-100 text-gray-500",
        dotColor: "bg-gray-400",
      },
    };
    return configs[status] || configs.inactive;
  },

  /**
   * Format member joined date
   */
  formatJoinedDate: (date: Date | string): string => {
    const d = typeof date === "string" ? new Date(date) : date;
    const now = new Date();
    const diffTime = Math.abs(now.getTime() - d.getTime());
    const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24));

    if (diffDays === 0) {
      return "Today";
    } else if (diffDays === 1) {
      return "Yesterday";
    } else if (diffDays < 7) {
      return `${diffDays} days ago`;
    } else if (diffDays < 30) {
      const weeks = Math.floor(diffDays / 7);
      return `${weeks} ${weeks === 1 ? "week" : "weeks"} ago`;
    } else if (diffDays < 365) {
      const months = Math.floor(diffDays / 30);
      return `${months} ${months === 1 ? "month" : "months"} ago`;
    } else {
      return d.toLocaleDateString("en-US", {
        month: "short",
        day: "numeric",
        year: "numeric",
      });
    }
  },
};
