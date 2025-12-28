package testutil

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// NewMockDB creates a new sqlmock database connection for testing
func NewMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock db: %v", err)
	}
	return db, mock
}

// NewMockDBWithQueryMatcher creates a mock with custom query matcher
func NewMockDBWithQueryMatcher(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("Failed to create mock db: %v", err)
	}
	return db, mock
}

// ExpectClose sets up expectation for database close
func ExpectClose(mock sqlmock.Sqlmock) {
	mock.ExpectClose()
}

// AssertExpectations verifies all expectations were met
func AssertExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}
