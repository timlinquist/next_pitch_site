# Camp Location with Google Maps Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a physical location to each camp, surface it on a new `/camps/:slug` detail page with a keyless Google Maps iframe, and provide an admin inline editor with dropdown-based reuse of existing locations.

**Architecture:** Normalized `locations` table with `UNIQUE(street, zip)` for dedup via Postgres `ON CONFLICT` upsert. Go service/controller for locations, nested location object on camp request/response payloads. New React detail page + two small components (`LocationMapEmbed`, `LocationPicker`). Phase-1 migration leaves `camps.location_id` nullable; API requires the field from day one. Phase-2 migration to set NOT NULL is out of scope.

**Tech Stack:** Go 1.x + Gin + `database/sql` + Postgres; React 18 + Vite + React Router v6 + Vitest; Playwright for E2E.

**Spec:** `specs/2026-04-25-camp-location-map-design.md`

---

## File Structure

### Backend (new)
- `backend/db/migrations/000014_create_locations_table.up.sql`
- `backend/db/migrations/000014_create_locations_table.down.sql`
- `backend/models/location.go` — `Location` struct
- `backend/services/location_service.go` — `LocationService` with `GetAll`, `UpsertByStreetZip`
- `backend/services/location_service_test.go`
- `backend/controllers/location_controller.go` — `GET /api/locations`

### Backend (modified)
- `backend/models/camp.go` — add `LocationID *int` + `Location *Location` (on response)
- `backend/services/camp_service.go` — join locations on reads, accept optional location FK on create/update
- `backend/controllers/camp_controller.go` — accept nested `location` in request, upsert via `LocationService`, include in `buildCampWithSpots`
- `backend/main.go` — register `LocationController` + route
- `backend/test/*` or inline controller tests (follow existing pattern — extend `backend/main_test.go`)

### Frontend (new)
- `frontend/src/pages/CampDetailPage.jsx`
- `frontend/src/components/LocationMapEmbed.jsx`
- `frontend/src/components/LocationPicker.jsx`
- `frontend/src/utils/formatAddress.js`
- `frontend/src/components/__tests__/LocationMapEmbed.test.jsx`
- `frontend/src/components/__tests__/LocationPicker.test.jsx`
- `frontend/src/pages/CampDetailPage.test.jsx`
- `frontend/tests/e2e/camp-location.spec.ts`
- `frontend/styles/camp-detail.css` (optional — can reuse `camps.css` if simpler)

### Frontend (modified)
- `frontend/src/App.jsx` — add `/camps/:slug` route
- `frontend/src/pages/CampsPage.jsx` — button "Register Now" → "View Details", route target changes
- `frontend/src/pages/AdminCampsPage.jsx` — integrate `LocationPicker`, fetch locations on mount
- `frontend/tests/e2e/helpers/db.ts` — extend `seedTestCamp` to optionally attach a location + add cleanup for locations table
- `frontend/tests/e2e/admin-camps.spec.ts` — extend existing create flows to fill location fields (so they still pass)

---

## Task 1: Database migration

**Files:**
- Create: `backend/db/migrations/000014_create_locations_table.up.sql`
- Create: `backend/db/migrations/000014_create_locations_table.down.sql`

- [ ] **Step 1: Write up migration**

Create `backend/db/migrations/000014_create_locations_table.up.sql`:

```sql
CREATE TABLE IF NOT EXISTS locations (
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

ALTER TABLE camps ADD COLUMN location_id INTEGER REFERENCES locations(id);
```

- [ ] **Step 2: Write down migration**

Create `backend/db/migrations/000014_create_locations_table.down.sql`:

```sql
ALTER TABLE camps DROP COLUMN IF EXISTS location_id;
DROP TRIGGER IF EXISTS update_locations_updated_at ON locations;
DROP TABLE IF EXISTS locations;
```

- [ ] **Step 3: Run migration**

Run: `make migrate`

Expected: migration 000014 applies without error. Verify by connecting to DB and running `\d locations` and `\d camps` (new `location_id` column present).

- [ ] **Step 4: Commit**

```bash
git add backend/db/migrations/000014_create_locations_table.up.sql backend/db/migrations/000014_create_locations_table.down.sql
git commit -m "Add locations table and camps.location_id FK"
```

---

## Task 2: Location model

**Files:**
- Create: `backend/models/location.go`
- Modify: `backend/models/camp.go`

- [ ] **Step 1: Write Location model**

Create `backend/models/location.go`:

```go
package models

import "time"

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

- [ ] **Step 2: Extend Camp model**

Edit `backend/models/camp.go`. Replace the struct with:

```go
package models

import (
	"time"
)

