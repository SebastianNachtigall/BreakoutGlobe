import { useState } from 'react';
import type { POIData } from './MapContainer';

interface POISidebarProps {
  pois: POIData[];
  onPOIClick: (poi: POIData) => void;
  currentUserPOI: string | null;
}

export function POISidebar({ pois, onPOIClick, currentUserPOI }: POISidebarProps) {
  const [isCollapsed, setIsCollapsed] = useState(false);

  const sortedPOIs = [...pois].sort((a, b) => {
    // Sort by participant count (descending), then by name
    if (b.participantCount !== a.participantCount) {
      return b.participantCount - a.participantCount;
    }
    return a.name.localeCompare(b.name);
  });

  return (
    <>
      {/* Sidebar */}
      <div
        className={`absolute left-0 top-0 bottom-0 bg-white shadow-lg transition-transform duration-300 z-10 ${
          isCollapsed ? '-translate-x-full' : 'translate-x-0'
        } w-80 md:w-96`}
      >
        <div className="h-full flex flex-col">
          {/* Header */}
          <div className="bg-blue-600 text-white p-4 flex justify-between items-center">
            <h2 className="text-lg font-semibold">Points of Interest</h2>
            <span className="bg-blue-700 px-2 py-1 rounded text-sm">
              {pois.length}
            </span>
          </div>

          {/* POI List */}
          <div className="flex-1 overflow-y-auto">
            {sortedPOIs.length === 0 ? (
              <div className="p-6 text-center text-gray-500">
                <div className="text-4xl mb-2">📍</div>
                <p className="text-sm">No POIs on the map yet</p>
                <p className="text-xs mt-1">Right-click on the map to create one</p>
              </div>
            ) : (
              <div className="divide-y divide-gray-200">
                {sortedPOIs.map((poi) => (
                  <button
                    key={poi.id}
                    onClick={() => onPOIClick(poi)}
                    className={`w-full p-4 text-left hover:bg-gray-50 transition-colors ${
                      currentUserPOI === poi.id ? 'bg-blue-50 border-l-4 border-blue-600' : ''
                    }`}
                  >
                    <div className="flex items-start gap-3">
                      {/* POI Image or Icon */}
                      {poi.imageUrl ? (
                        <img
                          src={poi.imageUrl}
                          alt={poi.name}
                          className="w-12 h-12 rounded object-cover flex-shrink-0"
                          onError={(e) => {
                            e.currentTarget.style.display = 'none';
                          }}
                        />
                      ) : (
                        <div className="w-12 h-12 bg-blue-100 rounded flex items-center justify-center flex-shrink-0">
                          <span className="text-2xl">📍</span>
                        </div>
                      )}

                      {/* POI Info */}
                      <div className="flex-1 min-w-0">
                        <h3 className="font-semibold text-gray-900 truncate">
                          {poi.name}
                        </h3>
                        {poi.description && (
                          <p className="text-sm text-gray-600 line-clamp-2 mt-1">
                            {poi.description}
                          </p>
                        )}
                        <div className="flex items-center gap-3 mt-2 text-xs text-gray-500">
                          <span className="flex items-center gap-1">
                            👥 {poi.participantCount} {poi.participantCount === 1 ? 'person' : 'people'}
                          </span>
                          {currentUserPOI === poi.id && (
                            <span className="text-blue-600 font-medium">
                              • You're here
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
      </div>

      {/* Toggle Button */}
      <button
        onClick={() => setIsCollapsed(!isCollapsed)}
        className={`absolute left-0 top-1/2 -translate-y-1/2 bg-white shadow-lg rounded-r-lg p-2 z-20 transition-transform duration-300 hover:bg-gray-50 ${
          isCollapsed ? 'translate-x-0' : 'translate-x-80 md:translate-x-96'
        }`}
        aria-label={isCollapsed ? 'Show POI list' : 'Hide POI list'}
      >
        <svg
          className={`w-5 h-5 text-gray-600 transition-transform duration-300 ${
            isCollapsed ? '' : 'rotate-180'
          }`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9 5l7 7-7 7"
          />
        </svg>
      </button>
    </>
  );
}
