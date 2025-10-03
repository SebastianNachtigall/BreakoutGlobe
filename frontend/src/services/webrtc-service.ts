export interface MediaConstraints {
  video: boolean;
  audio: boolean;
}

export interface WebRTCConfig extends RTCConfiguration {
  iceServers: RTCIceServer[];
}

export interface PeerConnectionCallbacks {
  onLocalStream?: (stream: MediaStream) => void;
  onRemoteStream?: (stream: MediaStream) => void;
  onIceCandidate?: (candidate: RTCIceCandidate) => void;
  onConnectionStateChange?: (state: RTCPeerConnectionState) => void;
  onError?: (error: Error) => void;
}

export class WebRTCService {
  private peerConnection: RTCPeerConnection | null = null;
  private localStream: MediaStream | null = null;
  private callbacks: PeerConnectionCallbacks = {};
  
  protected readonly defaultConfig: WebRTCConfig = {
    iceServers: [
      // STUN servers for NAT discovery
      { urls: 'stun:stun.l.google.com:19302' },
      { urls: 'stun:stun1.l.google.com:19302' },
      { urls: 'stun:stun2.l.google.com:19302' },
      { urls: 'stun:stun.cloudflare.com:3478' },
      // Free TURN servers for traffic relay (when STUN isn't enough)
      { 
        urls: 'turn:openrelay.metered.ca:80',
        username: 'openrelayproject',
        credential: 'openrelayproject'
      },
      { 
        urls: 'turn:openrelay.metered.ca:443',
        username: 'openrelayproject', 
        credential: 'openrelayproject'
      },
      {
        urls: 'turn:openrelay.metered.ca:443?transport=tcp',
        username: 'openrelayproject',
        credential: 'openrelayproject'
      }
    ],
    // ICE configuration for better connectivity
    iceTransportPolicy: 'all',
    bundlePolicy: 'max-bundle',
    rtcpMuxPolicy: 'require'
  };

  constructor(config?: Partial<WebRTCConfig>) {
    const finalConfig = { ...this.defaultConfig, ...config };
    this.initializePeerConnection(finalConfig);
  }

  protected initializePeerConnection(config: WebRTCConfig): void {
    try {
      this.peerConnection = new RTCPeerConnection(config);
      this.setupPeerConnectionHandlers();
      console.log('🔗 WebRTC: Peer connection initialized');
    } catch (error) {
      console.error('❌ WebRTC: Failed to initialize peer connection:', error);
      this.callbacks.onError?.(error as Error);
    }
  }

  private setupPeerConnectionHandlers(): void {
    if (!this.peerConnection) return;

    // Handle ICE candidates
    this.peerConnection.onicecandidate = (event) => {
      if (event.candidate) {
        const candidate = event.candidate;
        const candidateType = candidate.candidate.includes('typ relay') ? 'TURN' : 
                             candidate.candidate.includes('typ srflx') ? 'STUN' : 
                             candidate.candidate.includes('typ host') ? 'HOST' : 'UNKNOWN';
        console.log(`🧊 WebRTC: ICE candidate generated (${candidateType}):`, candidate.candidate.substring(0, 50) + '...');
        this.callbacks.onIceCandidate?.(event.candidate);
      }
    };

    // Handle remote stream
    this.peerConnection.ontrack = (event) => {
      console.log('📺 WebRTC: Remote stream received');
      const [remoteStream] = event.streams;
      this.callbacks.onRemoteStream?.(remoteStream);
    };

    // Handle connection state changes
    this.peerConnection.onconnectionstatechange = () => {
      const state = this.peerConnection?.connectionState;
      console.log('🔄 WebRTC: Connection state changed to:', state);
      
      if (state === 'connected') {
        console.log('🎉 WebRTC: Peer connection fully established');
      } else if (state === 'disconnected') {
        console.warn('⚠️ WebRTC: Peer connection disconnected');
      } else if (state === 'failed') {
        console.error('❌ WebRTC: Peer connection failed');
      }
      
      if (state) {
        this.callbacks.onConnectionStateChange?.(state);
      }
    };

    // Handle ICE connection state changes with retry logic
    this.peerConnection.oniceconnectionstatechange = () => {
      const state = this.peerConnection?.iceConnectionState;
      console.log('🧊 WebRTC: ICE connection state changed to:', state);
      
      if (state === 'failed') {
        console.error('❌ WebRTC: ICE connection failed');
        this.callbacks.onError?.(new Error('ICE connection failed'));
      } else if (state === 'disconnected') {
        console.warn('⚠️ WebRTC: ICE connection disconnected - attempting to reconnect...');
        // Don't immediately fail on disconnected - give it more time to reconnect with TURN
        setTimeout(() => {
          const currentState = this.peerConnection?.iceConnectionState;
          if (currentState === 'disconnected' || currentState === 'failed') {
            console.error('❌ WebRTC: ICE connection failed after extended retry timeout');
            this.callbacks.onError?.(new Error('ICE connection failed after retry'));
          } else if (currentState === 'connected' || currentState === 'completed') {
            console.log('✅ WebRTC: ICE connection recovered successfully');
          }
        }, 10000); // Wait 10 seconds before giving up (TURN servers need more time)
      } else if (state === 'connected') {
        console.log('✅ WebRTC: ICE connection established');
      } else if (state === 'completed') {
        console.log('🎉 WebRTC: ICE connection completed (optimal path found)');
      }
    };

    // Handle ICE gathering state changes
    this.peerConnection.onicegatheringstatechange = () => {
      const state = this.peerConnection?.iceGatheringState;
      console.log('🧊 WebRTC: ICE gathering state changed to:', state);
    };
  }

