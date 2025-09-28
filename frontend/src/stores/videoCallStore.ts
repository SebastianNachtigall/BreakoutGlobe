import { create } from 'zustand';
import { WebRTCService, GroupWebRTCService, MediaConstraints } from '../services/webrtc-service';
import { avatarStore } from './avatarStore';

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

interface GroupCallParticipant {
  userId: string;
  displayName: string;
  avatarURL?: string;
}

interface VideoCallState {
  // Current call state
  callState: CallState;
  currentCall: CallInfo | null;
  
  // Group call state
  currentPOI: string | null;
  isGroupCallActive: boolean;
  groupCallParticipants: Map<string, GroupCallParticipant>;
  remoteStreams: Map<string, MediaStream>;
  
  // WebRTC state
  webrtcService: WebRTCService | null;
  groupWebRTCService: GroupWebRTCService | null;
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
  
  // Group call actions
  joinPOICall: (poiId: string) => void;
  leavePOICall: () => void;
  addGroupCallParticipant: (userId: string, participant: GroupCallParticipant) => void;
  removeGroupCallParticipant: (userId: string) => void;
  setRemoteStreamForUser: (userId: string, stream: MediaStream) => void;
  initializeGroupWebRTC: () => Promise<void>;
  addPeerToGroupCall: (userId: string) => Promise<void>;
  removePeerFromGroupCall: (userId: string) => void;
  
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
  
  // Group call initial state
  currentPOI: null,
  isGroupCallActive: false,
  groupCallParticipants: new Map(),
  remoteStreams: new Map(),
  
  // WebRTC initial state
  webrtcService: null,
  groupWebRTCService: null,
  localStream: null,
  remoteStream: null,
  isAudioEnabled: true,
  isVideoEnabled: true,
  
