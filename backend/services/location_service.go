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
