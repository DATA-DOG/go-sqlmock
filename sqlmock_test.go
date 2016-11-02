package sqlmock

import (
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"
)

func cancelOrder(db *sql.DB, orderID int) error {
	tx, _ := db.Begin()
	_, _ = tx.Query("SELECT * FROM orders {0} FOR UPDATE", orderID)
	_ = tx.Rollback()
	return nil
}

func Example() {
	// Open new mock database
	db, mock, err := New()
	if err != nil {
		fmt.Println("error creating mock database")
		return
	}
	// columns to be used for result
	columns := []string{"id", "status"}
	// expect transaction begin
	mock.ExpectBegin()
	// expect query to fetch order, match it with regexp
	mock.ExpectQuery("SELECT (.+) FROM orders (.+) FOR UPDATE").
		WithArgs(1).
		WillReturnRows(NewRows(columns).AddRow(1, 1))
	// expect transaction rollback, since order status is "cancelled"
	mock.ExpectRollback()

	// run the cancel order function
	someOrderID := 1
	// call a function which executes expected database operations
	err = cancelOrder(db, someOrderID)
	if err != nil {
		fmt.Printf("unexpected error: %s", err)
		return
	}

	// ensure all expectations have been met
	if err = mock.ExpectationsWereMet(); err != nil {
		fmt.Printf("unmet expectation error: %s", err)
	}
	// Output:
}

