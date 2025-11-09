import { test, expect } from '@playwright/test';

test.describe('OAuth2 Authentication Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should display login page with OAuth2 features', async ({ page }) => {
    // Check if login page is displayed
    await expect(page.locator('h1')).toContainText('Todo App with SSO');
    
    // Check for OAuth2 features
    await expect(page.locator('text=HttpOnly Cookies')).toBeVisible();
    await expect(page.locator('text=PKCE Flow')).toBeVisible();
    await expect(page.locator('text=Auto Token Refresh')).toBeVisible();
    await expect(page.locator('text=Memory-only Access Tokens')).toBeVisible();
    
    // Check for login button
    const loginButton = page.locator('button:has-text("Login with OAuth2")');
    await expect(loginButton).toBeVisible();
  });

  test('should redirect to OAuth2 server when clicking login', async ({ page }) => {
    // Click login button
    const loginButton = page.locator('button:has-text("Login with OAuth2")');
    await loginButton.click();
    
    // Wait for navigation
    await page.waitForURL(/.*/, { timeout: 5000 });
    
    // Should redirect to backend auth endpoint or OAuth2 server
    const currentUrl = page.url();
    expect(currentUrl).not.toBe('http://localhost:3000/');
  });

  test('should display error message when OAuth2 error occurs', async ({ page }) => {
    // Navigate to login page with error parameter
    await page.goto('/login?error=access_denied');
    
    // Check for error message
    await expect(page.locator('text=Access was denied. Please try again.')).toBeVisible();
  });

  test('should handle invalid_request error', async ({ page }) => {
    await page.goto('/login?error=invalid_request');
    await expect(page.locator('text=Invalid request. Please try again.')).toBeVisible();
  });

  test('should handle server_error', async ({ page }) => {
    await page.goto('/login?error=server_error');
    await expect(page.locator('text=Server error occurred. Please try again later.')).toBeVisible();
  });
});

test.describe('Protected Routes', () => {
  test('should redirect to login when accessing dashboard without authentication', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Should redirect to login page
    await page.waitForURL('**/login', { timeout: 5000 });
    
    // Verify we're on the login page
    await expect(page.locator('h1')).toContainText('Todo App with SSO');
    await expect(page.locator('button:has-text("Login with OAuth2")')).toBeVisible();
  });
});
