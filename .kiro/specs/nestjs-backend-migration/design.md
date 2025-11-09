# Design Document - NestJS Backend Migration

## Overview

เอกสารนี้อธิบาย design สำหรับการสร้าง Backend for Frontend (BFF) application ด้วย NestJS framework เพื่อทดแทน Express version ปัจจุบัน โดยจะใช้ NestJS architecture patterns เช่น Modules, Controllers, Services, Guards, และ Interceptors เพื่อสร้าง scalable และ maintainable codebase

Application นี้จะทำหน้าที่เป็น BFF layer ระหว่าง React frontend และ OAuth2 server โดยจัดการ OAuth2/OIDC authentication flow, session management, และ Todo CRUD operations พร้อม MongoDB integration

## Architecture

### High-Level Architecture

```
┌─────────────────┐
│  React Frontend │
│  (Port 5173)    │
└────────┬────────┘
         │ HTTP/HTTPS
         │ (CORS enabled)
         ▼
┌─────────────────────────────────────┐
│      NestJS BFF Backend             │
│         (Port 3001)                 │
│                                     │
│  ┌──────────────────────────────┐  │
│  │      Auth Module             │  │
│  │  - AuthController            │  │
│  │  - AuthService               │  │
│  │  - SessionService            │  │
│  │  - Guards (Auth, Refresh)    │  │
│  └──────────────────────────────┘  │
│                                     │
│  ┌──────────────────────────────┐  │
│  │      Todos Module            │  │
│  │  - TodosController           │  │
│  │  - TodosService              │  │
│  └──────────────────────────────┘  │
│                                     │
│  ┌──────────────────────────────┐  │
│  │      Database Module         │  │
│  │  - MongoDBService            │  │
│  └──────────────────────────────┘  │
│                                     │
│  ┌──────────────────────────────┐  │
│  │      Shared Module           │  │
│  │  - OidcService               │  │
│  │  - TokenService              │  │
│  │  - CryptoService             │  │
│  └──────────────────────────────┘  │
└────────┬────────────────────┬──────┘
         │                    │
         │ OAuth2/OIDC        │ MongoDB
         ▼                    ▼
┌─────────────────┐   ┌──────────────┐
│  OAuth2 Server  │   │   MongoDB    │
│  (Port 8080)    │   │ (Port 27017) │
└─────────────────┘   └──────────────┘
```

### Module Structure

```
src/
├── main.ts                      # Application entry point
├── app.module.ts                # Root module
├── config/
│   └── configuration.ts         # Configuration factory
├── auth/
│   ├── auth.module.ts
│   ├── auth.controller.ts       # OAuth2/OIDC endpoints
│   ├── auth.service.ts          # OAuth2 flow logic
│   ├── session.service.ts       # Session management
│   ├── guards/
│   │   ├── auth.guard.ts        # Access token validation
│   │   └── refresh.guard.ts     # Refresh token validation
│   └── dto/
│       ├── login-response.dto.ts
│       ├── token-response.dto.ts
│       └── userinfo.dto.ts
├── todos/
│   ├── todos.module.ts
│   ├── todos.controller.ts      # Todo CRUD endpoints
│   ├── todos.service.ts         # Todo business logic
│   ├── entities/
│   │   └── todo.entity.ts
│   └── dto/
│       ├── create-todo.dto.ts
│       ├── update-todo.dto.ts
│       └── update-status.dto.ts
├── database/
│   ├── database.module.ts
│   └── database.service.ts      # MongoDB connection
├── shared/
│   ├── shared.module.ts
│   ├── services/
│   │   ├── oidc.service.ts      # OIDC utilities
│   │   ├── token.service.ts     # Token validation
│   │   └── crypto.service.ts    # Crypto utilities
│   └── filters/
│       └── http-exception.filter.ts
└── common/
    ├── decorators/
    │   └── user.decorator.ts    # Extract user from token
    └── interfaces/
        └── session.interface.ts
```

## Components and Interfaces

### 1. App Module (Root Module)

**Purpose**: รวม modules ทั้งหมดและ configure global settings