func TestIssue14EscapeSQL(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mock.ExpectExec("INSERT INTO mytable\\(a, b\\)").
		WithArgs("A", "B").
		WillReturnResult(NewResult(1, 1))

	_, err = db.Exec("INSERT INTO mytable(a, b) VALUES (?, ?)", "A", "B")
	if err != nil {
		t.Errorf("error '%s' was not expected, while inserting a row", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

// test the case when db is not triggered and expectations
// are not asserted on close
func TestIssue4(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("some sql query which will not be called").
		WillReturnRows(NewRows([]string{"id"}))

	if err := mock.ExpectationsWereMet(); err == nil {
		t.Errorf("was expecting an error since query was not triggered")
	}
}

func TestMockQuery(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rs := NewRows([]string{"id", "title"}).FromCSVString("5,hello world")

	mock.ExpectQuery("SELECT (.+) FROM articles WHERE id = ?").
		WithArgs(5).
		WillReturnRows(rs)

	rows, err := db.Query("SELECT (.+) FROM articles WHERE id = ?", 5)
	if err != nil {
		t.Errorf("error '%s' was not expected while retrieving mock rows", err)
	}

	defer func() {
		if er := rows.Close(); er != nil {
			t.Error("Unexpected error while trying to close rows")
		}
	}()

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

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestMockQueryTypes(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	columns := []string{"id", "timestamp", "sold"}

	timestamp := time.Now()
	rs := NewRows(columns)
	rs.AddRow(5, timestamp, true)

	mock.ExpectQuery("SELECT (.+) FROM sales WHERE id = ?").
		WithArgs(5).
		WillReturnRows(rs)

	rows, err := db.Query("SELECT (.+) FROM sales WHERE id = ?", 5)
	if err != nil {
		t.Errorf("error '%s' was not expected while retrieving mock rows", err)
	}
	defer func() {
		if er := rows.Close(); er != nil {
			t.Error("Unexpected error while trying to close rows")
		}
	}()
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

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestTransactionExpectations(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// begin and commit
	mock.ExpectBegin()
	mock.ExpectCommit()

	tx, err := db.Begin()
	if err != nil {
		t.Errorf("an error '%s' was not expected when beginning a transaction", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Errorf("an error '%s' was not expected when commiting a transaction", err)
	}

	// begin and rollback
	mock.ExpectBegin()
	mock.ExpectRollback()

	tx, err = db.Begin()
	if err != nil {
		t.Errorf("an error '%s' was not expected when beginning a transaction", err)
	}

	err = tx.Rollback()
	if err != nil {
		t.Errorf("an error '%s' was not expected when rolling back a transaction", err)
	}

	// begin with an error
	mock.ExpectBegin().WillReturnError(fmt.Errorf("some err"))

	tx, err = db.Begin()
	if err == nil {
		t.Error("an error was expected when beginning a transaction, but got none")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestPrepareExpectations(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectPrepare("SELECT (.+) FROM articles WHERE id = ?")

	stmt, err := db.Prepare("SELECT (.+) FROM articles WHERE id = ?")
	if err != nil {
		t.Errorf("error '%s' was not expected while creating a prepared statement", err)
	}
	if stmt == nil {
		t.Errorf("stmt was expected while creating a prepared statement")
	}

	// expect something else, w/o ExpectPrepare()
	var id int
	var title string
	rs := NewRows([]string{"id", "title"}).FromCSVString("5,hello world")

	mock.ExpectQuery("SELECT (.+) FROM articles WHERE id = ?").
		WithArgs(5).
		WillReturnRows(rs)

	err = stmt.QueryRow(5).Scan(&id, &title)
	if err != nil {
		t.Errorf("error '%s' was not expected while retrieving mock rows", err)
	}

	mock.ExpectPrepare("SELECT (.+) FROM articles WHERE id = ?").
		WillReturnError(fmt.Errorf("Some DB error occurred"))

	stmt, err = db.Prepare("SELECT id FROM articles WHERE id = ?")
	if err == nil {
		t.Error("error was expected while creating a prepared statement")
	}
	if stmt != nil {
		t.Errorf("stmt was not expected while creating a prepared statement returning error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestPreparedQueryExecutions(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectPrepare("SELECT (.+) FROM articles WHERE id = ?")

	rs1 := NewRows([]string{"id", "title"}).FromCSVString("5,hello world")
	mock.ExpectQuery("SELECT (.+) FROM articles WHERE id = ?").
		WithArgs(5).
		WillReturnRows(rs1)

	rs2 := NewRows([]string{"id", "title"}).FromCSVString("2,whoop")
	mock.ExpectQuery("SELECT (.+) FROM articles WHERE id = ?").
		WithArgs(2).
		WillReturnRows(rs2)

	stmt, err := db.Prepare("SELECT id, title FROM articles WHERE id = ?")
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

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestUnexpectedOperations(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectPrepare("SELECT (.+) FROM articles WHERE id = ?")
	stmt, err := db.Prepare("SELECT id, title FROM articles WHERE id = ?")
	if err != nil {
		t.Errorf("error '%s' was not expected while creating a prepared statement", err)
	}

	var id int
	var title string

	err = stmt.QueryRow(5).Scan(&id, &title)
	if err == nil {
		t.Error("error was expected querying row, since there was no such expectation")
	}

	mock.ExpectRollback()

	if err := mock.ExpectationsWereMet(); err == nil {
		t.Errorf("was expecting an error since query was not triggered")
	}
}

func TestWrongExpectations(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin()

	rs1 := NewRows([]string{"id", "title"}).FromCSVString("5,hello world")
	mock.ExpectQuery("SELECT (.+) FROM articles WHERE id = ?").
		WithArgs(5).
		WillReturnRows(rs1)

	mock.ExpectCommit().WillReturnError(fmt.Errorf("deadlock occured"))
	mock.ExpectRollback() // won't be triggered

	var id int
	var title string

	err = db.QueryRow("SELECT id, title FROM articles WHERE id = ? FOR UPDATE", 5).Scan(&id, &title)
	if err == nil {
		t.Error("error was expected while querying row, since there begin transaction expectation is not fulfilled")
	}

	// lets go around and start transaction
	tx, err := db.Begin()
	if err != nil {
		t.Errorf("an error '%s' was not expected when beginning a transaction", err)
	}

	err = db.QueryRow("SELECT id, title FROM articles WHERE id = ? FOR UPDATE", 5).Scan(&id, &title)
	if err != nil {
		t.Errorf("error '%s' was not expected while querying row, since transaction was started", err)
	}

	err = tx.Commit()
	if err == nil {
		t.Error("a deadlock error was expected when commiting a transaction", err)
	}

	if err := mock.ExpectationsWereMet(); err == nil {
		t.Errorf("was expecting an error since query was not triggered")
	}
}

func TestExecExpectations(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	result := NewResult(1, 1)
	mock.ExpectExec("^INSERT INTO articles").
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

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestRowBuilderAndNilTypes(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rs := NewRows([]string{"id", "active", "created", "status"}).
		AddRow(1, true, time.Now(), 5).
		AddRow(2, false, nil, nil)

	mock.ExpectQuery("SELECT (.+) FROM sales").WillReturnRows(rs)

	rows, err := db.Query("SELECT * FROM sales")
	if err != nil {
		t.Errorf("error '%s' was not expected while retrieving mock rows", err)
	}
	defer func() {
		if er := rows.Close(); er != nil {
			t.Error("Unexpected error while trying to close rows")
		}
	}()

	// NullTime and NullInt are used from stubs_test.go
	var (
		id      int
		active  bool
		created NullTime
		status  NullInt
	)

	if !rows.Next() {
		t.Error("it must have had row in rows, but got empty result set instead")
	}

	err = rows.Scan(&id, &active, &created, &status)
	if err != nil {
		t.Errorf("error '%s' was not expected while trying to scan row", err)
	}

	if id != 1 {
		t.Errorf("expected mocked id to be 1, but got %d instead", id)
	}

	if !active {
		t.Errorf("expected 'active' to be 'true', but got '%v' instead", active)
	}

	if !created.Valid {
		t.Errorf("expected 'created' to be valid, but it %+v is not", created)
	}

	if !status.Valid {
		t.Errorf("expected 'status' to be valid, but it %+v is not", status)
	}

	if status.Integer != 5 {
		t.Errorf("expected 'status' to be '5', but got '%d'", status.Integer)
	}

	// test second row
	if !rows.Next() {
		t.Error("it must have had row in rows, but got empty result set instead")
	}

	err = rows.Scan(&id, &active, &created, &status)
	if err != nil {
		t.Errorf("error '%s' was not expected while trying to scan row", err)
	}

	if id != 2 {
		t.Errorf("expected mocked id to be 2, but got %d instead", id)
	}

	if active {
		t.Errorf("expected 'active' to be 'false', but got '%v' instead", active)
	}

	if created.Valid {
		t.Errorf("expected 'created' to be invalid, but it %+v is not", created)
	}

	if status.Valid {
		t.Errorf("expected 'status' to be invalid, but it %+v is not", status)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestArgumentReflectValueTypeError(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rs := NewRows([]string{"id"}).AddRow(1)

	mock.ExpectQuery("SELECT (.+) FROM sales").WithArgs(5.5).WillReturnRows(rs)

	_, err = db.Query("SELECT * FROM sales WHERE x = ?", 5)
	if err == nil {
		t.Error("Expected error, but got none")
	}
}

func TestGoroutineExecutionWithUnorderedExpectationMatching(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// note this line is important for unordered expectation matching
	mock.MatchExpectationsInOrder(false)

	result := NewResult(1, 1)

	mock.ExpectExec("^UPDATE one").WithArgs("one").WillReturnResult(result)
	mock.ExpectExec("^UPDATE two").WithArgs("one", "two").WillReturnResult(result)
	mock.ExpectExec("^UPDATE three").WithArgs("one", "two", "three").WillReturnResult(result)

	var wg sync.WaitGroup
	queries := map[string][]interface{}{
		"one":   []interface{}{"one"},
		"two":   []interface{}{"one", "two"},
		"three": []interface{}{"one", "two", "three"},
	}

	wg.Add(len(queries))
	for table, args := range queries {
		go func(tbl string, a []interface{}) {
			if _, err := db.Exec("UPDATE "+tbl, a...); err != nil {
				t.Errorf("error was not expected: %s", err)
			}
			wg.Done()
		}(table, args)
	}

	wg.Wait()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func ExampleSqlmock_goroutines() {
	db, mock, err := New()
	if err != nil {
		fmt.Println("failed to open sqlmock database:", err)
	}
	defer db.Close()

	// note this line is important for unordered expectation matching
	mock.MatchExpectationsInOrder(false)

	result := NewResult(1, 1)

	mock.ExpectExec("^UPDATE one").WithArgs("one").WillReturnResult(result)
	mock.ExpectExec("^UPDATE two").WithArgs("one", "two").WillReturnResult(result)
	mock.ExpectExec("^UPDATE three").WithArgs("one", "two", "three").WillReturnResult(result)

	var wg sync.WaitGroup
	queries := map[string][]interface{}{
		"one":   []interface{}{"one"},
		"two":   []interface{}{"one", "two"},
		"three": []interface{}{"one", "two", "three"},
	}

	wg.Add(len(queries))
	for table, args := range queries {
		go func(tbl string, a []interface{}) {
			if _, err := db.Exec("UPDATE "+tbl, a...); err != nil {
				fmt.Println("error was not expected:", err)
			}
			wg.Done()
		}(table, args)
	}

	wg.Wait()

	if err := mock.ExpectationsWereMet(); err != nil {
		fmt.Println("there were unfulfilled expections:", err)
	}
	// Output:
}

// False Positive - passes despite mismatched Exec
// see #37 issue
func TestRunExecsWithOrderedShouldNotMeetAllExpectations(t *testing.T) {
	db, dbmock, _ := New()
	dbmock.ExpectExec("THE FIRST EXEC")
	dbmock.ExpectExec("THE SECOND EXEC")

	_, _ = db.Exec("THE FIRST EXEC")
	_, _ = db.Exec("THE WRONG EXEC")

	err := dbmock.ExpectationsWereMet()
	if err == nil {
		t.Fatal("was expecting an error, but there wasn't any")
	}
}

// False Positive - passes despite mismatched Exec
// see #37 issue
func TestRunQueriesWithOrderedShouldNotMeetAllExpectations(t *testing.T) {
	db, dbmock, _ := New()
	dbmock.ExpectQuery("THE FIRST QUERY")
	dbmock.ExpectQuery("THE SECOND QUERY")

	_, _ = db.Query("THE FIRST QUERY")
	_, _ = db.Query("THE WRONG QUERY")

	err := dbmock.ExpectationsWereMet()
	if err == nil {
		t.Fatal("was expecting an error, but there wasn't any")
	}
}

func TestRunExecsWithExpectedErrorMeetsExpectations(t *testing.T) {
	db, dbmock, _ := New()
	dbmock.ExpectExec("THE FIRST EXEC").WillReturnError(fmt.Errorf("big bad bug"))
	dbmock.ExpectExec("THE SECOND EXEC").WillReturnResult(NewResult(0, 0))

	_, _ = db.Exec("THE FIRST EXEC")
	_, _ = db.Exec("THE SECOND EXEC")

	err := dbmock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("all expectations should be met: %s", err)
	}
}

func TestRunQueryWithExpectedErrorMeetsExpectations(t *testing.T) {
	db, dbmock, _ := New()
	dbmock.ExpectQuery("THE FIRST QUERY").WillReturnError(fmt.Errorf("big bad bug"))
	dbmock.ExpectQuery("THE SECOND QUERY").WillReturnRows(NewRows([]string{"col"}).AddRow(1))

	_, _ = db.Query("THE FIRST QUERY")
	_, _ = db.Query("THE SECOND QUERY")

	err := dbmock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("all expectations should be met: %s", err)
	}
}

func TestEmptyRowSet(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rs := NewRows([]string{"id", "title"})

	mock.ExpectQuery("SELECT (.+) FROM articles WHERE id = ?").
		WithArgs(5).
		WillReturnRows(rs)

	rows, err := db.Query("SELECT (.+) FROM articles WHERE id = ?", 5)
	if err != nil {
		t.Errorf("error '%s' was not expected while retrieving mock rows", err)
	}

	defer func() {
		if er := rows.Close(); er != nil {
			t.Error("Unexpected error while trying to close rows")
		}
	}()

	if rows.Next() {
		t.Error("expected no rows but got one")
	}

	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("all expectations should be met: %s", err)
	}
}

// Based on issue #50
func TestPrepareExpectationNotFulfilled(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectPrepare("^BADSELECT$")

	if _, err := db.Prepare("SELECT"); err == nil {
		t.Fatal("prepare should not match expected query string")
	}

	if err := mock.ExpectationsWereMet(); err == nil {
		t.Errorf("was expecting an error, since prepared statement query does not match, but there was none")
	}
}
