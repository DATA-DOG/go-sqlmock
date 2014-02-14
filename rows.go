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
	driver.Rows // composed interface, supports sql driver.Rows
	AddRow(...driver.Value) Rows
	FromCSVString(s string) Rows
}

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

// NewRows allows Rows to be created from a group of
// sql driver.Value or from the CSV string and
// to be used as sql driver.Rows
func NewRows(columns []string) Rows {
	return &rows{cols: columns}
}

// AddRow adds a row which is built from arguments
// in the same column order, returns sql driver.Rows
// compatible interface
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

// FromCSVString adds rows from CSV string.
// Returns sql driver.Rows compatible interface
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

// RowsFromCSVString creates Rows from CSV string
// to be used for mocked queries. Returns sql driver Rows interface
// ** DEPRECATED ** will be removed in the future, use Rows.FromCSVString
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
