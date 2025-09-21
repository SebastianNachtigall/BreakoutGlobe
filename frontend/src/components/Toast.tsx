import React, { useEffect } from 'react';

export interface ToastAction {
  label: string;
  onClick: () => void;
}

export interface ToastData {
  id: string;
  message: string;
  type: 'success' | 'error' | 'warning' | 'info';
  duration?: number;
  action?: ToastAction;
}

export interface ToastProps extends ToastData {
  onDismiss: (id: string) => void;
}

export interface ToastContainerProps {
  toasts: ToastData[];
  onDismiss: (id: string) => void;
  position?: 'top-right' | 'top-left' | 'bottom-right' | 'bottom-left' | 'top-center' | 'bottom-center';
  maxVisible?: number;
}

export const Toast: React.FC<ToastProps> = ({
  id,
  message,
  type,
  duration = 3000,
  action,
  onDismiss
}) => {
  useEffect(() => {
    if (duration > 0) {
      const timer = setTimeout(() => {
        onDismiss(id);
      }, duration);

      return () => clearTimeout(timer);
    }
  }, [id, duration, onDismiss]);

  const getIcon = () => {
    switch (type) {
      case 'success':
        return '✅';
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

  const handleDismiss = () => {
    onDismiss(id);
  };

  const handleAction = () => {
    if (action) {
      action.onClick();
    }
  };

  return (
    <div
      className={`toast ${type}`}
      data-testid="toast"
    >
      <div className="toast-content">
        <span className="toast-icon">{getIcon()}</span>
        <span className="toast-message">{message}</span>
      </div>
      
      <div className="toast-actions">
        {action && (
          <button
            onClick={handleAction}
            className="toast-action-button"
          >
            {action.label}
          </button>
        )}
        <button
          onClick={handleDismiss}
          className="toast-dismiss-button"
          aria-label="Dismiss toast"
        >
          ✕
        </button>
      </div>
    </div>
  );
};

export const ToastContainer: React.FC<ToastContainerProps> = ({
  toasts,
  onDismiss,
  position = 'top-right',
  maxVisible = 5
}) => {
  // Show most recent toasts first, limited by maxVisible
  const visibleToasts = toasts
    .slice(-maxVisible)
    .reverse();

  return (
    <div
      className={`toast-container ${position}`}
      data-testid="toast-container"
    >
      {visibleToasts.map(toast => (
        <Toast
          key={toast.id}
          {...toast}
          onDismiss={onDismiss}
        />
      ))}
    </div>
  );
};