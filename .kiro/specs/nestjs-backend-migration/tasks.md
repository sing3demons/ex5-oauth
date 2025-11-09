# Implementation Plan

- [x] 1. Setup NestJS project structure and dependencies
  - Initialize NestJS project in `oauth2-bff-app/backend-nestjs` directory
  - Install required dependencies: @nestjs/core, @nestjs/common, @nestjs/config, @nestjs/platform-express, @nestjs/axios, mongodb, cookie-parser, uuid, class-validator, class-transformer
  - Install dev dependencies: @types/node, @types/cookie-parser, @types/uuid, typescript, ts-node
  - Configure tsconfig.json with strict mode and path aliases
  - Create .env.example file with all required environment variables
  - Copy .env from Express backend for initial configuration
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

- [x] 2. Implement configuration and shared utilities
  - [x] 2.1 Create configuration module
    - Create `src/config/configuration.ts` with configuration factory
    - Configure ConfigModule as global module in app.module.ts
    - Define TypeScript interfaces for configuration structure
    - _Requirements: 1.3_
  
  - [x] 2.2 Implement crypto utilities
    - Create `src/shared/services/crypto.service.ts`
    - Implement generateState() method for OAuth state generation
    - Implement generateNonce() method for OIDC nonce generation
    - Implement base64URLEncode() helper method
    - _Requirements: 2.1, 2.5_
  
  - [x] 2.3 Implement OIDC utilities
    - Create `src/shared/services/oidc.service.ts`
    - Implement decodeJWT() method to decode JWT tokens
    - Implement validateIDToken() method with issuer, audience, expiration, and nonce validation
    - Implement fetchDiscovery() method to get OIDC discovery document
    - _Requirements: 2.3, 14.1, 14.2, 14.3, 14.4, 14.5, 18.1, 18.2_
  
  - [x] 2.4 Implement token utilities
    - Create `src/shared/services/token.service.ts`
    - Implement getUserIdFromToken() method to extract user ID from JWT claims
    - Use OidcService for JWT decoding
    - _Requirements: 6.2, 7.2_
  
  - [x] 2.5 Create shared module
    - Create `src/shared/shared.module.ts`
    - Export CryptoService, OidcService, and TokenService
    - Configure as global module
    - _Requirements: 1.2_

- [x] 3. Implement database module
  - [x] 3.1 Create database service
    - Create `src/database/database.service.ts`
    - Implement onModuleInit() to connect to MongoDB on startup
    - Implement getDatabase() method to provide Db instance
    - Implement createIndexes() to create todos collection index on userId and createdAt
    - Implement onModuleDestroy() for graceful shutdown
    - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5_
  
  - [x] 3.2 Create database module
    - Create `src/database/database.module.ts`
    - Export DatabaseService as global provider
    - _Requirements: 11.5_

- [x] 4. Implement authentication guards
  - [x] 4.1 Create auth guard
    - Create `src/auth/guards/auth.guard.ts`
    - Implement CanActivate interface
    - Validate Authorization header presence and Bearer format
    - Throw UnauthorizedException if token is missing
    - _Requirements: 13.1, 13.2, 13.3, 13.4_
  
  - [x] 4.2 Create refresh token guard
    - Create `src/auth/guards/refresh.guard.ts`
    - Implement CanActivate interface
    - Validate refresh_token cookie presence
    - Throw UnauthorizedException if cookie is missing
    - _Requirements: 3.5, 13.5_

- [x] 5. Implement session management
  - [x] 5.1 Create session service
    - Create `src/auth/session.service.ts`
    - Implement in-memory Map for session storage
    - Implement createSession() method to store state, nonce, and timestamp
    - Implement getSession() method to retrieve session by state
    - Implement deleteSession() method to remove session after use
    - _Requirements: 17.1, 17.2, 17.3, 17.4_
  
  - [x] 5.2 Add session cleanup
    - Implement cleanupExpiredSessions() method
    - Use setInterval to run cleanup every 60 seconds
    - Remove sessions older than 10 minutes
    - _Requirements: 17.5_

