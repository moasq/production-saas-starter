/**
 * Profile Repository - Fetch current user profile with computed permissions
 */

import { apiClient } from "../client/api-client";
import type { ProfileResponseDto } from "../dto/profile.dto";

class ProfileRepository {
  // The server forwards the verified request cookie; browsers use their HttpOnly cookie.
  async getProfile(cookieHeader?: string): Promise<ProfileResponseDto> {
    const options = cookieHeader
      ? {
          headers: {
            Cookie: cookieHeader,
          },
        }
      : undefined;

    type ProfileApiResponse =
      | ProfileResponseDto
      | {
          data?: ProfileResponseDto;
          success?: boolean;
          message?: string;
        };

    const response = await apiClient.get<ProfileApiResponse>(
      "/auth/profile/me",
      options
    );

    // Backend wraps response in { data: {...}, success: true }
    // Extract the actual profile data
    const profileDto =
      (response as { data?: ProfileResponseDto }).data ??
      (response as ProfileResponseDto);

    if (!profileDto || !profileDto.member_id) {
      const errorMessage =
        (response as { message?: string }).message ||
        "Profile response did not include member information";
      throw new Error(errorMessage);
    }

    return profileDto;
  }
}

export const profileRepository = new ProfileRepository();
