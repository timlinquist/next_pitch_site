import { test } from '@playwright/test';

// PayPal renders in an iframe/popup that requires sandbox buyer credentials
// and is notoriously brittle to automate. Options to revisit:
//   1. Drive the PayPal sandbox popup UI (fragile, depends on PayPal markup)
//   2. Intercept at API level and mock the capture response
//   3. Test PayPal flow manually, only automate the pre-PayPal form submission

test.skip('PayPal registration - full happy path', async ({ page }) => {
  // TODO: Implement after Stripe tests are green and PayPal approach is decided
  // - Fill athlete form
  // - Select PayPal tab
  // - Click PayPal button → sandbox popup
  // - Log in with sandbox buyer, approve payment
  // - Assert success confirmation
  // - Assert DB registration with payment_status = 'paid'
});
