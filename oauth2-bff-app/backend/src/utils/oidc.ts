import axios from 'axios';
import crypto from 'crypto';

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

/**
 * Fetch OIDC Discovery Document
 */
export async function fetchDiscovery(issuer: string): Promise<DiscoveryDocument> {
  const discoveryUrl = `${issuer}/.well-known/openid-configuration`;
  const response = await axios.get<DiscoveryDocument>(discoveryUrl);
  return response.data;
}

/**
 * Fetch JWKS (JSON Web Key Set)
 */
export async function fetchJWKS(jwksUri: string): Promise<JWKS> {
  const response = await axios.get<JWKS>(jwksUri);
  return response.data;
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
