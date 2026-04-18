import { type Page } from '@playwright/test';

export async function fillStripeCard(
  page: Page,
  cardNumber = '4242424242424242',
  expiry = '1230',
  cvc = '123',
  zip = '90210'
) {
  const stripeFrame = page
    .frameLocator('iframe[name*="__privateStripeFrame"]')
    .first();

  const cardField = stripeFrame.locator('[name="cardnumber"]');
  await cardField.click();
  await cardField.pressSequentially(cardNumber, { delay: 50 });

  const expField = stripeFrame.locator('[name="exp-date"]');
  await expField.click();
  await expField.pressSequentially(expiry, { delay: 50 });

  const cvcField = stripeFrame.locator('[name="cvc"]');
  await cvcField.click();
  await cvcField.pressSequentially(cvc, { delay: 50 });

  // ZIP field — Tab from CVC within the iframe to move focus to postal code
  await cvcField.press('Tab');
  await page.waitForTimeout(200);
  // Type into whatever is now focused inside the stripe iframe
  await stripeFrame.locator(':focus').pressSequentially(zip, { delay: 50 });
}
