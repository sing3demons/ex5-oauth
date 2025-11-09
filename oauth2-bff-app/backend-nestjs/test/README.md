# Backend NestJS E2E Tests

This directory contains end-to-end tests for the NestJS backend API.

## Overview

The E2E test suite covers:
- **Health Check**: Application health endpoint
- **Authentication**: OAuth2/OIDC authentication flow
- **Todos API**: CRUD operations for todos
- **Authorization**: User-specific data access
- **Validation**: Input validation and error handling

## Test Files

### `app.e2e-spec.ts`
Tests the application health endpoint and basic routing:
- Health check endpoint
- 404 handling
- Timestamp format validation

### `auth.e2e-spec.ts`
Tests OAuth2/OIDC authentication endpoints:
- Login initiation (PKCE flow)
- OAuth2 callback handling
- Token refresh
- Logout
- User info retrieval
- OIDC discovery
- JWKS endpoint
- Token validation and decoding

### `todos.e2e-spec.ts`
Tests todos API without authentication:
- Authentication requirements
- Endpoint structure
- HTTP method validation
- Content-Type handling
- Request validation

### `todos-authenticated.e2e-spec.ts`
Tests todos API with authentication (requires setup):
- CRUD operations
- Input validation
- Authorization checks
- Status transitions
- Sorting and filtering

## Running Tests

### Run all E2E tests
```bash
npm run test:e2e
```

### Run specific test file
```bash
npm run test:e2e -- app.e2e-spec
npm run test:e2e -- auth.e2e-spec
npm run test:e2e -- todos.e2e-spec
```

### Run with coverage
```bash
npm run test:e2e -- --coverage
```

### Run in watch mode
```bash
npm run test:e2e -- --watch
```

### Run with verbose output
```bash
npm run test:e2e -- --verbose
```

## Prerequisites

Before running E2E tests, ensure:

1. **MongoDB is running** on `mongodb://localhost:27017`
2. **OAuth2 server is running** on `http://localhost:8080`
3. **Environment variables are set** (see `.env.example`)

## Configuration

### Environment Variables

Create a `.env.test` file for test-specific configuration:

```bash
# OAuth2 Configuration
OAUTH2_SERVER_URL=http://localhost:8080
OAUTH2_CLIENT_ID=your-test-client-id
OAUTH2_CLIENT_SECRET=your-test-client-secret
OAUTH2_REDIRECT_URI=http://localhost:4000/auth/callback

# Database
MONGODB_URI=mongodb://localhost:27017/todo_app_test

# Server
PORT=4001
NODE_ENV=test

# Test User Credentials (for authenticated tests)
TEST_USERNAME=test@example.com
TEST_PASSWORD=testpassword123
```

### Jest Configuration

E2E tests use a separate Jest configuration in `jest-e2e.json`:

```json
{
  "moduleFileExtensions": ["js", "json", "ts"],
  "rootDir": ".",
  "testEnvironment": "node",
  "testRegex": ".e2e-spec.ts$",
  "transform": {
    "^.+\\.(t|j)s$": "ts-jest"
  }
}
```

## Test Structure

### Basic Test Pattern

```typescript
describe('Feature (e2e)', () => {
  let app: INestApplication;

  beforeAll(async () => {
    const moduleFixture = await Test.createTestingModule({
      imports: [AppModule],
    }).compile();

    app = moduleFixture.createNestApplication();
    await app.init();
  });

  afterAll(async () => {
    await app.close();
  });

  it('should test something', () => {
    return request(app.getHttpServer())
      .get('/endpoint')
      .expect(200)
      .expect((res) => {
        expect(res.body).toHaveProperty('property');
      });
  });
});
```

### Authenticated Test Pattern

```typescript
it('should perform authenticated action', async () => {
  const authHeader = await getAuthHeader();
  
  return request(app.getHttpServer())
    .post('/api/todos')
    .set(authHeader)
    .send({ title: 'Test' })
    .expect(201);
});
```

## Authentication in Tests

### Current Status

Most authenticated tests are marked with `.skip` because they require valid OAuth2 tokens. The tests that run verify:
- Endpoint structure
- Authentication requirements
- Error handling
- Request validation

### Enabling Authenticated Tests

To enable full E2E testing with authentication:

#### Option 1: Use Test OAuth2 Server

1. **Create test user** in OAuth2 server:
```bash
# Register test user
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "test@example.com",
    "password": "testpassword123"
  }'
```

2. **Implement `getTestAccessToken()`** in `test/helpers/auth.helper.ts`:
```typescript
export async function getTestAccessToken(): Promise<string> {
  const response = await axios.post('http://localhost:8080/oauth2/token', {
    grant_type: 'password',
    username: process.env.TEST_USERNAME,
    password: process.env.TEST_PASSWORD,
    client_id: process.env.OAUTH2_CLIENT_ID,
    client_secret: process.env.OAUTH2_CLIENT_SECRET,
    scope: 'openid profile email'
  });
  return response.data.access_token;
}
```

3. **Remove `.skip`** from test suites in `todos-authenticated.e2e-spec.ts`

#### Option 2: Mock OAuth2 Responses

1. **Install nock** for HTTP mocking:
```bash
npm install -D nock
```

2. **Mock OAuth2 endpoints**:
```typescript
import * as nock from 'nock';

beforeAll(() => {
  nock('http://localhost:8080')
    .post('/oauth2/token')
    .reply(200, {
      access_token: 'mock-access-token',
      token_type: 'Bearer',
      expires_in: 3600
    });

  nock('http://localhost:8080')
    .get('/oauth2/userinfo')
    .reply(200, {
      sub: 'test-user-id',
      email: 'test@example.com',
      name: 'Test User'
    });
});
```

