import crypto from 'crypto';

/**
 * Base64 URL encode (without padding)
 */
function base64URLEncode(buffer: Buffer): string {
  return buffer
    .toString('base64')
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
}

/**
 * Generate random state for CSRF protection
 */
export function generateState(): string {
  return base64URLEncode(crypto.randomBytes(32));
}
