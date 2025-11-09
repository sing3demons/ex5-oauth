import axios from 'axios';
import crypto from 'crypto';
import jwt from 'jsonwebtoken';

interface DiscoveryDocument {
  issuer: string;
  authorization_endpoint: string;
  token_endpoint: string;
  userinfo_endpoint: string;
  jwks_uri: string;
  response_types_supported: string[];
  subject_types_supported: string[];
  id_token_signing_alg_values_supported: string[];
  scopes_supported: string[];
  token_endpoint_auth_methods_supported: string[];
  claims_supported: string[];
}

interface JWK {
  kty: string;
  use?: string;
  kid?: string;
  n?: string;
  e?: string;
  alg?: string;
}

interface JWKS {
  keys: JWK[];
}

// Cache for JWKS
let jwksCache: { keys: JWK[]; timestamp: number } | null = null;
const JWKS_CACHE_TTL = 3600000; // 1 hour

/**
 * Fetch OIDC Discovery Document
 */
export async function fetchDiscovery(issuer: string): Promise<DiscoveryDocument> {
  const discoveryUrl = `${issuer}/.well-known/openid-configuration`;
  const response = await axios.get<DiscoveryDocument>(discoveryUrl);
  return response.data;
}

/**
 * Fetch JWKS (JSON Web Key Set) with caching
 */
export async function fetchJWKS(jwksUri: string): Promise<JWKS> {
  const now = Date.now();
  
  // Return cached JWKS if still valid
  if (jwksCache && (now - jwksCache.timestamp) < JWKS_CACHE_TTL) {
    return { keys: jwksCache.keys };
  }
  
  // Fetch fresh JWKS
  const response = await axios.get<JWKS>(jwksUri);
  jwksCache = {
    keys: response.data.keys,
    timestamp: now
  };
  
  return response.data;
}

/**
 * Convert JWK to PEM format for RSA public key
 */
function jwkToPem(jwk: JWK): string {
  if (jwk.kty !== 'RSA' || !jwk.n || !jwk.e) {
    throw new Error('Only RSA keys are supported');
  }
  
  // Convert base64url to buffer
  const modulus = Buffer.from(jwk.n, 'base64url');
  const exponent = Buffer.from(jwk.e, 'base64url');
  
  // Create PEM format
  const modulusHex = modulus.toString('hex');
  const exponentHex = exponent.toString('hex');
  
  // Build ASN.1 structure for RSA public key
  const modulusLength = modulus.length;
  const exponentLength = exponent.length;
  
  // Simple PEM construction (for production, use a proper library)
  const key = crypto.createPublicKey({
    key: {
      kty: 'RSA',
      n: jwk.n,
      e: jwk.e
    },
    format: 'jwk'
  });
  
  return key.export({ type: 'spki', format: 'pem' }) as string;
}

/**
 * Get public key from JWKS by kid
 */
export async function getPublicKey(jwksUri: string, kid?: string): Promise<string> {
  const jwks = await fetchJWKS(jwksUri);
  
  let jwk: JWK | undefined;
  
  if (kid) {
    jwk = jwks.keys.find(key => key.kid === kid);
  } else {
    // Use first key if no kid specified
    jwk = jwks.keys[0];
  }
  
  if (!jwk) {
    throw new Error('No matching key found in JWKS');
  }
  
  return jwkToPem(jwk);
}

/**
 * Decode JWT without verification (for inspection)
 */
export function decodeJWT(token: string): any {
  const parts = token.split('.');
  if (parts.length !== 3) {
    throw new Error('Invalid JWT format');
  }
  
  const payload = Buffer.from(parts[1], 'base64').toString('utf-8');
  return JSON.parse(payload);
}

/**
 * Validate ID Token claims
 */
export function validateIDToken(
  idToken: string,
  clientId: string,
  issuer: string,
  nonce?: string
): { valid: boolean; error?: string; claims?: any } {
  try {
    const claims = decodeJWT(idToken);
    
    // Validate issuer
    if (claims.iss !== issuer) {
      return { valid: false, error: 'Invalid issuer' };
    }
    
    // Validate audience
    if (claims.aud !== clientId && !claims.aud?.includes(clientId)) {
      return { valid: false, error: 'Invalid audience' };
    }
    
    // Validate expiration
    const now = Math.floor(Date.now() / 1000);
    if (claims.exp && claims.exp < now) {
      return { valid: false, error: 'Token expired' };
    }
    
    // Validate issued at
    if (claims.iat && claims.iat > now + 60) {
      return { valid: false, error: 'Token issued in the future' };
    }
    
    // Validate nonce if provided
    if (nonce && claims.nonce !== nonce) {
      return { valid: false, error: 'Invalid nonce' };
    }
    
    return { valid: true, claims };
  } catch (error: any) {
    return { valid: false, error: error.message };
  }
}

/**
 * Generate nonce for OIDC
 */
export function generateNonce(): string {
  return crypto.randomBytes(32).toString('base64url');
}

/**
 * Verify JWT token with JWKS
 */
export async function verifyJWT(
  token: string,
  jwksUri: string,
  options?: {
    issuer?: string;
    audience?: string;
    algorithms?: string[];
  }
): Promise<{ valid: boolean; error?: string; claims?: any }> {
  try {
    // Decode header to get kid
    const parts = token.split('.');
    if (parts.length !== 3) {
      return { valid: false, error: 'Invalid JWT format' };
    }
    
    const header = JSON.parse(Buffer.from(parts[0], 'base64url').toString());
    const kid = header.kid;
    
    // Get public key from JWKS
    const publicKey = await getPublicKey(jwksUri, kid);
    
    // Verify token
    const verifyOptions: jwt.VerifyOptions = {
      algorithms: (options?.algorithms || ['RS256']) as jwt.Algorithm[],
    };
    
    if (options?.issuer) {
      verifyOptions.issuer = options.issuer;
    }
    
    if (options?.audience) {
      verifyOptions.audience = options.audience;
    }
    
    const claims = jwt.verify(token, publicKey, verifyOptions);
    
    return { valid: true, claims };
  } catch (error: any) {
    return { valid: false, error: error.message };
  }
}

/**
 * Validate access token by calling token introspection endpoint
 */
export async function introspectToken(
  token: string,
  introspectionEndpoint: string,
  clientId: string,
  clientSecret: string
): Promise<{ active: boolean; [key: string]: any }> {
  try {
    const response = await axios.post(
      introspectionEndpoint,
      new URLSearchParams({
        token,
        client_id: clientId,
        client_secret: clientSecret
      }),
      {
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded'
        }
      }
    );
    
    return response.data;
  } catch (error) {
    return { active: false };
  }
}

/**
 * Extract user information from token claims
 */
export function extractUserInfo(claims: any): {
  id: string;
  email?: string;
  name?: string;
  picture?: string;
} {
  return {
    id: claims.sub || claims.user_id,
    email: claims.email,
    name: claims.name,
    picture: claims.picture
  };
}
