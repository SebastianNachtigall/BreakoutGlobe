import { describe, it, expect } from 'vitest'
import {
  validateCreatePOIForm,
  validateCreateMapForm,
  validateMoveAvatarForm,
  sanitizeString,
  sanitizePOIName,
  sanitizeMapName,
  parseCoordinates,
  parseCoordinateString
} from './forms'
import type { CreatePOIForm, CreateMapForm, MoveAvatarForm } from './forms'

describe('Form Validation', () => {
  describe('validateCreatePOIForm', () => {
    it('should validate correct POI form', () => {
      const form: CreatePOIForm = {
        name: 'Meeting Room A',
        description: 'A great place to meet',
        position: { lat: 40.7128, lng: -74.0060 }
      }
      
      const result = validateCreatePOIForm(form)
      
      expect(result.isValid).toBe(true)
      expect(result.errors).toEqual([])
    })

    it('should validate POI form without description', () => {
      const form: CreatePOIForm = {
        name: 'Meeting Room A',
        position: { lat: 40.7128, lng: -74.0060 }
      }
      
      const result = validateCreatePOIForm(form)
      
      expect(result.isValid).toBe(true)
      expect(result.errors).toEqual([])
    })

    it('should reject POI form with empty name', () => {
      const form: CreatePOIForm = {
        name: '',
        position: { lat: 40.7128, lng: -74.0060 }
      }
      
      const result = validateCreatePOIForm(form)
      
      expect(result.isValid).toBe(false)
      expect(result.errors).toContain('POI name is required')
    })

    it('should reject POI form with name too long', () => {
      const form: CreatePOIForm = {
        name: 'A'.repeat(256),
        position: { lat: 40.7128, lng: -74.0060 }
      }
      
      const result = validateCreatePOIForm(form)
      
      expect(result.isValid).toBe(false)
      expect(result.errors).toContain('POI name must be 255 characters or less')
    })

    it('should reject POI form with invalid position', () => {
      const form: CreatePOIForm = {
        name: 'Meeting Room A',
        position: { lat: 91, lng: -74.0060 }
      }
      
      const result = validateCreatePOIForm(form)
      
      expect(result.isValid).toBe(false)
      expect(result.errors).toContain('Latitude must be between -90 and 90')
    })
  })

  describe('validateCreateMapForm', () => {
    it('should validate correct map form', () => {
      const form: CreateMapForm = {
        name: 'Workshop Map 1',
        description: 'A map for the morning workshop'
      }
      
      const result = validateCreateMapForm(form)
      
      expect(result.isValid).toBe(true)
      expect(result.errors).toEqual([])
    })

    it('should validate map form without description', () => {
      const form: CreateMapForm = {
        name: 'Workshop Map 1'
      }
      
      const result = validateCreateMapForm(form)
      
      expect(result.isValid).toBe(true)
      expect(result.errors).toEqual([])
    })

    it('should reject map form with empty name', () => {
      const form: CreateMapForm = {
        name: ''
      }
      
      const result = validateCreateMapForm(form)
      
      expect(result.isValid).toBe(false)
      expect(result.errors).toContain('Map name is required')
    })

    it('should reject map form with name too long', () => {
      const form: CreateMapForm = {
        name: 'A'.repeat(256)
      }
      
      const result = validateCreateMapForm(form)
      
      expect(result.isValid).toBe(false)
      expect(result.errors).toContain('Map name must be 255 characters or less')
    })
  })

  describe('validateMoveAvatarForm', () => {
    it('should validate correct avatar move form', () => {
      const form: MoveAvatarForm = {
        position: { lat: 40.7128, lng: -74.0060 }
      }
      
      const result = validateMoveAvatarForm(form)
      
      expect(result.isValid).toBe(true)
      expect(result.errors).toEqual([])
    })

    it('should reject avatar move form with invalid position', () => {
      const form: MoveAvatarForm = {
        position: { lat: 91, lng: -74.0060 }
      }
      
      const result = validateMoveAvatarForm(form)
      
      expect(result.isValid).toBe(false)
      expect(result.errors).toContain('Latitude must be between -90 and 90')
    })
  })
})

describe('Input Sanitization', () => {
  describe('sanitizeString', () => {
    it('should trim whitespace', () => {
      const result = sanitizeString('  hello world  ')
      
      expect(result).toBe('hello world')
    })

    it('should collapse multiple spaces', () => {
      const result = sanitizeString('hello    world')
      
      expect(result).toBe('hello world')
    })

    it('should handle mixed whitespace', () => {
      const result = sanitizeString('  hello   world  ')
      
      expect(result).toBe('hello world')
    })

    it('should handle empty string', () => {
      const result = sanitizeString('')
      
      expect(result).toBe('')
    })
  })

  describe('sanitizePOIName', () => {
    it('should sanitize and limit length', () => {
      const longName = '  ' + 'A'.repeat(300) + '  '
      
      const result = sanitizePOIName(longName)
      
      expect(result).toBe('A'.repeat(255))
    })

    it('should handle normal names', () => {
      const result = sanitizePOIName('  Meeting Room A  ')
      
      expect(result).toBe('Meeting Room A')
    })
  })

  describe('sanitizeMapName', () => {
    it('should sanitize and limit length', () => {
      const longName = '  ' + 'A'.repeat(300) + '  '
      
      const result = sanitizeMapName(longName)
      
      expect(result).toBe('A'.repeat(255))
    })

    it('should handle normal names', () => {
      const result = sanitizeMapName('  Workshop Map 1  ')
      
      expect(result).toBe('Workshop Map 1')
    })
  })
})

describe('Coordinate Parsing', () => {
  describe('parseCoordinates', () => {
    it('should parse valid coordinates', () => {
      const result = parseCoordinates('40.7128', '-74.0060')
      
      expect(result).toEqual({ lat: 40.7128, lng: -74.0060 })
    })

    it('should return null for invalid numbers', () => {
      const result = parseCoordinates('invalid', '-74.0060')
      
      expect(result).toBeNull()
    })

    it('should return null for out-of-bounds coordinates', () => {
      const result = parseCoordinates('91', '-74.0060')
      
      expect(result).toBeNull()
    })

    it('should handle boundary coordinates', () => {
      const result = parseCoordinates('90', '180')
      
      expect(result).toEqual({ lat: 90, lng: 180 })
    })
  })

  describe('parseCoordinateString', () => {
    it('should parse valid coordinate string', () => {
      const result = parseCoordinateString('40.7128, -74.0060')
      
      expect(result).toEqual({ lat: 40.7128, lng: -74.0060 })
    })

    it('should handle coordinate string without spaces', () => {
      const result = parseCoordinateString('40.7128,-74.0060')
      
      expect(result).toEqual({ lat: 40.7128, lng: -74.0060 })
    })

    it('should return null for invalid format', () => {
      const result = parseCoordinateString('40.7128')
      
      expect(result).toBeNull()
    })

    it('should return null for too many parts', () => {
      const result = parseCoordinateString('40.7128, -74.0060, 100')
      
      expect(result).toBeNull()
    })

    it('should return null for invalid coordinates', () => {
      const result = parseCoordinateString('91, -74.0060')
      
      expect(result).toBeNull()
    })
  })
})