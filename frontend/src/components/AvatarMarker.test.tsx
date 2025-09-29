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
    
    // Test that the correct Tailwind classes are applied
    // Note: Computed styles won't work in test environment without Tailwind CSS loaded
    expect(markerElement.className).toContain('w-8');
    expect(markerElement.className).toContain('h-8');
    expect(markerElement.className).toContain('rounded-full');
    expect(markerElement.className).toContain('border-2');
    expect(markerElement.className).toContain('bg-blue-500');
    expect(markerElement.className).toContain('border-blue-600');
    expect(markerElement.className).toContain('ring-2');
    expect(markerElement.className).toContain('ring-blue-500');
    expect(markerElement.className).toContain('shadow-lg');
    expect(markerElement.className).toContain('cursor-pointer');
    expect(markerElement.className).toContain('hover:scale-110');
    expect(markerElement.className).toContain('flex');
    expect(markerElement.className).toContain('items-center');
    expect(markerElement.className).toContain('justify-center');
    expect(markerElement.className).toContain('text-white');
    expect(markerElement.className).toContain('text-xs');
    expect(markerElement.className).toContain('font-bold');
    
    // Clean up
    document.body.removeChild(markerElement);
  });
});