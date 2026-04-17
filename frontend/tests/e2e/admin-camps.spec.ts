import { test, expect } from '@playwright/test';
import { authStoragePath } from './helpers/auth';
import { seedTestCamp, seedTestRegistration, type TestCamp } from './helpers/db';

test.use({ storageState: authStoragePath });

test('admin creates a camp', async ({ page }) => {
  await page.goto('/admin/camps');

  // Should see admin page, not "Access Denied"
  await expect(page.getByRole('heading', { name: 'Manage Camps' })).toBeVisible({ timeout: 15_000 });

  // Open create form
  await page.getByRole('button', { name: 'Create New Camp' }).click();
  await expect(page.getByRole('heading', { name: 'New Camp' })).toBeVisible();

  // Fill camp form
  await page.locator('#camp-name').fill('E2E Test Admin Camp');
  await page.locator('#camp-desc').fill('Created by admin E2E test');
  await page.locator('#camp-start').fill('2026-08-01');
  await page.locator('#camp-end').fill('2026-08-03');
  await page.locator('#camp-price').fill('7500');
  await page.locator('#camp-cap').fill('15');

  // Submit
  await page.getByRole('button', { name: 'Create' }).click();

  // Assert camp appears in admin list
  await expect(page.locator('.admin-camp-item', { hasText: 'E2E Test Admin Camp' })).toBeVisible({ timeout: 10_000 });

  // Verify it appears on public camps page too
  await page.goto('/camps');
  await expect(page.locator('.service-card', { hasText: 'E2E Test Admin Camp' })).toBeVisible();
});

test('admin views registrations', async ({ page }) => {
  // Seed a camp with a registration
  const camp = await seedTestCamp({
    name: 'E2E Test Reg View Camp',
    price_cents: 5000,
  });
  await seedTestRegistration(camp.id, {
    athleteName: 'E2E Test Viewable Athlete',
    parentEmail: 'e2e-view@example.com',
    paymentStatus: 'paid',
  });

  await page.goto('/admin/camps');
  await expect(page.getByRole('heading', { name: 'Manage Camps' })).toBeVisible({ timeout: 15_000 });

  // Find the camp and expand registrations
  const campItem = page.locator('.admin-camp-item', { hasText: camp.name });
  await campItem.getByRole('button', { name: 'Registrations' }).click();

  // Assert registration table is visible with expected data
  const table = campItem.locator('.registrations-table');
  await expect(table).toBeVisible({ timeout: 10_000 });
  await expect(table).toContainText('E2E Test Viewable Athlete');
  await expect(table).toContainText('e2e-view@example.com');
  await expect(table).toContainText('paid');
});
