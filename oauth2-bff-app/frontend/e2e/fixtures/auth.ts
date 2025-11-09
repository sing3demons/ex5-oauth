import { Page } from '@playwright/test';

/**
 * Mock authentication helper for E2E tests
 * In a real scenario, you would either:
 * 1. Use a test OAuth2 server
 * 2. Mock the OAuth2 responses
 * 3. Use API calls to create test sessions
 */
export async function mockAuthentication(page: Page) {
  // Set up mock session storage/cookies
  await page.addInitScript(() => {
    // Mock authenticated state
    window.localStorage.setItem('auth_test_mode', 'true');
  });
}

/**
 * Helper to wait for navigation and ensure page is loaded
 */
export async function waitForPageLoad(page: Page) {
  await page.waitForLoadState('networkidle');
  await page.waitForLoadState('domcontentloaded');
}

/**
 * Helper to check if user is on login page
 */
export async function isOnLoginPage(page: Page): Promise<boolean> {
  const url = page.url();
  return url.includes('/login') || url === 'http://localhost:3000/';
}

/**
 * Helper to check if user is on dashboard
 */
export async function isOnDashboard(page: Page): Promise<boolean> {
  const url = page.url();
  return url.includes('/dashboard');
}
