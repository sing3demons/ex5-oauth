import { Injectable } from '@nestjs/common';
import * as crypto from 'crypto';

const ALGORITHM = 'aes-256-gcm';
const IV_LENGTH = 16;
const AUTH_TAG_LENGTH = 16;

export interface OAuthStateData {
  state: string;
  nonce: string;
  codeVerifier: string;
  redirectUri: string;
  timestamp: number;
}

@Injectable()
export class CryptoService {
  private getEncryptionKey(): Buffer {
    const secret = process.env.SESSION_SECRET || 'change-this-secret-in-production';
    return crypto.createHash('sha256').update(secret).digest();
  }

  encrypt(data: object): string {
    const key = this.getEncryptionKey();
    const iv = crypto.randomBytes(IV_LENGTH);

    const cipher = crypto.createCipheriv(ALGORITHM, key, iv);

    const jsonData = JSON.stringify(data);
    let encrypted = cipher.update(jsonData, 'utf8', 'hex');
    encrypted += cipher.final('hex');

    const authTag = cipher.getAuthTag();

    const combined = Buffer.concat([
      iv,
      authTag,
      Buffer.from(encrypted, 'hex'),
    ]);

    return combined
      .toString('base64')
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=/g, '');
  }

  decrypt(encryptedData: string): object {
    const key = this.getEncryptionKey();

    let base64 = encryptedData.replace(/-/g, '+').replace(/_/g, '/');

    while (base64.length % 4) {
      base64 += '=';
    }

    const combined = Buffer.from(base64, 'base64');

    const iv = combined.subarray(0, IV_LENGTH);
    const authTag = combined.subarray(IV_LENGTH, IV_LENGTH + AUTH_TAG_LENGTH);
    const encrypted = combined.subarray(IV_LENGTH + AUTH_TAG_LENGTH);

    const decipher = crypto.createDecipheriv(ALGORITHM, key, iv);
    decipher.setAuthTag(authTag);

    let decrypted = decipher.update(encrypted.toString('hex'), 'hex', 'utf8');
    decrypted += decipher.final('utf8');

    return JSON.parse(decrypted);
  }

  encryptOAuthState(
    data: Omit<OAuthStateData, 'timestamp'>,
  ): string {
    const stateData: OAuthStateData = {
      ...data,
      timestamp: Date.now(),
    };
    return this.encrypt(stateData);
  }

  decryptOAuthState(encryptedState: string): OAuthStateData | null {
    try {
      const data = this.decrypt(encryptedState) as OAuthStateData;

      const now = Date.now();
      const expiry = 10 * 60 * 1000; // 10 minutes

      if (now - data.timestamp > expiry) {
        console.error('OAuth state expired');
        return null;
      }

      return data;
    } catch (error) {
      console.error('Failed to decrypt OAuth state:', error);
      return null;
    }
  }

  /**
   * Generate random state parameter
   */
  generateState(): string {
    return crypto.randomBytes(32).toString('base64url');
  }

  /**
   * Generate random nonce
   */
  generateNonce(): string {
    return crypto.randomBytes(32).toString('base64url');
  }

  /**
   * Generate PKCE code verifier
   */
  generateCodeVerifier(): string {
    return crypto.randomBytes(32).toString('base64url');
  }

  /**
   * Generate PKCE code challenge from verifier
   */
  generateCodeChallenge(verifier: string): string {
    return crypto
      .createHash('sha256')
      .update(verifier)
      .digest('base64url');
  }
}
