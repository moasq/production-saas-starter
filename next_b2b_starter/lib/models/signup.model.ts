export interface SignupOwner {
  fullName: string;
  email: string;
}

export interface SignupOrganization {
  displayName: string;
}

export interface SignupDraft {
  owner: SignupOwner;
  organization: SignupOrganization;
}

export interface CreatedUser {
  id: number;
  email: string;
  fullName: string;
  createdAt: Date;
}

export interface AuthTokens {
  accessToken: string;
  tokenType: string;
  expiresIn?: number;
  refreshToken?: string;
}

export interface CreatedOrganization {
  id: number;
  name: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface SignupResult {
  orgId: string;
  orgName: string;
  displayName: string;
  ownerUserId: string;
  ownerEmail: string;
  ownerName: string;
  loginUrl: string;
}
