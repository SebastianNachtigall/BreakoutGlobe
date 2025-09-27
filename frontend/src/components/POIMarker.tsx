import React from 'react';
import { POIData } from './MapContainer';

export interface POIMarkerProps {
  poi: POIData;
  onPOIClick?: (poiId: string) => void;
}

export const POIMarker: React.FC<POIMarkerProps> = ({ poi, onPOIClick }) => {
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
        shadow-lg hover:scale-105 transition-transform duration-200
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
    </div>
  );
};

// Utility function to create a DOM element using the same logic as the React component
export const createPOIMarkerElement = (poi: POIData, onPOIClick?: (poiId: string) => void): HTMLElement => {
  console.log(`🖼️ Creating POI marker for "${poi.name}" with imageUrl:`, poi.imageUrl);

  const markerElement = document.createElement('div');
  const isFull = poi.participantCount >= poi.maxParticipants;

  // If POI has an image, create circular image-based marker
  if (poi.imageUrl) {
    markerElement.className = `
      flex flex-col items-center cursor-pointer hover:scale-105 transition-transform duration-200
      ${isFull ? 'cursor-not-allowed' : ''}
    `;

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
      w-20 h-12 rounded-lg border-2 cursor-pointer
      flex flex-col items-center justify-center text-white text-xs font-bold
      shadow-lg hover:scale-105 transition-transform duration-200
      ${isFull
        ? 'bg-red-500 border-red-600 cursor-not-allowed'
        : 'bg-green-500 border-green-600'
      }
    `;

    markerElement.innerHTML = `
      <div class="text-center leading-tight">
        <div class="truncate max-w-20">${poi.name}</div>
        <div class="text-xs opacity-90">
          ${poi.participantCount}/${poi.maxParticipants}
        </div>
      </div>
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