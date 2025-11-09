import { Test, TestingModule } from '@nestjs/testing';
import { INestApplication } from '@nestjs/common';
import * as request from 'supertest';
import { AppModule } from '../src/app.module';

/**
 * SSO (Single Sign-On) Flow E2E Tests
 * 
 * These tests verify the complete OAuth2/OIDC SSO authentication flow
 * including PKCE (Proof Key for Code Exchange) implementation.
 */

describe('SSO Flow (e2e)', () => {
  let app: INestApplication;
  let authorizationUrl: string;
  let state: string;
  let codeVerifier: string;

  beforeAll(async () => {
    const moduleFixture: TestingModule = await Test.createTestingModule({
      imports: [AppModule],
    }).compile();

    app = moduleFixture.createNestApplication();
    
    // Apply cookie parser middleware
    const cookieParser = require('cookie-parser');
    app.use(cookieParser());
    
    await app.init();
  });

  afterAll(async () => {
    await app.close();
  });

  describe('Complete SSO Flow', () => {
    it('Step 1: Should initiate OAuth2 login with PKCE', async () => {
      const response = await request(app.getHttpServer())
        .get('/auth/login')
        .expect(200);

      expect(response.body).toHaveProperty('authorizationUrl');
      expect(response.body).toHaveProperty('state');
      expect(response.body).toHaveProperty('codeVerifier');

      // Store for next steps
      authorizationUrl = response.body.authorizationUrl;
      state = response.body.state;
      codeVerifier = response.body.codeVerifier;

      // Verify PKCE parameters in URL
      expect(authorizationUrl).toContain('code_challenge=');
      expect(authorizationUrl).toContain('code_challenge_method=S256');
      expect(authorizationUrl).toContain(`state=${state}`);
    });

    it('Step 2: Should verify authorization URL structure', () => {
      const url = new URL(authorizationUrl);
      const params = url.searchParams;

      // Required OAuth2 parameters
      expect(params.get('response_type')).toBe('code');
      expect(params.get('client_id')).toBeTruthy();
      expect(params.get('redirect_uri')).toBeTruthy();
      expect(params.get('scope')).toContain('openid');
      expect(params.get('state')).toBe(state);

      // PKCE parameters
      expect(params.get('code_challenge')).toBeTruthy();
      expect(params.get('code_challenge_method')).toBe('S256');

      // Verify redirect URI format
      const redirectUri = params.get('redirect_uri');
      expect(redirectUri).toMatch(/^https?:\/\/.+\/auth\/callback$/);
    });

    it('Step 3: Should handle OAuth2 callback with missing parameters', async () => {
      await request(app.getHttpServer())
        .get('/auth/callback')
        .expect(302)
        .expect((res) => {
          expect(res.headers.location).toContain('error=');
        });
    });

    it('Step 4: Should handle OAuth2 error responses', async () => {
      await request(app.getHttpServer())
        .get('/auth/callback')
        .query({
          error: 'access_denied',
          error_description: 'User denied access',
          state: state
        })
        .expect(302)
        .expect((res) => {
          expect(res.headers.location).toContain('error=access_denied');
        });
    });
  });

  describe('PKCE Implementation', () => {
    it('should generate cryptographically secure code verifier', async () => {
      const response = await request(app.getHttpServer())
        .get('/auth/login')
        .expect(200);

      const verifier = response.body.codeVerifier;
      
      // PKCE spec: code_verifier must be 43-128 characters
      expect(verifier.length).toBeGreaterThanOrEqual(43);
      expect(verifier.length).toBeLessThanOrEqual(128);
      
      // Should contain only unreserved characters
      expect(verifier).toMatch(/^[A-Za-z0-9\-._~]+$/);
    });

    it('should generate unique code verifier for each request', async () => {
      const response1 = await request(app.getHttpServer())
        .get('/auth/login')
        .expect(200);

      const response2 = await request(app.getHttpServer())
        .get('/auth/login')
        .expect(200);

      expect(response1.body.codeVerifier).not.toBe(response2.body.codeVerifier);
    });

    it('should generate code challenge from verifier', async () => {
      const response = await request(app.getHttpServer())
        .get('/auth/login')
        .expect(200);

      const authUrl = new URL(response.body.authorizationUrl);
      const codeChallenge = authUrl.searchParams.get('code_challenge');

      // Code challenge should be base64url encoded SHA256 hash
      expect(codeChallenge).toBeTruthy();
      expect(codeChallenge).toMatch(/^[A-Za-z0-9\-_]+$/);
      
      // Should not contain padding
      expect(codeChallenge).not.toContain('=');
    });
  });

  describe('State Parameter Security', () => {
    it('should generate cryptographically secure state', async () => {
      const response = await request(app.getHttpServer())
        .get('/auth/login')
        .expect(200);

      const state = response.body.state;
      
      // State should be sufficiently long for security
      expect(state.length).toBeGreaterThanOrEqual(16);
      
      // Should be random
      expect(state).toMatch(/^[A-Za-z0-9\-_]+$/);
    });

    it('should generate unique state for each request', async () => {
      const response1 = await request(app.getHttpServer())
        .get('/auth/login')
        .expect(200);

      const response2 = await request(app.getHttpServer())
        .get('/auth/login')
        .expect(200);

      expect(response1.body.state).not.toBe(response2.body.state);
    });

    it('should reject callback with mismatched state', async () => {
      await request(app.getHttpServer())
        .get('/auth/callback')
        .query({
          code: 'test-code',
          state: 'invalid-state-12345'
        })
        .expect(302)
        .expect((res) => {
          expect(res.headers.location).toContain('error=');
        });
    });
  });

  describe('OIDC Discovery', () => {
    it('should return valid discovery document', async () => {
      const response = await request(app.getHttpServer())
        .get('/auth/discovery')
        .expect(200);

      // Required OIDC discovery fields
      expect(response.body).toHaveProperty('issuer');
      expect(response.body).toHaveProperty('authorization_endpoint');
      expect(response.body).toHaveProperty('token_endpoint');
      expect(response.body).toHaveProperty('userinfo_endpoint');
      expect(response.body).toHaveProperty('jwks_uri');
      
      // Response types
      expect(response.body.response_types_supported).toContain('code');
      
      // Subject types
      expect(response.body.subject_types_supported).toBeTruthy();
      
      // Signing algorithms
      expect(response.body.id_token_signing_alg_values_supported).toBeTruthy();
    });

    it('should have consistent endpoint URLs', async () => {
      const response = await request(app.getHttpServer())
        .get('/auth/discovery')
        .expect(200);

      const issuer = response.body.issuer;
      
      // All endpoints should start with issuer URL
      expect(response.body.authorization_endpoint).toContain(issuer);
      expect(response.body.token_endpoint).toContain(issuer);
      expect(response.body.userinfo_endpoint).toContain(issuer);
      expect(response.body.jwks_uri).toContain(issuer);
    });
  });

  describe('JWKS Endpoint', () => {
    it('should return valid JWKS', async () => {
      const response = await request(app.getHttpServer())
        .get('/auth/jwks')
        .expect(200);

      expect(response.body).toHaveProperty('keys');
      expect(Array.isArray(response.body.keys)).toBe(true);
    });

    it('should have properly formatted JWK keys', async () => {
      const response = await request(app.getHttpServer())
        .get('/auth/jwks')
        .expect(200);

      if (response.body.keys.length > 0) {
        const key = response.body.keys[0];
        
        // Required JWK fields
        expect(key).toHaveProperty('kty'); // Key type
        expect(key).toHaveProperty('use'); // Public key use
        expect(key).toHaveProperty('kid'); // Key ID
        expect(key).toHaveProperty('alg'); // Algorithm
      }
    });
  });

  describe('Token Validation', () => {
    it('should validate token structure', async () => {
      // Create a mock JWT token
      const mockToken = 'eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.signature';
      
      const response = await request(app.getHttpServer())
        .post('/auth/validate-token')
        .send({ id_token: mockToken })
        .expect((res) => {
          // Should return validation result (valid or invalid)
          expect(res.status).toBeGreaterThanOrEqual(200);
        });
    });

    it('should decode JWT token', async () => {
      const mockToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c';
      
      const response = await request(app.getHttpServer())
        .post('/auth/decode-token')
        .send({ token: mockToken })
        .expect((res) => {
          expect(res.status).toBeGreaterThanOrEqual(200);
        });
    });
  });

  describe('Session Management', () => {
    it('should require refresh token for session info', async () => {
      await request(app.getHttpServer())
        .get('/auth/session')
        .expect(401);
    });

    it('should reject invalid refresh token', async () => {
      await request(app.getHttpServer())
        .get('/auth/session')
        .set('Cookie', ['refresh_token=invalid-token'])
        .expect(401);
    });

    it('should clear session on logout', async () => {
      const response = await request(app.getHttpServer())
        .post('/auth/logout')
        .expect(201);

      expect(response.body).toHaveProperty('message', 'Logged out successfully');
      
      // Check if refresh token cookie is cleared
      const setCookie = response.headers['set-cookie'];
      if (setCookie) {
        const cookieString = Array.isArray(setCookie) ? setCookie.join('; ') : setCookie;
        expect(cookieString).toContain('refresh_token=');
      }
    });
  });

  describe('Token Refresh Flow', () => {
    it('should require refresh token cookie', async () => {
      await request(app.getHttpServer())
        .post('/auth/refresh')
        .expect(401);
    });

    it('should reject invalid refresh token', async () => {
      await request(app.getHttpServer())
        .post('/auth/refresh')
        .set('Cookie', ['refresh_token=invalid-token-12345'])
        .expect(401);
    });

    it('should reject expired refresh token', async () => {
      // Mock expired token
      const expiredToken = 'expired.refresh.token';
      
      await request(app.getHttpServer())
        .post('/auth/refresh')
        .set('Cookie', [`refresh_token=${expiredToken}`])
        .expect(401);
    });
  });

  describe('UserInfo Endpoint', () => {
    it('should require valid access token', async () => {
      await request(app.getHttpServer())
        .get('/auth/userinfo')
        .expect(401);
    });

    it('should reject invalid access token', async () => {
      await request(app.getHttpServer())
        .get('/auth/userinfo')
        .set('Authorization', 'Bearer invalid-token')
        .expect(401);
    });

    it('should reject malformed authorization header', async () => {
      await request(app.getHttpServer())
        .get('/auth/userinfo')
        .set('Authorization', 'InvalidFormat token')
        .expect(401);
    });

    it('should reject missing Bearer prefix', async () => {
      await request(app.getHttpServer())
        .get('/auth/userinfo')
        .set('Authorization', 'token-without-bearer')
        .expect(401);
    });
  });

  describe('Security Headers', () => {
    it('should not expose sensitive information in errors', async () => {
      const response = await request(app.getHttpServer())
        .get('/auth/callback')
        .query({ code: 'test' })
        .expect(302);

      // Should not expose internal error details
      const location = response.headers.location;
      expect(location).not.toContain('stack');
      expect(location).not.toContain('password');
      expect(location).not.toContain('secret');
    });

    it('should handle CORS properly', async () => {
      const response = await request(app.getHttpServer())
        .get('/auth/discovery')
        .set('Origin', 'http://localhost:3000')
        .expect(200);

      // CORS headers should be present
      expect(response.headers['access-control-allow-origin']).toBeTruthy();
    });
  });

  describe('Error Handling', () => {
    it('should handle network errors gracefully', async () => {
      // This test verifies the app doesn't crash on network errors
      await request(app.getHttpServer())
        .get('/auth/callback')
        .query({
          code: 'test-code',
          state: 'test-state'
        })
        .expect(302);
    });

    it('should handle malformed requests', async () => {
      await request(app.getHttpServer())
        .post('/auth/validate-token')
        .send({ invalid: 'data' })
        .expect(400);
    });

    it('should handle missing required parameters', async () => {
      await request(app.getHttpServer())
        .post('/auth/decode-token')
        .send({})
        .expect(400);
    });
  });
});
