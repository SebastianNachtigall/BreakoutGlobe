import { describe, it, expect } from 'vitest';

describe('Avatar Marker CSS Classes', () => {
  it('should create a marker element with proper CSS classes', () => {
    // Simulate the exact code from MapContainer
    const markerElement = document.createElement('div');
    
    const baseClasses = `
      w-8 h-8 rounded-full border-2 
      bg-blue-500 border-blue-600
      ring-2 ring-blue-500 ring-opacity-50
      shadow-lg cursor-pointer hover:scale-110
      flex items-center justify-center text-white text-xs font-bold
      relative overflow-hidden
    `;

    markerElement.className = baseClasses;
    markerElement.textContent = 'AN';
    
    // Add to DOM to test visibility
    document.body.appendChild(markerElement);
    
    // Test that element exists and has content
    expect(markerElement.textContent).toBe('AN');
    expect(markerElement.className).toContain('bg-blue-500');
    expect(markerElement.className).toContain('text-white');
    expect(markerElement.className).toContain('flex');
    
    // Test computed styles (this will show if Tailwind is working)
    const computedStyle = window.getComputedStyle(markerElement);
    console.log('🎨 Computed styles:', {
      backgroundColor: computedStyle.backgroundColor,
      color: computedStyle.color,
      display: computedStyle.display,
      width: computedStyle.width,
      height: computedStyle.height,
      borderRadius: computedStyle.borderRadius,
    });
    
    // These tests will show if Tailwind CSS is working
    // If Tailwind isn't loaded, these will be default values
    expect(computedStyle.display).toBe('flex'); // Should be 'flex' if Tailwind works
    expect(computedStyle.backgroundColor).not.toBe('rgba(0, 0, 0, 0)'); // Should have blue background
    expect(computedStyle.color).not.toBe('rgb(0, 0, 0)'); // Should be white text
    
    // Clean up
    document.body.removeChild(markerElement);
  });
});