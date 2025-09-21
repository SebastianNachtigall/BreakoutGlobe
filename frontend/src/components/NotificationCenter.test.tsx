import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, act } from '@testing-library/react';
import { NotificationCenter } from './NotificationCenter';
import { errorStore } from '../stores/errorStore';

describe('NotificationCenter', () => {
    beforeEach(() => {
        errorStore.getState().clearAllErrors();
        vi.clearAllMocks();
    });

    it('should not render when no errors exist', () => {
        render(<NotificationCenter />);
        
        expect(screen.queryByTestId('notification-center')).not.toBeInTheDocument();
    });

    it('should render error notifications', () => {
        const error = {
            id: 'test-error',
            message: 'Test error message',
            type: 'websocket' as const,
            timestamp: new Date(),
            severity: 'error' as const
        };

        act(() => {
            errorStore.getState().addError(error);
        });

        render(<NotificationCenter />);

        expect(screen.getByTestId('notification-center')).toBeInTheDocument();
        expect(screen.getByText('Test error message')).toBeInTheDocument();
    });

    it('should render multiple notifications', () => {
        const errors = [
            {
                id: 'error-1',
                message: 'First error',
                type: 'websocket' as const,
                timestamp: new Date(),
                severity: 'error' as const
            },
            {
                id: 'error-2',
                message: 'Second error',
                type: 'api' as const,
                timestamp: new Date(),
                severity: 'warning' as const
            }
        ];

        act(() => {
            errors.forEach(error => errorStore.getState().addError(error));
        });

        render(<NotificationCenter />);

        expect(screen.getByText('First error')).toBeInTheDocument();
        expect(screen.getByText('Second error')).toBeInTheDocument();
    });

    it('should show different styles for different severity levels', () => {
        const errors = [
            {
                id: 'error-1',
                message: 'Error message',
                type: 'websocket' as const,
                timestamp: new Date(),
                severity: 'error' as const
            },
            {
                id: 'warning-1',
                message: 'Warning message',
                type: 'api' as const,
                timestamp: new Date(),
                severity: 'warning' as const
            },
            {
                id: 'info-1',
                message: 'Info message',
                type: 'validation' as const,
                timestamp: new Date(),
                severity: 'info' as const
            }
        ];

        act(() => {
            errors.forEach(error => errorStore.getState().addError(error));
        });

        render(<NotificationCenter />);

        const errorNotification = screen.getByText('Error message').closest('.notification');
        const warningNotification = screen.getByText('Warning message').closest('.notification');
        const infoNotification = screen.getByText('Info message').closest('.notification');

        expect(errorNotification).toHaveClass('error');
        expect(warningNotification).toHaveClass('warning');
        expect(infoNotification).toHaveClass('info');
    });

    it('should allow dismissing notifications', () => {
        const error = {
            id: 'dismissible-error',
            message: 'Dismissible error',
            type: 'websocket' as const,
            timestamp: new Date(),
            severity: 'error' as const
        };

        act(() => {
            errorStore.getState().addError(error);
        });

        render(<NotificationCenter />);

        expect(screen.getByText('Dismissible error')).toBeInTheDocument();

        const dismissButton = screen.getByLabelText(/dismiss/i);
        act(() => {
            dismissButton.click();
        });

        expect(screen.queryByText('Dismissible error')).not.toBeInTheDocument();
    });

    it('should show retry button for retryable errors', () => {
        const retryAction = vi.fn();
        const error = {
            id: 'retryable-error',
            message: 'Retryable error',
            type: 'websocket' as const,
            timestamp: new Date(),
            severity: 'error' as const,
            retryable: true,
            retryAction
        };

        act(() => {
            errorStore.getState().addError(error);
        });

        render(<NotificationCenter />);

        const retryButton = screen.getByRole('button', { name: /retry/i });
        expect(retryButton).toBeInTheDocument();

        act(() => {
            retryButton.click();
        });

        expect(retryAction).toHaveBeenCalled();
    });

    it('should not show retry button for non-retryable errors', () => {
        const error = {
            id: 'non-retryable-error',
            message: 'Non-retryable error',
            type: 'validation' as const,
            timestamp: new Date(),
            severity: 'error' as const,
            retryable: false
        };

        act(() => {
            errorStore.getState().addError(error);
        });

        render(<NotificationCenter />);

        expect(screen.queryByRole('button', { name: /retry/i })).not.toBeInTheDocument();
    });

    it('should limit number of visible notifications', () => {
        const errors = Array.from({ length: 10 }, (_, i) => ({
            id: `error-${i}`,
            message: `Error ${i}`,
            type: 'websocket' as const,
            timestamp: new Date(Date.now() + i * 1000), // Ensure different timestamps
            severity: 'error' as const
        }));

        act(() => {
            errors.forEach(error => errorStore.getState().addError(error));
        });

        render(<NotificationCenter maxVisible={5} />);

        // Should only show 5 notifications
        const notifications = screen.getAllByText(/Error \d/);
        expect(notifications).toHaveLength(5);

        // Should show the most recent ones (Error 9, 8, 7, 6, 5)
        expect(screen.getByText('Error 9')).toBeInTheDocument();
        expect(screen.getByText('Error 5')).toBeInTheDocument();
        expect(screen.queryByText('Error 4')).not.toBeInTheDocument();
    });

    it('should show clear all button when multiple notifications exist', () => {
        const errors = [
            {
                id: 'error-1',
                message: 'First error',
                type: 'websocket' as const,
                timestamp: new Date(),
                severity: 'error' as const
            },
            {
                id: 'error-2',
                message: 'Second error',
                type: 'api' as const,
                timestamp: new Date(),
                severity: 'warning' as const
            }
        ];

        act(() => {
            errors.forEach(error => errorStore.getState().addError(error));
        });

        render(<NotificationCenter />);

        const clearAllButton = screen.getByText(/clear all/i);
        expect(clearAllButton).toBeInTheDocument();

        act(() => {
            clearAllButton.click();
        });

        expect(screen.queryByTestId('notification-center')).not.toBeInTheDocument();
    });

    it('should auto-hide notifications after timeout', async () => {
        const error = {
            id: 'auto-hide-error',
            message: 'Auto hide error',
            type: 'websocket' as const,
            timestamp: new Date(),
            severity: 'info' as const,
            autoRemoveAfter: 100
        };

        act(() => {
            errorStore.getState().addError(error);
        });

        render(<NotificationCenter />);

        expect(screen.getByText('Auto hide error')).toBeInTheDocument();

        // Wait for auto-removal
        await act(async () => {
            await new Promise(resolve => setTimeout(resolve, 150));
        });

        expect(screen.queryByText('Auto hide error')).not.toBeInTheDocument();
    });
});