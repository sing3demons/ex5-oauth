import { Request, Response, NextFunction } from 'express';

/**
 * Middleware to check if access token exists in request
 */
export function requireAuth(req: Request, res: Response, next: NextFunction) {
  const authHeader = req.headers.authorization;
  
  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return res.status(401).json({
      error: 'unauthorized',
      message: 'No access token provided'
    });
  }
  
  // Token validation will be done by OAuth2 server
  // BFF just forwards the token
  next();
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
