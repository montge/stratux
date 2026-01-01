// @ts-check
const { test, expect } = require('@playwright/test');

/**
 * API endpoint tests for Stratux web interface.
 * These tests verify the backend API is accessible and returns valid data.
 */

test.describe('REST API Endpoints', () => {
  test('getStatus returns valid JSON', async ({ request }) => {
    const response = await request.get('/getStatus');
    expect(response.ok()).toBeTruthy();

    const data = await response.json();
    expect(data).toHaveProperty('Version');
    expect(data).toHaveProperty('Build');
    expect(data).toHaveProperty('Devices');
  });

  test('getSettings returns valid JSON', async ({ request }) => {
    const response = await request.get('/getSettings');
    expect(response.ok()).toBeTruthy();

    const data = await response.json();
    // Settings should have various configuration options
    expect(typeof data).toBe('object');
    expect(data).toHaveProperty('UAT_Enabled');
  });

  test('setSettings can update and persist settings', async ({ request }) => {
    // Get current settings
    const getResponse = await request.get('/getSettings');
    expect(getResponse.ok()).toBeTruthy();
    const originalSettings = await getResponse.json();

    // Toggle a safe setting (DarkMode) to test save/retrieve
    const testValue = !originalSettings.DarkMode;
    const updateResponse = await request.post('/setSettings', {
      data: { DarkMode: testValue }
    });
    expect(updateResponse.ok()).toBeTruthy();

    // Verify the setting was saved
    const verifyResponse = await request.get('/getSettings');
    expect(verifyResponse.ok()).toBeTruthy();
    const newSettings = await verifyResponse.json();
    expect(newSettings.DarkMode).toBe(testValue);

    // Restore original setting
    await request.post('/setSettings', {
      data: { DarkMode: originalSettings.DarkMode }
    });
  });

  test('getSituation returns valid JSON', async ({ request }) => {
    const response = await request.get('/getSituation');
    expect(response.ok()).toBeTruthy();

    const data = await response.json();
    expect(typeof data).toBe('object');
  });

  test('getTowers returns valid JSON', async ({ request }) => {
    const response = await request.get('/getTowers');
    expect(response.ok()).toBeTruthy();

    const data = await response.json();
    expect(typeof data).toBe('object');
  });
});

test.describe('WebSocket Connections', () => {
  test('status WebSocket connects', async ({ page }) => {
    await page.goto('/');

    // Wait for the page to establish WebSocket connection
    await expect(page.locator('body[ng-app="stratux"]')).toBeVisible();

    // Check that status data is being received (Version should update)
    await expect(page.locator('text=Version')).toBeVisible({ timeout: 10000 });
  });
});
