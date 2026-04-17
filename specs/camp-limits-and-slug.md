# Camp Slug URLs + Age-Range Capacity Limits

## What this builds

Two features on the existing camp system:

1. **Slug-based URLs** — camps get a URL slug (e.g., camp "ELL" → `/camps/ell`). Admin sets slug when creating/editing; auto-generated from name, editable. Public registration pages use slug instead of numeric ID.

2. **Age-range capacity limits** — two mutually exclusive capacity modes per camp:
   - **Simple mode** (existing): single `max_capacity` for the whole camp
   - **Age-range mode**: per-age-group limits (e.g., ages 8-10: 10 spots, ages 11-13: 10 spots). Athlete's age must fall in a defined group to register. Each group enforces its own capacity independently.

## Schema

Migrations continue existing numbering (latest is `000010`). Reuse `update_updated_at_column()` trigger.

**Migration 000011** — add slug to camps:
```sql
ALTER TABLE camps ADD COLUMN slug VARCHAR(255);
CREATE UNIQUE INDEX idx_camps_slug ON camps(slug);
```
Nullable so existing rows aren't broken. App enforces on create/update.

**Migration 000012** — age groups table:
```sql
CREATE TABLE IF NOT EXISTS camp_age_groups (
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
Presence of rows in this table for a camp = age-range mode. No rows = simple mode (uses existing `max_capacity` on camps table).

Write down migrations for both.

## Backend

Go/Gin. Raw SQL via `database/sql`. Follow existing patterns in `services/camp_service.go` and `controllers/camp_controller.go`.

### Models

**`models/camp.go`** — add field:
```go
Slug *string `json:"slug"`
```

**`models/camp_age_group.go`** (new file):
```go
type CampAgeGroup struct {
    ID          int       `json:"id"`
    CampID      int       `json:"camp_id"`
    MinAge      int       `json:"min_age" binding:"required"`
    MaxAge      int       `json:"max_age" binding:"required"`
    MaxCapacity int       `json:"max_capacity" binding:"required"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### Service changes (`services/camp_service.go`)

All existing SELECT/INSERT/UPDATE queries must add `slug` column + `&camp.Slug` in Scan calls. Affected methods: `GetActiveCamps`, `GetAllCamps`, `GetCampByID`, `CreateCamp`, `UpdateCamp`.

New methods:
- `GetCampBySlug(slug string) (*Camp, error)` — like `GetCampByID` but `WHERE slug = $1`
- `GenerateSlug(name string) string` — lowercase, replace non-alphanumeric with hyphens, collapse multiples, trim edges
- `GetAgeGroupsByCampID(campID int) ([]CampAgeGroup, error)` — SELECT ordered by min_age
- `SetAgeGroups(campID int, groups []CampAgeGroup) error` — transaction: DELETE existing for camp, INSERT new ones
- `ValidateAgeGroups(groups []CampAgeGroup) error` — sort by min_age, check pairwise overlap (`groups[i].min_age <= groups[i-1].max_age` = overlap)
- `GetAgeGroupRegistrationCount(campID, minAge, maxAge int) (int, error)` — count registrations joining athletes where `age BETWEEN minAge AND maxAge` and `payment_status IN ('pending','paid')`

`CreateCamp`: auto-generate slug from name if not provided. INSERT includes slug.
`UpdateCamp`: include slug in UPDATE SET clause.

### Registration capacity check (`services/registration_service.go`)

`CreateAthleteAndRegistration` (line 47-56) currently only checks flat `max_capacity`. Replace with:

1. Fetch age groups for camp
2. If age groups exist → find group matching `athlete.Age`. No match → reject. Match found → check per-group count via `GetAgeGroupRegistrationCount`. Full → reject.
3. Else if `max_capacity` set → existing flat check
4. Else → unlimited

### Controller changes (`controllers/camp_controller.go`)

New response types:
```go
type AgeGroupWithSpots struct {
    models.CampAgeGroup
    RegisteredCount int `json:"registered_count"`
    SpotsRemaining  int `json:"spots_remaining"`
}

type CampWithSpots struct {
    models.Camp
    RegisteredCount int                 `json:"registered_count"`
    SpotsRemaining  *int               `json:"spots_remaining"`
    AgeGroups       []AgeGroupWithSpots `json:"age_groups,omitempty"`
}
```

`GetActiveCamps` / `GetCampByID`: for each camp, fetch age groups. If groups exist, compute per-group counts; top-level `spots_remaining` becomes nil.

`GetCampBySlug` (new): same as `GetCampByID` but takes slug param.

`CreateCamp` / `UpdateCamp`: accept request struct with optional `age_groups []CampAgeGroup`. Modes are mutually exclusive — if age groups provided, clear max_capacity; if both set, return 400. Validate groups (no overlaps) then call `SetAgeGroups`.

### Routes (`main.go`)

Add public route:
```go
r.GET("/api/camps/by-slug/:slug", campController.GetCampBySlug)
```
Keep `/api/camps/:id` for admin use. Separate path avoids Gin ambiguity between numeric ID and string slug.

## Frontend

React + Vite SPA. Follow existing patterns.

### Routes (`src/App.jsx`)

Change: `<Route path="/camps/:campId/register" ...>` → `<Route path="/camps/:slug/register" ...>`

### CampsPage (`src/pages/CampsPage.jsx`)

- Links change from `/camps/${camp.id}/register` to `/camps/${camp.slug}/register`
- When camp has `age_groups`, show per-group spots instead of single `spots_remaining`
- Disable Register Now if all groups full

### CampRegistrationPage (`src/pages/CampRegistrationPage.jsx`)

- `useParams()`: `{ campId }` → `{ slug }`
- Fetch: `camps/${campId}` → `camps/by-slug/${slug}`
- When age_groups present: show per-group availability in camp info banner
- After athlete enters age: inline feedback showing which group they're in and spots remaining, or "age not eligible"

### AdminCampsPage (`src/pages/AdminCampsPage.jsx`)

Form additions:
- **Slug field** below name — auto-generates from name on change (lowercase, hyphens), admin can override
- **Capacity mode toggle** (radio): "Simple" vs "Age Range"
  - Simple: existing max_capacity input
  - Age Range: hides max_capacity, shows dynamic row editor
- **Age group editor** (visible in Age Range mode):
  - Each row: min_age (number), max_age (number), max_capacity (number), remove button
  - "Add Age Group" button
  - Client-side overlap validation before submit

Form state additions: `slug`, `capacity_mode` ('simple'|'age_range'), `age_groups` array.

Payload: in age_range mode send `age_groups` and null `max_capacity`. In simple mode send empty `age_groups` (clears existing).

Edit: detect mode from response — if `age_groups` has entries, set age_range mode.

Camp list: show slug next to name, show per-group capacity when applicable.

### CSS (`src/styles/camps.css`)

Add styles for: `.capacity-mode-toggle`, `.age-group-editor`, `.age-group-row`, `.age-group-spots`, `.slug-field`

## E2E Tests

- `tests/e2e/helpers/db.ts`: add `slug` to `seedTestCamp` INSERT + `TestCamp` type, add `seedTestCampAgeGroups` helper
- `tests/e2e/camps-listing.spec.ts`: update URL assertions to use slug
- `tests/e2e/admin-camps.spec.ts`: update create test for slug, add test for age-range mode

## Verification

1. Migrations apply: `cd backend && go run cmd/migrate/main.go -action up`
2. Backend tests: `cd backend && go test ./...`
3. Admin: create camp → slug auto-generates and is editable
4. Admin: create camp with age groups → overlapping ranges rejected client + server side
5. Public: `/camps` shows slug-based links
6. Public: `/camps/ell/register` loads correct camp
7. Registration: athlete in valid age group → succeeds
8. Registration: athlete age outside all groups → rejected
9. Registration: age group at capacity → rejected
10. Simple mode camps still work as before
11. E2E tests: `cd frontend && npm run test:e2e`
