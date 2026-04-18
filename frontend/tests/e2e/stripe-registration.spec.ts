import { test, expect } from '@playwright/test';
import { seedTestCamp, getRegistration, type TestCamp } from './helpers/db';
import { fillStripeCard } from './helpers/stripe';

let camp: TestCamp;

test.beforeAll(async () => {
  camp = await seedTestCamp({
    name: 'E2E Test Stripe Camp',
    price_cents: 5000,
  });
});

async function fillAthleteForm(page: import('@playwright/test').Page) {
  await page.locator('#name').fill('E2E Test Athlete');
  await page.locator('#age').fill('12');
  await page.locator('#yearsPlayed').fill('3');
  await page.locator('#position').fill('Pitcher');
  await page.locator('#parentEmail').fill('e2e-test@example.com');
  await page.locator('#parentPhone').fill('555-0199');
}

test('form validation blocks incomplete submissions', async ({ page }) => {
  await page.goto(`/camps/${camp.slug}/register`);
  await expect(page.getByRole('heading', { name: `Register for ${camp.name}` })).toBeVisible();

  const payButton = page.locator('button.register-btn');
  await payButton.click();

  // HTML5 validation should prevent submission — name field is required and empty
  const nameInput = page.locator('#name');
  await expect(nameInput).toHaveAttribute('required', '');

  // Page should still be on registration (no navigation, no success)
  await expect(page).toHaveURL(new RegExp(`/camps/${camp.slug}/register`));
  await expect(page.locator('.registration-success')).not.toBeVisible();
});

test('successful Stripe payment', async ({ page }) => {
  await page.goto(`/camps/${camp.slug}/register`);
  await expect(page.getByRole('heading', { name: `Register for ${camp.name}` })).toBeVisible();

  await fillAthleteForm(page);

  // Credit Card tab should be active by default
  await expect(page.locator('.payment-tab.active')).toContainText('Credit Card');

  // Wait for Stripe iframe to load
  await page.waitForSelector('iframe[name*="__privateStripeFrame"]', { timeout: 15_000 });
  await fillStripeCard(page);

  const payButton = page.locator('button.register-btn');
  await payButton.click();

  // Wait for success confirmation
  await expect(page.locator('.registration-success')).toBeVisible({ timeout: 30_000 });
  await expect(page.getByRole('heading', { name: 'Registration Confirmed!' })).toBeVisible();
  await expect(page.locator('.registration-success')).toContainText(camp.name);
  await expect(page.locator('.success-details')).toContainText('$50.00');

  // Verify DB state
  const reg = await getRegistration(camp.id);
  expect(reg).toBeTruthy();
  expect(reg.payment_status).toBe('paid');
  expect(reg.athlete_name).toBe('E2E Test Athlete');
  expect(reg.parent_email).toBe('e2e-test@example.com');
});

test('declined card shows error', async ({ page }) => {
  await page.goto(`/camps/${camp.slug}/register`);
  await expect(page.getByRole('heading', { name: `Register for ${camp.name}` })).toBeVisible();

  await fillAthleteForm(page);

  await page.waitForSelector('iframe[name*="__privateStripeFrame"]', { timeout: 15_000 });
  // 4000000000000002 is Stripe's generic decline test card
  await fillStripeCard(page, '4000000000000002');

  const payButton = page.locator('button.register-btn');
  await payButton.click();

  // Should show error, not success
  await expect(page.locator('.payment-error').first()).toBeVisible({ timeout: 30_000 });
  await expect(page.locator('.registration-success')).not.toBeVisible();
});
