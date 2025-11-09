import { Injectable, UnauthorizedException, BadRequestException } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { HttpService } from '@nestjs/axios';
import { Response } from 'express';
import { firstValueFrom } from 'rxjs';
import { SessionService } from './session.service';
import { OidcService } from '../shared/services/oidc.service';
import { CryptoService } from '../shared/services/crypto.service';
import { LoginResponseDto } from './dto/login-response.dto';
import { TokenResponseDto } from './dto/token-response.dto';
import { UserInfoDto } from './dto/userinfo.dto';
import { ValidationResultDto } from './dto/validation-result.dto';

@Injectable()
export class AuthService {
  private readonly oauth2ServerUrl: string;
  private readonly clientId: string;
  private readonly clientSecret: string;
  private readonly frontendUrl: string;
  private readonly redirectUri: string;

  constructor(
    private readonly configService: ConfigService,
    private readonly sessionService: SessionService,
    private readonly oidcService: OidcService,
    private readonly cryptoService: CryptoService,
    private readonly httpService: HttpService,
  ) {
    this.oauth2ServerUrl =
      this.configService.get<string>('OAUTH2_SERVER_URL') || 'http://localhost:8080';
    this.clientId = this.configService.get<string>('CLIENT_ID') || '';
    this.clientSecret = this.configService.get<string>('CLIENT_SECRET') || '';
    this.frontendUrl = this.configService.get<string>('FRONTEND_URL') || 'http://localhost:5173';
    const port = this.configService.get<string>('PORT') || '3001';
    this.redirectUri = `http://localhost:${port}/auth/callback`;
  }

  /**
   * Initiate OAuth2/OIDC login flow with encrypted state
   */
  async initiateLogin(): Promise<LoginResponseDto> {
    try {
      const state = this.cryptoService.generateState();
      const nonce = this.cryptoService.generateNonce();
      const codeVerifier = this.cryptoService.generateCodeVerifier();
      const codeChallenge = this.cryptoService.generateCodeChallenge(codeVerifier);

      // Encrypt OAuth state data into the state parameter
      const encryptedState = this.cryptoService.encryptOAuthState({
        state,
        nonce,
        codeVerifier,
        redirectUri: this.redirectUri,
      });

      console.log('🔐 Generated encrypted OAuth state:', {
        originalState: state,
        encryptedState: encryptedState.substring(0, 20) + '...',
        nonce,
        codeVerifier: codeVerifier.substring(0, 10) + '...',
      });

      // Build authorization URL with OIDC and PKCE parameters
      const authUrl = new URL(`${this.oauth2ServerUrl}/oauth/authorize`);
      authUrl.searchParams.set('response_type', 'code');
      authUrl.searchParams.set('client_id', this.clientId);
      authUrl.searchParams.set('redirect_uri', this.redirectUri);
      authUrl.searchParams.set('scope', 'openid profile email');
      authUrl.searchParams.set('state', encryptedState);
      authUrl.searchParams.set('nonce', nonce);
      authUrl.searchParams.set('code_challenge', codeChallenge);
      authUrl.searchParams.set('code_challenge_method', 'S256');
      authUrl.searchParams.set('response_mode', 'query');

      return {
        authorization_url: authUrl.toString(),
      };
    } catch (error) {
      console.error('Login error:', error);
      throw new BadRequestException('Failed to initiate login');
    }
  }

  /**
   * Handle OAuth2 callback and exchange code for tokens
   */
  async handleCallback(code: string, state: string, error: string, res: Response): Promise<any> {
    try {
      if (error) {
        return res.json({ error });
      }

      if (!code || !state) {
        return res.json({ error: 'invalid_request' });
      }

      // Decrypt and validate OAuth state from the state parameter
      const oauthState = this.cryptoService.decryptOAuthState(state);

      if (!oauthState) {
        console.error('❌ Failed to decrypt or validate OAuth state');
        return res.json({ error: 'invalid_state' });
      }

      console.log('🔓 Decrypted OAuth state:', {
        originalState: oauthState.state,
        nonce: oauthState.nonce,
        codeVerifier: oauthState.codeVerifier.substring(0, 10) + '...',
        age: Date.now() - oauthState.timestamp + 'ms',
      });

      const { nonce: storedNonce, codeVerifier, redirectUri: storedRedirectUri } = oauthState;

      // Exchange authorization code for tokens (with client_secret and PKCE)
      const tokenResponse = await firstValueFrom(
        this.httpService.post(
          `${this.oauth2ServerUrl}/oauth/token`,
          new URLSearchParams({
            grant_type: 'authorization_code',
            code,
            redirect_uri: storedRedirectUri,
            client_id: this.clientId,
            client_secret: this.clientSecret,
            code_verifier: codeVerifier,
          }),
          {
            headers: {
              'Content-Type': 'application/x-www-form-urlencoded',
            },
          },
        ),
      );

      const tokens = tokenResponse.data;

      // Validate ID token if present
      if (tokens.id_token) {
        const validation = this.oidcService.validateIDToken(
          tokens.id_token,
          this.clientId,
          this.oauth2ServerUrl,
          storedNonce,
        );

        if (!validation.valid) {
          console.error('ID Token validation failed:', validation.error);
          return res.json({ error: 'invalid_id_token', message: validation.error });
        }

        console.log('✅ ID Token validated successfully:', validation.claims);
      }

      // Store refresh token in HttpOnly cookie
      if (tokens.refresh_token) {
        res.cookie('refresh_token', tokens.refresh_token, {
          httpOnly: true,
          secure: this.configService.get<string>('NODE_ENV') === 'production',
          sameSite: 'lax',
          maxAge: 7 * 24 * 60 * 60 * 1000, // 7 days
          path: '/',
        });
      }

      // Build frontend callback URL with tokens
      const callbackUrl = new URL(`${this.frontendUrl}/callback`);
      callbackUrl.searchParams.set('access_token', tokens.access_token);
      callbackUrl.searchParams.set('expires_in', tokens.expires_in.toString());
      if (tokens.id_token) {
        callbackUrl.searchParams.set('id_token', tokens.id_token);
      }

      // Always return JSON with redirect_uri
      // OAuth2 server login page will handle the redirect
      return res.json({
        success: true,
        redirect_uri: callbackUrl.toString(),
        access_token: tokens.access_token,
        expires_in: tokens.expires_in,
        id_token: tokens.id_token,
        token_type: tokens.token_type || 'Bearer',
      });
    } catch (error: any) {
      console.error('Callback error:', error.response?.data || error.message);
      return res.json({
        error: 'token_exchange_failed',
        message: error.response?.data?.error_description || error.message,
      });
    }
  }