- [x] 6. Implement auth service and DTOs
  - [x] 6.1 Create auth DTOs
    - Create `src/auth/dto/login-response.dto.ts` with authorization_url field
    - Create `src/auth/dto/token-response.dto.ts` with access_token, expires_in, token_type fields
    - Create `src/auth/dto/userinfo.dto.ts` for user claims
    - Create `src/auth/dto/validation-result.dto.ts` for token validation
    - _Requirements: 2.1, 2.5, 5.3, 19.2_
  
  - [x] 6.2 Implement auth service - login flow
    - Create `src/auth/auth.service.ts`
    - Inject ConfigService, SessionService, OidcService, CryptoService, and HttpService
    - Implement initiateLogin() method
    - Generate state and nonce using CryptoService
    - Store session with SessionService
    - Build authorization URL with response_type=code, client_id, redirect_uri, scope=openid profile email, state, nonce, response_mode=query
    - Return LoginResponseDto with authorization URL
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_
  
  - [x] 6.3 Implement auth service - callback flow
    - Implement handleCallback() method
    - Validate code and state parameters
    - Retrieve session using state
    - Delete session after retrieval
    - Exchange authorization code for tokens using POST to /oauth/token with grant_type=authorization_code, code, redirect_uri, client_id, client_secret
    - Validate ID token if present using OidcService
    - Set refresh_token in HttpOnly cookie with secure, sameSite=lax, maxAge=7 days
    - Redirect to frontend with access_token, expires_in, and id_token in URL parameters
    - _Requirements: 2.2, 2.3, 2.4, 2.5_
  
  - [x] 6.4 Implement auth service - refresh flow
    - Implement refreshToken() method
    - Exchange refresh_token for new access_token using POST to /oauth/token with grant_type=refresh_token, refresh_token, client_id, client_secret
    - Update refresh_token cookie if new token is provided (token rotation)
    - Return TokenResponseDto with access_token, expires_in, token_type
    - Clear cookie and throw UnauthorizedException if refresh fails
    - _Requirements: 3.1, 3.2, 3.3, 3.4_
  
  - [x] 6.5 Implement auth service - logout and userinfo
    - Implement logout() method to clear refresh_token cookie
    - Implement getUserInfo() method to forward request to OAuth2 server /oauth/userinfo endpoint
    - Implement getDiscovery() method to fetch .well-known/openid-configuration
    - Implement validateIDToken() method using OidcService
    - _Requirements: 4.1, 4.2, 4.3, 5.1, 5.2, 5.3, 5.4, 18.1, 18.2, 18.3, 19.1, 19.2, 19.3_

- [x] 7. Implement auth controller
  - [x] 7.1 Create auth controller endpoints - login and callback
    - Create `src/auth/auth.controller.ts`
    - Inject AuthService
    - Implement GET /auth/login endpoint calling authService.initiateLogin()
    - Implement GET /auth/callback endpoint with @Query decorators for code, state, error
    - Use @Res decorator to handle redirect response
    - _Requirements: 2.1, 2.2, 2.5_
  
  - [x] 7.2 Create auth controller endpoints - token management
    - Implement POST /auth/refresh endpoint with @UseGuards(RefreshGuard)
    - Use @Req and @Res decorators to access cookies and response
    - Implement POST /auth/logout endpoint calling authService.logout()
    - _Requirements: 3.1, 3.5, 4.1, 4.2_
  
  - [x] 7.3 Create auth controller endpoints - user info and discovery
    - Implement GET /auth/userinfo endpoint with @UseGuards(AuthGuard)
    - Use @Headers decorator to get authorization header
    - Implement GET /auth/discovery endpoint calling authService.getDiscovery()
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 18.1, 18.4, 18.5_
  
  - [x] 7.4 Create auth controller endpoints - token utilities
    - Implement POST /auth/validate-token endpoint with @Body decorator for id_token
    - Implement POST /auth/decode-token endpoint with @Body decorator for token
    - Implement GET /auth/session endpoint with @UseGuards(RefreshGuard)
    - Use @Req decorator to access refresh_token cookie
    - Decode token and return session info with expiration, user_id, scope
    - _Requirements: 19.1, 19.2, 19.3, 19.4, 19.5, 20.1, 20.2, 20.3, 20.4, 20.5_

- [x] 8. Create auth module
  - Create `src/auth/auth.module.ts`
  - Import ConfigModule, HttpModule, SharedModule
  - Provide AuthService, SessionService
  - Export AuthGuard and RefreshGuard
  - Register AuthController
  - _Requirements: 1.2_

- [x] 9. Implement todo entities and DTOs
  - [x] 9.1 Create todo entity
    - Create `src/todos/entities/todo.entity.ts`
    - Define Todo interface with id, userId, title, description, status, priority, createdAt, updatedAt
    - Define TodoStatus type as 'todo' | 'in_progress' | 'done'
    - Define TodoPriority type as 'low' | 'medium' | 'high'
    - _Requirements: 6.4, 7.4, 8.4, 10.4_
  
  - [x] 9.2 Create todo DTOs
    - Create `src/todos/dto/create-todo.dto.ts` with validation decorators
    - Add @IsString() @IsNotEmpty() for title
    - Add @IsString() @IsOptional() for description
    - Add @IsEnum(['low', 'medium', 'high']) @IsOptional() for priority
    - Create `src/todos/dto/update-todo.dto.ts` with optional fields
    - Create `src/todos/dto/update-status.dto.ts` with @IsEnum(['todo', 'in_progress', 'done']) for status
    - _Requirements: 6.1, 6.4, 8.1, 8.3, 10.2_

