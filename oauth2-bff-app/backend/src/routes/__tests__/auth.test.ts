import request from 'supertest';
import express from 'express';
import nock from 'nock';
import cookieParser from 'cookie-parser';
import session from 'express-session';
import authRoutes from '../auth';
import config from '../../config';

// Create test app
const createTestApp = () => {
  const app = express();
  app.use(express.json());
  app.use(cookieParser());
  app.use(session({
    secret: 'test-secret',
    resave: false,
    saveUninitialized: false,
    cookie: {
      httpOnly: true,
      secure: false,
      sameSite: 'lax',
      maxAge: 10 * 60 * 1000,
    },
  }));
  app.use('/auth', authRoutes);
  return app;
};

describe('Auth Routes', () => {
  let app: express.Application;

  beforeEach(() => {
    app = createTestApp();
    nock.cleanAll();
  });

  afterEach(() => {
    nock.cleanAll();
  });

  describe('GET /auth/login', () => {
    it('should return authorization URL with PKCE parameters', async () => {
      const response = await request(app)
        .get('/auth/login')
        .expect(200);

      expect(response.body).toHaveProperty('authorization_url');
      
      const authUrl = new URL(response.body.authorization_url);
      expect(authUrl.origin).toBe(config.OAUTH2_SERVER);
      expect(authUrl.pathname).toBe('/oauth/authorize');
      expect(authUrl.searchParams.get('response_type')).toBe('code');
      expect(authUrl.searchParams.get('client_id')).toBe(config.CLIENT_ID);
      expect(authUrl.searchParams.get('redirect_uri')).toBe(config.REDIRECT_URI);
      expect(authUrl.searchParams.get('scope')).toBe('openid profile email');
      expect(authUrl.searchParams.get('state')).toBeTruthy();
      expect(authUrl.searchParams.get('nonce')).toBeTruthy();
      expect(authUrl.searchParams.get('code_challenge')).toBeTruthy();
      expect(authUrl.searchParams.get('code_challenge_method')).toBe('S256');
    });
  });

  describe('GET /auth/callback', () => {
    it('should exchange authorization code for tokens', async () => {
      // Create agent to maintain session
      const agent = request.agent(app);
      
      // First, initiate login to set up session
      const loginResponse = await agent
        .get('/auth/login')
        .expect(200);

      const authUrl = new URL(loginResponse.body.authorization_url);
      const state = authUrl.searchParams.get('state');

      // Mock token exchange
      nock(config.OAUTH2_SERVER)
        .post('/oauth/token')
        .reply(200, {
          access_token: 'test_access_token',
          token_type: 'Bearer',
          expires_in: 3600,
          refresh_token: 'test_refresh_token',
          id_token: 'eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwODAiLCJzdWIiOiJ1c2VyMTIzIiwiYXVkIjoidGVzdC1jbGllbnQiLCJleHAiOjk5OTk5OTk5OTksImlhdCI6MTYwMDAwMDAwMCwibm9uY2UiOiJ0ZXN0LW5vbmNlIn0.test'
        });

      const response = await agent
        .get('/auth/callback')
        .query({ code: 'test_code', state })
        .expect(302);

      expect(response.headers.location).toContain(config.FRONTEND_URL);
      expect(response.headers.location).toContain('access_token=test_access_token');
      expect(response.headers['set-cookie']).toBeDefined();
      
      // Check refresh token cookie
      const setCookieHeader = response.headers['set-cookie'];
      const responseCookies = Array.isArray(setCookieHeader) ? setCookieHeader : [setCookieHeader];
      const refreshTokenCookie = responseCookies.find((c: string) => 
        c.startsWith('refresh_token=')
      );
      expect(refreshTokenCookie).toBeDefined();
      expect(refreshTokenCookie).toContain('HttpOnly');
    });

    it('should return error for invalid state', async () => {
      const response = await request(app)
        .get('/auth/callback')
        .query({ code: 'test_code', state: 'invalid_state' })
        .expect(302);

      expect(response.headers.location).toContain('error=invalid_state');
    });

    it('should return error when code is missing', async () => {
      const response = await request(app)
        .get('/auth/callback')
        .query({ state: 'test_state' })
        .expect(302);

      expect(response.headers.location).toContain('error=invalid_request');
    });
  });

  describe('POST /auth/refresh', () => {
    it('should refresh access token using refresh token cookie', async () => {
      // Mock token refresh
      nock(config.OAUTH2_SERVER)
        .post('/oauth/token')
        .reply(200, {
          access_token: 'new_access_token',
          token_type: 'Bearer',
          expires_in: 3600
        });

      const response = await request(app)
        .post('/auth/refresh')
        .set('Cookie', ['refresh_token=test_refresh_token'])
        .expect(200);

      expect(response.body).toHaveProperty('access_token', 'new_access_token');
      expect(response.body).toHaveProperty('expires_in', 3600);
      expect(response.body).toHaveProperty('token_type', 'Bearer');
    });

    it('should return 401 when refresh token is missing', async () => {
      const response = await request(app)
        .post('/auth/refresh')
        .expect(401);

      expect(response.body).toHaveProperty('error', 'unauthorized');
    });

    it('should clear cookie and return 401 when refresh token is invalid', async () => {
      // Mock failed token refresh
      nock(config.OAUTH2_SERVER)
        .post('/oauth/token')
        .reply(400, {
          error: 'invalid_grant'
        });

      const response = await request(app)
        .post('/auth/refresh')
        .set('Cookie', ['refresh_token=invalid_token'])
        .expect(401);

      expect(response.body).toHaveProperty('error', 'invalid_grant');
      
      // Check that refresh token cookie is cleared
      const setCookieHeader = response.headers['set-cookie'];
      const cookies = Array.isArray(setCookieHeader) ? setCookieHeader : [setCookieHeader];
      const clearCookie = cookies.find((c: string) => 
        c.startsWith('refresh_token=')
      );
      expect(clearCookie).toMatch(/Max-Age=0|Expires=Thu, 01 Jan 1970/);
    });
  });

  describe('POST /auth/logout', () => {
    it('should clear refresh token cookie', async () => {
      const response = await request(app)
        .post('/auth/logout')
        .set('Cookie', ['refresh_token=test_token'])
        .expect(200);

      expect(response.body).toHaveProperty('message', 'Logged out successfully');
      
      // Check that refresh token cookie is cleared
      const setCookieHeader = response.headers['set-cookie'];
      const cookies = Array.isArray(setCookieHeader) ? setCookieHeader : [setCookieHeader];
      const clearCookie = cookies.find((c: string) => 
        c.startsWith('refresh_token=')
      );
      expect(clearCookie).toBeDefined();
      expect(clearCookie).toMatch(/Max-Age=0|Expires=Thu, 01 Jan 1970/);
    });
  });

  describe('GET /auth/userinfo', () => {
    it('should return user info from OAuth2 server', async () => {
      const mockUserInfo = {
        sub: 'user123',
        email: 'user@example.com',
        name: 'Test User',
        email_verified: true
      };

      nock(config.OAUTH2_SERVER)
        .get('/oauth/userinfo')
        .reply(200, mockUserInfo);

      const response = await request(app)
        .get('/auth/userinfo')
        .set('Authorization', 'Bearer test_access_token')
        .expect(200);

      expect(response.body).toEqual(mockUserInfo);
    });

    it('should return 401 when access token is missing', async () => {
      const response = await request(app)
        .get('/auth/userinfo')
        .expect(401);

      expect(response.body).toHaveProperty('error', 'unauthorized');
    });
  });

  describe('GET /auth/discovery', () => {
    it('should return OIDC discovery document', async () => {
      const mockDiscovery = {
        issuer: config.OAUTH2_SERVER,
        authorization_endpoint: `${config.OAUTH2_SERVER}/oauth/authorize`,
        token_endpoint: `${config.OAUTH2_SERVER}/oauth/token`,
        userinfo_endpoint: `${config.OAUTH2_SERVER}/oauth/userinfo`,
        jwks_uri: `${config.OAUTH2_SERVER}/.well-known/jwks.json`
      };

      nock(config.OAUTH2_SERVER)
        .get('/.well-known/openid-configuration')
        .reply(200, mockDiscovery);

      const response = await request(app)
        .get('/auth/discovery')
        .expect(200);

      expect(response.body).toEqual(mockDiscovery);
    });
  });
});
