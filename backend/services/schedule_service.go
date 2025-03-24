package services

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"nextpitch.com/backend/models"
)

type DB interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	Begin() (*sql.Tx, error)
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
		       se.created_at, se.updated_at, se.recurrence, se.parent_event_id
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
			&entry.ParentEventID,
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

func (s *ScheduleService) generateRecurringInstances(parent *models.ScheduleEntry) ([]models.ScheduleEntry, error) {
	if parent.Recurrence == models.RecurrenceNone || parent.RecurrenceEndDate == nil {
		return nil, nil
	}

	var instances []models.ScheduleEntry
	duration := parent.EndTime.Sub(parent.StartTime)

	// Calculate the number of instances based on recurrence type and end date
	var numInstances int
	switch parent.Recurrence {
	case models.RecurrenceWeekly:
		weeks := int(parent.RecurrenceEndDate.Sub(parent.StartTime).Hours() / (24 * 7))
		if weeks < 0 {
			return nil, nil
		}
		numInstances = weeks
	case models.RecurrenceBiweekly:
		weeks := int(parent.RecurrenceEndDate.Sub(parent.StartTime).Hours() / (24 * 14))
		if weeks < 0 {
			return nil, nil
		}
		numInstances = weeks
	case models.RecurrenceMonthly:
		// For monthly recurrence, we need to count the number of months
		months := (parent.RecurrenceEndDate.Year()-parent.StartTime.Year())*12 +
			int(parent.RecurrenceEndDate.Month()-parent.StartTime.Month())
		if months < 0 {
			return nil, nil
		}
		numInstances = months
	}

	// Generate instances
	for i := 1; i <= numInstances; i++ { // Start from 1 to skip parent event
		instance := *parent
		instance.ID = 0                             // Reset ID for new instance
		instance.Recurrence = models.RecurrenceNone // Instances don't recur
		instance.RecurrenceEndDate = nil
		instance.ParentEventID = &parent.ID

		// Calculate instance dates
		switch parent.Recurrence {
		case models.RecurrenceWeekly:
			instance.StartTime = parent.StartTime.AddDate(0, 0, i*7)
		case models.RecurrenceBiweekly:
			instance.StartTime = parent.StartTime.AddDate(0, 0, i*14)
		case models.RecurrenceMonthly:
			instance.StartTime = parent.StartTime.AddDate(0, i, 0)
		}

		// Only add the instance if it's before or at the end date
		if !instance.StartTime.After(*parent.RecurrenceEndDate) {
			instance.EndTime = instance.StartTime.Add(duration)
			instances = append(instances, instance)
		}
	}

	return instances, nil
}

type FailedInstance struct {
	Event  models.ScheduleEntry
	Reason string
}

type BulkInsertResult struct {
	SuccessfullyInserted int
	FailedInstances      []FailedInstance
}