**Dependencies**:
- `@nestjs/config` - Environment configuration
- `@nestjs/common` - Common utilities
- `@nestjs/platform-express` - Express platform

**Configuration**:
```typescript
@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
      load: [configuration],
    }),
    AuthModule,
    TodosModule,
    DatabaseModule,
    SharedModule,
  ],
  controllers: [AppController],
  providers: [
    {
      provide: APP_FILTER,
      useClass: HttpExceptionFilter,
    },
  ],
})
export class AppModule {}
```

### 2. Auth Module

#### AuthController

**Endpoints**:

| Method | Path | Description | Auth Required |
|--------|------|-------------|---------------|
| GET | `/auth/login` | Initiate OAuth2 flow | No |
| GET | `/auth/callback` | Handle OAuth2 callback | No |
| POST | `/auth/refresh` | Refresh access token | Refresh Token |
| POST | `/auth/logout` | Clear session | No |
| GET | `/auth/userinfo` | Get user info | Access Token |
| GET | `/auth/discovery` | Get OIDC discovery | No |
| POST | `/auth/validate-token` | Validate ID token | No |
| POST | `/auth/decode-token` | Decode JWT | No |
| GET | `/auth/session` | Get session info | Refresh Token |

**Implementation Pattern**:
```typescript
@Controller('auth')
export class AuthController {
  constructor(
    private readonly authService: AuthService,
    private readonly sessionService: SessionService,
  ) {}

  @Get('login')
  async login(): Promise<LoginResponseDto> {
    return this.authService.initiateLogin();
  }

  @Get('callback')
  async callback(
    @Query('code') code: string,
    @Query('state') state: string,
    @Query('error') error: string,
    @Res() res: Response,
  ): Promise<void> {
    return this.authService.handleCallback(code, state, error, res);
  }

  @Post('refresh')
  @UseGuards(RefreshGuard)
  async refresh(
    @Req() req: Request,
    @Res() res: Response,
  ): Promise<TokenResponseDto> {
    return this.authService.refreshToken(req.cookies.refresh_token, res);
  }

  @Post('logout')
  async logout(@Res() res: Response): Promise<void> {
    return this.authService.logout(res);
  }

  @Get('userinfo')
  @UseGuards(AuthGuard)
  async getUserInfo(@Headers('authorization') auth: string): Promise<UserInfoDto> {
    return this.authService.getUserInfo(auth);
  }
}
```

#### AuthService

**Responsibilities**:
- Generate authorization URLs with state and nonce
- Exchange authorization code for tokens
- Validate ID tokens
- Refresh access tokens
- Manage refresh token cookies
- Forward userinfo requests

**Key Methods**:
```typescript
export class AuthService {
  async initiateLogin(): Promise<LoginResponseDto>
  async handleCallback(code: string, state: string, error: string, res: Response): Promise<void>
  async refreshToken(refreshToken: string, res: Response): Promise<TokenResponseDto>
  async logout(res: Response): Promise<void>
  async getUserInfo(authHeader: string): Promise<UserInfoDto>
  async getDiscovery(): Promise<DiscoveryDocument>
  async validateIDToken(idToken: string): Promise<ValidationResult>
}
```

#### SessionService

**Responsibilities**:
- Store OAuth state and nonce in memory
- Retrieve and validate sessions
- Clean up expired sessions

**Data Structure**:
```typescript
interface SessionData {
  redirect_uri: string;
  nonce: string;
  timestamp: number;
}

export class SessionService {
  private sessions: Map<string, SessionData> = new Map();
  
  createSession(state: string, data: SessionData): void
  getSession(state: string): SessionData | undefined
  deleteSession(state: string): void
  cleanupExpiredSessions(): void
}
```

**Session Cleanup**: ทำงานทุก 60 วินาทีเพื่อลบ sessions ที่เก่ากว่า 10 นาที

#### AuthGuard

**Purpose**: Validate access token in Authorization header

**Implementation**:
```typescript
@Injectable()
export class AuthGuard implements CanActivate {
  canActivate(context: ExecutionContext): boolean {
    const request = context.switchToHttp().getRequest();
    const authHeader = request.headers.authorization;
    
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      throw new UnauthorizedException('No access token provided');
    }
    
    return true;
  }
}
```

