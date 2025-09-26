import { create } from 'zustand';
import { WebRTCService, MediaConstraints } from '../services/webrtc-service';

// Get WebSocket client instance (will be set by App.tsx)
let wsClient: any = null;

export const setWebSocketClient = (client: any) => {
  wsClient = client;
};

export type CallState = 'idle' | 'calling' | 'ringing' | 'connecting' | 'connected' | 'ended';

interface CallInfo {
  callId: string;
  targetUserId: string;
  targetUserName: string;
  targetUserAvatar?: string;
  isIncoming: boolean;
  createdAt: Date;
}

interface VideoCallState {
  // Current call state
  callState: CallState;
  currentCall: CallInfo | null;
  
  // WebRTC state
  webrtcService: WebRTCService | null;
  localStream: MediaStream | null;
  remoteStream: MediaStream | null;
  isAudioEnabled: boolean;
  isVideoEnabled: boolean;
  
  // Actions
  initiateCall: (targetUserId: string, targetUserName: string, targetUserAvatar?: string) => void;
  receiveCall: (callId: string, fromUserId: string, fromUserName: string, fromUserAvatar?: string) => void;
  acceptCall: () => void;
  rejectCall: () => void;
  endCall: () => void;
  setCallState: (state: CallState) => void;
  clearCall: () => void;
  
  // WebRTC actions
  toggleAudio: () => void;
  toggleVideo: () => void;
  setLocalStream: (stream: MediaStream | null) => void;
  setRemoteStream: (stream: MediaStream | null) => void;
}

