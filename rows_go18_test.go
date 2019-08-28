// +build go1.8

package sqlmock

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
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
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestQueryRowBytesInvalidatedByNext_jsonRawMessageIntoRawBytes(t *testing.T) {
	t.Parallel()
	replace := []byte(invalid)
	rows := NewRows([]string{"raw"}).
		AddRow(json.RawMessage(`{"thing": "one", "thing2": "two"}`)).
		AddRow(json.RawMessage(`{"that": "foo", "this": "bar"}`))
	scan := func(rs *sql.Rows) ([]byte, error) {
		var raw sql.RawBytes
		return raw, rs.Scan(&raw)
	}
	want := []struct {
		Initial  []byte
		Replaced []byte
	}{
		{Initial: []byte(`{"thing": "one", "thing2": "two"}`), Replaced: replace[:len(replace)-6]},
		{Initial: []byte(`{"that": "foo", "this": "bar"}`), Replaced: replace[:len(replace)-9]},
	}
	queryRowBytesInvalidatedByNext(t, rows, scan, want)
}

func TestQueryRowBytesNotInvalidatedByNext_jsonRawMessageIntoBytes(t *testing.T) {
	t.Parallel()
	rows := NewRows([]string{"raw"}).
		AddRow(json.RawMessage(`{"thing": "one", "thing2": "two"}`)).
		AddRow(json.RawMessage(`{"that": "foo", "this": "bar"}`))
	scan := func(rs *sql.Rows) ([]byte, error) {
		var b []byte
		return b, rs.Scan(&b)
	}
	want := [][]byte{[]byte(`{"thing": "one", "thing2": "two"}`), []byte(`{"that": "foo", "this": "bar"}`)}
	queryRowBytesNotInvalidatedByNext(t, rows, scan, want)
}

func TestQueryRowBytesNotInvalidatedByNext_bytesIntoCustomBytes(t *testing.T) {
	t.Parallel()
	rows := NewRows([]string{"raw"}).
		AddRow([]byte(`one binary value with some text!`)).
		AddRow([]byte(`two binary value with even more text than the first one`))
	scan := func(rs *sql.Rows) ([]byte, error) {
		type customBytes []byte
		var b customBytes
		return b, rs.Scan(&b)
	}
	want := [][]byte{[]byte(`one binary value with some text!`), []byte(`two binary value with even more text than the first one`)}
	queryRowBytesNotInvalidatedByNext(t, rows, scan, want)
}

func TestQueryRowBytesNotInvalidatedByNext_jsonRawMessageIntoCustomBytes(t *testing.T) {
	t.Parallel()
	rows := NewRows([]string{"raw"}).
		AddRow(json.RawMessage(`{"thing": "one", "thing2": "two"}`)).
		AddRow(json.RawMessage(`{"that": "foo", "this": "bar"}`))
	scan := func(rs *sql.Rows) ([]byte, error) {
		type customBytes []byte
		var b customBytes
		return b, rs.Scan(&b)
	}
	want := [][]byte{[]byte(`{"thing": "one", "thing2": "two"}`), []byte(`{"that": "foo", "this": "bar"}`)}
	queryRowBytesNotInvalidatedByNext(t, rows, scan, want)
}

func TestQueryRowBytesNotInvalidatedByClose_bytesIntoCustomBytes(t *testing.T) {
	t.Parallel()
	rows := NewRows([]string{"raw"}).AddRow([]byte(`one binary value with some text!`))
	scan := func(rs *sql.Rows) ([]byte, error) {
		type customBytes []byte
		var b customBytes
		return b, rs.Scan(&b)
	}
	queryRowBytesNotInvalidatedByClose(t, rows, scan, []byte(`one binary value with some text!`))
}

func TestQueryRowBytesInvalidatedByClose_jsonRawMessageIntoRawBytes(t *testing.T) {
	t.Parallel()
	replace := []byte(invalid)
	rows := NewRows([]string{"raw"}).AddRow(json.RawMessage(`{"thing": "one", "thing2": "two"}`))
	scan := func(rs *sql.Rows) ([]byte, error) {
		var raw sql.RawBytes
		return raw, rs.Scan(&raw)
	}
	want := struct {
		Initial  []byte
		Replaced []byte
	}{
		Initial:  []byte(`{"thing": "one", "thing2": "two"}`),
		Replaced: replace[:len(replace)-6],
	}
	queryRowBytesInvalidatedByClose(t, rows, scan, want)
}

