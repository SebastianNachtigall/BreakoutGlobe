import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { getBreakpoint, isScreenSize, getResponsiveValue } from './responsive';

// Mock window.matchMedia
const mockMatchMedia = (matches: boolean) => ({
  matches,
  media: '',
  onchange: null,
  addListener: () => {},
  removeListener: () => {},
  addEventListener: () => {},
  removeEventListener: () => {},
  dispatchEvent: () => true,
});

describe('Responsive utilities', () => {
  beforeEach(() => {
    // Reset window size
    Object.defineProperty(window, 'innerWidth', {
      writable: true,
      configurable: true,
      value: 1024,
    });
    Object.defineProperty(window, 'innerHeight', {
      writable: true,
      configurable: true,
      value: 768,
    });
  });

  afterEach(() => {
    // Clean up
    vi.restoreAllMocks();
  });

  describe('getBreakpoint', () => {
    it('should return mobile for small screens', () => {
      Object.defineProperty(window, 'innerWidth', { value: 320 });
      expect(getBreakpoint()).toBe('mobile');
    });

    it('should return tablet for medium screens', () => {
      Object.defineProperty(window, 'innerWidth', { value: 768 });
      expect(getBreakpoint()).toBe('tablet');
    });

    it('should return desktop for large screens', () => {
      Object.defineProperty(window, 'innerWidth', { value: 1024 });
      expect(getBreakpoint()).toBe('desktop');
    });

    it('should return desktop for extra large screens', () => {
      Object.defineProperty(window, 'innerWidth', { value: 1920 });
      expect(getBreakpoint()).toBe('desktop');
    });
  });

  describe('isScreenSize', () => {
    it('should correctly identify mobile screens', () => {
      Object.defineProperty(window, 'innerWidth', { value: 480 });
      expect(isScreenSize('mobile')).toBe(true);
      expect(isScreenSize('tablet')).toBe(false);
      expect(isScreenSize('desktop')).toBe(false);
    });

    it('should correctly identify tablet screens', () => {
      Object.defineProperty(window, 'innerWidth', { value: 768 });
      expect(isScreenSize('mobile')).toBe(false);
      expect(isScreenSize('tablet')).toBe(true);
      expect(isScreenSize('desktop')).toBe(false);
    });

    it('should correctly identify desktop screens', () => {
      Object.defineProperty(window, 'innerWidth', { value: 1200 });
      expect(isScreenSize('mobile')).toBe(false);
      expect(isScreenSize('tablet')).toBe(false);
      expect(isScreenSize('desktop')).toBe(true);
    });
  });

  describe('getResponsiveValue', () => {
    it('should return mobile value for mobile screens', () => {
      Object.defineProperty(window, 'innerWidth', { value: 480 });
      const values = { mobile: 'small', tablet: 'medium', desktop: 'large' };
      expect(getResponsiveValue(values)).toBe('small');
    });

    it('should return tablet value for tablet screens', () => {
      Object.defineProperty(window, 'innerWidth', { value: 768 });
      const values = { mobile: 'small', tablet: 'medium', desktop: 'large' };
      expect(getResponsiveValue(values)).toBe('medium');
    });

    it('should return desktop value for desktop screens', () => {
      Object.defineProperty(window, 'innerWidth', { value: 1200 });
      const values = { mobile: 'small', tablet: 'medium', desktop: 'large' };
      expect(getResponsiveValue(values)).toBe('large');
    });

    it('should fallback to mobile value if breakpoint not defined', () => {
      Object.defineProperty(window, 'innerWidth', { value: 768 });
      const values = { mobile: 'small', desktop: 'large' };
      expect(getResponsiveValue(values)).toBe('small');
    });

    it('should handle single value', () => {
      const value = 'constant';
      expect(getResponsiveValue(value)).toBe('constant');
    });
  });
});