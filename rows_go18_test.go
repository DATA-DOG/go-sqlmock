//go:build go1.8
// +build go1.8

package sqlmock

import (
	"database/sql"
	"encoding/json"
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