type Camp struct {
	ID          int       `json:"id"`
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date" binding:"required"`
	EndDate     time.Time `json:"end_date" binding:"required"`
	Price       *float64  `json:"price"`
	MaxCapacity *int      `json:"max_capacity"`
	Slug        *string   `json:"slug"`
	IsActive    bool      `json:"is_active"`
	LocationID  *int      `json:"location_id,omitempty"`
	Location    *Location `json:"location,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
```

- [ ] **Step 3: Build backend**

Run: `make build` (or `cd backend && go build ./...`)

Expected: compiles with no errors.

- [ ] **Step 4: Commit**

```bash
git add backend/models/location.go backend/models/camp.go
git commit -m "Add Location model and location fields on Camp"
```

---

## Task 3: Location service — upsert by (street, zip)

**Files:**
- Create: `backend/services/location_service.go`
- Create: `backend/services/location_service_test.go`

- [ ] **Step 1: Write failing test for upsert (insert path)**

Create `backend/services/location_service_test.go`:

```go
package services

import (
	"testing"

	"nextpitch.com/backend/models"
)

func TestLocationService_UpsertByStreetZip_Insert(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	svc := NewLocationService(db)

	venue := "Springfield HS"
	loc := &models.Location{
		VenueName: &venue,
		Street:    "123 Main St",
		City:      "Springfield",
		State:     "IL",
		Zip:       "62701",
		Country:   "US",
	}

	id, err := svc.UpsertByStreetZip(loc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero id")
	}

	got, err := svc.GetByID(id)
	if err != nil {
		t.Fatalf("GetByID error: %v", err)
	}
	if got.Street != "123 Main St" || got.Zip != "62701" {
		t.Errorf("fields mismatch: %+v", got)
	}
	if got.VenueName == nil || *got.VenueName != "Springfield HS" {
		t.Errorf("venue_name mismatch: %+v", got.VenueName)
	}
}
```

Note: This uses `setupTestDB` / `teardownTestDB` from the existing `backend/services/test_utils.go` (same pattern as other `_test.go` files in the package).

- [ ] **Step 2: Run test — expect compile failure**

Run: `cd backend && go test ./services/ -run TestLocationService_UpsertByStreetZip_Insert`

Expected: FAIL — `undefined: NewLocationService`.

- [ ] **Step 3: Write LocationService**

Create `backend/services/location_service.go`:

```go
package services

import (
	"database/sql"
	"errors"

	"nextpitch.com/backend/models"
)

type LocationService struct {
	db DB
}

func NewLocationService(db DB) *LocationService {
	return &LocationService{db: db}
}

func (s *LocationService) UpsertByStreetZip(loc *models.Location) (int, error) {
	country := loc.Country
	if country == "" {
		country = "US"
	}
	var id int
	err := s.db.QueryRow(`
		INSERT INTO locations (venue_name, street, city, state, zip, country)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (street, zip) DO UPDATE SET
			venue_name = EXCLUDED.venue_name,
			city = EXCLUDED.city,
			state = EXCLUDED.state,
			country = EXCLUDED.country,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id
	`, loc.VenueName, loc.Street, loc.City, loc.State, loc.Zip, country).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *LocationService) GetByID(id int) (*models.Location, error) {
	var loc models.Location
	err := s.db.QueryRow(`
		SELECT id, venue_name, street, city, state, zip, country, created_at, updated_at
		FROM locations
		WHERE id = $1
	`, id).Scan(
		&loc.ID, &loc.VenueName, &loc.Street, &loc.City, &loc.State,
		&loc.Zip, &loc.Country, &loc.CreatedAt, &loc.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("location not found")
	}
	if err != nil {
		return nil, err
	}
	return &loc, nil
}

func (s *LocationService) GetAll() ([]models.Location, error) {
	rows, err := s.db.Query(`
		SELECT id, venue_name, street, city, state, zip, country, created_at, updated_at
		FROM locations
		ORDER BY city ASC, street ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Location
	for rows.Next() {
		var loc models.Location
		err := rows.Scan(
			&loc.ID, &loc.VenueName, &loc.Street, &loc.City, &loc.State,
			&loc.Zip, &loc.Country, &loc.CreatedAt, &loc.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, loc)
	}
	return out, nil
}
```

- [ ] **Step 4: Run test — expect pass**

Run: `cd backend && go test ./services/ -run TestLocationService_UpsertByStreetZip_Insert -v`

Expected: PASS.

- [ ] **Step 5: Write failing test for upsert (conflict path)**

Append to `backend/services/location_service_test.go`:

```go
func TestLocationService_UpsertByStreetZip_UpdatesOnConflict(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	svc := NewLocationService(db)

	first := "First Venue"
	id1, err := svc.UpsertByStreetZip(&models.Location{
		VenueName: &first,
		Street:    "500 Oak Ave",
		City:      "Chicago",
		State:     "IL",
		Zip:       "60601",
	})
	if err != nil {
		t.Fatalf("first upsert error: %v", err)
	}

	second := "Renamed Venue"
	id2, err := svc.UpsertByStreetZip(&models.Location{
		VenueName: &second,
		Street:    "500 Oak Ave",
		City:      "Chicago",
		State:     "IL",
		Zip:       "60601",
	})
	if err != nil {
		t.Fatalf("second upsert error: %v", err)
	}

	if id1 != id2 {
		t.Errorf("expected same id, got %d vs %d", id1, id2)
	}

	loc, err := svc.GetByID(id2)
	if err != nil {
		t.Fatalf("GetByID error: %v", err)
	}
	if loc.VenueName == nil || *loc.VenueName != "Renamed Venue" {
		t.Errorf("expected venue_name updated, got %+v", loc.VenueName)
	}
}

func TestLocationService_GetAll(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)
	svc := NewLocationService(db)

	_, err := svc.UpsertByStreetZip(&models.Location{
		Street: "1 A St", City: "Aville", State: "IL", Zip: "10001",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.UpsertByStreetZip(&models.Location{
		Street: "2 B St", City: "Bville", State: "IL", Zip: "10002",
	})
	if err != nil {
		t.Fatal(err)
	}

	all, err := svc.GetAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) < 2 {
		t.Errorf("expected at least 2 rows, got %d", len(all))
	}
}
```

- [ ] **Step 6: Run tests — expect all pass**

Run: `cd backend && go test ./services/ -run TestLocationService -v`

Expected: all three tests PASS.

- [ ] **Step 7: Commit**

```bash
git add backend/services/location_service.go backend/services/location_service_test.go
git commit -m "Add LocationService with upsert by (street, zip)"
```

---

## Task 4: Wire location into CampService reads

**Files:**
- Modify: `backend/services/camp_service.go`

- [ ] **Step 1: Update scanCamp to include location_id and join location**

Edit `backend/services/camp_service.go`. Replace the `scanCamp` function and each of the four SELECT query strings.

Change `scanCamp` signature and body:

```go
func scanCamp(scanner interface{ Scan(...any) error }) (models.Camp, error) {
	var camp models.Camp
	var priceCents *int
	var loc models.Location
	var locID sql.NullInt64
	var locVenueName sql.NullString
	var locStreet, locCity, locState, locZip, locCountry sql.NullString
	var locCreatedAt, locUpdatedAt sql.NullTime
	err := scanner.Scan(
		&camp.ID,
		&camp.Name,
		&camp.Description,
		&camp.StartDate,
		&camp.EndDate,
		&priceCents,
		&camp.MaxCapacity,
		&camp.Slug,
		&camp.IsActive,
		&camp.LocationID,
		&camp.CreatedAt,
		&camp.UpdatedAt,
		&locID,
		&locVenueName,
		&locStreet,
		&locCity,
		&locState,
		&locZip,
		&locCountry,
		&locCreatedAt,
		&locUpdatedAt,
	)
	camp.Price = centsToDollars(priceCents)
	if locID.Valid {
		loc.ID = int(locID.Int64)
		if locVenueName.Valid {
			s := locVenueName.String
			loc.VenueName = &s
		}
		loc.Street = locStreet.String
		loc.City = locCity.String
		loc.State = locState.String
		loc.Zip = locZip.String
		loc.Country = locCountry.String
		if locCreatedAt.Valid {
			loc.CreatedAt = locCreatedAt.Time
		}
		if locUpdatedAt.Valid {
			loc.UpdatedAt = locUpdatedAt.Time
		}
		camp.Location = &loc
	}
	return camp, err
}
```

Update the four SELECT queries (`GetActiveCamps`, `GetAllCamps`, `GetCampByID`, `GetCampBySlug`) to:

```sql
SELECT c.id, c.name, c.description, c.start_date, c.end_date, c.price_cents,
       c.max_capacity, c.slug, c.is_active, c.location_id, c.created_at, c.updated_at,
       l.id, l.venue_name, l.street, l.city, l.state, l.zip, l.country,
       l.created_at, l.updated_at
