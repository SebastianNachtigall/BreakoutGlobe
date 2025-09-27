import React from 'react';
import { POIData, POIParticipant } from './MapContainer';

export interface POIMarkerProps {
  poi: POIData;
  onPOIClick?: (poiId: string) => void;
}

// Utility function to generate initials from name
const generateInitials = (name: string): string => {
  const words = name.trim().split(/\s+/);
  if (words.length >= 2) {
    return (words[0][0] + words[1][0]).toUpperCase();
  } else if (words.length === 1 && words[0].length >= 2) {
    return words[0].substring(0, 2).toUpperCase();
  } else {
    return words[0][0].toUpperCase();
  }
};

// Avatar Badge Component
interface AvatarBadgeProps {
  participant: POIParticipant;
  angle: number;
  radius: number;
}

const AvatarBadge: React.FC<AvatarBadgeProps> = ({ participant, angle, radius }) => {
  const initials = generateInitials(participant.name);
  
  return (
    <div
      data-testid="avatar-badge"
      className="absolute w-6 h-6 rounded-full border border-white shadow-md bg-gray-500 flex items-center justify-center text-white text-xs font-bold overflow-hidden"
      style={{
        transform: `rotate(${angle}deg) translateY(-${radius}px) rotate(${-angle}deg)`,
        transformOrigin: 'center center'
      }}
      title={participant.name}
    >
      {participant.avatarUrl ? (
        <img 
          src={participant.avatarUrl} 
          alt={participant.name}
          className="w-full h-full object-cover rounded-full"
          onError={(e) => {
            // Fallback to initials if image fails to load
            const target = e.target as HTMLImageElement;
            target.style.display = 'none';
          }}
        />
      ) : null}
      
      {/* Initials - shown as fallback or when no avatar image */}
      <span className={participant.avatarUrl ? 'absolute inset-0 flex items-center justify-center' : ''}>
        {initials}
      </span>
    </div>
  );
};

// Avatar Badges Container Component
interface AvatarBadgesProps {
  participants: POIParticipant[];
}

const AvatarBadges: React.FC<AvatarBadgesProps> = ({ participants }) => {
  // Debug logging
  console.log('🎯 AvatarBadges render:', { participants, length: participants?.length });
  
  if (!participants || participants.length === 0) {
    console.log('🎯 AvatarBadges: No participants, returning null');
    return null;
  }

  // Limit to maximum 8 badges to avoid overcrowding
  const displayParticipants = participants.slice(0, 8);
  const radius = 32; // Distance from POI center
  
  // Calculate angles starting at 11 o'clock (-60 degrees) and going counter-clockwise
  const startAngle = -60;
  const angleStep = 45; // 45-degree increments for up to 8 positions
  
  return (
    <div data-testid="avatar-badges-container" className="absolute inset-0 pointer-events-none">
      {displayParticipants.map((participant, index) => {
        let angle = startAngle - (angleStep * index);
        // Normalize angle to be between -360 and 0
        while (angle <= -360) {
          angle += 360;
        }
        return (
          <AvatarBadge
            key={participant.id}
            participant={participant}
            angle={angle}
            radius={radius}
          />
        );
      })}
    </div>
  );
};

