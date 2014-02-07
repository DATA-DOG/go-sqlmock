package sqlmock

// Result satisfies sql driver Result, which
// holds last insert id and rows affected
// by Exec queries
type Result struct {
	lastInsertId int64
	rowsAffected int64
}

// Creates a new Result for Exec based query mocks
func NewResult(lastInsertId int64, rowsAffected int64) *Result {
	return &Result{
		lastInsertId,
		rowsAffected,
	}
}

// Retrieve last inserted id
func (res *Result) LastInsertId() (int64, error) {
	return res.lastInsertId, nil
}

// Retrieve number rows affected
func (res *Result) RowsAffected() (int64, error) {
	return res.rowsAffected, nil
}
