# Camp Registration with Stripe + PayPal

## What this builds

Public camp registration: athlete (name/age/years played/position) registers for a camp, pays via embedded Stripe CC form or PayPal modal. Parent (User) association is optional — parent can sign up later. Payment is required. Admin UI for camp CRUD + viewing registrations. Confirmation emails to parent + admin on payment.

## Schema

Three tables. Migrations go in `backend/db/migrations/` continuing existing numbering. Reuse `update_updated_at_column()` trigger from migration 000001.

**camps**: id, name, description, start_date (DATE), end_date (DATE), price_cents (INT), max_capacity (INT nullable=unlimited), is_active (BOOL default true), created_at, updated_at

**athletes**: id, name, age, years_played (default 0), position (VARCHAR(100)), user_id (FK users nullable, ON DELETE SET NULL), parent_email (NOT NULL), parent_phone, created_at, updated_at. Indexes on user_id and parent_email.

**camp_registrations**: id, athlete_id (FK athletes CASCADE), camp_id (FK camps CASCADE), payment_status (ENUM: pending/paid/failed/refunded), payment_method (ENUM: stripe/paypal), stripe_payment_intent_id, paypal_order_id, amount_cents (snapshot of price), parent_email, created_at, updated_at. UNIQUE(athlete_id, camp_id). Indexes on camp_id, athlete_id, stripe_payment_intent_id.

## Backend

Go/Gin. Raw SQL via `database/sql`. Follow existing patterns: services use `DB` interface from `services/schedule_service.go:12`, controllers use `*gin.Context`.

### Dependencies
- `github.com/stripe/stripe-go/v76` (PayPal uses `net/http` directly)

### Env vars
```
STRIPE_SECRET_KEY, STRIPE_WEBHOOK_SECRET
PAYPAL_CLIENT_ID, PAYPAL_CLIENT_SECRET, PAYPAL_API_BASE (default: https://api-m.sandbox.paypal.com)
```

### Files to create

