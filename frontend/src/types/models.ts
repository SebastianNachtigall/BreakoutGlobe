// Core data types matching the backend models

export interface LatLng {
  lat: number
  lng: number
}

export interface Session {
  id: string
  userId: string
  mapId: string
  avatarPosition: LatLng
  createdAt: Date
  lastActive: Date
  isActive: boolean
}

export interface POI {
  id: string
  mapId: string
  name: string
  description?: string
  position: LatLng
  createdBy: string
  maxParticipants: number
  participantCount: number
  createdAt: Date
}

export interface Map {
  id: string
  name: string
  description?: string
  createdBy: string
  isActive: boolean
  createdAt: Date
  updatedAt: Date
}

export interface UserProfile {
  id: string
  displayName: string
  email?: string
  avatarURL?: string
  aboutMe?: string
  accountType: 'guest' | 'full'
  role: 'user' | 'admin' | 'superadmin'
  isActive: boolean
  emailVerified: boolean
  createdAt: Date
  lastActiveAt?: Date
}

// API response types (with string dates)
export interface SessionAPI {
  id: string
  userId: string
  mapId: string
  avatarPosition: LatLng
  createdAt: string
  lastActive: string
  isActive: boolean
}

export interface POIAPI {
  id: string
  mapId: string
  name: string
  description?: string
  position: LatLng
  createdBy: string
  maxParticipants: number
  participantCount: number
  createdAt: string
}

export interface MapAPI {
  id: string
  name: string
  description?: string
  createdBy: string
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface UserProfileAPI {
  id: string
  displayName: string
  email?: string
  avatarUrl?: string  // Note: backend uses lowercase 'u'
  aboutMe?: string
  accountType: 'guest' | 'full'
  role: 'user' | 'admin' | 'superadmin'
  isActive: boolean
  emailVerified: boolean
  createdAt: string
  lastActiveAt?: string
}

// Validation result type
export interface ValidationResult {
  isValid: boolean
  errors: string[]
}

// Validation functions
export function validateLatLng(coords: LatLng): ValidationResult {
  const errors: string[] = []

  if (coords.lat < -90 || coords.lat > 90) {
    errors.push('Latitude must be between -90 and 90')
  }

  if (coords.lng < -180 || coords.lng > 180) {
    errors.push('Longitude must be between -180 and 180')
  }

  return {
    isValid: errors.length === 0,
    errors
  }
}

export function validateSession(session: Session): ValidationResult {
  const errors: string[] = []

  if (!session.id || session.id.trim() === '') {
    errors.push('Session ID is required')
  }

  if (!session.userId || session.userId.trim() === '') {
    errors.push('User ID is required')
  }

  if (!session.mapId || session.mapId.trim() === '') {
    errors.push('Map ID is required')
  }

  // Validate avatar position
  const positionValidation = validateLatLng(session.avatarPosition)
  if (!positionValidation.isValid) {
    errors.push(...positionValidation.errors)
  }

  return {
    isValid: errors.length === 0,
    errors
  }
}

export function validatePOI(poi: POI): ValidationResult {
  const errors: string[] = []

  if (!poi.id || poi.id.trim() === '') {
    errors.push('POI ID is required')
  }

  if (!poi.mapId || poi.mapId.trim() === '') {
    errors.push('Map ID is required')
  }

  if (!poi.name || poi.name.trim() === '') {
    errors.push('POI name is required')
  }

  if (poi.name && poi.name.length > 255) {
    errors.push('POI name must be 255 characters or less')
  }

  if (!poi.createdBy || poi.createdBy.trim() === '') {
    errors.push('Created by is required')
  }

  if (poi.maxParticipants < 1 || poi.maxParticipants > 50) {
    errors.push('Max participants must be between 1 and 50')
  }

  // Validate position
  const positionValidation = validateLatLng(poi.position)
  if (!positionValidation.isValid) {
    errors.push(...positionValidation.errors)
  }

  return {
    isValid: errors.length === 0,
    errors
  }
}

export function validateMap(map: Map): ValidationResult {
  const errors: string[] = []

  if (!map.id || map.id.trim() === '') {
    errors.push('Map ID is required')
  }

  if (!map.name || map.name.trim() === '') {
    errors.push('Map name is required')
  }

  if (map.name && map.name.length > 255) {
    errors.push('Map name must be 255 characters or less')
  }

  if (!map.createdBy || map.createdBy.trim() === '') {
    errors.push('Created by is required')
  }

  return {
    isValid: errors.length === 0,
    errors
  }
}

// Data transformation functions for API responses
export function transformSessionFromAPI(apiSession: SessionAPI): Session {
  return {
    ...apiSession,
    createdAt: new Date(apiSession.createdAt),
    lastActive: new Date(apiSession.lastActive)
  }
}

export function transformPOIFromAPI(apiPOI: POIAPI): POI {
  return {
    ...apiPOI,
    createdAt: new Date(apiPOI.createdAt)
  }
}

export function transformMapFromAPI(apiMap: MapAPI): Map {
  return {
    ...apiMap,
    createdAt: new Date(apiMap.createdAt),
    updatedAt: new Date(apiMap.updatedAt)
  }
}

export function transformUserProfileFromAPI(apiProfile: UserProfileAPI): UserProfile {
  // Convert relative avatar URL to absolute URL
  let avatarURL = apiProfile.avatarUrl;
  if (avatarURL && !avatarURL.startsWith('http')) {
    const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
    avatarURL = `${API_BASE_URL}${avatarURL}`;
  }

  return {
    ...apiProfile,
    avatarURL: avatarURL, // Transform avatarUrl to avatarURL with absolute URL
    createdAt: new Date(apiProfile.createdAt),
    lastActiveAt: apiProfile.lastActiveAt ? new Date(apiProfile.lastActiveAt) : undefined
  };
}

// Utility functions for coordinate operations
export function calculateDistance(point1: LatLng, point2: LatLng): number {
  if (point1.lat === point2.lat && point1.lng === point2.lng) {
    return 0
  }

  const earthRadius = 6371 // Earth's radius in kilometers

  // Convert degrees to radians
  const lat1Rad = (point1.lat * Math.PI) / 180
  const lng1Rad = (point1.lng * Math.PI) / 180
  const lat2Rad = (point2.lat * Math.PI) / 180
  const lng2Rad = (point2.lng * Math.PI) / 180

  // Haversine formula
  const dlat = lat2Rad - lat1Rad
  const dlng = lng2Rad - lng1Rad

  const a =
    Math.sin(dlat / 2) * Math.sin(dlat / 2) +
    Math.cos(lat1Rad) * Math.cos(lat2Rad) * Math.sin(dlng / 2) * Math.sin(dlng / 2)

  const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a))

  return earthRadius * c
}

export function isWithinRadius(center: LatLng, point: LatLng, radiusKm: number): boolean {
  return calculateDistance(center, point) <= radiusKm
}

export function formatLatLng(coords: LatLng): string {
  return `${coords.lat.toFixed(4)},${coords.lng.toFixed(4)}`
}