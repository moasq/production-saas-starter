import { parseInvitationResponse } from "../dto/invitation-response.ts";
// lib/api/api/repositories/member-repository.ts

import { apiClient, type ApiClient } from "../client/api-client.ts";
import type {
  MemberListResponseDto,
  InviteMemberRequestDto,
  InviteMemberResponseDto,
  MemberDto,
  UpdateProfileRequestDto,
  UpdateProfileResponseDto,
  ResendInvitationResponseDto,
} from "../dto/member.dto";
import type { ProfileResponseDto } from "../dto/profile.dto";
import type {
  OrganizationMember,
  UserProfile,
  InviteMemberRequest,
  InviteMemberResponse,
  UpdateProfileRequest,
  MemberListResponse,
  MemberRole,
} from "@/lib/models/member.model";

export class MemberRepository {
  private client: ApiClient;
  constructor(client: ApiClient = apiClient) { this.client = client; }
  /**
   * Get current user profile
   */
  async getProfile(): Promise<UserProfile> {
    type ProfileApiResponse =
      | ProfileResponseDto
      | {
          data?: ProfileResponseDto;
          profile?: ProfileResponseDto;
          success?: boolean;
          message?: string;
        };

    const response = await this.client.get<ProfileApiResponse>("/auth/profile/me");

    const profileDto =
      (response as { data?: ProfileResponseDto }).data ??
      (response as { profile?: ProfileResponseDto }).profile ??
      (response as ProfileResponseDto);

    if (!profileDto || !profileDto.member_id) {
      const errorMessage =
        (response as { message?: string }).message ||
        "Profile response did not include member information";
      throw new Error(errorMessage);
    }

    return this.toUserProfile(profileDto);
  }

  /**
   * Update user profile
   */
  async updateProfile(request: UpdateProfileRequest): Promise<UserProfile> {
    const payload: UpdateProfileRequestDto = {
      name: request.name,
    };

    type UpdateProfileApiResponse = UpdateProfileResponseDto & {
      profile?: ProfileResponseDto;
      data?: ProfileResponseDto;
    };

    const response = await this.client.put<UpdateProfileApiResponse>(
      "/auth/profile/me",
      payload
    );

    if (response.profile || response.data) {
      return this.toUserProfile(response.profile ?? response.data!);
    }

    if (!response.success) {
      throw new Error(response.message || "Failed to update profile");
    }

    // Fetch updated profile after successful update
    return this.getProfile();
  }

  /**
   * Get organization members list
   * You can optionally scope results by organization and pagination settings.
   */
  async getMembers(options?: {
    organizationId?: string;
    page?: number;
    pageSize?: number;
  }): Promise<MemberListResponse> {
    const params = new URLSearchParams();

    if (options?.organizationId) {
      params.append("organization_id", options.organizationId);
    }

    if (options?.page) {
      params.append("page", String(options.page));
    }

    if (options?.pageSize) {
      params.append("page_size", String(options.pageSize));
    }

    const queryString = params.toString();
    const endpoint = queryString ? `/auth/members?${queryString}` : "/auth/members";

    type MemberListApiResponse =
      | MemberListResponseDto
      | {
          data?: MemberListResponseDto;
          members?: MemberDto[];
          total?: number;
          success?: boolean;
        };

    const response = await this.client.get<MemberListApiResponse>(endpoint);

    const dto: MemberListResponseDto = "data" in response && response.data
      ? response.data
      : {
          members:
            (response as { members?: MemberDto[] }).members ??
            (response as MemberListResponseDto).members,
          total:
            (response as { total?: number }).total ??
            (response as MemberListResponseDto).total,
        };

    return {
      members: (dto.members ?? []).map((item) => this.toOrganizationMember(item)),
      totalCount: dto.total ?? dto.members?.length ?? 0,
    };
  }

  /**
   * Invite a new member
   */
  async inviteMember(
    request: InviteMemberRequest,
    _organizationId: string
  ): Promise<InviteMemberResponse> {
    void _organizationId; // Scope comes from the verified session, never this UI argument.
    const payload: InviteMemberRequestDto = {
      email: request.email,
      name: request.name,
      role_slug: request.role,
    };

    const response = await this.client.post<InviteMemberResponseDto>(
      "/auth/members",
      payload
    );

    return parseInvitationResponse(response);
  }

  /**
   * Remove member from organization
   */
  async removeMember(memberId: string): Promise<boolean> {
    await this.client.delete<void>(`/auth/members/${encodeURIComponent(memberId)}`);
    return true;
  }

  async updateRole(memberId: string, role: MemberRole): Promise<void> {
    await this.client.put(`/auth/members/${encodeURIComponent(memberId)}`, { role });
  }

  /**
   * Resend invitation to pending member
   */
  async resendInvitation(memberId: string): Promise<boolean> {
    const response = await this.client.post<ResendInvitationResponseDto>(
      `/auth/members/${encodeURIComponent(memberId)}/resend-invitation`,
      { member_id: memberId }
    );

    if (!response.success || !response.data?.invite_sent) throw new Error(response.message || "Invitation could not be sent.");
    return true;
  }

  /**
   * Transform DTO to UserProfile model
   * Extracts first non-provider role from roles array as the primary role
   */
  private toUserProfile(dto: ProfileResponseDto): UserProfile {
    // The bridge returns one canonical organization role
    const roles = dto.roles || [];
    const primaryRole =
      roles[0] || "member";
    const normalizedRole: MemberRole = ["admin", "manager", "member"].includes(
      primaryRole
    )
      ? (primaryRole as MemberRole)
      : "member";

    return {
      id: dto.member_id,
      email: dto.email,
      name: dto.name,
      role: normalizedRole,
      organizationId: dto.organization?.organization_id || "",
      organizationName: dto.organization?.name || "",
    };
  }

  /**
   * Transform DTO to OrganizationMember model
   * Extracts first non-provider role from roles array as the primary role
   */
  private toOrganizationMember(dto: MemberDto): OrganizationMember {
    // The bridge returns one canonical organization role
    const roles = dto.roles || [];
    const primaryRole =
      roles[0] || "member";
    const normalizedRole: MemberRole = ["admin", "manager", "member"].includes(
      primaryRole
    )
      ? (primaryRole as MemberRole)
      : "member";

    return {
      id: dto.member_id,
      email: dto.email,
      name: dto.name,
      role: normalizedRole,
      status: dto.status as "active" | "pending" | "inactive",
      joinedAt: new Date(dto.created_at),
    };
  }
}

export const memberRepository = new MemberRepository();
