import { NestFactory } from '@nestjs/core';
import { ValidationPipe } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import * as cookieParser from 'cookie-parser';
import { AppModule } from './app.module';

async function bootstrap() {
  const app = await NestFactory.create(AppModule);

  // Get configuration
  const configService = app.get(ConfigService);
  const port = configService.get<number>('port') || 3001;
  const frontendUrl = configService.get<string>('frontend.url');
  const oauth2ServerUrl = configService.get<string>('oauth2.serverUrl');
  const clientId = configService.get<string>('oauth2.clientId');

  // Enable CORS
  app.enableCors({
    origin: frontendUrl,
    credentials: true,
    methods: ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS'],
    allowedHeaders: ['Content-Type', 'Authorization'],
  });

  // Use cookie parser middleware
  app.use(cookieParser());

  // Enable global validation pipe
  app.useGlobalPipes(
    new ValidationPipe({
      whitelist: true,
      transform: true,
    }),
  );

  // Start server
  await app.listen(port);

  // Log startup information
  console.log(`🚀 NestJS BFF Backend started successfully`);
  console.log(`📡 Server listening on port: ${port}`);
  console.log(`🌐 Frontend URL: ${frontendUrl}`);
  console.log(`🔐 OAuth2 Server URL: ${oauth2ServerUrl}`);
  console.log(`🆔 Client ID: ${clientId}`);
  console.log(`✅ Health check available at: http://localhost:${port}/health`);
}
bootstrap();
