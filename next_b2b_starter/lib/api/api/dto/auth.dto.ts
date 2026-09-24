export interface SignupMagicLinkResponseDto {
  message: string;
  org_id: string;
  org_name?: string;
  display_name: string;
  owner_email: string;
  owner_name: string;
  magic_link_sent: boolean;
}
