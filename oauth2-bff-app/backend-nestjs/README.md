# OAuth2 BFF Backend - NestJS

A modern, production-ready Backend for Frontend (BFF) implementation using NestJS framework with TypeScript. This application handles OAuth2/OIDC authentication flows and provides secure API endpoints for a React frontend.

## Features

- 🚀 **NestJS Framework** - Scalable, maintainable architecture with dependency injection
- 🔐 **OAuth2/OIDC Authentication** - Full authorization code flow with PKCE support
- 🍪 **Secure Token Management** - HttpOnly cookies for refresh tokens
- ✅ **ID Token Validation** - Complete OIDC compliance with issuer, audience, nonce validation
- 🔄 **Automatic Token Refresh** - Seamless token rotation
- 📝 **Todo CRUD Operations** - Complete REST API with MongoDB
- 🛡️ **Guards & Decorators** - Route protection and user extraction
- 🎯 **TypeScript** - Full type safety throughout
- 📊 **MongoDB Integration** - Efficient data persistence
- 🧪 **Testing Ready** - Jest configuration included

## Architecture

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
│  │  - OAuth2/OIDC flows         │  │
│  │  - Session management        │  │
│  │  - Token validation          │  │
│  └──────────────────────────────┘  │
│                                     │
│  ┌──────────────────────────────┐  │
│  │      Todos Module            │  │
│  │  - CRUD operations           │  │
│  │  - User authorization        │  │
│  └──────────────────────────────┘  │
│                                     │
│  ┌──────────────────────────────┐  │
│  │      Database Module         │  │
│  │  - MongoDB connection        │  │
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

## Prerequisites

- Node.js 18+ and npm
- MongoDB 6+
- OAuth2 Server running on port 8080

## Installation

```bash
# Install dependencies
npm install

# Copy environment file
cp .env.example .env

# Edit .env with your configuration
```

## Environment Variables

Create a `.env` file in the root directory:

```env
# Server Configuration
PORT=3001
NODE_ENV=development

# OAuth2 Server Configuration
OAUTH2_SERVER_URL=http://localhost:8080
CLIENT_ID=your-client-id
CLIENT_SECRET=your-client-secret

# Frontend Configuration
FRONTEND_URL=http://localhost:5173

# Database Configuration
MONGODB_URI=mongodb://localhost:27017
MONGODB_DB=oauth2_bff_app

# Session Configuration (optional)
SESSION_SECRET=change-this-to-a-random-secret-key-in-production
```

### Environment Variable Details

| Variable | Description | Example |
|----------|-------------|---------|
| `PORT` | Port for NestJS server | `3001` |
| `NODE_ENV` | Environment mode | `development` or `production` |
| `OAUTH2_SERVER_URL` | OAuth2 authorization server URL | `http://localhost:8080` |
| `CLIENT_ID` | OAuth2 client identifier | `your-client-id` |
| `CLIENT_SECRET` | OAuth2 client secret (confidential client) | `your-client-secret` |
| `FRONTEND_URL` | React frontend URL for CORS | `http://localhost:5173` |
| `MONGODB_URI` | MongoDB connection string | `mongodb://localhost:27017` |
| `MONGODB_DB` | MongoDB database name | `oauth2_bff_app` |

## Running the Application

### Development Mode

```bash
# Start with hot-reload
npm run start:dev

# Start with debug mode
npm run start:debug
```

### Production Mode

```bash
# Build the application
npm run build

# Start production server
npm run start:prod
```

### Other Commands

```bash
# Format code
npm run format

# Lint code
npm run lint

# Run tests
npm run test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:cov
```

## API Endpoints

### Authentication Endpoints

#### `GET /auth/login`
Initiates OAuth2 authorization code flow.

**Response:**
```json
{
  "authorization_url": "http://localhost:8080/oauth/authorize?..."
}
```

#### `GET /auth/callback`
Handles OAuth2 callback and exchanges code for tokens.

**Query Parameters:**
- `code` - Authorization code
- `state` - CSRF protection state
- `error` - Error code (if any)

**Response:** Redirects to frontend with tokens in URL parameters

#### `POST /auth/refresh`
Refreshes access token using refresh token from HttpOnly cookie.

**Authentication:** Requires `refresh_token` cookie

**Response:**
```json
{
  "access_token": "eyJhbGc...",
  "expires_in": 900,
  "token_type": "Bearer"
}
```

#### `POST /auth/logout`
Clears refresh token cookie.

**Response:**
```json
{
  "message": "Logged out successfully"
}
```

#### `GET /auth/userinfo`
Retrieves user information from OAuth2 server.

**Authentication:** Requires `Authorization: Bearer <token>` header

**Response:**
```json
{
  "sub": "user-id",
  "email": "user@example.com",
  "name": "User Name",
  ...
}
```

#### `GET /auth/discovery`
Fetches OIDC discovery document.

**Response:**
```json
{
  "issuer": "http://localhost:8080",
  "authorization_endpoint": "...",
  "token_endpoint": "...",
  ...
}
```

#### `POST /auth/validate-token`
Validates an ID token.

**Request Body:**
```json
{
  "id_token": "eyJhbGc..."
}
```

**Response:**
```json
{
  "valid": true,
  "claims": { ... }
}
```

#### `POST /auth/decode-token`
Decodes a JWT token without validation.

**Request Body:**
```json
{
  "token": "eyJhbGc..."
}
```

**Response:**
```json
{
  "header": { ... },
  "payload": { ... }
}
```

#### `GET /auth/session`
Gets current session information.

**Authentication:** Requires `refresh_token` cookie

**Response:**
```json
{
  "user_id": "user-id",
  "expires_at": "2024-01-01T00:00:00Z",
  "scope": "openid profile email"
}
```

### Todo Endpoints

