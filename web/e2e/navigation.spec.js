// @ts-check
const { test, expect } = require('@playwright/test');

/**
 * Navigation tests for Stratux web interface.
 * These tests verify that all pages load correctly after library upgrades.
 */

test.describe('Page Navigation', () => {
  test('home page (status) loads', async ({ page }) => {
    await page.goto('/');

    // Wait for AngularJS to bootstrap
    await expect(page.locator('body[ng-app="stratux"]')).toBeVisible();

    // Status page should show version info
    await expect(page.locator('text=Version')).toBeVisible({ timeout: 10000 });
  });

  test('traffic page loads', async ({ page }) => {
    await page.goto('/#/traffic');

    await expect(page.locator('body[ng-app="stratux"]')).toBeVisible();
    // Traffic page has the traffic header panel
    await expect(page.locator('.panel_label:has-text("ADS-B")')).toBeVisible({ timeout: 10000 });
  });

  test('gps page loads', async ({ page }) => {
    await page.goto('/#/gps');

    await expect(page.locator('body[ng-app="stratux"]')).toBeVisible();
    // GPS page shows AHRS panel
    await expect(page.locator('.panel_label:has-text("AHRS")')).toBeVisible({ timeout: 10000 });
  });

  test('settings page loads', async ({ page }) => {
    await page.goto('/#/settings');

    await expect(page.locator('body[ng-app="stratux"]')).toBeVisible();
    // Settings page has WiFi Settings panel
    await expect(page.locator('.panel-heading:has-text("WiFi Settings")')).toBeVisible({ timeout: 10000 });
  });

  test('logs page loads', async ({ page }) => {
    await page.goto('/#/logs');

    await expect(page.locator('body[ng-app="stratux"]')).toBeVisible();
    // Logs page has a Logs heading
    await expect(page.locator('h2:has-text("Logs")')).toBeVisible({ timeout: 10000 });
  });

  test('radar page loads', async ({ page }) => {
    await page.goto('/#/radar');

    await expect(page.locator('body[ng-app="stratux"]')).toBeVisible();
    // Radar page - verify state changed correctly
    await page.waitForTimeout(2000);
    const state = await page.evaluate(() => {
      const injector = window.angular?.element(document.body).injector();
      if (!injector) return null;
      const $state = injector.get('$state');
      return $state.current.name;
    });
    expect(state).toBe('radar');
  });

  test('map page loads', async ({ page }) => {
    await page.goto('/#/map');

    await expect(page.locator('body[ng-app="stratux"]')).toBeVisible();
    // Map should have the OpenLayers container
    await expect(page.locator('#map_display')).toBeVisible({ timeout: 10000 });
  });
});

test.describe('Menu Navigation', () => {
  test('can navigate using menu', async ({ page }) => {
    await page.goto('/');

    // Wait for app to load
    await expect(page.locator('body[ng-app="stratux"]')).toBeVisible();
    await expect(page.locator('text=Version')).toBeVisible({ timeout: 10000 });

    // Open menu
    await page.click('text=Menu');

    // Click on Traffic
    await page.click('.sidebar >> text=Traffic');

    // Should navigate to traffic page
    await expect(page).toHaveURL(/\/#\/traffic/);
  });
});
