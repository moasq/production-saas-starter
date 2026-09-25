/**
 * Profile DTOs - Mirror backend ProfileResponse structure
 */

export interface ProfileOrganizationDto {
  organization_id: string;
  slug: string;
  name: string;
  status: string;
}

export interface ProfileResponseDto {
  // Verified member details
  member_id: string;
  email: string;
  name: string;
  roles: string[];
  permissions: string[];
  email_verified: boolean;
  status: string;

  // Organization details
  organization: ProfileOrganizationDto;

  // Internal account details
  account_id: number;
  created_at: string;
  updated_at: string;
}