func TestQueryRowBytesNotInvalidatedByClose_jsonRawMessageIntoBytes(t *testing.T) {
	t.Parallel()
	rows := NewRows([]string{"raw"}).AddRow(json.RawMessage(`{"thing": "one", "thing2": "two"}`))
	scan := func(rs *sql.Rows) ([]byte, error) {
		var b []byte
		return b, rs.Scan(&b)
	}
	queryRowBytesNotInvalidatedByClose(t, rows, scan, []byte(`{"thing": "one", "thing2": "two"}`))
}

func TestQueryRowBytesNotInvalidatedByClose_jsonRawMessageIntoCustomBytes(t *testing.T) {
	t.Parallel()
	rows := NewRows([]string{"raw"}).AddRow(json.RawMessage(`{"thing": "one", "thing2": "two"}`))
	scan := func(rs *sql.Rows) ([]byte, error) {
		type customBytes []byte
		var b customBytes
		return b, rs.Scan(&b)
	}
	queryRowBytesNotInvalidatedByClose(t, rows, scan, []byte(`{"thing": "one", "thing2": "two"}`))
}

func TestNewColumnWithDefinition(t *testing.T) {

	t.Run("with one ResultSet", func(t *testing.T) {
		db, mock, _ := New()
		column1 := mock.NewColumn("test", "VARCHAR", "", true, 100, 0, 0)
		column2 := mock.NewColumn("number", "DECIMAL", float64(0.0), false, 0, 10, 4)
		rows := mock.NewRowsWithColumnDefinition(column1, column2)
		rows.AddRow("foo.bar", float64(10.123))

		mQuery := mock.ExpectQuery("SELECT test, number from dummy")
		isQuery := mQuery.WillReturnRows(rows)
		isQueryClosed := mQuery.RowsWillBeClosed()
		isDbClosed := mock.ExpectClose()

		query, _ := db.Query("SELECT test, number from dummy")

		if false == isQuery.fulfilled() {
			t.Fatal("Query is not executed")
		}

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

		query.Close()
		if false == isQueryClosed.fulfilled() {
			t.Fatal("Query is not executed")
		}

		db.Close()
		if false == isDbClosed.fulfilled() {
			t.Fatal("Query is not closed")
		}
	})

	t.Run("with more then one ResultSet", func(t *testing.T) {
		db, mock, _ := New()
		column1 := mock.NewColumn("test", "VARCHAR", "", true, 100, 0, 0)
		column2 := mock.NewColumn("number", "DECIMAL", float64(0.0), false, 0, 10, 4)
		rows := mock.NewRowsWithColumnDefinition(column1, column2)
		rows.AddRow("foo.bar", float64(10.123))
		rows.AddRow("bar.foo", float64(123.10))
		rows.AddRow("lollipop", float64(10.321))

		mQuery := mock.ExpectQuery("SELECT test, number from dummy")
		isQuery := mQuery.WillReturnRows(rows)
		isQueryClosed := mQuery.RowsWillBeClosed()
		isDbClosed := mock.ExpectClose()

		query, _ := db.Query("SELECT test, number from dummy")

		if false == isQuery.fulfilled() {
			t.Fatal("Query is not executed")
		}

		rowsSi := 0

		if query.Next() {
			var test string
			var number float64

			if queryError := query.Scan(&test, &number); queryError != nil {
				t.Fatal(queryError)

			} else if rowsSi == 0 && test != "foo.bar" {
				t.Fatal("field test is not 'foo.bar'")
			} else if rowsSi == 0 && number != float64(10.123) {
				t.Fatal("field number is not '10.123'")

			} else if rowsSi == 1 && test != "bar.foo" {
				t.Fatal("field test is not 'bar.bar'")
			} else if rowsSi == 1 && number != float64(123.10) {
				t.Fatal("field number is not '123.10'")

			} else if rowsSi == 2 && test != "lollipop" {
				t.Fatal("field test is not 'lollipop'")
			} else if rowsSi == 2 && number != float64(10.321) {
				t.Fatal("field number is not '10.321'")
			}

			rowsSi++

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

		query.Close()
		if false == isQueryClosed.fulfilled() {
			t.Fatal("Query is not executed")
		}

		db.Close()
		if false == isDbClosed.fulfilled() {
			t.Fatal("Query is not closed")
		}
	})
}