  /**
   * Refresh access token using refresh token
   */
  async refreshToken(refreshToken: string, res: Response): Promise<TokenResponseDto> {
    try {
      // Exchange refresh token for new access token
      const tokenResponse = await firstValueFrom(
        this.httpService.post(
          `${this.oauth2ServerUrl}/oauth/token`,
          new URLSearchParams({
            grant_type: 'refresh_token',
            refresh_token: refreshToken,
            client_id: this.clientId,
            client_secret: this.clientSecret,
          }),
          {
            headers: {
              'Content-Type': 'application/x-www-form-urlencoded',
            },
          },
        ),
      );

      const tokens = tokenResponse.data;

      // Update refresh token cookie if rotated
      if (tokens.refresh_token) {
        res.cookie('refresh_token', tokens.refresh_token, {
          httpOnly: true,
          secure: this.configService.get<string>('NODE_ENV') === 'production',
          sameSite: 'lax',
          maxAge: 7 * 24 * 60 * 60 * 1000, // 7 days
          path: '/',
        });
      }

      return {
        access_token: tokens.access_token,
        expires_in: tokens.expires_in,
        token_type: tokens.token_type,
      };
    } catch (error: any) {
      console.error('Refresh error:', error.response?.data || error.message);

      // Clear invalid refresh token
      res.clearCookie('refresh_token', {
        httpOnly: true,
        secure: this.configService.get<string>('NODE_ENV') === 'production',
        sameSite: 'lax',
        path: '/',
      });

      throw new UnauthorizedException('Refresh token expired or invalid');
    }
  }

  /**
   * Logout and clear refresh token cookie
   */
  async logout(res: Response): Promise<void> {
    res.clearCookie('refresh_token', {
      httpOnly: true,
      secure: this.configService.get<string>('NODE_ENV') === 'production',
      sameSite: 'lax',
      path: '/',
    });
  }

  /**
   * Get user info from OAuth2 server
   */
  async getUserInfo(authHeader: string): Promise<UserInfoDto> {
    try {
      const userInfoResponse = await firstValueFrom(
        this.httpService.get(`${this.oauth2ServerUrl}/oauth/userinfo`, {
          headers: {
            Authorization: authHeader,
          },
        }),
      );

      return userInfoResponse.data;
    } catch (error: any) {
      console.error('UserInfo error:', error.response?.data || error.message);
      throw new UnauthorizedException('Failed to get user info');
    }
  }

  /**
   * Get OIDC discovery document
   */
  async getDiscovery(): Promise<any> {
    try {
      const response = await firstValueFrom(
        this.httpService.get(`${this.oauth2ServerUrl}/.well-known/openid-configuration`),
      );
      return response.data;
    } catch (error: any) {
      console.error('Discovery error:', error.response?.data || error.message);
      throw new BadRequestException('Failed to fetch discovery document');
    }
  }

  /**
   * Get JWKS (JSON Web Key Set) from OAuth2 server
   */
  async getJwks(): Promise<any> {
    try {
      const response = await firstValueFrom(
        this.httpService.get(`${this.oauth2ServerUrl}/.well-known/jwks.json`),
      );
      return response.data;
    } catch (error: any) {
      console.error('JWKS error:', error.response?.data || error.message);
      throw new BadRequestException('Failed to fetch JWKS');
    }
  }

  /**
   * Validate ID token
   */
  async validateIDToken(idToken: string): Promise<ValidationResultDto> {
    try {
      if (!idToken) {
        throw new BadRequestException('id_token required');
      }

      const validation = this.oidcService.validateIDToken(
        idToken,
        this.clientId,
        this.oauth2ServerUrl,
      );

      if (!validation.valid) {
        return {
          valid: false,
          error: validation.error,
        };
      }

      return {
        valid: true,
        claims: validation.claims,
      };
    } catch (error: any) {
      throw new BadRequestException(error.message);
    }
  }

  /**
   * Decode JWT token without validation
   */
  decodeToken(token: string): any {
    try {
      if (!token) {
        throw new BadRequestException('token required');
      }

      return this.oidcService.decodeJWT(token);
    } catch (error: any) {
      throw new BadRequestException(error.message);
    }
  }

  /**
   * Get session info from refresh token
   */
  getSessionInfo(refreshToken: string): any {
    try {
      const decoded = this.oidcService.decodeJWT(refreshToken);

      return {
        has_refresh_token: true,
        expires_at: decoded.exp ? new Date(decoded.exp * 1000).toISOString() : null,
        user_id: decoded.sub || decoded.user_id,
        scope: decoded.scope,
      };
    } catch {
      return {
        has_refresh_token: true,
        error: 'Could not decode token',
      };
    }
  }
}
