import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Toast, ToastContainer } from './Toast';

describe('Toast', () => {
  const mockOnDismiss = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render toast with message', () => {
    render(
      <Toast
        id="test-toast"
        message="Success message"
        type="success"
        onDismiss={mockOnDismiss}
      />
    );

    expect(screen.getByText('Success message')).toBeInTheDocument();
    expect(screen.getByTestId('toast')).toHaveClass('success');
  });

  it('should render different toast types', () => {
    const { rerender } = render(
      <Toast
        id="test-toast"
        message="Success"
        type="success"
        onDismiss={mockOnDismiss}
      />
    );

    expect(screen.getByTestId('toast')).toHaveClass('success');

    rerender(
      <Toast
        id="test-toast"
        message="Error"
        type="error"
        onDismiss={mockOnDismiss}
      />
    );

    expect(screen.getByTestId('toast')).toHaveClass('error');

    rerender(
      <Toast
        id="test-toast"
        message="Warning"
        type="warning"
        onDismiss={mockOnDismiss}
      />
    );

    expect(screen.getByTestId('toast')).toHaveClass('warning');

    rerender(
      <Toast
        id="test-toast"
        message="Info"
        type="info"
        onDismiss={mockOnDismiss}
      />
    );

    expect(screen.getByTestId('toast')).toHaveClass('info');
  });

  it('should show appropriate icons for different types', () => {
    const { rerender } = render(
      <Toast
        id="test-toast"
        message="Success"
        type="success"
        onDismiss={mockOnDismiss}
      />
    );

    expect(screen.getByText('✅')).toBeInTheDocument();

    rerender(
      <Toast
        id="test-toast"
        message="Error"
        type="error"
        onDismiss={mockOnDismiss}
      />
    );

    expect(screen.getByText('❌')).toBeInTheDocument();
  });

  it('should call onDismiss when close button is clicked', async () => {
    const user = userEvent.setup();
    render(
      <Toast
        id="test-toast"
        message="Test message"
        type="success"
        onDismiss={mockOnDismiss}
      />
    );

    const closeButton = screen.getByLabelText(/dismiss/i);
    await user.click(closeButton);

    expect(mockOnDismiss).toHaveBeenCalledWith('test-toast');
  });

  it('should auto-dismiss after duration', async () => {
    vi.useFakeTimers();

    render(
      <Toast
        id="test-toast"
        message="Auto dismiss"
        type="success"
        duration={1000}
        onDismiss={mockOnDismiss}
      />
    );

    expect(mockOnDismiss).not.toHaveBeenCalled();

    act(() => {
      vi.advanceTimersByTime(1000);
    });

    expect(mockOnDismiss).toHaveBeenCalledWith('test-toast');

    vi.useRealTimers();
  });

  it('should not auto-dismiss when duration is 0', async () => {
    vi.useFakeTimers();

    render(
      <Toast
        id="test-toast"
        message="No auto dismiss"
        type="success"
        duration={0}
        onDismiss={mockOnDismiss}
      />
    );

    act(() => {
      vi.advanceTimersByTime(5000);
    });

    expect(mockOnDismiss).not.toHaveBeenCalled();

    vi.useRealTimers();
  });

  it('should render with action button', async () => {
    const mockAction = vi.fn();
    const user = userEvent.setup();

    render(
      <Toast
        id="test-toast"
        message="With action"
        type="info"
        action={{ label: 'Retry', onClick: mockAction }}
        onDismiss={mockOnDismiss}
      />
    );

    const actionButton = screen.getByText('Retry');
    expect(actionButton).toBeInTheDocument();

    await user.click(actionButton);
    expect(mockAction).toHaveBeenCalled();
  });
});

describe('ToastContainer', () => {
  it('should render empty container when no toasts', () => {
    render(<ToastContainer toasts={[]} onDismiss={vi.fn()} />);

    const container = screen.getByTestId('toast-container');
    expect(container).toBeInTheDocument();
    expect(container).toBeEmptyDOMElement();
  });

  it('should render multiple toasts', () => {
    const toasts = [
      {
        id: 'toast-1',
        message: 'First toast',
        type: 'success' as const,
        duration: 3000
      },
      {
        id: 'toast-2',
        message: 'Second toast',
        type: 'error' as const,
        duration: 3000
      }
    ];

    render(<ToastContainer toasts={toasts} onDismiss={vi.fn()} />);

    expect(screen.getByText('First toast')).toBeInTheDocument();
    expect(screen.getByText('Second toast')).toBeInTheDocument();
  });

  it('should limit number of visible toasts', () => {
    const toasts = Array.from({ length: 10 }, (_, i) => ({
      id: `toast-${i}`,
      message: `Toast ${i}`,
      type: 'info' as const,
      duration: 3000
    }));

    render(<ToastContainer toasts={toasts} maxVisible={5} onDismiss={vi.fn()} />);

    // Should only show 5 toasts
    const visibleToasts = screen.getAllByTestId('toast');
    expect(visibleToasts).toHaveLength(5);
  });

  it('should position toasts correctly', () => {
    const toasts = [
      {
        id: 'toast-1',
        message: 'Test toast',
        type: 'success' as const,
        duration: 3000
      }
    ];

    const { rerender } = render(
      <ToastContainer toasts={toasts} position="top-right" onDismiss={vi.fn()} />
    );

    expect(screen.getByTestId('toast-container')).toHaveClass('top-right');

    rerender(
      <ToastContainer toasts={toasts} position="bottom-left" onDismiss={vi.fn()} />
    );

    expect(screen.getByTestId('toast-container')).toHaveClass('bottom-left');
  });
});