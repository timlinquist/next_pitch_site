import { test as setup, expect } from '@playwright/test';
import { getUser, promoteToAdmin, closeDb } from './helpers/db';
import { authStoragePath } from './helpers/auth';
import fs from 'fs';
import path from 'path';

const email = process.env.E2E_AUTH0_TEST_EMAIL;
const password = process.env.E2E_AUTH0_TEST_PASSWORD;

setup.skip(!email || !password, 'E2E_AUTH0_TEST_EMAIL / E2E_AUTH0_TEST_PASSWORD not set');

setup('authenticate via Auth0 and cache session', async ({ page }) => {
  // Ensure .auth directory exists
  const authDir = path.dirname(authStoragePath);
  if (!fs.existsSync(authDir)) {
    fs.mkdirSync(authDir, { recursive: true });
  }

  // Navigate to the app
  await page.goto('http://localhost:5173');

  // Click "My Account" to trigger Auth0 login redirect
  await page.getByRole('button', { name: 'My Account' }).first().click({ force: true });

  // Wait for Auth0 Universal Login page
  await page.waitForURL(/auth0\.com/, { timeout: 15_000 });

  // Prefer login (user likely exists from previous runs). If on signup page, switch to login.
  const loginLink = page.getByText('Log in', { exact: true });
  if (await loginLink.isVisible({ timeout: 3_000 }).catch(() => false)) {
    await loginLink.click();
    await page.waitForTimeout(500);
  }

  // Fill credentials and submit
  await page.locator('input[name="username"], input[name="email"]').first().fill(email!);
  await page.locator('input[name="password"]').fill(password!);
  await page.locator('button[data-action-button-primary="true"]').click();

  // If login fails (user doesn't exist yet), switch to signup and retry
  const errorBanner = page.locator('[id*="error"], [class*="error"], [role="alert"]');
  if (await errorBanner.isVisible({ timeout: 5_000 }).catch(() => false)) {
    const signUpLink = page.getByText('Sign up', { exact: true });
    if (await signUpLink.isVisible({ timeout: 2_000 }).catch(() => false)) {
      await signUpLink.click();
      await page.waitForTimeout(500);
      await page.locator('input[name="username"], input[name="email"]').first().fill(email!);
      await page.locator('input[name="password"]').fill(password!);
      await page.locator('button[data-action-button-primary="true"]').click();
    }
  }

  // Auth0 may show an "Accept" / authorize prompt — handle it
  const acceptButton = page.getByRole('button', { name: /accept|authorize|continue/i });
  if (await acceptButton.isVisible({ timeout: 5_000 }).catch(() => false)) {
    await acceptButton.click();
  }

  // Wait for redirect back to app
  await page.waitForURL(/localhost:5173/, { timeout: 30_000 });

  // Give the Auth0 SDK time to process tokens and store in localStorage
  await page.waitForTimeout(2_000);

  // Navigate to admin page — this triggers GET /api/users/me which auto-creates the user record
  await page.goto('http://localhost:5173/admin/camps');
  await page.waitForTimeout(3_000);

  // Verify user record was created in DB
  const user = await getUser(email!);
  if (!user) {
    setup.info().annotations.push({
      type: 'issue',
      description: 'User record was NOT created in DB after Auth0 signup — GetCurrentUser bug confirmed',
    });
  }
  expect(user, 'User record should exist in DB after Auth0 signup/login').toBeTruthy();

  // Promote to admin for subsequent tests
  await promoteToAdmin(email!);

  // Save authenticated state
  await page.context().storageState({ path: authStoragePath });

  await closeDb();
});
