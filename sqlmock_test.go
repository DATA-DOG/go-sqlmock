package sqlmock

import (
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func TestMockQuery(t *testing.T) {
	db, err := sql.Open("mock", "")
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}

	rs := RowsFromCSVString([]string{"id", "title"}, "5,hello world")

	ExpectQuery("SELECT (.+) FROM articles WHERE id = ?").
		WithArgs(5).
		WillReturnRows(rs)

	rows, err := db.Query("SELECT (.+) FROM articles WHERE id = ?", 5)
	if err != nil {
		t.Errorf("error '%s' was not expected while retrieving mock rows", err)
	}
	defer rows.Close()
	if !rows.Next() {
		t.Error("it must have had one row as result, but got empty result set instead")
	}

	var id int
	var title string

	err = rows.Scan(&id, &title)
	if err != nil {
		t.Errorf("error '%s' was not expected while trying to scan row", err)
	}

	if id != 5 {
		t.Errorf("expected mocked id to be 5, but got %d instead", id)
	}

	if title != "hello world" {
		t.Errorf("expected mocked title to be 'hello world', but got '%s' instead", title)
	}

	if err = db.Close(); err != nil {
		t.Errorf("error '%s' was not expected while closing the database", err)
	}
}

func TestMockQueryTypes(t *testing.T) {
	db, err := sql.Open("mock", "")
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}

	columns := []string{"id", "timestamp", "sold"}

	timestamp := time.Now()
	rs := NewRows(columns)
	rs.AddRow(5, timestamp, true)

	ExpectQuery("SELECT (.+) FROM sales WHERE id = ?").
		WithArgs(5).
		WillReturnRows(rs)

	rows, err := db.Query("SELECT (.+) FROM sales WHERE id = ?", 5)
	if err != nil {
		t.Errorf("error '%s' was not expected while retrieving mock rows", err)
	}
	defer rows.Close()
	if !rows.Next() {
		t.Error("it must have had one row as result, but got empty result set instead")
	}

	var id int
	var time time.Time
	var sold bool

	err = rows.Scan(&id, &time, &sold)
	if err != nil {
		t.Errorf("error '%s' was not expected while trying to scan row", err)
	}

	if id != 5 {
		t.Errorf("expected mocked id to be 5, but got %d instead", id)
	}

	if time != timestamp {
		t.Errorf("expected mocked time to be %s, but got '%s' instead", timestamp, time)
	}

	if sold != true {
		t.Errorf("expected mocked boolean to be true, but got %v instead", sold)
	}

	if err = db.Close(); err != nil {
		t.Errorf("error '%s' was not expected while closing the database", err)
	}
}

func TestTransactionExpectations(t *testing.T) {
	db, err := sql.Open("mock", "")
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}

	// begin and commit
	ExpectBegin()
	ExpectCommit()

	tx, err := db.Begin()
	if err != nil {
		t.Errorf("an error '%s' was not expected when beginning a transaction", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Errorf("an error '%s' was not expected when commiting a transaction", err)
	}

	// begin and rollback
	ExpectBegin()
	ExpectRollback()

	tx, err = db.Begin()
	if err != nil {
		t.Errorf("an error '%s' was not expected when beginning a transaction", err)
	}

	err = tx.Rollback()
	if err != nil {
		t.Errorf("an error '%s' was not expected when rolling back a transaction", err)
	}

	// begin with an error
	ExpectBegin().WillReturnError(fmt.Errorf("some err"))

	tx, err = db.Begin()
	if err == nil {
		t.Error("an error was expected when beginning a transaction, but got none")
	}

	if err = db.Close(); err != nil {
		t.Errorf("error '%s' was not expected while closing the database", err)
	}
}

