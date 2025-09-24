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
  const formData = new FormData();
  formData.append('displayName', request.displayName);
  
  if (request.aboutMe) {
    formData.append('aboutMe', request.aboutMe);
  }
  
  if (request.avatarFile) {
    formData.append('avatar', request.avatarFile);
  }

  const response = await fetch(`${API_BASE_URL}/api/users/profile`, {
    method: 'POST',
    body: formData,
  });

  const apiProfile = await handleResponse<UserProfileAPI>(response);
  return transformUserProfileFromAPI(apiProfile);
}

export async function getCurrentUserProfile(): Promise<UserProfile | null> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/users/profile`, {
      method: 'GET',
      credentials: 'include',
    });

    if (response.status === 404) {
      return null; // No profile exists
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

export async function updateUserProfile(updates: Partial<Pick<UserProfile, 'displayName' | 'aboutMe'>>): Promise<UserProfile> {
  const response = await fetch(`${API_BASE_URL}/api/users/profile`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
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