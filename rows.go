package sqlmock

import (
	"database/sql/driver"
	"encoding/csv"
	"io"
	"strings"
)

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