export const POIMarker: React.FC<POIMarkerProps> = ({ poi, onPOIClick }) => {
  // Debug logging
  console.log('🏷️ POIMarker render:', { 
    poiId: poi.id, 
    participantCount: poi.participantCount, 
    participants: poi.participants,
    participantsLength: poi.participants?.length 
  });
  
  const isFull = poi.participantCount >= poi.maxParticipants;

  const handleClick = (event: React.MouseEvent) => {
    event.stopPropagation(); // Prevent map click handler from firing
    if (!isFull && onPOIClick) {
      onPOIClick(poi.id);
    }
  };

  // If POI has an image, render it as circular image with name underneath and red badge
  if (poi.imageUrl) {
    return (
      <div
        data-testid="poi-marker"
        className="flex flex-col items-center cursor-pointer hover:scale-105 transition-transform duration-200"
        onClick={handleClick}
        title={`${poi.name} - ${poi.participantCount}/${poi.maxParticipants} participants`}
      >
        <div className="relative p-1">
          <img
            src={poi.imageUrl}
            alt={poi.name}
            className="w-16 h-16 rounded-full object-cover border-2 border-white shadow-lg"
            onError={(e) => {
              // Fallback to default marker if image fails to load
              const target = e.target as HTMLImageElement;
              target.style.display = 'none';
              console.error(`❌ Image failed to load for POI ${poi.name}: ${poi.imageUrl}`);
            }}
            onLoad={() => {
              console.log(`✅ Image loaded successfully for POI ${poi.name}`);
            }}
          />
          {/* Red badge with participant count */}
          <div
            data-testid="participant-badge"
            className="absolute top-0 right-0 bg-red-500 text-white text-xs font-bold rounded-full w-6 h-6 flex items-center justify-center border-2 border-white shadow-md"
          >
            {poi.participantCount}
          </div>
          
          {/* Avatar badges around the POI */}
          <AvatarBadges participants={poi.participants || []} />
          
          {/* Temporary test badge to verify rendering works */}
          {poi.participantCount > 0 && (!poi.participants || poi.participants.length === 0) && (
            <div 
              className="absolute w-6 h-6 rounded-full bg-blue-500 border border-white shadow-md flex items-center justify-center text-white text-xs font-bold"
              style={{
                transform: 'rotate(-60deg) translateY(-32px) rotate(60deg)',
                transformOrigin: 'center center'
              }}
              title="Test badge - participants data missing"
            >
              ?
            </div>
          )}
        </div>
        {/* Name underneath */}
        <div className="mt-1 text-center">
          <div className="text-xs font-semibold text-gray-800 max-w-20 truncate">
            {poi.name}
          </div>
        </div>
      </div>
    );
  }

  // Default marker without image
  return (
    <div
      data-testid="poi-marker"
      className={`
        w-20 h-12 rounded-lg border-2 cursor-pointer
        flex flex-col items-center justify-center text-white text-xs font-bold
        shadow-lg hover:scale-105 transition-transform duration-200 relative
        ${isFull
          ? 'bg-red-500 border-red-600 cursor-not-allowed'
          : 'bg-green-500 border-green-600'
        }
      `}
      onClick={handleClick}
      title={`${poi.name} - ${poi.participantCount}/${poi.maxParticipants} participants`}
    >
      <div className="text-center leading-tight">
        <div className="truncate max-w-20">{poi.name}</div>
        <div className="text-xs opacity-90">
          {poi.participantCount}/{poi.maxParticipants}
        </div>
      </div>
      
      {/* Avatar badges around the POI */}
      <AvatarBadges participants={poi.participants || []} />
      
      {/* Temporary test badge to verify rendering works */}
      {poi.participantCount > 0 && (!poi.participants || poi.participants.length === 0) && (
        <div 
          className="absolute w-6 h-6 rounded-full bg-blue-500 border border-white shadow-md flex items-center justify-center text-white text-xs font-bold"
          style={{
            transform: 'rotate(-60deg) translateY(-32px) rotate(60deg)',
            transformOrigin: 'center center'
          }}
          title="Test badge - participants data missing"
        >
          ?
        </div>
      )}
    </div>
  );
};

// Helper function to create avatar badges HTML
const createAvatarBadgesHTML = (participants: POIParticipant[]): string => {
  if (!participants || participants.length === 0) {
    return '';
  }

  const displayParticipants = participants.slice(0, 8);
  const radius = 40; // Increased radius to make badges more visible
  const startAngle = -60;
  const angleStep = 45;

  return displayParticipants.map((participant, index) => {
    let angle = startAngle - (angleStep * index);
    // Normalize angle to be between -360 and 0
    while (angle <= -360) {
      angle += 360;
    }

    const initials = generateInitials(participant.name);
    const avatarImg = participant.avatarUrl 
      ? `<img src="${participant.avatarUrl}" alt="${participant.name}" class="w-full h-full object-cover rounded-full" onerror="this.style.display='none';" />`
      : '';

    return `
      <div 
        class="absolute w-8 h-8 rounded-full border-2 border-white shadow-lg bg-gray-500 flex items-center justify-center text-white text-xs font-bold overflow-hidden z-10"
        style="transform: rotate(${angle}deg) translateY(-${radius}px) rotate(${-angle}deg); transform-origin: center center; top: 50%; left: 50%; margin-left: -16px; margin-top: -16px;"
        title="${participant.name}"
      >
        ${avatarImg}
        <span class="${participant.avatarUrl ? 'absolute inset-0 flex items-center justify-center' : ''}">${initials}</span>
      </div>
    `;
  }).join('');
};

