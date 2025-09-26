import { create } from 'zustand';

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
    
    // TODO: Send call request via WebSocket in Phase 2
    // For now, just simulate the calling state
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
    
    set({ 
      callState: 'ended',
    });
    
    // Auto-clear after showing "ended" state briefly
    setTimeout(() => {
      get().clearCall();
    }, 2000);
    
    // TODO: Send rejection via WebSocket in Phase 2
  },
  
  endCall: () => {
    const { currentCall } = get();
    if (!currentCall) {
      console.warn('No current call to end');
      return;
    }
    
    console.log('📵 Call ended');
    
    set({ callState: 'ended' });
    
    // Auto-clear after showing "ended" state briefly
    setTimeout(() => {
      get().clearCall();
    }, 2000);
    
    // TODO: Send end call via WebSocket in Phase 2
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