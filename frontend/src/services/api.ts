import { UserProfile, UserProfileAPI, transformUserProfileFromAPI } from '../types/models';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

// POI API Types
export interface CreatePOIRequest {
  mapId: string;
  name: string;
  description: string;
  position: { lat: number; lng: number };
  createdBy: string;
  maxParticipants: number;
  image?: File; // Optional image file
}

export interface POIResponse {
  id: string;
  mapId: string;
  name: string;
  description: string;
  position: { lat: number; lng: number };
  createdBy: string;
  maxParticipants: number;
  participantCount?: number;
  participants?: Array<{ id: string; name: string; avatarUrl: string }>;
  imageUrl?: string;
  
  // Discussion timer fields
  discussionStartTime?: string;
  isDiscussionActive?: boolean;
  
  createdAt: string;
}

export interface POIListResponse {
  mapId: string;
  pois: POIResponse[];
  count: number;
}

export interface UpdatePOIRequest {
  name?: string;
  description?: string;
  maxParticipants?: number;
}

export interface JoinPOIRequest {
  userId: string;
}

export interface POIParticipantsResponse {
  poiId: string;
  participants: string[];
  count: number;
}

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
  console.log('🌐 API: createGuestProfile called with:', request);
  
  // First create the profile
  const requestBody = {
    displayName: request.displayName,
    accountType: 'guest', // Required by backend
    aboutMe: request.aboutMe || ''
  };

  console.log('📡 API: Sending request body to backend:', requestBody);

  const response = await fetch(`${API_BASE_URL}/api/users/profile`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(requestBody),
  });

  console.log('📨 API: Response status:', response.status);

  let profile = await handleResponse<UserProfileAPI>(response);
  console.log('📦 API: Raw response from backend:', profile);
  
  let transformedProfile = transformUserProfileFromAPI(profile);
  console.log('🔄 API: Transformed profile:', transformedProfile);

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

export async function getCurrentUserProfile(userId?: string): Promise<UserProfile | null> {
  try {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };
    
    // Add X-User-ID header if userId is provided
    if (userId) {
      headers['X-User-ID'] = userId;
    }
    
    const response = await fetch(`${API_BASE_URL}/api/users/profile`, {
      method: 'GET',
      headers,
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

export async function uploadAvatar(avatarFile: File, userId?: string): Promise<UserProfile> {
  const formData = new FormData();
  formData.append('avatar', avatarFile);

  const headers: Record<string, string> = {};
  
  // Add user ID header if provided
  if (userId) {
    headers['X-User-ID'] = userId;
  }

  const response = await fetch(`${API_BASE_URL}/api/users/avatar`, {
    method: 'POST',
    headers,
    credentials: 'include',
    body: formData,
  });

  const apiProfile = await handleResponse<UserProfileAPI>(response);
  return transformUserProfileFromAPI(apiProfile);
}

// POI API Functions

export async function createPOI(request: CreatePOIRequest): Promise<POIResponse> {
  console.log('🌐 API: createPOI called with:', request);

  let response: Response;

  // If image is provided, use multipart form data
  if (request.image) {
    const formData = new FormData();
    formData.append('mapId', request.mapId);
    formData.append('name', request.name);
    formData.append('description', request.description);
    formData.append('position.lat', request.position.lat.toString());
    formData.append('position.lng', request.position.lng.toString());
    formData.append('createdBy', request.createdBy);
    formData.append('maxParticipants', request.maxParticipants.toString());
    formData.append('image', request.image);

    response = await fetch(`${API_BASE_URL}/api/pois`, {
      method: 'POST',
      body: formData,
    });
  } else {
    // Use JSON for requests without image (existing functionality)
    const { image, ...jsonRequest } = request;
    response = await fetch(`${API_BASE_URL}/api/pois`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(jsonRequest),
    });
  }

  console.log('📨 API: POI creation response status:', response.status);

  const poiResponse = await handleResponse<POIResponse>(response);
  console.log('📦 API: POI created:', poiResponse);

  return poiResponse;
}

export async function getPOIs(mapId: string): Promise<POIResponse[]> {
  console.log('🌐 API: getPOIs called for mapId:', mapId);

  const response = await fetch(`${API_BASE_URL}/api/pois?mapId=${encodeURIComponent(mapId)}`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
  });

  const listResponse = await handleResponse<POIListResponse>(response);
  console.log('📦 API: POIs loaded:', listResponse.count, 'POIs');

  return listResponse.pois;
}

export async function getPOI(poiId: string): Promise<POIResponse> {
  console.log('🌐 API: getPOI called for poiId:', poiId);

  const response = await fetch(`${API_BASE_URL}/api/pois/${encodeURIComponent(poiId)}`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
  });

  const poiResponse = await handleResponse<POIResponse>(response);
  console.log('📦 API: POI loaded:', poiResponse);

  return poiResponse;
}