  public setCallbacks(callbacks: PeerConnectionCallbacks): void {
    this.callbacks = { ...this.callbacks, ...callbacks };
  }

  public async initializeLocalMedia(constraints: MediaConstraints = { video: true, audio: true }): Promise<MediaStream> {
    try {
      console.log('🎥 WebRTC: Requesting local media access...', constraints);
      
      const stream = await navigator.mediaDevices.getUserMedia(constraints);
      this.localStream = stream;
      
      console.log('✅ WebRTC: Local media stream obtained');
      this.callbacks.onLocalStream?.(stream);
      
      // Add local stream tracks to peer connection
      if (this.peerConnection) {
        stream.getTracks().forEach(track => {
          this.peerConnection!.addTrack(track, stream);
        });
        console.log('📤 WebRTC: Local tracks added to peer connection');
      }
      
      return stream;
    } catch (error) {
      console.error('❌ WebRTC: Failed to get local media:', error);
      this.callbacks.onError?.(error as Error);
      throw error;
    }
  }

  public async createOffer(): Promise<RTCSessionDescriptionInit> {
    if (!this.peerConnection) {
      throw new Error('Peer connection not initialized');
    }

    try {
      console.log('📝 WebRTC: Creating offer...');
      const offer = await this.peerConnection.createOffer();
      await this.peerConnection.setLocalDescription(offer);
      console.log('✅ WebRTC: Offer created and set as local description');
      return offer;
    } catch (error) {
      console.error('❌ WebRTC: Failed to create offer:', error);
      this.callbacks.onError?.(error as Error);
      throw error;
    }
  }

  public async createAnswer(): Promise<RTCSessionDescriptionInit> {
    if (!this.peerConnection) {
      throw new Error('Peer connection not initialized');
    }

    try {
      console.log('📝 WebRTC: Creating answer...');
      const answer = await this.peerConnection.createAnswer();
      await this.peerConnection.setLocalDescription(answer);
      console.log('✅ WebRTC: Answer created and set as local description');
      return answer;
    } catch (error) {
      console.error('❌ WebRTC: Failed to create answer:', error);
      this.callbacks.onError?.(error as Error);
      throw error;
    }
  }

  public async setRemoteDescription(description: RTCSessionDescriptionInit): Promise<void> {
    if (!this.peerConnection) {
      throw new Error('Peer connection not initialized');
    }

    try {
      console.log('📥 WebRTC: Setting remote description...', description.type);
      await this.peerConnection.setRemoteDescription(description);
      console.log('✅ WebRTC: Remote description set successfully');
    } catch (error) {
      console.error('❌ WebRTC: Failed to set remote description:', error);
      this.callbacks.onError?.(error as Error);
      throw error;
    }
  }

