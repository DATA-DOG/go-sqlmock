// +build go1.8

package sqlmock

import (
	"fmt"
	"testing"
	"reflect"
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
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestNewColumnWithDefinition(t *testing.T) {
	column1 := NewColumn("test", "VARCHAR", "", true, 100, 0, 0)
	column2 := NewColumn("number", "DECIMAL", float64(0.0), false, 0, 10, 4)
	rows := NewRowsWithColumnDefiniton(column1, column2).AddRow("foo.bar", float64(10.123))

	db, mock, _ := New()
	mock.ExpectQuery("SELECT test, number from dummy").WillReturnRows(rows)

	query, _ := db.Query("SELECT test, number from dummy")

	if query.Next() {
		var test string
		var number float64

		if queryError := query.Scan(&test, &number); queryError != nil {
			t.Fatal(queryError)
		} else if test != "foo.bar" {
			t.Fatal("field test is not 'foo.bar'")
		} else if number != float64(10.123) {
			t.Fatal("field number is not '10.123'")
		}

		if columnTypes, colTypErr := query.ColumnTypes(); colTypErr != nil {
			t.Fatal(colTypErr)
		} else if len(columnTypes) != 2 {
			t.Fatal("number of columnTypes")
		} else if name := columnTypes[0].Name(); name != "test" {
			t.Fatalf("field 'test' has a wrong name '%s'", name)
		} else if dbTyp := columnTypes[0].DatabaseTypeName(); dbTyp != "VARCHAR" {
			t.Fatalf("field 'test' has a wrong db type '%s'", dbTyp)
		} else if columnTypes[0].ScanType().Kind() != reflect.String {
			t.Fatal("field 'test' has a wrong scanType")
		} else if precision, scale, _ := columnTypes[0].DecimalSize(); precision != 0 || scale != 0 {
			t.Fatal("field 'test' has a wrong precision, scale")
		} else if length, _ := columnTypes[0].Length(); length != 100 {
			t.Fatalf("field 'test' has a wrong length '%d'", length)
		} else if name := columnTypes[1].Name(); name != "number" {
			t.Fatalf("field 'number' has a wrong name '%s'", name)
		} else if dbTyp := columnTypes[1].DatabaseTypeName(); dbTyp != "DECIMAL" {
			t.Fatalf("field 'number' has a wrong db type '%s'", dbTyp)
		} else if columnTypes[1].ScanType().Kind() != reflect.Float64 {
			t.Fatal("field 'number' has a wrong scanType")
		} else if precision, scale, _ := columnTypes[1].DecimalSize(); precision != int64(10) || scale != int64(4) {
			t.Fatal("field 'number' has a wrong precision, scale")
		} else if length, _ := columnTypes[1].Length(); length != 0 {
			t.Fatal("field 'number' has a wrong length")
		}
	} else {
		t.Fatal("no result set")
	}

}
