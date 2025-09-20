import { describe, it, expect } from 'vitest'
import { 
  validateLatLng, 
  validateSession, 
  validatePOI, 
  validateMap,
  transformSessionFromAPI,
  transformPOIFromAPI,
  transformMapFromAPI
} from './models'
import type { LatLng, Session, POI, Map } from './models'

describe('LatLng Validation', () => {
  it('should validate correct coordinates', () => {
    const validCoords: LatLng = { lat: 40.7128, lng: -74.0060 }
    
    const result = validateLatLng(validCoords)
    
    expect(result.isValid).toBe(true)
    expect(result.errors).toEqual([])
  })

  it('should validate boundary coordinates', () => {
    const boundaryCoords: LatLng = { lat: 90, lng: 180 }
    
    const result = validateLatLng(boundaryCoords)
    
    expect(result.isValid).toBe(true)
    expect(result.errors).toEqual([])
  })

  it('should reject invalid latitude - too high', () => {
    const invalidCoords: LatLng = { lat: 91, lng: 0 }
    
    const result = validateLatLng(invalidCoords)
    
    expect(result.isValid).toBe(false)
    expect(result.errors).toContain('Latitude must be between -90 and 90')
  })

  it('should reject invalid latitude - too low', () => {
    const invalidCoords: LatLng = { lat: -91, lng: 0 }
    
    const result = validateLatLng(invalidCoords)
    
    expect(result.isValid).toBe(false)
    expect(result.errors).toContain('Latitude must be between -90 and 90')
  })

  it('should reject invalid longitude - too high', () => {
    const invalidCoords: LatLng = { lat: 0, lng: 181 }
    
    const result = validateLatLng(invalidCoords)
    
    expect(result.isValid).toBe(false)
    expect(result.errors).toContain('Longitude must be between -180 and 180')
  })

  it('should reject invalid longitude - too low', () => {
    const invalidCoords: LatLng = { lat: 0, lng: -181 }
    
    const result = validateLatLng(invalidCoords)
    
    expect(result.isValid).toBe(false)
    expect(result.errors).toContain('Longitude must be between -180 and 180')
  })

  it('should reject multiple invalid coordinates', () => {
    const invalidCoords: LatLng = { lat: 91, lng: 181 }
    
    const result = validateLatLng(invalidCoords)
    
    expect(result.isValid).toBe(false)
    expect(result.errors).toHaveLength(2)
    expect(result.errors).toContain('Latitude must be between -90 and 90')
    expect(result.errors).toContain('Longitude must be between -180 and 180')
  })
})

describe('Session Validation', () => {
  it('should validate correct session', () => {
    const validSession: Session = {
      id: 'session-123',
      userId: 'user-456',
      mapId: 'map-789',
      avatarPosition: { lat: 40.7128, lng: -74.0060 },
      createdAt: new Date(),
      lastActive: new Date(),
      isActive: true
    }
    
    const result = validateSession(validSession)
    
    expect(result.isValid).toBe(true)
    expect(result.errors).toEqual([])
  })

  it('should reject session with empty ID', () => {
    const invalidSession: Session = {
      id: '',
      userId: 'user-456',
      mapId: 'map-789',
      avatarPosition: { lat: 40.7128, lng: -74.0060 },
      createdAt: new Date(),
      lastActive: new Date(),
      isActive: true
    }
    
    const result = validateSession(invalidSession)
    
    expect(result.isValid).toBe(false)
    expect(result.errors).toContain('Session ID is required')
  })

  it('should reject session with empty user ID', () => {
    const invalidSession: Session = {
      id: 'session-123',
      userId: '',
      mapId: 'map-789',
      avatarPosition: { lat: 40.7128, lng: -74.0060 },
      createdAt: new Date(),
      lastActive: new Date(),
      isActive: true
    }
    
    const result = validateSession(invalidSession)
    
    expect(result.isValid).toBe(false)
    expect(result.errors).toContain('User ID is required')
  })

  it('should reject session with empty map ID', () => {
    const invalidSession: Session = {
      id: 'session-123',
      userId: 'user-456',
      mapId: '',
      avatarPosition: { lat: 40.7128, lng: -74.0060 },
      createdAt: new Date(),
      lastActive: new Date(),
      isActive: true
    }
    
    const result = validateSession(invalidSession)
    
    expect(result.isValid).toBe(false)
    expect(result.errors).toContain('Map ID is required')
  })

  it('should reject session with invalid avatar position', () => {
    const invalidSession: Session = {
      id: 'session-123',
      userId: 'user-456',
      mapId: 'map-789',
      avatarPosition: { lat: 91, lng: -74.0060 },
      createdAt: new Date(),
      lastActive: new Date(),
      isActive: true
    }
    
    const result = validateSession(invalidSession)
    
    expect(result.isValid).toBe(false)
    expect(result.errors).toContain('Latitude must be between -90 and 90')
  })
})

