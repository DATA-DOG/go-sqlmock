package sqlmock

import (
	"bytes"
	"database/sql/driver"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"
)

const invalidate = "☠☠☠ MEMORY OVERWRITTEN ☠☠☠ "

// CSVColumnParser is a function which converts trimmed csv
// column string to a []byte representation. Currently
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
	raw  [][]byte
}

func (rs *rowSets) Columns() []string {
	return rs.sets[rs.pos].cols
}

func (rs *rowSets) Close() error {
	rs.invalidateRaw()
	rs.ex.rowsWereClosed = true
	return rs.sets[rs.pos].closeErr
}

// advances to next row
func (rs *rowSets) Next(dest []driver.Value) error {
	r := rs.sets[rs.pos]
	r.pos++
	rs.invalidateRaw()
	if r.pos > len(r.rows) {
		return io.EOF // per interface spec
	}

	for i, col := range r.rows[r.pos-1] {
		if b, ok := rawBytes(col); ok {
			rs.raw = append(rs.raw, b)
			dest[i] = b
			continue
		}
		dest[i] = col
	}

	return r.nextErr[r.pos-1]
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

func rawBytes(col driver.Value) (_ []byte, ok bool) {
	val, ok := col.([]byte)
	if !ok || len(val) == 0 {
		return nil, false
	}
	// Copy the bytes from the mocked row into a shared raw buffer, which we'll replace the content of later
	// This allows scanning into sql.RawBytes to correctly become invalid on subsequent calls to Next(), Scan() or Close()
	b := make([]byte, len(val))
	copy(b, val)
	return b, true
}

// Bytes that could have been scanned as sql.RawBytes are only valid until the next call to Next, Scan or Close.
// If those occur, we must replace their content to simulate the shared memory to expose misuse of sql.RawBytes
func (rs *rowSets) invalidateRaw() {
	// Replace the content of slices previously returned
	b := []byte(invalidate)
	for _, r := range rs.raw {
		copy(r, bytes.Repeat(b, len(r)/len(b)+1))
	}
	// Start with new slices for the next scan
	rs.raw = nil
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

// NewRowsFromStruct new Rows from struct reflect with tagName
// tagName default "json"
func NewRowsFromStruct(m interface{}, tagName ...string) (*Rows, error) {
	if m == nil {
		return nil, errors.New("param m is nil")
	}
	val := reflect.ValueOf(m).Elem()
	if val.Kind() != reflect.Struct {
		return nil, errors.New("param type must be struct")
	}
	num := val.NumField()
	if num == 0 {
		return nil, errors.New("no properties available")
	}
	columns := make([]string, 0, num)
	var values []driver.Value
	tag := "json"
	if len(tagName) > 0 {
		tag = tagName[0]
	}
	for i := 0; i < num; i++ {
		f := val.Type().Field(i)
		column := f.Tag.Get(tag)
		if len(column) > 0 {
			columns = append(columns, column)
			values = append(values, val.Field(i))
		}
	}
	if len(columns) == 0 {
		return nil, errors.New("tag not match")
	}
	rows := &Rows{
		cols:      columns,
		nextErr:   make(map[int]error),
		converter: reflectTypeConverter{},
	}
	return rows.AddRow(values...), nil
}

var timeKind = reflect.TypeOf(time.Time{}).Kind()

type reflectTypeConverter struct{}

func (reflectTypeConverter) ConvertValue(v interface{}) (driver.Value, error) {
	rv := v.(reflect.Value)
	switch rv.Kind() {
	case reflect.Ptr:
		// indirect pointers
		if rv.IsNil() {
			return nil, nil
		} else {
			return driver.DefaultParameterConverter.ConvertValue(rv.Elem().Interface())
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return int64(rv.Uint()), nil
	case reflect.Uint64:
		u64 := rv.Uint()
		if u64 >= 1<<63 {
			return nil, fmt.Errorf("uint64 values with high bit set are not supported")
		}
		return int64(u64), nil
	case reflect.Float32, reflect.Float64:
		return rv.Float(), nil
	case reflect.Bool:
		return rv.Bool(), nil
	case reflect.Slice:
		ek := rv.Type().Elem().Kind()
		if ek == reflect.Uint8 {
			return rv.Bytes(), nil
		}
		return nil, fmt.Errorf("unsupported type %T, a slice of %s", v, ek)
	case reflect.String:
		return rv.String(), nil
	case timeKind:
		return rv.Interface().(time.Time), nil
	}
	return nil, fmt.Errorf("unsupported type %T, a %s", v, rv.Kind())
}

// NewRowsFromStructs new Rows from struct slice reflect with tagName
// NOTE: arr must be of the same type
// tagName default "json"
func NewRowsFromStructs(tagName string, arr ...interface{}) (*Rows, error) {
	if len(arr) == 0 {
		return nil, errors.New("param arr is nil")
	}
	typ := reflect.TypeOf(arr[0]).Elem()
	if typ.Kind() != reflect.Struct {
		return nil, errors.New("param type must be struct")
	}
	if typ.NumField() == 0 {
		return nil, errors.New("no properties available")
	}
	var columns []string
	tag := "json"
	if len(tagName) > 0 {
		tag = tagName
	}
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		column := f.Tag.Get(tag)
		if len(column) > 0 {
			columns = append(columns, column)
		}
	}
	if len(columns) == 0 {
		return nil, errors.New("tag not match")
	}
	rows := &Rows{
		cols:      columns,
		nextErr:   make(map[int]error),
		converter: reflectTypeConverter{},
	}
	for _, m := range arr {
		v := m
		val := reflect.ValueOf(v).Elem()
		var values []driver.Value
		for _, column := range columns {
			values = append(values, val.FieldByName(column))
		}
		rows.AddRow(values...)
	}
	return rows, nil
}

// NewRows allows Rows to be created from a
// sql driver.Value slice or from the CSV string and
// to be used as sql driver.Rows.
// Use Sqlmock.NewRows instead if using a custom converter
func NewRows(columns []string) *Rows {
	return &Rows{
		cols:      columns,
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

// AddRows adds multiple rows composed from database driver.Value slice and
// returns the same instance to perform subsequent actions.
func (r *Rows) AddRows(values ...[]driver.Value) *Rows {
	for _, value := range values {
		r.AddRow(value...)
	}

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