// Utility function to create a DOM element using the same logic as the React component
export const createPOIMarkerElement = (poi: POIData, onPOIClick?: (poiId: string) => void): HTMLElement => {
  console.log(`🖼️ Creating POI marker for "${poi.name}" with imageUrl:`, poi.imageUrl);

  const markerElement = document.createElement('div');
  const isFull = poi.participantCount >= poi.maxParticipants;

  // If POI has an image, create circular image-based marker
  if (poi.imageUrl) {
    markerElement.className = `
      flex flex-col items-center cursor-pointer hover:scale-105 transition-transform duration-200 relative
      ${isFull ? 'cursor-not-allowed' : ''}
    `;
    
    // Add padding to prevent badge clipping
    markerElement.style.padding = '20px';

    const avatarBadgesHTML = createAvatarBadgesHTML(poi.participants || []);
    
    markerElement.innerHTML = `
      <div class="relative p-1">
        <img 
          src="${poi.imageUrl}" 
          alt="${poi.name}"
          class="w-16 h-16 rounded-full object-cover border-2 border-white shadow-lg"
          onerror="this.style.display='none'; console.error('❌ Image failed to load for POI ${poi.name}: ${poi.imageUrl}');"
          onload="console.log('✅ Image loaded successfully for POI ${poi.name}');"
        />
        <div class="absolute top-0 right-0 bg-red-500 text-white text-xs font-bold rounded-full w-6 h-6 flex items-center justify-center border-2 border-white shadow-md">
          ${poi.participantCount}
        </div>
        
        <!-- Avatar badges around the POI -->
        ${avatarBadgesHTML}
        
        <!-- Temporary test badge if no participant data -->
        ${poi.participantCount > 0 && (!poi.participants || poi.participants.length === 0) ? `
          <div 
            class="absolute w-6 h-6 rounded-full bg-blue-500 border border-white shadow-md flex items-center justify-center text-white text-xs font-bold"
            style="transform: rotate(-60deg) translateY(-32px) rotate(60deg); transform-origin: center center;"
            title="Test badge - participants data missing"
          >
            ?
          </div>
        ` : ''}
      </div>
      <div class="mt-1 text-center">
        <div class="text-xs font-semibold text-gray-800 max-w-20 truncate">
          ${poi.name}
        </div>
      </div>
    `;
  } else {
    // Default marker without image
    markerElement.className = `
      w-20 h-12 rounded-lg border-2 cursor-pointer relative
      flex flex-col items-center justify-center text-white text-xs font-bold
      shadow-lg hover:scale-105 transition-transform duration-200
      ${isFull
        ? 'bg-red-500 border-red-600 cursor-not-allowed'
        : 'bg-green-500 border-green-600'
      }
    `;
    
    // Add padding to prevent badge clipping
    markerElement.style.padding = '20px';

    const avatarBadgesHTML = createAvatarBadgesHTML(poi.participants || []);
    
    markerElement.innerHTML = `
      <div class="text-center leading-tight">
        <div class="truncate max-w-20">${poi.name}</div>
        <div class="text-xs opacity-90">
          ${poi.participantCount}/${poi.maxParticipants}
        </div>
      </div>
      
      <!-- Avatar badges around the POI -->
      ${avatarBadgesHTML}
      
      <!-- Temporary test badge if no participant data -->
      ${poi.participantCount > 0 && (!poi.participants || poi.participants.length === 0) ? `
        <div 
          class="absolute w-6 h-6 rounded-full bg-blue-500 border border-white shadow-md flex items-center justify-center text-white text-xs font-bold"
          style="transform: rotate(-60deg) translateY(-32px) rotate(60deg); transform-origin: center center;"
          title="Test badge - participants data missing"
        >
          ?
        </div>
      ` : ''}
    `;
  }

  markerElement.title = `${poi.name} - ${poi.participantCount}/${poi.maxParticipants} participants`;
  markerElement.setAttribute('data-testid', 'poi-marker');

  // Add click handler
  markerElement.addEventListener('click', (event) => {
    event.stopPropagation(); // Prevent map click handler from firing
    if (!isFull && onPOIClick) {
      onPOIClick(poi.id);
    }
  });

  // Performance optimizations
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

  return markerElement;
};

