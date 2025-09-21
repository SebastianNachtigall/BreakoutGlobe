import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type ErrorType = 'websocket' | 'api' | 'validation' | 'network' | 'unknown';
export type ErrorSeverity = 'error' | 'warning' | 'info';

export interface AppError {
    id: string;
    message: string;
    type: ErrorType;
    timestamp: Date;
    severity: ErrorSeverity;
    details?: string;
    autoRemoveAfter?: number; // milliseconds
    retryable?: boolean;
    retryAction?: () => void;
}

interface ErrorState {
    errors: AppError[];
    
    // Actions
    addError: (error: AppError) => void;
    removeError: (id: string) => void;
    clearAllErrors: () => void;
    
    // Selectors
    getErrorsByType: (type: ErrorType) => AppError[];
    getErrorsBySeverity: (severity: ErrorSeverity) => AppError[];
    hasErrorsOfType: (type: ErrorType) => boolean;
    getErrorCount: () => number;
    getLatestError: () => AppError | null;
}

export const errorStore = create<ErrorState>()(
    persist(
        (set, get) => ({
            errors: [],

            addError: (error: AppError) => {
                set((state) => ({
                    errors: [...state.errors, error]
                }));

                // Auto-remove error if specified
                if (error.autoRemoveAfter) {
                    setTimeout(() => {
                        get().removeError(error.id);
                    }, error.autoRemoveAfter);
                }
            },

            removeError: (id: string) => {
                set((state) => ({
                    errors: state.errors.filter(error => error.id !== id)
                }));
            },

            clearAllErrors: () => {
                set({ errors: [] });
            },

            getErrorsByType: (type: ErrorType) => {
                return get().errors.filter(error => error.type === type);
            },

            getErrorsBySeverity: (severity: ErrorSeverity) => {
                return get().errors.filter(error => error.severity === severity);
            },

            hasErrorsOfType: (type: ErrorType) => {
                return get().errors.some(error => error.type === type);
            },

            getErrorCount: () => {
                return get().errors.length;
            },

            getLatestError: () => {
                const errors = get().errors;
                if (errors.length === 0) return null;
                
                return errors.reduce((latest, current) => 
                    current.timestamp > latest.timestamp ? current : latest
                );
            }
        }),
        {
            name: 'error-store',
            partialize: (state) => ({
                // Don't persist errors - they should be fresh on each session
                errors: []
            })
        }
    )
);