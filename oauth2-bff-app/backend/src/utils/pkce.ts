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

/**
 * Generate PKCE code verifier
 */
export function generateCodeVerifier(): string {
  return base64URLEncode(crypto.randomBytes(32));
}

/**
 * Generate PKCE code challenge from verifier
 */
export function generateCodeChallenge(verifier: string): string {
  const hash = crypto.createHash('sha256').update(verifier).digest();
  return base64URLEncode(hash);
}

/**
 * Validate state parameter
 */
export function validateState(state: string, expectedState: string): boolean {
  return state === expectedState;
}
