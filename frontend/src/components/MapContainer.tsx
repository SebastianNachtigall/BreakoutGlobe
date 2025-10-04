import { useEffect, useRef, useCallback, useState } from 'react';
import { Map, NavigationControl, ScaleControl, Marker } from 'maplibre-gl';
import 'maplibre-gl/dist/maplibre-gl.css';
import { POIContextMenu } from './POIContextMenu';
import { ProfileCard } from './ProfileCard';
import { createPOIMarkerElement, updatePOIMarkerElement } from './POIMarker';
import { createAvatarMarkerElement } from './AvatarMarker';

export interface AvatarData {
  sessionId: string;
  userId?: string;
  displayName?: string;
  avatarURL?: string;
  aboutMe?: string;
  position: {
    lat: number;
    lng: number;
  };
  isCurrentUser: boolean;
  isMoving?: boolean;
  isInCall?: boolean;
  role?: 'user' | 'admin' | 'superadmin';
}

export interface POIParticipant {
  id: string;
  name: string;
  avatarUrl?: string | null;
}

export interface POIData {
  id: string;
  name: string;
  description?: string;
  position: {
    lat: number;
    lng: number;
  };
  participantCount: number;
  maxParticipants: number;
  participants?: POIParticipant[];
  imageUrl?: string;
  thumbnailUrl?: string; // Thumbnail for map display (200x200)
  createdBy: string;
  createdAt: Date;
  // Discussion timer fields - backend only tracks when 2+ users are present
  discussionStartTime?: Date | null;
  isDiscussionActive?: boolean;
  discussionDuration?: number; // Duration in seconds (for testing or when provided directly)
}

export interface MapContainerProps {
  initialCenter?: [number, number];
  initialZoom?: number;
  avatars?: AvatarData[];
  pois?: POIData[];
  onMapClick?: (event: { lngLat: { lng: number; lat: number } }) => void;
  onMapReady?: (map: Map) => void;
  onAvatarMove?: (position: { lat: number; lng: number }) => void;
  onPOIClick?: (poiId: string) => void;
  onPOICreate?: (position: { lat: number; lng: number }) => void;
  onAvatarClick?: (userId: string, clickPosition: { x: number; y: number }) => void;
  showProfileCard?: boolean;
  selectedUserProfile?: import('../types/models').UserProfile;
  onProfileCardClose?: () => void;
}

