import { apiClient } from "../client/api-client";
import type { SignupMagicLinkResponseDto } from "../dto/auth.dto";
import type { SignupOrganization, SignupOwner, SignupResult } from "@/lib/models/signup.model";
export const signupRepository = {
 async createOrganizationWithMagicLink(owner: SignupOwner, organization: SignupOrganization): Promise<SignupResult> {
  const dto = await apiClient.post<SignupMagicLinkResponseDto>("/auth/signup", {
    org_display_name: organization.displayName, owner_email: owner.email, owner_name: owner.fullName,
  }, { skipAuth: true });
  if (!dto.magic_link_sent) throw new Error("Workspace created, but the sign-in email could not be sent. Try signing in again.");
  return { orgId: dto.org_id, orgName: dto.org_name ?? dto.display_name, displayName: dto.display_name,
    ownerUserId: "", ownerEmail: dto.owner_email, ownerName: dto.owner_name, loginUrl: "/auth" };
 }
};
