package services

import (
	"database/sql"
	"errors"
	"time"

	"nextpitch.com/backend/models"
)

type CampService struct {
	db DB
}

func NewCampService(db DB) *CampService {
	return &CampService{db: db}
}

func (s *CampService) GetActiveCamps() ([]models.Camp, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, start_date, end_date, price_cents,
		       max_capacity, is_active, created_at, updated_at
		FROM camps
		WHERE is_active = true
		ORDER BY start_date ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var camps []models.Camp
	for rows.Next() {
		var camp models.Camp
		err := rows.Scan(
			&camp.ID,
			&camp.Name,
			&camp.Description,
			&camp.StartDate,
			&camp.EndDate,
			&camp.PriceCents,
			&camp.MaxCapacity,
			&camp.IsActive,
			&camp.CreatedAt,
			&camp.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		camps = append(camps, camp)
	}

	return camps, nil
}

func (s *CampService) GetAllCamps() ([]models.Camp, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, start_date, end_date, price_cents,
		       max_capacity, is_active, created_at, updated_at
		FROM camps
		ORDER BY start_date ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var camps []models.Camp
	for rows.Next() {
		var camp models.Camp
		err := rows.Scan(
			&camp.ID,
			&camp.Name,
			&camp.Description,
			&camp.StartDate,
			&camp.EndDate,
			&camp.PriceCents,
			&camp.MaxCapacity,
			&camp.IsActive,
			&camp.CreatedAt,
			&camp.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		camps = append(camps, camp)
	}

	return camps, nil
}

func (s *CampService) GetCampByID(id int) (*models.Camp, error) {
	var camp models.Camp
	err := s.db.QueryRow(`
		SELECT id, name, description, start_date, end_date, price_cents,
		       max_capacity, is_active, created_at, updated_at
		FROM camps
		WHERE id = $1
	`, id).Scan(
		&camp.ID,
		&camp.Name,
		&camp.Description,
		&camp.StartDate,
		&camp.EndDate,
		&camp.PriceCents,
		&camp.MaxCapacity,
		&camp.IsActive,
		&camp.CreatedAt,
		&camp.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("camp not found")
	}
	if err != nil {
		return nil, err
	}

	return &camp, nil
}

func (s *CampService) CreateCamp(camp *models.Camp) error {
	now := time.Now()
	err := s.db.QueryRow(`
		INSERT INTO camps (name, description, start_date, end_date, price_cents, max_capacity, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		RETURNING id
	`,
		camp.Name,
		camp.Description,
		camp.StartDate,
		camp.EndDate,
		camp.PriceCents,
		camp.MaxCapacity,
		true,
		now,
	).Scan(&camp.ID)

	if err != nil {
		return err
	}

	camp.IsActive = true
	camp.CreatedAt = now
	camp.UpdatedAt = now
	return nil
}

func (s *CampService) UpdateCamp(camp *models.Camp) error {
	now := time.Now()
	result, err := s.db.Exec(`
		UPDATE camps
		SET name = $1, description = $2, start_date = $3, end_date = $4,
		    price_cents = $5, max_capacity = $6, updated_at = $7
		WHERE id = $8
	`,
		camp.Name,
		camp.Description,
		camp.StartDate,
		camp.EndDate,
		camp.PriceCents,
		camp.MaxCapacity,
		now,
		camp.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("camp not found")
	}

	camp.UpdatedAt = now
	return nil
}

func (s *CampService) DeactivateCamp(id int) error {
	result, err := s.db.Exec(`
		UPDATE camps SET is_active = false, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("camp not found")
	}

	return nil
}

func (s *CampService) GetCampRegistrationCount(campID int) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*)
		FROM camp_registrations
		WHERE camp_id = $1 AND payment_status IN ('pending', 'paid')
	`, campID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
