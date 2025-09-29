export class SessionService {
  private heartbeatInterval: number | null = null;
  private readonly HEARTBEAT_INTERVAL = 5 * 60 * 1000; // 5 minutes

  constructor(private sessionId: string) {}

  startHeartbeat(): void {
    if (this.heartbeatInterval) {
      return; // Already started
    }

    console.log('🫀 SessionService: Starting heartbeat for session', this.sessionId);
    
    // Send initial heartbeat
    this.sendHeartbeat();
    
    // Set up interval
    this.heartbeatInterval = window.setInterval(() => {
      this.sendHeartbeat();
    }, this.HEARTBEAT_INTERVAL);
  }

  stopHeartbeat(): void {
    if (this.heartbeatInterval) {
      console.log('🫀 SessionService: Stopping heartbeat for session', this.sessionId);
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  private async sendHeartbeat(): Promise<void> {
    try {
      console.log('🫀 SessionService: Sending heartbeat for session', this.sessionId);
      
      const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
      console.log('🔧 SessionService environment check:', {
        VITE_API_BASE_URL: import.meta.env.VITE_API_BASE_URL,
        API_BASE_URL,
        heartbeatUrl: `${API_BASE_URL}/api/sessions/${this.sessionId}/heartbeat`
      });
      const response = await fetch(`${API_BASE_URL}/api/sessions/${this.sessionId}/heartbeat`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        console.error('❌ SessionService: Heartbeat failed', response.status, response.statusText);
        return;
      }

      console.log('✅ SessionService: Heartbeat successful for session', this.sessionId);
      
      // Update local heartbeat timestamp
      const { sessionStore } = await import('../stores/sessionStore');
      sessionStore.getState().updateHeartbeat();
      
    } catch (error) {
      console.error('❌ SessionService: Heartbeat error', error);
    }
  }

  updateSessionId(newSessionId: string): void {
    this.sessionId = newSessionId;
  }
}