package sqlmock

// a structure which implements database/sql/driver.Result
// holds last insert id and rows affected
// should be returned by Exec queries
type Result struct {
	lastInsertId int64
	rowsAffected int64
}

// creates a new result for Exec based query mocks
func NewResult(lastInsertId int64, rowsAffected int64) *Result {
	return &Result{
		lastInsertId,
		rowsAffected,
	}
}

// get last insert id
func (res *Result) LastInsertId() (int64, error) {
	return res.lastInsertId, nil
}

// get rows affected
func (res *Result) RowsAffected() (int64, error) {
	return res.rowsAffected, nil
}
