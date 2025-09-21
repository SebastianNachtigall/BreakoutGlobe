import React from 'react';
import { useWebSocket } from '../hooks/useWebSocket';

interface ConnectionStatusProps {
    url: string;
    sessionId: string;
    compact?: boolean;
    className?: string;
}

export const ConnectionStatus: React.FC<ConnectionStatusProps> = ({
    url,
    sessionId,
    compact = false,
    className = ''
}) => {
    const {
        connectionStatus,
        isConnected,
        isConnecting,
        isReconnecting,
        queuedMessageCount,
        lastError,
        connect
    } = useWebSocket({ url, sessionId, autoConnect: true });

    const getStatusText = () => {
        switch (connectionStatus) {
            case 'connected':
                return 'Connected';
            case 'connecting':
                return 'Connecting...';
            case 'reconnecting':
                return 'Reconnecting...';
            case 'disconnected':
                return 'Disconnected';
            default:
                return 'Unknown';
        }
    };

    const getStatusIcon = () => {
        switch (connectionStatus) {
            case 'connected':
                return '🟢';
            case 'connecting':
                return '🟡';
            case 'reconnecting':
                return '🟠';
            case 'disconnected':
                return '🔴';
            default:
                return '⚪';
        }
    };

    const handleRetry = () => {
        connect();
    };

    const statusClasses = [
        'connection-status',
        connectionStatus,
        compact && 'compact',
        className
    ].filter(Boolean).join(' ');

    if (compact) {
        return (
            <div 
                className={statusClasses}
                data-testid="connection-indicator"
                title={getStatusText()}
            >
                <span className="status-icon">{getStatusIcon()}</span>
            </div>
        );
    }

    return (
        <div className={statusClasses} data-testid="connection-indicator">
            <div className="status-content">
                <span className="status-icon">{getStatusIcon()}</span>
                <span className="status-text">{getStatusText()}</span>
                
                {queuedMessageCount > 0 && (
                    <span className="queued-count">
                        ({queuedMessageCount} queued)
                    </span>
                )}
            </div>

            {lastError && (
                <div className="error-message">
                    {lastError.message}
                </div>
            )}

            {!isConnected && !isConnecting && !isReconnecting && (
                <button 
                    onClick={handleRetry}
                    className="retry-button"
                >
                    Retry
                </button>
            )}
        </div>
    );
};