- [x] 10. Implement todos service
  - [x] 10.1 Create todos service - query operations
    - Create `src/todos/todos.service.ts`
    - Inject DatabaseService and TokenService
    - Implement findAllByUser() method to query todos by userId sorted by createdAt descending
    - Implement findOne() method to get specific todo by id
    - Implement verifyOwnership() private method to check if todo belongs to user
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 8.2, 9.2_
  
  - [x] 10.2 Create todos service - create and update operations
    - Implement create() method to generate UUID, set userId, status=todo, priority=medium, timestamps
    - Insert todo into MongoDB todos collection
    - Implement update() method to validate ownership and update allowed fields
    - Set updatedAt timestamp on updates
    - Validate title is not empty when updating
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 8.1, 8.2, 8.3, 8.4, 8.5_
  
  - [x] 10.3 Create todos service - delete and status operations
    - Implement remove() method to verify ownership and delete from MongoDB
    - Implement updateStatus() method to validate status enum and update todo
    - Verify ownership before status update
    - Set updatedAt timestamp on status change
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 10.1, 10.2, 10.3, 10.4, 10.5_

- [x] 11. Implement user decorator
  - Create `src/common/decorators/user.decorator.ts`
  - Use createParamDecorator to extract user ID from Authorization header
  - Decode JWT and return sub or user_id claim
  - Throw UnauthorizedException if token is missing
  - _Requirements: 6.2, 7.2_

- [x] 12. Implement todos controller
  - [x] 12.1 Create todos controller - query endpoints
    - Create `src/todos/todos.controller.ts`
    - Add @Controller('api/todos') and @UseGuards(AuthGuard) decorators
    - Inject TodosService
    - Implement GET / endpoint using @User() decorator to get userId
    - Implement GET /:id endpoint with @Param('id') decorator
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_
  
  - [x] 12.2 Create todos controller - mutation endpoints
    - Implement POST / endpoint with @Body() CreateTodoDto and @User() decorator
    - Return 201 status code for created todo
    - Implement PUT /:id endpoint with @Param('id'), @Body() UpdateTodoDto, and @User() decorator
    - Implement DELETE /:id endpoint with @HttpCode(204) decorator
    - _Requirements: 6.1, 6.5, 8.1, 8.5, 9.1, 9.4_
  
  - [x] 12.3 Create todos controller - status endpoint
    - Implement PATCH /:id/status endpoint with @Param('id'), @Body() UpdateStatusDto, and @User() decorator
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5_

- [x] 13. Create todos module
  - Create `src/todos/todos.module.ts`
  - Import DatabaseModule, SharedModule
  - Provide TodosService
  - Register TodosController
  - _Requirements: 1.2_

- [x] 14. Implement global error handling
  - Create `src/common/filters/http-exception.filter.ts`
  - Implement ExceptionFilter interface with @Catch() decorator
  - Log errors with stack trace
  - Return formatted error response with error code, message, timestamp, path
  - Use appropriate status codes (default to 500)
  - _Requirements: 16.1, 16.2, 16.3, 16.4, 16.5_

- [x] 15. Setup app module and main entry point
  - [x] 15.1 Create app controller
    - Create `src/app.controller.ts`
    - Implement GET /health endpoint returning status=ok and timestamp
    - _Requirements: 15.1, 15.2, 15.3, 15.4, 15.5_
  
  - [x] 15.2 Create app module
    - Create `src/app.module.ts`
    - Import ConfigModule.forRoot with isGlobal=true
    - Import AuthModule, TodosModule, DatabaseModule, SharedModule
    - Register AppController
    - Provide HttpExceptionFilter as APP_FILTER
    - _Requirements: 1.1, 1.2, 1.3_
  
  - [x] 15.3 Create main entry point
    - Create `src/main.ts`
    - Bootstrap NestJS application
    - Enable CORS with origin from config, credentials=true, allowed methods and headers
    - Use cookie-parser middleware
    - Enable validation pipe globally with whitelist=true, transform=true
    - Listen on configured port
    - Log startup information including port, frontend URL, OAuth2 server URL, client ID
    - _Requirements: 12.1, 12.2, 12.3, 12.4, 12.5_

- [x] 16. Configure build and development scripts
  - Update package.json with scripts: start:dev, start:debug, start:prod, build, format, lint
  - Configure nest-cli.json for build options
  - Add .gitignore for dist/, node_modules/, .env
  - _Requirements: 1.1, 1.4_

- [x] 17. Create documentation and setup guide
  - Create README.md with setup instructions, environment variables, and API endpoints
  - Create MIGRATION.md documenting differences from Express version
  - Update main oauth2-bff-app README to mention NestJS version
  - _Requirements: 1.1_

- [x] 18. Test with existing frontend
  - Start NestJS backend on port 3001
  - Verify OAuth2 login flow works
  - Verify token refresh works
  - Verify todo CRUD operations work
  - Verify drag & drop status updates work
  - Verify error handling works correctly
  - _Requirements: All requirements_
