package sqlmock

import (
	"database/sql/driver"
	"encoding/csv"
	"io"
	"strings"
)

// Rows interface allows to construct rows
// which also satisfies database/sql/driver.Rows interface
type Rows interface {
	// composed interface, supports sql driver.Rows
	driver.Rows

	// AddRow composed from database driver.Value slice
	// return the same instance to perform subsequent actions.
	// Note that the number of values must match the number
	// of columns
	AddRow(...driver.Value) Rows

	// FromCSVString build rows from csv string.
	// return the same instance to perform subsequent actions.
	// Note that the number of values must match the number
	// of columns
	FromCSVString(s string) Rows
}

type rows struct {
	cols []string
	rows [][]driver.Value
	pos  int
}

func (r *rows) Columns() []string {
	return r.cols
}

func (r *rows) Close() error {
	return nil
}

func (r *rows) Err() error {
	return nil
}

// advances to next row
func (r *rows) Next(dest []driver.Value) error {
	r.pos++
	if r.pos > len(r.rows) {
		return io.EOF // per interface spec
	}

	for i, col := range r.rows[r.pos-1] {
		dest[i] = col
	}

	return nil
}

// NewRows allows Rows to be created from a
// sql driver.Value slice or from the CSV string and
// to be used as sql driver.Rows
func NewRows(columns []string) Rows {
	return &rows{cols: columns}
}

func (r *rows) AddRow(values ...driver.Value) Rows {
	if len(values) != len(r.cols) {
		panic("Expected number of values to match number of columns")
	}

	row := make([]driver.Value, len(r.cols))
	for i, v := range values {
		row[i] = v
	}

	r.rows = append(r.rows, row)
	return r
}

func (r *rows) FromCSVString(s string) Rows {
	res := strings.NewReader(strings.TrimSpace(s))
	csvReader := csv.NewReader(res)

	for {
		res, err := csvReader.Read()
		if err != nil || res == nil {
			break
		}

		row := make([]driver.Value, len(r.cols))
		for i, v := range res {
			row[i] = []byte(strings.TrimSpace(v))
		}
		r.rows = append(r.rows, row)
	}
	return r
}
