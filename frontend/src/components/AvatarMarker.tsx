import React from 'react';
import { AvatarData } from './MapContainer';

export interface AvatarMarkerProps {
  avatar: AvatarData;
  onAvatarClick?: (userId: string, clickPosition: { x: number; y: number }) => void;
}

// Utility function to generate initials from display name or fallback to sessionId
const generateInitials = (displayName?: string, sessionId?: string): string => {
  const name = displayName || sessionId || 'U';
  const words = name.trim().split(/\s+/);
  if (words.length >= 2) {
    return (words[0][0] + words[1][0]).toUpperCase();
  } else if (words.length === 1 && words[0].length >= 2) {
    return words[0].substring(0, 2).toUpperCase();
  } else {
    return words[0][0].toUpperCase();
  }
};

export const AvatarMarker: React.FC<AvatarMarkerProps> = ({ avatar, onAvatarClick }) => {
  // Determine role-based styling
  const getRoleRing = (role?: string) => {
    switch (role) {
      case 'admin':
        return 'ring-yellow-400';
      case 'superadmin':
        return 'ring-red-400';
      default:
        return avatar.isCurrentUser ? 'ring-blue-500' : 'ring-gray-400';
    }
  };

  const handleClick = (event: React.MouseEvent) => {
    event.stopPropagation();
    if (onAvatarClick && avatar.userId) {
      const clickPosition = {
        x: event.clientX,
        y: event.clientY
      };
      onAvatarClick(avatar.userId, clickPosition);
    }
  };

  const initials = generateInitials(avatar.displayName, avatar.sessionId);

  return (
    <div
      data-testid="avatar-marker"
      className={`
        w-8 h-8 rounded-full border-2 
        ${avatar.isCurrentUser
          ? 'bg-blue-500 border-blue-600'
          : 'bg-gray-500 border-gray-600'
        }
        ring-2 ${getRoleRing(avatar.role)} ring-opacity-50
        shadow-lg cursor-pointer hover:scale-110
        flex items-center justify-center text-white text-xs font-bold
        relative overflow-hidden
      `}
      onClick={handleClick}
      title={avatar.displayName || avatar.sessionId}
    >
      {avatar.avatarURL ? (
        <img 
          src={avatar.avatarURL} 
          alt={avatar.displayName || avatar.sessionId}
          className="w-full h-full object-cover rounded-full"
          onError={(e) => {
            // Fallback to initials if image fails to load
            const target = e.target as HTMLImageElement;
            target.style.display = 'none';
            // The initials will show as fallback content
          }}
        />
      ) : null}
      
      {/* Initials - shown as fallback or when no avatar image */}
      <span className={avatar.avatarURL ? 'absolute inset-0 flex items-center justify-center' : ''}>
        {initials}
      </span>
    </div>
  );
};

// Utility function to create a DOM element using the same logic as the React component
export const createAvatarMarkerElement = (
  avatar: AvatarData, 
  onAvatarClick?: (userId: string, clickPosition: { x: number; y: number }) => void
): HTMLElement => {
  const markerElement = document.createElement('div');
  
  // Determine role-based styling
  const getRoleRing = (role?: string) => {
    switch (role) {
      case 'admin':
        return 'ring-yellow-400';
      case 'superadmin':
        return 'ring-red-400';
      default:
        return avatar.isCurrentUser ? 'ring-blue-500' : 'ring-gray-400';
    }
  };

  const baseClasses = `
    w-8 h-8 rounded-full border-2 
    ${avatar.isCurrentUser
      ? 'bg-blue-500 border-blue-600'
      : 'bg-gray-500 border-gray-600'
    }
    ring-2 ${getRoleRing(avatar.role)} ring-opacity-50
    shadow-lg cursor-pointer hover:scale-110
    flex items-center justify-center text-white text-xs font-bold
    relative overflow-hidden
  `;

  markerElement.className = baseClasses;
  markerElement.title = avatar.displayName || avatar.sessionId;
  markerElement.setAttribute('data-testid', 'avatar-marker');

  // Handle avatar image or initials
  if (avatar.avatarURL) {
    // Show loading state initially
    markerElement.classList.add('animate-pulse');
    
    const avatarImg = document.createElement('img');
    avatarImg.src = avatar.avatarURL;
    avatarImg.className = 'w-full h-full object-cover rounded-full';
    avatarImg.alt = avatar.displayName || avatar.sessionId;
    
    avatarImg.onload = () => {
      markerElement.classList.remove('animate-pulse');
      markerElement.textContent = ''; // Clear any existing content
      markerElement.appendChild(avatarImg);
    };
    
    avatarImg.onerror = () => {
      markerElement.classList.remove('animate-pulse');
      const initials = generateInitials(avatar.displayName, avatar.sessionId);
      markerElement.textContent = initials;
    };
    
    // Set initial fallback while loading
    const initialInitials = generateInitials(avatar.displayName, avatar.sessionId);
    markerElement.textContent = initialInitials;
  } else {
    // Display initials
    const initials = generateInitials(avatar.displayName, avatar.sessionId);
    markerElement.textContent = initials;
  }

  // Add click handler for profile card
  markerElement.addEventListener('click', (event) => {
    event.stopPropagation();
    if (onAvatarClick && avatar.userId) {
      const clickPosition = {
        x: event.clientX,
        y: event.clientY
      };
      onAvatarClick(avatar.userId, clickPosition);
    }
  });

  // Performance optimizations - NO CSS transitions for position
  markerElement.style.willChange = 'transform';
  markerElement.style.backfaceVisibility = 'hidden';
  markerElement.style.transform = 'translateZ(0)';
  markerElement.style.contain = 'layout style paint';
  markerElement.style.pointerEvents = 'auto';
  // CRITICAL: Only allow hover transitions, no position transitions
  markerElement.style.transition = 'transform 0.2s ease'; // Only for hover scale

  // CRITICAL FIX: Add explicit positioning for MapLibre markers
  markerElement.style.position = 'absolute';
  markerElement.style.zIndex = '1000';
  
  // Add explicit dimensions and styling as fallback for Tailwind classes
  markerElement.style.width = '32px';  // w-8
  markerElement.style.height = '32px'; // h-8
  markerElement.style.borderRadius = '50%'; // rounded-full

  return markerElement;
};