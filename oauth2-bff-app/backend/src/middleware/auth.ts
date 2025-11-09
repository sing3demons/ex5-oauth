import { Request, Response, NextFunction } from 'express';
import { verifyJWT, decodeJWT, extractUserInfo } from '../utils/oidc';
import config from '../config';

// Extend Express Request to include user
declare global {
  namespace Express {
    interface Request {
      user?: {
        id: string;
        email?: string;
        name?: string;
        picture?: string;
      };
    }
  }
}

/**
 * Middleware to validate access token and extract user information
 */
export async function requireAuth(req: Request, res: Response, next: NextFunction) {
  const authHeader = req.headers.authorization;
  
  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return res.status(401).json({
      error: 'unauthorized',
      message: 'No access token provided'
    });
  }
  
  const token = authHeader.substring(7); // Remove 'Bearer ' prefix
  
  try {
    // First try to verify with JWKS (if available)
    const jwksUri = `${config.OAUTH2_SERVER}/.well-known/jwks.json`;
    
    const verification = await verifyJWT(token, jwksUri, {
      issuer: config.OAUTH2_SERVER,
      algorithms: ['RS256']
    });
    
    if (verification.valid && verification.claims) {
      // Extract user information from claims
      req.user = extractUserInfo(verification.claims);
      return next();
    }
    
    // If JWKS verification fails, try decoding (for development)
    if (process.env.NODE_ENV === 'development') {
      console.warn('JWT verification failed, falling back to decode:', verification.error);
      const claims = decodeJWT(token);
      
      // Check expiration
      if (claims.exp && claims.exp < Math.floor(Date.now() / 1000)) {
        return res.status(401).json({
          error: 'token_expired',
          message: 'Access token has expired'
        });
      }
      
      req.user = extractUserInfo(claims);
      return next();
    }
    
    return res.status(401).json({
      error: 'invalid_token',
      message: verification.error || 'Token validation failed'
    });
  } catch (error: any) {
    console.error('Token validation error:', error);
    return res.status(401).json({
      error: 'invalid_token',
      message: 'Failed to validate access token'
    });
  }
}

/**
 * Middleware to check if refresh token cookie exists
 */
export function requireRefreshToken(req: Request, res: Response, next: NextFunction) {
  const refreshToken = req.cookies.refresh_token;
  
  if (!refreshToken) {
    return res.status(401).json({
      error: 'unauthorized',
      message: 'No refresh token found'
    });
  }
  
  next();
}
