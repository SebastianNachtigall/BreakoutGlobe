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