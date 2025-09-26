import { sessionStore } from '../stores/sessionStore';
import { poiStore } from '../stores/poiStore';
import { avatarStore } from '../stores/avatarStore';
import type { POIData, AvatarData } from '../components/MapContainer';

export enum ConnectionStatus {
  DISCONNECTED = 'disconnected',
  CONNECTING = 'connecting',
  CONNECTED = 'connected',
  RECONNECTING = 'reconnecting'
}

export interface WebSocketMessage {
  type: string;
  data: any;
  timestamp: Date;
}

export interface WebSocketError {
  message: string;
  code?: number;
  timestamp: Date;
}

export interface StateSync {
  type: 'avatar' | 'poi' | 'session';
  data: any;
  timestamp: Date;
}

export class WebSocketClient {
  private ws: WebSocket | null = null;
  private url: string;
  private sessionId: string;
  private connectionStatus: ConnectionStatus = ConnectionStatus.DISCONNECTED;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000; // Start with 1 second
  private reconnectTimer: number | null = null;
  private messageQueue: WebSocketMessage[] = [];
  private statusChangeCallbacks: ((status: ConnectionStatus) => void)[] = [];
  private messageCallbacks: ((message: WebSocketMessage) => void)[] = [];
  private errorCallbacks: ((error: WebSocketError) => void)[] = [];
  private stateSyncCallbacks: ((data: StateSync) => void)[] = [];

  constructor(url: string, sessionId: string) {
    this.url = url;
    this.sessionId = sessionId;
  }

  // Connection Management
  async connect(): Promise<void> {
    if (this.connectionStatus === ConnectionStatus.CONNECTED || this.connectionStatus === ConnectionStatus.CONNECTING) {
      return;
    }

    this.connectionStatus = ConnectionStatus.CONNECTING;
    this.notifyStatusChange();

    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.url);

        this.ws.onopen = () => {
          console.log('🔌 WebSocket: Connected successfully to', this.url);
          this.connectionStatus = ConnectionStatus.CONNECTED;
          this.reconnectAttempts = 0;
          this.reconnectDelay = 1000; // Reset delay
          this.notifyStatusChange();
          this.sendQueuedMessages();
          resolve();
        };

        this.ws.onclose = (event) => {
          this.connectionStatus = ConnectionStatus.DISCONNECTED;
          this.ws = null;
          this.notifyStatusChange();

          // Auto-reconnect on unexpected closure
          if (event.code !== 1000 && event.code !== 1001) {
            this.scheduleReconnect();
          }
        };

        this.ws.onerror = (error) => {
          this.connectionStatus = ConnectionStatus.DISCONNECTED;
          const wsError: WebSocketError = {
            message: 'WebSocket connection error',
            timestamp: new Date()
          };
          this.notifyError(wsError);
          reject(error);
        };

