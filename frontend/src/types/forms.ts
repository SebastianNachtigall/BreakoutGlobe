// Form validation utilities and types

import type { LatLng, ValidationResult } from './models'
import { validateLatLng } from './models'

// Form data types
export interface CreatePOIForm {
  name: string
  description?: string
  position: LatLng
}

export interface CreateMapForm {
  name: string
  description?: string
}

export interface MoveAvatarForm {
  position: LatLng
}

// Form validation functions
export function validateCreatePOIForm(form: CreatePOIForm): ValidationResult {
  const errors: string[] = []

  if (!form.name || form.name.trim() === '') {
    errors.push('POI name is required')
  }

  if (form.name && form.name.length > 255) {
    errors.push('POI name must be 255 characters or less')
  }

  // Validate position
  const positionValidation = validateLatLng(form.position)
  if (!positionValidation.isValid) {
    errors.push(...positionValidation.errors)
  }

  return {
    isValid: errors.length === 0,
    errors
  }
}

export function validateCreateMapForm(form: CreateMapForm): ValidationResult {
  const errors: string[] = []

  if (!form.name || form.name.trim() === '') {
    errors.push('Map name is required')
  }

  if (form.name && form.name.length > 255) {
    errors.push('Map name must be 255 characters or less')
  }

  return {
    isValid: errors.length === 0,
    errors
  }
}

export function validateMoveAvatarForm(form: MoveAvatarForm): ValidationResult {
  return validateLatLng(form.position)
}

// Input sanitization functions
export function sanitizeString(input: string): string {
  return input.trim().replace(/\s+/g, ' ')
}

export function sanitizePOIName(name: string): string {
  return sanitizeString(name).substring(0, 255)
}

export function sanitizeMapName(name: string): string {
  return sanitizeString(name).substring(0, 255)
}

// Coordinate parsing and validation
export function parseCoordinates(latStr: string, lngStr: string): LatLng | null {
  const lat = parseFloat(latStr)
  const lng = parseFloat(lngStr)

  if (isNaN(lat) || isNaN(lng)) {
    return null
  }

  const coords: LatLng = { lat, lng }
  const validation = validateLatLng(coords)

  return validation.isValid ? coords : null
}

export function parseCoordinateString(coordStr: string): LatLng | null {
  const parts = coordStr.split(',').map(s => s.trim())
  
  if (parts.length !== 2) {
    return null
  }

  return parseCoordinates(parts[0], parts[1])
}