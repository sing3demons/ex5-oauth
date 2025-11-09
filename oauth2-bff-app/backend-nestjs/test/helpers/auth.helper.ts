/**
 * Authentication Helper for E2E Tests
 * 
 * This helper provides utilities for testing authenticated endpoints.
 * 
 * To enable full E2E testing with authentication:
 * 1. Set up a test OAuth2 server or use mocks
 * 2. Implement getTestAccessToken() to return a valid token
 * 3. Implement getTestRefreshToken() to return a valid refresh token
 */

/**
 * Get a test access token
 * 
 * Implementation options:
 * 1. Use a test OAuth2 server with test credentials
 * 2. Mock the OAuth2 token endpoint
 * 3. Generate a test JWT token
 */
export async function getTestAccessToken(): Promise<string> {
  // TODO: Implement based on your test setup
  // Example with test OAuth2 server:
  // const response = await axios.post('http://localhost:8080/oauth2/token', {
  //   grant_type: 'password',
  //   username: 'test@example.com',
  //   password: 'testpassword',
  //   client_id: process.env.TEST_CLIENT_ID,
  //   client_secret: process.env.TEST_CLIENT_SECRET,
  // });
  // return response.data.access_token;
  
  throw new Error('getTestAccessToken not implemented. See test/helpers/auth.helper.ts');
}

/**
 * Get a test refresh token
 */
export async function getTestRefreshToken(): Promise<string> {
  // TODO: Implement based on your test setup
  throw new Error('getTestRefreshToken not implemented. See test/helpers/auth.helper.ts');
}

/**
 * Get test user ID
 */
export function getTestUserId(): string {
  return 'test-user-id-12345';
}

/**
 * Create authorization header with test token
 */
export async function getAuthHeader(): Promise<{ Authorization: string }> {
  const token = await getTestAccessToken();
  return { Authorization: `Bearer ${token}` };
}

/**
 * Create mock authorization header (for testing without real OAuth2)
 */
export function getMockAuthHeader(): { Authorization: string } {
  // This will fail authentication but is useful for testing endpoint structure
  return { Authorization: 'Bearer mock-test-token-12345' };
}

/**
 * Mock user info response
 */
export function getMockUserInfo() {
  return {
    sub: getTestUserId(),
    email: 'test@example.com',
    name: 'Test User',
    email_verified: true,
  };
}

/**
 * Wait for async operations to complete
 */
export function wait(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * Clean up test data
 */
export async function cleanupTestData(userId: string): Promise<void> {
  // TODO: Implement cleanup logic
  // Example: Delete all todos for test user
  // await todosService.deleteAllByUser(userId);
}