  public async addIceCandidate(candidate: RTCIceCandidateInit): Promise<void> {
    if (!this.peerConnection) {
      throw new Error('Peer connection not initialized');
    }

    try {
      console.log('🧊 WebRTC: Adding ICE candidate...');
      await this.peerConnection.addIceCandidate(candidate);
      console.log('✅ WebRTC: ICE candidate added successfully');
    } catch (error) {
      console.error('❌ WebRTC: Failed to add ICE candidate:', error);
      this.callbacks.onError?.(error as Error);
      throw error;
    }
  }

  public toggleAudio(enabled?: boolean): boolean {
    if (!this.localStream) {
      console.warn('⚠️ WebRTC: No local stream available for audio toggle');
      return false;
    }

    const audioTracks = this.localStream.getAudioTracks();
    if (audioTracks.length === 0) {
      console.warn('⚠️ WebRTC: No audio tracks available');
      return false;
    }

    const newState = enabled !== undefined ? enabled : !audioTracks[0].enabled;
    audioTracks.forEach(track => {
      track.enabled = newState;
    });

    console.log(`🎤 WebRTC: Audio ${newState ? 'enabled' : 'disabled'}`);
    return newState;
  }

  public toggleVideo(enabled?: boolean): boolean {
    if (!this.localStream) {
      console.warn('⚠️ WebRTC: No local stream available for video toggle');
      return false;
    }

    const videoTracks = this.localStream.getVideoTracks();
    if (videoTracks.length === 0) {
      console.warn('⚠️ WebRTC: No video tracks available');
      return false;
    }

    const newState = enabled !== undefined ? enabled : !videoTracks[0].enabled;
    videoTracks.forEach(track => {
      track.enabled = newState;
    });

    console.log(`📹 WebRTC: Video ${newState ? 'enabled' : 'disabled'}`);
    return newState;
  }

  public getConnectionState(): RTCPeerConnectionState | null {
    return this.peerConnection?.connectionState || null;
  }

  public getLocalStream(): MediaStream | null {
    return this.localStream;
  }

  public cleanup(): void {
    console.log('🧹 WebRTC: Cleaning up resources...');

    // Stop local media tracks
    if (this.localStream) {
      this.localStream.getTracks().forEach(track => {
        console.log(`🛑 Stopping ${track.kind} track`);
        track.stop();
      });
      this.localStream = null;
    }

    // Properly close peer connection with all event handlers
    if (this.peerConnection) {
      // Remove all event listeners to prevent callbacks during cleanup
      this.peerConnection.onicecandidate = null;
      this.peerConnection.ontrack = null;
      this.peerConnection.onconnectionstatechange = null;
      this.peerConnection.oniceconnectionstatechange = null;
      this.peerConnection.onicegatheringstatechange = null;
      
      // Close the connection
      this.peerConnection.close();
      this.peerConnection = null;
      console.log('🔌 Peer connection closed and nullified');
    }

    // Clear callbacks
    this.callbacks = {};

    console.log('✅ WebRTC: Cleanup completed');
  }
}

export interface GroupPeerConnectionCallbacks extends PeerConnectionCallbacks {
  onRemoteStreamForUser?: (userId: string, stream: MediaStream) => void;
  onPeerConnectionStateChange?: (userId: string, state: RTCPeerConnectionState) => void;
}

export class GroupWebRTCService extends WebRTCService {
  public peerConnections: Map<string, RTCPeerConnection> = new Map();
  public remoteStreams: Map<string, MediaStream> = new Map();
  private groupCallbacks: GroupPeerConnectionCallbacks = {};

  constructor(config?: Partial<WebRTCConfig>) {
    super(config);
  }

  // Override to prevent parent class from creating a main peer connection
  protected initializePeerConnection(config: WebRTCConfig): void {
    // Don't create the main peer connection for group calls
    // We'll create individual peer connections as needed
    console.log('🔗 WebRTC: Group service initialized (no main peer connection)');
  }

  public setCallbacks(callbacks: GroupPeerConnectionCallbacks): void {
    super.setCallbacks(callbacks);
    this.groupCallbacks = { ...this.groupCallbacks, ...callbacks };
  }

  public setWebSocketClient(wsClient: any): void {
    this.wsClient = wsClient;
  }

  private wsClient: any = null;
  private currentUserId: string | null = null;

  public setCurrentUserId(userId: string): void {
    this.currentUserId = userId;
  }

  private getCurrentUserId(): string | null {
    return this.currentUserId;
  }

