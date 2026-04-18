package services

import (
	"database/sql"
	"errors"
	"regexp"
	"sort"
	"strings"
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
		       max_capacity, slug, is_active, created_at, updated_at
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
			&camp.Slug,
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
		       max_capacity, slug, is_active, created_at, updated_at
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
			&camp.Slug,
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
		       max_capacity, slug, is_active, created_at, updated_at
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
		&camp.Slug,
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

func (s *CampService) GetCampBySlug(slug string) (*models.Camp, error) {
	var camp models.Camp
	err := s.db.QueryRow(`
		SELECT id, name, description, start_date, end_date, price_cents,
		       max_capacity, slug, is_active, created_at, updated_at
		FROM camps
		WHERE slug = $1
	`, slug).Scan(
		&camp.ID,
		&camp.Name,
		&camp.Description,
		&camp.StartDate,
		&camp.EndDate,
		&camp.PriceCents,
		&camp.MaxCapacity,
		&camp.Slug,
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

var slugRegexp = regexp.MustCompile(`[^a-z0-9]+`)

func (s *CampService) GenerateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = slugRegexp.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}

func (s *CampService) CreateCamp(camp *models.Camp) error {
	if camp.Slug == nil || *camp.Slug == "" {
		generated := s.GenerateSlug(camp.Name)
		camp.Slug = &generated
	}

	now := time.Now()
	err := s.db.QueryRow(`
		INSERT INTO camps (name, description, start_date, end_date, price_cents, max_capacity, slug, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
		RETURNING id
	`,
		camp.Name,
		camp.Description,
		camp.StartDate,
		camp.EndDate,
		camp.PriceCents,
		camp.MaxCapacity,
		camp.Slug,
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
		    price_cents = $5, max_capacity = $6, slug = $7, updated_at = $8
		WHERE id = $9
	`,
		camp.Name,
		camp.Description,
		camp.StartDate,
		camp.EndDate,
		camp.PriceCents,
		camp.MaxCapacity,
		camp.Slug,
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

// Age group methods

func (s *CampService) GetAgeGroupsByCampID(campID int) ([]models.CampAgeGroup, error) {
	rows, err := s.db.Query(`
		SELECT id, camp_id, min_age, max_age, max_capacity, price_cents, created_at, updated_at
		FROM camp_age_groups
		WHERE camp_id = $1
		ORDER BY min_age ASC
	`, campID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.CampAgeGroup
	for rows.Next() {
		var g models.CampAgeGroup
		err := rows.Scan(&g.ID, &g.CampID, &g.MinAge, &g.MaxAge, &g.MaxCapacity, &g.PriceCents, &g.CreatedAt, &g.UpdatedAt)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}

	return groups, nil
}

func (s *CampService) SetAgeGroups(campID int, groups []models.CampAgeGroup) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM camp_age_groups WHERE camp_id = $1`, campID)
	if err != nil {
		return err
	}

	for _, g := range groups {
		_, err = tx.Exec(`
			INSERT INTO camp_age_groups (camp_id, min_age, max_age, max_capacity, price_cents)
			VALUES ($1, $2, $3, $4, $5)
		`, campID, g.MinAge, g.MaxAge, g.MaxCapacity, g.PriceCents)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *CampService) ValidateAgeGroups(groups []models.CampAgeGroup) error {
	if len(groups) == 0 {
		return nil
	}

	sorted := make([]models.CampAgeGroup, len(groups))
	copy(sorted, groups)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].MinAge < sorted[j].MinAge
	})

	for i := 1; i < len(sorted); i++ {
		if sorted[i].MinAge <= sorted[i-1].MaxAge {
			return errors.New("age groups must not overlap")
		}
	}

	return nil
}

func (s *CampService) GetAgeGroupRegistrationCount(campID, minAge, maxAge int) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*)
		FROM camp_registrations cr
		JOIN athletes a ON a.id = cr.athlete_id
		WHERE cr.camp_id = $1
		  AND cr.payment_status IN ('pending', 'paid')
		  AND a.age BETWEEN $2 AND $3
	`, campID, minAge, maxAge).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
