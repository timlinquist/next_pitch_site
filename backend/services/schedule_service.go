package services

import (
	"database/sql"
	"errors"
	"time"

	"nextpitch.com/backend/models"
)

type DB interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type ScheduleService struct {
	db DB
}

func NewScheduleService(db DB) *ScheduleService {
	return &ScheduleService{db: db}
}

func (s *ScheduleService) GetScheduleEntries() ([]models.ScheduleEntry, error) {
	rows, err := s.db.Query(`
		SELECT se.id, se.title, se.start_time, se.end_time, se.description, u.email as user_email,
		       se.created_at, se.updated_at, se.recurrence
		FROM schedule_entries se
		JOIN users u ON se.user_id = u.id
		ORDER BY se.start_time ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.ScheduleEntry
	for rows.Next() {
		var entry models.ScheduleEntry
		err := rows.Scan(
			&entry.ID,
			&entry.Title,
			&entry.StartTime,
			&entry.EndTime,
			&entry.Description,
			&entry.UserEmail,
			&entry.CreatedAt,
			&entry.UpdatedAt,
			&entry.Recurrence,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (s *ScheduleService) getUserIDOrCreate(email string) (int, error) {
	// First try to get the existing user
	var userID int
	err := s.db.QueryRow(`
		SELECT id FROM users WHERE email = $1
	`, email).Scan(&userID)

	if err == nil {
		// User found, return the ID
		return userID, nil
	}

	if err != sql.ErrNoRows {
		// Unexpected error
		return 0, err
	}

	// User not found, create a new one with default values
	now := time.Now()
	err = s.db.QueryRow(`
		INSERT INTO users (
			email, 
			name,
			is_admin, 
			created_at, 
			updated_at
		)
		VALUES ($1, '', false, $2, $2)
		RETURNING id
	`, email, now).Scan(&userID)

	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (s *ScheduleService) CreateScheduleEntry(entry *models.ScheduleEntry, userEmail string, isAdmin bool) error {
	// Validate event duration for non-admin users
	if !isAdmin {
		duration := entry.EndTime.Sub(entry.StartTime)
		if duration > 2*time.Hour {
			return errors.New("event duration exceeds maximum allowed duration for non-admin users")
		}
	}

	// Check for overlapping events
	overlapping, err := s.checkOverlappingEvents(entry.StartTime, entry.EndTime)
	if err != nil {
		return err
	}
	if overlapping {
		return errors.New("event overlaps with existing events")
	}

	// Get or create user ID from email
	userID, err := s.getUserIDOrCreate(userEmail)
	if err != nil {
		return errors.New("failed to process user")
	}

	now := time.Now()
	err = s.db.QueryRow(`
		INSERT INTO schedule_entries (title, start_time, end_time, description, user_id, created_at, updated_at, recurrence)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, entry.Title, entry.StartTime, entry.EndTime, entry.Description, userID, now, now, entry.Recurrence).Scan(&entry.ID)

	if err != nil {
		return err
	}

	entry.UserEmail = userEmail
	entry.CreatedAt = now
	entry.UpdatedAt = now
	return nil
}

func (s *ScheduleService) checkOverlappingEvents(startTime, endTime time.Time) (bool, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*)
		FROM schedule_entries
		WHERE (start_time < $2 AND end_time > $1) OR  -- Event spans over our new event
			  (start_time >= $1 AND start_time < $2) OR  -- Event starts during our new event
			  (end_time > $1 AND end_time <= $2)  -- Event ends during our new event
	`, startTime, endTime).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *ScheduleService) UpdateScheduleEntry(entry *models.ScheduleEntry, userEmail string, isAdmin bool) error {
	// Validate event duration for non-admin users
	if !isAdmin {
		duration := entry.EndTime.Sub(entry.StartTime)
		if duration > 2*time.Hour {
			return errors.New("event duration exceeds maximum allowed duration for non-admin users")
		}
	}

	// Check for overlapping events, excluding the current event
	overlapping, err := s.checkOverlappingEventsForUpdate(entry.ID, entry.StartTime, entry.EndTime)
	if err != nil {
		return err
	}
	if overlapping {
		return errors.New("event overlaps with existing events")
	}

	// Get or create user ID from email
	userID, err := s.getUserIDOrCreate(userEmail)
	if err != nil {
		return errors.New("failed to process user")
	}

	now := time.Now()
	err = s.db.QueryRow(`
		UPDATE schedule_entries
		SET title = $1, start_time = $2, end_time = $3, description = $4, updated_at = $5, recurrence = $6
		WHERE id = $7 AND user_id = $8
		RETURNING id
	`, entry.Title, entry.StartTime, entry.EndTime, entry.Description, now, entry.Recurrence, entry.ID, userID).Scan(&entry.ID)

	if err == sql.ErrNoRows {
		return errors.New("schedule entry not found or unauthorized")
	}
	if err != nil {
		return err
	}

	entry.UpdatedAt = now
	return nil
}

func (s *ScheduleService) checkOverlappingEventsForUpdate(entryID int, startTime, endTime time.Time) (bool, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*)
		FROM schedule_entries
		WHERE id != $1 AND
			  ((start_time < $3 AND end_time > $2) OR  -- Event spans over our updated event
			   (start_time >= $2 AND start_time < $3) OR  -- Event starts during our updated event
			   (end_time > $2 AND end_time <= $3))  -- Event ends during our updated event
	`, entryID, startTime, endTime).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *ScheduleService) DeleteScheduleEntry(id int64, userEmail string) error {
	// Get or create user ID from email
	userID, err := s.getUserIDOrCreate(userEmail)
	if err != nil {
		return errors.New("failed to process user")
	}

	result, err := s.db.Exec(`
		DELETE FROM schedule_entries
		WHERE id = $1 AND user_id = $2
	`, id, userID)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("schedule entry not found or unauthorized")
	}

	return nil
}

func (s *ScheduleService) GetScheduleEntry(id int64) (*models.ScheduleEntry, error) {
	var entry models.ScheduleEntry
	err := s.db.QueryRow(`
		SELECT se.id, se.title, se.start_time, se.end_time, se.description, u.email as user_email,
		       se.created_at, se.updated_at, se.recurrence
		FROM schedule_entries se
		JOIN users u ON se.user_id = u.id
		WHERE se.id = $1
	`, id).Scan(
		&entry.ID,
		&entry.Title,
		&entry.StartTime,
		&entry.EndTime,
		&entry.Description,
		&entry.UserEmail,
		&entry.CreatedAt,
		&entry.UpdatedAt,
		&entry.Recurrence,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("schedule entry not found")
	}
	if err != nil {
		return nil, err
	}

	return &entry, nil
}

func (s *ScheduleService) GetUpcomingAppointmentsByEmail(email string) ([]models.ScheduleEntry, error) {
	rows, err := s.db.Query(`
		SELECT se.id, se.title, se.description, se.start_time, se.end_time, u.email as user_email,
		       se.created_at, se.updated_at, se.recurrence
		FROM schedule_entries se
		JOIN users u ON se.user_id = u.id
		WHERE u.email = $1 AND se.start_time >= NOW()
		ORDER BY se.start_time ASC
	`, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.ScheduleEntry
	for rows.Next() {
		var entry models.ScheduleEntry
		err := rows.Scan(
			&entry.ID,
			&entry.Title,
			&entry.Description,
			&entry.StartTime,
			&entry.EndTime,
			&entry.UserEmail,
			&entry.CreatedAt,
			&entry.UpdatedAt,
			&entry.Recurrence,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}
