import { describe, it, expect } from 'vitest'
import { calculateDistance, isWithinRadius, formatLatLng } from './models'
import type { LatLng } from './models'

describe('Coordinate Utilities', () => {
  describe('calculateDistance', () => {
    it('should return 0 for same point', () => {
      const point: LatLng = { lat: 40.7128, lng: -74.0060 }
      
      const distance = calculateDistance(point, point)
      
      expect(distance).toBe(0)
    })

    it('should calculate distance between NYC and LA', () => {
      const nyc: LatLng = { lat: 40.7128, lng: -74.0060 }
      const la: LatLng = { lat: 34.0522, lng: -118.2437 }
      
      const distance = calculateDistance(nyc, la)
      
      // Distance between NYC and LA is approximately 3944 km
      expect(distance).toBeGreaterThan(3900)
      expect(distance).toBeLessThan(4000)
    })

    it('should calculate distance between close points', () => {
      const point1: LatLng = { lat: 40.7128, lng: -74.0060 }
      const point2: LatLng = { lat: 40.7130, lng: -74.0062 }
      
      const distance = calculateDistance(point1, point2)
      
      // Should be a very small distance (less than 1 km)
      expect(distance).toBeLessThan(1)
      expect(distance).toBeGreaterThan(0)
    })
  })

  describe('isWithinRadius', () => {
    it('should return true for same point', () => {
      const point: LatLng = { lat: 40.7128, lng: -74.0060 }
      
      const result = isWithinRadius(point, point, 1)
      
      expect(result).toBe(true)
    })

    it('should return true for point within radius', () => {
      const center: LatLng = { lat: 40.7128, lng: -74.0060 }
      const nearby: LatLng = { lat: 40.7130, lng: -74.0062 }
      
      const result = isWithinRadius(center, nearby, 1) // 1 km radius
      
      expect(result).toBe(true)
    })

    it('should return false for point outside radius', () => {
      const nyc: LatLng = { lat: 40.7128, lng: -74.0060 }
      const la: LatLng = { lat: 34.0522, lng: -118.2437 }
      
      const result = isWithinRadius(nyc, la, 1000) // 1000 km radius
      
      expect(result).toBe(false)
    })

    it('should handle edge case at exact radius boundary', () => {
      const center: LatLng = { lat: 0, lng: 0 }
      const point: LatLng = { lat: 0, lng: 1 } // Approximately 111 km at equator
      
      const result = isWithinRadius(center, point, 111.2) // Slightly larger than distance
      
      expect(result).toBe(true)
    })
  })

  describe('formatLatLng', () => {
    it('should format coordinates with 4 decimal places', () => {
      const coords: LatLng = { lat: 40.7128, lng: -74.0060 }
      
      const formatted = formatLatLng(coords)
      
      expect(formatted).toBe('40.7128,-74.0060')
    })

    it('should format coordinates with rounding', () => {
      const coords: LatLng = { lat: 40.712845, lng: -74.006012 }
      
      const formatted = formatLatLng(coords)
      
      expect(formatted).toBe('40.7128,-74.0060')
    })

    it('should handle negative coordinates', () => {
      const coords: LatLng = { lat: -40.7128, lng: 74.0060 }
      
      const formatted = formatLatLng(coords)
      
      expect(formatted).toBe('-40.7128,74.0060')
    })

    it('should handle zero coordinates', () => {
      const coords: LatLng = { lat: 0, lng: 0 }
      
      const formatted = formatLatLng(coords)
      
      expect(formatted).toBe('0.0000,0.0000')
    })
  })
})