All todo endpoints require `Authorization: Bearer <token>` header.

#### `GET /api/todos`
Get all todos for authenticated user.

**Response:**
```json
[
  {
    "id": "uuid",
    "userId": "user-id",
    "title": "Task title",
    "description": "Task description",
    "status": "todo",
    "priority": "medium",
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  }
]
```

#### `GET /api/todos/:id`
Get specific todo by ID.

**Response:**
```json
{
  "id": "uuid",
  "userId": "user-id",
  "title": "Task title",
  ...
}
```

#### `POST /api/todos`
Create a new todo.

**Request Body:**
```json
{
  "title": "Task title",
  "description": "Task description (optional)",
  "priority": "low|medium|high (optional, default: medium)"
}
```

**Response:** `201 Created` with created todo object

#### `PUT /api/todos/:id`
Update a todo.

**Request Body:**
```json
{
  "title": "Updated title (optional)",
  "description": "Updated description (optional)",
  "status": "todo|in_progress|done (optional)",
  "priority": "low|medium|high (optional)"
}
```

**Response:** Updated todo object

#### `DELETE /api/todos/:id`
Delete a todo.

**Response:** `204 No Content`

#### `PATCH /api/todos/:id/status`
Update todo status (for drag & drop).

**Request Body:**
```json
{
  "status": "todo|in_progress|done"
}
```

**Response:** Updated todo object

### Health Check

#### `GET /health`
Check application health.

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## Project Structure

```
src/
├── main.ts                      # Application entry point
├── app.module.ts                # Root module
├── app.controller.ts            # Health check controller
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
│       ├── userinfo.dto.ts
│       └── validation-result.dto.ts
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
│   └── services/
│       ├── oidc.service.ts      # OIDC utilities
│       ├── token.service.ts     # Token validation
│       └── crypto.service.ts    # Crypto utilities
└── common/
    ├── decorators/
    │   └── user.decorator.ts    # Extract user from token
    └── filters/
        └── http-exception.filter.ts
```

## Security Features

### Token Management
- **Refresh Tokens**: Stored in HttpOnly cookies with secure, sameSite flags
- **Access Tokens**: Never stored server-side, passed from frontend
- **Token Rotation**: New refresh token issued on every refresh

### Cookie Security
```typescript
{
  httpOnly: true,           // Prevent JavaScript access
  secure: true,             // HTTPS only in production
  sameSite: 'lax',         // CSRF protection
  maxAge: 7 * 24 * 60 * 60 * 1000,   // 7 days
  path: '/'
}
```

### CORS Configuration
- Allows only configured frontend URL
- Enables credentials for cookie support
- Restricts allowed methods and headers

### Input Validation
- DTOs with class-validator decorators
- Automatic validation pipe
- Whitelist unknown properties

### Authorization
- Guards protect all sensitive endpoints
- User ownership verification for todos
- Token-based user identification

## Error Handling

All errors follow a consistent format:

```json
{
  "error": "error_code",
  "message": "Human-readable error message",
  "timestamp": "2024-01-01T00:00:00Z",
  "path": "/api/endpoint"
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
| `server_error` | 500 | Internal server error |

## Testing

```bash
# Run unit tests
npm run test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:cov

# Run e2e tests
npm run test:e2e
```

## Deployment

### Production Checklist

- [ ] Set `NODE_ENV=production`
- [ ] Use HTTPS for all connections
- [ ] Configure production MongoDB instance
- [ ] Set strong `CLIENT_SECRET`
- [ ] Configure proper CORS origins
- [ ] Enable request logging
- [ ] Set up health check monitoring
- [ ] Configure graceful shutdown
- [ ] Use environment-specific secrets
- [ ] Enable rate limiting (optional)

### Docker Deployment (Optional)

```dockerfile
FROM node:18-alpine

WORKDIR /app

COPY package*.json ./
RUN npm ci --only=production

COPY . .
RUN npm run build

EXPOSE 3001

CMD ["npm", "run", "start:prod"]
```

## Troubleshooting

### MongoDB Connection Issues
```bash
# Check if MongoDB is running
mongosh --eval "db.adminCommand('ping')"

# Check connection string in .env
echo $MONGODB_URI
```

### OAuth2 Flow Issues
```bash
# Verify OAuth2 server is running
curl http://localhost:8080/health

# Check client credentials
curl http://localhost:8080/.well-known/openid-configuration
```

### CORS Issues
- Verify `FRONTEND_URL` matches your React app URL
- Check browser console for CORS errors
- Ensure credentials are enabled in frontend requests

## Development Tips

### Hot Reload
The application uses NestJS watch mode for automatic reloading during development.

### Debugging
```bash
# Start with debug mode
npm run start:debug

# Attach debugger in VS Code (launch.json)
{
  "type": "node",
  "request": "attach",
  "name": "Attach NestJS",
  "port": 9229
}
```

### Database Inspection
```bash
# Connect to MongoDB
mongosh mongodb://localhost:27017/oauth2_bff_app

# View todos
db.todos.find().pretty()

# View indexes
db.todos.getIndexes()
```

## Contributing

1. Follow NestJS best practices
2. Use TypeScript strict mode
3. Add tests for new features
4. Update documentation
5. Follow existing code style

## License

MIT

## Support

For issues and questions:
- Check the [MIGRATION.md](./MIGRATION.md) for differences from Express version
- Review the main [oauth2-bff-app README](../README.md)
- Check OAuth2 server documentation

## Related Documentation

- [NestJS Documentation](https://docs.nestjs.com/)
- [OAuth2 RFC 6749](https://tools.ietf.org/html/rfc6749)
- [OpenID Connect Core](https://openid.net/specs/openid-connect-core-1_0.html)
- [MongoDB Node.js Driver](https://www.mongodb.com/docs/drivers/node/)
