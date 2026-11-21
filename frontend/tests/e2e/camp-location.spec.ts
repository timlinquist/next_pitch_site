import { test, expect } from '@playwright/test';
import { authStoragePath } from './helpers/auth';
import { countLocations, seedTestCamp, seedTestLocation } from './helpers/db';

test.describe('camp location admin + public', () => {
  test.use({ storageState: authStoragePath });

  test('admin reuses existing location via dropdown (no duplicate row)', async ({ page }) => {
    const sharedLoc = await seedTestLocation({
      venue_name: 'E2E Test Shared Venue',
      street: '888 Shared Ave',
      city: 'Sharedtown',
      state: 'IL',
      zip: '61111',
    });

    const before = await countLocations();

    await page.goto('/admin/camps');
    await expect(page.getByRole('heading', { name: 'Manage Camps' })).toBeVisible({ timeout: 15_000 });
    await page.getByRole('button', { name: 'Create New Camp' }).click();

    await page.locator('#camp-name').fill('E2E Test Location Reuse Camp');
    await page.locator('#camp-desc').fill('shared loc');
    await page.locator('#camp-start').fill('2026-10-01');
    await page.locator('#camp-end').fill('2026-10-02');
    await page.locator('#camp-price').fill('50');

    await page.locator('#location-existing').selectOption({ value: String(sharedLoc.id) });

    await page.getByRole('button', { name: 'Create' }).click();
    await expect(page.locator('.admin-camp-item', { hasText: 'E2E Test Location Reuse Camp' })).toBeVisible({ timeout: 10_000 });

    const after = await countLocations();
    expect(after).toBe(before);
  });

  test('public detail page shows address, map iframe, and register link', async ({ browser }) => {
    const anonymous = await browser.newContext();
    const anonPage = await anonymous.newPage();

    const loc = await seedTestLocation({
      venue_name: 'E2E Test Public Venue',
      street: '42 Public Way',
      city: 'Pubtown',
      state: 'IL',
      zip: '62000',
    });
    const camp = await seedTestCamp({
      name: 'E2E Test Detail Page Camp',
      price_cents: 7500,
      location_id: loc.id,
    });

    await anonPage.goto(`/camps/${camp.slug}`);
    await expect(anonPage.getByRole('heading', { name: camp.name })).toBeVisible({ timeout: 15_000 });
    await expect(anonPage.getByText('E2E Test Public Venue')).toBeVisible();
    await expect(anonPage.getByText(/42 Public Way, Pubtown, IL 62000/)).toBeVisible();
    const iframe = anonPage.locator('iframe[src*="google.com/maps"]');
    await expect(iframe).toHaveAttribute('src', /google\.com\/maps.*output=embed/);

    const register = anonPage.getByRole('link', { name: /Register Now/i });
    await expect(register).toHaveAttribute('href', `/camps/${camp.slug}/register`);

    await anonymous.close();
  });

  test('list page routes View Details to detail page', async ({ browser }) => {
    const loc = await seedTestLocation({
      venue_name: 'E2E Test List Venue',
      street: '9 List St',
      city: 'Listville',
      state: 'IL',
      zip: '62999',
    });
    const camp = await seedTestCamp({
      name: 'E2E Test List Detail Camp',
      price_cents: 6000,
      location_id: loc.id,
    });

    const anonymous = await browser.newContext();
    const anonPage = await anonymous.newPage();

    await anonPage.goto('/camps');
    const card = anonPage.locator('.service-card', { hasText: camp.name });
    await expect(card).toBeVisible({ timeout: 15_000 });
    await card.getByRole('link', { name: 'View Details' }).click();
    await expect(anonPage).toHaveURL(new RegExp(`/camps/${camp.slug}$`));

    await anonymous.close();
  });
});
