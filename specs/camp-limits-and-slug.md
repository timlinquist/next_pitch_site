# Camp Slug URLs + Age-Range Capacity & Pricing

## Overview

Two capacity modes per camp:
- **Simple**: single `max_capacity` + single `price_cents` on the camp
- **Age-range**: per-age-group rows, each with its own `max_capacity` and `price_cents`. Camp-level `price_cents` is null.

Camps use URL slugs (auto-generated from name, editable) instead of numeric IDs in public routes.

## Schema

**Migration 000011** — slug on camps:
```sql
ALTER TABLE camps ADD COLUMN slug VARCHAR(255);
CREATE UNIQUE INDEX idx_camps_slug ON camps(slug);
```

**Migration 000012** — age groups table:
```sql
CREATE TABLE camp_age_groups (
    id SERIAL PRIMARY KEY,
    camp_id INTEGER NOT NULL REFERENCES camps(id) ON DELETE CASCADE,
    min_age INTEGER NOT NULL,
    max_age INTEGER NOT NULL,
    max_capacity INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_age_range CHECK (min_age <= max_age),
    CONSTRAINT chk_capacity_positive CHECK (max_capacity > 0)
);
CREATE INDEX idx_camp_age_groups_camp_id ON camp_age_groups(camp_id);
```

**Migration 000013** — price on age groups, nullable price on camps:
```sql
ALTER TABLE camp_age_groups ADD COLUMN price_cents INTEGER NOT NULL;
ALTER TABLE camp_age_groups ADD CONSTRAINT chk_price_cents_positive CHECK (price_cents > 0);
ALTER TABLE camps ALTER COLUMN price_cents DROP NOT NULL;
```

Presence of `camp_age_groups` rows = age-range mode. No rows = simple mode.

## Backend

Go/Gin, raw SQL via `database/sql`.

### Models

`Camp.PriceCents` is `*int` (nullable). `Camp.Slug` is `*string`.

`CampAgeGroup` has: `ID`, `CampID`, `MinAge`, `MaxAge`, `MaxCapacity`, `PriceCents` (all `int`), `CreatedAt`, `UpdatedAt`.

### Camp service

All camp queries include `slug` column. Additional methods:
- `GetCampBySlug(slug)` — `WHERE slug = $1`
- `GenerateSlug(name)` — lowercase, non-alphanumeric → hyphens, trim
- `GetAgeGroupsByCampID(campID)` — SELECT including `price_cents`, ordered by `min_age`
- `SetAgeGroups(campID, groups)` — transaction: DELETE then INSERT (includes `price_cents`)
- `ValidateAgeGroups(groups)` — sort by `min_age`, reject pairwise overlap
- `GetAgeGroupRegistrationCount(campID, minAge, maxAge)` — count via JOIN athletes WHERE age BETWEEN

`CreateCamp` auto-generates slug from name if not provided.

### Registration service

`CreateAthleteAndRegistration` capacity + pricing logic:
1. Fetch age groups for camp
2. If age groups exist → match by `athlete.Age`. No match → reject. Match → check per-group count. Full → reject. **Use matched group's `PriceCents`** for the registration.
3. Else → use camp-level `PriceCents` and flat `max_capacity` check
4. Snapshot price into `registration.AmountCents`

### Controller

Response wraps camps with computed availability:
- `CampWithSpots`: camp + `registered_count` + `spots_remaining` (nil when age-range) + `age_groups` array
- `AgeGroupWithSpots`: group + `registered_count` + `spots_remaining` + `price_cents`

Create/Update accept `age_groups []CampAgeGroup`. Mutually exclusive: age groups + `max_capacity` both set → 400. Age-range mode nulls `max_capacity`.

### Routes

```
GET  /api/camps/by-slug/:slug  (public, before :id route)
```

All other existing camp routes unchanged.

## Frontend

### Routing

`/camps/:slug/register` (was `:campId`). Fetch via `camps/by-slug/${slug}`.

### CampsPage

- Links use `camp.slug`
- Age-range camps: show per-group line with price + spots (e.g. "Ages 8-10: $75.00 — 5 spots remaining")
- Simple camps: show single price + spots
- Full = all groups full OR flat `spots_remaining === 0`

### CampRegistrationPage

- `useParams()` extracts `{ slug }`
- `getMatchedAgeGroup()` finds group by athlete age
- `getDisplayPrice()` returns matched group's `price_cents` or camp's
- Camp info banner: per-group prices + availability, or single price
- Age input feedback: shows matched group price + spots, or "not eligible"
- Stripe pay button + success page use `getDisplayPrice()`

### AdminCampsPage

- Slug field: auto-generates from name, editable
- Capacity mode radio: Simple / Age Range
- **Simple mode**: price input (dollars, `step="0.01"`) + max capacity
- **Age-range mode**: dynamic row editor — each row: min age, max age, capacity, price (dollars). "Add Age Group" button.
- **Dollar ↔ cents conversion**: form displays/accepts dollars. On submit: `Math.round(parseFloat(price) * 100)`. On edit load: `price_cents / 100`.
- Payload: age-range sends `age_groups` with `price_cents` + null camp `price_cents`. Simple sends camp `price_cents` + empty `age_groups`.
- Client-side overlap validation before submit.
- Camp list: shows per-group info (price + registered/capacity) or single price.

## E2E Tests

- `helpers/db.ts`: `TestCamp` includes `slug`. `seedTestCamp` generates slug. `seedTestCampAgeGroups` accepts `price_cents`.
- `camps-listing.spec.ts`: URL assertions use slug
- `admin-camps.spec.ts`: simple camp test fills dollar price. Age-range test fills per-group dollar prices.
- `stripe-registration.spec.ts`: URLs use `camp.slug`

## Verification

Run `make test-all` (backend unit + frontend unit + E2E). The E2E suite covers:
1. Admin creates simple camp (slug auto-generates, dollar price)
2. Admin creates age-range camp (per-group prices in dollars)
3. Admin views registrations
4. Public camp listing shows slug-based links
5. Registration navigates via slug
6. Stripe payment flow end-to-end
7. Declined card error handling