  public async addPeer(userId: string): Promise<void> {
    if (this.peerConnections.has(userId)) {
      console.warn(`⚠️ WebRTC: Peer connection already exists for user: ${userId}`);
      return;
    }

    console.log(`🔗 WebRTC: Adding peer connection for user: ${userId}`);
    
    try {
      const peerConnection = new RTCPeerConnection(this.defaultConfig);
      this.setupGroupPeerConnectionHandlers(peerConnection, userId);
      
      // Add local stream tracks if available
      const localStream = this.getLocalStream();
      if (localStream) {
        localStream.getTracks().forEach(track => {
          peerConnection.addTrack(track, localStream);
        });
        console.log(`📤 WebRTC: Local tracks added to peer connection for user: ${userId}`);
      }
      
      this.peerConnections.set(userId, peerConnection);
      console.log(`✅ WebRTC: Peer connection created for user: ${userId}`);
      
      // Only create offer if current user should be the "caller" (lexicographically smaller ID)
      // This prevents both peers from creating offers simultaneously
      const currentUserId = this.getCurrentUserId();
      if (currentUserId && currentUserId < userId) {
        console.log(`📝 WebRTC: Current user (${currentUserId}) should initiate call to user (${userId})`);
        this.createOfferForPeer(userId).then((offer) => {
          console.log(`📝 WebRTC: Offer created for user: ${userId}, sending via WebSocket`);
          
          // Send offer via WebSocket for POI group calls
          if (this.wsClient && this.wsClient.isConnected()) {
            import('../stores/videoCallStore').then(({ videoCallStore }) => {
              const currentPOI = videoCallStore.getState().currentPOI;
              if (currentPOI) {
                this.wsClient.sendPOICallOffer(currentPOI, userId, offer);
              }
            });
          }
        }).catch((error) => {
          console.error(`❌ WebRTC: Failed to create offer for user ${userId}:`, error);
        });
      } else {
        console.log(`📞 WebRTC: Waiting for offer from user (${userId}) as they should initiate the call`);
      }
      
    } catch (error) {
      console.error(`❌ WebRTC: Failed to create peer connection for user ${userId}:`, error);
      this.groupCallbacks.onError?.(error as Error);
      throw error;
    }
  }

  public removePeer(userId: string): void {
    console.log(`🗑️ WebRTC: Removing peer connection for user: ${userId}`);
    
    const peerConnection = this.peerConnections.get(userId);
    if (peerConnection) {
      // Clean up event handlers
      peerConnection.onicecandidate = null;
      peerConnection.ontrack = null;
      peerConnection.onconnectionstatechange = null;
      peerConnection.oniceconnectionstatechange = null;
      peerConnection.onicegatheringstatechange = null;
      
      // Close connection
      peerConnection.close();
      this.peerConnections.delete(userId);
    }
    
    // Remove remote stream
    this.remoteStreams.delete(userId);
    
    console.log(`✅ WebRTC: Peer connection removed for user: ${userId}`);
  }

  public async createOfferForPeer(userId: string): Promise<RTCSessionDescriptionInit> {
    const peerConnection = this.peerConnections.get(userId);
    if (!peerConnection) {
      throw new Error(`Peer connection not found for user: ${userId}`);
    }

    try {
      console.log(`📝 WebRTC: Creating offer for user: ${userId}`);
      const offer = await peerConnection.createOffer();
      await peerConnection.setLocalDescription(offer);
      console.log(`✅ WebRTC: Offer created for user: ${userId}`);
      return offer;
    } catch (error) {
      console.error(`❌ WebRTC: Failed to create offer for user ${userId}:`, error);
      this.groupCallbacks.onError?.(error as Error);
      throw error;
    }
  }

  public async handleAnswerFromPeer(userId: string, answer: RTCSessionDescriptionInit): Promise<void> {
    const peerConnection = this.peerConnections.get(userId);
    if (!peerConnection) {
      throw new Error(`Peer connection not found for user: ${userId}`);
    }

    try {
      console.log(`📥 WebRTC: Setting remote description (answer) for user: ${userId}`);
      await peerConnection.setRemoteDescription(answer);
      console.log(`✅ WebRTC: Answer processed for user: ${userId}`);
    } catch (error) {
      console.error(`❌ WebRTC: Failed to handle answer for user ${userId}:`, error);
      this.groupCallbacks.onError?.(error as Error);
      throw error;
    }
  }

