// +build go1.3

package sqlmock

import (
	"database/sql"
	"testing"
)

func TestQueryRowBytesNotInvalidatedByNext_stringIntoRawBytes(t *testing.T) {
	t.Parallel()
	rows := NewRows([]string{"raw"}).
		AddRow(`one binary value with some text!`).
		AddRow(`two binary value with even more text than the first one`)
	scan := func(rs *sql.Rows) ([]byte, error) {
		var raw sql.RawBytes
		return raw, rs.Scan(&raw)
	}
	want := [][]byte{[]byte(`one binary value with some text!`), []byte(`two binary value with even more text than the first one`)}
	queryRowBytesNotInvalidatedByNext(t, rows, scan, want)
}

func TestQueryRowBytesNotInvalidatedByClose_stringIntoRawBytes(t *testing.T) {
	t.Parallel()
	rows := NewRows([]string{"raw"}).AddRow(`one binary value with some text!`)
	scan := func(rs *sql.Rows) ([]byte, error) {
		var raw sql.RawBytes
		return raw, rs.Scan(&raw)
	}
	queryRowBytesNotInvalidatedByClose(t, rows, scan, []byte(`one binary value with some text!`))
}
