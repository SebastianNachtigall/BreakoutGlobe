import React, { useEffect, useRef } from 'react';

export type CallState = 'idle' | 'calling' | 'ringing' | 'connecting' | 'connected' | 'ended';

interface VideoCallModalProps {
  isOpen: boolean;
  onClose: () => void;
  callState: CallState;
  targetUser: {
    id: string;
    displayName: string;
    avatarURL?: string;
  };
  localStream?: MediaStream | null;
  remoteStream?: MediaStream | null;
  isAudioEnabled?: boolean;
  isVideoEnabled?: boolean;
  onAcceptCall?: () => void;
  onRejectCall?: () => void;
  onEndCall?: () => void;
  onToggleAudio?: () => void;
  onToggleVideo?: () => void;
}

export const VideoCallModal: React.FC<VideoCallModalProps> = ({
  isOpen,
  onClose,
  callState,
  targetUser,
  localStream,
  remoteStream,
  isAudioEnabled = true,
  isVideoEnabled = true,
  onAcceptCall,
  onRejectCall,
  onEndCall,
  onToggleAudio,
  onToggleVideo
}) => {
  const localVideoRef = useRef<HTMLVideoElement>(null);
  const remoteVideoRef = useRef<HTMLVideoElement>(null);

  // Set up video streams
  useEffect(() => {
    if (localVideoRef.current && localStream) {
      localVideoRef.current.srcObject = localStream;
    }
  }, [localStream]);

  useEffect(() => {
    if (remoteVideoRef.current && remoteStream) {
      remoteVideoRef.current.srcObject = remoteStream;
    }
  }, [remoteStream]);

  if (!isOpen) return null;

  const getCallStateText = () => {
    switch (callState) {
      case 'calling':
        return 'Calling...';
      case 'ringing':
        return 'Incoming video call';
      case 'connecting':
        return 'Connecting...';
      case 'connected':
        return 'Connected';
      case 'ended':
        return 'Call ended';
      default:
        return '';
    }
  };

  const renderUserAvatar = () => {
    if (targetUser.avatarURL) {
      return (
        <img 
          src={targetUser.avatarURL} 
          alt={targetUser.displayName}
          className="w-24 h-24 rounded-full mx-auto mb-4"
        />
      );
    } else {
      return (
        <div className="w-24 h-24 bg-gray-600 rounded-full flex items-center justify-center text-2xl font-bold mx-auto mb-4 text-white">
          {targetUser.displayName.charAt(0).toUpperCase()}
        </div>
      );
    }
  };

  const renderCallControls = () => {
    if (callState === 'ringing') {
      // Incoming call controls
      return (
        <div className="flex justify-center space-x-6">
          <button
            onClick={onAcceptCall}
            className="bg-green-500 hover:bg-green-600 text-white p-4 rounded-full text-2xl transition-colors"
            title="Accept call"
          >
            📞
          </button>
          <button
            onClick={onRejectCall}
            className="bg-red-500 hover:bg-red-600 text-white p-4 rounded-full text-2xl transition-colors"
            title="Reject call"
          >
            📵
          </button>
        </div>
      );
    } else if (callState === 'connected') {
      // Active call controls
      return (
        <div className="flex justify-center space-x-4">
          <button
            onClick={onToggleAudio}
            className={`p-3 rounded-full text-xl transition-colors ${
              isAudioEnabled 
                ? 'bg-gray-300 hover:bg-gray-400 text-gray-700' 
                : 'bg-red-500 hover:bg-red-600 text-white'
            }`}
            title={isAudioEnabled ? 'Mute' : 'Unmute'}
          >
            {isAudioEnabled ? '🎤' : '🔇'}
          </button>
          <button
            onClick={onToggleVideo}
            className={`p-3 rounded-full text-xl transition-colors ${
              isVideoEnabled 
                ? 'bg-gray-300 hover:bg-gray-400 text-gray-700' 
                : 'bg-red-500 hover:bg-red-600 text-white'
            }`}
            title={isVideoEnabled ? 'Turn off camera' : 'Turn on camera'}
          >
            {isVideoEnabled ? '🎥' : '📹'}
          </button>
          <button
            onClick={onEndCall}
            className="bg-red-500 hover:bg-red-600 text-white p-3 rounded-full text-xl transition-colors"
            title="End call"
          >
            📵
          </button>
        </div>
      );
    } else if (callState === 'calling' || callState === 'connecting') {
      // Outgoing call controls
      return (
        <div className="flex justify-center">
          <button
            onClick={onEndCall}
            className="bg-red-500 hover:bg-red-600 text-white p-4 rounded-full text-2xl transition-colors"
            title="Cancel call"
          >
            📵
          </button>
        </div>
      );
    } else if (callState === 'ended') {
      // Call ended - just close button
      return (
        <div className="flex justify-center">
          <button
            onClick={onClose}
            className="bg-gray-500 hover:bg-gray-600 text-white px-6 py-2 rounded-lg transition-colors"
          >
            Close
          </button>
        </div>
      );
    }
    
    return null;
  };

  const renderVideoArea = () => {
    if (callState === 'connected') {
      return (
        <div className="relative bg-gray-900 h-80 rounded-lg overflow-hidden mb-6">
          {/* Remote video */}
          {remoteStream ? (
            <video
              ref={remoteVideoRef}
              autoPlay
              playsInline
              className="w-full h-full object-cover"
            />
          ) : (
            <div className="w-full h-full flex items-center justify-center text-white">
              <div className="text-center">
                {renderUserAvatar()}
                <p className="text-lg">{targetUser.displayName}</p>
                <p className="text-sm text-gray-300">Waiting for video...</p>
              </div>
            </div>
          )}
          
          {/* Local video (picture-in-picture) */}
          <div className="absolute top-4 right-4 w-32 h-24 bg-gray-800 rounded-lg border-2 border-white overflow-hidden">
            {localStream && isVideoEnabled ? (
              <video
                ref={localVideoRef}
                autoPlay
                playsInline
                muted
                className="w-full h-full object-cover"
              />
            ) : (
              <div className="w-full h-full flex items-center justify-center">
                <span className="text-white text-xs">
                  {!isVideoEnabled ? 'Camera off' : 'No video'}
                </span>
              </div>
            )}
          </div>
        </div>
      );
    } else {
      // Call state display
      return (
        <div className="text-center py-8">
          {renderUserAvatar()}
          <h3 className="text-xl font-semibold mb-2 text-gray-800">{targetUser.displayName}</h3>
          <p className="text-gray-600 mb-6">{getCallStateText()}</p>
          
          {/* Loading animation for calling/connecting states */}
          {(callState === 'calling' || callState === 'connecting') && (
            <div className="flex justify-center mb-6">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
            </div>
          )}
        </div>
      );
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-hidden">
        {/* Header */}
        <div className="bg-gray-800 text-white p-4 flex justify-between items-center">
          <div className="flex items-center space-x-3">
            <div className="w-8 h-8 bg-blue-600 rounded-full flex items-center justify-center">
              📹
            </div>
            <div>
              <h3 className="font-semibold">Video Call</h3>
              <p className="text-sm text-gray-300">{getCallStateText()}</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-white text-xl transition-colors"
            title="Close"
          >
            ✕
          </button>
        </div>

        {/* Video/Content Area */}
        <div className="p-6">
          {renderVideoArea()}
          {renderCallControls()}
        </div>
      </div>
    </div>
  );
};