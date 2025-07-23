/**
 * Simple logout broadcast service for multi-tab synchronization.
 * Uses BroadcastChannel API to notify other tabs when user logs out.
 */

const AUTH_BROADCAST_CHANNEL_NAME = 'nebraska-auth-sync';

interface AuthMessage {
  type: 'LOGOUT';
}

class LogoutBroadcastService {
  private channel: BroadcastChannel | null = null;
  private listeners = new Set<() => void>();

  constructor() {
    if (typeof BroadcastChannel !== 'undefined') {
      try {
        this.channel = new BroadcastChannel(AUTH_BROADCAST_CHANNEL_NAME);
        this.channel.onmessage = (event: MessageEvent<AuthMessage>) => {
          if (event.data?.type === 'LOGOUT') {
            this.notifyListeners();
          }
        };
      } catch (error) {
        console.warn('BroadcastChannel not available:', error);
      }
    }

    if (typeof window !== 'undefined') {
      window.addEventListener('unload', () => this.destroy());
    }
  }

  broadcastLogout(): void {
    const message: AuthMessage = { type: 'LOGOUT' };
    this.channel?.postMessage(message);
  }

  onLogout(callback: () => void): () => void {
    this.listeners.add(callback);
    return () => this.listeners.delete(callback);
  }

  private notifyListeners(): void {
    this.listeners.forEach(callback => {
      try {
        callback();
      } catch (error) {
        console.error('Error in logout listener:', error);
      }
    });
  }

  destroy(): void {
    this.channel?.close();
    this.listeners.clear();
  }
}

export const authBroadcast = new LogoutBroadcastService();
