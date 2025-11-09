import { Injectable, UnauthorizedException } from '@nestjs/common';
import { HttpService } from '@nestjs/axios';
import { firstValueFrom } from 'rxjs';

export interface ValidationResult {
  valid: boolean;
  claims?: any;
  error?: string;
}

export interface DiscoveryDocument {
  issuer: string;
  authorization_endpoint: string;
  token_endpoint: string;
  userinfo_endpoint: string;
  jwks_uri: string;
  [key: string]: any;
}

@Injectable()
export class OidcService {
  constructor(private readonly httpService: HttpService) {}

  /**
   * Decode a JWT token without validation
   */
  decodeJWT(token: string): any {
    try {
      const parts = token.split('.');
      if (parts.length !== 3) {
        throw new Error('Invalid JWT format');
      }

      const payload = parts[1];
      const decoded = Buffer.from(payload, 'base64url').toString('utf-8');
      return JSON.parse(decoded);
    } catch {
      throw new UnauthorizedException('Failed to decode JWT');
    }
  }

  /**
   * Validate an ID token according to OIDC specification
   */
  validateIDToken(
    idToken: string,
    clientId: string,
    issuer: string,
    nonce?: string,
  ): ValidationResult {
    try {
      const claims = this.decodeJWT(idToken);

      // Verify issuer
      if (claims.iss !== issuer) {
        return {
          valid: false,
          error: `Invalid issuer. Expected ${issuer}, got ${claims.iss}`,
        };
      }

      // Verify audience
      const audience = Array.isArray(claims.aud) ? claims.aud : [claims.aud];
      if (!audience.includes(clientId)) {
        return {
          valid: false,
          error: `Invalid audience. Expected ${clientId}`,
        };
      }

      // Verify expiration
      const now = Math.floor(Date.now() / 1000);
      if (claims.exp && claims.exp <= now) {
        return {
          valid: false,
          error: 'Token has expired',
        };
      }

      // Verify issued at (with 60 second clock skew tolerance)
      if (claims.iat && claims.iat > now + 60) {
        return {
          valid: false,
          error: 'Token issued in the future',
        };
      }

      // Verify nonce if provided
      if (nonce && claims.nonce !== nonce) {
        return {
          valid: false,
          error: 'Invalid nonce',
        };
      }

      return {
        valid: true,
        claims,
      };
    } catch (error) {
      return {
        valid: false,
        error: error instanceof Error ? error.message : 'Token validation failed',
      };
    }
  }

  /**
   * Fetch OIDC discovery document from the authorization server
   */
  async fetchDiscovery(issuer: string): Promise<DiscoveryDocument> {
    try {
      const discoveryUrl = `${issuer}/.well-known/openid-configuration`;
      const response = await firstValueFrom(this.httpService.get<DiscoveryDocument>(discoveryUrl));
      return response.data;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      throw new Error(`Failed to fetch discovery document: ${message}`);
    }
  }
}
