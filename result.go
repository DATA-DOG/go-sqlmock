package sqlmock

import (
    "database/sql/driver"
)

// Result satisfies sql driver Result, which
// holds last insert id and rows affected
// by Exec queries
type result struct {
	insertID int64
	rowsAffected int64
}

// NewResult creates a new sql driver Result
// for Exec based query mocks.
func NewResult(lastInsertID int64, rowsAffected int64) driver.Result {
	return &result{
		lastInsertID,
		rowsAffected,
	}
}

func (r *result) LastInsertId() (int64, error) {
	return r.insertID, nil
}

func (r *result) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}
