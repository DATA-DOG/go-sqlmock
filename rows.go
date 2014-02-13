package sqlmock

import (
	"database/sql/driver"
	"encoding/csv"
	"io"
	"strings"
)

// a struct which implements database/sql/driver.Rows
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

func (r *rows) AddRow(values ...interface{}) {
	if len(values) != len(r.cols) {
		panic("Expected number of values to match number of columns")
	}

	row := make([]driver.Value, len(r.cols))
	for i, v := range values {
		row[i] = v
	}

	r.rows = append(r.rows, row)
}

// NewRows allows Rows to be created manually to use
// any of the types sql/driver.Value supports
func NewRows(columns []string) *rows {
	rs := &rows{}
	rs.cols = columns
	return rs
}

// RowsFromCSVString creates Rows from CSV string
// to be used for mocked queries. Returns sql driver Rows interface
func RowsFromCSVString(columns []string, s string) driver.Rows {
	rs := &rows{}
	rs.cols = columns

	r := strings.NewReader(strings.TrimSpace(s))
	csvReader := csv.NewReader(r)

	for {
		r, err := csvReader.Read()
		if err != nil || r == nil {
			break
		}

		row := make([]driver.Value, len(columns))
		for i, v := range r {
			v := strings.TrimSpace(v)
			row[i] = []byte(v)
		}
		rs.rows = append(rs.rows, row)
	}
	return rs
}
