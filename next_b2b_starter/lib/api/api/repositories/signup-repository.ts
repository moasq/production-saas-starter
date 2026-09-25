import { signup } from "@/lib/actions/auth/signup";
import type { SignupOrganization, SignupOwner, SignupResult } from "@/lib/models/signup.model";
export const signupRepository = {
 async createOrganizationWithMagicLink(owner: SignupOwner, organization: SignupOrganization): Promise<SignupResult> {
  const result = await signup({ email: owner.email, name: owner.fullName, organizationName: organization.displayName });
  if (!result.success) throw new Error(result.error);
  return { orgId: "", orgName: organization.displayName, displayName: organization.displayName,
    ownerUserId: "", ownerEmail: owner.email, ownerName: owner.fullName, loginUrl: "/auth" };
 }
};
