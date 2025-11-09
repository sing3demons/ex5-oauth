import { Injectable } from '@nestjs/common';
import { OidcService } from './oidc.service';

@Injectable()
export class TokenService {
  constructor(private readonly oidcService: OidcService) {}

  /**
   * Extract user ID from JWT token
   * Looks for 'sub' or 'user_id' claim
   */
  getUserIdFromToken(authHeader: string): string | null {
    try {
      if (!authHeader || !authHeader.startsWith('Bearer ')) {
        return null;
      }

      const token = authHeader.replace('Bearer ', '');
      const claims = this.oidcService.decodeJWT(token);

      return claims.sub || claims.user_id || null;
    } catch {
      return null;
    }
  }
}
