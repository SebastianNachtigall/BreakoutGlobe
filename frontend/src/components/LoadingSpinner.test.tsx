import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { LoadingSpinner } from './LoadingSpinner';

describe('LoadingSpinner', () => {
  it('should render spinner with default props', () => {
    render(<LoadingSpinner />);
    
    const spinner = screen.getByTestId('loading-spinner');
    expect(spinner).toBeInTheDocument();
    expect(spinner).toHaveClass('loading-spinner');
  });

  it('should render with custom size', () => {
    render(<LoadingSpinner size="large" />);
    
    const spinner = screen.getByTestId('loading-spinner');
    expect(spinner).toHaveClass('large');
  });

  it('should render with custom color', () => {
    render(<LoadingSpinner color="primary" />);
    
    const spinner = screen.getByTestId('loading-spinner');
    expect(spinner).toHaveClass('primary');
  });

  it('should render with loading text', () => {
    render(<LoadingSpinner text="Loading data..." />);
    
    expect(screen.getByText('Loading data...')).toBeInTheDocument();
  });

  it('should render inline variant', () => {
    render(<LoadingSpinner inline />);
    
    const spinner = screen.getByTestId('loading-spinner');
    expect(spinner).toHaveClass('inline');
  });

  it('should render overlay variant', () => {
    render(<LoadingSpinner overlay />);
    
    const spinner = screen.getByTestId('loading-spinner');
    expect(spinner).toHaveClass('overlay');
  });

  it('should apply custom className', () => {
    render(<LoadingSpinner className="custom-class" />);
    
    const spinner = screen.getByTestId('loading-spinner');
    expect(spinner).toHaveClass('custom-class');
  });

  it('should have proper accessibility attributes', () => {
    render(<LoadingSpinner text="Loading..." />);
    
    const spinner = screen.getByTestId('loading-spinner');
    expect(spinner).toHaveAttribute('role', 'status');
    expect(spinner).toHaveAttribute('aria-label', 'Loading...');
  });

  it('should have default aria-label when no text provided', () => {
    render(<LoadingSpinner />);
    
    const spinner = screen.getByTestId('loading-spinner');
    expect(spinner).toHaveAttribute('aria-label', 'Loading');
  });
});