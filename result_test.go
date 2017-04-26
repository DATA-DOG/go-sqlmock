package sqlmock

import (
	"fmt"
	"testing"
)

// used for examples
var mock = &sqlmock{}

func ExampleNewErrorResult() {
	db, mock, _ := New()
	result := NewErrorResult(fmt.Errorf("some error"))
	mock.ExpectExec("^INSERT (.+)").WillReturnResult(result)
	res, _ := db.Exec("INSERT something")
	_, err := res.LastInsertId()
	fmt.Println(err)
	// Output: some error
}

func ExampleNewResult() {
	var lastInsertID, affected int64
	result := NewResult(lastInsertID, affected)
	mock.ExpectExec("^INSERT (.+)").WillReturnResult(result)
	fmt.Println(mock.ExpectationsWereMet())
	// Output: there is a remaining expectation which was not matched: ExpectedExec => expecting Exec or ExecContext which:
	//   - matches sql: '^INSERT (.+)'
	//   - is without arguments
	//   - should return Result having:
	//       LastInsertId: 0
	//       RowsAffected: 0
}

func TestShouldReturnValidSqlDriverResult(t *testing.T) {
	result := NewResult(1, 2)
	id, err := result.LastInsertId()
	if 1 != id {
		t.Errorf("Expected last insert id to be 1, but got: %d", id)
	}
	if err != nil {
		t.Errorf("expected no error, but got: %s", err)
	}
	affected, err := result.RowsAffected()
	if 2 != affected {
		t.Errorf("Expected affected rows to be 2, but got: %d", affected)
	}
	if err != nil {
		t.Errorf("expected no error, but got: %s", err)
	}
}

func TestShouldReturnErroeSqlDriverResult(t *testing.T) {
	result := NewErrorResult(fmt.Errorf("some error"))
	_, err := result.LastInsertId()
	if err == nil {
		t.Error("expected error, but got none")
	}
	_, err = result.RowsAffected()
	if err == nil {
		t.Error("expected error, but got none")
	}
}