FROM camps c
LEFT JOIN locations l ON l.id = c.location_id
WHERE c.is_active = true
ORDER BY c.start_date ASC
```

Adjust the `WHERE` clause per existing method:
- `GetActiveCamps`: `WHERE c.is_active = true ORDER BY c.start_date ASC`
- `GetAllCamps`: `ORDER BY c.start_date ASC` (no WHERE)
- `GetCampByID`: `WHERE c.id = $1`
- `GetCampBySlug`: `WHERE c.slug = $1`

- [ ] **Step 2: Update CreateCamp to persist location_id**

In `backend/services/camp_service.go`, replace the `CreateCamp` INSERT:

```go
err := s.db.QueryRow(`
    INSERT INTO camps (name, description, start_date, end_date, price_cents, max_capacity, slug, is_active, location_id, created_at, updated_at)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10)
    RETURNING id
`,
    camp.Name,
    camp.Description,
    camp.StartDate,
    camp.EndDate,
    dollarsToCents(camp.Price),
    camp.MaxCapacity,
    camp.Slug,
    true,
    camp.LocationID,
    now,
).Scan(&camp.ID)
```

- [ ] **Step 3: Update UpdateCamp to persist location_id**

In `backend/services/camp_service.go`, replace the `UpdateCamp` UPDATE:

```go
result, err := s.db.Exec(`
    UPDATE camps
    SET name = $1, description = $2, start_date = $3, end_date = $4,
        price_cents = $5, max_capacity = $6, slug = $7, location_id = $8, updated_at = $9
    WHERE id = $10
`,
    camp.Name,
    camp.Description,
    camp.StartDate,
    camp.EndDate,
    dollarsToCents(camp.Price),
    camp.MaxCapacity,
    camp.Slug,
    camp.LocationID,
    now,
    camp.ID,
)
```

- [ ] **Step 4: Run existing camp tests — expect pass**

Run: `cd backend && go test ./services/ -v`

Expected: all existing tests still PASS (camps without location have `LocationID = nil`, LEFT JOIN yields NULLs handled by `sql.Null*`).

If any pre-existing test fails due to the new column positions in scanCamp, the test is likely asserting on row scanning directly — rare; update if needed.

- [ ] **Step 5: Commit**

```bash
git add backend/services/camp_service.go
git commit -m "Join location in camp reads, persist location_id on write"
```

---

## Task 5: CampController accepts nested location

**Files:**
- Modify: `backend/controllers/camp_controller.go`
- Modify: `backend/main.go`

- [ ] **Step 1: Add LocationService to CampController**

Edit `backend/controllers/camp_controller.go`. Update struct and constructor:

```go
type CampController struct {
	campService     *services.CampService
	locationService *services.LocationService
	userService     *services.UserService
}

func NewCampController(campService *services.CampService, locationService *services.LocationService, userService *services.UserService) *CampController {
	return &CampController{
		campService:     campService,
		locationService: locationService,
		userService:     userService,
	}
}
```

- [ ] **Step 2: Update createCampRequest to include location**

In `backend/controllers/camp_controller.go`, replace the `createCampRequest` type:

```go
type createCampRequest struct {
	models.Camp
	AgeGroups []models.CampAgeGroup `json:"age_groups"`
	Location  *models.Location      `json:"location" binding:"required"`
}
```

- [ ] **Step 3: Upsert location and set camp FK in CreateCamp**

In `backend/controllers/camp_controller.go`, inside `CreateCamp` after JSON bind validation and before `ctrl.campService.CreateCamp`, add:

```go
	if req.Location == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "location is required"})
		return
	}
	if req.Location.Street == "" || req.Location.City == "" || req.Location.State == "" || req.Location.Zip == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "location street, city, state, and zip are required"})
		return
	}
	locID, err := ctrl.locationService.UpsertByStreetZip(req.Location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save location: " + err.Error()})
		return
	}
	req.Camp.LocationID = &locID
```

Apply the same block inside `UpdateCamp` in the same relative position (after bind, before `campService.UpdateCamp`).

- [ ] **Step 4: Reload camp with joined location for response**

Replace the final JSON responses in `CreateCamp` and `UpdateCamp` with a re-fetch so the nested `location` is populated:

```go
	created, err := ctrl.campService.GetCampByID(req.Camp.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload camp"})
		return
	}
	c.JSON(http.StatusCreated, ctrl.buildCampWithSpots(*created))
```

(For `UpdateCamp`, use `http.StatusOK`.)

- [ ] **Step 5: Update main.go to pass LocationService**

Edit `backend/main.go`. After the line that creates `campService`, add:

```go
	locationService := services.NewLocationService(db.DB)
```

Update the `NewCampController` call:

```go
	campController := controllers.NewCampController(campService, locationService, userService)
```

- [ ] **Step 6: Build backend**

Run: `make build`

Expected: compiles cleanly.

- [ ] **Step 7: Commit**

```bash
git add backend/controllers/camp_controller.go backend/main.go
git commit -m "Upsert location on camp create/update, return nested location"
```

---

## Task 6: Location controller + GET /api/locations route

**Files:**
- Create: `backend/controllers/location_controller.go`
- Modify: `backend/main.go`

- [ ] **Step 1: Write LocationController**

Create `backend/controllers/location_controller.go`:

```go
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"nextpitch.com/backend/services"
)

type LocationController struct {
	locationService *services.LocationService
	userService     *services.UserService
}

func NewLocationController(locationService *services.LocationService, userService *services.UserService) *LocationController {
	return &LocationController{
		locationService: locationService,
		userService:     userService,
	}
}

func (ctrl *LocationController) GetAll(c *gin.Context) {
	userEmail, exists := c.Get("user_email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	isAdmin, err := ctrl.userService.IsAdmin(userEmail.(string))
	if err != nil || !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	locs, err := ctrl.locationService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch locations"})
		return
	}
	if locs == nil {
		locs = []models.Location{}
	}
	c.JSON(http.StatusOK, locs)
}
```

Note: remove the `models.Location{}` fallback line and `models` import if the import is flagged unused — simpler:

```go
	c.JSON(http.StatusOK, locs)
```

and remove the `if locs == nil` block. Gin encodes nil slices as `null`; frontend handles both.

- [ ] **Step 2: Register route**

Edit `backend/main.go`. After `locationService := ...` line, add:

```go
	locationController := controllers.NewLocationController(locationService, userService)
```

Inside the `protected` group (same block that registers `POST /camps`), add:

```go
		protected.GET("/locations", locationController.GetAll)