  // Actions
  initiateCall: async (targetUserId, targetUserName, targetUserAvatar) => {
    // Clean up any existing call first
    const { webrtcService, callState } = get();
    if (webrtcService || callState !== 'idle') {
      console.log('🧹 Cleaning up existing call before starting new one');
      get().clearCall();
      // Wait a bit for cleanup to complete
      await new Promise(resolve => setTimeout(resolve, 500));
    }
    
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
    
    // Mark target user as in call
    avatarStore.getState().updateAvatarCallStatus(targetUserId, true);
    
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
          const currentState = get();
          // Only end call on connection failure if we're still in an active call
          if (state === 'failed' && currentState.callState !== 'ended' && currentState.callState !== 'idle') {
            console.log('❌ WebRTC connection failed, ending call');
            get().endCall();
          }
          // Don't auto-end on 'disconnected' as this happens during normal cleanup
        },
        onError: (error) => {
          console.error('❌ WebRTC error:', error);
          const currentState = get();
          if (currentState.callState !== 'ended' && currentState.callState !== 'idle') {
            get().endCall();
          }
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
    
    // Clean up any existing call first
    const { webrtcService, callState } = get();
    if (webrtcService || callState !== 'idle') {
      console.log('🧹 Cleaning up existing call before receiving new one');
      get().clearCall();
    }
    
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
    
    // Mark caller as in call
    avatarStore.getState().updateAvatarCallStatus(fromUserId, true);
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
          const currentState = get();
          // Only end call on connection failure if we're still in an active call
          if (state === 'failed' && currentState.callState !== 'ended' && currentState.callState !== 'idle') {
            console.log('❌ WebRTC connection failed, ending call');
            get().endCall();
          }
          // Don't auto-end on 'disconnected' as this happens during normal cleanup
        },
        onError: (error) => {
          console.error('❌ WebRTC error:', error);
          const currentState = get();
          if (currentState.callState !== 'ended' && currentState.callState !== 'idle') {
            get().endCall();
          }
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
    
    // Mark target user as no longer in call
    avatarStore.getState().updateAvatarCallStatus(currentCall.targetUserId, false);
    
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
    const { currentCall, webrtcService, callState } = get();
    if (!currentCall || callState === 'ended' || callState === 'idle') {
      console.warn('No current call to end');
      return;
    }
    
    console.log('📵 Call ended');
    
    // Send call end via WebSocket (only if we're ending the call, not if we received end)
    if (wsClient && wsClient.isConnected() && callState !== 'ended') {
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
    const { webrtcService, currentCall } = get();
    
    // Mark target user as no longer in call
    if (currentCall) {
      avatarStore.getState().updateAvatarCallStatus(currentCall.targetUserId, false);
    }
    
    // Ensure WebRTC service is properly cleaned up
    if (webrtcService) {
      webrtcService.cleanup();
    }
    
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
  
  // Group call actions
  joinPOICall: (poiId: string) => {
    console.log('🏢 Joining POI group call:', poiId);
    
    // Clean up any existing call first
    const { webrtcService } = get();
    if (webrtcService) {
      webrtcService.cleanup();
    }
    
    // Set the group call state
    set({
      currentPOI: poiId,
      isGroupCallActive: true,
      callState: 'connecting',
      webrtcService: null,
      groupWebRTCService: null,
      localStream: null,
      remoteStream: null,
      groupCallParticipants: new Map(),
      remoteStreams: new Map(),
      isAudioEnabled: true,
      isVideoEnabled: true
    });
  },
  
  leavePOICall: () => {
    console.log('🚪 Leaving POI group call');
    const { webrtcService, groupWebRTCService } = get();
    
    // Clean up WebRTC resources
    if (webrtcService) {
      webrtcService.cleanup();
    }
    if (groupWebRTCService) {
      groupWebRTCService.cleanup();
    }
    
    set({
      currentPOI: null,
      isGroupCallActive: false,
      callState: 'idle',
      webrtcService: null,
      groupWebRTCService: null,
      localStream: null,
      remoteStream: null,
      groupCallParticipants: new Map(),
      remoteStreams: new Map(),
      isAudioEnabled: true,
      isVideoEnabled: true
    });
  },

  addGroupCallParticipant: (userId: string, participant: GroupCallParticipant) => {
    console.log('👥 Adding group call participant:', {
      userId,
      displayName: participant.displayName,
      avatarUrl: participant.avatarUrl,
      participant
    });
    const { groupCallParticipants } = get();
    const newParticipants = new Map(groupCallParticipants);
    newParticipants.set(userId, participant);
    set({ groupCallParticipants: newParticipants });
    
    console.log('👥 Group call participants after adding:', Array.from(newParticipants.entries()).map(([id, p]) => ({
      userId: id,
      displayName: p.displayName
    })));
  },

  removeGroupCallParticipant: (userId: string) => {
    console.log('👋 Removing group call participant:', userId);
    const { groupCallParticipants, remoteStreams } = get();
    
    // Remove participant
    const newParticipants = new Map(groupCallParticipants);
    newParticipants.delete(userId);
    
    // Remove associated stream
    const newStreams = new Map(remoteStreams);
    newStreams.delete(userId);
    
    set({ 
      groupCallParticipants: newParticipants,
      remoteStreams: newStreams
    });
  },

  setRemoteStreamForUser: (userId: string, stream: MediaStream) => {
    console.log('📺 Setting remote stream for user:', userId);
    const { remoteStreams } = get();
    const newStreams = new Map(remoteStreams);
    newStreams.set(userId, stream);
    set({ remoteStreams: newStreams });
  },

  initializeGroupWebRTC: async () => {
    console.log('🔗 Initializing group WebRTC service');
    try {
      const groupWebRTCService = new GroupWebRTCService();
      
      groupWebRTCService.setCallbacks({
        onLocalStream: (stream) => {
          console.log('📹 Group call local stream received');
          get().setLocalStream(stream);
        },
        onRemoteStreamForUser: (userId, stream) => {
          console.log('📺 Group call remote stream received from user:', userId);
          get().setRemoteStreamForUser(userId, stream);
        },
        onIceCandidate: (candidate) => {
          console.log('🧊 Group call ICE candidate generated:', candidate);
          // TODO: Send ICE candidate via websocket
        },
        onPeerConnectionStateChange: (userId, state) => {
          console.log(`🔄 Group call peer connection state changed for user ${userId}:`, state);
          if (state === 'connected') {
            set({ callState: 'connected' });
          }
        },
        onError: (error) => {
          console.error('❌ Group call WebRTC error:', error);
          // TODO: Handle error appropriately
        }
      });

      // Initialize local media
      await groupWebRTCService.initializeLocalMedia({ video: true, audio: true });
      
      // Set WebSocket client for signaling
      const wsClient = (window as any).wsClient;
      if (wsClient) {
        groupWebRTCService.setWebSocketClient(wsClient);
      }
      
      set({ groupWebRTCService });
      console.log('✅ Group WebRTC service initialized');
    } catch (error) {
      console.error('❌ Failed to initialize group WebRTC service:', error);
      throw error;
    }
  },

  addPeerToGroupCall: async (userId: string) => {
    console.log('👥 Adding peer to group call:', userId);
    const { groupWebRTCService } = get();
    
    if (!groupWebRTCService) {
      throw new Error('Group WebRTC service not initialized');
    }

    try {
      await groupWebRTCService.addPeer(userId);
      console.log('✅ Peer added to group call:', userId);
    } catch (error) {
      console.error('❌ Failed to add peer to group call:', error);
      throw error;
    }
  },

  removePeerFromGroupCall: (userId: string) => {
    console.log('👋 Removing peer from group call:', userId);
    const { groupWebRTCService } = get();
    
    if (groupWebRTCService) {
      groupWebRTCService.removePeer(userId);
    }
    
    // Also remove from participants and streams
    get().removeGroupCallParticipant(userId);
    console.log('✅ Peer removed from group call:', userId);
  },
  
  // WebRTC actions
  toggleAudio: () => {
    const { webrtcService, groupWebRTCService, isGroupCallActive, isAudioEnabled } = get();
    if (isGroupCallActive && groupWebRTCService) {
      // For group calls, use groupWebRTCService
      const newState = groupWebRTCService.toggleAudio();
      set({ isAudioEnabled: newState });
    } else if (webrtcService) {
      // For regular calls, use webrtcService
      const newState = webrtcService.toggleAudio();
      set({ isAudioEnabled: newState });
    }
  },
  
  toggleVideo: () => {
    const { webrtcService, groupWebRTCService, isGroupCallActive, isVideoEnabled } = get();
    if (isGroupCallActive && groupWebRTCService) {
      // For group calls, use groupWebRTCService
      const newState = groupWebRTCService.toggleVideo();
      set({ isVideoEnabled: newState });
    } else if (webrtcService) {
      // For regular calls, use webrtcService
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