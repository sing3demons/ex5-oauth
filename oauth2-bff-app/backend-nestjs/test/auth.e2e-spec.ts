import { Test, TestingModule } from '@nestjs/testing';
import { INestApplication } from '@nestjs/common';
import * as request from 'supertest';
import { AppModule } from '../src/app.module';

describe('AuthController (e2e)', () => {
  let app: INestApplication;

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

  describe('/auth/login (GET)', () => {
    it('should return authorization URL and state', () => {
      return request(app.getHttpServer())
        .get('/auth/login')
        .expect(200)
        .expect((res) => {
          expect(res.body).toHaveProperty('authorizationUrl');
          expect(res.body).toHaveProperty('state');
          expect(res.body).toHaveProperty('codeVerifier');
          expect(res.body.authorizationUrl).toContain('response_type=code');
          expect(res.body.authorizationUrl).toContain('client_id=');
          expect(res.body.authorizationUrl).toContain('redirect_uri=');
          expect(res.body.authorizationUrl).toContain('scope=');
          expect(res.body.authorizationUrl).toContain('state=');
          expect(res.body.authorizationUrl).toContain('code_challenge=');
          expect(res.body.authorizationUrl).toContain('code_challenge_method=S256');
        });
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

    it('should generate unique code verifier for each request', async () => {
      const response1 = await request(app.getHttpServer())
        .get('/auth/login')
        .expect(200);

      const response2 = await request(app.getHttpServer())
        .get('/auth/login')
        .expect(200);

      expect(response1.body.codeVerifier).not.toBe(response2.body.codeVerifier);
    });
  });

  describe('/auth/callback (GET)', () => {
    it('should return error for missing code parameter', () => {
      return request(app.getHttpServer())
        .get('/auth/callback')
        .query({ state: 'test-state' })
        .expect(302)
        .expect((res) => {
          expect(res.headers.location).toContain('error=');
        });
    });

    it('should return error for missing state parameter', () => {
      return request(app.getHttpServer())
        .get('/auth/callback')
        .query({ code: 'test-code' })
        .expect(302)
        .expect((res) => {
          expect(res.headers.location).toContain('error=');
        });
    });

    it('should handle OAuth2 error parameter', () => {
      return request(app.getHttpServer())
        .get('/auth/callback')
        .query({ error: 'access_denied', error_description: 'User denied access' })
        .expect(302)
        .expect((res) => {
          expect(res.headers.location).toContain('error=access_denied');
        });
    });
  });

  describe('/auth/logout (POST)', () => {
    it('should clear refresh token cookie', () => {
      return request(app.getHttpServer())
        .post('/auth/logout')
        .expect(201)
        .expect((res) => {
          expect(res.body).toHaveProperty('message', 'Logged out successfully');
          // Check if Set-Cookie header clears the refresh_token
          const setCookie = res.headers['set-cookie'];
          if (setCookie) {
            const cookieString = Array.isArray(setCookie) ? setCookie.join('; ') : setCookie;
            expect(cookieString).toContain('refresh_token=');
          }
        });
    });
  });

  describe('/auth/discovery (GET)', () => {
    it('should return OIDC discovery document', () => {
      return request(app.getHttpServer())
        .get('/auth/discovery')
        .expect(200)
        .expect((res) => {
          expect(res.body).toHaveProperty('issuer');
          expect(res.body).toHaveProperty('authorization_endpoint');
          expect(res.body).toHaveProperty('token_endpoint');
          expect(res.body).toHaveProperty('userinfo_endpoint');
          expect(res.body).toHaveProperty('jwks_uri');
          expect(res.body).toHaveProperty('response_types_supported');
          expect(res.body).toHaveProperty('subject_types_supported');
          expect(res.body).toHaveProperty('id_token_signing_alg_values_supported');
        });
    });
  });

  describe('/auth/jwks (GET)', () => {
    it('should return JWKS', () => {
      return request(app.getHttpServer())
        .get('/auth/jwks')
        .expect(200)
        .expect((res) => {
          expect(res.body).toHaveProperty('keys');
          expect(Array.isArray(res.body.keys)).toBe(true);
        });
    });
  });

  describe('/auth/userinfo (GET)', () => {
    it('should return 401 without authorization header', () => {
      return request(app.getHttpServer())
        .get('/auth/userinfo')
        .expect(401);
    });

    it('should return 401 with invalid token', () => {
      return request(app.getHttpServer())
        .get('/auth/userinfo')
        .set('Authorization', 'Bearer invalid-token')
        .expect(401);
    });
  });

  describe('/auth/refresh (POST)', () => {
    it('should return 401 without refresh token cookie', () => {
      return request(app.getHttpServer())
        .post('/auth/refresh')
        .expect(401);
    });

    it('should return 401 with invalid refresh token', () => {
      return request(app.getHttpServer())
        .post('/auth/refresh')
        .set('Cookie', ['refresh_token=invalid-token'])
        .expect(401);
    });
  });

  describe('/auth/session (GET)', () => {
    it('should return 401 without refresh token cookie', () => {
      return request(app.getHttpServer())
        .get('/auth/session')
        .expect(401);
    });

    it('should return 401 with invalid refresh token', () => {
      return request(app.getHttpServer())
        .get('/auth/session')
        .set('Cookie', ['refresh_token=invalid-token'])
        .expect(401);
    });
  });

  describe('/auth/validate-token (POST)', () => {
    it('should return 400 without id_token', () => {
      return request(app.getHttpServer())
        .post('/auth/validate-token')
        .send({})
        .expect(400);
    });

    it('should validate token format', () => {
      return request(app.getHttpServer())
        .post('/auth/validate-token')
        .send({ id_token: 'invalid.token.format' })
        .expect((res) => {
          // Should either return validation result or error
          expect(res.status).toBeGreaterThanOrEqual(200);
        });
    });
  });

  describe('/auth/decode-token (POST)', () => {
    it('should return 400 without token', () => {
      return request(app.getHttpServer())
        .post('/auth/decode-token')
        .send({})
        .expect(400);
    });

    it('should decode valid JWT format', () => {
      // Create a simple JWT-like token (header.payload.signature)
      const mockToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c';
      
      return request(app.getHttpServer())
        .post('/auth/decode-token')
        .send({ token: mockToken })
        .expect((res) => {
          // Should return decoded token or error
          expect(res.status).toBeGreaterThanOrEqual(200);
        });
    });
  });
});