**models/**: `camp.go`, `athlete.go`, `camp_registration.go` — structs with json/binding tags, PaymentStatus/PaymentMethod string type constants.

**services/camp_service.go**: GetActiveCamps (is_active=true, ORDER BY start_date), GetAllCamps, GetCampByID, CreateCamp, UpdateCamp, DeactivateCamp (soft-delete), GetCampRegistrationCount (count where status IN pending/paid).

**services/registration_service.go**:
- `CreateAthleteAndRegistration(athlete, campID, paymentMethod)` — transaction: validate camp active + check capacity, insert athlete, insert pending registration. Return registration.
- `InitiateStripePayment(regID)` — create PaymentIntent (amount from reg, USD, metadata with reg/camp/athlete IDs), store PI ID on reg, return client_secret.
- `ConfirmStripePayment(regID)` — verify PI status via Stripe API, update to paid. Idempotent.
- `HandleStripeWebhook(payload, sig)` — verify sig, handle payment_intent.succeeded/failed. Update by stripe_payment_intent_id. Skip if already paid (idempotent).
- `CreatePayPalOrder(regID)` — OAuth2 client_credentials for token (cache with expiry + mutex), POST /v2/checkout/orders, store order ID, return it.
- `CapturePayPalOrder(regID, orderID)` — POST /v2/checkout/orders/{id}/capture, verify COMPLETED, update to paid.
- `GetRegistrationByID`, `GetRegistrationsByCampID` (JOIN athletes), `GetAthleteByID`

**controllers/camp_controller.go**: GetActiveCamps (public, include registered_count + spots_remaining), GetCampByID (public), CreateCamp/UpdateCamp/DeactivateCamp (admin-only via userService.IsAdmin check).

**controllers/registration_controller.go**:
- `RegisterForCamp` — PUBLIC. Bind `{athlete, camp_id, parent_email, parent_phone, payment_method}`. Creates athlete+reg, then either InitiateStripePayment (returns `{registration_id, client_secret}`) or CreatePayPalOrder (returns `{registration_id, paypal_order_id}`).
- `ConfirmStripePayment` — PUBLIC. Bind `{registration_id}`. Verifies+updates. Sends emails.
- `HandleStripeWebhook` — PUBLIC. Read raw body with `io.ReadAll(c.Request.Body)` (NOT ShouldBindJSON — need raw for sig verification). Sends emails on paid.
- `CapturePayPalPayment` — PUBLIC. Bind `{registration_id, paypal_order_id}`. Captures+updates. Sends emails.
- `GetCampRegistrations` — PROTECTED+ADMIN.

### Routes (in main.go)
```
Public:  GET /api/camps, GET /api/camps/:id
         POST /api/register, POST /api/register/stripe-confirm, POST /api/register/paypal-capture
         POST /api/webhooks/stripe
Admin:   POST /api/camps, PUT /api/camps/:id, DELETE /api/camps/:id, GET /api/camps/:id/registrations
```

### Emails
Add to `EmailServiceInterface`: `SendCampRegistrationConfirmation(reg, athlete, camp)` and `SendAdminCampRegistrationNotification(reg, athlete, camp)`. Add EmailType constants + update String(). Add templates to load map. Create `CampRegistrationEmailData` struct (Athlete, Camp, Amount string, RegTime). Templates: `templates/email/camp_registration_confirmation.html`, `templates/email/admin_camp_registration.html`. Update any test mocks implementing EmailServiceInterface.

## Frontend

React 19 + Vite. Vanilla fetch, vanilla CSS, React Router 7, Auth0 React SDK.

### Dependencies
```
@stripe/react-stripe-js @stripe/stripe-js @paypal/react-paypal-js
```

### Env vars
```
VITE_STRIPE_PUBLISHABLE_KEY, VITE_PAYPAL_CLIENT_ID
```

### Pages

**pages/CampsPage.jsx**: Fetch GET /api/camps (no auth). Render grid of camp cards (reuse `services-grid`/`service-card` CSS). Show name, dates, price (cents->dollars), description, spots remaining. "Register Now" links to `/camps/:id/register`.

**pages/CampRegistrationPage.jsx**: Fetch camp by ID. Three sections: camp info banner (read-only), athlete form (name/age/years_played/position/parent_email/parent_phone using `form-group` CSS), payment. Payment section has tab toggle (Credit Card / PayPal). Stripe tab: wrap in `<Elements stripe={stripePromise}>`, render `<CardElement>`, on submit POST /api/register then `stripe.confirmCardPayment(clientSecret)` then POST /api/register/stripe-confirm. PayPal tab: `<PayPalScriptProvider>` + `<PayPalButtons>`, createOrder calls POST /api/register (store registration_id in state), onApprove calls POST /api/register/paypal-capture. Success state shows confirmation.

**pages/AdminCampsPage.jsx**: Protected route. Check isAdmin via GET /api/users/me. Camp CRUD form. List all camps with edit/deactivate. Expandable registration viewer per camp (GET /api/camps/:id/registrations as admin).

### Routing
Add to App.jsx: `/camps` -> CampsPage, `/camps/:campId/register` -> CampRegistrationPage, `/admin/camps` -> ProtectedRoute > AdminCampsPage. Add "Camps" link to Nav.jsx.

### CSS (styles/camps.css)
Payment method tabs (border-bottom active indicator), card-element-wrapper (border + padding for Stripe CardElement), registration success state, admin camp list + registrations table, status badges (paid=green, pending=yellow, failed=red).

## Key gotchas
- Stripe webhook must read raw body before any parsing (for signature verification)
- PayPal OAuth token needs caching with mutex (expires, re-fetch when stale)
- Store amount_cents on registration at creation time (snapshot, not live camp price)
- PayPal createOrder returns orderID to PayPal SDK, but registration_id must be stored in React state for the onApprove callback
- Capacity check should count pending+paid registrations
- UNIQUE(athlete_id, camp_id) prevents duplicate registrations
