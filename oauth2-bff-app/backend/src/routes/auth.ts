import express, { Request, Response } from 'express';
import axios from 'axios';
import { generateState } from '../utils/pkce';
import { generateNonce, validateIDToken, decodeJWT } from '../utils/oidc';
import { requireRefreshToken } from '../middleware/auth';
import { TokenResponse, UserInfo } from '../types';
import config from '../config';

const router = express.Router();


// In-memory session store (use Redis in production)
const sessions = new Map<string, any>();

/**
 * GET /auth/login
 * Initiate OIDC authorization flow (Confidential Client)
 */
router.get('/login', (req: Request, res: Response) => {
  try {
    const state = generateState();
    const nonce = generateNonce();
    
    // Store state and nonce in session for validation
    sessions.set(state, {
      redirect_uri: config.REDIRECT_URI,
      nonce,
      timestamp: Date.now()
    });
    
    // Build authorization URL with OIDC parameters
    const authUrl = new URL(`${config.OAUTH2_SERVER}/oauth/authorize`);
    authUrl.searchParams.set('response_type', 'code');
    authUrl.searchParams.set('client_id', config.CLIENT_ID);
    authUrl.searchParams.set('redirect_uri', config.REDIRECT_URI);
    authUrl.searchParams.set('scope', 'openid profile email');
    authUrl.searchParams.set('state', state);
    authUrl.searchParams.set('nonce', nonce);
    authUrl.searchParams.set('response_mode', 'query');
    
    res.json({
      authorization_url: authUrl.toString()
    });
  } catch (error) {
    console.error('Login error:', error);
    res.status(500).json({
      error: 'server_error',
      message: 'Failed to initiate login'
    });
  }
});

/**
 * GET /auth/callback
 * Handle OAuth2 callback and exchange code for tokens (Confidential Client)
 */
router.get('/callback', async (req: Request, res: Response) => {
  try {
    const { code, state, error } = req.query;
    
    if (error) {
      return res.redirect(`${config.FRONTEND_URL}?error=${error}`);
    }
    
    if (!code || !state) {
      return res.redirect(`${config.FRONTEND_URL}?error=invalid_request`);
    }
    
    // Retrieve session
    const session = sessions.get(state as string);
    if (!session) {
      return res.redirect(`${config.FRONTEND_URL}?error=invalid_state`);
    }
    
    // Clean up session
    sessions.delete(state as string);
    
    // Exchange authorization code for tokens (with client_secret)
    const tokenResponse = await axios.post<TokenResponse>(
      `${config.OAUTH2_SERVER}/oauth/token`,
      new URLSearchParams({
        grant_type: 'authorization_code',
        code: code as string,
        redirect_uri: session.redirect_uri,
        client_id: config.CLIENT_ID,
        client_secret: config.CLIENT_SECRET
      }),
      {
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded'
        }
      }
    );
    
    const tokens = tokenResponse.data;
    
    // Validate ID Token if present (OIDC)
    if (tokens.id_token) {
      const validation = validateIDToken(
        tokens.id_token,
        config.CLIENT_ID,
        config.OAUTH2_SERVER,
        session.nonce
      );
      
      if (!validation.valid) {
        console.error('ID Token validation failed:', validation.error);
        return res.redirect(`${config.FRONTEND_URL}?error=invalid_id_token`);
      }
      
      console.log('ID Token validated successfully:', validation.claims);
    }
    
    // Store refresh token in HttpOnly cookie
    if (tokens.refresh_token) {
      res.cookie('refresh_token', tokens.refresh_token, {
        httpOnly: true,
        secure: process.env.NODE_ENV === 'production',
        sameSite: 'lax',
        maxAge: 7 * 24 * 60 * 60 * 1000, // 7 days
        path: '/'
      });
    }
    
    // Redirect to frontend with access token
    const redirectUrl = new URL(config.FRONTEND_URL);
    redirectUrl.searchParams.set('access_token', tokens.access_token);
    redirectUrl.searchParams.set('expires_in', tokens.expires_in.toString());
    if (tokens.id_token) {
      redirectUrl.searchParams.set('id_token', tokens.id_token);
    }
    
    res.redirect(redirectUrl.toString());
  } catch (error: any) {
    console.error('Callback error:', error.response?.data || error.message);
    res.redirect(`${config.FRONTEND_URL}?error=token_exchange_failed`);
  }
});

/**
 * POST /auth/refresh
 * Refresh access token using refresh token from cookie (Confidential Client)
 */
