import { useEffect, useRef } from 'react';
import { Map, NavigationControl, ScaleControl, Marker } from 'maplibre-gl';
import 'maplibre-gl/dist/maplibre-gl.css';

export interface AvatarData {
  sessionId: string;
  position: {
    lat: number;
    lng: number;
  };
  isCurrentUser: boolean;
}

export interface MapContainerProps {
  initialCenter?: [number, number];
  initialZoom?: number;
  avatars?: AvatarData[];
  onMapClick?: (event: { lngLat: { lng: number; lat: number } }) => void;
  onMapReady?: (map: Map) => void;
}

export const MapContainer: React.FC<MapContainerProps> = ({
  initialCenter = [0, 0],
  initialZoom = 2,
  avatars = [],
  onMapClick,
  onMapReady
}) => {
  const mapContainer = useRef<HTMLDivElement>(null);
  const map = useRef<Map | null>(null);
  const markers = useRef<globalThis.Map<string, Marker>>(new globalThis.Map());

  useEffect(() => {
    if (!mapContainer.current || map.current) return;

    // Initialize map
    map.current = new Map({
      container: mapContainer.current,
      style: 'https://demotiles.maplibre.org/style.json',
      center: initialCenter,
      zoom: initialZoom,
      attributionControl: false
    });

    // Add controls
    map.current.addControl(new NavigationControl({}), 'top-right');
    map.current.addControl(new ScaleControl({}), 'bottom-left');

    // Add event listeners
    if (onMapClick) {
      map.current.on('click', onMapClick);
    }

    // Notify parent that map is ready
    if (onMapReady) {
      onMapReady(map.current);
    }

    // Cleanup function
    return () => {
      if (map.current) {
        if (onMapClick) {
          map.current.off('click', onMapClick);
        }

        // Remove all markers
        markers.current.forEach(marker => marker.remove());
        markers.current.clear();

        map.current.remove();
        map.current = null;
      }
    };
  }, [initialCenter, initialZoom, onMapClick, onMapReady]);

  // Update markers when avatars change
  useEffect(() => {
    if (!map.current) return;

    // Remove markers that no longer exist
    const currentSessionIds = new Set(avatars.map(avatar => avatar.sessionId));
    markers.current.forEach((marker: Marker, sessionId: string) => {
      if (!currentSessionIds.has(sessionId)) {
        marker.remove();
        markers.current.delete(sessionId);
      }
    });

    // Add or update markers
    avatars.forEach(avatar => {
      let marker = markers.current.get(avatar.sessionId);

      if (!marker) {
        // Create new marker
        const markerElement = document.createElement('div');
        markerElement.className = `
          w-8 h-8 rounded-full border-2 
          ${avatar.isCurrentUser
            ? 'bg-blue-500 border-blue-600 ring-2 ring-blue-500 ring-opacity-50'
            : 'bg-gray-500 border-gray-600 ring-2 ring-gray-400 ring-opacity-50'
          }
          shadow-lg cursor-pointer hover:scale-110 transition-transform duration-200
          flex items-center justify-center text-white text-xs font-bold
        `;
        markerElement.textContent = avatar.sessionId.charAt(0).toUpperCase();
        markerElement.title = avatar.sessionId;

        marker = new Marker({ element: markerElement })
          .setLngLat([avatar.position.lng, avatar.position.lat])
          .addTo(map.current!);

        markers.current.set(avatar.sessionId, marker);
      } else {
        // Update existing marker position
        marker.setLngLat([avatar.position.lng, avatar.position.lat]);

        // Update marker styling if current user status changed
        const markerElement = marker.getElement();
        markerElement.className = `
          w-8 h-8 rounded-full border-2 
          ${avatar.isCurrentUser
            ? 'bg-blue-500 border-blue-600 ring-2 ring-blue-500 ring-opacity-50'
            : 'bg-gray-500 border-gray-600 ring-2 ring-gray-400 ring-opacity-50'
          }
          shadow-lg cursor-pointer hover:scale-110 transition-transform duration-200
          flex items-center justify-center text-white text-xs font-bold
        `;
      }
    });
  }, [avatars]);

  return (
    <div
      ref={mapContainer}
      data-testid="map-container"
      className="w-full h-full"
    />
  );
};