#### RefreshGuard

**Purpose**: Validate refresh token in HttpOnly cookie

**Implementation**:
```typescript
@Injectable()
export class RefreshGuard implements CanActivate {
  canActivate(context: ExecutionContext): boolean {
    const request = context.switchToHttp().getRequest();
    const refreshToken = request.cookies?.refresh_token;
    
    if (!refreshToken) {
      throw new UnauthorizedException('No refresh token found');
    }
    
    return true;
  }
}
```

### 3. Todos Module

#### TodosController

**Endpoints**:

| Method | Path | Description | Auth Required |
|--------|------|-------------|---------------|
| GET | `/api/todos` | Get all user todos | Access Token |
| GET | `/api/todos/:id` | Get specific todo | Access Token |
| POST | `/api/todos` | Create new todo | Access Token |
| PUT | `/api/todos/:id` | Update todo | Access Token |
| DELETE | `/api/todos/:id` | Delete todo | Access Token |
| PATCH | `/api/todos/:id/status` | Update status | Access Token |

**Implementation Pattern**:
```typescript
@Controller('api/todos')
@UseGuards(AuthGuard)
export class TodosController {
  constructor(private readonly todosService: TodosService) {}

  @Get()
  async findAll(@User() userId: string): Promise<Todo[]> {
    return this.todosService.findAllByUser(userId);
  }

  @Post()
  async create(
    @User() userId: string,
    @Body() createTodoDto: CreateTodoDto,
  ): Promise<Todo> {
    return this.todosService.create(userId, createTodoDto);
  }

  @Put(':id')
  async update(
    @User() userId: string,
    @Param('id') id: string,
    @Body() updateTodoDto: UpdateTodoDto,
  ): Promise<Todo> {
    return this.todosService.update(userId, id, updateTodoDto);
  }

  @Delete(':id')
  @HttpCode(204)
  async remove(
    @User() userId: string,
    @Param('id') id: string,
  ): Promise<void> {
    return this.todosService.remove(userId, id);
  }

  @Patch(':id/status')
  async updateStatus(
    @User() userId: string,
    @Param('id') id: string,
    @Body() updateStatusDto: UpdateStatusDto,
  ): Promise<Todo> {
    return this.todosService.updateStatus(userId, id, updateStatusDto.status);
  }
}
```

#### TodosService

**Responsibilities**:
- CRUD operations for todos
- Validate todo ownership
- Extract user ID from tokens
- Interact with MongoDB

**Key Methods**:
```typescript
export class TodosService {
  constructor(
    private readonly databaseService: DatabaseService,
    private readonly tokenService: TokenService,
  ) {}

  async findAllByUser(userId: string): Promise<Todo[]>
  async findOne(userId: string, id: string): Promise<Todo>
  async create(userId: string, createTodoDto: CreateTodoDto): Promise<Todo>
  async update(userId: string, id: string, updateTodoDto: UpdateTodoDto): Promise<Todo>
  async remove(userId: string, id: string): Promise<void>
  async updateStatus(userId: string, id: string, status: TodoStatus): Promise<Todo>
  
  private async verifyOwnership(userId: string, todoId: string): Promise<Todo>
}
```

#### Todo Entity

```typescript
export interface Todo {
  id: string;              // UUID
  userId: string;          // From token claims
  title: string;
  description?: string;
  status: 'todo' | 'in_progress' | 'done';
  priority: 'low' | 'medium' | 'high';
  createdAt: Date;
  updatedAt: Date;
}
```

#### DTOs

**CreateTodoDto**:
```typescript
export class CreateTodoDto {
  @IsString()
  @IsNotEmpty()
  title: string;

  @IsString()
  @IsOptional()
  description?: string;

  @IsEnum(['low', 'medium', 'high'])
  @IsOptional()
  priority?: 'low' | 'medium' | 'high';
}
```

