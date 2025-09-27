import React from 'react';
import { POIData } from './MapContainer';

export interface POIMarkerProps {
  poi: POIData;
  onPOIClick: (poiId: string) => void;
}

export const POIMarker: React.FC<POIMarkerProps> = ({ poi, onPOIClick }) => {
  const isFull = poi.participantCount >= poi.maxParticipants;

  const handleClick = (event: React.MouseEvent) => {
    event.stopPropagation(); // Prevent map click handler from firing
    if (!isFull) {
      onPOIClick(poi.id);
    }
  };

  // If POI has an image, render it with overlay text
  if (poi.imageUrl) {
    return (
      <div
        data-testid="poi-marker"
        className={`
          w-24 h-16 rounded-lg border-2 cursor-pointer relative overflow-hidden
          shadow-lg hover:scale-105 transition-transform duration-200
          ${isFull 
            ? 'border-red-600 cursor-not-allowed' 
            : 'border-green-600'
          }
        `}
        onClick={handleClick}
        title={`${poi.name} - ${poi.participantCount}/${poi.maxParticipants} participants`}
      >
        <img 
          src={poi.imageUrl} 
          alt={poi.name}
          className="w-full h-full object-cover"
          onError={(e) => {
            // Fallback to default marker if image fails to load
            const target = e.target as HTMLImageElement;
            target.style.display = 'none';
          }}
        />
        <div className={`
          absolute bottom-0 left-0 right-0 
          ${isFull ? 'bg-red-500/90' : 'bg-green-500/90'}
          text-white text-xs font-bold px-1 py-0.5
        `}>
          <div className="text-center leading-tight">
            <div className="truncate">{poi.name}</div>
            <div className="text-xs opacity-90">
              {poi.participantCount}/{poi.maxParticipants}
            </div>
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