export const MapContainer: React.FC<MapContainerProps> = ({
  initialCenter = [0, 0],
  initialZoom = 2,
  avatars = [],
  pois = [],
  onMapClick,
  onMapReady,
  onAvatarMove,
  onPOIClick,
  onPOICreate,
  onAvatarClick,
  showProfileCard = false,
  selectedUserProfile,
  onProfileCardClose
}) => {
  const mapContainer = useRef<HTMLDivElement>(null);
  const map = useRef<Map | null>(null);
  const markers = useRef<globalThis.Map<string, Marker>>(new globalThis.Map());
  const poiMarkers = useRef<globalThis.Map<string, Marker>>(new globalThis.Map());
  const animationTimeouts = useRef<globalThis.Map<string, number>>(new globalThis.Map());

  // Context menu state
  const [contextMenu, setContextMenu] = useState<{
    position: { x: number; y: number };
    mapPosition: { lat: number; lng: number };
  } | null>(null);



  // Memoize click handler to prevent re-renders
  const handleMapClick = useCallback((event: { lngLat: { lng: number; lat: number } }) => {
    if (onMapClick) {
      onMapClick(event);
    }
    if (onAvatarMove) {
      onAvatarMove({
        lat: event.lngLat.lat,
        lng: event.lngLat.lng
      });
    }
  }, [onMapClick, onAvatarMove]);







  // Ultra-simple animation system
  const animateMarkerTo = useCallback((marker: Marker, newPosition: [number, number], sessionId: string) => {
    // Clear any existing animation
    const existingTimeout = animationTimeouts.current.get(sessionId);
    if (existingTimeout) {
      clearTimeout(existingTimeout);
    }

    const startPosition = marker.getLngLat();
    const startTime = Date.now();
    const duration = 500; // Match App timeout

    const animate = () => {
      const elapsed = Date.now() - startTime;
      const progress = Math.min(elapsed / duration, 1);

      // Simple linear easing
      const lng = startPosition.lng + (newPosition[0] - startPosition.lng) * progress;
      const lat = startPosition.lat + (newPosition[1] - startPosition.lat) * progress;

      marker.setLngLat([lng, lat]);

      if (progress < 1) {
        const timeoutId = setTimeout(animate, 16);
        animationTimeouts.current.set(sessionId, timeoutId);
      } else {
        animationTimeouts.current.delete(sessionId);
      }
    };

    animate();
  }, []);

  // Initialize map only once
  useEffect(() => {
    if (!mapContainer.current || map.current) return;

    // Initialize map with maximum performance settings
    map.current = new Map({
      container: mapContainer.current,
      style: {
        version: 8,
        sources: {
          'osm': {
            type: 'raster',
            tiles: [
              'https://tile.openstreetmap.org/{z}/{x}/{y}.png'
            ],
            tileSize: 256,
            attribution: '© OpenStreetMap contributors'
          }
        },
        layers: [
          {
            id: 'osm',
            type: 'raster',
            source: 'osm'
          }
        ]
      },
      center: initialCenter,
      zoom: initialZoom,
      attributionControl: false,
      // Performance optimizations
      preserveDrawingBuffer: false,
      antialias: false,
      maxZoom: 18,
      minZoom: 1,
      // Additional performance settings
      renderWorldCopies: true, // Enable world wrapping to prevent grey areas
      fadeDuration: 0, // Disable fade animations for faster rendering
      crossSourceCollisions: false // Disable collision detection between sources
    });

    // Add controls
    map.current.addControl(new NavigationControl({}), 'top-right');
    map.current.addControl(new ScaleControl({}), 'bottom-left');

    // Add click event listener
    map.current.on('click', handleMapClick);

    // Add right-click context menu
    const handleContextMenu = (event: any) => {
      event.preventDefault();

      if (onPOICreate) {
        setContextMenu({
          position: {
            x: event.point.x,
            y: event.point.y
          },
          mapPosition: {
            lat: event.lngLat.lat,
            lng: event.lngLat.lng
          }
        });
      }
    };

    map.current.on('contextmenu', handleContextMenu);

    // Optimize marker rendering during map movements
    map.current.on('movestart', () => {
      // Disable hover transitions during map movement for better performance
      // Apply to both avatar markers and POI markers
      markers.current.forEach(marker => {
        const element = marker.getElement();
        element.style.transition = 'none';
      });
      poiMarkers.current.forEach(marker => {
        const element = marker.getElement();
        element.style.transition = 'none';
      });
    });

    map.current.on('moveend', () => {
      // Re-enable only hover transitions (not position transitions)
      // Apply to both avatar markers and POI markers
      setTimeout(() => {
        markers.current.forEach(marker => {
          const element = marker.getElement();
          element.style.transition = 'transform 0.2s ease'; // Only for hover
        });
        poiMarkers.current.forEach(marker => {
          const element = marker.getElement();
          element.style.transition = 'transform 0.2s ease'; // Only for hover
        });
      }, 50);
    });

    // Notify parent that map is ready
    if (onMapReady) {
      onMapReady(map.current);
    }

    // Cleanup function
    return () => {
      if (map.current) {
        // Clear all animation timeouts
        animationTimeouts.current.forEach(timeout => clearTimeout(timeout));
        animationTimeouts.current.clear();

        // Remove all markers
        markers.current.forEach(marker => marker.remove());
        markers.current.clear();

        // Remove all POI markers
        poiMarkers.current.forEach(marker => marker.remove());
        poiMarkers.current.clear();

        map.current.remove();
        map.current = null;
      }
    };
  }, []); // Only run once on mount

  // Avatar marker management - only animate when explicitly requested
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
      const newPosition: [number, number] = [avatar.position.lng, avatar.position.lat];

      if (!marker) {
        // Create new marker
        const markerElement = createAvatarMarkerElement(avatar, onAvatarClick);
        marker = new Marker({
          element: markerElement,
          pitchAlignment: 'viewport',
          rotationAlignment: 'viewport',
          draggable: false
        })
          .setLngLat(newPosition)
          .addTo(map.current!);

        markers.current.set(avatar.sessionId, marker);
      } else {
        // Update existing marker position
        const currentPos = marker.getLngLat();
        const hasPositionChanged =
          Math.abs(currentPos.lng - newPosition[0]) > 0.000001 ||
          Math.abs(currentPos.lat - newPosition[1]) > 0.000001;

        if (hasPositionChanged) {
          if (avatar.isMoving) {
            // Animate to new position
            animateMarkerTo(marker, newPosition, avatar.sessionId);
          } else {
            // Instant position update
            marker.setLngLat(newPosition);
          }
        }
      }
    });
  }, [avatars, onAvatarClick, animateMarkerTo]);

  // POI marker management
  useEffect(() => {
    if (!map.current) return;

    // Remove POI markers that no longer exist
    const currentPOIIds = new Set((pois || []).map(poi => poi.id));
    poiMarkers.current.forEach((marker: Marker, poiId: string) => {
      if (!currentPOIIds.has(poiId)) {
        marker.remove();
        poiMarkers.current.delete(poiId);
      }
    });

    // Add or update POI markers
    pois.forEach(poi => {
      let marker = poiMarkers.current.get(poi.id);
      const position: [number, number] = [poi.position.lng, poi.position.lat];

      if (!marker) {
        // Create new POI marker
        const markerElement = createPOIMarkerElement(poi, onPOIClick);
        marker = new Marker({
          element: markerElement,
          pitchAlignment: 'viewport',
          rotationAlignment: 'viewport',
          draggable: false
        })
          .setLngLat(position)
          .addTo(map.current!);

        poiMarkers.current.set(poi.id, marker);
      } else {
        // Update existing POI marker content using the update function
        const currentElement = marker.getElement();
        const currentPos = marker.getLngLat();
        const newPosition: [number, number] = [poi.position.lng, poi.position.lat];

        // Update the element with new POI data (including participant badges)
        updatePOIMarkerElement(currentElement, poi);

        // Update position if changed
        const hasPositionChanged =
          Math.abs(currentPos.lng - newPosition[0]) > 0.000001 ||
          Math.abs(currentPos.lat - newPosition[1]) > 0.000001;

        if (hasPositionChanged) {
          marker.setLngLat(newPosition);
        }
      }
    });
  }, [pois, onPOIClick]);

  const handlePOIClick = (poiId: string) => {
    if (onPOIClick) {
      onPOIClick(poiId);
    }
  };

  const handlePOICreate = (position: { lat: number; lng: number }) => {
    if (onPOICreate) {
      onPOICreate(position);
    }
    setContextMenu(null);
  };

  const handleContextMenuClose = () => {
    setContextMenu(null);
  };

  return (
    <div className="relative w-full h-full">
      <div
        ref={mapContainer}
        data-testid="map-container"
        className="w-full h-full"
      />

      {/* Context menu */}
      {contextMenu && (
        <POIContextMenu
          position={contextMenu.position}
          mapPosition={contextMenu.mapPosition}
          onCreatePOI={handlePOICreate}
          onClose={handleContextMenuClose}
        />
      )}

      {/* Profile card */}
      {showProfileCard && selectedUserProfile && onProfileCardClose && (
        <ProfileCard
          userProfile={selectedUserProfile}
          onClose={onProfileCardClose}
        />
      )}
    </div>
  );
};