/**
 * Simple Event Bus for Group Video Call Coordination
 * 
 * Used to decouple video call initialization from WebSocket events
 * and eliminate race conditions when users join during WebRTC setup.
 */

type EventCallback = (data?: any) => void;

class EventBus {
  private listeners = new Map<string, Set<EventCallback>>();
  private debug = true; // Enable logging for debugging

  /**
   * Subscribe to an event
   * @returns Unsubscribe function
   */
  on(event: string, callback: EventCallback): () => void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set());
    }
    this.listeners.get(event)!.add(callback);

    // Return unsubscribe function
    return () => this.off(event, callback);
  }

  /**
   * Unsubscribe from an event
   */
  off(event: string, callback: EventCallback): void {
    this.listeners.get(event)?.delete(callback);
  }

  /**
   * Emit an event to all listeners
   */
  emit(event: string, data?: any): void {
    if (this.debug) {
      console.log(`📢 Event: ${event}`, data);
    }

    const listeners = this.listeners.get(event);
    if (!listeners || listeners.size === 0) {
      if (this.debug) {
        console.log(`⚠️ No listeners for event: ${event}`);
      }
      return;
    }

    // Call each listener with error handling
    listeners.forEach((callback) => {
      try {
        callback(data);
      } catch (error) {
        console.error(`❌ Error in ${event} listener:`, error);
      }
    });
  }

  /**
   * Remove all listeners (useful for cleanup/testing)
   */
  clear(): void {
    this.listeners.clear();
  }

  /**
   * Get listener count for an event (useful for debugging)
   */
  listenerCount(event: string): number {
    return this.listeners.get(event)?.size || 0;
  }
}

// Singleton instance
export const eventBus = new EventBus();

/**
 * Event constants for type safety and autocomplete
 */
export const GroupCallEvents = {
  // User joined a POI (relevant for video calls)
  USER_JOINED_POI: 'group_call:user_joined_poi',
  
  // WebRTC service is initialized and ready to accept peers
  WEBRTC_READY: 'group_call:webrtc_ready',
  
  // Peer connection was successfully added (optional, for logging)
  PEER_ADDED: 'group_call:peer_added',
} as const;

/**
 * Type definitions for event data
 */
export interface UserJoinedPOIEvent {
  poiId: string;
  userId: string;
  displayName: string;
  avatarURL?: string;
  participants: any[];
}

export interface WebRTCReadyEvent {
  poiId: string;
}

export interface PeerAddedEvent {
  userId: string;
  poiId: string;
}
