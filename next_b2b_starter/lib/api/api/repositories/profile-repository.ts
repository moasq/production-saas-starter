import { apiClient, unwrap } from "../client/api-client.ts";
import type { ProfileResponseDto } from "../dto/profile.dto.ts";

class ProfileRepository {
  async getProfile(cookieHeader?: string): Promise<ProfileResponseDto> {
    const response = unwrap(await apiClient.GET("/auth/profile/me", {
      headers: cookieHeader ? { Cookie: cookieHeader } : undefined,
    }));
    return response.data;
  }
}
export const profileRepository = new ProfileRepository();
