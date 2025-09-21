import { useEffect, useRef, useCallback, useMemo } from 'react';
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
  const animationTimeouts = useRef<globalThis.Map<string, number>>(new globalThis.Map());

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

  // Helper function to detect and resolve avatar collisions
  const resolveCollisions = useCallback((avatars: AvatarData[]): AvatarData[] => {
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
  }, []);

  // Create marker element with NO conflicting animations
  const createMarkerElement = useCallback((avatar: AvatarData) => {
    const markerElement = document.createElement('div');
    markerElement.className = `
      w-8 h-8 rounded-full border-2 
      ${avatar.isCurrentUser
        ? 'bg-blue-500 border-blue-600 ring-2 ring-blue-500 ring-opacity-50'
        : 'bg-gray-500 border-gray-600 ring-2 ring-gray-400 ring-opacity-50'
      }
      shadow-lg cursor-pointer hover:scale-110
      flex items-center justify-center text-white text-xs font-bold
    `;
    markerElement.textContent = avatar.sessionId.charAt(0).toUpperCase();
    markerElement.title = avatar.sessionId;

    // Performance optimizations - NO CSS transitions for position
    markerElement.style.willChange = 'transform';
    markerElement.style.backfaceVisibility = 'hidden';
    markerElement.style.transform = 'translateZ(0)';
    markerElement.style.contain = 'layout style paint';
    markerElement.style.pointerEvents = 'auto';
    // CRITICAL: Only allow hover transitions, no position transitions
    markerElement.style.transition = 'transform 0.2s ease'; // Only for hover scale

    return markerElement;
  }, []);

  // Single, clean animation system with proper easing
  const animateMarkerTo = useCallback((marker: Marker, newPosition: [number, number], sessionId: string) => {
    // Clear any existing animation timeout to prevent conflicts
    const existingTimeout = animationTimeouts.current.get(sessionId);
    if (existingTimeout) {
      clearTimeout(existingTimeout);
      animationTimeouts.current.delete(sessionId);
    }

    const currentLngLat = marker.getLngLat();
    const startTime = Date.now();
    const duration = 600; // Slightly longer for smoother feel

    const animate = () => {
      const elapsed = Date.now() - startTime;
      const progress = Math.min(elapsed / duration, 1);

      // Clean ease-out-quart for very natural movement
      const easeOutQuart = (t: number) => 1 - Math.pow(1 - t, 4);
      const easedProgress = easeOutQuart(progress);

      // Interpolate position
      const lng = currentLngLat.lng + (newPosition[0] - currentLngLat.lng) * easedProgress;
      const lat = currentLngLat.lat + (newPosition[1] - currentLngLat.lat) * easedProgress;

      marker.setLngLat([lng, lat]);

      if (progress < 1) {
        const timeoutId = setTimeout(animate, 16); // ~60fps
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
      style: 'https://demotiles.maplibre.org/style.json',
      center: initialCenter,
      zoom: initialZoom,
      attributionControl: false,
      // Performance optimizations
      preserveDrawingBuffer: false,
      antialias: false,
      maxZoom: 18,
      minZoom: 1,
      // Additional performance settings
      renderWorldCopies: false, // Don't render multiple world copies
      fadeDuration: 0, // Disable fade animations for faster rendering
      crossSourceCollisions: false // Disable collision detection between sources
    });

    // Add controls
    map.current.addControl(new NavigationControl({}), 'top-right');
    map.current.addControl(new ScaleControl({}), 'bottom-left');

    // Add click event listener
    map.current.on('click', handleMapClick);

    // Optimize marker rendering during map movements
    map.current.on('movestart', () => {
      // Disable hover transitions during map movement for better performance
      markers.current.forEach(marker => {
        const element = marker.getElement();
        element.style.transition = 'none';
      });
    });

    map.current.on('moveend', () => {
      // Re-enable only hover transitions (not position transitions)
      setTimeout(() => {
        markers.current.forEach(marker => {
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

        map.current.remove();
        map.current = null;
      }
    };
  }, []); // Only run once on mount

  // Memoize processed avatars to prevent unnecessary recalculations
  const processedAvatars = useMemo(() => {
    const resolved = resolveCollisions(avatars);
    // Sort avatars to render current user last (on top)
    return [...resolved].sort((a, b) => {
      if (a.isCurrentUser && !b.isCurrentUser) return 1;
      if (!a.isCurrentUser && b.isCurrentUser) return -1;
      return 0;
    });
  }, [avatars, resolveCollisions]);

  // Update markers when avatars change
  useEffect(() => {
    if (!map.current) return;

    // Remove markers that no longer exist
    const currentSessionIds = new Set(processedAvatars.map(avatar => avatar.sessionId));
    markers.current.forEach((marker: Marker, sessionId: string) => {
      if (!currentSessionIds.has(sessionId)) {
        // Clear any pending animation
        const timeout = animationTimeouts.current.get(sessionId);
        if (timeout) {
          clearTimeout(timeout);
          animationTimeouts.current.delete(sessionId);
        }

        marker.remove();
        markers.current.delete(sessionId);
      }
    });

    // Add or update markers
    processedAvatars.forEach(avatar => {
      let marker = markers.current.get(avatar.sessionId);
      const newPosition: [number, number] = [avatar.position.lng, avatar.position.lat];

      if (!marker) {
        // Create new marker
        const markerElement = createMarkerElement(avatar);
        marker = new Marker({
          element: markerElement,
          // Critical: Use viewport alignment for better performance during zoom/pan
          pitchAlignment: 'viewport',
          rotationAlignment: 'viewport',
          // Disable marker dragging for better performance
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

        // Update marker styling if needed (no position transitions)
        const markerElement = marker.getElement();
        const expectedClass = `
          w-8 h-8 rounded-full border-2 
          ${avatar.isCurrentUser
            ? 'bg-blue-500 border-blue-600 ring-2 ring-blue-500 ring-opacity-50'
            : 'bg-gray-500 border-gray-600 ring-2 ring-gray-400 ring-opacity-50'
          }
          shadow-lg cursor-pointer hover:scale-110
          flex items-center justify-center text-white text-xs font-bold
        `.replace(/\s+/g, ' ').trim();

        if (markerElement.className.replace(/\s+/g, ' ').trim() !== expectedClass) {
          markerElement.className = expectedClass;
          // Ensure only hover transitions, no position transitions
          markerElement.style.transition = 'transform 0.2s ease';
        }
      }
    });
  }, [processedAvatars, createMarkerElement, animateMarkerTo]);

  return (
    <div
      ref={mapContainer}
      data-testid="map-container"
      className="w-full h-full"
    />
  );
};