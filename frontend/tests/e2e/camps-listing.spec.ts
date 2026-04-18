import { test, expect } from '@playwright/test';
import { seedTestCamp, type TestCamp } from './helpers/db';

let camp: TestCamp;

test.beforeAll(async () => {
  camp = await seedTestCamp({
    name: 'E2E Test Summer Hitting Camp',
    description: 'A camp for testing the listing page',
    price_cents: 15000,
  });
});

test('camp listing page loads and shows camp cards', async ({ page }) => {
  await page.goto('/camps');

  await expect(page.getByRole('heading', { name: 'Upcoming Camps' })).toBeVisible();

  const card = page.locator('.service-card', { hasText: camp.name });
  await expect(card).toBeVisible();
  await expect(card.locator('.price')).toContainText('$150.00');
  await expect(card.locator('.duration')).toBeVisible();
  await expect(card.locator('.description')).toContainText('A camp for testing the listing page');
});

test('Register Now navigates to registration form', async ({ page }) => {
  await page.goto('/camps');

  const card = page.locator('.service-card', { hasText: camp.name });
  await card.getByRole('link', { name: 'Register Now' }).click();

  await expect(page).toHaveURL(new RegExp(`/camps/${camp.slug}/register`));
  await expect(page.getByRole('heading', { name: `Register for ${camp.name}` })).toBeVisible();
  await expect(page.locator('.camp-info-banner')).toContainText('$150.00');
});
