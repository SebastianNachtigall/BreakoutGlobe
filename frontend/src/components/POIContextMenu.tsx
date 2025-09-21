import React, { useEffect } from 'react';

export interface POIContextMenuProps {
  position: { x: number; y: number };
  mapPosition: { lat: number; lng: number };
  onCreatePOI: (position: { lat: number; lng: number }) => void;
  onClose: () => void;
}

export const POIContextMenu: React.FC<POIContextMenuProps> = ({
  position,
  mapPosition,
  onCreatePOI,
  onClose
}) => {
  // Close menu on outside click or Escape key
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Element;
      if (!target.closest('[data-testid="poi-context-menu"]')) {
        onClose();
      }
    };

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose();
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    document.addEventListener('keydown', handleKeyDown);

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [onClose]);

  const handleCreatePOI = () => {
    onCreatePOI(mapPosition);
    onClose();
  };

  return (
    <div
      data-testid="poi-context-menu"
      className="absolute z-50 bg-white border border-gray-300 rounded-lg shadow-lg py-2 min-w-32"
      style={{
        left: `${position.x}px`,
        top: `${position.y}px`
      }}
    >
      <button
        className="w-full px-4 py-2 text-left hover:bg-gray-100 text-sm font-medium text-gray-700"
        onClick={handleCreatePOI}
      >
        Create POI
      </button>
    </div>
  );
};