```

- [ ] **Step 3: Build + test**

Run: `make build && make test-backend`

Expected: compiles, all backend tests PASS.

- [ ] **Step 4: Commit**

```bash
git add backend/controllers/location_controller.go backend/main.go
git commit -m "Add GET /api/locations admin endpoint"
```

---

## Task 7: Backend controller tests for location-required validation

**Files:**
- Modify: `backend/main_test.go` (or the existing controller test file — check which pattern is used)

- [ ] **Step 1: Identify existing test harness**

Run: `ls backend/*_test.go backend/**/*_test.go 2>/dev/null`

If `backend/main_test.go` contains HTTP-level tests using `httptest`, extend it. Otherwise, follow the service-level test pattern and skip this task. (Controllers are thin wrappers; service-level tests cover the logic.)

- [ ] **Step 2: Add test for missing location**

If extending HTTP tests, add:

```go
func TestCreateCamp_RequiresLocation(t *testing.T) {
	router, _ := setupTestRouter(t) // follow existing helper
	body := `{"name":"No Loc","description":"x","start_date":"2026-06-01T00:00:00Z","end_date":"2026-06-02T00:00:00Z","price":50}`
	req := httptest.NewRequest("POST", "/api/camps", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// attach admin auth per existing pattern
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d. body: %s", rec.Code, rec.Body.String())
	}
}
```

If no HTTP test harness exists in this repo, skip this task — service + E2E coverage is sufficient. Note this skip in the commit message of Task 8.

- [ ] **Step 3: Run tests**

Run: `make test-backend`

Expected: PASS.

- [ ] **Step 4: Commit (only if tests added)**

```bash
git add backend/main_test.go
git commit -m "Add controller test for required location on camp create"
```

---

## Task 8: formatAddress utility

**Files:**
- Create: `frontend/src/utils/formatAddress.js`
- Create: `frontend/src/utils/__tests__/formatAddress.test.js` (if co-located pattern used) or similar

- [ ] **Step 1: Write failing test**

Check existing utility test pattern first:

```bash
find frontend/src/utils -name "*.test.*"
```

Create `frontend/src/utils/formatAddress.test.js`:

```js
import { describe, it, expect } from 'vitest';
import { formatAddress } from './formatAddress';

describe('formatAddress', () => {
  it('formats US address without country', () => {
    expect(formatAddress({
      street: '123 Main St',
      city: 'Springfield',
      state: 'IL',
      zip: '62701',
      country: 'US',
    })).toBe('123 Main St, Springfield, IL 62701');
  });

  it('appends country when not US', () => {
    expect(formatAddress({
      street: '1 Queen St',
      city: 'Toronto',
      state: 'ON',
      zip: 'M5H',
      country: 'Canada',
    })).toBe('1 Queen St, Toronto, ON M5H, Canada');
  });

  it('returns empty string when location is null', () => {
    expect(formatAddress(null)).toBe('');
  });
});
```

- [ ] **Step 2: Run test — expect fail**

Run: `cd frontend && npx vitest run src/utils/formatAddress.test.js`

Expected: FAIL — module not found.

- [ ] **Step 3: Write implementation**

Create `frontend/src/utils/formatAddress.js`:

```js
export function formatAddress(loc) {
  if (!loc) return '';
  const base = `${loc.street}, ${loc.city}, ${loc.state} ${loc.zip}`;
  if (loc.country && loc.country !== 'US') {
    return `${base}, ${loc.country}`;
  }
  return base;
}
```

- [ ] **Step 4: Run test — expect pass**

Run: `cd frontend && npx vitest run src/utils/formatAddress.test.js`

Expected: all 3 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/utils/formatAddress.js frontend/src/utils/formatAddress.test.js
git commit -m "Add formatAddress helper"
```

---

## Task 9: LocationMapEmbed component

**Files:**
- Create: `frontend/src/components/LocationMapEmbed.jsx`
- Create: `frontend/src/components/__tests__/LocationMapEmbed.test.jsx`

- [ ] **Step 1: Write failing test**

Create `frontend/src/components/__tests__/LocationMapEmbed.test.jsx`:

```jsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import LocationMapEmbed from '../LocationMapEmbed';

describe('LocationMapEmbed', () => {
  const address = '123 Main St, Springfield, IL 62701';

  it('renders iframe with correct embed src', () => {
    render(<LocationMapEmbed address={address} />);
    const iframe = screen.getByTitle(/Map of 123 Main St/i);
    expect(iframe.tagName).toBe('IFRAME');
    expect(iframe.getAttribute('src')).toBe(
      `https://www.google.com/maps?q=${encodeURIComponent(address)}&output=embed`
    );
  });

  it('renders external Open in Google Maps link', () => {
    render(<LocationMapEmbed address={address} />);
    const link = screen.getByRole('link', { name: /Open in Google Maps/i });
    expect(link.getAttribute('href')).toBe(
      `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(address)}`
    );
    expect(link.getAttribute('target')).toBe('_blank');
    expect(link.getAttribute('rel')).toContain('noopener');
  });
});
```

- [ ] **Step 2: Run test — expect fail**

Run: `cd frontend && npx vitest run src/components/__tests__/LocationMapEmbed.test.jsx`

Expected: FAIL — module not found.

- [ ] **Step 3: Write component**

Create `frontend/src/components/LocationMapEmbed.jsx`:

```jsx
import React from 'react';

const LocationMapEmbed = ({ address }) => {
  if (!address) return null;
  const encoded = encodeURIComponent(address);
  const embedSrc = `https://www.google.com/maps?q=${encoded}&output=embed`;
  const viewHref = `https://www.google.com/maps/search/?api=1&query=${encoded}`;

  return (
    <div className="location-map-embed">
      <iframe
        title={`Map of ${address}`}
        src={embedSrc}
        width="100%"
        height="300"
        style={{ border: 0 }}
        loading="lazy"
        referrerPolicy="no-referrer-when-downgrade"
      />
      <a
        className="location-map-link"
        href={viewHref}
        target="_blank"
        rel="noopener noreferrer"
      >
        Open in Google Maps ↗
      </a>
    </div>
  );
};

export default LocationMapEmbed;
```

- [ ] **Step 4: Run test — expect pass**

Run: `cd frontend && npx vitest run src/components/__tests__/LocationMapEmbed.test.jsx`

Expected: both tests PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/LocationMapEmbed.jsx frontend/src/components/__tests__/LocationMapEmbed.test.jsx
git commit -m "Add LocationMapEmbed component"
```

---

## Task 10: LocationPicker component

**Files:**
- Create: `frontend/src/components/LocationPicker.jsx`
- Create: `frontend/src/components/__tests__/LocationPicker.test.jsx`

- [ ] **Step 1: Write failing tests**

Create `frontend/src/components/__tests__/LocationPicker.test.jsx`:

```jsx
import React, { useState } from 'react';
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import LocationPicker from '../LocationPicker';

const existing = [
  {
    id: 1,
    venue_name: 'Springfield HS',
    street: '123 Main St',
    city: 'Springfield',
    state: 'IL',
    zip: '62701',
    country: 'US',
  },
];

function Harness({ initial = {}, onChange }) {
  const [value, setValue] = useState(initial);
  return (
    <LocationPicker
      value={value}
      onChange={(v) => {
        setValue(v);
        onChange && onChange(v);
      }}
      existingLocations={existing}
    />
  );
}

describe('LocationPicker', () => {
  it('fills fields when selecting an existing location', () => {
    const onChange = vi.fn();
    render(<Harness onChange={onChange} />);
    const select = screen.getByLabelText(/Use existing location/i);
    fireEvent.change(select, { target: { value: '1' } });
    expect(screen.getByLabelText(/Street/i).value).toBe('123 Main St');
    expect(screen.getByLabelText(/City/i).value).toBe('Springfield');
    expect(screen.getByLabelText(/^State/i).value).toBe('IL');
    expect(screen.getByLabelText(/Zip/i).value).toBe('62701');
    expect(onChange).toHaveBeenCalled();
  });

  it('keeps fields editable after selecting', () => {
    render(<Harness />);
    fireEvent.change(screen.getByLabelText(/Use existing location/i), { target: { value: '1' } });
    const streetInput = screen.getByLabelText(/Street/i);
    fireEvent.change(streetInput, { target: { value: '456 Different Rd' } });
    expect(streetInput.value).toBe('456 Different Rd');
  });

  it('shows dedup hint when street+zip match existing', () => {
    render(<Harness />);
    fireEvent.change(screen.getByLabelText(/Street/i), { target: { value: '123 Main St' } });
    fireEvent.change(screen.getByLabelText(/Zip/i), { target: { value: '62701' } });
    expect(screen.getByText(/Matches existing location/i)).toBeInTheDocument();
  });

  it('emits merged object on field change', () => {
    const onChange = vi.fn();
    render(<Harness onChange={onChange} />);
    fireEvent.change(screen.getByLabelText(/Street/i), { target: { value: '9 Elm' } });
    const lastCall = onChange.mock.calls.at(-1)[0];
    expect(lastCall.street).toBe('9 Elm');
  });
});
```

- [ ] **Step 2: Run — expect fail**

Run: `cd frontend && npx vitest run src/components/__tests__/LocationPicker.test.jsx`

Expected: FAIL — module not found.

- [ ] **Step 3: Write component**

Create `frontend/src/components/LocationPicker.jsx`:

```jsx
import React from 'react';

const EMPTY = {
  venue_name: '',
  street: '',
  city: '',
  state: '',
  zip: '',
  country: 'US',
};

const LocationPicker = ({ value, onChange, existingLocations = [] }) => {
  const current = { ...EMPTY, ...(value || {}) };

  const update = (patch) => {
    onChange({ ...current, ...patch });
  };

  const handleSelectExisting = (e) => {
    const id = e.target.value;
    if (!id) return;
    const match = existingLocations.find((l) => String(l.id) === id);
    if (!match) return;
    onChange({
      venue_name: match.venue_name || '',
      street: match.street,
      city: match.city,
      state: match.state,
      zip: match.zip,
      country: match.country || 'US',
    });
  };

  const matchesExisting = existingLocations.some(
    (l) => l.street === current.street && l.zip === current.zip && current.street && current.zip
  );

  return (
    <div className="location-picker">
      <div className="form-group">
        <label htmlFor="location-existing">Use existing location (optional)</label>
        <select id="location-existing" onChange={handleSelectExisting} defaultValue="">
          <option value="">— Enter new address below —</option>
          {existingLocations.map((l) => (
            <option key={l.id} value={l.id}>
              {(l.venue_name || l.street)} — {l.street}, {l.city} {l.state} {l.zip}
            </option>
          ))}
        </select>
        <small>Selecting fills fields below. Edits create new entry unless street+zip match.</small>
      </div>

      <div className="form-group">
        <label htmlFor="location-venue">Venue name (optional)</label>
        <input
          id="location-venue"
          type="text"
          value={current.venue_name || ''}
          onChange={(e) => update({ venue_name: e.target.value })}
        />
      </div>

      <div className="form-group">
        <label htmlFor="location-street">Street</label>
        <input
          id="location-street"
          type="text"
          value={current.street}
          onChange={(e) => update({ street: e.target.value })}
          required
        />
      </div>

      <div className="form-group location-row">
        <div>
          <label htmlFor="location-city">City</label>
          <input
            id="location-city"
            type="text"
            value={current.city}
            onChange={(e) => update({ city: e.target.value })}
            required
          />
        </div>
        <div>
          <label htmlFor="location-state">State</label>
          <input
            id="location-state"
            type="text"
            value={current.state}
            onChange={(e) => update({ state: e.target.value })}
            required
          />
        </div>
        <div>
          <label htmlFor="location-zip">Zip</label>
          <input
            id="location-zip"
            type="text"
            value={current.zip}
            onChange={(e) => update({ zip: e.target.value })}
            required
          />
        </div>
      </div>

      <div className="form-group">
        <label htmlFor="location-country">Country</label>
        <input
          id="location-country"
          type="text"
          value={current.country || 'US'}
          onChange={(e) => update({ country: e.target.value })}
        />
      </div>

      {matchesExisting && (
        <div className="location-dedup-hint">
          ✓ Matches existing location — will reuse on save
        </div>
      )}
    </div>
  );
};

export default LocationPicker;
```

- [ ] **Step 4: Run — expect pass**

Run: `cd frontend && npx vitest run src/components/__tests__/LocationPicker.test.jsx`

Expected: all 4 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/LocationPicker.jsx frontend/src/components/__tests__/LocationPicker.test.jsx
git commit -m "Add LocationPicker component with existing dropdown"
```

---

## Task 11: CampDetailPage

**Files:**
- Create: `frontend/src/pages/CampDetailPage.jsx`
- Create: `frontend/src/pages/CampDetailPage.test.jsx`
- Modify: `frontend/src/App.jsx`

- [ ] **Step 1: Add route to App.jsx**

Edit `frontend/src/App.jsx`. Add import:

```jsx
import CampDetailPage from './pages/CampDetailPage';
```

Add route before `/camps/:slug/register`:

```jsx
<Route path="/camps/:slug" element={<CampDetailPage />} />
```

- [ ] **Step 2: Write failing test**

Create `frontend/src/pages/CampDetailPage.test.jsx`:

```jsx
import React from 'react';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import CampDetailPage from './CampDetailPage';

const mockCamp = {
  id: 1,
  name: 'Summer Pitching Camp',
  description: 'desc',
  slug: 'summer-pitching-camp',
  start_date: '2026-06-10T00:00:00Z',
  end_date: '2026-06-14T00:00:00Z',
  price: 299,
  spots_remaining: 10,
  location: {
    id: 1,
    venue_name: 'Springfield HS',
    street: '123 Main St',
    city: 'Springfield',
    state: 'IL',
    zip: '62701',
    country: 'US',
  },
};

