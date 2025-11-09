import request from 'supertest';
import express from 'express';
import cors from 'cors';
import cookieParser from 'cookie-parser';
import session from 'express-session';
import { csrfProtection } from '../middleware/csrf';
import { requireAuth } from '../middleware/auth';
import config from '../config';

describe('Security Tests', () => {
  describe('CSRF Protection', () => {
    let app: express.Application;

    beforeEach(() => {
      app = express();
      app.use(express.json());
      app.use(cookieParser());
      app.use(session({
        secret: 'test-secret',
        resave: false,
        saveUninitialized: false,
      }));

      // CSRF token endpoint
      app.get('/csrf-token', csrfProtection, (req, res) => {
        res.json({ csrfToken: req.csrfToken() });
      });

      // Protected endpoint
      app.post('/protected', csrfProtection, (req, res) => {
        res.json({ message: 'Success' });
      });
    });

    it('should provide CSRF token via /csrf-token endpoint', async () => {
      const response = await request(app)
        .get('/csrf-token')
        .expect(200);

      expect(response.body).toHaveProperty('csrfToken');
      expect(typeof response.body.csrfToken).toBe('string');
      expect(response.body.csrfToken.length).toBeGreaterThan(0);
    });

    it('should reject POST request without CSRF token', async () => {
      const response = await request(app)
        .post('/protected')
        .send({ data: 'test' })
        .expect(403);

      // CSRF error response format may vary
      expect(response.status).toBe(403);
    });

    it('should accept POST request with valid CSRF token', async () => {
      const agent = request.agent(app);

      // Get CSRF token
      const tokenResponse = await agent
        .get('/csrf-token')
        .expect(200);

      const csrfToken = tokenResponse.body.csrfToken;

      // Make request with CSRF token
      const response = await agent
        .post('/protected')
        .set('X-CSRF-Token', csrfToken)
        .send({ data: 'test' })
        .expect(200);

      expect(response.body).toHaveProperty('message', 'Success');
    });

    it('should reject POST request with invalid CSRF token', async () => {
      const response = await request(app)
        .post('/protected')
        .set('X-CSRF-Token', 'invalid-token')
        .send({ data: 'test' })
        .expect(403);

      // CSRF error response format may vary
      expect(response.status).toBe(403);
    });
  });

  describe('CORS Configuration', () => {
    let app: express.Application;

    beforeEach(() => {
      app = express();
      app.use(express.json());

      // Apply CORS with config
      app.use(cors({
        origin: (origin, callback) => {
          if (!origin) {
            return callback(null, true);
          }
          
          if (config.CORS_ORIGINS.includes(origin)) {
            callback(null, true);
          } else {
            callback(new Error('Not allowed by CORS'));
          }
        },
        credentials: true,
        methods: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS'],
        allowedHeaders: ['Content-Type', 'Authorization', 'X-CSRF-Token'],
        exposedHeaders: ['X-CSRF-Token'],
      }));

      app.get('/test', (req, res) => {
        res.json({ message: 'OK' });
      });
    });

    it('should allow requests from configured origins', async () => {
      const allowedOrigin = config.CORS_ORIGINS[0];

      const response = await request(app)
        .get('/test')
        .set('Origin', allowedOrigin)
        .expect(200);

      expect(response.headers['access-control-allow-origin']).toBe(allowedOrigin);
      expect(response.headers['access-control-allow-credentials']).toBe('true');
    });

    it('should allow requests with no origin', async () => {
      const response = await request(app)
        .get('/test')
        .expect(200);

      expect(response.body).toHaveProperty('message', 'OK');
    });

    it('should include correct CORS headers in preflight response', async () => {
      const allowedOrigin = config.CORS_ORIGINS[0];

      const response = await request(app)
        .options('/test')
        .set('Origin', allowedOrigin)
        .set('Access-Control-Request-Method', 'POST')
        .set('Access-Control-Request-Headers', 'Content-Type,Authorization')
        .expect(204);

      expect(response.headers['access-control-allow-origin']).toBe(allowedOrigin);
      expect(response.headers['access-control-allow-methods']).toContain('POST');
      expect(response.headers['access-control-allow-headers']).toContain('Authorization');
      expect(response.headers['access-control-allow-credentials']).toBe('true');
    });
  });

  describe('Token Validation and Authorization', () => {
    let app: express.Application;

    beforeEach(() => {
      app = express();
      app.use(express.json());

      // Protected endpoint
      app.get('/protected', requireAuth, (req, res) => {
        res.json({ 
          message: 'Success',
          userId: req.user?.id 
        });
      });
    });

    it('should reject requests without Authorization header', async () => {
      const response = await request(app)
        .get('/protected')
        .expect(401);

      expect(response.body).toHaveProperty('error', 'unauthorized');
      expect(response.body.message).toMatch(/access token/i);
    });

    it('should reject requests with invalid token format', async () => {
      const response = await request(app)
        .get('/protected')
        .set('Authorization', 'InvalidFormat')
        .expect(401);

      expect(response.body).toHaveProperty('error', 'unauthorized');
    });

    it('should reject requests with Bearer but no token', async () => {
      const response = await request(app)
        .get('/protected')
        .set('Authorization', 'Bearer ')
        .expect(401);

      expect(response.body).toHaveProperty('error', 'unauthorized');
    });
  });

  describe('Cookie Security', () => {
    let app: express.Application;

    beforeEach(() => {
      app = express();
      app.use(express.json());
      app.use(cookieParser());

      app.post('/set-cookie', (req, res) => {
        res.cookie('refresh_token', 'test_token', {
          httpOnly: true,
          secure: process.env.NODE_ENV === 'production',
          sameSite: 'lax',
          maxAge: 7 * 24 * 60 * 60 * 1000,
        });
        res.json({ message: 'Cookie set' });
      });
    });

    it('should set cookies with httpOnly flag', async () => {
      const response = await request(app)
        .post('/set-cookie')
        .expect(200);

      const setCookieHeader = response.headers['set-cookie'];
      const cookies = Array.isArray(setCookieHeader) ? setCookieHeader : [setCookieHeader];
      const refreshTokenCookie = cookies.find((c: string) => 
        c.startsWith('refresh_token=')
      );

      expect(refreshTokenCookie).toBeDefined();
      expect(refreshTokenCookie).toContain('HttpOnly');
      expect(refreshTokenCookie).toContain('SameSite=Lax');
    });

    it('should set secure flag in production', async () => {
      const originalEnv = process.env.NODE_ENV;
      process.env.NODE_ENV = 'production';

      const response = await request(app)
        .post('/set-cookie')
        .expect(200);

      const setCookieHeader = response.headers['set-cookie'];
      const cookies = Array.isArray(setCookieHeader) ? setCookieHeader : [setCookieHeader];
      const refreshTokenCookie = cookies.find((c: string) => 
        c.startsWith('refresh_token=')
      );

      expect(refreshTokenCookie).toContain('Secure');

      process.env.NODE_ENV = originalEnv;
    });
  });
});