describe('POI Validation', () => {
  it('should validate correct POI', () => {
    const validPOI: POI = {
      id: 'poi-123',
      mapId: 'map-789',
      name: 'Meeting Room A',
      description: 'A great place to meet',
      position: { lat: 40.7128, lng: -74.0060 },
      createdBy: 'user-456',
      maxParticipants: 10,
      participantCount: 0,
      createdAt: new Date()
    }
    
    const result = validatePOI(validPOI)
    
    expect(result.isValid).toBe(true)
    expect(result.errors).toEqual([])
  })

  it('should validate POI without description', () => {
    const validPOI: POI = {
      id: 'poi-123',
      mapId: 'map-789',
      name: 'Meeting Room A',
      position: { lat: 40.7128, lng: -74.0060 },
      createdBy: 'user-456',
      maxParticipants: 10,
      participantCount: 0,
      createdAt: new Date()
    }
    
    const result = validatePOI(validPOI)
    
    expect(result.isValid).toBe(true)
    expect(result.errors).toEqual([])
  })

  it('should reject POI with empty name', () => {
    const invalidPOI: POI = {
      id: 'poi-123',
      mapId: 'map-789',
      name: '',
      position: { lat: 40.7128, lng: -74.0060 },
      createdBy: 'user-456',
      maxParticipants: 10,
      participantCount: 0,
      createdAt: new Date()
    }
    
    const result = validatePOI(invalidPOI)
    
    expect(result.isValid).toBe(false)
    expect(result.errors).toContain('POI name is required')
  })

  it('should reject POI with name too long', () => {
    const longName = 'A'.repeat(256) // 256 characters
    const invalidPOI: POI = {
      id: 'poi-123',
      mapId: 'map-789',
      name: longName,
      position: { lat: 40.7128, lng: -74.0060 },
      createdBy: 'user-456',
      maxParticipants: 10,
      participantCount: 0,
      createdAt: new Date()
    }
    
    const result = validatePOI(invalidPOI)
    
    expect(result.isValid).toBe(false)
    expect(result.errors).toContain('POI name must be 255 characters or less')
  })

  it('should reject POI with invalid max participants', () => {
    const invalidPOI: POI = {
      id: 'poi-123',
      mapId: 'map-789',
      name: 'Meeting Room A',
      position: { lat: 40.7128, lng: -74.0060 },
      createdBy: 'user-456',
      maxParticipants: 0,
      participantCount: 0,
      createdAt: new Date()
    }
    
    const result = validatePOI(invalidPOI)
    
    expect(result.isValid).toBe(false)
    expect(result.errors).toContain('Max participants must be between 1 and 50')
  })
})

