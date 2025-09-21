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
  isMoving?: boolean;
}

export interface MapContainerProps {
  initialCenter?: [number, number];
  initialZoom?: number;
  avatars?: AvatarData[];
  onMapClick?: (event: { lngLat: { lng: number; lat: number } }) => void;
  onMapReady?: (map: Map) => void;
  onAvatarMove?: (position: { lat: number; lng: number }) => void;
}

export const MapContainer: React.FC<MapContainerProps> = ({
  initialCenter = [0, 0],
  initialZoom = 2,
  avatars = [],
  onMapClick,
  onMapReady,
  onAvatarMove
}) => {
  const mapContainer = useRef<HTMLDivElement>(null);
  const map = useRef<Map | null>(null);
  const markers = useRef<globalThis.Map<string, Marker>>(new globalThis.Map());

  // Helper function to detect and resolve avatar collisions
  const resolveCollisions = (avatars: AvatarData[]): AvatarData[] => {
    const resolved = [...avatars];
    const positionMap = new globalThis.Map<string, number>();
    
    resolved.forEach((avatar, index) => {
      const posKey = `${avatar.position.lat.toFixed(6)},${avatar.position.lng.toFixed(6)}`;
      const existingCount = positionMap.get(posKey) || 0;
      
      if (existingCount > 0) {
        // Apply small offset to prevent overlap
        const offsetDistance = 0.0001; // ~11 meters
        const angle = (existingCount * 60) * (Math.PI / 180); // 60 degrees apart
        
        resolved[index] = {
          ...avatar,
          position: {
            lat: avatar.position.lat + (Math.sin(angle) * offsetDistance),
            lng: avatar.position.lng + (Math.cos(angle) * offsetDistance)
          }
        };
      }
      
      positionMap.set(posKey, existingCount + 1);
    });
    
    return resolved;
  };

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
    if (onMapClick || onAvatarMove) {
      const handleMapClick = (event: { lngLat: { lng: number; lat: number } }) => {
        // Call the original click handler if provided
        if (onMapClick) {
          onMapClick(event);
        }
        
        // Handle avatar movement if provided
        if (onAvatarMove) {
          onAvatarMove({
            lat: event.lngLat.lat,
            lng: event.lngLat.lng
          });
        }
      };
      
      map.current.on('click', handleMapClick);
    }

    // Notify parent that map is ready
    if (onMapReady) {
      onMapReady(map.current);
    }

    // Cleanup function
    return () => {
      if (map.current) {
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

    // Resolve collisions and optimize positioning
    const resolvedAvatars = resolveCollisions(avatars);
    
    // Sort avatars to render current user last (on top)
    const sortedAvatars = [...resolvedAvatars].sort((a, b) => {
      if (a.isCurrentUser && !b.isCurrentUser) return 1;
      if (!a.isCurrentUser && b.isCurrentUser) return -1;
      return 0;
    });

    // Remove markers that no longer exist
    const currentSessionIds = new Set(sortedAvatars.map(avatar => avatar.sessionId));
    markers.current.forEach((marker: Marker, sessionId: string) => {
      if (!currentSessionIds.has(sessionId)) {
        marker.remove();
        markers.current.delete(sessionId);
      }
    });

    // Add or update markers
    sortedAvatars.forEach(avatar => {
      let marker = markers.current.get(avatar.sessionId);

      if (!marker) {
        // Create new marker
        const markerElement = document.createElement('div');
        const animationClasses = avatar.isMoving ? 'transition-all duration-500 ease-in-out' : '';
        
        markerElement.className = `
          w-8 h-8 rounded-full border-2 
          ${avatar.isCurrentUser
            ? 'bg-blue-500 border-blue-600 ring-2 ring-blue-500 ring-opacity-50'
            : 'bg-gray-500 border-gray-600 ring-2 ring-gray-400 ring-opacity-50'
          }
          shadow-lg cursor-pointer hover:scale-110 transition-transform duration-200
          flex items-center justify-center text-white text-xs font-bold
          ${animationClasses}
        `;
        markerElement.textContent = avatar.sessionId.charAt(0).toUpperCase();
        markerElement.title = avatar.sessionId;

        marker = new Marker({ element: markerElement })
          .setLngLat([avatar.position.lng, avatar.position.lat])
          .addTo(map.current!);

        markers.current.set(avatar.sessionId, marker);
      } else {
        // Update existing marker position with smooth animation
        const markerElement = marker.getElement();
        const animationClasses = avatar.isMoving ? 'transition-all duration-500 ease-in-out' : '';
        
        // Update position
        marker.setLngLat([avatar.position.lng, avatar.position.lat]);

        // Update marker styling and animation classes
        markerElement.className = `
          w-8 h-8 rounded-full border-2 
          ${avatar.isCurrentUser
            ? 'bg-blue-500 border-blue-600 ring-2 ring-blue-500 ring-opacity-50'
            : 'bg-gray-500 border-gray-600 ring-2 ring-gray-400 ring-opacity-50'
          }
          shadow-lg cursor-pointer hover:scale-110 transition-transform duration-200
          flex items-center justify-center text-white text-xs font-bold
          ${animationClasses}
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