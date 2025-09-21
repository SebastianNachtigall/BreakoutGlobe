import { describe, it, expect, beforeEach } from 'vitest';
import { errorStore } from './errorStore';

describe('errorStore', () => {
    beforeEach(() => {
        errorStore.getState().clearAllErrors();
    });

    describe('Error Management', () => {
        it('should add error to store', () => {
            const error = {
                id: 'test-error',
                message: 'Test error message',
                type: 'websocket' as const,
                timestamp: new Date(),
                severity: 'error' as const
            };

            errorStore.getState().addError(error);

            const state = errorStore.getState();
            expect(state.errors).toHaveLength(1);
            expect(state.errors[0]).toEqual(error);
        });

        it('should remove error by id', () => {
            const error1 = {
                id: 'error-1',
                message: 'First error',
                type: 'websocket' as const,
                timestamp: new Date(),
                severity: 'error' as const
            };
            const error2 = {
                id: 'error-2',
                message: 'Second error',
                type: 'api' as const,
                timestamp: new Date(),
                severity: 'warning' as const
            };

            errorStore.getState().addError(error1);
            errorStore.getState().addError(error2);
            errorStore.getState().removeError('error-1');

            const state = errorStore.getState();
            expect(state.errors).toHaveLength(1);
            expect(state.errors[0].id).toBe('error-2');
        });

        it('should clear all errors', () => {
            const error1 = {
                id: 'error-1',
                message: 'First error',
                type: 'websocket' as const,
                timestamp: new Date(),
                severity: 'error' as const
            };
            const error2 = {
                id: 'error-2',
                message: 'Second error',
                type: 'api' as const,
                timestamp: new Date(),
                severity: 'warning' as const
            };

            errorStore.getState().addError(error1);
            errorStore.getState().addError(error2);
            errorStore.getState().clearAllErrors();

            const state = errorStore.getState();
            expect(state.errors).toHaveLength(0);
        });

        it('should auto-remove errors after timeout', async () => {
            const error = {
                id: 'auto-remove-error',
                message: 'Auto remove error',
                type: 'websocket' as const,
                timestamp: new Date(),
                severity: 'info' as const,
                autoRemoveAfter: 100 // 100ms
            };

            errorStore.getState().addError(error);
            expect(errorStore.getState().errors).toHaveLength(1);

            // Wait for auto-removal
            await new Promise(resolve => setTimeout(resolve, 150));
            expect(errorStore.getState().errors).toHaveLength(0);
        });
    });

    describe('Error Filtering', () => {
        beforeEach(() => {
            const errors = [
                {
                    id: 'websocket-error',
                    message: 'WebSocket connection failed',
                    type: 'websocket' as const,
                    timestamp: new Date(),
                    severity: 'error' as const
                },
                {
                    id: 'api-error',
                    message: 'API request failed',
                    type: 'api' as const,
                    timestamp: new Date(),
                    severity: 'error' as const
                },
                {
                    id: 'validation-warning',
                    message: 'Invalid input',
                    type: 'validation' as const,
                    timestamp: new Date(),
                    severity: 'warning' as const
                }
            ];

            errors.forEach(error => errorStore.getState().addError(error));
        });

        it('should get errors by type', () => {
            const websocketErrors = errorStore.getState().getErrorsByType('websocket');
            expect(websocketErrors).toHaveLength(1);
            expect(websocketErrors[0].id).toBe('websocket-error');
        });

        it('should get errors by severity', () => {
            const errorSeverityErrors = errorStore.getState().getErrorsBySeverity('error');
            expect(errorSeverityErrors).toHaveLength(2);
            
            const warningErrors = errorStore.getState().getErrorsBySeverity('warning');
            expect(warningErrors).toHaveLength(1);
            expect(warningErrors[0].id).toBe('validation-warning');
        });

        it('should check if has errors of type', () => {
            expect(errorStore.getState().hasErrorsOfType('websocket')).toBe(true);
            expect(errorStore.getState().hasErrorsOfType('unknown' as any)).toBe(false);
        });
    });

    describe('Error Statistics', () => {
        it('should get error count', () => {
            expect(errorStore.getState().getErrorCount()).toBe(0);

            errorStore.getState().addError({
                id: 'test-1',
                message: 'Test 1',
                type: 'websocket' as const,
                timestamp: new Date(),
                severity: 'error' as const
            });

            expect(errorStore.getState().getErrorCount()).toBe(1);
        });

        it('should get latest error', () => {
            const error1 = {
                id: 'error-1',
                message: 'First error',
                type: 'websocket' as const,
                timestamp: new Date(Date.now() - 1000),
                severity: 'error' as const
            };
            const error2 = {
                id: 'error-2',
                message: 'Second error',
                type: 'api' as const,
                timestamp: new Date(),
                severity: 'warning' as const
            };

            errorStore.getState().addError(error1);
            errorStore.getState().addError(error2);

            const latestError = errorStore.getState().getLatestError();
            expect(latestError?.id).toBe('error-2');
        });
    });
});