export async function updatePOI(poiId: string, updates: UpdatePOIRequest): Promise<POIResponse> {
  console.log('🌐 API: updatePOI called for poiId:', poiId, 'with updates:', updates);

  const response = await fetch(`${API_BASE_URL}/api/pois/${encodeURIComponent(poiId)}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(updates),
  });

  const poiResponse = await handleResponse<POIResponse>(response);
  console.log('📦 API: POI updated:', poiResponse);

  return poiResponse;
}

export async function deletePOI(poiId: string): Promise<void> {
  console.log('🌐 API: deletePOI called for poiId:', poiId);

  const response = await fetch(`${API_BASE_URL}/api/pois/${encodeURIComponent(poiId)}`, {
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json',
    },
  });

  await handleResponse<{ success: boolean; message: string }>(response);
  console.log('✅ API: POI deleted successfully');
}

export async function joinPOI(poiId: string, userId: string): Promise<void> {
  console.log('🌐 API: joinPOI called for poiId:', poiId, 'userId:', userId);

  const request: JoinPOIRequest = { userId };

  const response = await fetch(`${API_BASE_URL}/api/pois/${encodeURIComponent(poiId)}/join`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(request),
  });

  await handleResponse<{ success: boolean; poiId: string; userId: string }>(response);
  console.log('✅ API: Joined POI successfully');
}

export async function leavePOI(poiId: string, userId: string): Promise<void> {
  console.log('🌐 API: leavePOI called for poiId:', poiId, 'userId:', userId);

  const request = { userId };

  const response = await fetch(`${API_BASE_URL}/api/pois/${encodeURIComponent(poiId)}/leave`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(request),
  });

  await handleResponse<{ success: boolean; poiId: string; userId: string }>(response);
  console.log('✅ API: Left POI successfully');
}

export async function getPOIParticipants(poiId: string): Promise<string[]> {
  console.log('🌐 API: getPOIParticipants called for poiId:', poiId);

  const response = await fetch(`${API_BASE_URL}/api/pois/${encodeURIComponent(poiId)}/participants`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
  });

  const participantsResponse = await handleResponse<POIParticipantsResponse>(response);
  console.log('📦 API: POI participants loaded:', participantsResponse.count, 'participants');

  return participantsResponse.participants;
}

// Data Transformation Utilities

export function transformToCreatePOIRequest(
  formData: {
    name: string;
    description: string;
    maxParticipants: number;
    position: { lat: number; lng: number };
    image?: File;
  },
  userId: string,
  mapId: string = 'default-map'
): CreatePOIRequest {
  return {
    mapId,
    name: formData.name,
    description: formData.description,
    position: formData.position,
    createdBy: userId,
    maxParticipants: formData.maxParticipants,
    image: formData.image
  };
}

export function transformFromPOIResponse(apiResponse: POIResponse): {
  id: string;
  name: string;
  description?: string;
  position: { lat: number; lng: number };
  participantCount: number;
  maxParticipants: number;
  participants?: Array<{ id: string; name: string }>;
  imageUrl?: string;
  createdBy: string;
  createdAt: Date;
  // Discussion timer fields
  discussionStartTime?: Date | null;
  isDiscussionActive?: boolean;
} {
  return {
    id: apiResponse.id,
    name: apiResponse.name,
    description: apiResponse.description,
    position: apiResponse.position,
    participantCount: apiResponse.participantCount || 0,
    maxParticipants: apiResponse.maxParticipants,
    participants: apiResponse.participants,
    imageUrl: apiResponse.imageUrl,
    createdBy: apiResponse.createdBy,
    createdAt: new Date(apiResponse.createdAt),
    // Transform discussion timer fields - backend only tracks when 2+ users are present
    discussionStartTime: apiResponse.discussionStartTime ? new Date(apiResponse.discussionStartTime) : null,
    isDiscussionActive: apiResponse.isDiscussionActive || false
  };
}// D
evelopment helper function to clear all POIs
export async function clearAllPOIs(mapId: string = 'default-map'): Promise<void> {
  console.log('🧹 API: clearAllPOIs called for mapId:', mapId);

  const response = await fetch(`${API_BASE_URL}/api/pois/dev/clear-all?mapId=${encodeURIComponent(mapId)}`, {
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json',
    },
  });

  await handleResponse<{ success: boolean; message: string; mapId: string }>(response);
  console.log('🧹 API: All POIs cleared successfully');
}