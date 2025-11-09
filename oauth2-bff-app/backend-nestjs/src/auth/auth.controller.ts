import { Controller, Get, Post, Query, Body, Headers, Req, Res, UseGuards } from '@nestjs/common';
import { Request, Response } from 'express';
import { AuthService } from './auth.service';
import { AuthGuard } from './guards/auth.guard';
import { RefreshGuard } from './guards/refresh.guard';
import { LoginResponseDto } from './dto/login-response.dto';
import { TokenResponseDto } from './dto/token-response.dto';
import { UserInfoDto } from './dto/userinfo.dto';
import { ValidationResultDto } from './dto/validation-result.dto';

@Controller('auth')
export class AuthController {
  constructor(private readonly authService: AuthService) {}

  /**
   * Initiate OAuth2/OIDC login flow
   * GET /auth/login
   */
  @Get('login')
  async login(): Promise<LoginResponseDto> {
    return this.authService.initiateLogin();
  }

  /**
   * Handle OAuth2 callback
   * GET /auth/callback
   */
  @Get('callback')
  async callback(
    @Query('code') code: string,
    @Query('state') state: string,
    @Query('error') error: string,
    @Res() res: Response,
  ): Promise<any> {
    return this.authService.handleCallback(code, state, error, res);
  }

  /**
   * Refresh access token using refresh token
   * POST /auth/refresh
   */
  @Post('refresh')
  @UseGuards(RefreshGuard)
  async refresh(
    @Req() req: Request,
    @Res({ passthrough: true }) res: Response,
  ): Promise<TokenResponseDto> {
    const refreshToken = req.cookies.refresh_token;
    return this.authService.refreshToken(refreshToken, res);
  }

  /**
   * Logout and clear refresh token cookie
   * POST /auth/logout
   */
  @Post('logout')
  async logout(@Res({ passthrough: true }) res: Response): Promise<{ message: string }> {
    await this.authService.logout(res);
    return { message: 'Logged out successfully' };
  }

  /**
   * Get user info from OAuth2 server
   * GET /auth/userinfo
   */
  @Get('userinfo')
  @UseGuards(AuthGuard)
  async getUserInfo(@Headers('authorization') authorization: string): Promise<UserInfoDto> {
    return this.authService.getUserInfo(authorization);
  }

  /**
   * Get OIDC discovery document
   * GET /auth/discovery
   */
  @Get('discovery')
  async getDiscovery(): Promise<any> {
    return this.authService.getDiscovery();
  }

  /**
   * Get JWKS (JSON Web Key Set) from OAuth2 server
   * GET /auth/jwks
   */
  @Get('jwks')
  async getJwks(): Promise<any> {
    return this.authService.getJwks();
  }

  /**
   * Validate ID token
   * POST /auth/validate-token
   */
  @Post('validate-token')
  async validateToken(@Body('id_token') idToken: string): Promise<ValidationResultDto> {
    return this.authService.validateIDToken(idToken);
  }

  /**
   * Decode JWT token without validation
   * POST /auth/decode-token
   */
  @Post('decode-token')
  async decodeToken(@Body('token') token: string): Promise<any> {
    return this.authService.decodeToken(token);
  }

  /**
   * Get session info from refresh token
   * GET /auth/session
   */
  @Get('session')
  @UseGuards(RefreshGuard)
  async getSession(@Req() req: Request): Promise<any> {
    const refreshToken = req.cookies.refresh_token;
    return this.authService.getSessionInfo(refreshToken);
  }
}
