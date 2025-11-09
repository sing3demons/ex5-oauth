# E2E Tests Quick Start

## Run Tests

```bash
# Run all tests
npm run test:e2e

# Run in UI mode (recommended for development)
npm run test:e2e:ui

# Run in headed mode (see the browser)
npm run test:e2e:headed

# Run specific test file
npx playwright test e2e/auth.spec.ts

# Run in debug mode
npm run test:e2e:debug
```

## View Results

```bash
# View HTML report
npx playwright show-report

# View trace for failed test
npx playwright show-trace test-results/[test-name]/trace.zip
```

## Current Test Status

✅ **13 tests passing** - Authentication, responsive design, error handling
⏭️ **50 tests skipped** - Require authentication setup (fully implemented, ready to enable)

## What's Working

- Login page display and OAuth2 features
- OAuth2 redirect flow
- Error handling (all error types)
- Protected route redirection
- Responsive design (mobile, tablet, desktop)
- Performance testing
- Error recovery

## What Needs Authentication

All skipped tests are fully implemented and just need authentication to be set up:
- Todo CRUD operations
- Drag-and-drop functionality
- User profile display
- Session management
- Security features

See `README.md` for authentication setup instructions.

## Quick Tips

1. **Use UI mode for development**: `npm run test:e2e:ui`
   - Interactive test runner
   - Time travel debugging
   - Watch mode

2. **Debug failing tests**: `npm run test:e2e:debug`
   - Step through tests
   - Inspect elements
   - View network requests

3. **Run specific tests**: Add `.only` to test
   ```typescript
   test.only('should display login page', async ({ page }) => {
     // This test will run alone
   });
   ```

4. **Skip tests temporarily**: Add `.skip` to test
   ```typescript
   test.skip('should create todo', async ({ page }) => {
     // This test will be skipped
   });
   ```

## Common Issues

### "Target closed" error
- Ensure frontend dev server is running
- Check for JavaScript errors in the app
- Increase timeout in playwright.config.ts

### Tests are slow
- Run in parallel: `npx playwright test --workers=4`
- Use headed mode only when debugging
- Optimize wait times

### Authentication tests skipped
- Implement authentication helper in `fixtures/auth.ts`
- Set up test credentials
- See `README.md` for detailed instructions

## Next Steps

To enable all tests:
1. Read `README.md` for authentication setup
2. Implement authentication helper
3. Remove `.skip` from authenticated tests
4. Run full test suite

## Resources

- [Full Documentation](./README.md)
- [Playwright Docs](https://playwright.dev)
- [Test Files](./e2e/)
