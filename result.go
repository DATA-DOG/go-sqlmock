package sqlmock

type Result struct {
	lastInsertId int64
	rowsAffected int64
}

func NewResult(lastInsertId int64, rowsAffected int64) *Result {
	return &Result{
		lastInsertId,
		rowsAffected,
	}
}

func (res *Result) LastInsertId() (int64, error) {
	return res.lastInsertId, nil
}

func (res *Result) RowsAffected() (int64, error) {
	return res.rowsAffected, nil
}
