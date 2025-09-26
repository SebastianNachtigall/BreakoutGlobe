import { create } from 'zustand';

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
  
  // Actions
  initiateCall: (targetUserId: string, targetUserName: string, targetUserAvatar?: string) => void;
  receiveCall: (callId: string, fromUserId: string, fromUserName: string, fromUserAvatar?: string) => void;
  acceptCall: () => void;
  rejectCall: () => void;
  endCall: () => void;
  setCallState: (state: CallState) => void;
  clearCall: () => void;
}

export const videoCallStore = create<VideoCallState>((set, get) => ({
  // Initial state
  callState: 'idle',
  currentCall: null,
  
  // Actions
  initiateCall: (targetUserId, targetUserName, targetUserAvatar) => {
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
  
  acceptCall: () => {
    const { currentCall } = get();
    if (!currentCall) {
      console.warn('No current call to accept');
      return;
    }
    
    console.log('✅ Call accepted');
    set({ callState: 'connecting' });
    
    // Send call accept via WebSocket
    if (wsClient && wsClient.isConnected()) {
      wsClient.sendCallAccept(currentCall.callId, currentCall.targetUserId);
    }
    
    // Simulate connection process
    setTimeout(() => {
      const { callState } = get();
      if (callState === 'connecting') {
        set({ callState: 'connected' });
        console.log('🎥 Call connected');
      }
    }, 2000);
    
    // TODO: Initialize WebRTC connection in Phase 3
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
    const { currentCall } = get();
    if (!currentCall) {
      console.warn('No current call to end');
      return;
    }
    
    console.log('📵 Call ended');
    
    // Send call end via WebSocket
    if (wsClient && wsClient.isConnected()) {
      wsClient.sendCallEnd(currentCall.callId, currentCall.targetUserId);
    }
    
    set({ callState: 'ended' });
    
    // Auto-clear after showing "ended" state briefly
    setTimeout(() => {
      get().clearCall();
    }, 2000);
    
    // TODO: Clean up WebRTC resources in Phase 3
  },
  
  setCallState: (state) => {
    console.log('📞 Call state changed to:', state);
    set({ callState: state });
  },
  
  clearCall: () => {
    console.log('🧹 Clearing call state');
    set({
      callState: 'idle',
      currentCall: null
    });
  }
}));