func (s *ScheduleService) bulkInsertEvents(events []models.ScheduleEntry) (*BulkInsertResult, error) {
	// Start a transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	result := &BulkInsertResult{
		SuccessfullyInserted: 0,
		FailedInstances:      make([]FailedInstance, 0),
	}

	now := time.Now()
	for _, event := range events {
		// Check for overlapping events
		overlapping, err := s.checkOverlappingEvents(event.StartTime, event.EndTime)
		if err != nil {
			result.FailedInstances = append(result.FailedInstances, FailedInstance{
				Event:  event,
				Reason: err.Error(),
			})
			continue
		}

		if overlapping {
			result.FailedInstances = append(result.FailedInstances, FailedInstance{
				Event:  event,
				Reason: "overlap",
			})
			continue
		}

		// Get or create user ID from email
		userID, err := s.getUserIDOrCreate(event.UserEmail)
		if err != nil {
			result.FailedInstances = append(result.FailedInstances, FailedInstance{
				Event:  event,
				Reason: fmt.Sprintf("Failed to get user ID: %v", err),
			})
			continue
		}

		// Set default recurrence type if not set
		recurrence := event.Recurrence
		if recurrence == "" {
			recurrence = models.RecurrenceNone
		}

		// Insert the event
		_, err = tx.Exec(`
			INSERT INTO schedule_entries (
				title, start_time, end_time, description, user_id,
				created_at, updated_at, recurrence, recurrence_end_date,
				parent_event_id
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, event.Title, event.StartTime, event.EndTime, event.Description, userID,
			now, now, recurrence, event.RecurrenceEndDate, event.ParentEventID)

		if err != nil {
			result.FailedInstances = append(result.FailedInstances, FailedInstance{
				Event:  event,
				Reason: fmt.Sprintf("Failed to insert event: %v", err),
			})
			continue
		}

		result.SuccessfullyInserted++
	}

	// If we have any successful inserts, commit the transaction
	if result.SuccessfullyInserted > 0 {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	return result, nil
}

func (s *ScheduleService) checkOverlappingEvents(startTime, endTime time.Time, parentID ...int) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM schedule_entries
		WHERE (
			(start_time <= $1 AND end_time > $1) OR  -- Event starts before or at our start and ends after our start
			(start_time < $2 AND end_time >= $2) OR  -- Event starts before our end and ends at or after our end
			(start_time >= $1 AND end_time <= $2)    -- Event is completely contained within our time range
		) AND (parent_event_id IS NULL OR parent_event_id = id)  -- Only check against parent events or non-recurring events
	`

	// If parentID is provided, exclude it from the overlap check
	if len(parentID) > 0 {
		query += " AND id != $3"
		err := s.db.QueryRow(query, startTime, endTime, parentID[0]).Scan(&count)
		if err != nil {
			return false, fmt.Errorf("failed to check overlapping events: %w", err)
		}
	} else {
		err := s.db.QueryRow(query, startTime, endTime).Scan(&count)
		if err != nil {
			return false, fmt.Errorf("failed to check overlapping events: %w", err)
		}
	}

	return count > 0, nil
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

	// Start a transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()

	// Set default recurrence type if not set
	if entry.Recurrence == "" {
		entry.Recurrence = models.RecurrenceNone
	}

	// Insert the parent event
	err = tx.QueryRow(`
		INSERT INTO schedule_entries (
			title, start_time, end_time, description, user_id,
			created_at, updated_at, recurrence, recurrence_end_date
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`, entry.Title, entry.StartTime, entry.EndTime, entry.Description, userID,
		now, now, entry.Recurrence, entry.RecurrenceEndDate).Scan(&entry.ID)

	if err != nil {
		return fmt.Errorf("failed to insert parent event: %w", err)
	}

	// If this is a recurring event, generate and insert instances
	if entry.Recurrence != models.RecurrenceNone && entry.RecurrenceEndDate != nil {
		instances, err := s.generateRecurringInstances(entry)
		if err != nil {
			return fmt.Errorf("failed to generate recurring instances: %w", err)
		}

		// Insert each instance within the same transaction
		for _, instance := range instances {
			// Check for overlapping events for each instance, excluding the parent event
			overlapping, err := s.checkOverlappingEvents(instance.StartTime, instance.EndTime, entry.ID)
			if err != nil {
				return fmt.Errorf("failed to check overlapping events for instance: %w", err)
			}
			if overlapping {
				continue // Skip this instance if it overlaps
			}

			_, err = tx.Exec(`
				INSERT INTO schedule_entries (
					title, start_time, end_time, description, user_id,
					created_at, updated_at, recurrence, recurrence_end_date,
					parent_event_id
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			`, instance.Title, instance.StartTime, instance.EndTime, instance.Description, userID,
				now, now, instance.Recurrence, instance.RecurrenceEndDate, instance.ParentEventID)

			if err != nil {
				return fmt.Errorf("failed to insert instance: %w", err)
			}
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	entry.UserEmail = userEmail
	entry.CreatedAt = now
	entry.UpdatedAt = now
	return nil
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

func (s *ScheduleService) DeleteScheduleEntry(id int64, userEmail string, deleteFollowing bool) error {
	// Get or create user ID from email
	userID, err := s.getUserIDOrCreate(userEmail)
	if err != nil {
		return errors.New("failed to process user")
	}

	// Start a transaction
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get the event details and check if it's a parent event
	var parentEventID *int64
	var startTime time.Time
	var hasChildren bool
	var recurrence models.RecurrenceType
	err = tx.QueryRow(`
		SELECT se.parent_event_id, se.start_time, se.recurrence,
			   EXISTS (SELECT 1 FROM schedule_entries WHERE parent_event_id = se.id) as has_children
		FROM schedule_entries se
		WHERE se.id = $1 AND se.user_id = $2
	`, id, userID).Scan(&parentEventID, &startTime, &recurrence, &hasChildren)

	if err == sql.ErrNoRows {
		return errors.New("schedule entry not found or unauthorized")
	}
	if err != nil {
		return err
	}

	// If this is a parent event (has children)
	if hasChildren {
		if deleteFollowing {
			// Delete all child events first
			_, err = tx.Exec(`
				DELETE FROM schedule_entries
				WHERE parent_event_id = $1
			`, id)
			if err != nil {
				return err
			}
		} else {
			// Find the next event in the series to be the new parent
			var newParentID int64
			err = tx.QueryRow(`
				SELECT id
				FROM schedule_entries
				WHERE parent_event_id = $1
				AND start_time > $2
				ORDER BY start_time ASC
				LIMIT 1
			`, id, startTime).Scan(&newParentID)

			if err != nil && err != sql.ErrNoRows {
				return fmt.Errorf("failed to find new parent event: %w", err)
			}

			if err == sql.ErrNoRows {
				// No next event found, set all remaining events to have no parent
				_, err = tx.Exec(`
					UPDATE schedule_entries
					SET parent_event_id = NULL
					WHERE parent_event_id = $1
				`, id)
				if err != nil {
					return fmt.Errorf("failed to update remaining events: %w", err)
				}
			} else {
				// Update all subsequent events to point to the new parent
				_, err = tx.Exec(`
					UPDATE schedule_entries
					SET parent_event_id = $1
					WHERE parent_event_id = $2
					AND start_time > $3
				`, newParentID, id, startTime)
				if err != nil {
					return fmt.Errorf("failed to update parent references: %w", err)
				}

				// Update the new parent event to have parent_event_id = NULL and inherit the recurrence type
				_, err = tx.Exec(`
					UPDATE schedule_entries
					SET parent_event_id = NULL, recurrence = $2
					WHERE id = $1
				`, newParentID, recurrence)
				if err != nil {
					return fmt.Errorf("failed to update new parent event: %w", err)
				}
			}
		}
	}

	// If this is a child event and deleteFollowing is true, delete all following events
	if deleteFollowing && parentEventID != nil {
		// Delete all events with the same parent_id that start after this event
		_, err = tx.Exec(`
			DELETE FROM schedule_entries
			WHERE parent_event_id = $1 AND start_time > $2
		`, parentEventID, startTime)
		if err != nil {
			return err
		}
	}

	// Delete the event itself
	result, err := tx.Exec(`
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

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return err
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