describe('Map Validation', () => {
  it('should validate correct map', () => {
    const validMap: Map = {
      id: 'map-123',
      name: 'Workshop Map 1',
      description: 'A map for the morning workshop',
      createdBy: 'facilitator-456',
      isActive: true,
      createdAt: new Date(),
      updatedAt: new Date()
    }
    
    const result = validateMap(validMap)
    
    expect(result.isValid).toBe(true)
    expect(result.errors).toEqual([])
  })

  it('should reject map with empty name', () => {
    const invalidMap: Map = {
      id: 'map-123',
      name: '',
      createdBy: 'facilitator-456',
      isActive: true,
      createdAt: new Date(),
      updatedAt: new Date()
    }
    
    const result = validateMap(invalidMap)
    
    expect(result.isValid).toBe(false)
    expect(result.errors).toContain('Map name is required')
  })

  it('should reject map with name too long', () => {
    const longName = 'A'.repeat(256) // 256 characters
    const invalidMap: Map = {
      id: 'map-123',
      name: longName,
      createdBy: 'facilitator-456',
      isActive: true,
      createdAt: new Date(),
      updatedAt: new Date()
    }
    
    const result = validateMap(invalidMap)
    
    expect(result.isValid).toBe(false)
    expect(result.errors).toContain('Map name must be 255 characters or less')
  })
})

describe('API Data Transformation', () => {
  it('should transform session from API format', () => {
    const apiSession = {
      id: 'session-123',
      userId: 'user-456',
      mapId: 'map-789',
      avatarPosition: { lat: 40.7128, lng: -74.0060 },
      createdAt: '2024-01-15T10:30:00Z',
      lastActive: '2024-01-15T10:35:00Z',
      isActive: true
    }
    
    const transformed = transformSessionFromAPI(apiSession)
    
    expect(transformed.id).toBe(apiSession.id)
    expect(transformed.userId).toBe(apiSession.userId)
    expect(transformed.mapId).toBe(apiSession.mapId)
    expect(transformed.avatarPosition).toEqual(apiSession.avatarPosition)
    expect(transformed.createdAt).toBeInstanceOf(Date)
    expect(transformed.lastActive).toBeInstanceOf(Date)
    expect(transformed.isActive).toBe(apiSession.isActive)
  })

  it('should transform POI from API format', () => {
    const apiPOI = {
      id: 'poi-123',
      mapId: 'map-789',
      name: 'Meeting Room A',
      description: 'A great place to meet',
      position: { lat: 40.7128, lng: -74.0060 },
      createdBy: 'user-456',
      maxParticipants: 10,
      participantCount: 0,
      createdAt: '2024-01-15T10:30:00Z'
    }
    
    const transformed = transformPOIFromAPI(apiPOI)
    
    expect(transformed.id).toBe(apiPOI.id)
    expect(transformed.mapId).toBe(apiPOI.mapId)
    expect(transformed.name).toBe(apiPOI.name)
    expect(transformed.description).toBe(apiPOI.description)
    expect(transformed.position).toEqual(apiPOI.position)
    expect(transformed.createdBy).toBe(apiPOI.createdBy)
    expect(transformed.maxParticipants).toBe(apiPOI.maxParticipants)
    expect(transformed.participantCount).toBe(apiPOI.participantCount)
    expect(transformed.createdAt).toBeInstanceOf(Date)
  })

  it('should transform map from API format', () => {
    const apiMap = {
      id: 'map-123',
      name: 'Workshop Map 1',
      description: 'A map for the morning workshop',
      createdBy: 'facilitator-456',
      isActive: true,
      createdAt: '2024-01-15T10:30:00Z',
      updatedAt: '2024-01-15T10:35:00Z'
    }
    
    const transformed = transformMapFromAPI(apiMap)
    
    expect(transformed.id).toBe(apiMap.id)
    expect(transformed.name).toBe(apiMap.name)
    expect(transformed.description).toBe(apiMap.description)
    expect(transformed.createdBy).toBe(apiMap.createdBy)
    expect(transformed.isActive).toBe(apiMap.isActive)
    expect(transformed.createdAt).toBeInstanceOf(Date)
    expect(transformed.updatedAt).toBeInstanceOf(Date)
  })
})