// Utility function to update an existing POI marker element with new data
export const updatePOIMarkerElement = (element: HTMLElement, poi: POIData): void => {
  console.log(`🔄 Updating POI marker for "${poi.name}" with participant data:`, {
    participantCount: poi.participantCount,
    participants: poi.participants,
    participantsLength: poi.participants?.length
  });

  const isFull = poi.participantCount >= poi.maxParticipants;
  const avatarBadgesHTML = createAvatarBadgesHTML(poi.participants || []);

  // Update based on whether POI has an image or not
  if (poi.imageUrl) {
    // Update image-based marker
    element.className = `
      flex flex-col items-center cursor-pointer hover:scale-105 transition-transform duration-200 relative
      ${isFull ? 'cursor-not-allowed' : ''}
    `;
    
    element.style.padding = '20px';

    element.innerHTML = `
      <div class="relative p-1">
        <img 
          src="${poi.imageUrl}" 
          alt="${poi.name}"
          class="w-16 h-16 rounded-full object-cover border-2 border-white shadow-lg"
          onerror="this.style.display='none'; console.error('❌ Image failed to load for POI ${poi.name}: ${poi.imageUrl}');"
          onload="console.log('✅ Image loaded successfully for POI ${poi.name}');"
        />
        <div class="absolute top-0 right-0 bg-red-500 text-white text-xs font-bold rounded-full w-6 h-6 flex items-center justify-center border-2 border-white shadow-md">
          ${poi.participantCount}
        </div>
        
        <!-- Avatar badges around the POI -->
        ${avatarBadgesHTML}
        
        <!-- Temporary test badge if no participant data -->
        ${poi.participantCount > 0 && (!poi.participants || poi.participants.length === 0) ? `
          <div 
            class="absolute w-6 h-6 rounded-full bg-blue-500 border border-white shadow-md flex items-center justify-center text-white text-xs font-bold"
            style="transform: rotate(-60deg) translateY(-32px) rotate(60deg); transform-origin: center center;"
            title="Test badge - participants data missing"
          >
            ?
          </div>
        ` : ''}
      </div>
      <div class="mt-1 text-center">
        <div class="text-xs font-semibold text-gray-800 max-w-20 truncate">
          ${poi.name}
        </div>
      </div>
    `;
  } else {
    // Update default marker
    element.className = `
      w-20 h-12 rounded-lg border-2 cursor-pointer relative
      flex flex-col items-center justify-center text-white text-xs font-bold
      shadow-lg hover:scale-105 transition-transform duration-200
      ${isFull
        ? 'bg-red-500 border-red-600 cursor-not-allowed'
        : 'bg-green-500 border-green-600'
      }
    `;
    
    element.style.padding = '20px';

    element.innerHTML = `
      <div class="text-center leading-tight">
        <div class="truncate max-w-20">${poi.name}</div>
        <div class="text-xs opacity-90">
          ${poi.participantCount}/${poi.maxParticipants}
        </div>
      </div>
      
      <!-- Avatar badges around the POI -->
      ${avatarBadgesHTML}
      
      <!-- Temporary test badge if no participant data -->
      ${poi.participantCount > 0 && (!poi.participants || poi.participants.length === 0) ? `
        <div 
          class="absolute w-6 h-6 rounded-full bg-blue-500 border border-white shadow-md flex items-center justify-center text-white text-xs font-bold"
          style="transform: rotate(-60deg) translateY(-32px) rotate(60deg); transform-origin: center center;"
          title="Test badge - participants data missing"
        >
          ?
        </div>
      ` : ''}
    `;
  }

  element.title = `${poi.name} - ${poi.participantCount}/${poi.maxParticipants} participants`;
};