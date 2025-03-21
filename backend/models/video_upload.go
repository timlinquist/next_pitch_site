package models

import (
	"database/sql"
	"time"
)

// VideoUploadStatus represents the status of a video upload
type VideoUploadStatus string

const (
	VideoUploadStatusUploaded  VideoUploadStatus = "uploaded"
	VideoUploadStatusProcessed VideoUploadStatus = "processed"
	VideoUploadStatusPaid      VideoUploadStatus = "paid"
)

// VideoUpload represents a video upload record in the database
type VideoUpload struct {
	ID         int               `json:"id"`
	UserID     int               `json:"user_id"`
	DropboxURL string            `json:"dropbox_url"`
	FileName   string            `json:"file_name"`
	Status     VideoUploadStatus `json:"status"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// CreateVideoUpload creates a new video upload record
func CreateVideoUpload(db *sql.DB, upload *VideoUpload) error {
	query := `
		INSERT INTO video_uploads (user_id, dropbox_url, file_name, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	return db.QueryRow(
		query,
		upload.UserID,
		upload.DropboxURL,
		upload.FileName,
		upload.Status,
	).Scan(&upload.ID, &upload.CreatedAt, &upload.UpdatedAt)
}

// GetVideoUploadByID retrieves a video upload by its ID
func GetVideoUploadByID(db *sql.DB, id int) (*VideoUpload, error) {
	upload := &VideoUpload{}
	query := `
		SELECT id, user_id, dropbox_url, file_name, status, created_at, updated_at
		FROM video_uploads
		WHERE id = $1
	`
	err := db.QueryRow(query, id).Scan(
		&upload.ID,
		&upload.UserID,
		&upload.DropboxURL,
		&upload.FileName,
		&upload.Status,
		&upload.CreatedAt,
		&upload.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return upload, nil
}

// GetVideoUploadsByUserID retrieves all video uploads for a specific user
func GetVideoUploadsByUserID(db *sql.DB, userID int) ([]*VideoUpload, error) {
	query := `
		SELECT id, user_id, dropbox_url, file_name, status, created_at, updated_at
		FROM video_uploads
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var uploads []*VideoUpload
	for rows.Next() {
		upload := &VideoUpload{}
		err := rows.Scan(
			&upload.ID,
			&upload.UserID,
			&upload.DropboxURL,
			&upload.FileName,
			&upload.Status,
			&upload.CreatedAt,
			&upload.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		uploads = append(uploads, upload)
	}
	return uploads, nil
}

// UpdateVideoUploadStatus updates the status of a video upload
func UpdateVideoUploadStatus(db *sql.DB, id int, status VideoUploadStatus) error {
	query := `
		UPDATE video_uploads
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	result, err := db.Exec(query, status, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
