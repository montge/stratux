// @ts-check
const { defineConfig, devices } = require('@playwright/test');

/**
 * Playwright configuration for Stratux web interface testing.
 *
 * Usage:
 *   npm run test:e2e:device  - Test against real device at http://10.0.1.53
 *   npm run test:e2e:mock    - Test against mock server (for CI)
 *   npm run test:e2e         - Test against STRATUX_URL env var or localhost:9090
 */

const baseURL = process.env.STRATUX_URL || 'http://localhost:9090';

module.exports = defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',

  use: {
    baseURL,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },

  /* Use only Chromium by default. Run `npx playwright install` to add more browsers. */
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],

  /* No webServer config - we test against external device or mock server */
});
