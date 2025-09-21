export type Breakpoint = 'mobile' | 'tablet' | 'desktop';

export interface ResponsiveValue<T> {
  mobile: T;
  tablet?: T;
  desktop?: T;
}

// Breakpoint definitions (in pixels)
export const BREAKPOINTS = {
  mobile: 0,
  tablet: 768,
  desktop: 1024,
} as const;

/**
 * Get the current breakpoint based on window width
 */
export function getBreakpoint(): Breakpoint {
  const width = window.innerWidth;
  
  if (width >= BREAKPOINTS.desktop) {
    return 'desktop';
  } else if (width >= BREAKPOINTS.tablet) {
    return 'tablet';
  } else {
    return 'mobile';
  }
}

/**
 * Check if the current screen matches a specific breakpoint
 */
export function isScreenSize(breakpoint: Breakpoint): boolean {
  return getBreakpoint() === breakpoint;
}

/**
 * Get a responsive value based on current breakpoint
 */
export function getResponsiveValue<T>(
  value: T | ResponsiveValue<T>
): T {
  // If it's not a responsive value object, return as-is
  if (typeof value !== 'object' || value === null || !('mobile' in value)) {
    return value as T;
  }
  
  const breakpoint = getBreakpoint();
  const responsiveValue = value as ResponsiveValue<T>;
  
  // Return value for current breakpoint, fallback to mobile
  return responsiveValue[breakpoint] ?? responsiveValue.mobile;
}

/**
 * Create a media query string for a breakpoint
 */
export function createMediaQuery(breakpoint: Breakpoint): string {
  const minWidth = BREAKPOINTS[breakpoint];
  return `(min-width: ${minWidth}px)`;
}

/**
 * Hook-like function to get current breakpoint with updates
 * (This would typically be a React hook, but keeping it simple for now)
 */
export function useBreakpoint(): Breakpoint {
  return getBreakpoint();
}