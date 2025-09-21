import React from 'react';

export interface LoadingSpinnerProps {
  size?: 'small' | 'medium' | 'large';
  color?: 'primary' | 'secondary' | 'white';
  text?: string;
  inline?: boolean;
  overlay?: boolean;
  className?: string;
}

export const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({
  size = 'medium',
  color = 'primary',
  text,
  inline = false,
  overlay = false,
  className = ''
}) => {
  const spinnerClasses = [
    'loading-spinner',
    size,
    color,
    inline && 'inline',
    overlay && 'overlay',
    className
  ].filter(Boolean).join(' ');

  const ariaLabel = text || 'Loading';

  return (
    <div
      className={spinnerClasses}
      data-testid="loading-spinner"
      role="status"
      aria-label={ariaLabel}
    >
      <div className="spinner-circle">
        <div className="spinner-inner"></div>
      </div>
      {text && (
        <div className="spinner-text">{text}</div>
      )}
    </div>
  );
};