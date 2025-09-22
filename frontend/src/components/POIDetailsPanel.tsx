import React, { useEffect } from 'react';

export interface POIParticipant {
  id: string;
  name: string;
}

export interface POIData {
  id: string;
  name: string;
  description: string;
  maxParticipants: number;
  participantCount: number;
  position: { lat: number; lng: number };
  participants: POIParticipant[];
}

export interface POIDetailsPanelProps {
  poi: POIData;
  currentUserId: string;
  isUserParticipant: boolean;
  onJoin: (poiId: string) => void;
  onLeave: (poiId: string) => void;
  onClose: () => void;
  isLoading?: boolean;
  position?: { x: number; y: number };
}

export const POIDetailsPanel: React.FC<POIDetailsPanelProps> = ({
  poi,
  currentUserId,
  isUserParticipant,
  onJoin,
  onLeave,
  onClose,
  isLoading = false,
  position
}) => {
  const isFull = poi.participantCount >= poi.maxParticipants;
  const isNearFull = poi.participantCount >= poi.maxParticipants - 1 && !isFull;

  // Handle escape key
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [onClose]);

  const handleJoin = () => {
    if (!isFull && !isLoading) {
      onJoin(poi.id);
    }
  };

  const handleLeave = () => {
    if (!isLoading) {
      onLeave(poi.id);
    }
  };

  const getActionButtonText = () => {
    if (isLoading) {
      return isUserParticipant ? 'Leaving...' : 'Joining...';
    }
    if (isUserParticipant) {
      return 'Leave';
    }
    if (isFull) {
      return 'Join (Full)';
    }
    return 'Join';
  };

  const panelStyle = position ? {
    position: 'absolute' as const,
    left: `${position.x + 20}px`, // Offset to the right of the POI marker
    top: `${position.y - 50}px`,  // Offset above the POI marker
    zIndex: 1000,
  } : {
    position: 'absolute' as const,
    top: '20px',
    right: '20px',
    zIndex: 1000,
  };

  return (
    <div 
      className="poi-details-panel bg-white border border-gray-300 rounded-lg shadow-lg p-4 max-w-sm"
      style={panelStyle}
    >
      <div className="poi-header flex justify-between items-start mb-4">
        <div>
          <h3 className="text-lg font-semibold text-gray-900">{poi.name}</h3>
          <div className="text-sm text-gray-600 mt-1">
            <span className={`font-medium ${isFull ? 'text-red-600' : isNearFull ? 'text-yellow-600' : 'text-green-600'}`}>
              {poi.participantCount}/{poi.maxParticipants} participants
            </span>
            {isFull && <span className="ml-2 text-red-600 font-medium">(Full)</span>}
            {isNearFull && <span className="ml-2 text-yellow-600 font-medium">(Almost Full)</span>}
          </div>
        </div>
        <button
          onClick={onClose}
          className="text-gray-400 hover:text-gray-600 text-xl font-bold"
          aria-label="Close panel"
        >
          ✕
        </button>
      </div>

      <div className="poi-description mb-4">
        <p className="text-gray-700 text-sm">{poi.description}</p>
      </div>

      <div className="poi-coordinates mb-4">
        <div className="text-xs text-gray-500">
          <span>Lat: {poi.position.lat.toFixed(4)}</span>
          <span className="ml-3">Lng: {poi.position.lng.toFixed(4)}</span>
        </div>
      </div>

      <div className="poi-participants mb-4">
        <h4 className="text-sm font-medium text-gray-900 mb-2">Participants</h4>
        {poi.participants && poi.participants.length > 0 ? (
          <ul className="space-y-1">
            {poi.participants.map((participant) => (
              <li 
                key={participant.id} 
                className="text-sm text-gray-700"
                data-testid={participant.id === currentUserId ? 'current-user' : undefined}
              >
                {participant.name}
                {participant.id === currentUserId && <span className="text-blue-600 font-medium"> (You)</span>}
              </li>
            ))}
          </ul>
        ) : (
          <p className="text-sm text-gray-500 italic">No participants yet</p>
        )}
      </div>

      <div className="poi-actions">
        <button
          onClick={isUserParticipant ? handleLeave : handleJoin}
          disabled={(!isUserParticipant && isFull) || isLoading}
          className={`w-full py-2 px-4 rounded-md font-medium text-sm transition-colors ${
            isUserParticipant
              ? 'bg-red-600 hover:bg-red-700 text-white disabled:bg-red-400'
              : isFull
              ? 'bg-gray-300 text-gray-500 cursor-not-allowed'
              : 'bg-green-600 hover:bg-green-700 text-white disabled:bg-green-400'
          }`}
        >
          {getActionButtonText()}
        </button>
      </div>
    </div>
  );
};