**UpdateTodoDto**:
```typescript
export class UpdateTodoDto {
  @IsString()
  @IsOptional()
  title?: string;

  @IsString()
  @IsOptional()
  description?: string;

  @IsEnum(['todo', 'in_progress', 'done'])
  @IsOptional()
  status?: 'todo' | 'in_progress' | 'done';

  @IsEnum(['low', 'medium', 'high'])
  @IsOptional()
  priority?: 'low' | 'medium' | 'high';
}
```

**UpdateStatusDto**:
```typescript
export class UpdateStatusDto {
  @IsEnum(['todo', 'in_progress', 'done'])
  status: 'todo' | 'in_progress' | 'done';
}
```

### 4. Database Module

#### DatabaseService

**Responsibilities**:
- Manage MongoDB connection
- Provide database instance to other modules
- Create indexes
- Handle graceful shutdown

**Implementation**:
```typescript
@Injectable()
export class DatabaseService implements OnModuleInit, OnModuleDestroy {
  private client: MongoClient;
  private db: Db;

  constructor(private configService: ConfigService) {}

  async onModuleInit(): Promise<void> {
    await this.connect();
  }

  async onModuleDestroy(): Promise<void> {
    await this.disconnect();
  }

  private async connect(): Promise<void> {
    const uri = this.configService.get<string>('MONGODB_URI');
    const dbName = this.configService.get<string>('MONGODB_DB');
    
    this.client = new MongoClient(uri);
    await this.client.connect();
    this.db = this.client.db(dbName);
    
    await this.createIndexes();
  }

  private async createIndexes(): Promise<void> {
    await this.db.collection('todos').createIndex({ 
      userId: 1, 
      createdAt: -1 
    });
  }

  getDatabase(): Db {
    if (!this.db) {
      throw new Error('Database not initialized');
    }
    return this.db;
  }

  async disconnect(): Promise<void> {
    if (this.client) {
      await this.client.close();
    }
  }
}
```

### 5. Shared Module

#### OidcService

**Responsibilities**:
- Validate ID tokens
- Decode JWTs
- Generate nonce
- Fetch discovery documents

**Key Methods**:
```typescript
@Injectable()
export class OidcService {
  validateIDToken(
    idToken: string,
    clientId: string,
    issuer: string,
    nonce?: string,
  ): ValidationResult
  
  decodeJWT(token: string): any
  
  generateNonce(): string
  
  async fetchDiscovery(issuer: string): Promise<DiscoveryDocument>
}
```

**Validation Logic**:
1. Decode JWT
2. Verify issuer matches OAuth2 server
3. Verify audience matches client ID
4. Verify expiration (exp > now)
5. Verify issued at (iat <= now + 60s)
6. Verify nonce if provided

#### TokenService

**Responsibilities**:
- Extract user ID from tokens
- Decode tokens for user info

**Implementation**:
```typescript
@Injectable()
export class TokenService {
  constructor(private readonly oidcService: OidcService) {}

  getUserIdFromToken(authHeader: string): string | null {
    try {
      const token = authHeader.replace('Bearer ', '');
      const claims = this.oidcService.decodeJWT(token);
      return claims.sub || claims.user_id || null;
    } catch {
      return null;
    }
  }
}
```

#### CryptoService

**Responsibilities**:
- Generate random state
- Generate nonce
- Base64 URL encoding

**Implementation**:
```typescript
@Injectable()
export class CryptoService {
  generateState(): string {
    return this.base64URLEncode(crypto.randomBytes(32));
  }

  generateNonce(): string {
    return crypto.randomBytes(32).toString('base64url');
  }

  private base64URLEncode(buffer: Buffer): string {
    return buffer
      .toString('base64')
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=/g, '');
  }
}
```

### 6. Common Utilities

#### User Decorator

**Purpose**: Extract user ID from access token

**Implementation**:
```typescript
export const User = createParamDecorator(
  (data: unknown, ctx: ExecutionContext): string => {
    const request = ctx.switchToHttp().getRequest();
    const authHeader = request.headers.authorization;
    
    if (!authHeader) {
      throw new UnauthorizedException('No access token');
    }
    
    const token = authHeader.replace('Bearer ', '');
    const claims = decodeJWT(token);
    return claims.sub || claims.user_id;
  },
);
```

#### HttpExceptionFilter