export const videoCallStore = create<VideoCallState>((set, get) => ({
  // Initial state
  callState: 'idle',
  currentCall: null,
  
  // WebRTC initial state
  webrtcService: null,
  localStream: null,
  remoteStream: null,
  isAudioEnabled: true,
  isVideoEnabled: true,
  
  // Actions
  initiateCall: async (targetUserId, targetUserName, targetUserAvatar) => {
    const callId = `call-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    
    console.log('📞 Initiating call to:', targetUserName);
    
    set({
      callState: 'calling',
      currentCall: {
        callId,
        targetUserId,
        targetUserName,
        targetUserAvatar,
        isIncoming: false,
        createdAt: new Date()
      }
    });
    
    // Initialize WebRTC for outgoing call (caller)
    try {
      const webrtcService = new WebRTCService();
      
      webrtcService.setCallbacks({
        onLocalStream: (stream) => {
          console.log('📹 Local stream received');
          get().setLocalStream(stream);
        },
        onRemoteStream: (stream) => {
          console.log('📺 Remote stream received');
          get().setRemoteStream(stream);
          set({ callState: 'connected' });
        },
        onIceCandidate: (candidate) => {
          console.log('🧊 ICE candidate generated for outgoing call:', candidate);
          const currentCall = get().currentCall;
          if (currentCall && wsClient && wsClient.isConnected()) {
            wsClient.sendICECandidate(currentCall.callId, currentCall.targetUserId, candidate);
          }
        },
        onConnectionStateChange: (state) => {
          console.log('🔄 Connection state:', state);
          if (state === 'failed' || state === 'disconnected') {
            get().endCall();
          }
        },
        onError: (error) => {
          console.error('❌ WebRTC error:', error);
          get().endCall();
        }
      });
      
      // Get local media
      await webrtcService.initializeLocalMedia({ video: true, audio: true });
      
      set({ webrtcService });
      
      console.log('🎥 WebRTC initialized for outgoing call');
    } catch (error) {
      console.error('❌ Failed to initialize WebRTC:', error);
      set({ callState: 'ended' });
      setTimeout(() => get().clearCall(), 2000);
      return;
    }
    
    // Send call request via WebSocket
    if (wsClient && wsClient.isConnected()) {
      wsClient.sendCallRequest(targetUserId, callId, targetUserName);
    } else {
      console.warn('📞 WebSocket not connected, cannot send call request');
      // Simulate call failure after a delay
      setTimeout(() => {
        const { callState } = get();
        if (callState === 'calling') {
          set({ callState: 'ended' });
          setTimeout(() => get().clearCall(), 2000);
        }
      }, 3000);
    }
  },
  
  receiveCall: (callId, fromUserId, fromUserName, fromUserAvatar) => {
    console.log('📞 Receiving call from:', fromUserName);
    
    set({
      callState: 'ringing',
      currentCall: {
        callId,
        targetUserId: fromUserId,
        targetUserName: fromUserName,
        targetUserAvatar: fromUserAvatar,
        isIncoming: true,
        createdAt: new Date()
      }
    });
  },
  
  acceptCall: async () => {
    const { currentCall } = get();
    if (!currentCall) {
      console.warn('No current call to accept');
      return;
    }
    
    console.log('✅ Call accepted - initializing WebRTC first');
    set({ callState: 'connecting' });
    
    // Initialize WebRTC for incoming call (answerer) BEFORE sending accept
    try {
      const webrtcService = new WebRTCService();
      
      webrtcService.setCallbacks({
        onLocalStream: (stream) => {
          console.log('📹 Local stream received');
          get().setLocalStream(stream);
        },
        onRemoteStream: (stream) => {
          console.log('📺 Remote stream received');
          get().setRemoteStream(stream);
          set({ callState: 'connected' });
        },
        onIceCandidate: (candidate) => {
          console.log('🧊 ICE candidate generated for incoming call:', candidate);
          const currentCall = get().currentCall;
          if (currentCall && wsClient && wsClient.isConnected()) {
            wsClient.sendICECandidate(currentCall.callId, currentCall.targetUserId, candidate);
          }
        },
        onConnectionStateChange: (state) => {
          console.log('🔄 Connection state:', state);
          if (state === 'failed' || state === 'disconnected') {
            get().endCall();
          }
        },
        onError: (error) => {
          console.error('❌ WebRTC error:', error);
          get().endCall();
        }
      });
      
      // Get local media
      await webrtcService.initializeLocalMedia({ video: true, audio: true });
      
      set({ webrtcService });
      
      console.log('🎥 WebRTC initialized for incoming call');
      
      // NOW send call accept via WebSocket after WebRTC is ready
      if (wsClient && wsClient.isConnected()) {
        console.log('📤 Sending call accept after WebRTC initialization');
        wsClient.sendCallAccept(currentCall.callId, currentCall.targetUserId);
      }
      
      // The caller will send an offer after receiving our accept
    } catch (error) {
      console.error('❌ Failed to initialize WebRTC:', error);
      set({ callState: 'ended' });
      setTimeout(() => get().clearCall(), 2000);
    }
  },
  
  rejectCall: () => {
    const { currentCall } = get();
    if (!currentCall) {
      console.warn('No current call to reject');
      return;
    }
    
    console.log('❌ Call rejected');
    
    // Send call reject via WebSocket
    if (wsClient && wsClient.isConnected()) {
      wsClient.sendCallReject(currentCall.callId, currentCall.targetUserId);
    }
    
    set({ 
      callState: 'ended',
    });
    
    // Auto-clear after showing "ended" state briefly
    setTimeout(() => {
      get().clearCall();
    }, 2000);
  },
  
  endCall: () => {
    const { currentCall, webrtcService } = get();
    if (!currentCall) {
      console.warn('No current call to end');
      return;
    }
    
    console.log('📵 Call ended');
    
    // Send call end via WebSocket
    if (wsClient && wsClient.isConnected()) {
      wsClient.sendCallEnd(currentCall.callId, currentCall.targetUserId);
    }
    
    // Clean up WebRTC resources
    if (webrtcService) {
      webrtcService.cleanup();
    }
    
    set({ 
      callState: 'ended',
      webrtcService: null,
      localStream: null,
      remoteStream: null,
      isAudioEnabled: true,
      isVideoEnabled: true
    });
    
    // Auto-clear after showing "ended" state briefly
    setTimeout(() => {
      get().clearCall();
    }, 2000);
  },
  
  setCallState: (state) => {
    console.log('📞 Call state changed to:', state);
    set({ callState: state });
  },
  
  clearCall: () => {
    console.log('🧹 Clearing call state');
    set({
      callState: 'idle',
      currentCall: null,
      webrtcService: null,
      localStream: null,
      remoteStream: null,
      isAudioEnabled: true,
      isVideoEnabled: true
    });
  },
  
  // WebRTC actions
  toggleAudio: () => {
    const { webrtcService, isAudioEnabled } = get();
    if (webrtcService) {
      const newState = webrtcService.toggleAudio();
      set({ isAudioEnabled: newState });
    }
  },
  
  toggleVideo: () => {
    const { webrtcService, isVideoEnabled } = get();
    if (webrtcService) {
      const newState = webrtcService.toggleVideo();
      set({ isVideoEnabled: newState });
    }
  },
  
  setLocalStream: (stream: MediaStream | null) => {
    set({ localStream: stream });
  },
  
  setRemoteStream: (stream: MediaStream | null) => {
    set({ remoteStream: stream });
  }
}));