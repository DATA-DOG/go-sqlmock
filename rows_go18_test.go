// +build go1.8

package sqlmock

import (
	"fmt"
	"testing"
)

func TestQueryMultiRows(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rs1 := NewRows([]string{"id", "title"}).AddRow(5, "hello world")
	rs2 := NewRows([]string{"name"}).AddRow("gopher").AddRow("john").AddRow("jane").RowError(2, fmt.Errorf("error"))

	mock.ExpectQuery("SELECT (.+) FROM articles WHERE id = \\?;SELECT name FROM users").
		WithArgs(5).
		WillReturnRows(rs1, rs2)

	rows, err := db.Query("SELECT id, title FROM articles WHERE id = ?;SELECT name FROM users", 5)
	if err != nil {
		t.Errorf("error was not expected, but got: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Error("expected a row to be available in first result set")
	}

	var id int
	var name string

	err = rows.Scan(&id, &name)
	if err != nil {
		t.Errorf("error was not expected, but got: %v", err)
	}

	if id != 5 || name != "hello world" {
		t.Errorf("unexpected row values id: %v name: %v", id, name)
	}

	if rows.Next() {
		t.Error("was not expecting next row in first result set")
	}

	if !rows.NextResultSet() {
		t.Error("had to have next result set")
	}

	if !rows.Next() {
		t.Error("expected a row to be available in second result set")
	}

	err = rows.Scan(&name)
	if err != nil {
		t.Errorf("error was not expected, but got: %v", err)
	}

	if name != "gopher" {
		t.Errorf("unexpected row name: %v", name)
	}

	if !rows.Next() {
		t.Error("expected a row to be available in second result set")
	}

	err = rows.Scan(&name)
	if err != nil {
		t.Errorf("error was not expected, but got: %v", err)
	}

	if name != "john" {
		t.Errorf("unexpected row name: %v", name)
	}

	if rows.Next() {
		t.Error("expected next row to produce error")
	}

	if rows.Err() == nil {
		t.Error("expected an error, but there was none")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}