**Purpose**: Global error handling

**Implementation**:
```typescript
@Catch()
export class HttpExceptionFilter implements ExceptionFilter {
  catch(exception: unknown, host: ArgumentsHost) {
    const ctx = host.switchToHttp();
    const response = ctx.getResponse<Response>();
    const request = ctx.getRequest<Request>();

    const status =
      exception instanceof HttpException
        ? exception.getStatus()
        : HttpStatus.INTERNAL_SERVER_ERROR;

    const message =
      exception instanceof HttpException
        ? exception.message
        : 'Internal server error';

    console.error('Error:', exception);

    response.status(status).json({
      error: exception instanceof HttpException 
        ? exception.name 
        : 'server_error',
      message,
      timestamp: new Date().toISOString(),
      path: request.url,
    });
  }
}
```

## Data Models

### Todo Collection (MongoDB)

```typescript
{
  id: string,              // UUID v4
  userId: string,          // From JWT claims (sub or user_id)
  title: string,           // Required, non-empty
  description?: string,    // Optional
  status: 'todo' | 'in_progress' | 'done',
  priority: 'low' | 'medium' | 'high',
  createdAt: Date,
  updatedAt: Date
}
```

**Indexes**:
- `{ userId: 1, createdAt: -1 }` - For efficient user todo queries

### Session Store (In-Memory Map)

```typescript
Map<string, {
  redirect_uri: string,
  nonce: string,
  timestamp: number
}>
```

**Key**: OAuth state parameter
**TTL**: 10 minutes
**Cleanup**: Every 60 seconds

## Error Handling

### Error Response Format

```typescript
{
  error: string,        // Error code
  message: string,      // Human-readable message
  timestamp: string,    // ISO timestamp
  path: string         // Request path
}
```

### Common Error Codes

| Code | Status | Description |
|------|--------|-------------|
| `unauthorized` | 401 | Missing or invalid token |
| `forbidden` | 403 | User not authorized for resource |
| `not_found` | 404 | Resource not found |
| `invalid_request` | 400 | Invalid request parameters |
| `invalid_grant` | 401 | Invalid or expired refresh token |
| `invalid_state` | 400 | Invalid OAuth state |
| `invalid_id_token` | 401 | ID token validation failed |
| `server_error` | 500 | Internal server error |

### Exception Handling Strategy

1. **Controller Level**: Validate DTOs using class-validator
2. **Service Level**: Throw appropriate HttpExceptions
3. **Global Level**: Catch all exceptions with HttpExceptionFilter
4. **Logging**: Log all errors with stack traces

## Testing Strategy

### Unit Tests

**Auth Module**:
- AuthService methods (login, callback, refresh, logout)
- SessionService (create, get, delete, cleanup)
- Guards (AuthGuard, RefreshGuard)
- OidcService (token validation, JWT decode)

**Todos Module**:
- TodosService CRUD operations
- Ownership verification
- Status updates

**Database Module**:
- Connection management
- Index creation

**Shared Module**:
- Token extraction
- Crypto utilities

### Integration Tests

**Auth Flow**:
- Complete OAuth2 authorization code flow
- Token refresh flow
- Logout flow
- ID token validation

**Todo Operations**:
- Create, read, update, delete todos
- Status updates via drag & drop
- Authorization checks

**Error Scenarios**:
- Invalid tokens
- Missing cookies
- Unauthorized access
- Database errors

### E2E Tests

- Full authentication flow from frontend to backend
- Todo CRUD operations with real MongoDB
- Session management
- CORS configuration

### Test Tools

- Jest (unit and integration tests)
- Supertest (HTTP testing)
- MongoDB Memory Server (database testing)

## Security Considerations

### 1. Token Storage

- **Refresh Token**: Stored in HttpOnly cookie with secure, sameSite flags
- **Access Token**: Never stored in backend, passed from frontend
- **ID Token**: Validated but not stored

### 2. CORS Configuration

- Allow only configured frontend URL
- Enable credentials for cookie support
- Restrict allowed methods and headers

### 3. Cookie Security

