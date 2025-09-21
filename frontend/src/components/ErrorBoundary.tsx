import React, { Component, ReactNode, ErrorInfo } from 'react';
import { errorStore } from '../stores/errorStore';

interface Props {
    children: ReactNode;
    fallback?: ReactNode;
    onError?: (error: Error, errorInfo: ErrorInfo) => void;
    resetKeys?: Array<string | number>;
}

interface State {
    hasError: boolean;
    error: Error | null;
    errorInfo: ErrorInfo | null;
}

export class ErrorBoundary extends Component<Props, State> {
    private resetTimeoutId: number | null = null;

    constructor(props: Props) {
        super(props);
        this.state = {
            hasError: false,
            error: null,
            errorInfo: null
        };
    }

    static getDerivedStateFromError(error: Error): Partial<State> {
        return {
            hasError: true,
            error
        };
    }

    componentDidCatch(error: Error, errorInfo: ErrorInfo) {
        this.setState({
            error,
            errorInfo
        });

        // Add error to global error store
        errorStore.getState().addError({
            id: `error-boundary-${Date.now()}`,
            message: error.message,
            type: 'unknown',
            timestamp: new Date(),
            severity: 'error',
            details: errorInfo.componentStack,
            retryable: true,
            retryAction: this.handleRetry
        });

        // Call custom error handler if provided
        if (this.props.onError) {
            this.props.onError(error, errorInfo);
        }

        console.error('ErrorBoundary caught an error:', error, errorInfo);
    }

    componentDidUpdate(prevProps: Props) {
        const { resetKeys } = this.props;
        const { hasError } = this.state;

        // Reset error state if resetKeys have changed
        if (hasError && resetKeys && prevProps.resetKeys) {
            const hasResetKeyChanged = resetKeys.some(
                (key, index) => key !== prevProps.resetKeys![index]
            );

            if (hasResetKeyChanged) {
                this.resetErrorBoundary();
            }
        }
    }

    componentWillUnmount() {
        if (this.resetTimeoutId) {
            clearTimeout(this.resetTimeoutId);
        }
    }

    handleRetry = () => {
        this.resetErrorBoundary();
    };

    resetErrorBoundary = () => {
        this.setState({
            hasError: false,
            error: null,
            errorInfo: null
        });
    };

    render() {
        const { hasError, error } = this.state;
        const { children, fallback } = this.props;

        if (hasError) {
            // Use custom fallback if provided
            if (fallback) {
                return fallback;
            }

            // Default error UI
            return (
                <div className="error-boundary">
                    <div className="error-boundary-content">
                        <h2>Something went wrong</h2>
                        <p>
                            {error?.message || 'An unexpected error occurred'}
                        </p>
                        <button
                            onClick={this.handleRetry}
                            className="error-boundary-retry-button"
                        >
                            Retry
                        </button>
                    </div>
                </div>
            );
        }

        return children;
    }
}