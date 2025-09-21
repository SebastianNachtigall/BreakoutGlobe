import React from 'react';
import { errorStore, AppError } from '../stores/errorStore';

interface NotificationCenterProps {
    maxVisible?: number;
    className?: string;
}

interface NotificationProps {
    error: AppError;
    onDismiss: (id: string) => void;
}

const Notification: React.FC<NotificationProps> = ({ error, onDismiss }) => {
    const handleDismiss = () => {
        onDismiss(error.id);
    };

    const handleRetry = () => {
        if (error.retryAction) {
            error.retryAction();
        }
        onDismiss(error.id);
    };

    const getSeverityIcon = () => {
        switch (error.severity) {
            case 'error':
                return '❌';
            case 'warning':
                return '⚠️';
            case 'info':
                return 'ℹ️';
            default:
                return '📝';
        }
    };

    return (
        <div className={`notification ${error.severity}`}>
            <div className="notification-content">
                <span className="notification-icon">{getSeverityIcon()}</span>
                <div className="notification-message">
                    <div className="message-text">{error.message}</div>
                    {error.details && (
                        <div className="message-details">{error.details}</div>
                    )}
                </div>
            </div>
            
            <div className="notification-actions">
                {error.retryable && (
                    <button 
                        onClick={handleRetry}
                        className="retry-button"
                    >
                        Retry
                    </button>
                )}
                <button 
                    onClick={handleDismiss}
                    className="dismiss-button"
                    aria-label="Dismiss notification"
                >
                    ✕
                </button>
            </div>
        </div>
    );
};

export const NotificationCenter: React.FC<NotificationCenterProps> = ({
    maxVisible = 5,
    className = ''
}) => {
    const errors = errorStore(state => state.errors);
    const removeError = errorStore(state => state.removeError);
    const clearAllErrors = errorStore(state => state.clearAllErrors);

    if (errors.length === 0) {
        return null;
    }

    // Show most recent errors first, limited by maxVisible
    const visibleErrors = errors
        .sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime())
        .slice(0, maxVisible);

    const handleDismiss = (id: string) => {
        removeError(id);
    };

    const handleClearAll = () => {
        clearAllErrors();
    };

    return (
        <div 
            className={`notification-center ${className}`}
            data-testid="notification-center"
        >
            <div className="notification-header">
                {errors.length > 1 && (
                    <button 
                        onClick={handleClearAll}
                        className="clear-all-button"
                    >
                        Clear All
                    </button>
                )}
            </div>
            
            <div className="notifications-list">
                {visibleErrors.map(error => (
                    <Notification
                        key={error.id}
                        error={error}
                        onDismiss={handleDismiss}
                    />
                ))}
            </div>
            
            {errors.length > maxVisible && (
                <div className="notification-overflow">
                    +{errors.length - maxVisible} more notifications
                </div>
            )}
        </div>
    );
};