package controllers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"nextpitch.com/backend/db"
	"nextpitch.com/backend/models"
)

type ScheduleController struct {
	db *sql.DB
}

func NewScheduleController() *ScheduleController {
	return &ScheduleController{
		db: db.DB,
	}
}

func (sc *ScheduleController) GetScheduleEntries(c *gin.Context) {
	rows, err := sc.db.Query(`
		SELECT se.id, se.title, se.description, se.start_time, se.end_time, 
			   se.created_at, se.updated_at, u.email
		FROM schedule_entries se
		LEFT JOIN users u ON se.user_id = u.id
		ORDER BY se.start_time ASC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch schedule entries"})
		return
	}
	defer rows.Close()

	var entries []models.ScheduleEntry
	for rows.Next() {
		var entry models.ScheduleEntry
		var email sql.NullString
		err := rows.Scan(
			&entry.ID, &entry.Title, &entry.Description,
			&entry.StartTime, &entry.EndTime,
			&entry.CreatedAt, &entry.UpdatedAt,
			&email,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan schedule entry"})
			return
		}
		if email.Valid {
			entry.UserEmail = email.String
		}
		entries = append(entries, entry)
	}

	c.JSON(http.StatusOK, entries)
}

func (sc *ScheduleController) CreateScheduleEntry(c *gin.Context) {
	var entry models.ScheduleEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get or create user by email
	var userID int
	err := sc.db.QueryRow(`
		WITH new_user AS (
			INSERT INTO users (email, first_name, last_name)
			VALUES ($1, '', '') -- Placeholder names, can be updated later
			ON CONFLICT (email) DO NOTHING
			RETURNING id
		)
		SELECT id FROM new_user
		UNION ALL
		SELECT id FROM users WHERE email = $1
		LIMIT 1
	`, entry.UserEmail).Scan(&userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process user"})
		return
	}

	err = sc.db.QueryRow(`
		INSERT INTO schedule_entries (title, description, start_time, end_time, user_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`, entry.Title, entry.Description, entry.StartTime, entry.EndTime, userID).
		Scan(&entry.ID, &entry.CreatedAt, &entry.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create schedule entry"})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

func (sc *ScheduleController) UpdateScheduleEntry(c *gin.Context) {
	id := c.Param("id")
	var entry models.ScheduleEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := sc.db.QueryRow(`
		UPDATE schedule_entries 
		SET title = $1, description = $2, start_time = $3, end_time = $4
		WHERE id = $5
		RETURNING id, created_at, updated_at
	`, entry.Title, entry.Description, entry.StartTime, entry.EndTime, id).
		Scan(&entry.ID, &entry.CreatedAt, &entry.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Schedule entry not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update schedule entry"})
		return
	}

	c.JSON(http.StatusOK, entry)
}

func (sc *ScheduleController) DeleteScheduleEntry(c *gin.Context) {
	id := c.Param("id")
	result, err := sc.db.Exec("DELETE FROM schedule_entries WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete schedule entry"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rows affected"})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Schedule entry not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Schedule entry deleted successfully"})
}

func (sc *ScheduleController) GetUpcomingAppointmentsByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	rows, err := sc.db.Query(`
		SELECT se.id, se.title, se.description, se.start_time, se.end_time, 
			   se.created_at, se.updated_at 
		FROM schedule_entries se
		JOIN users u ON se.user_id = u.id
		WHERE u.email = $1 AND se.start_time >= NOW()
		ORDER BY se.start_time ASC
	`, email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch appointments"})
		return
	}
	defer rows.Close()

	var entries []models.ScheduleEntry
	for rows.Next() {
		var entry models.ScheduleEntry
		err := rows.Scan(
			&entry.ID, &entry.Title, &entry.Description,
			&entry.StartTime, &entry.EndTime,
			&entry.CreatedAt, &entry.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan appointment"})
			return
		}
		entries = append(entries, entry)
	}

	c.JSON(http.StatusOK, entries)
}
