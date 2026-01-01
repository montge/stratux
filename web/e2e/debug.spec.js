// @ts-check
const { test, expect } = require('@playwright/test');

/**
 * Debug test to capture console errors from the web interface.
 */

test.describe('Debug Console Errors', () => {
  test('capture all console messages', async ({ page }) => {
    const consoleMessages = [];
    const consoleErrors = [];

    page.on('console', msg => {
      consoleMessages.push({ type: msg.type(), text: msg.text() });
      if (msg.type() === 'error') {
        consoleErrors.push(msg.text());
      }
    });

    page.on('pageerror', error => {
      consoleErrors.push(`Page Error: ${error.message}`);
    });

    await page.goto('/');

    // Wait a bit for any async errors
    await page.waitForTimeout(5000);

    console.log('=== All Console Messages ===');
    consoleMessages.forEach(msg => {
      console.log(`[${msg.type}] ${msg.text}`);
    });

    console.log('\n=== Console Errors ===');
    consoleErrors.forEach(err => {
      console.log(`ERROR: ${err}`);
    });

    // If there are errors, log them clearly
    if (consoleErrors.length > 0) {
      console.log('\n=== JAVASCRIPT ERRORS FOUND ===');
      consoleErrors.forEach((err, i) => {
        console.log(`${i + 1}. ${err}`);
      });
    }

    // This test always passes - it's just for debugging
    expect(true).toBe(true);
  });

  test('check page HTML structure', async ({ page }) => {
    await page.goto('/');
    await page.waitForTimeout(3000);

    const html = await page.content();

    // Check if Angular app bootstrapped
    const hasNgApp = html.includes('ng-app="stratux"');
    console.log(`Angular app attribute present: ${hasNgApp}`);

    // Check ui-view content
    const uiViewContent = await page.locator('[ui-view]').innerHTML();
    console.log(`ui-view content length: ${uiViewContent.length}`);
    console.log(`ui-view content: ${uiViewContent.substring(0, 500)}`);

    // Check if Angular rendered any ng- attributes
    const ngBindings = await page.locator('[ng-bind], [ng-show], [ng-hide], [ng-if]').count();
    console.log(`Angular binding elements: ${ngBindings}`);

    expect(true).toBe(true);
  });
});
