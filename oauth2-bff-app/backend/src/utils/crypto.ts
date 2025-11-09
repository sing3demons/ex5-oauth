import crypto from 'crypto';

const ALGORITHM = 'aes-256-gcm';
const KEY_LENGTH = 32;
const IV_LENGTH = 16;
const AUTH_TAG_LENGTH = 16;

/**
 * Get encryption key from environment or generate one
 */
function getEncryptionKey(): Buffer {
  const secret = process.env.SESSION_SECRET || 'change-this-secret-in-production';
  // Derive a 32-byte key from the secret
  return crypto.createHash('sha256').update(secret).digest();
}

/**
 * Encrypt data and return base64url encoded string
 */
export function encrypt(data: object): string {
  const key = getEncryptionKey();
  const iv = crypto.randomBytes(IV_LENGTH);
  
  const cipher = crypto.createCipheriv(ALGORITHM, key, iv);
  
  const jsonData = JSON.stringify(data);
  let encrypted = cipher.update(jsonData, 'utf8', 'hex');
  encrypted += cipher.final('hex');
  
  const authTag = cipher.getAuthTag();
  
  // Combine iv + authTag + encrypted data
  const combined = Buffer.concat([
    iv,
    authTag,
    Buffer.from(encrypted, 'hex')
  ]);
  
  // Return base64url encoded (URL-safe)
  return combined.toString('base64')
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
}

/**
 * Decrypt base64url encoded string and return data
 */
export function decrypt(encryptedData: string): object {
  const key = getEncryptionKey();
  
  // Convert base64url to base64
  let base64 = encryptedData
    .replace(/-/g, '+')
    .replace(/_/g, '/');
  
  // Add padding if needed
  while (base64.length % 4) {
    base64 += '=';
  }
  
  const combined = Buffer.from(base64, 'base64');
  
  // Extract iv, authTag, and encrypted data
  const iv = combined.subarray(0, IV_LENGTH);
  const authTag = combined.subarray(IV_LENGTH, IV_LENGTH + AUTH_TAG_LENGTH);
  const encrypted = combined.subarray(IV_LENGTH + AUTH_TAG_LENGTH);
  
  const decipher = crypto.createDecipheriv(ALGORITHM, key, iv);
  decipher.setAuthTag(authTag);
  
  let decrypted = decipher.update(encrypted.toString('hex'), 'hex', 'utf8');
  decrypted += decipher.final('utf8');
  
  return JSON.parse(decrypted);
}

/**
 * Encrypt OAuth state data
 */
export interface OAuthStateData {
  state: string;
  nonce: string;
  codeVerifier: string;
  redirectUri: string;
  timestamp: number;
}

export function encryptOAuthState(data: Omit<OAuthStateData, 'timestamp'>): string {
  const stateData: OAuthStateData = {
    ...data,
    timestamp: Date.now()
  };
  return encrypt(stateData);
}

/**
 * Decrypt and validate OAuth state data
 */
export function decryptOAuthState(encryptedState: string): OAuthStateData | null {
  try {
    const data = decrypt(encryptedState) as OAuthStateData;
    
    // Validate timestamp (10 minutes expiry)
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
