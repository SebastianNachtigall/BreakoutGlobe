import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ConnectionStatus } from './ConnectionStatus';
import { useWebSocket } from '../hooks/useWebSocket';

// Mock the useWebSocket hook
vi.mock('../hooks/useWebSocket');

const mockUseWebSocket = vi.mocked(useWebSocket);

describe('ConnectionStatus', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('should display connected status', () => {
        mockUseWebSocket.mockReturnValue({
            connectionStatus: 'connected' as any,
            isConnected: true,
            isConnecting: false,
            isReconnecting: false,
            queuedMessageCount: 0,
            lastError: null,
            connect: vi.fn(),
            disconnect: vi.fn(),
            sendMessage: vi.fn(),
            onMessage: vi.fn(),
            onStateSync: vi.fn()
        });

        render(<ConnectionStatus url="ws://test" sessionId="test" />);

        expect(screen.getByText(/connected/i)).toBeInTheDocument();
        expect(screen.getByTestId('connection-indicator')).toHaveClass('connected');
    });

    it('should display connecting status', () => {
        mockUseWebSocket.mockReturnValue({
            connectionStatus: 'connecting' as any,
            isConnected: false,
            isConnecting: true,
            isReconnecting: false,
            queuedMessageCount: 0,
            lastError: null,
            connect: vi.fn(),
            disconnect: vi.fn(),
            sendMessage: vi.fn(),
            onMessage: vi.fn(),
            onStateSync: vi.fn()
        });

        render(<ConnectionStatus url="ws://test" sessionId="test" />);

        expect(screen.getByText(/connecting/i)).toBeInTheDocument();
        expect(screen.getByTestId('connection-indicator')).toHaveClass('connecting');
    });

    it('should display reconnecting status', () => {
        mockUseWebSocket.mockReturnValue({
            connectionStatus: 'reconnecting' as any,
            isConnected: false,
            isConnecting: false,
            isReconnecting: true,
            queuedMessageCount: 0,
            lastError: null,
            connect: vi.fn(),
            disconnect: vi.fn(),
            sendMessage: vi.fn(),
            onMessage: vi.fn(),
            onStateSync: vi.fn()
        });

        render(<ConnectionStatus url="ws://test" sessionId="test" />);

        expect(screen.getByText(/reconnecting/i)).toBeInTheDocument();
        expect(screen.getByTestId('connection-indicator')).toHaveClass('reconnecting');
    });

    it('should display disconnected status', () => {
        mockUseWebSocket.mockReturnValue({
            connectionStatus: 'disconnected' as any,
            isConnected: false,
            isConnecting: false,
            isReconnecting: false,
            queuedMessageCount: 0,
            lastError: null,
            connect: vi.fn(),
            disconnect: vi.fn(),
            sendMessage: vi.fn(),
            onMessage: vi.fn(),
            onStateSync: vi.fn()
        });

        render(<ConnectionStatus url="ws://test" sessionId="test" />);

        expect(screen.getByText(/disconnected/i)).toBeInTheDocument();
        expect(screen.getByTestId('connection-indicator')).toHaveClass('disconnected');
    });

    it('should show queued message count when messages are queued', () => {
        mockUseWebSocket.mockReturnValue({
            connectionStatus: 'disconnected' as any,
            isConnected: false,
            isConnecting: false,
            isReconnecting: false,
            queuedMessageCount: 3,
            lastError: null,
            connect: vi.fn(),
            disconnect: vi.fn(),
            sendMessage: vi.fn(),
            onMessage: vi.fn(),
            onStateSync: vi.fn()
        });

        render(<ConnectionStatus url="ws://test" sessionId="test" />);

        expect(screen.getByText(/3 queued/i)).toBeInTheDocument();
    });

    it('should not show queued message count when no messages are queued', () => {
        mockUseWebSocket.mockReturnValue({
            connectionStatus: 'connected' as any,
            isConnected: true,
            isConnecting: false,
            isReconnecting: false,
            queuedMessageCount: 0,
            lastError: null,
            connect: vi.fn(),
            disconnect: vi.fn(),
            sendMessage: vi.fn(),
            onMessage: vi.fn(),
            onStateSync: vi.fn()
        });

        render(<ConnectionStatus url="ws://test" sessionId="test" />);

        expect(screen.queryByText(/queued/i)).not.toBeInTheDocument();
    });

    it('should show error message when there is a connection error', () => {
        const mockError = {
            message: 'Connection failed',
            timestamp: new Date()
        };

        mockUseWebSocket.mockReturnValue({
            connectionStatus: 'disconnected' as any,
            isConnected: false,
            isConnecting: false,
            isReconnecting: false,
            queuedMessageCount: 0,
            lastError: mockError,
            connect: vi.fn(),
            disconnect: vi.fn(),
            sendMessage: vi.fn(),
            onMessage: vi.fn(),
            onStateSync: vi.fn()
        });

        render(<ConnectionStatus url="ws://test" sessionId="test" />);

        expect(screen.getByText(/connection failed/i)).toBeInTheDocument();
    });

    it('should show retry button when disconnected', () => {
        const mockConnect = vi.fn();

        mockUseWebSocket.mockReturnValue({
            connectionStatus: 'disconnected' as any,
            isConnected: false,
            isConnecting: false,
            isReconnecting: false,
            queuedMessageCount: 0,
            lastError: null,
            connect: mockConnect,
            disconnect: vi.fn(),
            sendMessage: vi.fn(),
            onMessage: vi.fn(),
            onStateSync: vi.fn()
        });

        render(<ConnectionStatus url="ws://test" sessionId="test" />);

        const retryButton = screen.getByText(/retry/i);
        expect(retryButton).toBeInTheDocument();

        retryButton.click();
        expect(mockConnect).toHaveBeenCalled();
    });

    it('should not show retry button when connected', () => {
        mockUseWebSocket.mockReturnValue({
            connectionStatus: 'connected' as any,
            isConnected: true,
            isConnecting: false,
            isReconnecting: false,
            queuedMessageCount: 0,
            lastError: null,
            connect: vi.fn(),
            disconnect: vi.fn(),
            sendMessage: vi.fn(),
            onMessage: vi.fn(),
            onStateSync: vi.fn()
        });

        render(<ConnectionStatus url="ws://test" sessionId="test" />);

        expect(screen.queryByText(/retry/i)).not.toBeInTheDocument();
    });

    it('should show compact version when compact prop is true', () => {
        mockUseWebSocket.mockReturnValue({
            connectionStatus: 'connected' as any,
            isConnected: true,
            isConnecting: false,
            isReconnecting: false,
            queuedMessageCount: 0,
            lastError: null,
            connect: vi.fn(),
            disconnect: vi.fn(),
            sendMessage: vi.fn(),
            onMessage: vi.fn(),
            onStateSync: vi.fn()
        });

        render(<ConnectionStatus url="ws://test" sessionId="test" compact />);

        const indicator = screen.getByTestId('connection-indicator');
        expect(indicator).toHaveClass('compact');
        // In compact mode, should only show status icon, not text
        expect(screen.queryByText(/connected/i)).not.toBeInTheDocument();
    });
});