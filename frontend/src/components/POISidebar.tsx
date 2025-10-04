import { useState } from 'react';
import type { POIData } from './MapContainer';

interface POISidebarProps {
  pois: POIData[];
  onPOIClick: (poi: POIData) => void;
  currentUserPOI: string | null;
}

export function POISidebar({ pois, onPOIClick, currentUserPOI }: POISidebarProps) {
  const [isExpanded, setIsExpanded] = useState(false);

  const sortedPOIs = [...pois].sort((a, b) => {
    // Sort by participant count (descending), then by name
    if (b.participantCount !== a.participantCount) {
      return b.participantCount - a.participantCount;
    }
    return a.name.localeCompare(b.name);
  });

  const handlePOIClick = (poi: POIData) => {
    onPOIClick(poi);
    setIsExpanded(false); // Auto-collapse after selection
  };

  return (
    <div className="absolute left-2 top-2 z-[1000]">
      {/* Collapsed Header - Always Visible */}
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="bg-white shadow-lg rounded-lg px-3 py-2 flex items-center gap-2 hover:bg-gray-50 transition-colors"
      >
        <span className="text-sm font-medium text-gray-700">Points of Interest</span>
        <span className="bg-blue-600 text-white text-xs px-2 py-0.5 rounded-full font-medium">
          {pois.length}
        </span>
        <svg
          className={`w-4 h-4 text-gray-600 transition-transform duration-200 ${
            isExpanded ? 'rotate-180' : ''
          }`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M19 9l-7 7-7-7"
          />
        </svg>
      </button>

      {/* Expanded POI List */}
      {isExpanded && (
        <div className="mt-2 bg-white shadow-lg rounded-lg overflow-hidden w-64 max-h-96">
          <div className="overflow-y-auto max-h-96">
            {sortedPOIs.length === 0 ? (
              <div className="p-4 text-center text-gray-500">
                <div className="text-2xl mb-1">📍</div>
                <p className="text-xs">No POIs yet</p>
              </div>
            ) : (
              <div className="divide-y divide-gray-200">
                {sortedPOIs.map((poi) => (
                  <button
                    key={poi.id}
                    onClick={() => handlePOIClick(poi)}
                    className={`w-full p-3 text-left hover:bg-gray-50 transition-colors ${
                      currentUserPOI === poi.id ? 'bg-blue-50' : ''
                    }`}
                  >
                    <div className="flex items-start gap-2">
                      {/* POI Image or Icon */}
                      {(poi.thumbnailUrl || poi.imageUrl) ? (
                        <img
                          src={poi.thumbnailUrl || poi.imageUrl}
                          alt={poi.name}
                          className="w-10 h-10 rounded-full object-cover flex-shrink-0"
                          onError={(e) => {
                            e.currentTarget.style.display = 'none';
                          }}
                        />
                      ) : (
                        <div className="w-10 h-10 bg-blue-100 rounded-full flex items-center justify-center flex-shrink-0">
                          <span className="text-lg">📍</span>
                        </div>
                      )}

                      {/* POI Info */}
                      <div className="flex-1 min-w-0">
                        <h3 className="font-medium text-sm text-gray-900 truncate">
                          {poi.name}
                        </h3>
                        <div className="flex items-center gap-2 mt-1 text-xs text-gray-500">
                          <span className="flex items-center gap-1">
                            👥 {poi.participantCount}
                          </span>
                          {currentUserPOI === poi.id && (
                            <span className="text-blue-600 font-medium">
                              • You
                            </span>
                          )}
                        </div>
                      </div>
                    </div>
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