router.post('/refresh', requireRefreshToken, async (req: Request, res: Response) => {
  try {
    const refreshToken = req.cookies.refresh_token;
    
    // Exchange refresh token for new access token (with client_secret)
    const tokenResponse = await axios.post<TokenResponse>(
      `${config.OAUTH2_SERVER}/oauth/token`,
      new URLSearchParams({
        grant_type: 'refresh_token',
        refresh_token: refreshToken,
        client_id: config.CLIENT_ID,
        client_secret: config.CLIENT_SECRET
      }),
      {
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded'
        }
      }
    );
    
    const tokens = tokenResponse.data;
    
    // Update refresh token cookie if rotated
    if (tokens.refresh_token) {
      res.cookie('refresh_token', tokens.refresh_token, {
        httpOnly: true,
        secure: process.env.NODE_ENV === 'production',
        sameSite: 'lax',
        maxAge: 7 * 24 * 60 * 60 * 1000,
        path: '/'
      });
    }
    
    res.json({
      access_token: tokens.access_token,
      expires_in: tokens.expires_in,
      token_type: tokens.token_type
    });
  } catch (error: any) {
    console.error('Refresh error:', error.response?.data || error.message);
    
    // Clear invalid refresh token
    res.clearCookie('refresh_token');
    
    res.status(401).json({
      error: 'invalid_grant',
      message: 'Refresh token expired or invalid'
    });
  }
});

/**
 * POST /auth/logout
 * Clear refresh token cookie
 */
router.post('/logout', (req: Request, res: Response) => {
  res.clearCookie('refresh_token', {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    path: '/'
  });
  
  res.json({
    message: 'Logged out successfully'
  });
});

/**
 * GET /auth/userinfo
 * Get user info from OAuth2 server
 */
router.get('/userinfo', async (req: Request, res: Response) => {
  try {
    const authHeader = req.headers.authorization;
    
    if (!authHeader) {
      return res.status(401).json({
        error: 'unauthorized',
        message: 'No access token provided'
      });
    }
    
    // Forward request to OAuth2 server
    const userInfoResponse = await axios.get<UserInfo>(
      `${config.OAUTH2_SERVER}/oauth/userinfo`,
      {
        headers: {
          Authorization: authHeader
        }
      }
    );
    
    res.json(userInfoResponse.data);
  } catch (error: any) {
    console.error('UserInfo error:', error.response?.data || error.message);
    res.status(error.response?.status || 500).json({
      error: 'server_error',
      message: 'Failed to get user info'
    });
  }
});

/**
 * GET /auth/discovery
 * Get OIDC discovery document from OAuth2 server
 */
router.get('/discovery', async (req: Request, res: Response) => {
  try {
    const response = await axios.get(
      `${config.OAUTH2_SERVER}/.well-known/openid-configuration`
    );
    res.json(response.data);
  } catch (error: any) {
    console.error('Discovery error:', error.response?.data || error.message);
    res.status(500).json({
      error: 'server_error',
      message: 'Failed to fetch discovery document'
    });
  }
});

/**
 * POST /auth/validate-token
 * Validate and decode ID token
 */
router.post('/validate-token', (req: Request, res: Response) => {
  try {
    const { id_token } = req.body;
    
    if (!id_token) {
      return res.status(400).json({
        error: 'invalid_request',
        message: 'id_token required'
      });
    }
    
    const validation = validateIDToken(
      id_token,
      config.CLIENT_ID,
      config.OAUTH2_SERVER
    );
    
    if (!validation.valid) {
      return res.status(401).json({
        error: 'invalid_token',
        message: validation.error
      });
    }
    
    res.json({
      valid: true,
      claims: validation.claims
    });
  } catch (error: any) {
    res.status(500).json({
      error: 'server_error',
      message: error.message
    });
  }
});

/**
 * POST /auth/decode-token
 * Decode JWT without validation (for debugging)
 */
router.post('/decode-token', (req: Request, res: Response) => {
  try {
    const { token } = req.body;
    
    if (!token) {
      return res.status(400).json({
        error: 'invalid_request',
        message: 'token required'
      });
    }
    
    const decoded = decodeJWT(token);
    res.json(decoded);
  } catch (error: any) {
    res.status(400).json({
      error: 'invalid_token',
      message: error.message
    });
  }
});

/**
 * GET /auth/session
 * Get current session info (for debugging)
 */
router.get('/session', requireRefreshToken, (req: Request, res: Response) => {
  const refreshToken = req.cookies.refresh_token;
  
  try {
    // Decode refresh token to get info (without validation)
    const decoded = decodeJWT(refreshToken);
    
    res.json({
      has_refresh_token: true,
      expires_at: decoded.exp ? new Date(decoded.exp * 1000).toISOString() : null,
      user_id: decoded.sub || decoded.user_id,
      scope: decoded.scope
    });
  } catch (error) {
    res.json({
      has_refresh_token: true,
      error: 'Could not decode token'
    });
  }
});

// Clean up expired sessions periodically
setInterval(() => {
  const now = Date.now();
  const maxAge = 10 * 60 * 1000; // 10 minutes
  
  for (const [state, session] of sessions.entries()) {
    if (now - session.timestamp > maxAge) {
      sessions.delete(state);
    }
  }
}, 60 * 1000); // Run every minute

export default router;