```typescript
{
  httpOnly: true,           // Prevent JavaScript access
  secure: NODE_ENV === 'production',  // HTTPS only in production
  sameSite: 'lax',         // CSRF protection
  maxAge: 7 * 24 * 60 * 60 * 1000,   // 7 days
  path: '/'
}
```

### 4. Input Validation

- Use class-validator for DTO validation
- Sanitize user inputs
- Validate enum values

### 5. Authorization

- Verify todo ownership before operations
- Extract user ID from validated tokens
- Use guards for endpoint protection

### 6. Session Management

- Store minimal data in sessions
- Clean up expired sessions
- Use cryptographically secure random values

### 7. Error Messages

- Don't expose sensitive information
- Use generic error messages for security errors
- Log detailed errors server-side only

## Configuration

### Environment Variables

```env
# Server
PORT=3001
NODE_ENV=development

# OAuth2
OAUTH2_SERVER_URL=http://localhost:8080
CLIENT_ID=your-client-id
CLIENT_SECRET=your-client-secret

# Frontend
FRONTEND_URL=http://localhost:5173

# Database
MONGODB_URI=mongodb://localhost:27017
MONGODB_DB=oauth2_bff_app
```

### Configuration Module

```typescript
export default () => ({
  port: parseInt(process.env.PORT, 10) || 3001,
  nodeEnv: process.env.NODE_ENV || 'development',
  oauth2: {
    serverUrl: process.env.OAUTH2_SERVER_URL,
    clientId: process.env.CLIENT_ID,
    clientSecret: process.env.CLIENT_SECRET,
  },
  frontend: {
    url: process.env.FRONTEND_URL,
  },
  database: {
    uri: process.env.MONGODB_URI,
    name: process.env.MONGODB_DB,
  },
});
```

## Deployment Considerations

### Production Checklist

1. Set `NODE_ENV=production`
2. Use HTTPS for all connections
3. Set secure cookie flags
4. Use production MongoDB instance
5. Configure proper CORS origins
6. Enable request logging
7. Set up health checks
8. Configure graceful shutdown
9. Use environment-specific secrets
10. Enable rate limiting

### Health Check

```typescript
@Controller()
export class AppController {
  @Get('health')
  health() {
    return {
      status: 'ok',
      timestamp: new Date().toISOString(),
    };
  }
}
```

### Graceful Shutdown

- Close MongoDB connections
- Finish pending requests
- Clean up resources

## Migration Path from Express

### Phase 1: Setup
1. Create NestJS project structure
2. Install dependencies
3. Configure environment

### Phase 2: Core Modules
1. Implement Database Module
2. Implement Shared Module (utilities)
3. Set up configuration

### Phase 3: Auth Module
1. Implement AuthService
2. Implement SessionService
3. Create Guards
4. Add AuthController endpoints

### Phase 4: Todos Module
1. Implement TodosService
2. Create DTOs
3. Add TodosController endpoints

### Phase 5: Testing & Validation
1. Write unit tests
2. Write integration tests
3. Test with existing frontend
4. Validate all flows

### Phase 6: Deployment
1. Update documentation
2. Configure production environment
3. Deploy alongside Express version
4. Gradual migration
5. Deprecate Express version

## Performance Optimizations

1. **Connection Pooling**: MongoDB connection reuse
2. **Caching**: Consider Redis for session storage in production
3. **Compression**: Enable gzip compression
4. **Request Validation**: Early validation to reject bad requests
5. **Async Operations**: Use async/await throughout
6. **Database Indexes**: Optimize queries with proper indexes

## Monitoring and Logging

### Logging Strategy

- Use NestJS Logger
- Log levels: error, warn, log, debug, verbose
- Structured logging for production
- Mask sensitive data (tokens, secrets)

### Metrics to Monitor

- Request rate and latency
- Error rates by endpoint
- Database query performance
- OAuth2 flow success rate
- Token refresh rate
- Active sessions count

## API Compatibility

The NestJS backend will maintain 100% API compatibility with the Express version:

- Same endpoint paths
- Same request/response formats
- Same error codes
- Same cookie behavior
- Same CORS configuration

This ensures the existing React frontend works without any changes.
