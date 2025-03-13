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
		SELECT id, title, start_time, end_time, description, user_email
		FROM schedule_entries
		ORDER BY start_time ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.ScheduleEntry
	for rows.Next() {
		var entry models.ScheduleEntry
		err := rows.Scan(&entry.ID, &entry.Title, &entry.StartTime, &entry.EndTime, &entry.Description, &entry.UserEmail)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (s *ScheduleService) CreateScheduleEntry(entry *models.ScheduleEntry, userEmail string, isAdmin bool) error {
	// Validate event duration for non-admin users
	if !isAdmin {
		duration := entry.EndTime.Sub(entry.StartTime)
		if duration > 2*time.Hour {
			return errors.New("non-admin users cannot create events longer than 2 hours")
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

	now := time.Now()
	err = s.db.QueryRow(`
		INSERT INTO schedule_entries (title, start_time, end_time, description, user_email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, entry.Title, entry.StartTime, entry.EndTime, entry.Description, userEmail, now, now).Scan(&entry.ID)

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
		WHERE (start_time <= $1 AND end_time > $1) OR
			  (start_time < $2 AND end_time >= $2) OR
			  (start_time >= $1 AND end_time <= $2)
	`, endTime, startTime).Scan(&count)

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
			return errors.New("non-admin users cannot create events longer than 2 hours")
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

	now := time.Now()
	err = s.db.QueryRow(`
		UPDATE schedule_entries
		SET title = $1, start_time = $2, end_time = $3, description = $4, updated_at = $5
		WHERE id = $6 AND user_email = $7
		RETURNING id
	`, entry.Title, entry.StartTime, entry.EndTime, entry.Description, now, entry.ID, userEmail).Scan(&entry.ID)

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
			  ((start_time <= $2 AND end_time > $2) OR
			   (start_time < $3 AND end_time >= $3) OR
			   (start_time >= $2 AND end_time <= $3))
	`, entryID, endTime, startTime).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *ScheduleService) DeleteScheduleEntry(id int64, userEmail string) error {
	result, err := s.db.Exec(`
		DELETE FROM schedule_entries
		WHERE id = $1 AND user_email = $2
	`, id, userEmail)

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
