import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import WelcomeScreen from '../WelcomeScreen';

describe('WelcomeScreen', () => {
  it('renders welcome screen when open', () => {
    const mockOnGetStarted = vi.fn();
    
    render(
      <WelcomeScreen
        isOpen={true}
        onGetStarted={mockOnGetStarted}
      />
    );

    expect(screen.getByText('Welcome')).toBeInTheDocument();
    expect(screen.getByText('Join POIs on the map to initiate video calls.')).toBeInTheDocument();
    expect(screen.getByText('Useful for user-driven breakout sessions in a workshop scenario.')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Get Started' })).toBeInTheDocument();
  });

  it('does not render when closed', () => {
    const mockOnGetStarted = vi.fn();
    
    render(
      <WelcomeScreen
        isOpen={false}
        onGetStarted={mockOnGetStarted}
      />
    );

    expect(screen.queryByText('Welcome')).not.toBeInTheDocument();
  });

  it('calls onGetStarted when Get Started button is clicked', () => {
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

  it('renders map illustration with BreakoutGlobe SVG', () => {
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
    expect(mapImage).toHaveAttribute('src', '/BreakoutGlobe2.svg');
  });

  it('has proper styling and layout', () => {
    const mockOnGetStarted = vi.fn();
    
    render(
      <WelcomeScreen
        isOpen={true}
        onGetStarted={mockOnGetStarted}
      />
    );

    // Check main container has proper z-index and positioning with scrollable overflow
    const mainContainer = screen.getByText('Welcome').closest('.fixed');
    expect(mainContainer).toHaveClass('fixed', 'inset-0', 'bg-gray-50', 'z-[9999]', 'overflow-y-auto');

    // Check button styling with mobile optimizations
    const button = screen.getByRole('button', { name: 'Get Started' });
    expect(button).toHaveClass('w-full', 'bg-blue-600', 'text-white', 'touch-manipulation');

    // Check map container styling with responsive borders
    const mapContainer = screen.getByAltText('BreakoutGlobe map illustration showing POIs and video call functionality').closest('div');
    expect(mapContainer).toHaveClass('relative', 'rounded-xl', 'sm:rounded-2xl', 'border', 'border-gray-200', 'bg-white');
  });

  it('has responsive text sizing for mobile devices', () => {
    const mockOnGetStarted = vi.fn();
    
    render(
      <WelcomeScreen
        isOpen={true}
        onGetStarted={mockOnGetStarted}
      />
    );

    // Check responsive title sizing
    const title = screen.getByText('Welcome');
    expect(title).toHaveClass('text-4xl', 'sm:text-5xl', 'md:text-6xl');

    // Check responsive subtitle sizing
    const subtitle = screen.getByText('Join POIs on the map to initiate video calls.');
    expect(subtitle).toHaveClass('text-xl', 'sm:text-2xl');

    // Check responsive description sizing
    const description = screen.getByText('Useful for user-driven breakout sessions in a workshop scenario.');
    expect(description).toHaveClass('text-base', 'sm:text-lg');
  });
});