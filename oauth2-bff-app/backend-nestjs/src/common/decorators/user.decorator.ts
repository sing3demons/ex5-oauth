import { createParamDecorator, ExecutionContext, UnauthorizedException } from '@nestjs/common';

/**
 * Custom decorator to extract user ID from Authorization header
 * Decodes JWT and returns sub or user_id claim
 *
 * Usage:
 * @Get()
 * async findAll(@User() userId: string) {
 *   // userId is extracted from the JWT token
 * }
 */
export const User = createParamDecorator((data: unknown, ctx: ExecutionContext): string => {
  const request = ctx.switchToHttp().getRequest();
  const authHeader = request.headers.authorization;

  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    throw new UnauthorizedException('No access token provided');
  }

  try {
    const token = authHeader.replace('Bearer ', '');

    // Decode JWT (without verification - verification is done by AuthGuard)
    const base64Payload = token.split('.')[1];
    if (!base64Payload) {
      throw new UnauthorizedException('Invalid token format');
    }

    const payload = JSON.parse(Buffer.from(base64Payload, 'base64').toString('utf-8'));

    const userId = payload.sub || payload.user_id;

    if (!userId) {
      throw new UnauthorizedException('User ID not found in token');
    }

    return userId;
  } catch (error) {
    if (error instanceof UnauthorizedException) {
      throw error;
    }
    throw new UnauthorizedException('Failed to extract user ID from token');
  }
});
