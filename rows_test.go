package sqlmock

import (
	"database/sql"
	"fmt"
	"testing"
)

func ExampleRows() {
	db, mock, err := New()
	if err != nil {
		fmt.Println("failed to open sqlmock database:", err)
	}
	defer db.Close()

	rows := NewRows([]string{"id", "title"}).
		AddRow(1, "one").
		AddRow(2, "two")

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	rs, _ := db.Query("SELECT")
	defer rs.Close()

	for rs.Next() {
		var id int
		var title string
		rs.Scan(&id, &title)
		fmt.Println("scanned id:", id, "and title:", title)
	}

	if rs.Err() != nil {
		fmt.Println("got rows error:", rs.Err())
	}
	// Output: scanned id: 1 and title: one
	// scanned id: 2 and title: two
}

func ExampleRows_rowError() {
	db, mock, err := New()
	if err != nil {
		fmt.Println("failed to open sqlmock database:", err)
	}
	defer db.Close()

	rows := NewRows([]string{"id", "title"}).
		AddRow(0, "one").
		AddRow(1, "two").
		RowError(1, fmt.Errorf("row error"))
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	rs, _ := db.Query("SELECT")
	defer rs.Close()

	for rs.Next() {
		var id int
		var title string
		rs.Scan(&id, &title)
		fmt.Println("scanned id:", id, "and title:", title)
	}

	if rs.Err() != nil {
		fmt.Println("got rows error:", rs.Err())
	}
	// Output: scanned id: 0 and title: one
	// got rows error: row error
}

func ExampleRows_closeError() {
	db, mock, err := New()
	if err != nil {
		fmt.Println("failed to open sqlmock database:", err)
	}
	defer db.Close()

	rows := NewRows([]string{"id", "title"}).CloseError(fmt.Errorf("close error"))
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	rs, _ := db.Query("SELECT")

	// Note: that close will return error only before rows EOF
	// that is a default sql package behavior. If you run rs.Next()
	// it will handle the error internally and return nil bellow
	if err := rs.Close(); err != nil {
		fmt.Println("got error:", err)
	}

	// Output: got error: close error
}

func TestAllowsToSetRowsErrors(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := NewRows([]string{"id", "title"}).
		AddRow(0, "one").
		AddRow(1, "two").
		RowError(1, fmt.Errorf("error"))
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	rs, err := db.Query("SELECT")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer rs.Close()

	if !rs.Next() {
		t.Fatal("expected the first row to be available")
	}
	if rs.Err() != nil {
		t.Fatalf("unexpected error: %s", rs.Err())
	}

	if rs.Next() {
		t.Fatal("was not expecting the second row, since there should be an error")
	}
	if rs.Err() == nil {
		t.Fatal("expected an error, but got none")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRowsCloseError(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := NewRows([]string{"id"}).CloseError(fmt.Errorf("close error"))
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	rs, err := db.Query("SELECT")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if err := rs.Close(); err == nil {
		t.Fatal("expected a close error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestQuerySingleRow(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := NewRows([]string{"id"}).
		AddRow(1).
		AddRow(2)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	var id int
	if err := db.QueryRow("SELECT").Scan(&id); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	mock.ExpectQuery("SELECT").WillReturnRows(NewRows([]string{"id"}))
	if err := db.QueryRow("SELECT").Scan(&id); err != sql.ErrNoRows {
		t.Fatal("expected sql no rows error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRowsScanError(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	r := NewRows([]string{"col1", "col2"}).AddRow("one", "two").AddRow("one", nil)
	mock.ExpectQuery("SELECT").WillReturnRows(r)

	rs, err := db.Query("SELECT")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer rs.Close()

	var one, two string
	if !rs.Next() || rs.Err() != nil || rs.Scan(&one, &two) != nil {
		t.Fatal("unexpected error on first row scan")
	}

	if !rs.Next() || rs.Err() != nil {
		t.Fatal("unexpected error on second row read")
	}

	err = rs.Scan(&one, &two)
	if err == nil {
		t.Fatal("expected an error for scan, but got none")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCSVRowParser(t *testing.T) {
	t.Parallel()
	rs := NewRows([]string{"col1", "col2"}).FromCSVString("a,NULL")
	db, mock, err := New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT").WillReturnRows(rs)

	rw, err := db.Query("SELECT")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer rw.Close()
	var col1 string
	var col2 []byte

	rw.Next()
	if err = rw.Scan(&col1, &col2); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if col1 != "a" {
		t.Fatalf("expected col1 to be 'a', but got [%T]:%+v", col1, col1)
	}
	if col2 != nil {
		t.Fatalf("expected col2 to be nil, but got [%T]:%+v", col2, col2)
	}
}

func TestWrongNumberOfValues(t *testing.T) {
	// Open new mock database
	db, mock, err := New()
	if err != nil {
		fmt.Println("error creating mock database")
		return
	}
	defer db.Close()
	defer func() {
		recover()
	}()
	mock.ExpectQuery("SELECT ID FROM TABLE").WithArgs(101).WillReturnRows(NewRows([]string{"ID"}).AddRow(101, "Hello"))
	db.Query("SELECT ID FROM TABLE", 101)
	// shouldn't reach here
	t.Error("expected panic from query")
}

func TestEmptyRowSets(t *testing.T) {
	rs1 := NewRows([]string{"a"}).AddRow("a")
	rs2 := NewRows([]string{"b"})
	rs3 := NewRows([]string{"c"})

	set1 := &rowSets{sets: []*Rows{rs1, rs2}}
	set2 := &rowSets{sets: []*Rows{rs3, rs2}}
	set3 := &rowSets{sets: []*Rows{rs2}}

	if set1.empty() {
		t.Fatalf("expected rowset 1, not to be empty, but it was")
	}
	if !set2.empty() {
		t.Fatalf("expected rowset 2, to be empty, but it was not")
	}
	if !set3.empty() {
		t.Fatalf("expected rowset 3, to be empty, but it was not")
	}
}
