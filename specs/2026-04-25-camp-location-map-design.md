# Camp Location with Google Maps — Design Spec

**Date:** 2026-04-25
**Status:** Approved (ready for implementation plan)

## Summary

Add a physical location to each camp. Display address and an embedded Google Map on a new camp detail page, with a click-through link to full Google Maps. Location data is stored in its own table with deduplication by `(street, zip)`, so multiple camps can reuse the same location.

## Goals

- Admins enter a required street address (plus city/state/zip, optional venue name) when creating or editing a camp.
- Public-facing users see address + an embedded keyless Google Maps iframe on a new camp detail page.
- Addresses are reusable across camps (no duplicate rows for identical venues).
- No Google Maps API key required.

## Non-Goals

- Geocoding to lat/lng (not needed — iframe accepts address strings).
- Interactive map on camp list page (address text on list not included; list links straight to detail page).
- International support beyond a `country` field defaulting to `US`.
- A standalone Locations admin CRUD page (creation/edit happens embedded in the camp form).

## Data Model

### New table: `locations`

```sql
CREATE TABLE locations (
    id SERIAL PRIMARY KEY,
    venue_name VARCHAR(255),
    street VARCHAR(255) NOT NULL,
    city VARCHAR(255) NOT NULL,
    state VARCHAR(64) NOT NULL,
    zip VARCHAR(32) NOT NULL,
    country VARCHAR(64) NOT NULL DEFAULT 'US',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (street, zip)
);

CREATE TRIGGER update_locations_updated_at
    BEFORE UPDATE ON locations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

- `venue_name` optional (e.g. "Springfield HS Baseball Field").
- `(street, zip)` unique — enforces dedup at DB level.
- Upsert strategy for inserts/updates: `INSERT ... ON CONFLICT (street, zip) DO UPDATE SET venue_name = EXCLUDED.venue_name, city = EXCLUDED.city, state = EXCLUDED.state, country = EXCLUDED.country, updated_at = CURRENT_TIMESTAMP RETURNING id`.

### Alter `camps`

```sql
ALTER TABLE camps ADD COLUMN location_id INTEGER REFERENCES locations(id);
```

Nullable at DB level. Two-phase migration approach:

- **Phase 1 (this spec):** add nullable column; admin backfills existing camps via UI.
- **Phase 2 (separate, later):** `ALTER COLUMN location_id SET NOT NULL` once all camps have locations assigned.

API-level validation requires `location` on camp create and update from day one.

### Go model

```go
// backend/models/location.go
type Location struct {
    ID        int       `json:"id"`
    VenueName *string   `json:"venue_name"`
    Street    string    `json:"street" binding:"required"`
    City      string    `json:"city" binding:"required"`
    State     string    `json:"state" binding:"required"`
    Zip       string    `json:"zip" binding:"required"`
    Country   string    `json:"country"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

`Camp` struct adds `LocationID *int` (nullable FK). Response DTO (`CampWithSpots` or similar) adds nested `Location *Location`.

## API

### New endpoints

**`GET /api/locations`** — admin only (JWT + `isAdmin` check).

Returns all locations for the admin dropdown:

```json
[
  {
    "id": 1,
    "venue_name": "Springfield HS Baseball Field",
    "street": "123 Main St",
    "city": "Springfield",
    "state": "IL",
    "zip": "62701",
    "country": "US"
  }
]
```

No separate `POST/PUT/DELETE /api/locations` endpoints — locations are created/updated as a side effect of camp create/update.

### Changed endpoints

`POST /api/camps` and `PUT /api/camps/:id` request body adds a required `location` object:

```json
{
  "name": "Summer Pitching Camp",
  "description": "...",
  "start_date": "2026-06-10T00:00:00Z",
  "end_date": "2026-06-14T00:00:00Z",
  "price": 299.00,
  "max_capacity": 30,
  "slug": "summer-pitching-camp",
  "age_groups": [],
  "location": {
    "id": null,
    "venue_name": "Springfield HS Baseball Field",
    "street": "123 Main St",
    "city": "Springfield",
    "state": "IL",
    "zip": "62701",
    "country": "US"
  }
}
```

Server handling:

1. Validate all required fields on `location`. Return 400 if missing or invalid.
2. Server ignores `location.id` in the request body. Always performs upsert keyed by `(street, zip)`. Dedup is enforced by the unique constraint, not by the client-provided id.
3. Upsert via `INSERT ... ON CONFLICT (street, zip) DO UPDATE SET venue_name = EXCLUDED.venue_name, city = EXCLUDED.city, state = EXCLUDED.state, country = EXCLUDED.country, updated_at = CURRENT_TIMESTAMP RETURNING id` returns the row id (existing or new).
4. Set `camp.location_id` to the returned id.

`GET /api/camps` and `GET /api/camps/:slug` responses include the nested `location` object (LEFT JOIN, so camps without location return `"location": null`).

## Frontend

### New routes/files

- **`frontend/src/pages/CampDetailPage.jsx`** — route `/camps/:slug`.
  - Fetches `GET /api/camps/:slug`.
  - Renders camp name, date range, price (or age group breakdown), description.
  - Location block: venue name (if present), formatted address, embedded map, "Open in Google Maps" link.
  - "Register Now" button → `/camps/:slug/register` (disabled when camp is full).

- **`frontend/src/components/LocationMapEmbed.jsx`** — presentational component.
  - Props: `{ address }` (pre-formatted string).
  - Renders `<iframe src="https://www.google.com/maps?q={encodedAddress}&output=embed">` with responsive sizing.
  - Below iframe: external link to `https://www.google.com/maps/search/?api=1&query={encodedAddress}` with `target="_blank"` and `rel="noopener noreferrer"`.

- **`frontend/src/components/LocationPicker.jsx`** — admin sub-form.
  - Props: `{ value, onChange, existingLocations }`.
  - Dropdown: "— Enter new address below —" + one option per existing location (display: `"{venue_name || street} — {street}, {city} {state} {zip}"`).
  - Selecting a dropdown option fills the form fields but keeps them editable.
  - Fields: `venue_name` (optional), `street`, `city`, `state`, `zip`, `country` (default "US").
  - Shows "✓ Matches existing location — will reuse on save" when `(street, zip)` matches any row in `existingLocations`.
  - Emits full location object to parent via `onChange`.

- **Address formatting helper:** `frontend/src/utils/formatAddress.js`
  - `formatAddress(location)` → `"{street}, {city}, {state} {zip}"` (append `, {country}` if not "US").
  - Used for iframe query and on-page display.

### Changed files

- **`frontend/src/App.jsx`** — add `<Route path="/camps/:slug" element={<CampDetailPage />} />`. React Router v6 matches the more specific route automatically, so this coexists with the existing `/camps/:slug/register`.

- **`frontend/src/pages/CampsPage.jsx`** — change the card button:
  - Text: "Register Now" → "View Details".
  - `to`: `/camps/${camp.slug}/register` → `/camps/${camp.slug}`.
  - Full-state behavior unchanged (button still disabled when full, text "Full" stays).

- **`frontend/src/pages/AdminCampsPage.jsx`** —
  - On mount (when `isAdmin`), fetch `GET /api/locations` and store in state.
  - Render `<LocationPicker>` inside the camp form (below date fields, above capacity mode toggle).
  - Include `location` in create/update payload.
  - `handleEdit` seeds the picker with `camp.location`.
  - Validate location fields before submit (parallel to existing age-group validation).

## Admin UX Flow

1. Admin opens "Create New Camp".
2. Fills name, description, dates, slug.
3. In location section: either picks an existing location from the dropdown (fields auto-fill, still editable) or enters a new address directly.
4. If an edit matches an existing `(street, zip)`, UI shows the reuse hint.
5. Saves. Server upserts location, sets camp FK.

For existing camps (pre-migration), admin edits each camp and adds a location. Until then, the camp's public detail page shows "Location TBA" placeholder text.

## Public UX Flow

1. User visits `/camps` — cards show name, dates, price, description, "View Details" button.
2. Click "View Details" → `/camps/:slug` detail page.
3. Detail page shows all camp info + location block (venue, address, embedded map).
4. User can click the iframe or "Open in Google Maps" link to open full Google Maps in a new tab.
5. "Register Now" button → `/camps/:slug/register` (existing flow unchanged).

## Error Handling

- Missing location fields on POST/PUT camp: 400 with field-level error messages.
- Camp with `location_id` null on detail page: show "Location TBA" in place of map block. Do not render iframe.
- Iframe fails to load: browser shows its own iframe error; external "Open in Google Maps" link still works as fallback.
- `GET /api/locations` failure on admin page: fall back to blank dropdown (admin can still enter new address manually).

## Migrations

File: `backend/db/migrations/000014_create_locations_table.up.sql` / `.down.sql`

Up:
```sql
CREATE TABLE locations (...);  -- as above
CREATE TRIGGER update_locations_updated_at ...;
ALTER TABLE camps ADD COLUMN location_id INTEGER REFERENCES locations(id);
```

Down:
```sql
ALTER TABLE camps DROP COLUMN location_id;
DROP TABLE locations;
```

Phase-2 migration (future, separate PR, not part of this spec) will add the `NOT NULL` constraint.

## Testing

### Backend unit

- `backend/services/location_service_test.go`
  - Upsert inserts new row when `(street, zip)` is new.
  - Upsert returns existing id and updates `venue_name` when `(street, zip)` matches.
  - `GetAllLocations` returns all rows.

- `backend/services/camp_service_test.go` (extend)
  - Create/update camp with location correctly sets `location_id` FK.
  - `GetCampBySlug` and `GetActiveCamps` return nested location via LEFT JOIN.

- `backend/controllers/camp_controller_test.go` (extend)
  - POST/PUT camp missing `location` → 400.
  - POST camp with valid location → 201 and correct FK.
  - POST camp with missing location sub-field (e.g. `zip`) → 400.

### Frontend unit (Vitest)

- `LocationMapEmbed.test.jsx` — renders iframe with correct encoded src; renders external link with `target="_blank"`.
- `LocationPicker.test.jsx` — dropdown fills fields; fields stay editable post-select; dedup hint toggles on match; `onChange` emits merged object.
- `CampDetailPage.test.jsx` — renders camp fields, location block, Register link; handles null location (shows "Location TBA").
- `CampsPage.test.jsx` (extend) — button text is "View Details" with correct href.

### E2E (Playwright)

- `frontend/tests/camp-location.spec.ts`
  - Admin: log in, create camp with new location (inline fields), verify camp persisted with location.
  - Admin: create second camp, select existing location from dropdown, submit, verify reuse (API check: location count unchanged).
  - Public: navigate `/camps` → click "View Details" → detail page shows address + iframe + Register link → Register button routes to `/camps/:slug/register`.

All tiers must pass via `make test-all`.

## Rollout

1. Merge migration + backend + frontend together in one PR.
2. Deploy via existing Render pipeline (`/deploy` skill).
3. Admin backfills location on all existing camps through the admin UI.
4. Future separate PR: phase-2 migration to set `location_id NOT NULL`.
