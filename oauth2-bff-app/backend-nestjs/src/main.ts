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

  // Enable CORS with localhost wildcard for development
  const corsOrigins = [frontendUrl, oauth2ServerUrl];
  console.log('🔒 CORS allowed origins:', corsOrigins);

  app.enableCors({
    origin: (origin, callback) => {
      // Allow requests with no origin (like mobile apps or curl requests)
      if (!origin) {
        console.log('✅ CORS: Allowing request with no origin');
        return callback(null, true);
      }

      // For development: allow localhost origins
      if (origin.startsWith('http://localhost:') || origin.startsWith('http://127.0.0.1:')) {
        console.log(`✅ CORS: Allowing localhost origin: ${origin}`);
        return callback(null, true);
      }

      // Check if origin is in allowed list
      if (corsOrigins.includes(origin)) {
        console.log(`✅ CORS: Allowing origin: ${origin}`);
        callback(null, true);
      } else {
        console.warn(`❌ CORS blocked origin: ${origin}`);
        console.warn(`   Allowed origins: ${corsOrigins.join(', ')}`);
        callback(new Error('Not allowed by CORS'));
      }
    },
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
