import { parseInvitationResponse } from "../dto/invitation-response.ts";
import { apiClient, unwrap, type ApiClient } from "../client/api-client.ts";
import type { MemberDto } from "../dto/member.dto.ts";
import type { ProfileResponseDto } from "../dto/profile.dto.ts";
import type { OrganizationMember, UserProfile, InviteMemberRequest, InviteMemberResponse, UpdateProfileRequest, MemberListResponse, MemberRole } from "@/lib/models/member.model";

export class MemberRepository {
  private client: ApiClient;
  constructor(client: ApiClient = apiClient) { this.client = client; }

  async getProfile(): Promise<UserProfile> {
    return this.toUserProfile(unwrap(await this.client.GET("/auth/profile/me")).data);
  }
  async updateProfile(request: UpdateProfileRequest): Promise<UserProfile> {
    if (!request.name) throw new Error("Name is required");
    unwrap(await this.client.PUT("/auth/profile/me", { body: { name: request.name } }));
    return this.getProfile();
  }
  async getMembers(): Promise<MemberListResponse> {
    const { data } = unwrap(await this.client.GET("/auth/members"));
    return { members: data.members.map(item => this.toOrganizationMember(item)), totalCount: data.total };
  }
  async inviteMember(request: InviteMemberRequest, _organizationId: string): Promise<InviteMemberResponse> {
    void _organizationId; // Tenant comes only from the verified session.
    return parseInvitationResponse(unwrap(await this.client.POST("/auth/members", {
      body: { email: request.email, name: request.name, role_slug: request.role },
    })));
  }
  async removeMember(memberId: string): Promise<boolean> {
    unwrap(await this.client.DELETE("/auth/members/{member_id}", { params: { path: { member_id: memberId } } }));
    return true;
  }
  async updateRole(memberId: string, role: MemberRole): Promise<void> {
    unwrap(await this.client.PUT("/auth/members/{member_id}", { params: { path: { member_id: memberId } }, body: { role } }));
  }
  async resendInvitation(memberId: string): Promise<boolean> {
    const response = unwrap(await this.client.POST("/auth/members/{member_id}/resend-invitation", { params: { path: { member_id: memberId } } }));
    if (!response.data.invite_sent) throw new Error("Invitation could not be sent.");
    return true;
  }
  private toUserProfile(dto: ProfileResponseDto): UserProfile {
    return { id: dto.member_id, email: dto.email, name: dto.name, role: dto.roles[0] || "member", organizationId: dto.organization.organization_id, organizationName: dto.organization.name };
  }
  private toOrganizationMember(dto: MemberDto): OrganizationMember {
    return { id: dto.member_id, email: dto.email, name: dto.name, role: dto.roles[0] || "member", status: dto.status, joinedAt: new Date(dto.created_at) };
  }
}
export const memberRepository = new MemberRepository();
