import { Injectable } from '@nestjs/common';
import * as crypto from 'crypto';

@Injectable()
export class CryptoService {
  /**
   * Generate a cryptographically secure random state for OAuth2 flow
   */
  generateState(): string {
    return this.base64URLEncode(crypto.randomBytes(32));
  }

  /**
   * Generate a cryptographically secure random nonce for OIDC
   */
  generateNonce(): string {
    return crypto.randomBytes(32).toString('base64url');
  }

  /**
   * Base64 URL encode a buffer
   */
  private base64URLEncode(buffer: Buffer): string {
    return buffer.toString('base64').replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
  }
}