describe('CampDetailPage', () => {
  beforeEach(() => {
    global.fetch = vi.fn(() =>
      Promise.resolve({ ok: true, json: () => Promise.resolve(mockCamp) })
    );
  });

  it('renders camp info and location block', async () => {
    render(
      <MemoryRouter initialEntries={['/camps/summer-pitching-camp']}>
        <Routes>
          <Route path="/camps/:slug" element={<CampDetailPage />} />
        </Routes>
      </MemoryRouter>
    );
    await waitFor(() => expect(screen.getByText('Summer Pitching Camp')).toBeInTheDocument());
    expect(screen.getByText(/Springfield HS/)).toBeInTheDocument();
    expect(screen.getByText(/123 Main St, Springfield, IL 62701/)).toBeInTheDocument();
    const registerLink = screen.getByRole('link', { name: /Register Now/i });
    expect(registerLink.getAttribute('href')).toBe('/camps/summer-pitching-camp/register');
  });

  it('shows Location TBA when location missing', async () => {
    global.fetch = vi.fn(() =>
      Promise.resolve({ ok: true, json: () => Promise.resolve({ ...mockCamp, location: null }) })
    );
    render(
      <MemoryRouter initialEntries={['/camps/summer-pitching-camp']}>
        <Routes>
          <Route path="/camps/:slug" element={<CampDetailPage />} />
        </Routes>
      </MemoryRouter>
    );
    await waitFor(() => expect(screen.getByText(/Location TBA/i)).toBeInTheDocument());
  });
});
```

- [ ] **Step 3: Run — expect fail**

Run: `cd frontend && npx vitest run src/pages/CampDetailPage.test.jsx`

Expected: FAIL — module not found.

- [ ] **Step 4: Write CampDetailPage**

Create `frontend/src/pages/CampDetailPage.jsx`:

```jsx
import React, { useState, useEffect } from 'react';
import { Link, useParams } from 'react-router-dom';
import { getApiUrl } from '../utils/api';
import { formatAddress } from '../utils/formatAddress';
import LocationMapEmbed from '../components/LocationMapEmbed';
import '../styles/camps.css';

const formatDate = (dateStr) =>
  new Date(dateStr).toLocaleDateString('en-US', { month: 'long', day: 'numeric', year: 'numeric' });

const formatPrice = (dollars) => `$${Number(dollars).toFixed(2)}`;

const CampDetailPage = () => {
  const { slug } = useParams();
  const [camp, setCamp] = useState(null);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchCamp = async () => {
      try {
        const res = await fetch(getApiUrl(`camps/by-slug/${slug}`));
        if (!res.ok) throw new Error('not found');
        setCamp(await res.json());
      } catch (e) {
        setError('Camp not found.');
      } finally {
        setLoading(false);
      }
    };
    fetchCamp();
  }, [slug]);

  if (loading) {
    return <div className="container"><div className="section"><p>Loading...</p></div></div>;
  }
  if (error || !camp) {
    return <div className="container"><div className="section"><p className="error">{error || 'Not found'}</p></div></div>;
  }

  const isFull = camp.age_groups && camp.age_groups.length > 0
    ? camp.age_groups.every((g) => g.spots_remaining <= 0)
    : camp.spots_remaining === 0;

  const address = formatAddress(camp.location);

  return (
    <div className="container">
      <div className="section">
        <h1>{camp.name}</h1>
        <p className="duration">
          {formatDate(camp.start_date)} - {formatDate(camp.end_date)}
        </p>
        {camp.age_groups && camp.age_groups.length > 0 ? (
          <div className="age-group-spots">
            {camp.age_groups.map((g, i) => (
              <p key={i} className="camp-spots">
                Ages {g.min_age}-{g.max_age}: {formatPrice(g.price)}
                {' — '}
                {g.spots_remaining > 0
                  ? `${g.spots_remaining} spot${g.spots_remaining !== 1 ? 's' : ''} remaining`
                  : 'Full'}
              </p>
            ))}
          </div>
        ) : (
          <>
            {camp.price && <div className="price">{formatPrice(camp.price)}</div>}
            {camp.spots_remaining !== null && camp.spots_remaining !== undefined && (
              <p className="camp-spots">
                {camp.spots_remaining > 0
                  ? `${camp.spots_remaining} spot${camp.spots_remaining !== 1 ? 's' : ''} remaining`
                  : 'Full'}
              </p>
            )}
          </>
        )}
        <p className="description">{camp.description}</p>
      </div>

      <div className="section">
        <h2>Location</h2>
        {camp.location ? (
          <div className="camp-location">
            {camp.location.venue_name && <p className="venue-name">{camp.location.venue_name}</p>}
            <p className="address">{address}</p>
            <LocationMapEmbed address={address} />
          </div>
        ) : (
          <p>Location TBA</p>
        )}
      </div>

      <div className="section">
        {isFull ? (
          <button className="btn" disabled>Full</button>
        ) : (
          <Link to={`/camps/${camp.slug}/register`} className="btn">Register Now</Link>
        )}
      </div>
    </div>
  );
};

export default CampDetailPage;
```

- [ ] **Step 5: Run — expect pass**

Run: `cd frontend && npx vitest run src/pages/CampDetailPage.test.jsx`

Expected: both tests PASS.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/App.jsx frontend/src/pages/CampDetailPage.jsx frontend/src/pages/CampDetailPage.test.jsx
git commit -m "Add CampDetailPage at /camps/:slug"
```

---

## Task 12: Update CampsPage button to "View Details"

**Files:**
- Modify: `frontend/src/pages/CampsPage.jsx`

- [ ] **Step 1: Check for existing test**

Run: `ls frontend/src/pages/CampsPage.test.jsx 2>/dev/null`

If exists, adapt existing assertions. If not, proceed without a new test (covered by E2E).

- [ ] **Step 2: Change button text and href**

Edit `frontend/src/pages/CampsPage.jsx`. Replace:

```jsx
<Link to={`/camps/${camp.slug}/register`} className="btn">
    Register Now
</Link>
```

with:

```jsx
<Link to={`/camps/${camp.slug}`} className="btn">
    View Details
</Link>
```

- [ ] **Step 3: Run unit tests**

Run: `make test-frontend`

