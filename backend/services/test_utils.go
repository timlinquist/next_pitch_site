package services

import (
	"database/sql"
	"database/sql/driver"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/mock"
)

// MockRow is a mock implementation of sql.Row
type MockRow struct {
	mock.Mock
}

func (m *MockRow) Scan(dest ...interface{}) error {
	args := m.Called(dest...)
	return args.Error(0)
}

// MockDB is a mock implementation of the database interface
type MockDB struct {
	mock.Mock
	db      *sql.DB
	sqlmock sqlmock.Sqlmock
}

func NewMockDB() (*MockDB, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	return &MockDB{
		db:      db,
		sqlmock: mock,
	}, nil
}

func (m *MockDB) QueryRow(query string, args ...interface{}) *sql.Row {
	// Get the mock result based on the query and args
	result := m.Called(query, args)

	// If a *sql.Row was directly returned, use it
	if row, ok := result.Get(0).(*sql.Row); ok {
		return row
	}

	// Otherwise, create a new row from the mock
	return m.db.QueryRow(query, args...)
}

func (m *MockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	result := m.Called(query, args)
	if rows, ok := result.Get(0).(*sql.Rows); ok {
		return rows, result.Error(1)
	}
	return m.db.Query(query, args...)
}

func (m *MockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	result := m.Called(query, args)
	if sqlResult, ok := result.Get(0).(sql.Result); ok {
		return sqlResult, result.Error(1)
	}
	return nil, result.Error(1)
}

func (m *MockDB) Begin() (*sql.Tx, error) {
	result := m.Called()
	if tx, ok := result.Get(0).(*sql.Tx); ok {
		return tx, result.Error(1)
	}
	return nil, result.Error(1)
}

func (m *MockDB) Close() error {
	result := m.Called()
	return result.Error(0)
}

func (m *MockDB) Driver() driver.Driver {
	result := m.Called()
	return result.Get(0).(driver.Driver)
}

func (m *MockDB) Ping() error {
	result := m.Called()
	return result.Error(0)
}

func (m *MockDB) Prepare(query string) (*sql.Stmt, error) {
	result := m.Called(query)
	if stmt, ok := result.Get(0).(*sql.Stmt); ok {
		return stmt, result.Error(1)
	}
	return nil, result.Error(1)
}

func (m *MockDB) SetConnMaxLifetime(d time.Duration) {
	m.Called(d)
}

func (m *MockDB) SetMaxIdleConns(n int) {
	m.Called(n)
}

func (m *MockDB) SetMaxOpenConns(n int) {
	m.Called(n)
}

func (m *MockDB) Stats() sql.DBStats {
	callArgs := m.Called()
	return callArgs.Get(0).(sql.DBStats)
}
