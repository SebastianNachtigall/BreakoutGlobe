import React, { useEffect, useRef } from 'react';

export type CallState = 'idle' | 'calling' | 'ringing' | 'connecting' | 'connected' | 'ended';

interface GroupCallParticipant {
  userId: string;
  displayName: string;
  avatarURL?: string;
}

interface GroupCallModalProps {
  isOpen: boolean;
  onClose: () => void;
  callState: CallState;
  poiId: string;
  poiName?: string;
  participants?: Map<string, GroupCallParticipant>;
  remoteStreams?: Map<string, MediaStream>;
  localStream?: MediaStream | null;
  isAudioEnabled?: boolean;
  isVideoEnabled?: boolean;
  onEndCall?: () => void;
  onToggleAudio?: () => void;
  onToggleVideo?: () => void;
}

export const GroupCallModal: React.FC<GroupCallModalProps> = ({
  isOpen,
  onClose,
  callState,
  poiId,
  poiName,
  participants = new Map(),
  remoteStreams = new Map(),
  localStream,
  isAudioEnabled = true,
  isVideoEnabled = true,
  onEndCall,
  onToggleAudio,
  onToggleVideo
}) => {

  const localVideoRef = useRef<HTMLVideoElement>(null);
  const remoteVideoRefs = useRef<Map<string, HTMLVideoElement>>(new Map());

  // Set up local video stream
  useEffect(() => {
    console.log('🎥 GroupCallModal: Local stream effect triggered', {
      hasVideoRef: !!localVideoRef.current,
      hasLocalStream: !!localStream,
      streamId: localStream?.id,
      participantsSize: participants.size
    });

    if (localStream) {
      // Use a small delay to ensure the video element is rendered
      const timer = setTimeout(() => {
        if (localVideoRef.current) {
          localVideoRef.current.srcObject = localStream;
        }
      }, 100);

      return () => clearTimeout(timer);
    }
  }, [localStream, participants.size]);

  // Set up remote video streams
  useEffect(() => {
    for (const [userId, stream] of remoteStreams) {
      const videoElement = remoteVideoRefs.current.get(userId);
      if (videoElement) {
        videoElement.srcObject = stream;
      }
    }
  }, [remoteStreams]);

  if (!isOpen) return null;

  const getCallStateText = () => {
    switch (callState) {
      case 'connecting':
        return 'Connecting to group call...';
      case 'connected':
        return 'Group call active';
      case 'ended':
        return 'Group call ended';
      default:
        return 'Group call';
    }
  };

  const renderCallControls = () => {
    if (callState === 'connecting') {
      return (
        <div className="flex justify-center">
          <button
            onClick={onEndCall}
            className="bg-red-500 hover:bg-red-600 text-white p-4 rounded-full text-2xl transition-colors"
            title="Leave call"
          >
            📵
          </button>
        </div>
      );
    } else if (callState === 'connected') {
      return (
        <div className="flex justify-center space-x-4">
          <button
            onClick={onToggleAudio}
            className={`p-3 rounded-full text-xl transition-colors ${isAudioEnabled
              ? 'bg-gray-300 hover:bg-gray-400 text-gray-700'
              : 'bg-red-500 hover:bg-red-600 text-white'
              }`}
            title={isAudioEnabled ? 'Mute' : 'Unmute'}
          >
            {isAudioEnabled ? '🎤' : '🔇'}
          </button>
          <button
            onClick={onToggleVideo}
            className={`p-3 rounded-full text-xl transition-colors ${isVideoEnabled
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
            title="Leave call"
          >
            📵
          </button>
        </div>
      );
    } else if (callState === 'ended') {
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

  // Calculate optimal grid layout based on participant count
  const getGridLayout = (participantCount: number) => {
    if (participantCount <= 1) return { cols: 1, rows: 1, gridClass: 'grid-cols-1', height: 'h-80' };
    if (participantCount <= 2) return { cols: 2, rows: 1, gridClass: 'grid-cols-2', height: 'h-80' };
    if (participantCount <= 4) return { cols: 2, rows: 2, gridClass: 'grid-cols-2', height: 'h-96' };
    if (participantCount <= 6) return { cols: 3, rows: 2, gridClass: 'grid-cols-3', height: 'h-96' };
    // For more than 6, still use 3x2 but some will be hidden/scrollable
    return { cols: 3, rows: 2, gridClass: 'grid-cols-3', height: 'h-96' };
  };

  const renderVideoArea = () => {
    if (participants.size > 0) {
      console.log('🎥 GroupCallModal: Rendering participants:', Array.from(participants.entries()).map(([userId, participant]) => ({
        userId,
        displayName: participant.displayName,
        avatarUrl: participant.avatarUrl
      })));
      
      const layout = getGridLayout(participants.size);
      
      return (
        <div className={`relative bg-gray-900 ${layout.height} rounded-lg overflow-hidden mb-6`}>
          {/* Remote video grid */}
          <div className={`grid ${layout.gridClass} gap-2 h-full p-2`}>
            {Array.from(participants.entries()).map(([userId, participant]) => {
              const stream = remoteStreams.get(userId);
              return (
                <div key={userId} className="relative bg-gray-800 rounded-lg overflow-hidden">
                  {stream ? (
                    <video
                      ref={(el) => {
                        if (el) {
                          remoteVideoRefs.current.set(userId, el);
                          el.srcObject = stream;
                        }
                      }}
                      autoPlay
                      playsInline
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center text-white">
                      <div className="text-center">
                        {participant.avatarURL ? (
                          <img
                            src={participant.avatarURL}
                            alt={participant.displayName}
                            className="w-16 h-16 rounded-full mx-auto mb-2"
                          />
                        ) : (
                          <div className="w-16 h-16 bg-gray-600 rounded-full flex items-center justify-center text-xl font-bold mx-auto mb-2">
                            {participant.displayName.charAt(0).toUpperCase()}
                          </div>
                        )}
                        <p className="text-sm">{participant.displayName}</p>
                        <p className="text-xs text-gray-300">Waiting for video...</p>
                      </div>
                    </div>
                  )}

                  {/* Participant name overlay */}
                  <div className="absolute bottom-2 left-2 bg-black bg-opacity-50 text-white px-2 py-1 rounded text-sm">
                    {participant.displayName}
                  </div>
                </div>
              );
            })}
          </div>

          {/* Local video (picture-in-picture) */}
          <div
            className="absolute top-4 right-4 w-32 h-24 bg-gray-800 rounded-lg border-2 border-white overflow-hidden"
            data-testid="local-video"
          >
            {localStream && isVideoEnabled ? (
              <video
                ref={(el) => {
                  localVideoRef.current = el;
                  if (el && localStream) {
                    el.srcObject = localStream;
                  }
                }}
                autoPlay
                playsInline
                muted
                className="w-full h-full object-cover"
                style={{ backgroundColor: '#1f2937' }} // Ensure visibility
              />
            ) : (
              <div className="w-full h-full flex items-center justify-center">
                <span className="text-white text-xs">
                  {!isVideoEnabled ? 'Camera off' : localStream ? 'Video disabled' : 'No stream'}
                </span>

              </div>
            )}
          </div>
        </div>
      );
    } else {
      // Call state display (connecting, etc.)
      return (
        <div className="text-center py-8">
          <div className="w-24 h-24 bg-blue-600 rounded-full flex items-center justify-center text-4xl mx-auto mb-4">
            🏢
          </div>
          <h3 className="text-xl font-semibold mb-2 text-gray-800">
            {poiName || `POI ${poiId.substring(0, 8)}`}
          </h3>
          <p className="text-gray-600 mb-6">{getCallStateText()}</p>

          {/* Loading animation for connecting state */}
          {callState === 'connecting' && (
            <div className="flex justify-center mb-6">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
            </div>
          )}

          {/* Simple message for now */}
          {callState === 'connecting' && (
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
              <p className="text-blue-800">
                Group call active in this POI! Video functionality coming soon.
              </p>
            </div>
          )}
        </div>
      );
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center z-[9999]">
      <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-hidden">
        {/* Header */}
        <div className="bg-gray-800 text-white p-4 flex justify-between items-center">
          <div className="flex items-center space-x-3">
            <div className="w-8 h-8 bg-blue-600 rounded-full flex items-center justify-center">
              🏢
            </div>
            <div>
              <h3 className="font-semibold">POI Group Call</h3>
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

        {/* Content Area */}
        <div className="p-6">
          {renderVideoArea()}
          {renderCallControls()}
        </div>
      </div>
    </div>
  );
};