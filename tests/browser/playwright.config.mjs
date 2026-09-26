import { defineConfig } from '@playwright/test';
import { fixture } from './evidence.mjs';
const run = fixture();
export default defineConfig({
  testDir: '.', testMatch: '*.spec.mjs', fullyParallel: false, workers: 1,
  timeout: 90_000, expect: { timeout: 15_000 }, retries: 0,
  reporter: [['list'], ['html', { open: 'never' }]],
  use: { baseURL: run.baseURL, browserName: 'chromium', locale: 'en-US', timezoneId: 'UTC', reducedMotion: 'reduce',
    launchOptions: process.env.BROWSER_EXECUTABLE ? { executablePath: process.env.BROWSER_EXECUTABLE } : {},
    screenshot: 'only-on-failure', trace: 'off', video: 'off' },
  projects: [390, 1440].flatMap(width => ['light', 'dark'].map(colorScheme => ({
    name: `${width}-${colorScheme}-preference`, use: { viewport: { width, height: width === 390 ? 844 : 1000 }, colorScheme },
  }))),
});
