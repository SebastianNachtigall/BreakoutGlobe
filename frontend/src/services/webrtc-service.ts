export interface MediaConstraints {
  video: boolean;
  audio: boolean;
}

export interface WebRTCConfig {
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
  
  private readonly defaultConfig: WebRTCConfig = {
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
    iceCandidatePoolSize: 10,
    iceTransportPolicy: 'all',
    bundlePolicy: 'max-bundle',
    rtcpMuxPolicy: 'require'
  };

  constructor(config?: Partial<WebRTCConfig>) {
    const finalConfig = { ...this.defaultConfig, ...config };
    this.initializePeerConnection(finalConfig);
  }

  private initializePeerConnection(config: WebRTCConfig): void {
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