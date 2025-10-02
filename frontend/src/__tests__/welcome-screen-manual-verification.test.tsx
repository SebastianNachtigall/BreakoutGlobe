import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import WelcomeScreen from '../components/WelcomeScreen';

describe('Welcome Screen Manual Verification', () => {
  it('renders the welcome screen with correct content and styling', () => {
    const mockOnGetStarted = vi.fn();
    
    render(
      <WelcomeScreen
        isOpen={true}
        onGetStarted={mockOnGetStarted}
      />
    );

    // Verify main title with responsive sizing
    const welcomeTitle = screen.getByText('Welcome');
    expect(welcomeTitle).toBeInTheDocument();
    expect(welcomeTitle).toHaveClass('text-4xl', 'sm:text-5xl', 'md:text-6xl', 'font-bold', 'text-gray-900');

    // Verify subtitle with responsive sizing
    const subtitle = screen.getByText('Join POIs on the map to initiate video calls.');
    expect(subtitle).toBeInTheDocument();
    expect(subtitle).toHaveClass('text-xl', 'sm:text-2xl', 'font-bold', 'text-gray-900');

    // Verify description with responsive sizing
    const description = screen.getByText('Useful for user-driven breakout sessions in a workshop scenario.');
    expect(description).toBeInTheDocument();
    expect(description).toHaveClass('text-base', 'sm:text-lg', 'text-gray-600');

    // Verify Get Started button with mobile optimizations
    const getStartedButton = screen.getByRole('button', { name: 'Get Started' });
    expect(getStartedButton).toBeInTheDocument();
    expect(getStartedButton).toHaveClass('w-full', 'bg-blue-600', 'text-white', 'touch-manipulation');

    // Verify BreakoutGlobe SVG image exists
    const mapImage = screen.getByAltText('BreakoutGlobe map illustration showing POIs and video call functionality');
    expect(mapImage).toBeInTheDocument();
    expect(mapImage).toHaveAttribute('src', '/BreakoutGlobe.svg');
  });

  it('handles Get Started button click correctly', () => {
    const mockOnGetStarted = vi.fn();
    
    render(
      <WelcomeScreen
        isOpen={true}
        onGetStarted={mockOnGetStarted}
      />
    );

    const getStartedButton = screen.getByRole('button', { name: 'Get Started' });
    fireEvent.click(getStartedButton);

    expect(mockOnGetStarted).toHaveBeenCalledTimes(1);
  });

  it('has proper full-screen overlay styling with mobile support', () => {
    const mockOnGetStarted = vi.fn();
    
    render(
      <WelcomeScreen
        isOpen={true}
        onGetStarted={mockOnGetStarted}
      />
    );

    // Find the main container (should be the fixed overlay with scrollable overflow)
    const mainContainer = screen.getByText('Welcome').closest('.fixed');
    expect(mainContainer).toHaveClass(
      'fixed',
      'inset-0',
      'bg-gray-50',
      'z-[9999]',
      'overflow-y-auto'
    );
  });

  it('renders BreakoutGlobe SVG with proper styling', () => {
    const mockOnGetStarted = vi.fn();
    
    render(
      <WelcomeScreen
        isOpen={true}
        onGetStarted={mockOnGetStarted}
      />
    );

    // Check for the BreakoutGlobe SVG image
    const mapImage = screen.getByAltText('BreakoutGlobe map illustration showing POIs and video call functionality');
    expect(mapImage).toBeInTheDocument();
    
    // Check image styling with mobile responsiveness
    expect(mapImage).toHaveClass('w-full', 'h-auto', 'max-w-xs', 'sm:max-w-sm', 'mx-auto');
    
    // Check container styling with responsive borders
    const mapContainer = mapImage.closest('div');
    expect(mapContainer).toHaveClass('relative', 'rounded-xl', 'sm:rounded-2xl', 'bg-white', 'border', 'border-gray-200');
  });

  it('renders BreakoutGlobe SVG with proper attributes', () => {
    const mockOnGetStarted = vi.fn();
    
    render(
      <WelcomeScreen
        isOpen={true}
        onGetStarted={mockOnGetStarted}
      />
    );

    // Check for the BreakoutGlobe SVG image with proper attributes
    const mapImage = screen.getByAltText('BreakoutGlobe map illustration showing POIs and video call functionality');
    expect(mapImage).toBeInTheDocument();
    expect(mapImage.tagName).toBe('IMG');
    expect(mapImage).toHaveAttribute('src', '/BreakoutGlobe.svg');
    expect(mapImage).toHaveAttribute('alt', 'BreakoutGlobe map illustration showing POIs and video call functionality');
  });

  it('has mobile-optimized spacing and padding', () => {
    const mockOnGetStarted = vi.fn();
    
    render(
      <WelcomeScreen
        isOpen={true}
        onGetStarted={mockOnGetStarted}
      />
    );

    // Check mobile-optimized spacing for title
    const title = screen.getByText('Welcome');
    expect(title).toHaveClass('mb-6', 'sm:mb-8');

    // Check mobile-optimized padding for description container
    const descriptionContainer = screen.getByText('Join POIs on the map to initiate video calls.').closest('div');
    expect(descriptionContainer).toHaveClass('mb-6', 'sm:mb-8', 'px-2', 'sm:px-0');

    // Check mobile-optimized button sizing
    const button = screen.getByRole('button', { name: 'Get Started' });
    expect(button).toHaveClass('py-3', 'sm:py-4', 'px-6', 'sm:px-8', 'text-base', 'sm:text-lg');
  });
});