func TestPreparedQueryExecutions(t *testing.T) {
	db, err := sql.Open("mock", "")
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}

	rs1 := RowsFromCSVString([]string{"id", "title"}, "5,hello world")
	ExpectQuery("SELECT (.+) FROM articles WHERE id = ?").
		WithArgs(5).
		WillReturnRows(rs1)

	rs2 := RowsFromCSVString([]string{"id", "title"}, "2,whoop")
	ExpectQuery("SELECT (.+) FROM articles WHERE id = ?").
		WithArgs(2).
		WillReturnRows(rs2)

	stmt, err := db.Prepare("SELECT (.+) FROM articles WHERE id = ?")
	if err != nil {
		t.Errorf("error '%s' was not expected while creating a prepared statement", err)
	}

	var id int
	var title string

	err = stmt.QueryRow(5).Scan(&id, &title)
	if err != nil {
		t.Errorf("error '%s' was not expected querying row from statement and scanning", err)
	}

	if id != 5 {
		t.Errorf("expected mocked id to be 5, but got %d instead", id)
	}

	if title != "hello world" {
		t.Errorf("expected mocked title to be 'hello world', but got '%s' instead", title)
	}

	err = stmt.QueryRow(2).Scan(&id, &title)
	if err != nil {
		t.Errorf("error '%s' was not expected querying row from statement and scanning", err)
	}

	if id != 2 {
		t.Errorf("expected mocked id to be 2, but got %d instead", id)
	}

	if title != "whoop" {
		t.Errorf("expected mocked title to be 'whoop', but got '%s' instead", title)
	}

	if err = db.Close(); err != nil {
		t.Errorf("error '%s' was not expected while closing the database", err)
	}
}

func TestUnexpectedOperations(t *testing.T) {
	db, err := sql.Open("mock", "")
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}

	stmt, err := db.Prepare("SELECT (.+) FROM articles WHERE id = ?")
	if err != nil {
		t.Errorf("error '%s' was not expected while creating a prepared statement", err)
	}

	var id int
	var title string

	err = stmt.QueryRow(5).Scan(&id, &title)
	if err == nil {
		t.Error("error was expected querying row, since there was no such expectation")
	}

	ExpectRollback()

	err = db.Close()
	if err == nil {
		t.Error("error was expected while closing the database, expectation was not fulfilled", err)
	}
}

func TestWrongExpectations(t *testing.T) {
	db, err := sql.Open("mock", "")
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}

	ExpectBegin()

	rs1 := RowsFromCSVString([]string{"id", "title"}, "5,hello world")
	ExpectQuery("SELECT (.+) FROM articles WHERE id = ?").
		WithArgs(5).
		WillReturnRows(rs1)

	ExpectCommit().WillReturnError(fmt.Errorf("deadlock occured"))
	ExpectRollback() // won't be triggered

	stmt, err := db.Prepare("SELECT (.+) FROM articles WHERE id = ? FOR UPDATE")
	if err != nil {
		t.Errorf("error '%s' was not expected while creating a prepared statement", err)
	}

	var id int
	var title string

	err = stmt.QueryRow(5).Scan(&id, &title)
	if err == nil {
		t.Error("error was expected while querying row, since there begin transaction expectation is not fulfilled")
	}

	// lets go around and start transaction
	tx, err := db.Begin()
	if err != nil {
		t.Errorf("an error '%s' was not expected when beginning a transaction", err)
	}

	err = stmt.QueryRow(5).Scan(&id, &title)
	if err != nil {
		t.Errorf("error '%s' was not expected while querying row, since transaction was started", err)
	}

	err = tx.Commit()
	if err == nil {
		t.Error("a deadlock error was expected when commiting a transaction", err)
	}

	err = db.Close()
	if err == nil {
		t.Error("error was expected while closing the database, expectation was not fulfilled", err)
	}
}

func TestExecExpectations(t *testing.T) {
	db, err := sql.Open("mock", "")
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}

	result := NewResult(1, 1)
	ExpectExec("^INSERT INTO articles").
		WithArgs("hello").
		WillReturnResult(result)

	res, err := db.Exec("INSERT INTO articles (title) VALUES (?)", "hello")
	if err != nil {
		t.Errorf("error '%s' was not expected, while inserting a row", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		t.Errorf("error '%s' was not expected, while getting a last insert id", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		t.Errorf("error '%s' was not expected, while getting affected rows", err)
	}

	if id != 1 {
		t.Errorf("expected last insert id to be 1, but got %d instead", id)
	}

	if affected != 1 {
		t.Errorf("expected affected rows to be 1, but got %d instead", affected)
	}

	if err = db.Close(); err != nil {
		t.Errorf("error '%s' was not expected while closing the database", err)
	}
}