        this.ws.onmessage = (event) => {
          this.handleMessage(event);
        };

      } catch (error) {
        this.connectionStatus = ConnectionStatus.DISCONNECTED;
        const wsError: WebSocketError = {
          message: error instanceof Error ? error.message : 'Unknown connection error',
          timestamp: new Date()
        };
        this.notifyError(wsError);
        reject(error);
      }
    });
  }

  disconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    if (this.ws) {
      this.ws.close(1000, 'Normal closure');
      this.ws = null;
    }

    this.connectionStatus = ConnectionStatus.DISCONNECTED;
    this.notifyStatusChange();
  }

  isConnected(): boolean {
    return this.connectionStatus === ConnectionStatus.CONNECTED;
  }

  isConnecting(): boolean {
    return this.connectionStatus === ConnectionStatus.CONNECTING;
  }

  isReconnecting(): boolean {
    return this.connectionStatus === ConnectionStatus.RECONNECTING;
  }

  getConnectionStatus(): ConnectionStatus {
    return this.connectionStatus;
  }

  getQueuedMessageCount(): number {
    return this.messageQueue.length;
  }

  // Callback registration methods
  onStatusChange(callback: (status: ConnectionStatus) => void): void {
    this.statusChangeCallbacks.push(callback);
  }

  onMessage(callback: (message: WebSocketMessage) => void): void {
    this.messageCallbacks.push(callback);
  }

  onError(callback: (error: WebSocketError) => void): void {
    this.errorCallbacks.push(callback);
  }

  onStateSync(callback: (data: StateSync) => void): void {
    this.stateSyncCallbacks.push(callback);
  }

  // Helper methods for notifications
  private notifyStatusChange(): void {
    this.statusChangeCallbacks.forEach(callback => {
      try {
        callback(this.connectionStatus);
      } catch (error) {
        console.error('Error in status change callback:', error);
      }
    });
  }

  private notifyMessage(message: WebSocketMessage): void {
    this.messageCallbacks.forEach(callback => {
      try {
        callback(message);
      } catch (error) {
        console.error('Error in message callback:', error);
      }
    });
  }

  private notifyError(error: WebSocketError): void {
    this.errorCallbacks.forEach(callback => {
      try {
        callback(error);
      } catch (error) {
        console.error('Error in error callback:', error);
      }
    });
  }

  private notifyStateSync(data: StateSync): void {
    this.stateSyncCallbacks.forEach(callback => {
      try {
        callback(data);
      } catch (error) {
        console.error('Error in state sync callback:', error);
      }
    });
  }

  // Reconnection Logic
  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      return;
    }

    this.connectionStatus = ConnectionStatus.RECONNECTING;
    this.reconnectAttempts++;
    this.notifyStatusChange();

    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1); // Exponential backoff

    this.reconnectTimer = setTimeout(() => {
      this.connect().catch(() => {
        // Connection failed, will trigger another reconnect attempt
      });
    }, delay);
  }

  setMaxReconnectAttempts(max: number): void {
    this.maxReconnectAttempts = max;
  }

  getReconnectAttempts(): number {
    return this.reconnectAttempts;
  }

  // Message Handling
  send(message: WebSocketMessage): void {
    console.log('📤 WebSocket: Sending message', {
      type: message.type,
      data: message.data,
      sessionId: this.sessionId,
      connected: this.isConnected(),
      queued: !this.isConnected()
    });

    if (this.isConnected() && this.ws) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.log('📋 WebSocket: Queueing message (not connected)', message.type);
      // Queue message for later
      this.messageQueue.push(message);
    }
  }

  private sendQueuedMessages(): void {
    while (this.messageQueue.length > 0 && this.isConnected() && this.ws) {
      const message = this.messageQueue.shift()!;
      this.ws.send(JSON.stringify(message));
    }
  }

  getQueuedMessages(): WebSocketMessage[] {
    return [...this.messageQueue];
  }

  private handleMessage(event: MessageEvent): void {
    try {
      const message: WebSocketMessage = JSON.parse(event.data);
      if (!message.timestamp) {
        message.timestamp = new Date();
      }

      console.log('📨 WebSocket: Raw message received', {
        type: message.type,
        data: message.data,
        sessionId: this.sessionId,
        timestamp: message.timestamp
      });

      this.notifyMessage(message);
      this.processMessage(message);
    } catch (error) {
      console.error('❌ WebSocket: Failed to parse message', error, event.data);
      this.notifyError({
        message: 'Failed to parse WebSocket message',
        timestamp: new Date()
      });
    }
  }

  private processMessage(message: WebSocketMessage): void {
    console.log('🔄 WebSocket: Processing message', {
      type: message.type,
      data: message.data,
      sessionId: this.sessionId,
      isOwnSession: message.data?.sessionId === this.sessionId
    });

    switch (message.type) {
      case 'welcome':
        this.handleWelcome(message.data);
        break;
      case 'error':
        this.handleError(message.data);
        break;
      case 'avatar_update':
        this.handleAvatarUpdate(message.data);
        break;
      case 'avatar_move_ack':
        this.handleAvatarMoveAck(message.data);
        break;
      case 'avatar_move_confirmed':
        this.handleAvatarMoveConfirmed(message.data);
        break;
      case 'avatar_move_rejected':
        this.handleAvatarMoveRejected(message.data);
        break;
      case 'avatar_moved':
        this.handleAvatarMoved(message.data);
        break;
      case 'user_joined':
        this.handleUserJoined(message.data);
        break;
      case 'user_left':
        this.handleUserLeft(message.data);
        break;
      case 'initial_users':
        this.handleInitialUsers(message.data);
        break;
      case 'poi_update':
        this.handlePOIUpdate(message.data);
        break;
      case 'poi_delete':
        this.handlePOIDelete(message.data);
        break;
      case 'poi_create_confirmed':
        this.handlePOICreateConfirmed(message.data);
        break;
      case 'poi_create_rejected':
        this.handlePOICreateRejected(message.data);
        break;
      case 'call_request':
        this.handleCallRequest(message.data);
        break;
      case 'call_accept':
        this.handleCallAccept(message.data);
        break;
      case 'call_reject':
        this.handleCallReject(message.data);
        break;
      case 'call_end':
        this.handleCallEnd(message.data);
        break;
      case 'webrtc_offer':
        this.handleWebRTCOffer(message.data);
        break;
      case 'webrtc_answer':
        this.handleWebRTCAnswer(message.data);
        break;
      case 'ice_candidate':
        this.handleICECandidate(message.data);
        break;
      case 'user_call_status':
        this.handleUserCallStatus(message.data);
        break;
      default:
        console.log('❓ WebSocket: Unknown message type', message.type);
        break;
    }
  }

  // Store Integration Handlers
  private handleWelcome(data: any): void {
    console.log('🎉 WebSocket: Welcome message received', data);
    // Welcome message - connection established successfully
  }

  private handleError(data: any): void {
    console.error('❌ WebSocket: Server error', data);
    this.notifyError({
      message: data.message || 'Server error',
      timestamp: new Date()
    });
  }

  private handleAvatarMoveAck(data: any): void {
    console.log('✅ WebSocket: Avatar move acknowledged', data);
    // Avatar move was acknowledged by server
    if (data.sessionId === this.sessionId) {
      sessionStore.getState().confirmAvatarPosition(data.position);
    }
  }

  private handleAvatarUpdate(data: any): void {
    if (data.sessionId === this.sessionId) {
      sessionStore.getState().confirmAvatarPosition(data.position);
      this.notifyStateSync({
        type: 'avatar',
        data: data.position,
        timestamp: new Date()
      });
    }
  }

  private handleAvatarMoveConfirmed(data: any): void {
    if (data.sessionId === this.sessionId) {
      sessionStore.getState().confirmAvatarPosition(data.position);
      this.notifyStateSync({
        type: 'avatar',
        data: data.position,
        timestamp: new Date()
      });
    }
  }

  private handleAvatarMoveRejected(data: any): void {
    if (data.sessionId === this.sessionId) {
      sessionStore.getState().rollbackAvatarPosition();
      this.notifyStateSync({
        type: 'avatar',
        data: { rejected: true, reason: data.reason },
        timestamp: new Date()
      });
    }
  }

  private handlePOIUpdate(data: POIData): void {
    poiStore.getState().handleRealtimeUpdate(data);
    this.notifyStateSync({
      type: 'poi',
      data,
      timestamp: new Date()
    });
  }

  private handlePOIDelete(data: { id: string }): void {
    poiStore.getState().handleRealtimeDelete(data.id);
    this.notifyStateSync({
      type: 'poi',
      data: { action: 'delete', id: data.id },
      timestamp: new Date()
    });
  }

  private handlePOICreateConfirmed(data: { tempId: string; poi: POIData }): void {
    // Remove the temporary POI and add the real one
    poiStore.getState().rollbackPOICreation(data.tempId);
    poiStore.getState().addPOI(data.poi);
    this.notifyStateSync({
      type: 'poi',
      data: { action: 'create_confirmed', poi: data.poi },
      timestamp: new Date()
    });
  }

  private handlePOICreateRejected(data: { tempId: string; reason: string }): void {
    poiStore.getState().rollbackPOICreation(data.tempId);
    this.notifyStateSync({
      type: 'poi',
      data: { action: 'create_rejected', tempId: data.tempId, reason: data.reason },
      timestamp: new Date()
    });
  }

  // Optimistic Update Methods
  moveAvatar(position: { lat: number; lng: number }): void {
    console.log('🚀 WebSocket: Sending avatar move', {
      sessionId: this.sessionId,
      position,
      connected: this.isConnected()
    });

    // Perform optimistic update
    sessionStore.getState().updateAvatarPosition(position, true);

    // Send to server
    this.send({
      type: 'avatar_move',
      data: {
        sessionId: this.sessionId,
        position
      },
      timestamp: new Date()
    });

    // Emit state sync
    this.notifyStateSync({
      type: 'avatar',
      data: { position },
      timestamp: new Date()
    });
  }

  createPOI(poi: POIData): void {
    // Perform optimistic update
    poiStore.getState().createPOIOptimistic(poi);

    // Send to server
    this.send({
      type: 'poi_create',
      data: {
        tempId: poi.id,
        poi: {
          ...poi,
          id: undefined // Server will assign real ID
        }
      },
      timestamp: new Date()
    });

    // Emit state sync
    this.notifyStateSync({
      type: 'poi',
      data: poi,
      timestamp: new Date()
    });
  }

  joinPOI(poiId: string): void {
    const userId = this.sessionId;

    // Perform optimistic update
    const success = poiStore.getState().joinPOIOptimistic(poiId, userId);

    if (success) {
      // Send to server
      this.send({
        type: 'poi_join',
        data: { poiId, userId },
        timestamp: new Date()
      });

      // Emit state sync
      this.notifyStateSync({
        type: 'poi',
        data: { action: 'join', poiId, userId },
        timestamp: new Date()
      });
    }
  }

  leavePOI(poiId: string): void {
    const userId = this.sessionId;

    // Perform optimistic update
    const success = poiStore.getState().leavePOI(poiId, userId);

    if (success) {
      // Send to server
      this.send({
        type: 'poi_leave',
        data: { poiId, userId },
        timestamp: new Date()
      });

      // Emit state sync
      this.notifyStateSync({
        type: 'poi',
        data: { action: 'leave', poiId, userId },
        timestamp: new Date()
      });
    }
  }

  // Auto-leave functionality for POI switching
  joinPOIWithAutoLeave(poiId: string): void {
    const userId = this.sessionId;

    // Perform optimistic update with auto-leave
    const success = poiStore.getState().joinPOIOptimisticWithAutoLeave(poiId, userId);

    if (success) {
      // Send to server
      this.send({
        type: 'poi_join',
        data: { poiId, userId },
        timestamp: new Date()
      });

      // Emit state sync
      this.notifyStateSync({
        type: 'poi',
        data: { action: 'join', poiId, userId },
        timestamp: new Date()
      });
    }
  }

  // Leave current POI (for map clicks)
  leaveCurrentPOI(): void {
    const currentPOI = poiStore.getState().getCurrentUserPOI();

    if (currentPOI) {
      this.leavePOI(currentPOI);
    }
  }

  // Multi-User Avatar Handlers
  private handleAvatarMoved(data: any): void {
    console.log('🏃 WebSocket: Received avatar_moved', {
      data,
      mySessionId: this.sessionId,
      isOwnMovement: data.sessionId === this.sessionId
    });

    // Handle other users' avatar movements
    if (data.sessionId !== this.sessionId) {
      console.log('👤 WebSocket: Updating other user avatar position', {
        sessionId: data.sessionId,
        userId: data.userId,
        position: data.position,
        mySessionId: this.sessionId
      });

      avatarStore.getState().updateAvatarPosition(
        data.sessionId,
        data.position,
        true // Mark as moving
      );

      this.notifyStateSync({
        type: 'avatar',
        data: { sessionId: data.sessionId, position: data.position },
        timestamp: new Date()
      });
    } else {
      console.log('🚫 WebSocket: Ignoring own avatar movement', {
        sessionId: data.sessionId,
        mySessionId: this.sessionId
      });
    }
  }

  private handleUserJoined(data: any): void {
    console.log('👋 WebSocket: Received user_joined', data);
    // Handle new user joining the map
    if (data.sessionId !== this.sessionId) {
      const avatarData: AvatarData = {
        sessionId: data.sessionId,
        userId: data.userId,
        displayName: data.displayName || data.userId,
        avatarURL: data.avatarURL,
        aboutMe: data.aboutMe,
        position: data.position || { lat: 0, lng: 0 },
        isCurrentUser: false,
        isMoving: false,
        role: data.role || 'user'
      };

      console.log('➕ WebSocket: Adding new user avatar', avatarData);
      avatarStore.getState().addOrUpdateAvatar(avatarData);

      this.notifyStateSync({
        type: 'avatar',
        data: { action: 'joined', avatar: avatarData },
        timestamp: new Date()
      });
    } else {
      console.log('🚫 WebSocket: Ignoring own user joined event');
    }
  }

  private handleUserLeft(data: any): void {
    // Handle user leaving the map
    if (data.sessionId !== this.sessionId) {
      avatarStore.getState().removeAvatar(data.sessionId);

      this.notifyStateSync({
        type: 'avatar',
        data: { action: 'left', sessionId: data.sessionId },
        timestamp: new Date()
      });
    }
  }

  private handleInitialUsers(data: any): void {
    console.log('📋 WebSocket: Received initial_users', data);
    // Handle initial user state when joining a map
    if (data.users && Array.isArray(data.users)) {
      const otherUsers: AvatarData[] = data.users
        .filter((user: any) => user.sessionId !== this.sessionId)
        .map((user: any) => ({
          sessionId: user.sessionId,
          userId: user.userId,
          displayName: user.displayName || user.userId,
          avatarURL: user.avatarURL,
          aboutMe: user.aboutMe,
          position: user.position || { lat: 0, lng: 0 },
          isCurrentUser: false,
          isMoving: false,
          role: user.role || 'user'
        }));

      console.log('📥 WebSocket: Loading initial users', otherUsers);
      avatarStore.getState().loadInitialUsers(otherUsers);

      this.notifyStateSync({
        type: 'avatar',
        data: { action: 'initial_load', users: otherUsers },
        timestamp: new Date()
      });
    }
  }

  // Request initial users when connecting
  requestInitialUsers(): void {
    console.log('📋 WebSocket: Requesting initial users');
    this.send({
      type: 'request_initial_users',
      data: { mapId: 'default-map' }, // TODO: Use actual map ID
      timestamp: new Date()
    });
  }

  // Video Call Methods
  sendCallRequest(targetUserId: string, callId: string, callerName?: string): void {
    console.log('📞 WebSocket: Sending call request', { targetUserId, callId });
    this.send({
      type: 'call_request',
      data: {
        targetUserId,
        callId,
        callerName
      },
      timestamp: new Date()
    });
  }

  sendCallAccept(callId: string, callerUserId: string): void {
    console.log('✅ WebSocket: Sending call accept', { callId, callerUserId });
    this.send({
      type: 'call_accept',
      data: {
        callId,
        callerUserId
      },
      timestamp: new Date()
    });
  }

  sendCallReject(callId: string, callerUserId: string): void {
    console.log('❌ WebSocket: Sending call reject', { callId, callerUserId });
    this.send({
      type: 'call_reject',
      data: {
        callId,
        callerUserId
      },
      timestamp: new Date()
    });
  }

  sendCallEnd(callId: string, otherUserId: string): void {
    console.log('📵 WebSocket: Sending call end', { callId, otherUserId });
    this.send({
      type: 'call_end',
      data: {
        callId,
        otherUserId
      },
      timestamp: new Date()
    });
  }

  // WebRTC Signaling Methods
  sendWebRTCOffer(callId: string, targetUserId: string, sdp: RTCSessionDescriptionInit): void {
    console.log('📝 WebSocket: Sending WebRTC offer', { callId, targetUserId });
    this.send({
      type: 'webrtc_offer',
      data: {
        callId,
        targetUserId,
        sdp
      },
      timestamp: new Date()
    });
  }

  sendWebRTCAnswer(callId: string, targetUserId: string, sdp: RTCSessionDescriptionInit): void {
    console.log('📋 WebSocket: Sending WebRTC answer', { callId, targetUserId });
    this.send({
      type: 'webrtc_answer',
      data: {
        callId,
        targetUserId,
        sdp
      },
      timestamp: new Date()
    });
  }

  sendICECandidate(callId: string, targetUserId: string, candidate: RTCIceCandidate): void {
    console.log('🧊 WebSocket: Sending ICE candidate', { callId, targetUserId });
    this.send({
      type: 'ice_candidate',
      data: {
        callId,
        targetUserId,
        candidate: {
          candidate: candidate.candidate,
          sdpMLineIndex: candidate.sdpMLineIndex,
          sdpMid: candidate.sdpMid
        }
      },
      timestamp: new Date()
    });
  }

  // Video Call Message Handlers
  private handleCallRequest(data: any): void {
    console.log('📞 WebSocket: Received call request', data);

    // Import stores dynamically to avoid circular dependencies
    import('../stores/videoCallStore').then(({ videoCallStore }) => {
      import('../stores/avatarStore').then(({ avatarStore }) => {
        const { callId, callerInfo } = data;

        if (callerInfo && callId) {
          // Mark caller as in call
          avatarStore.getState().updateAvatarCallStatus(callerInfo.userId, true);

          videoCallStore.getState().receiveCall(
            callId,
            callerInfo.userId,
            callerInfo.displayName || callerInfo.sessionId,
            callerInfo.avatarURL
          );
        }
      });
    });
  }

  private handleCallAccept(data: any): void {
    console.log('✅ WebSocket: Received call accept', data);

    import('../stores/videoCallStore').then(({ videoCallStore }) => {
      const { callId } = data;
      const state = videoCallStore.getState();

      // Only process if this is our current call and we're the caller
      if (state.currentCall?.callId === callId && state.currentCall && !state.currentCall.isIncoming) {
        console.log('📝 Call accepted - creating WebRTC offer as caller');
        state.setCallState('connecting');

        // Create and send WebRTC offer
        if (state.webrtcService) {
          state.webrtcService.createOffer().then((offer) => {
            console.log('📝 WebRTC offer created, sending to callee');
            const wsClient = (window as any).wsClient;
            if (wsClient && wsClient.isConnected() && state.currentCall) {
              wsClient.sendWebRTCOffer(state.currentCall.callId, state.currentCall.targetUserId, offer);
            }
          }).catch((error) => {
            console.error('❌ Failed to create WebRTC offer:', error);
            state.endCall();
          });
        }
      }
    });
  }

  private handleCallReject(data: any): void {
    console.log('❌ WebSocket: Received call reject', data);

    import('../stores/videoCallStore').then(({ videoCallStore }) => {
      import('../stores/avatarStore').then(({ avatarStore }) => {
        const { callId } = data;
        const state = videoCallStore.getState();

        // Only process if this is our current call
        if (state.currentCall?.callId === callId && state.currentCall) {
          // Mark target user as no longer in call
          avatarStore.getState().updateAvatarCallStatus(state.currentCall.targetUserId, false);

          // Don't call rejectCall() as it would send another message
          // Just update the state to ended and clear
          state.setCallState('ended');

          // Auto-clear after showing "ended" state briefly
          setTimeout(() => {
            state.clearCall();
          }, 2000);
        }
      });
    });
  }

  private handleCallEnd(data: any): void {
    console.log('📵 WebSocket: Received call end', data);

    import('../stores/videoCallStore').then(({ videoCallStore }) => {
      import('../stores/avatarStore').then(({ avatarStore }) => {
        const { callId } = data;
        const state = videoCallStore.getState();

        // Only process if this is our current call
        if (state.currentCall?.callId === callId && state.currentCall) {
          // Mark target user as no longer in call
          avatarStore.getState().updateAvatarCallStatus(state.currentCall.targetUserId, false);

          // Don't call endCall() as it would send another message
          // Just update the state to ended and clear
          state.setCallState('ended');

          // Auto-clear after showing "ended" state briefly
          setTimeout(() => {
            state.clearCall();
          }, 2000);
        }
      });
    });
  }

  // WebRTC Signaling Message Handlers
  private handleWebRTCOffer(data: any): void {
    console.log('📝 WebSocket: Received WebRTC offer', data);

    import('../stores/videoCallStore').then(({ videoCallStore }) => {
      const { callId, fromUserId, sdp } = data;
      const state = videoCallStore.getState();

      if (state.currentCall?.callId === callId && state.webrtcService) {
        console.log('📝 Processing WebRTC offer for current call');

        state.webrtcService.setRemoteDescription(sdp).then(() => {
          console.log('✅ Remote description set, creating answer...');
          return state.webrtcService!.createAnswer();
        }).then((answer) => {
          console.log('✅ Answer created, sending to caller...');
          const wsClient = (window as any).wsClient;
          if (wsClient && wsClient.isConnected()) {
            wsClient.sendWebRTCAnswer(callId, fromUserId, answer);
            console.log('📤 WebRTC answer sent');
          } else {
            console.error('❌ WebSocket not connected, cannot send answer');
          }
        }).catch((error) => {
          console.error('❌ Failed to handle WebRTC offer:', error);
          state.endCall();
        });
      } else {
        console.warn('⚠️ Received WebRTC offer but no matching call or WebRTC service');
      }
    });
  }

  private handleWebRTCAnswer(data: any): void {
    console.log('📋 WebSocket: Received WebRTC answer', data);

    import('../stores/videoCallStore').then(({ videoCallStore }) => {
      const { callId, sdp } = data;
      const state = videoCallStore.getState();

      if (state.currentCall?.callId === callId && state.webrtcService) {
        console.log('📋 Processing WebRTC answer for current call');

        state.webrtcService.setRemoteDescription(sdp).then(() => {
          console.log('✅ Remote description (answer) set successfully');
        }).catch((error) => {
          console.error('❌ Failed to handle WebRTC answer:', error);
          state.endCall();
        });
      } else {
        console.warn('⚠️ Received WebRTC answer but no matching call or WebRTC service');
      }
    });
  }

  private handleICECandidate(data: any): void {
    console.log('🧊 WebSocket: Received ICE candidate', data);

    import('../stores/videoCallStore').then(({ videoCallStore }) => {
      const { callId, candidate } = data;
      const state = videoCallStore.getState();

      if (state.currentCall?.callId === callId && state.webrtcService) {
        console.log('🧊 Processing ICE candidate for current call');

        state.webrtcService.addIceCandidate(candidate).catch((error) => {
          console.error('❌ Failed to handle ICE candidate:', error);
        });
      }
    });
  }

  // User Call Status Handler
  private handleUserCallStatus(data: any): void {
    console.log('📞 WebSocket: Received user call status', data);

    import('../stores/avatarStore').then(({ avatarStore }) => {
      const { userId, isInCall } = data;

      if (userId) {
        avatarStore.getState().updateAvatarCallStatus(userId, isInCall);
        console.log(`📞 Updated call status for user ${userId}: ${isInCall ? 'in call' : 'available'}`);
      }
    });
  }
}