#### Option 3: Use Test Database with Seeded Data

1. **Create test database** with pre-seeded users and tokens
2. **Bypass OAuth2** by directly setting session data
3. **Use internal service methods** to create test sessions

## Test Data Management

### Setup Test Data

```typescript
beforeEach(async () => {
  // Create test data
  const authHeader = await getAuthHeader();
  const response = await request(app.getHttpServer())
    .post('/api/todos')
    .set(authHeader)
    .send({ title: 'Test Todo' });
  
  testTodoId = response.body.id;
});
```

### Cleanup Test Data

```typescript
afterEach(async () => {
  // Clean up test data
  const authHeader = await getAuthHeader();
  await request(app.getHttpServer())
    .delete(`/api/todos/${testTodoId}`)
    .set(authHeader);
});
```

### Use Separate Test Database

```typescript
beforeAll(async () => {
  // Connect to test database
  process.env.MONGODB_URI = 'mongodb://localhost:27017/todo_app_test';
  
  const moduleFixture = await Test.createTestingModule({
    imports: [AppModule],
  }).compile();

  app = moduleFixture.createNestApplication();
  await app.init();
});

afterAll(async () => {
  // Drop test database
  const connection = app.get(DatabaseService);
  await connection.dropDatabase();
  await app.close();
});
```

## Best Practices

### 1. Test Independence
Each test should be independent and not rely on other tests:
```typescript
// ❌ Bad: Depends on previous test
it('should create todo', () => { /* creates todo */ });
it('should update todo', () => { /* uses todo from previous test */ });

// ✅ Good: Independent tests
it('should update todo', async () => {
  // Create todo in this test
  const todo = await createTestTodo();
  // Update it
  await updateTodo(todo.id);
  // Clean up
  await deleteTodo(todo.id);
});
```

### 2. Use Descriptive Test Names
```typescript
// ❌ Bad
it('should work', () => { /* ... */ });

// ✅ Good
it('should return 401 when accessing todos without authentication', () => { /* ... */ });
```

### 3. Test Error Cases
```typescript
it('should return 400 when creating todo without title', async () => {
  const authHeader = await getAuthHeader();
  
  await request(app.getHttpServer())
    .post('/api/todos')
    .set(authHeader)
    .send({ description: 'No title' })
    .expect(400)
    .expect((res) => {
      expect(res.body.message).toContain('title');
    });
});
```

### 4. Clean Up Resources
```typescript
afterEach(async () => {
  // Always clean up test data
  await cleanupTestData(getTestUserId());
});
```

### 5. Use Test Helpers
```typescript
// Create reusable helpers
async function createTestTodo(title: string) {
  const authHeader = await getAuthHeader();
  const response = await request(app.getHttpServer())
    .post('/api/todos')
    .set(authHeader)
    .send({ title, description: 'Test' });
  return response.body;
}
```

## Troubleshooting

### Tests Fail with "Cannot connect to MongoDB"
- Ensure MongoDB is running: `mongod --dbpath /path/to/data`
- Check connection string in `.env.test`
- Verify network connectivity

### Tests Fail with "OAuth2 server not responding"
- Ensure OAuth2 server is running on port 8080
- Check `OAUTH2_SERVER_URL` in environment variables
- Verify OAuth2 server is accessible

### Tests Timeout
- Increase Jest timeout in `jest-e2e.json`:
```json
{
  "testTimeout": 30000
}
```
- Or in individual tests:
```typescript
it('should do something', async () => {
  // ...
}, 30000); // 30 second timeout
```

### Authentication Tests Are Skipped
- Implement `getTestAccessToken()` in `test/helpers/auth.helper.ts`
- Set up test OAuth2 credentials
- Remove `.skip` from test suites

### Port Already in Use
- Change test port in `.env.test`
- Kill process using the port: `lsof -ti:4000 | xargs kill`

## CI/CD Integration

### GitHub Actions Example

```yaml
name: E2E Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      mongodb:
        image: mongo:6
        ports:
          - 27017:27017
    
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      
      - name: Install dependencies
        run: npm ci
        working-directory: oauth2-bff-app/backend-nestjs
      
      - name: Run E2E tests
        run: npm run test:e2e
        working-directory: oauth2-bff-app/backend-nestjs
        env:
          MONGODB_URI: mongodb://localhost:27017/todo_app_test
          OAUTH2_SERVER_URL: http://localhost:8080
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./oauth2-bff-app/backend-nestjs/coverage-e2e/lcov.info
```

## Test Coverage

### View Coverage Report

```bash
npm run test:e2e -- --coverage
```

Coverage report will be generated in `coverage-e2e/` directory.

### Open HTML Report

```bash
open coverage-e2e/lcov-report/index.html
```

## Resources

- [NestJS Testing Documentation](https://docs.nestjs.com/fundamentals/testing)
- [Jest Documentation](https://jestjs.io/docs/getting-started)
- [Supertest Documentation](https://github.com/visionmedia/supertest)
- [Testing Best Practices](https://github.com/goldbergyoni/javascript-testing-best-practices)

## Contributing

When adding new E2E tests:

1. Follow existing test patterns
2. Add descriptive test names
3. Include both success and error cases
4. Clean up test data
5. Update this README if needed
6. Ensure tests pass before committing

## Support

For issues or questions:
1. Check this README
2. Review existing tests for examples
3. Check NestJS documentation
4. Open an issue in the repository
