package sqlmock

import (
	"database/sql/driver"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"reflect"
)

// CSVColumnParser is a function which converts trimmed csv
// column string to a []byte representation. currently
// transforms NULL to nil
var CSVColumnParser = func(s string) []byte {
	switch {
	case strings.ToLower(s) == "null":
		return nil
	}
	return []byte(s)
}

type rowSets struct {
	sets []*Rows
	pos  int
	ex   *ExpectedQuery
}

func (rs *rowSets) Columns() []string {
	return rs.sets[rs.pos].cols
}

func (rs *rowSets) Close() error {
	rs.ex.rowsWereClosed = true
	return rs.sets[rs.pos].closeErr
}

// advances to next row
func (rs *rowSets) Next(dest []driver.Value) error {
	r := rs.sets[rs.pos]
	r.pos++
	if r.pos > len(r.rows) {
		return io.EOF // per interface spec
	}

	for i, col := range r.rows[r.pos-1] {
		dest[i] = col
	}

	return r.nextErr[r.pos-1]
}

// ColumnTypeLength is defined from driver.RowColumnTypeLength
func (rs *rowSets) ColumnTypeLength(index int) (length int64, ok bool) {
	return rs.sets[0].def[index].length, false
}

// ColumnTypeNullable is defined from driver.RowColumnTypeNullable
func (rs *rowSets) ColumnTypeNullable(index int) (nullable, ok bool) {
	return rs.sets[0].def[index].nullable, false
}

// ColumnTypePrecisionScale is defined from driver.RowColumnTypePrecisionScale
func (rs *rowSets) ColumnTypePrecisionScale(index int) (precision, scale int64, ok bool) {
	return rs.sets[0].def[index].precision, rs.sets[0].def[index].scale, false
}

// ColumnTypeScanType is defined from driver.RowsColumnTypeScanType
func (rs *rowSets) ColumnTypeScanType(index int) reflect.Type {
	return rs.sets[0].def[index].scanType
}

// ColumnTypeDatabaseTypeName is defined RowsColumnTypeDatabaseTypeName
func (rs *rowSets) ColumnTypeDatabaseTypeName(index int) string {
	return rs.sets[0].def[index].dbTyp
}

// transforms to debuggable printable string
func (rs *rowSets) String() string {
	if rs.empty() {
		return "with empty rows"
	}

	msg := "should return rows:\n"
	if len(rs.sets) == 1 {
		for n, row := range rs.sets[0].rows {
			msg += fmt.Sprintf("    row %d - %+v\n", n, row)
		}
		return strings.TrimSpace(msg)
	}
	for i, set := range rs.sets {
		msg += fmt.Sprintf("    result set: %d\n", i)
		for n, row := range set.rows {
			msg += fmt.Sprintf("      row %d - %+v\n", n, row)
		}
	}
	return strings.TrimSpace(msg)
}

func (rs *rowSets) empty() bool {
	for _, set := range rs.sets {
		if len(set.rows) > 0 {
			return false
		}
	}
	return true
}

// Column is a mocked column Metadate for rows.ColumnTypes()
type Column struct {
	name, dbTyp              string
	nullable                 bool
	length, precision, scale int64
	scanType                 reflect.Type
}

// Rows is a mocked collection of rows to
// return for Query result
type Rows struct {
	converter driver.ValueConverter
	cols      []string
	def       []*Column
	rows      [][]driver.Value
	pos       int
	nextErr   map[int]error
	closeErr  error
}

// New Column allows to create a Column Metadata definition
func NewColumn(name, dbTyp string, exampleValue interface{}, nullable bool, length, precision, scale int64) *Column {
	return &Column{name, dbTyp, nullable, length, precision, scale, reflect.TypeOf(exampleValue)}
}

func NewColumnSimple(name string) *Column {
	return &Column{name: name}
}

// NewRows allows Rows to be created from a
// sql driver.Value slice or from the CSV string and
// to be used as sql driver.Rows.
// Use Sqlmock.NewRows instead if using a custom converter
func NewRows(columns []string) *Rows {
	definition := make([]*Column, len(columns))
	for i, column := range columns {
		definition[i] = NewColumnSimple(column)
	}

	return &Rows{
		cols:      columns,
		def:       definition,
		nextErr:   make(map[int]error),
		converter: driver.DefaultParameterConverter,
	}
}

// NewRowsWithColumnDefiniton see PR-152
func NewRowsWithColumnDefiniton(columns ...*Column) *Rows {
	cols := make([]string, len(columns))
	for i, column := range columns {
		cols[i] = column.name
	}

	return &Rows{
		cols:      cols,
		def:       columns,
		nextErr:   make(map[int]error),
		converter: driver.DefaultParameterConverter,
	}
}

// CloseError allows to set an error
// which will be returned by rows.Close
// function.
//
// The close error will be triggered only in cases
// when rows.Next() EOF was not yet reached, that is
// a default sql library behavior
func (r *Rows) CloseError(err error) *Rows {
	r.closeErr = err
	return r
}

// RowError allows to set an error
// which will be returned when a given
// row number is read
func (r *Rows) RowError(row int, err error) *Rows {
	r.nextErr[row] = err
	return r
}

// AddRow composed from database driver.Value slice
// return the same instance to perform subsequent actions.
// Note that the number of values must match the number
// of columns
func (r *Rows) AddRow(values ...driver.Value) *Rows {
	if len(values) != len(r.cols) {
		panic("Expected number of values to match number of columns")
	}

	row := make([]driver.Value, len(r.cols))
	for i, v := range values {
		// Convert user-friendly values (such as int or driver.Valuer)
		// to database/sql native value (driver.Value such as int64)
		var err error
		v, err = r.converter.ConvertValue(v)
		if err != nil {
			panic(fmt.Errorf(
				"row #%d, column #%d (%q) type %T: %s",
				len(r.rows)+1, i, r.cols[i], values[i], err,
			))
		}

		row[i] = v
	}

	r.rows = append(r.rows, row)
	return r
}

// FromCSVString build rows from csv string.
// return the same instance to perform subsequent actions.
// Note that the number of values must match the number
// of columns
func (r *Rows) FromCSVString(s string) *Rows {
	res := strings.NewReader(strings.TrimSpace(s))
	csvReader := csv.NewReader(res)

	for {
		res, err := csvReader.Read()
		if err != nil || res == nil {
			break
		}

		row := make([]driver.Value, len(r.cols))
		for i, v := range res {
			row[i] = CSVColumnParser(strings.TrimSpace(v))
		}
		r.rows = append(r.rows, row)
	}
	return r
}
