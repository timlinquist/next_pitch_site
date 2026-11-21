package services

import (
	"testing"

	"nextpitch.com/backend/models"
	"nextpitch.com/backend/test/helpers"
)

func TestLocationService_UpsertByStreetZip_Insert(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()
	db.Exec(`TRUNCATE TABLE locations CASCADE`)
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

func TestLocationService_UpsertByStreetZip_UpdatesOnConflict(t *testing.T) {
	db := helpers.SetupTestDB(t)
	defer db.Close()
	db.Exec(`TRUNCATE TABLE locations CASCADE`)
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
	db := helpers.SetupTestDB(t)
	defer db.Close()
	db.Exec(`TRUNCATE TABLE locations CASCADE`)
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