Expected: PASS. If `CampsPage.test.jsx` exists and asserts on "Register Now" text, update assertion to "View Details" and the href to `/camps/:slug`.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/pages/CampsPage.jsx frontend/src/pages/CampsPage.test.jsx
git commit -m "Change camp card button to View Details"
```

---

## Task 13: AdminCampsPage integrates LocationPicker

**Files:**
- Modify: `frontend/src/pages/AdminCampsPage.jsx`

- [ ] **Step 1: Add location state + fetch existing**

Edit `frontend/src/pages/AdminCampsPage.jsx`.

Add import at top:

```jsx
import LocationPicker from '../components/LocationPicker';
```

Add `location` and `locations` to component state (initial form and existing-list):

```jsx
const EMPTY_LOCATION = {
  venue_name: '',
  street: '',
  city: '',
  state: '',
  zip: '',
  country: 'US',
};
```

Extend initial `formData` (inside `useState`):

```jsx
const [formData, setFormData] = useState({
    name: '',
    description: '',
    start_date: '',
    end_date: '',
    price: '',
    max_capacity: '',
    slug: '',
    capacity_mode: 'simple',
    age_groups: [],
    location: { ...EMPTY_LOCATION },
});
```

Add `existingLocations` state:

```jsx
const [existingLocations, setExistingLocations] = useState([]);
```

- [ ] **Step 2: Fetch existing locations on mount**

Inside `useEffect` that runs when `isAdmin`:

```jsx
useEffect(() => {
    if (userLoading) return;
    if (isAdmin) {
        Promise.all([fetchCamps(), fetchLocations()]).finally(() => setLoading(false));
    } else {
        setLoading(false);
    }
}, [isAdmin, userLoading]);
```

Add helper:

```jsx
const fetchLocations = async () => {
    try {
        const token = await getAccessTokenSilently();
        const res = await fetch(getApiUrl('locations'), {
            headers: { Authorization: `Bearer ${token}` },
        });
        if (!res.ok) return;
        const data = await res.json();
        setExistingLocations(data || []);
    } catch (err) {
        // non-fatal — admin can still enter new address
    }
};
```

- [ ] **Step 3: Render LocationPicker in form**

Inside the `<form>` in `AdminCampsPage`, add after the end date field and before the capacity-mode toggle:

```jsx
<div className="form-group">
    <label>Location</label>
    <LocationPicker
        value={formData.location}
        onChange={(location) => setFormData({ ...formData, location })}
        existingLocations={existingLocations}
    />
</div>
```

- [ ] **Step 4: Include location in payload**

In `handleSubmit`, before `const url = editingCamp ? ...`, add validation:

```jsx
const loc = formData.location || {};
if (!loc.street || !loc.city || !loc.state || !loc.zip) {
    setError('Location street, city, state, and zip are required');
    return;
}
```

Extend payload object:

```jsx
const payload = {
    name: formData.name,
    description: formData.description,
    start_date: formData.start_date + 'T00:00:00Z',
    end_date: formData.end_date + 'T00:00:00Z',
    price: formData.capacity_mode === 'simple' ? parseFloat(formData.price) : null,
    slug: formData.slug || null,
    max_capacity: formData.capacity_mode === 'simple' && formData.max_capacity
        ? parseInt(formData.max_capacity)
        : null,
    age_groups: formData.capacity_mode === 'age_range'
        ? formData.age_groups.map(g => ({
            min_age: parseInt(g.min_age),
            max_age: parseInt(g.max_age),
            max_capacity: parseInt(g.max_capacity),
            price: parseFloat(g.price),
        }))
        : [],
    location: {
        venue_name: loc.venue_name || null,
        street: loc.street,
        city: loc.city,
        state: loc.state,
        zip: loc.zip,
        country: loc.country || 'US',
    },
};
```

- [ ] **Step 5: Seed location in handleEdit**

In `handleEdit`, add `location` to `setFormData`:

```jsx
location: camp.location ? {
    venue_name: camp.location.venue_name || '',
    street: camp.location.street,
    city: camp.location.city,
    state: camp.location.state,
    zip: camp.location.zip,
    country: camp.location.country || 'US',
} : { ...EMPTY_LOCATION },
```

- [ ] **Step 6: Reset location in resetForm**

Update `resetForm` to include `location: { ...EMPTY_LOCATION }`:

```jsx
const resetForm = () => {
    setFormData({
        name: '', description: '', start_date: '', end_date: '',
        price: '', max_capacity: '', slug: '',
        capacity_mode: 'simple', age_groups: [],
        location: { ...EMPTY_LOCATION },
    });
    setEditingCamp(null);
    setShowForm(false);
};
```

After `fetchCamps()` call post-save, also refresh locations:

```jsx
resetForm();
await Promise.all([fetchCamps(), fetchLocations()]);
```

- [ ] **Step 7: Run unit tests + build**

Run: `make test-frontend && cd frontend && npm run build`

Expected: tests PASS, build succeeds.

- [ ] **Step 8: Commit**

```bash
git add frontend/src/pages/AdminCampsPage.jsx
git commit -m "Integrate LocationPicker into AdminCampsPage"
```

---

## Task 14: Minimal styling

**Files:**
- Modify: `frontend/styles/camps.css` (append)

- [ ] **Step 1: Add styles for location block**

Append to `frontend/styles/camps.css`:

```css
.location-map-embed {
    margin-top: 1rem;
}

.location-map-embed iframe {
    width: 100%;
    height: 300px;
    border: 0;
    border-radius: 4px;
}

.location-map-link {
    display: inline-block;
    margin-top: 0.5rem;
}

.camp-location .venue-name {
    font-weight: 600;
    margin-bottom: 0.25rem;
}

.camp-location .address {
    color: #555;
    margin-bottom: 0;
}

.location-picker .location-row {
    display: grid;
    grid-template-columns: 2fr 1fr 1fr;
    gap: 0.5rem;
}

.location-dedup-hint {
    background: #e0f2f1;
    color: #00695c;
    padding: 0.4rem 0.6rem;
    border-radius: 3px;
    font-size: 0.85rem;
    margin-top: 0.5rem;
}
```

- [ ] **Step 2: Visual check (manual)**

Run: `make dev:start`

Open `http://localhost:5173/camps` and confirm "View Details" button. Then admin flow (requires logged-in admin): create a camp with location, open `/camps/<slug>`, verify map iframe loads.

Stop with Ctrl-C.

- [ ] **Step 3: Commit**

```bash
git add frontend/styles/camps.css
git commit -m "Style camp location block and picker"
```

---

## Task 15: E2E test updates + new location spec

**Files:**
- Modify: `frontend/tests/e2e/helpers/db.ts`
- Modify: `frontend/tests/e2e/admin-camps.spec.ts`
- Create: `frontend/tests/e2e/camp-location.spec.ts`

- [ ] **Step 1: Extend seed helper to handle locations**

Edit `frontend/tests/e2e/helpers/db.ts`. Extend `seedTestCamp` to optionally accept a `location` and create/find it, OR (simpler) add a separate `seedTestLocation` helper + allow `location_id` override on `seedTestCamp`.

Add at bottom of helpers (before `closeDb`):

```ts
export async function seedTestLocation(overrides: Partial<{
  venue_name: string;
  street: string;
  city: string;
  state: string;
  zip: string;
}> = {}): Promise<{ id: number }> {
  const venue = overrides.venue_name ?? `${TEST_PREFIX} Venue ${Date.now()}`;
  const street = overrides.street ?? `${Date.now()} Test St`;
  const city = overrides.city ?? 'Testville';
  const state = overrides.state ?? 'IL';
  const zip = overrides.zip ?? '60000';

  const res = await db().query(
    `INSERT INTO locations (venue_name, street, city, state, zip)
     VALUES ($1, $2, $3, $4, $5)
     ON CONFLICT (street, zip) DO UPDATE SET venue_name = EXCLUDED.venue_name
     RETURNING id`,
    [venue, street, city, state, zip]
  );
  return { id: res.rows[0].id };
}
```

