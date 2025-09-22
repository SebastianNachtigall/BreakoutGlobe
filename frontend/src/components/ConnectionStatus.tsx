import React from 'react';
import { ConnectionStatus as WSConnectionStatus } from '../services/websocket-client';

interface ConnectionStatusProps {
    status: WSConnectionStatus;
    sessionId: string | null;
    compact?: boolean;
    className?: string;
}

export const ConnectionStatus: React.FC<ConnectionStatusProps> = ({
    status,
    sessionId,
    compact = false,
    className = ''
}) => {
    const connectionStatus = status;
    const isConnected = status === WSConnectionStatus.CONNECTED;
    const isConnecting = status === WSConnectionStatus.CONNECTING;
    const isReconnecting = status === WSConnectionStatus.RECONNECTING;

    const getStatusText = () => {
        switch (connectionStatus) {
            case WSConnectionStatus.CONNECTED:
                return 'Connected';
            case WSConnectionStatus.CONNECTING:
                return 'Connecting...';
            case WSConnectionStatus.RECONNECTING:
                return 'Reconnecting...';
            case WSConnectionStatus.DISCONNECTED:
                return 'Disconnected';
            default:
                return 'Unknown';
        }
    };

    const getStatusIcon = () => {
        switch (connectionStatus) {
            case WSConnectionStatus.CONNECTED:
                return '🟢';
            case WSConnectionStatus.CONNECTING:
                return '🟡';
            case WSConnectionStatus.RECONNECTING:
                return '🟠';
            case WSConnectionStatus.DISCONNECTED:
                return '🔴';
            default:
                return '⚪';
        }
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
                {sessionId && (
                    <span className="session-id">
                        ({sessionId.slice(-8)})
                    </span>
                )}
            </div>
        </div>
    );
};