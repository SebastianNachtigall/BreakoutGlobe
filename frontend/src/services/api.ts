import { UserProfile, UserProfileAPI, transformUserProfileFromAPI } from '../types/models';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

export interface CreateGuestProfileRequest {
  displayName: string;
  aboutMe?: string;
  avatarFile?: File;
}

export class APIError extends Error {
  constructor(
    message: string,
    public status: number,
    public code?: string
  ) {
    super(message);
    this.name = 'APIError';
  }
}

async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new APIError(
      errorData.message || `HTTP ${response.status}: ${response.statusText}`,
      response.status,
      errorData.code
    );
  }
  return response.json();
}

export async function createGuestProfile(request: CreateGuestProfileRequest): Promise<UserProfile> {
  // First create the profile
  const requestBody = {
    displayName: request.displayName,
    accountType: 'guest', // Required by backend
    aboutMe: request.aboutMe || ''
  };

  const response = await fetch(`${API_BASE_URL}/api/users/profile`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(requestBody),
  });

  let profile = await handleResponse<UserProfileAPI>(response);
  let transformedProfile = transformUserProfileFromAPI(profile);

  // If avatar file is provided, upload it after profile creation
  if (request.avatarFile) {
    try {
      // Set the user ID header for the avatar upload
      const formData = new FormData();
      formData.append('avatar', request.avatarFile);

      const avatarResponse = await fetch(`${API_BASE_URL}/api/users/avatar`, {
        method: 'POST',
        headers: {
          'X-User-ID': transformedProfile.id, // Backend expects user ID in header
        },
        body: formData,
      });

      const updatedProfile = await handleResponse<UserProfileAPI>(avatarResponse);
      transformedProfile = transformUserProfileFromAPI(updatedProfile);
    } catch (error) {
      console.warn('Avatar upload failed, but profile was created:', error);
      // Don't fail the entire profile creation if avatar upload fails
      // The user can upload an avatar later
    }
  }

  return transformedProfile;
}

export async function getCurrentUserProfile(): Promise<UserProfile | null> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/users/profile`, {
      method: 'GET',
      credentials: 'include',
    });

    if (response.status === 404) {
      // 404 is expected for new users - not an error
      return null;
    }

    const apiProfile = await handleResponse<UserProfileAPI>(response);
    return transformUserProfileFromAPI(apiProfile);
  } catch (error) {
    if (error instanceof APIError && error.status === 404) {
      return null;
    }
    throw error;
  }
}

export async function updateUserProfile(updates: Partial<Pick<UserProfile, 'displayName' | 'aboutMe'>>, userID?: string): Promise<UserProfile> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };
  
  // Add user ID header if provided
  if (userID) {
    headers['X-User-ID'] = userID;
  }

  const response = await fetch(`${API_BASE_URL}/api/users/profile`, {
    method: 'PUT',
    headers,
    credentials: 'include',
    body: JSON.stringify(updates),
  });

  const apiProfile = await handleResponse<UserProfileAPI>(response);
  return transformUserProfileFromAPI(apiProfile);
}

export async function uploadAvatar(avatarFile: File): Promise<UserProfile> {
  const formData = new FormData();
  formData.append('avatar', avatarFile);

  const response = await fetch(`${API_BASE_URL}/api/users/avatar`, {
    method: 'POST',
    credentials: 'include',
    body: formData,
  });

  const apiProfile = await handleResponse<UserProfileAPI>(response);
  return transformUserProfileFromAPI(apiProfile);
}