Extend `seedTestCamp` signature and INSERT:

```ts
export async function seedTestCamp(overrides: Partial<{
  name: string;
  description: string;
  start_date: string;
  end_date: string;
  price_cents: number;
  max_capacity: number | null;
  slug: string;
  location_id: number | null;
}> = {}): Promise<TestCamp> {
  // ... existing default values ...
  const locationId = overrides.location_id ?? null;

  const result = await db().query(
    `INSERT INTO camps (name, description, start_date, end_date, price_cents, max_capacity, slug, is_active, location_id)
     VALUES ($1, $2, $3, $4, $5, $6, $7, true, $8)
     RETURNING id, name, price_cents, slug`,
    [name, description, startDate, endDate, priceCents, maxCapacity, slug, locationId]
  );

  return result.rows[0];
}
```

Update `cleanupTestData` — before the `DELETE FROM camps`, null out FK and delete test-seeded locations:

```ts
export async function cleanupTestData() {
  await db().query(
    `DELETE FROM camp_registrations WHERE camp_id IN (SELECT id FROM camps WHERE name LIKE $1)`,
    [`${TEST_PREFIX}%`]
  );
  await db().query(
    `DELETE FROM camp_age_groups WHERE camp_id IN (SELECT id FROM camps WHERE name LIKE $1)`,
    [`${TEST_PREFIX}%`]
  );
  await db().query(
    `DELETE FROM athletes WHERE name LIKE $1`,
    [`${TEST_PREFIX}%`]
  );
  await db().query(
    `DELETE FROM camps WHERE name LIKE $1`,
    [`${TEST_PREFIX}%`]
  );
  await db().query(
    `DELETE FROM locations WHERE venue_name LIKE $1`,
    [`${TEST_PREFIX}%`]
  );
}
```

Also add a helper to count locations for the dedup assertion:

```ts
export async function countLocations(): Promise<number> {
  const r = await db().query(`SELECT COUNT(*)::int AS c FROM locations`);
  return r.rows[0].c;
}
```

- [ ] **Step 2: Fix existing admin-camps.spec.ts tests**

Edit `frontend/tests/e2e/admin-camps.spec.ts`. For each of the two admin-create tests (`admin creates a camp` and `admin creates a camp with age range capacity`), add location field fills after date fields but before submit:

```ts
await page.locator('#location-street').fill('123 Test St');
await page.locator('#location-city').fill('Testville');
await page.locator('#location-state').fill('IL');
await page.locator('#location-zip').fill('60000');
```

For `admin views registrations`, update the `seedTestCamp` call:

```ts
const loc = await seedTestLocation({
  venue_name: `${'E2E Test'} Venue View`,
  street: `${Date.now()}-view Ln`,
});
const camp = await seedTestCamp({
  name: 'E2E Test Reg View Camp',
  price_cents: 5000,
  location_id: loc.id,
});
```

Import `seedTestLocation` at top.

- [ ] **Step 3: Write new camp-location.spec.ts**

Create `frontend/tests/e2e/camp-location.spec.ts`:

```ts
import { test, expect } from '@playwright/test';
import { authStoragePath } from './helpers/auth';
import { countLocations, seedTestCamp, seedTestLocation } from './helpers/db';

test.describe('camp location admin + public', () => {
  test.use({ storageState: authStoragePath });

  test('admin reuses existing location via dropdown (no duplicate row)', async ({ page }) => {
    // Seed one location the dropdown can pick
    await seedTestLocation({
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

    await page.locator('#location-existing').selectOption({ label: /E2E Test Shared Venue/ });

    await page.getByRole('button', { name: 'Create' }).click();
    await expect(page.locator('.admin-camp-item', { hasText: 'E2E Test Location Reuse Camp' })).toBeVisible({ timeout: 10_000 });

    const after = await countLocations();
    expect(after).toBe(before); // no new location row
  });

  test('public detail page shows address, map iframe, and register link', async ({ page, browser }) => {
    // Use unauthenticated context — public pages should not require auth
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
    const iframe = anonPage.locator('iframe');
    await expect(iframe).toHaveAttribute('src', /google\.com\/maps.*output=embed/);

    const register = anonPage.getByRole('link', { name: /Register Now/i });
    await expect(register).toHaveAttribute('href', `/camps/${camp.slug}/register`);

    await anonymous.close();
  });

  test('list page routes View Details to detail page', async ({ page, browser }) => {
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
```

- [ ] **Step 4: Stop dev server, run E2E**

If dev server running, stop with `make dev:stop`.

Run: `make test-e2e`

Expected: all tests PASS, including the extended `admin-camps.spec.ts` and the new `camp-location.spec.ts`. If any pre-existing E2E fails, confirm it's because of the new required location field — then the fix in Step 2 was incomplete.

- [ ] **Step 5: Commit**

```bash
git add frontend/tests/e2e/helpers/db.ts frontend/tests/e2e/admin-camps.spec.ts frontend/tests/e2e/camp-location.spec.ts
git commit -m "E2E: cover camp location create, reuse, public detail"
```

---

## Task 16: Full test suite + final commit

- [ ] **Step 1: Run full suite**

Run: `make test-all`

Expected: backend + frontend unit + E2E all PASS.

- [ ] **Step 2: Fix any failures inline**

If anything fails, fix it in the minimal way and commit the fix with an explanatory message. Do not skip tests.

- [ ] **Step 3: Credential scan before push**

Run: `/credential-scan` (per CLAUDE.md).

- [ ] **Step 4: Push branch**

```bash
gh auth setup-git && git -c url."https://github.com/".insteadOf="git@github-personal:" push -u origin bug-fixes
```

---

## Self-Review Notes

- **Spec coverage:**
  - Data model (locations table, UNIQUE street+zip, camps.location_id nullable FK) → Task 1, Task 2
  - Upsert logic on server → Task 3
  - Camp service join + writes → Task 4
  - Camp controller validation + upsert + reload → Task 5
  - GET /api/locations → Task 6
  - Frontend helper formatAddress → Task 8
  - LocationMapEmbed → Task 9
  - LocationPicker + dedup hint → Task 10
  - CampDetailPage + route + "Location TBA" fallback → Task 11
  - CampsPage button change → Task 12
  - AdminCampsPage integration → Task 13
  - Styling → Task 14
  - E2E coverage (reuse via dropdown, public detail map, list→detail route) → Task 15
  - `make test-all` gate → Task 16

- **Phase-2 migration (NOT NULL)** is explicitly out of scope per spec; no task covers it.

- **Non-goals honored:** no geocoding, no list-page map, no standalone Locations admin CRUD, no international support beyond country string.
