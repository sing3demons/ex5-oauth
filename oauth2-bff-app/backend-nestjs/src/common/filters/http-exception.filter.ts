import { ExceptionFilter, Catch, ArgumentsHost, HttpException, HttpStatus } from '@nestjs/common';
import { Request, Response } from 'express';

@Catch()
export class HttpExceptionFilter implements ExceptionFilter {
  catch(exception: unknown, host: ArgumentsHost) {
    const ctx = host.switchToHttp();
    const response = ctx.getResponse<Response>();
    const request = ctx.getRequest<Request>();

    // Determine status code
    const status =
      exception instanceof HttpException ? exception.getStatus() : HttpStatus.INTERNAL_SERVER_ERROR;

    // Determine error message
    const message =
      exception instanceof HttpException ? exception.message : 'Internal server error';

    // Determine error code
    const errorCode =
      exception instanceof HttpException ? this.getErrorCode(exception) : 'server_error';

    // Log error with stack trace
    console.error('Error occurred:', {
      error: errorCode,
      message,
      path: request.url,
      method: request.method,
      timestamp: new Date().toISOString(),
      stack: exception instanceof Error ? exception.stack : undefined,
    });

    // Return formatted error response
    response.status(status).json({
      error: errorCode,
      message,
      timestamp: new Date().toISOString(),
      path: request.url,
    });
  }

  private getErrorCode(exception: HttpException): string {
    const status = exception.getStatus();
    const response = exception.getResponse();

    // If response has an error code, use it
    if (typeof response === 'object' && 'error' in response) {
      return (response as any).error;
    }

    // Map status codes to error codes
    switch (status) {
      case HttpStatus.UNAUTHORIZED:
        return 'unauthorized';
      case HttpStatus.FORBIDDEN:
        return 'forbidden';
      case HttpStatus.NOT_FOUND:
        return 'not_found';
      case HttpStatus.BAD_REQUEST:
        return 'invalid_request';
      default:
        return exception.name || 'http_error';
    }
  }
}
