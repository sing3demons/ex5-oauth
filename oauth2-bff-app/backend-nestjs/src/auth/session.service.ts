import { Injectable, Logger, OnModuleInit, OnModuleDestroy } from '@nestjs/common';

export interface SessionData {
  redirect_uri: string;
  nonce: string;
  timestamp: number;
}

@Injectable()
export class SessionService implements OnModuleInit, OnModuleDestroy {
  private readonly logger = new Logger(SessionService.name);
  private sessions: Map<string, SessionData> = new Map();
  private cleanupInterval?: NodeJS.Timeout;
  private readonly SESSION_TTL = 10 * 60 * 1000; // 10 minutes in milliseconds
  private readonly CLEANUP_INTERVAL = 60 * 1000; // 60 seconds in milliseconds

  onModuleInit() {
    this.logger.log('Starting session cleanup interval');
    this.cleanupInterval = setInterval(() => {
      this.cleanupExpiredSessions();
    }, this.CLEANUP_INTERVAL);
  }

  onModuleDestroy() {
    this.logger.log('Stopping session cleanup interval');
    if (this.cleanupInterval) {
      clearInterval(this.cleanupInterval);
    }
  }

  /**
   * Create a new session with state, nonce, and timestamp
   * Requirements: 17.1
   */
  createSession(state: string, data: Omit<SessionData, 'timestamp'>): void {
    const sessionData: SessionData = {
      ...data,
      timestamp: Date.now(),
    };

    this.sessions.set(state, sessionData);
    this.logger.debug(`Session created for state: ${state.substring(0, 8)}...`);
  }

  /**
   * Retrieve session by state
   * Requirements: 17.2
   */
  getSession(state: string): SessionData | undefined {
    const session = this.sessions.get(state);

    if (!session) {
      this.logger.debug(`Session not found for state: ${state.substring(0, 8)}...`);
      return undefined;
    }

    // Check if session is expired
    const now = Date.now();
    if (now - session.timestamp > this.SESSION_TTL) {
      this.logger.debug(`Session expired for state: ${state.substring(0, 8)}...`);
      this.sessions.delete(state);
      return undefined;
    }

    return session;
  }

  /**
   * Delete session after use
   * Requirements: 17.3, 17.4
   */
  deleteSession(state: string): void {
    const deleted = this.sessions.delete(state);
    if (deleted) {
      this.logger.debug(`Session deleted for state: ${state.substring(0, 8)}...`);
    }
  }

  /**
   * Clean up expired sessions
   * Requirements: 17.5
   */
  cleanupExpiredSessions(): void {
    const now = Date.now();
    let cleanedCount = 0;

    for (const [state, session] of this.sessions.entries()) {
      if (now - session.timestamp > this.SESSION_TTL) {
        this.sessions.delete(state);
        cleanedCount++;
      }
    }

    if (cleanedCount > 0) {
      this.logger.log(`Cleaned up ${cleanedCount} expired session(s)`);
    }
  }

  /**
   * Get current session count (for monitoring/debugging)
   */
  getSessionCount(): number {
    return this.sessions.size;
  }
}
