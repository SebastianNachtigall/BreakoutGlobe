import React from 'react';

interface AvatarTooltipProps {
  isOpen: boolean;
  position: { x: number; y: number };
  avatar: {
    sessionId: string;
    userId: string;
    displayName: string;
    avatarURL?: string;
    aboutMe?: string;
  };
  onClose: () => void;
  onStartCall: () => void;
}

export const AvatarTooltip: React.FC<AvatarTooltipProps> = ({
  isOpen,
  position,
  avatar,
  onClose,
  onStartCall
}) => {
  if (!isOpen) return null;

  const renderAvatar = () => {
    if (avatar.avatarURL) {
      return (
        <img 
          src={avatar.avatarURL} 
          alt={avatar.displayName}
          className="w-16 h-16 rounded-full object-cover"
        />
      );
    } else {
      return (
        <div className="w-16 h-16 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center text-white text-xl font-bold">
          {avatar.displayName.charAt(0).toUpperCase()}
        </div>
      );
    }
  };

  return (
    <>
      {/* Backdrop to close tooltip */}
      <div
        className="fixed inset-0 z-[8999]"
        onClick={onClose}
      />
      
      {/* Tooltip */}
      <div
        className="absolute z-[9000] bg-white rounded-lg shadow-xl border border-gray-200 p-4 min-w-64 max-w-80"
        style={{
          left: `${position.x}px`,
          top: `${position.y}px`,
          transform: 'translate(-50%, -100%)', // Center horizontally, position above click point
        }}
      >
        {/* Profile Section */}
        <div className="flex items-start space-x-3 mb-4">
          {renderAvatar()}
          <div className="flex-1 min-w-0">
            <h3 className="text-lg font-semibold text-gray-900 truncate">
              {avatar.displayName}
            </h3>
            <p className="text-sm text-gray-500 truncate">
              {avatar.userId}
            </p>
            <p className="text-sm text-gray-700 mt-2 line-clamp-3">
              {avatar.aboutMe || 'No bio available'}
            </p>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex space-x-2">
          <button
            onClick={onStartCall}
            className="flex-1 bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-medium transition-colors flex items-center justify-center space-x-2"
          >
            <span>📹</span>
            <span>Start Video Call</span>
          </button>
          <button
            onClick={onClose}
            className="px-3 py-2 text-gray-500 hover:text-gray-700 transition-colors"
            title="Close"
          >
            ✕
          </button>
        </div>

        {/* Tooltip Arrow */}
        <div className="absolute top-full left-1/2 transform -translate-x-1/2">
          <div className="w-0 h-0 border-l-8 border-r-8 border-t-8 border-l-transparent border-r-transparent border-t-white"></div>
          <div className="absolute top-[-9px] left-1/2 transform -translate-x-1/2 w-0 h-0 border-l-8 border-r-8 border-t-8 border-l-transparent border-r-transparent border-t-gray-200"></div>
        </div>
      </div>
    </>
  );
};