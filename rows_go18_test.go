// +build go1.8

package sqlmock

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"
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
	now, _ := time.Parse(time.RFC3339, "2020-06-20T22:08:41Z")

	t.Run("with one ResultSet", func(t *testing.T) {
		db, mock, _ := New()
		column1 := mock.NewColumn("test").OfType("VARCHAR", "").Nullable(true).WithLength(100)
		column2 := mock.NewColumn("number").OfType("DECIMAL", float64(0.0)).Nullable(false).WithPrecisionAndScale(10, 4)
		column3 := mock.NewColumn("when").OfType("TIMESTAMP", now)
		rows := mock.NewRowsWithColumnDefinition(column1, column2, column3)
		rows.AddRow("foo.bar", float64(10.123), now)

		mQuery := mock.ExpectQuery("SELECT test, number, when from dummy")
		isQuery := mQuery.WillReturnRows(rows)
		isQueryClosed := mQuery.RowsWillBeClosed()
		isDbClosed := mock.ExpectClose()

		query, _ := db.Query("SELECT test, number, when from dummy")

		if false == isQuery.fulfilled() {
			t.Error("Query is not executed")
		}

		if query.Next() {
			var test string
			var number float64
			var when time.Time

			if queryError := query.Scan(&test, &number, &when); queryError != nil {
				t.Error(queryError)
			} else if test != "foo.bar" {
				t.Error("field test is not 'foo.bar'")
			} else if number != float64(10.123) {
				t.Error("field number is not '10.123'")
			} else if when != now {
				t.Errorf("field when is not %v", now)
			}

			if columnTypes, colTypErr := query.ColumnTypes(); colTypErr != nil {
				t.Error(colTypErr)
			} else if len(columnTypes) != 3 {
				t.Error("number of columnTypes")
			} else if name := columnTypes[0].Name(); name != "test" {
				t.Errorf("field 'test' has a wrong name '%s'", name)
			} else if dbType := columnTypes[0].DatabaseTypeName(); dbType != "VARCHAR" {
				t.Errorf("field 'test' has a wrong db type '%s'", dbType)
			} else if columnTypes[0].ScanType().Kind() != reflect.String {
				t.Error("field 'test' has a wrong scanType")
			} else if _, _, ok := columnTypes[0].DecimalSize(); ok {
				t.Error("field 'test' should have not precision, scale")
			} else if length, ok := columnTypes[0].Length(); length != 100 || !ok {
				t.Errorf("field 'test' has a wrong length '%d'", length)
			} else if name := columnTypes[1].Name(); name != "number" {
				t.Errorf("field 'number' has a wrong name '%s'", name)
			} else if dbType := columnTypes[1].DatabaseTypeName(); dbType != "DECIMAL" {
				t.Errorf("field 'number' has a wrong db type '%s'", dbType)
			} else if columnTypes[1].ScanType().Kind() != reflect.Float64 {
				t.Error("field 'number' has a wrong scanType")
			} else if precision, scale, ok := columnTypes[1].DecimalSize(); precision != int64(10) || scale != int64(4) || !ok {
				t.Error("field 'number' has a wrong precision, scale")
			} else if _, ok := columnTypes[1].Length(); ok {
				t.Error("field 'number' is not variable length type")
			} else if _, ok := columnTypes[2].Nullable(); ok {
				t.Error("field 'when' should have nullability unknown")
			}
		} else {
			t.Error("no result set")
		}

		query.Close()
		if false == isQueryClosed.fulfilled() {
			t.Error("Query is not executed")
		}

		db.Close()
		if false == isDbClosed.fulfilled() {
			t.Error("Db is not closed")
		}
	})

	t.Run("with more then one ResultSet", func(t *testing.T) {
		db, mock, _ := New()
		column1 := mock.NewColumn("test").OfType("VARCHAR", "").Nullable(true).WithLength(100)
		column2 := mock.NewColumn("number").OfType("DECIMAL", float64(0.0)).Nullable(false).WithPrecisionAndScale(10, 4)
		column3 := mock.NewColumn("when").OfType("TIMESTAMP", now)
		rows1 := mock.NewRowsWithColumnDefinition(column1, column2, column3)
		rows1.AddRow("foo.bar", float64(10.123), now)
		rows2 := mock.NewRowsWithColumnDefinition(column1, column2, column3)
		rows2.AddRow("bar.foo", float64(123.10), now.Add(time.Second*10))
		rows3 := mock.NewRowsWithColumnDefinition(column1, column2, column3)
		rows3.AddRow("lollipop", float64(10.321), now.Add(time.Second*20))

		mQuery := mock.ExpectQuery("SELECT test, number, when from dummy")
		isQuery := mQuery.WillReturnRows(rows1, rows2, rows3)
		isQueryClosed := mQuery.RowsWillBeClosed()
		isDbClosed := mock.ExpectClose()

		query, _ := db.Query("SELECT test, number, when from dummy")

		if false == isQuery.fulfilled() {
			t.Error("Query is not executed")
		}

		rowsSi := 0

		for query.Next() {
			var test string
			var number float64
			var when time.Time

			if queryError := query.Scan(&test, &number, &when); queryError != nil {
				t.Error(queryError)

			} else if rowsSi == 0 && test != "foo.bar" {
				t.Error("field test is not 'foo.bar'")
			} else if rowsSi == 0 && number != float64(10.123) {
				t.Error("field number is not '10.123'")
			} else if rowsSi == 0 && when != now {
				t.Errorf("field when is not %v", now)

			} else if rowsSi == 1 && test != "bar.foo" {
				t.Error("field test is not 'bar.bar'")
			} else if rowsSi == 1 && number != float64(123.10) {
				t.Error("field number is not '123.10'")
			} else if rowsSi == 1 && when != now.Add(time.Second*10) {
				t.Errorf("field when is not %v", now)

			} else if rowsSi == 2 && test != "lollipop" {
				t.Error("field test is not 'lollipop'")
			} else if rowsSi == 2 && number != float64(10.321) {
				t.Error("field number is not '10.321'")
			} else if rowsSi == 2 && when != now.Add(time.Second*20) {
				t.Errorf("field when is not %v", now)
			}

			rowsSi++

			if columnTypes, colTypErr := query.ColumnTypes(); colTypErr != nil {
				t.Error(colTypErr)
			} else if len(columnTypes) != 3 {
				t.Error("number of columnTypes")
			} else if name := columnTypes[0].Name(); name != "test" {
				t.Errorf("field 'test' has a wrong name '%s'", name)
			} else if dbType := columnTypes[0].DatabaseTypeName(); dbType != "VARCHAR" {
				t.Errorf("field 'test' has a wrong db type '%s'", dbType)
			} else if columnTypes[0].ScanType().Kind() != reflect.String {
				t.Error("field 'test' has a wrong scanType")
			} else if _, _, ok := columnTypes[0].DecimalSize(); ok {
				t.Error("field 'test' should not have precision, scale")
			} else if length, ok := columnTypes[0].Length(); length != 100 || !ok {
				t.Errorf("field 'test' has a wrong length '%d'", length)
			} else if name := columnTypes[1].Name(); name != "number" {
				t.Errorf("field 'number' has a wrong name '%s'", name)
			} else if dbType := columnTypes[1].DatabaseTypeName(); dbType != "DECIMAL" {
				t.Errorf("field 'number' has a wrong db type '%s'", dbType)
			} else if columnTypes[1].ScanType().Kind() != reflect.Float64 {
				t.Error("field 'number' has a wrong scanType")
			} else if precision, scale, ok := columnTypes[1].DecimalSize(); precision != int64(10) || scale != int64(4) || !ok {
				t.Error("field 'number' has a wrong precision, scale")
			} else if _, ok := columnTypes[1].Length(); ok {
				t.Error("field 'number' is not variable length type")
			} else if _, ok := columnTypes[2].Nullable(); ok {
				t.Error("field 'when' should have nullability unknown")
			}
		}
		if rowsSi == 0 {
			t.Error("no result set")
		}

		query.Close()
		if false == isQueryClosed.fulfilled() {
			t.Error("Query is not executed")
		}

		db.Close()
		if false == isDbClosed.fulfilled() {
			t.Error("Db is not closed")
		}
	})
}
