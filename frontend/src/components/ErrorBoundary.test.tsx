import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ErrorBoundary } from './ErrorBoundary';
import { errorStore } from '../stores/errorStore';

// Component that throws an error for testing
const ThrowError = ({ shouldThrow }: { shouldThrow: boolean }) => {
    if (shouldThrow) {
        throw new Error('Test error');
    }
    return <div>No error</div>;
};

describe('ErrorBoundary', () => {
    beforeEach(() => {
        errorStore.getState().clearAllErrors();
        vi.clearAllMocks();
    });

    it('should render children when no error occurs', () => {
        render(
            <ErrorBoundary>
                <div>Test content</div>
            </ErrorBoundary>
        );

        expect(screen.getByText('Test content')).toBeInTheDocument();
    });

    it('should catch and display error when child component throws', () => {
        // Suppress console.error for this test
        const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

        render(
            <ErrorBoundary>
                <ThrowError shouldThrow={true} />
            </ErrorBoundary>
        );

        expect(screen.getByText(/something went wrong/i)).toBeInTheDocument();
        expect(screen.getByText(/test error/i)).toBeInTheDocument();
        
        consoleSpy.mockRestore();
    });

    it('should add error to error store when error occurs', () => {
        const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

        render(
            <ErrorBoundary>
                <ThrowError shouldThrow={true} />
            </ErrorBoundary>
        );

        const errors = errorStore.getState().errors;
        expect(errors).toHaveLength(1);
        expect(errors[0].message).toContain('Test error');
        expect(errors[0].type).toBe('unknown');
        expect(errors[0].severity).toBe('error');

        consoleSpy.mockRestore();
    });

    it('should show retry button and allow retry', () => {
        const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

        const { rerender } = render(
            <ErrorBoundary>
                <ThrowError shouldThrow={true} />
            </ErrorBoundary>
        );

        expect(screen.getByText(/retry/i)).toBeInTheDocument();

        // Click retry and render with no error
        screen.getByText(/retry/i).click();
        
        rerender(
            <ErrorBoundary>
                <ThrowError shouldThrow={false} />
            </ErrorBoundary>
        );

        expect(screen.getByText('No error')).toBeInTheDocument();

        consoleSpy.mockRestore();
    });

    it('should display custom fallback when provided', () => {
        const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
        const customFallback = <div>Custom error message</div>;

        render(
            <ErrorBoundary fallback={customFallback}>
                <ThrowError shouldThrow={true} />
            </ErrorBoundary>
        );

        expect(screen.getByText('Custom error message')).toBeInTheDocument();

        consoleSpy.mockRestore();
    });

    it('should call onError callback when provided', () => {
        const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
        const onErrorSpy = vi.fn();

        render(
            <ErrorBoundary onError={onErrorSpy}>
                <ThrowError shouldThrow={true} />
            </ErrorBoundary>
        );

        expect(onErrorSpy).toHaveBeenCalledWith(
            expect.any(Error),
            expect.objectContaining({
                componentStack: expect.any(String)
            })
        );

        consoleSpy.mockRestore();
    });

    it('should reset error state when resetKeys change', () => {
        const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

        const { rerender } = render(
            <ErrorBoundary resetKeys={['key1']}>
                <ThrowError shouldThrow={true} />
            </ErrorBoundary>
        );

        expect(screen.getByText(/something went wrong/i)).toBeInTheDocument();

        // Change resetKeys to trigger reset
        rerender(
            <ErrorBoundary resetKeys={['key2']}>
                <ThrowError shouldThrow={false} />
            </ErrorBoundary>
        );

        expect(screen.getByText('No error')).toBeInTheDocument();

        consoleSpy.mockRestore();
    });
});