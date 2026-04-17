import { test } from '@playwright/test';

// Admin tests require Auth0 login. Options to discuss:
//   1. Drive Auth0 Universal Login page (brittle, depends on Auth0 markup)
//   2. Inject a valid JWT into localStorage (reliable but needs token crafting)
//   3. Use Auth0 Resource Owner Password Grant for headless login
//   4. Log in once manually, seed is_admin via DB, reuse session storage

test.skip('admin creates a camp', async ({ page }) => {
  // TODO: Implement after auth strategy is decided
  // - Log in as admin (Auth0)
  // - Navigate to /admin/camps
  // - Fill camp form, submit
  // - Assert camp appears in admin list
  // - Assert camp appears on public /camps
});

test.skip('admin views registrations', async ({ page }) => {
  // TODO: Implement after auth strategy is decided
  // - Log in as admin
  // - Navigate to /admin/camps
  // - Click "Registrations" on a camp with registrations
  // - Assert table shows athlete name, parent email, payment status
});