  private setupGroupPeerConnectionHandlers(peerConnection: RTCPeerConnection, userId: string): void {
    // Handle ICE candidates
    peerConnection.onicecandidate = (event) => {
      if (event.candidate) {
        const candidate = event.candidate;
        const candidateType = candidate.candidate.includes('typ relay') ? 'TURN' : 
                             candidate.candidate.includes('typ srflx') ? 'STUN' : 
                             candidate.candidate.includes('typ host') ? 'HOST' : 'UNKNOWN';
        console.log(`🧊 WebRTC: ICE candidate generated for user ${userId} (${candidateType}):`, candidate.candidate.substring(0, 50) + '...');
        
        // Send ICE candidate via WebSocket for POI group calls
        if (this.wsClient && this.wsClient.isConnected()) {
          // Get current POI from video call store
          import('../stores/videoCallStore').then(({ videoCallStore }) => {
            const currentPOI = videoCallStore.getState().currentPOI;
            if (currentPOI) {
              this.wsClient.sendPOICallICECandidate(currentPOI, userId, event.candidate);
            }
          });
        }
        
        this.groupCallbacks.onIceCandidate?.(event.candidate);
      }
    };

    // Handle remote stream
    peerConnection.ontrack = (event) => {
      console.log(`📺 WebRTC: Remote stream received from user: ${userId}`);
      const [remoteStream] = event.streams;
      this.remoteStreams.set(userId, remoteStream);
      this.groupCallbacks.onRemoteStreamForUser?.(userId, remoteStream);
    };

    // Handle connection state changes
    peerConnection.onconnectionstatechange = () => {
      const state = peerConnection.connectionState;
      console.log(`🔄 WebRTC: Connection state changed for user ${userId}:`, state);
      
      if (state === 'connected') {
        console.log(`🎉 WebRTC: Peer connection established with user: ${userId}`);
      } else if (state === 'disconnected') {
        console.warn(`⚠️ WebRTC: Peer connection disconnected from user: ${userId}`);
      } else if (state === 'failed') {
        console.error(`❌ WebRTC: Peer connection failed with user: ${userId}`);
        this.groupCallbacks.onError?.(new Error(`Peer connection failed with user: ${userId}`));
      }
      
      this.groupCallbacks.onPeerConnectionStateChange?.(userId, state);
    };

    // Handle ICE connection state changes
    peerConnection.oniceconnectionstatechange = () => {
      const state = peerConnection.iceConnectionState;
      console.log(`🧊 WebRTC: ICE connection state changed for user ${userId}:`, state);
      
      if (state === 'failed') {
        console.error(`❌ WebRTC: ICE connection failed with user: ${userId}`);
        this.groupCallbacks.onError?.(new Error(`ICE connection failed with user: ${userId}`));
      } else if (state === 'connected') {
        console.log(`✅ WebRTC: ICE connection established with user: ${userId}`);
      } else if (state === 'completed') {
        console.log(`🎉 WebRTC: ICE connection completed with user: ${userId}`);
      }
    };

    // Handle ICE gathering state changes
    peerConnection.onicegatheringstatechange = () => {
      const state = peerConnection.iceGatheringState;
      console.log(`🧊 WebRTC: ICE gathering state changed for user ${userId}:`, state);
    };
  }

  public override cleanup(): void {
    console.log('🧹 WebRTC: Cleaning up group call resources...');

    // Clean up all peer connections
    for (const [userId, peerConnection] of this.peerConnections) {
      console.log(`🛑 Closing peer connection for user: ${userId}`);
      
      // Remove event listeners
      peerConnection.onicecandidate = null;
      peerConnection.ontrack = null;
      peerConnection.onconnectionstatechange = null;
      peerConnection.oniceconnectionstatechange = null;
      peerConnection.onicegatheringstatechange = null;
      
      // Close connection
      peerConnection.close();
    }

    // Clear collections
    this.peerConnections.clear();
    this.remoteStreams.clear();

    // Clear callbacks
    this.groupCallbacks = {};

    // Call parent cleanup for local stream and main peer connection
    super.cleanup();

    console.log('✅ WebRTC: Group call cleanup completed');
  }
}