package services

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"nextpitch.com/backend/models"
)

type testDB struct {
	db *sql.DB
}

func (m *testDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return m.db.QueryRow(query, args...)
}

func (m *testDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return m.db.Query(query, args...)
}

func (m *testDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return m.db.Exec(query, args...)
}

func TestCreateScheduleEntry(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	db := &testDB{db: mockDB}
	service := NewScheduleService(db)

	tests := []struct {
		name          string
		entry         *models.ScheduleEntry
		isAdmin       bool
		setupMock     func()
		expectedError string
	}{
		{
			name: "successful creation by admin",
			entry: &models.ScheduleEntry{
				Title:     "Test Event",
				StartTime: time.Now(),
				EndTime:   time.Now().Add(3 * time.Hour),
				UserEmail: "test@example.com",
			},
			isAdmin: true,
			setupMock: func() {
				mock.ExpectQuery(`
					SELECT COUNT\(\*\)
					FROM schedule_entries
					WHERE \(start_time <= \$1 AND end_time > \$1\) OR
						\(start_time < \$2 AND end_time >= \$2\) OR
						\(start_time >= \$1 AND end_time <= \$2\)
				`).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

				mock.ExpectQuery(`
					INSERT INTO schedule_entries \(title, start_time, end_time, description, user_email, created_at, updated_at\)
					VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7\)
					RETURNING id
				`).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
		},
		{
			name: "non-admin cannot create long event",
			entry: &models.ScheduleEntry{
				Title:     "Test Event",
				StartTime: time.Now(),
				EndTime:   time.Now().Add(3 * time.Hour),
				UserEmail: "test@example.com",
			},
			isAdmin:       false,
			expectedError: "non-admin users cannot create events longer than 2 hours",
		},
		{
			name: "non-admin can create short event",
			entry: &models.ScheduleEntry{
				Title:     "Test Event",
				StartTime: time.Now(),
				EndTime:   time.Now().Add(time.Hour),
				UserEmail: "test@example.com",
			},
			isAdmin: false,
			setupMock: func() {
				mock.ExpectQuery(`
					SELECT COUNT\(\*\)
					FROM schedule_entries
					WHERE \(start_time <= \$1 AND end_time > \$1\) OR
						\(start_time < \$2 AND end_time >= \$2\) OR
						\(start_time >= \$1 AND end_time <= \$2\)
				`).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

				mock.ExpectQuery(`
					INSERT INTO schedule_entries \(title, start_time, end_time, description, user_email, created_at, updated_at\)
					VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7\)
					RETURNING id
				`).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			err := service.CreateScheduleEntry(tt.entry, tt.entry.UserEmail, tt.isAdmin)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUpdateScheduleEntry(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	db := &testDB{db: mockDB}
	service := NewScheduleService(db)

	tests := []struct {
		name          string
		entry         *models.ScheduleEntry
		isAdmin       bool
		setupMock     func()
		expectedError string
	}{
		{
			name: "successful update by admin",
			entry: &models.ScheduleEntry{
				ID:        1,
				Title:     "Updated Event",
				StartTime: time.Now(),
				EndTime:   time.Now().Add(3 * time.Hour),
				UserEmail: "test@example.com",
			},
			isAdmin: true,
			setupMock: func() {
				mock.ExpectQuery(`
					SELECT COUNT\(\*\)
					FROM schedule_entries
					WHERE id != \$1 AND
						\(\(start_time <= \$2 AND end_time > \$2\) OR
						\(start_time < \$3 AND end_time >= \$3\) OR
						\(start_time >= \$2 AND end_time <= \$3\)\)
				`).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

				mock.ExpectQuery(`
					UPDATE schedule_entries
					SET title = \$1, start_time = \$2, end_time = \$3, description = \$4, updated_at = \$5
					WHERE id = \$6 AND user_email = \$7
					RETURNING id
				`).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
		},
		{
			name: "non-admin cannot update to long event",
			entry: &models.ScheduleEntry{
				ID:        1,
				Title:     "Updated Event",
				StartTime: time.Now(),
				EndTime:   time.Now().Add(3 * time.Hour),
				UserEmail: "test@example.com",
			},
			isAdmin:       false,
			expectedError: "non-admin users cannot create events longer than 2 hours",
		},
		{
			name: "non-admin can update to short event",
			entry: &models.ScheduleEntry{
				ID:        1,
				Title:     "Updated Event",
				StartTime: time.Now(),
				EndTime:   time.Now().Add(time.Hour),
				UserEmail: "test@example.com",
			},
			isAdmin: false,
			setupMock: func() {
				mock.ExpectQuery(`
					SELECT COUNT\(\*\)
					FROM schedule_entries
					WHERE id != \$1 AND
						\(\(start_time <= \$2 AND end_time > \$2\) OR
						\(start_time < \$3 AND end_time >= \$3\) OR
						\(start_time >= \$2 AND end_time <= \$3\)\)
				`).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

				mock.ExpectQuery(`
					UPDATE schedule_entries
					SET title = \$1, start_time = \$2, end_time = \$3, description = \$4, updated_at = \$5
					WHERE id = \$6 AND user_email = \$7
					RETURNING id
				`).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			err := service.UpdateScheduleEntry(tt.entry, tt.entry.UserEmail, tt.isAdmin)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCheckOverlappingEvents(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()

	db := &testDB{db: mockDB}
	service := NewScheduleService(db)

	tests := []struct {
		name          string
		startTime     time.Time
		endTime       time.Time
		setupMock     func()
		expectedError string
	}{
		{
			name:      "no overlapping events",
			startTime: time.Now(),
			endTime:   time.Now().Add(time.Hour),
			setupMock: func() {
				mock.ExpectQuery(`
					SELECT COUNT\(\*\)
					FROM schedule_entries
					WHERE \(start_time <= \$1 AND end_time > \$1\) OR
						\(start_time < \$2 AND end_time >= \$2\) OR
						\(start_time >= \$1 AND end_time <= \$2\)
				`).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
		},
		{
			name:      "overlapping events",
			startTime: time.Now(),
			endTime:   time.Now().Add(time.Hour),
			setupMock: func() {
				mock.ExpectQuery(`
					SELECT COUNT\(\*\)
					FROM schedule_entries
					WHERE \(start_time <= \$1 AND end_time > \$1\) OR
						\(start_time < \$2 AND end_time >= \$2\) OR
						\(start_time >= \$1 AND end_time <= \$2\)
				`).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			expectedError: "event overlaps with existing events",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			overlapping, err := service.checkOverlappingEvents(tt.startTime, tt.endTime)
			assert.NoError(t, err)

			if tt.expectedError != "" {
				assert.True(t, overlapping, "Expected overlapping events")
			} else {
				assert.False(t, overlapping, "Expected no overlapping events")
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
