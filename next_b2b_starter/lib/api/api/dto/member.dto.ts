// lib/api/api/dto/member.dto.ts

/**
 * Member Management DTOs
 * These match the expected backend API structure
 */

// Member List DTOs
export interface MemberListRequestDto {
  organization_id: string;
  page?: number;
  page_size?: number;
  status?: "active" | "pending" | "inactive";
}

export interface MemberListResponseDto {
  members: MemberDto[];
  total: number;
}

export interface MemberDto {
  member_id: string;
  email: string;
  name: string;
  roles: string[];
  status: string;
  email_verified: boolean;
  created_at: string;
  updated_at: string;
}

// Invite Member DTOs
export interface InviteMemberRequestDto {
  email: string;
  name: string;
  role_slug: string;
}

export interface InviteMemberResponseDto {
  success: boolean;
  data?: { member_id: string; invite_sent: boolean };
  message?: string;
}

// Remove Member DTOs
export interface RemoveMemberRequestDto {
  member_id: string;
  organization_id: string;
}

export interface RemoveMemberResponseDto {
  success: boolean;
  message?: string;
}

// Profile DTOs
export interface UpdateProfileRequestDto {
  name?: string;
  avatar_url?: string;
}

export interface UpdateProfileResponseDto {
  success: boolean;
  message?: string;
}

// Resend Invitation DTO
export interface ResendInvitationRequestDto {
  member_id: string;
}

export interface ResendInvitationResponseDto {
  data?: { invite_sent: boolean };
  success: boolean;
  message?: string;
}
