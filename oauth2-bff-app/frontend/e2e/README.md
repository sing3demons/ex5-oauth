# End-to-End Tests

This directory contains Playwright E2E tests for the Todo App with SSO.

## Overview

The E2E tests cover:
- **OAuth2 Authentication Flow**: Login, logout, error handling
- **Todo CRUD Operations**: Create, read, update, delete todos
- **Drag-and-Drop**: Moving todos between columns
- **Complete User Flows**: Full user journeys through the app
- **Responsive Design**: Mobile, tablet, and desktop layouts
- **Error Recovery**: Handling various error scenarios
- **Performance**: Load times and responsiveness
- **Security**: Token handling and CSRF protection

## Prerequisites

Before running E2E tests, ensure:

1. **OAuth2 Server is running** on `http://localhost:8080`
2. **Backend server is running** on `http://localhost:4000`
3. **MongoDB is running** on `mongodb://localhost:27017`
4. **Frontend dev server** will be started automatically by Playwright

## Installation

Install Playwright and browsers:

```bash
npm install -D @playwright/test
npx playwright install chromium
```

## Running Tests

### Run all E2E tests

```bash
npm run test:e2e
```

### Run tests in UI mode (interactive)

```bash
npm run test:e2e:ui
```

### Run tests in headed mode (see browser)

```bash
npx playwright test --headed
```

### Run specific test file

```bash
npx playwright test e2e/auth.spec.ts
```

### Run tests in debug mode

```bash
npx playwright test --debug
```

## Test Structure

### `auth.spec.ts`
Tests OAuth2 authentication flow:
- Login page display
- OAuth2 redirect
- Error handling
- Protected routes

### `todos.spec.ts`
Tests todo CRUD operations:
- Creating todos
- Editing todos
- Deleting todos
- Form validation
- Error handling

### `drag-and-drop.spec.ts`
Tests drag-and-drop functionality:
- Dragging between columns
- Visual feedback
- Touch gestures
- Performance

### `user-flow.spec.ts`
Tests complete user journeys:
- Full login to logout flow
- Todo management workflow
- Responsive design
- Error recovery
- Performance metrics

## Authentication in Tests

### Current Approach

Most tests are marked as `test.skip()` because they require authentication. The tests that run verify:
- Login page UI
- Error handling
- Responsive design
- Basic navigation

### Full E2E Testing

To run full E2E tests with authentication, you need to:

1. **Option A: Use Test Credentials**
   - Create a test user in the OAuth2 server
   - Store credentials in environment variables
   - Implement authentication helper in `fixtures/auth.ts`

2. **Option B: Mock OAuth2 Responses**
   - Use Playwright's route mocking
   - Mock OAuth2 endpoints
   - Set up test session cookies

3. **Option C: Use API to Create Sessions**
   - Call backend API to create test sessions
   - Set session cookies in tests
   - Bypass OAuth2 flow for testing

### Example: Implementing Authentication

```typescript
// In fixtures/auth.ts
export async function authenticateUser(page: Page) {
  // Option 1: Use test credentials
  await page.goto('/');
  await page.click('button:has-text("Login with OAuth2")');
  await page.fill('input[name="username"]', process.env.TEST_USERNAME!);
  await page.fill('input[name="password"]', process.env.TEST_PASSWORD!);
  await page.click('button[type="submit"]');
  await page.waitForURL('**/dashboard');
}

// In test file
test('should create todo', async ({ page }) => {
  await authenticateUser(page);
  // Now you can test authenticated features
});
```

## Configuration

### `playwright.config.ts`

Key configuration options:
- `baseURL`: Frontend URL (http://localhost:3000)
- `webServer`: Automatically starts dev server
- `use.trace`: Records traces on first retry
- `use.screenshot`: Takes screenshots on failure

### Environment Variables

Create `.env.test` for test-specific configuration:

```bash
VITE_BFF_URL=http://localhost:4000
TEST_USERNAME=testuser@example.com
TEST_PASSWORD=testpassword123
```

## Viewing Test Results

### HTML Report

After running tests, view the HTML report:

```bash
npx playwright show-report
```

### Traces

If a test fails, view the trace:

```bash
npx playwright show-trace trace.zip
```

## Best Practices

1. **Keep tests independent**: Each test should be able to run in isolation
2. **Use data-testid**: Add `data-testid` attributes for reliable selectors
3. **Wait for elements**: Use `waitFor` methods instead of fixed timeouts
4. **Clean up**: Reset state between tests
5. **Mock external services**: Don't depend on external APIs in tests
6. **Test user flows**: Focus on real user scenarios, not implementation details

## Troubleshooting

### Tests fail with "Target closed"
- Ensure all servers are running
- Check for JavaScript errors in the app
- Increase timeout in playwright.config.ts

### Authentication tests are skipped
- Implement authentication helper
- Set up test credentials
- Or use API to create test sessions

### Drag-and-drop tests fail
- Ensure @dnd-kit is properly configured
- Check for console errors
- Try running in headed mode to see what's happening

### Slow test execution
- Run tests in parallel: `npx playwright test --workers=4`
- Use `test.describe.configure({ mode: 'parallel' })`
- Optimize wait times and selectors

## CI/CD Integration

### GitHub Actions Example

```yaml
name: E2E Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      
      - name: Install dependencies
        run: npm ci
      
      - name: Install Playwright
        run: npx playwright install --with-deps chromium
      
      - name: Start services
        run: |
          docker-compose up -d
          npm run dev &
      
      - name: Run E2E tests
        run: npm run test:e2e
      
      - name: Upload test results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: playwright-report
          path: playwright-report/
```

## Future Improvements

1. **Visual Regression Testing**: Add screenshot comparison tests
2. **Accessibility Testing**: Use @axe-core/playwright for a11y tests
3. **Performance Testing**: Add Lighthouse CI integration
4. **API Mocking**: Use MSW for consistent API responses
5. **Test Data Management**: Create fixtures for test data
6. **Cross-browser Testing**: Add Firefox and WebKit
7. **Mobile Testing**: Add mobile device emulation tests

## Resources

- [Playwright Documentation](https://playwright.dev)
- [Best Practices](https://playwright.dev/docs/best-practices)
- [Debugging Tests](https://playwright.dev/docs/debug)
- [CI/CD Integration](https://playwright.dev/docs/ci)
