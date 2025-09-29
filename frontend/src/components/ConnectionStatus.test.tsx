import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ConnectionStatus } from './ConnectionStatus';
import { ConnectionStatus as WSConnectionStatus } from '../services/websocket-client';

describe('ConnectionStatus', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('should display connected status', () => {
        render(<ConnectionStatus status={WSConnectionStatus.CONNECTED} sessionId="test" />);

        expect(screen.getByText(/connected/i)).toBeInTheDocument();
        expect(screen.getByTestId('connection-indicator')).toHaveClass('connected');
    });

    it('should display connecting status', () => {
        render(<ConnectionStatus status={WSConnectionStatus.CONNECTING} sessionId="test" />);

        expect(screen.getByText(/connecting/i)).toBeInTheDocument();
        expect(screen.getByTestId('connection-indicator')).toHaveClass('connecting');
    });

    it('should display reconnecting status', () => {
        render(<ConnectionStatus status={WSConnectionStatus.RECONNECTING} sessionId="test" />);

        expect(screen.getByText(/reconnecting/i)).toBeInTheDocument();
        expect(screen.getByTestId('connection-indicator')).toHaveClass('reconnecting');
    });

    it('should display disconnected status', () => {
        render(<ConnectionStatus status={WSConnectionStatus.DISCONNECTED} sessionId="test" />);

        expect(screen.getByText(/disconnected/i)).toBeInTheDocument();
        expect(screen.getByTestId('connection-indicator')).toHaveClass('disconnected');
    });

    it('should show session ID when provided', () => {
        render(<ConnectionStatus status={WSConnectionStatus.CONNECTED} sessionId="test-session-123" />);

        expect(screen.getByText(/connected/i)).toBeInTheDocument();
        expect(screen.getByText(/(123)/)).toBeInTheDocument(); // Shows last 8 characters
    });

    it('should not show session ID when not provided', () => {
        render(<ConnectionStatus status={WSConnectionStatus.CONNECTED} sessionId={null} />);

        expect(screen.getByText(/connected/i)).toBeInTheDocument();
        expect(screen.queryByText(/\(/)).not.toBeInTheDocument(); // No session ID parentheses
    });

    it('should show correct icon for each status', () => {
        const { rerender } = render(<ConnectionStatus status={WSConnectionStatus.CONNECTED} sessionId="test" />);
        expect(screen.getByText('🟢')).toBeInTheDocument();

        rerender(<ConnectionStatus status={WSConnectionStatus.CONNECTING} sessionId="test" />);
        expect(screen.getByText('🟡')).toBeInTheDocument();

        rerender(<ConnectionStatus status={WSConnectionStatus.RECONNECTING} sessionId="test" />);
        expect(screen.getByText('🟠')).toBeInTheDocument();

        rerender(<ConnectionStatus status={WSConnectionStatus.DISCONNECTED} sessionId="test" />);
        expect(screen.getByText('🔴')).toBeInTheDocument();
    });

    it('should show compact version when compact prop is true', () => {
        render(<ConnectionStatus status={WSConnectionStatus.CONNECTED} sessionId="test" compact />);

        const indicator = screen.getByTestId('connection-indicator');
        expect(indicator).toHaveClass('compact');
        
        // In compact mode, should only show icon, not text
        expect(screen.getByText('🟢')).toBeInTheDocument();
        expect(screen.queryByText(/connected/i)).not.toBeInTheDocument();
    });

    it('should apply custom className when provided', () => {
        render(<ConnectionStatus status={WSConnectionStatus.CONNECTED} sessionId="test" className="custom-class" />);

        const indicator = screen.getByTestId('connection-indicator');
        expect(indicator).toHaveClass('custom-class');
    });

    it('should show unknown status for invalid status', () => {
        render(<ConnectionStatus status={'invalid' as any} sessionId="test" />);

        expect(screen.getByText(/unknown/i)).toBeInTheDocument();
        expect(screen.getByText('⚪')).toBeInTheDocument();
    });
});